package tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

// --- CONFIGURATION ---
const (
	IdpURL       = "http://localhost:8081"
	GatewayURL   = "http://localhost:8080"
	DbSuperuser  = "postgres://postgres:postgres_secure_password_2026@localhost:5432/fapi_db?sslmode=disable"
)

// --- CRYPTO HELPERS (Self-contained) ---

func generatePKCE() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return verifier, challenge, nil
}

func generateDPoPKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func calculateJWKThumbprint(pubKey *ecdsa.PublicKey) (string, error) {
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()
	xPad := make([]byte, 32)
	yPad := make([]byte, 32)
	copy(xPad[32-len(xBytes):], xBytes)
	copy(yPad[32-len(yBytes):], yBytes)

	xStr := base64.RawURLEncoding.EncodeToString(xPad)
	yStr := base64.RawURLEncoding.EncodeToString(yPad)

	jsonStr := fmt.Sprintf(`{"crv":"P-256","kty":"EC","x":"%s","y":"%s"}`, xStr, yStr)
	hash := sha256.Sum256([]byte(jsonStr))
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

func generateDPoPProof(key *ecdsa.PrivateKey, method, uri, accessToken string) (string, error) {
	now := time.Now()
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", err
	}
	jti := hex.EncodeToString(jtiBytes)

	claims := jwt.MapClaims{
		"htm": method,
		"htu": uri,
		"iat": now.Unix(),
		"jti": jti,
	}

	if accessToken != "" {
		hash := sha256.Sum256([]byte(accessToken))
		claims["ath"] = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	xBytes := key.PublicKey.X.Bytes()
	yBytes := key.PublicKey.Y.Bytes()
	xPad := make([]byte, 32)
	yPad := make([]byte, 32)
	copy(xPad[32-len(xBytes):], xBytes)
	copy(yPad[32-len(yBytes):], yBytes)

	token.Header["typ"] = "dpop+jwt"
	token.Header["jwk"] = map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(xPad),
		"y":   base64.RawURLEncoding.EncodeToString(yPad),
	}

	return token.SignedString(key)
}

// --- E2E HELPER FLOWS ---

func getDPoPBoundToken(t *testing.T, clientID, secret string, dpopKey *ecdsa.PrivateKey) (string, error) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		return "", err
	}

	// 1. Authorize
	authURL := fmt.Sprintf("%s/authorize?response_type=code&client_id=%s&client_secret=%s&code_challenge=%s&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback", IdpURL, clientID, secret, challenge)
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authorize failed: status %d, body %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	// 2. Token
	tokenURL := fmt.Sprintf("%s/token", IdpURL)
	dpopProof, err := generateDPoPProof(dpopKey, "POST", tokenURL, "")
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authResp.Code)
	form.Set("code_verifier", verifier)
	form.Set("client_secret", secret)

	tokenReq, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("DPoP", dpopProof)

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", err
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("token exchange failed: status %d, body %s", tokenResp.StatusCode, string(body))
	}

	var tr struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tr); err != nil {
		return "", err
	}

	return tr.AccessToken, nil
}

// --- INTEGRATION TESTS ---

