package telemetry

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type MetricsStore struct {
	mu            sync.RWMutex
	requestsTotal map[string]int64 // key: handler:status

	dpopVerifyUs  int64
	tokenVerifyUs int64
	zitiVerifyUs  int64

	dbRlsContextUs int64
	dbWormExecUs   int64
}

var globalStore = &MetricsStore{
	requestsTotal: make(map[string]int64),
}

// IncrementRequestCounter increments HTTP request count for a given handler and status code
func IncrementRequestCounter(handler string, status int) {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	key := fmt.Sprintf("%s:%d", handler, status)
	globalStore.requestsTotal[key]++
}

// RecordSecurityOverhead records the security verification latencies in microseconds
func RecordSecurityOverhead(dpop, token, ziti int64) {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	globalStore.dpopVerifyUs = dpop
	globalStore.tokenVerifyUs = token
	globalStore.zitiVerifyUs = ziti
}

// RecordDbLatency records database operation latencies in microseconds
func RecordDbLatency(rls, worm int64) {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	globalStore.dbRlsContextUs = rls
	globalStore.dbWormExecUs = worm
}

// ServeMetrics formats and writes the metrics in Prometheus text exposition format
func ServeMetrics(w http.ResponseWriter, r *http.Request) {
	globalStore.mu.RLock()
	defer globalStore.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// 1. Output HTTP request counts
	fmt.Fprintln(w, "# HELP gateway_requests_total Total number of HTTP requests processed by the API Gateway.")
	fmt.Fprintln(w, "# TYPE gateway_requests_total counter")
	for key, count := range globalStore.requestsTotal {
		handler := ""
		status := 0
		// Split by last colon
		idx := strings.LastIndex(key, ":")
		if idx != -1 {
			handler = key[:idx]
			fmt.Sscanf(key[idx+1:], "%d", &status)
		} else {
			handler = key
		}
		fmt.Fprintf(w, "gateway_requests_total{handler=\"%s\",status=\"%d\"} %d\n", handler, status, count)
	}
	fmt.Fprintln(w)

	// 2. Output Security Overheads
	fmt.Fprintln(w, "# HELP gateway_security_overhead_microseconds Latency of security layer checks in microseconds.")
	fmt.Fprintln(w, "# TYPE gateway_security_overhead_microseconds gauge")
	fmt.Fprintf(w, "gateway_security_overhead_microseconds{stage=\"dpop\"} %d\n", globalStore.dpopVerifyUs)
	fmt.Fprintf(w, "gateway_security_overhead_microseconds{stage=\"token\"} %d\n", globalStore.tokenVerifyUs)
	fmt.Fprintf(w, "gateway_security_overhead_microseconds{stage=\"ziti\"} %d\n", globalStore.zitiVerifyUs)
	fmt.Fprintln(w)

	// 3. Output DB Latencies
	fmt.Fprintln(w, "# HELP gateway_db_latency_microseconds Latency of database transaction and security operations in microseconds.")
	fmt.Fprintln(w, "# TYPE gateway_db_latency_microseconds gauge")
	fmt.Fprintf(w, "gateway_db_latency_microseconds{operation=\"rls_context\"} %d\n", globalStore.dbRlsContextUs)
	fmt.Fprintf(w, "gateway_db_latency_microseconds{operation=\"worm_exec\"} %d\n", globalStore.dbWormExecUs)
}
