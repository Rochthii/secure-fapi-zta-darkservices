package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"secure-fapi-zta-darkservices/idp/config"
	"secure-fapi-zta-darkservices/idp/crypto"
	"secure-fapi-zta-darkservices/idp/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenHandler xử lý yêu cầu đổi authorization_code lấy DPoP-bound Access Token
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse x-www-form-urlencoded
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid_request: failed to parse form parameters", http.StatusBadRequest)
		return
	}

	grantType := r.Form.Get("grant_type")
	code := r.Form.Get("code")
	codeVerifier := r.Form.Get("code_verifier")

	// 1. Xác thực tham số đầu vào
	if grantType != "authorization_code" {
		http.Error(w, "unsupported_grant_type: grant_type must be 'authorization_code'", http.StatusBadRequest)
		return
	}
	if code == "" {
		http.Error(w, "invalid_request: code is required", http.StatusBadRequest)
		return
	}
	if codeVerifier == "" {
		http.Error(w, "invalid_request: code_verifier is required", http.StatusBadRequest)
		return
	}

	// 2. Xác thực PKCE
	codeChallenge, ok := store.GetStore().GetAndRemoveAuthCode(code)
	if !ok {
		http.Error(w, "invalid_grant: authorization code is invalid or expired", http.StatusBadRequest)
		return
	}

	if !crypto.VerifyPKCE(codeVerifier, codeChallenge, "S256") {
		http.Error(w, "invalid_grant: PKCE verification failed", http.StatusBadRequest)
		return
	}

	// 3. Xác thực DPoP Proof
	dpopHeader := r.Header.Get("DPoP")
	if dpopHeader == "" {
		http.Error(w, "invalid_request: missing DPoP header proof", http.StatusBadRequest)
		return
	}

	// Tạo URL đầy đủ của endpoint token để xác thực htu
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fullURI := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)

	jkt, jti, err := crypto.VerifyDPoPProof(dpopHeader, "POST", fullURI, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid_dpop_proof: %v", err), http.StatusBadRequest)
		return
	}

	// Kiểm tra Replay Attack
	if store.GetStore().IsJTIUsedAndSave(jti, 2*time.Minute) {
		http.Error(w, "invalid_dpop_proof: replay attack detected (jti already used)", http.StatusBadRequest)
		return
	}

	// 4. Phát hành DPoP-bound Access Token (ES256)
	tokenExpiration := 60 * time.Second // Hạn dùng cực ngắn 60s
	now := time.Now()

	claims := jwt.MapClaims{
		"iss":       config.AppConfig.Issuer,
		"sub":       "user-123456", // Định danh người dùng
		"tenant_id": "88888888-8888-8888-8888-888888888888", // Định danh tenant giả lập
		"role":      "operator",
		"scope":     "transfer balance",
		"exp":       now.Add(tokenExpiration).Unix(),
		"iat":       now.Unix(),
		"cnf": jwt.MapClaims{
			"jkt": jkt, // Ràng buộc token với mã thumbprint public key của Client
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = "idp-signing-key"

	accessTokenString, err := token.SignedString(config.AppConfig.PrivateKey)
	if err != nil {
		http.Error(w, "server_error: failed to sign token", http.StatusInternalServerError)
		return
	}

	// Tạo Refresh Token giả lập dùng 1 lần (Opaque string)
	refreshToken := "rt_" + jti // Liên kết refresh token với jti để theo dõi

	resp := TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "DPoP",
		ExpiresIn:    int(tokenExpiration.Seconds()),
		RefreshToken: refreshToken,
	}

	json.NewEncoder(w).Encode(resp)
}
