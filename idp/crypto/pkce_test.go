package crypto

import (
	"testing"
)

func TestVerifyPKCE(t *testing.T) {
	// 1. Dữ liệu thử nghiệm chuẩn (S256)
	// code_verifier sinh ngẫu nhiên
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	// code_challenge = Base64URL(SHA-256(verifier))
	expectedChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	// Kiểm tra trường hợp thành công
	if !VerifyPKCE(verifier, expectedChallenge, "S256") {
		t.Error("Xác thực PKCE thất bại với cặp mã hợp lệ")
	}

	// 2. Kiểm tra trường hợp sai verifier
	invalidVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXy"
	if VerifyPKCE(invalidVerifier, expectedChallenge, "S256") {
		t.Error("Xác thực PKCE đáng nhẽ phải thất bại với verifier không hợp lệ")
	}

	// 3. Kiểm tra trường hợp sai phương thức mã hóa (Plain bị cấm trong FAPI 2.0)
	if VerifyPKCE(verifier, verifier, "plain") {
		t.Error("Xác thực PKCE đáng nhẽ phải thất bại với method 'plain'")
	}
}
