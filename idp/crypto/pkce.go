package crypto

import (
	"crypto/sha256"
	"encoding/base64"
)

// VerifyPKCE kiểm tra code_verifier có khớp với code_challenge theo chuẩn RFC 7636
func VerifyPKCE(codeVerifier, codeChallenge, method string) bool {
	// FAPI 2.0 bắt buộc sử dụng phương thức mã hóa SHA-256 (S256)
	if method != "S256" {
		return false
	}
	
	// Tính hash SHA-256 của code_verifier
	hash := sha256.Sum256([]byte(codeVerifier))
	
	// Mã hóa Base64URL không chứa padding (=)
	computedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	
	return computedChallenge == codeChallenge
}
