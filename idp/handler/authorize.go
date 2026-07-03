package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"secure-fapi-zta-darkservices/idp/store"
	"time"
)

// AuthorizeHandler xử lý luồng cấp authorization code (PKCE)
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	// Trích xuất các tham số từ Query Parameters
	q := r.URL.Query()
	responseType := q.Get("response_type")
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	// 1. Kiểm tra tính hợp lệ của luồng OAuth 2.1
	if responseType != "code" {
		http.Error(w, "invalid_request: response_type must be 'code'", http.StatusBadRequest)
		return
	}
	if clientID == "" {
		http.Error(w, "invalid_request: client_id is required", http.StatusBadRequest)
		return
	}
	if redirectURI == "" {
		http.Error(w, "invalid_request: redirect_uri is required", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" {
		http.Error(w, "invalid_request: code_challenge is required", http.StatusBadRequest)
		return
	}
	
	// FAPI 2.0 bắt buộc sử dụng phương thức S256 (SHA-256) để mã hóa verifier
	if codeChallengeMethod != "S256" {
		http.Error(w, "invalid_request: code_challenge_method must be 'S256'", http.StatusBadRequest)
		return
	}

	// 2. Sinh authorization_code ngẫu nhiên (16 bytes)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "server_error: failed to generate code", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(bytes)

	// 3. Lưu code và challenge vào RAMStore (Hạn dùng 5 phút)
	store.GetStore().SaveAuthCode(code, codeChallenge, 5*time.Minute)

	// 4. Thực hiện Redirect quay lại ứng dụng Client kèm code
	targetURL, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid_request: malformed redirect_uri", http.StatusBadRequest)
		return
	}

	targetQuery := targetURL.Query()
	targetQuery.Set("code", code)
	// Trả về state nếu client gửi lên
	if state := q.Get("state"); state != "" {
		targetQuery.Set("state", state)
	}
	targetURL.RawQuery = targetQuery.Encode()

	// Cho phép Client API gọi không cần redirect thực tế nếu yêu cầu JSON (phục vụ headless tests)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":"` + code + `"}`))
		return
	}

	http.Redirect(w, r, targetURL.String(), http.StatusFound)
}