// Test1_ValidFlow tests the successful end-to-end transaction flow for client-alice
func Test1_ValidFlow(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	token, err := getDPoPBoundToken(t, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Call Gateway
	targetURL := GatewayURL + "/api/balance"
	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof for Gateway: %v", err)
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
		t.Fatalf("Gateway request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Gateway returned status %d, body %s", resp.StatusCode, string(body))
	}

	var balanceResp struct {
		Balance float64 `json:"balance"`
		Role    string  `json:"role"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		t.Fatalf("Failed to parse balance: %v", err)
	}

	if balanceResp.Role != "operator" {
		t.Errorf("Expected role operator, got %s", balanceResp.Role)
	}
}

// Test2_ClientSpoofingRejected verifies that client spoofing with invalid secret is rejected
func Test2_ClientSpoofingRejected(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	// Test 1: Alice identity with Bob's secret
	_, err = getDPoPBoundToken(t, "client-alice", "bob-secure-secret-2026", dpopKey)
	if err == nil {
		t.Fatal("Expected error when using wrong secret for client-alice, but succeeded")
	}
	if !strings.Contains(err.Error(), "client authentication failed") {
		t.Errorf("Expected client authentication error, got: %v", err)
	}

	// Test 2: Unregistered client id
	_, err = getDPoPBoundToken(t, "unknown-client", "some-secret", dpopKey)
	if err == nil {
		t.Fatal("Expected error for unregistered client, but succeeded")
	}
	if !strings.Contains(err.Error(), "client_id is not registered") {
		t.Errorf("Expected client_id is not registered error, got: %v", err)
	}
}

// Test3_DPoPReplayRejected verifies DPoP Proof replay attack is rejected
func Test3_DPoPReplayRejected(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	token, err := getDPoPBoundToken(t, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	targetURL := GatewayURL + "/api/balance"
	// Generate ONE proof and use it twice
	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof: %v", err)
	}

	callGateway := func() (int, string) {
		req, _ := http.NewRequest("GET", targetURL, nil)
		req.Header.Set("Authorization", "DPoP "+token)
		req.Header.Set("DPoP", dpopProof)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Gateway request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}

	// First use -> OK
	status1, _ := callGateway()
	if status1 != http.StatusOK {
		t.Fatalf("First request failed with status: %d", status1)
	}

	// Second use (Replay) -> Should be rejected
	status2, body2 := callGateway()
	if status2 != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for DPoP replay, got %d. Body: %s", status2, body2)
	}
	if !strings.Contains(body2, "replay attack detected") {
		t.Errorf("Expected replay attack error message, got: %s", body2)
	}
}

// Test4_ZitiFailClosed verifies that standard TCP connection is rejected when Ziti check is enforced
func Test4_ZitiFailClosed(t *testing.T) {
	// Start a test gateway instance listening on port 8083 with ENFORCE_ZITI=true
	cmd := exec.Command("../gateway/gateway.exe")
	cmd.Dir = "../gateway"
	cmd.Env = append(os.Environ(), "USE_ZITI=false", "ENFORCE_ZITI=true", "PORT=8083")
	
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test gateway: %v", err)
	}
	// Ensure gateway is terminated after test
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Wait 1.5 seconds for server to start
	time.Sleep(1500 * time.Millisecond)

	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	token, err := getDPoPBoundToken(t, "client-alice", "alice-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Request Gateway running on 8083 (with enforced Ziti check) via TCP
	targetURL := "http://localhost:8083/api/balance"
	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, token)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Authorization", "DPoP "+token)
	req.Header.Set("DPoP", dpopProof)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to request Gateway: %v", err)
	}
	defer resp.Body.Close()

	// Should be 403 Forbidden due to missing Ziti Network Identity on TCP conn
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden for direct TCP call, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "missing OpenZiti network identity") {
		t.Errorf("Expected missing Ziti identity error message, got: %s", string(body))
	}
}

// Test5_CrossTenantIsolation verifies that Postgres RLS isolates client data
func Test5_CrossTenantIsolation(t *testing.T) {
	dpopKey, err := generateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP key: %v", err)
	}

	// Login as Bob
	bobToken, err := getDPoPBoundToken(t, "client-bob", "bob-secure-secret-2026", dpopKey)
	if err != nil {
		t.Fatalf("Failed to get token for Bob: %v", err)
	}

	// 1. Query Balance as Bob
	targetURL := GatewayURL + "/api/balance"
	dpopProof, err := generateDPoPProof(dpopKey, "GET", targetURL, bobToken)
	if err != nil {
		t.Fatalf("Failed to generate DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Authorization", "DPoP "+bobToken)
	req.Header.Set("DPoP", dpopProof)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Gateway request failed: %v", err)
	}
	defer resp.Body.Close()

	var balanceResp struct {
		Balance  float64 `json:"balance"`
		TenantID string  `json:"tenant_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		t.Fatalf("Failed to parse balance: %v", err)
	}

	// Bob's balance is 0, Alice's balance is 1000
	if balanceResp.Balance != 0 {
		t.Errorf("Bob should have balance 0, got %f", balanceResp.Balance)
	}
	if balanceResp.TenantID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("Expected Bob's tenant ID, got %s", balanceResp.TenantID)
	}

	// 2. Query Audit Logs as Bob
	logsURL := GatewayURL + "/api/audit-logs"
	dpopProofLogs, _ := generateDPoPProof(dpopKey, "GET", logsURL, bobToken)
	reqLogs, _ := http.NewRequest("GET", logsURL, nil)
	reqLogs.Header.Set("Authorization", "DPoP "+bobToken)
	reqLogs.Header.Set("DPoP", dpopProofLogs)

	respLogs, err := client.Do(reqLogs)
	if err != nil {
		t.Fatalf("Gateway logs request failed: %v", err)
	}
	defer respLogs.Body.Close()

	var logsResp struct {
		AuditLogs []map[string]interface{} `json:"audit_logs"`
	}
	if err := json.NewDecoder(respLogs.Body).Decode(&logsResp); err != nil {
		t.Fatalf("Failed to parse logs: %v", err)
	}
	logs := logsResp.AuditLogs

	// Verify that Bob's logs contain NO Alice records (Tenant ID: 11111111-1111-1111-1111-111111111111)
	for _, logRecord := range logs {
		tenant, _ := logRecord["tenant_id"].(string)
		if tenant == "11111111-1111-1111-1111-111111111111" {
			t.Errorf("Security Violation: Bob can view Alice's audit log record")
		}
	}
}

// Test6_WORMLedgerImmutability directly tests WORM database trigger limitations under superuser
func Test6_WORMLedgerImmutability(t *testing.T) {
	db, err := sql.Open("postgres", DbSuperuser)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Insert a dummy log record first to ensure the row-level trigger fires on it
	var logID int64
	err = db.QueryRow(`INSERT INTO audit_logs (actor_id, tenant_id, action, resource, details, prev_hash, block_hash) 
		VALUES ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'TEST', 'TEST', '{}', '0000000000000000000000000000000000000000000000000000000000000000', '0000000000000000000000000000000000000000000000000000000000000000') 
		RETURNING id`).Scan(&logID)
	if err != nil {
		t.Fatalf("Failed to insert dummy test audit log: %v", err)
	}

	// 1. Attempt to UPDATE the inserted audit log
	_, err = db.Exec("UPDATE audit_logs SET action = 'MALICIOUS_UPDATE' WHERE id = $1", logID)
	if err == nil {
		t.Fatal("Expected DB WORM trigger to reject UPDATE query, but it succeeded")
	}
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("Expected WORM immutability error, got: %v", err)
	}

	// 2. Attempt to DELETE the inserted audit log
	_, err = db.Exec("DELETE FROM audit_logs WHERE id = $1", logID)
	if err == nil {
		t.Fatal("Expected DB WORM trigger to reject DELETE query, but it succeeded")
	}
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("Expected WORM immutability error, got: %v", err)
	}
}
