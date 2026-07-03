package tests

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLatencyBreakdown(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	token, err := getDPoPBoundToken(t, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Target Gateway URL
	targetURL := GatewayURL + "/api/balance"

	// Measure client-perceived round-trip start
	clientStart := time.Now()

	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof: %v", err)
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "DPoP "+token)
	req.Header.Set("DPoP", dpopProof)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	clientDuration := time.Since(clientStart)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Parse custom performance headers from the response
	dpopUs, _ := strconv.ParseInt(resp.Header.Get("X-Perf-Dpop-Verify-Us"), 10, 64)
	tokenUs, _ := strconv.ParseInt(resp.Header.Get("X-Perf-Token-Verify-Us"), 10, 64)
	zitiUs, _ := strconv.ParseInt(resp.Header.Get("X-Perf-Ziti-Verify-Us"), 10, 64)
	dbContextUs, _ := strconv.ParseInt(resp.Header.Get("X-Perf-Db-Context-Us"), 10, 64)
	dbExecUs, _ := strconv.ParseInt(resp.Header.Get("X-Perf-Db-Exec-Us"), 10, 64)

	totalGatewayUs := dpopUs + tokenUs + zitiUs + dbContextUs + dbExecUs

	// Output latency breakdown report
	fmt.Println("========================================================")
	fmt.Println("             PERFORMANCE LATENCY BREAKDOWN              ")
	fmt.Println("========================================================")
	fmt.Printf("1. Client-side DPoP Proof Signing: %d µs\n", clientDuration.Microseconds()-totalGatewayUs) // Approximate
	fmt.Printf("2. Gateway DPoP Signature Verify:   %d µs\n", dpopUs)
	fmt.Printf("3. Gateway Token JWKS Verify:       %d µs\n", tokenUs)
	fmt.Printf("4. Gateway Ziti Identity Check:     %d µs\n", zitiUs)
	fmt.Printf("5. Postgres RLS Tenant Context Set: %d µs\n", dbContextUs)
	fmt.Printf("6. Postgres WORM Trigger & Hash:    %d µs\n", dbExecUs)
	fmt.Println("--------------------------------------------------------")
	fmt.Printf("Total Server Processing Time:       %d µs (%.2f ms)\n", totalGatewayUs, float64(totalGatewayUs)/1000.0)
	fmt.Printf("Client Round-Trip Duration:         %d µs (%.2f ms)\n", clientDuration.Microseconds(), float64(clientDuration.Milliseconds()))
	fmt.Println("========================================================")
}

func BenchmarkEndToEndFlow(b *testing.B) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		b.Fatalf("Failed to generate DPoP key: %v", err)
	}

	// Obtain a valid token to reuse (mimics client caching the JWT)
	var mockT testing.T
	token, err := getDPoPBoundToken(&mockT, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		b.Fatalf("Failed to pre-authenticate: %v", err)
	}

	targetURL := GatewayURL + "/api/balance"
	client := &http.Client{}

	b.ResetTimer()

	// Run parallel benchmark to test concurrency / throughput
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Generate a unique DPoP proof for each request (required to avoid JTI replay blocks)
			dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
			if err != nil {
				b.Fatalf("Failed to generate DPoP proof: %v", err)
			}

			req, _ := http.NewRequest("GET", targetURL, nil)
			req.Header.Set("Authorization", "DPoP "+token)
			req.Header.Set("DPoP", dpopProof)

			resp, err := client.Do(req)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
			}
		}
	})
}

func TestPrometheusMetrics(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	token, err := getDPoPBoundToken(t, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	targetURL := GatewayURL + "/api/balance"
	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Authorization", "DPoP "+token)
	req.Header.Set("DPoP", dpopProof)

	client := &http.Client{}
	resp1, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to query API: %v", err)
	}
	resp1.Body.Close()

	resp, err := http.Get(GatewayURL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to fetch /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	body := string(bodyBytes)

	// Check for Prometheus metrics
	requiredMetrics := []string{
		"gateway_requests_total",
		"gateway_security_overhead_microseconds",
		"gateway_db_latency_microseconds",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("Metrics output missing required metric name %q", metric)
		}
	}

	fmt.Println("========================================================")
	fmt.Println("             PROMETHEUS METRICS EXPOSITION              ")
	fmt.Println("========================================================")
	fmt.Print(body)
	fmt.Println("========================================================")
}
