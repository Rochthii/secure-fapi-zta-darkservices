package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	clientCrypto "client/crypto"
	"client/ziti"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var clientSecrets = map[string]string{
	"client-alice": "alice-secure-secret-2026",
	"client-bob":   "bob-secure-secret-2026",
	"client-evil":  "evil-secure-secret-2026",
}

func main() {
	// Parse CLI Arguments
	identityFlag := flag.String("identity", "client-alice", "OpenZiti client identity (client-alice, client-bob, client-evil)")
	secretFlag := flag.String("secret", "", "Client secret (optional, defaults to pre-registered secrets)")
	commandFlag := flag.String("cmd", "balance", "Command to execute (balance, transfer, logs)")
	amountFlag := flag.Float64("amount", 0.0, "Amount to transfer (only for transfer command)")
	descFlag := flag.String("desc", "Giao dich PTIT Thesis ZTA", "Description of transfer")
	idpURL := flag.String("idp", "http://localhost:8081", "Identity Provider base URL")
	gatewayURL := flag.String("gateway", "http://gateway.ziti", "API Gateway base URL on OpenZiti")
	useZitiFlag := flag.Bool("ziti", true, "Use OpenZiti network (set to false for local debugging)")
	flag.Parse()

	log.Printf("Initializing FAPI 2.0 ZTA Client for identity: %s...", *identityFlag)

	// 1. Khởi tạo khóa ký DPoP của Client (Unique per device session)
	dpopKey, err := clientCrypto.GenerateDPoPKey()
	if err != nil {
		log.Fatalf("CRYPTO ERROR: Failed to generate DPoP EC key: %v", err)
	}
	jkt, _ := clientCrypto.CalculateJWKThumbprint(&dpopKey.PublicKey)
	log.Printf("Generated DPoP public key thumbprint (JKT): %s", jkt)

	// Determine client secret
	secret := *secretFlag
	if secret == "" {
		secret = clientSecrets[*identityFlag]
	}

	// 2. CHẠY LUỒNG PKCE & ĐỔI ACCESS TOKEN TỪ IdP
	accessToken, err := getDPoPBoundToken(*idpURL, *identityFlag, secret, dpopKey)
	if err != nil {
		log.Fatalf("AUTHENTICATION ERROR: Failed to obtain DPoP-bound token: %v", err)
	}
	log.Println("Obtained DPoP-bound Access Token successfully.")

	// 3. THIẾT LẬP KẾT NỐI OPENZITI OVERLAY
	var httpClient *http.Client
	if *useZitiFlag {
		identityPath := fmt.Sprintf("docker/identities/%s.json", *identityFlag)
		log.Printf("Loading Ziti identity config: %s...", identityPath)
		zClient, err := ziti.NewZitiClient(identityPath)
		if err != nil {
			log.Fatalf("ZITI CONFIG ERROR: %v", err)
		}
		defer zClient.Close()

		httpClient = zClient.GetHTTPClient("financial-ledger-service")
		log.Println("Connected to OpenZiti overlay. Requests will be routed through tàng hình tunnel.")
	} else {
		// Fallback sang chạy HTTP/TCP thông thường phục vụ test/debug
		httpClient = &http.Client{}
		log.Println("Fallback mode: using standard TCP client.")
		if *gatewayURL == "http://gateway.ziti" {
			*gatewayURL = "http://localhost:8080"
		}
	}

	// 4. THỰC THI COMMAND NGHIỆP VỤ
	switch strings.ToLower(*commandFlag) {
	case "balance":
		executeGetBalance(httpClient, *gatewayURL, accessToken, dpopKey)
	case "transfer":
		if *amountFlag <= 0 {
			log.Fatalf("VALIDATION ERROR: Transfer amount must be greater than zero")
		}
		executeTransfer(httpClient, *gatewayURL, accessToken, dpopKey, *amountFlag, *descFlag)
	case "logs":
		executeGetAuditLogs(httpClient, *gatewayURL, accessToken, dpopKey)
	default:
		log.Fatalf("UNKNOWN COMMAND: Supported commands are balance, transfer, logs")
	}
}

