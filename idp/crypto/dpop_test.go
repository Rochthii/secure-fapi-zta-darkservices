package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// helper sinh khóa test và cấu trúc JWK
func generateTestKey(t *testing.T) (*ecdsa.PrivateKey, JWK) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	xBytes := priv.PublicKey.X.Bytes()
	yBytes := priv.PublicKey.Y.Bytes()

	// Pad coordinates to 32 bytes
	xPad := make([]byte, 32)
	yPad := make([]byte, 32)
	copy(xPad[32-len(xBytes):], xBytes)
	copy(yPad[32-len(yBytes):], yBytes)

	jwk := JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(xPad),
		Y:   base64.RawURLEncoding.EncodeToString(yPad),
	}

	return priv, jwk
}

// helper tạo DPoP Proof JWT cho test
func createTestDPoPProof(t *testing.T, priv *ecdsa.PrivateKey, jwk JWK, htm, htu, jti string, iat time.Time) string {
	claims := DPoPClaims{
		HTTPMethod: htm,
		HTTPURI:    htu,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(iat),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["typ"] = "dpop+jwt"
	token.Header["jwk"] = jwk

	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestVerifyDPoPProof(t *testing.T) {
	priv, jwk := generateTestKey(t)

	expectedMethod := "POST"
	expectedURI := "https://idp.internal/token"
	jti := "test-jti-12345"

	// 1. Kiểm tra trường hợp DPoP Proof hoàn toàn hợp lệ
	proof := createTestDPoPProof(t, priv, jwk, expectedMethod, expectedURI, jti, time.Now())
	jkt, returnedJti, err := VerifyDPoPProof(proof, expectedMethod, expectedURI, "")
	if err != nil {
		t.Fatalf("Xác thực DPoP hợp lệ thất bại: %v", err)
	}

	expectedJkt, _ := ComputeThumbprint(jwk)
	if jkt != expectedJkt {
		t.Errorf("Thumbprint (jkt) không khớp: expected %s, got %s", expectedJkt, jkt)
	}

	if returnedJti != jti {
		t.Errorf("JTI không khớp: expected %s, got %s", jti, returnedJti)
	}

	// 2. Kiểm tra trường hợp sai HTTP Method (htm)
	badMethodProof := createTestDPoPProof(t, priv, jwk, "GET", expectedURI, jti, time.Now())
	_, _, err = VerifyDPoPProof(badMethodProof, expectedMethod, expectedURI, "")
	if err == nil {
		t.Error("Xác thực DPoP đáng nhẽ phải thất bại do sai HTTP method (htm)")
	}

	// 3. Kiểm tra trường hợp sai URI (htu)
	badURIProof := createTestDPoPProof(t, priv, jwk, expectedMethod, "https://idp.internal/authorize", jti, time.Now())
	_, _, err = VerifyDPoPProof(badURIProof, expectedMethod, expectedURI, "")
	if err == nil {
		t.Error("Xác thực DPoP đáng nhẽ phải thất bại do sai URI (htu)")
	}

	// 4. Kiểm tra trường hợp Proof bị hết hạn (iat cũ hơn 60s)
	expiredProof := createTestDPoPProof(t, priv, jwk, expectedMethod, expectedURI, jti, time.Now().Add(-120*time.Second))
	_, _, err = VerifyDPoPProof(expiredProof, expectedMethod, expectedURI, "")
	if err == nil {
		t.Error("Xác thực DPoP đáng nhẽ phải thất bại do iat đã hết hạn")
	}
}