// getDPoPBoundToken thực hiện PKCE flow + DPoP token exchange
func getDPoPBoundToken(idpBaseURL, clientID, clientSecret string, dpopKey *ecdsa.PrivateKey) (string, error) {
	// A. Sinh PKCE Verifier & Challenge
	verifier, challenge, err := clientCrypto.GeneratePKCE()
	if err != nil {
		return "", err
	}

	// B. Bước 1: Authorization Request (lấy authorization code)
	// Sử dụng Accept: application/json để nhận thẳng code dạng JSON (headless CLI)
	authURL := fmt.Sprintf("%s/authorize?response_type=code&client_id=%s&client_secret=%s&code_challenge=%s&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback", idpBaseURL, clientID, clientSecret, challenge)
	
	req, _ := http.NewRequest("GET", authURL, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("authorize request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authorize returned status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to parse auth code response: %w", err)
	}

	// C. Bước 2: Token Exchange Request (Đổi code lấy DPoP-bound Token)
	tokenURL := fmt.Sprintf("%s/token", idpBaseURL)
	
	// Sinh DPoP Proof cho endpoint /token
	dpopProof, err := clientCrypto.GenerateDPoPProof(dpopKey, "POST", tokenURL, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate DPoP proof for token: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authResp.Code)
	form.Set("code_verifier", verifier)
	form.Set("client_secret", clientSecret)

	tokenReq, _ := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("DPoP", dpopProof) // Ràng buộc token bằng header DPoP

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("token exchange request failed: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("token exchange returned status %d: %s", tokenResp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tr.AccessToken, nil
}

// executeGetBalance gọi API Gateway truy vấn số dư
func executeGetBalance(client *http.Client, gatewayURL, accessToken string, dpopKey *ecdsa.PrivateKey) {
	apiPath := "/api/balance"
	targetURL := gatewayURL + apiPath

	// Sinh DPoP Proof gắn với Access Token (ath)
	dpopProof, err := clientCrypto.GenerateDPoPProof(dpopKey, "GET", targetURL, accessToken)
	if err != nil {
		log.Fatalf("CRYPTO ERROR: Failed to generate API DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Authorization", "DPoP "+accessToken)
	req.Header.Set("DPoP", dpopProof)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("GATEWAY CONNECTION ERROR: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	printAPIResponse(resp.StatusCode, body)
}

// executeTransfer gọi API Gateway thực thi giao dịch
func executeTransfer(client *http.Client, gatewayURL, accessToken string, dpopKey *ecdsa.PrivateKey, amount float64, desc string) {
	apiPath := "/api/transfer"
	targetURL := gatewayURL + apiPath

	reqPayload := map[string]interface{}{
		"amount":      amount,
		"description": desc,
	}
	payloadBytes, _ := json.Marshal(reqPayload)

	// Sinh DPoP Proof gắn với Access Token (ath)
	dpopProof, err := clientCrypto.GenerateDPoPProof(dpopKey, "POST", targetURL, accessToken)
	if err != nil {
		log.Fatalf("CRYPTO ERROR: Failed to generate API DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("POST", targetURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DPoP "+accessToken)
	req.Header.Set("DPoP", dpopProof)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("GATEWAY CONNECTION ERROR: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	printAPIResponse(resp.StatusCode, body)
}

// executeGetAuditLogs gọi API Gateway lấy audit logs
func executeGetAuditLogs(client *http.Client, gatewayURL, accessToken string, dpopKey *ecdsa.PrivateKey) {
	apiPath := "/api/audit-logs"
	targetURL := gatewayURL + apiPath

	// Sinh DPoP Proof gắn với Access Token (ath)
	dpopProof, err := clientCrypto.GenerateDPoPProof(dpopKey, "GET", targetURL, accessToken)
	if err != nil {
		log.Fatalf("CRYPTO ERROR: Failed to generate API DPoP proof: %v", err)
	}

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Authorization", "DPoP "+accessToken)
	req.Header.Set("DPoP", dpopProof)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("GATEWAY CONNECTION ERROR: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	printAPIResponse(resp.StatusCode, body)
}

func printAPIResponse(statusCode int, body []byte) {
	fmt.Printf("\n--- PHẢN HỒI TỪ API GATEWAY (HTTP STATUS: %d) ---\n", statusCode)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(body))
	}
	fmt.Println("------------------------------------------------")
}
