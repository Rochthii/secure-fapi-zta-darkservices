package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWK đại diện cho cấu trúc Public Key theo RFC 7517
// FAPI 2.0 bắt buộc dùng ECC P-256 (ES256)
type JWK struct {
	Crv string `json:"crv"`
	Kty string `json:"kty"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// DPoPClaims đại diện cho các claim đặc thù của DPoP Proof theo RFC 9449
type DPoPClaims struct {
	HTTPMethod string `json:"htm"`
	HTTPURI    string `json:"htu"`
	Nonce      string `json:"nonce,omitempty"`
	AccessTokenHash string `json:"ath,omitempty"`
	jwt.RegisteredClaims
}

// ComputeThumbprint tính toán JWK Thumbprint theo RFC 7638 để dùng làm cnf.jkt
func ComputeThumbprint(jwk JWK) (string, error) {
	// Re-marshal cấu trúc theo đúng thứ tự thuộc tính bảng chữ cái (crv, kty, x, y)
	// để tạo ra chuỗi JSON canonicalized chuẩn xác
	canonicalMap := map[string]string{
		"crv": jwk.Crv,
		"kty": jwk.Kty,
		"x":   jwk.X,
		"y":   jwk.Y,
	}
	
	data, err := json.Marshal(canonicalMap)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(data)
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

// JWKToPublicKey chuyển đổi từ cấu trúc JWK sang ecdsa.PublicKey
func JWKToPublicKey(jwk JWK) (*ecdsa.PublicKey, error) {
	if jwk.Kty != "EC" || jwk.Crv != "P-256" {
		return nil, errors.New("unsupported key type or curve, only EC P-256 is allowed")
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y coordinate: %w", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return pubKey, nil
}

// VerifyDPoPProof xác thực DPoP Proof JWT được gửi lên từ Client và trả về jkt, jti
func VerifyDPoPProof(proofJWT, expectedMethod, expectedURI, accessToken string) (string, string, error) {
	var clientJWK JWK

	// Parse và verify chữ ký của DPoP Proof JWT
	token, err := jwt.ParseWithClaims(proofJWT, &DPoPClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Kiểm tra thuật toán ký (chỉ chấp nhận ES256 cho ECC P-256)
		if t.Method.Alg() != "ES256" {
			return nil, fmt.Errorf("invalid signing algorithm: %v, must be ES256", t.Method.Alg())
		}

		// Trích xuất public key (jwk) nằm trong JWT Header
		jwkRaw, ok := t.Header["jwk"]
		if !ok {
			return nil, errors.New("missing jwk in DPoP header")
		}

		// Marshal ngược lại và Unmarshal vào cấu trúc JWK
		jwkBytes, err := json.Marshal(jwkRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse jwk from header: %w", err)
		}

		if err := json.Unmarshal(jwkBytes, &clientJWK); err != nil {
			return nil, fmt.Errorf("failed to unmarshal jwk: %w", err)
		}

		// Chuyển đổi thành ECDSA Public Key để verify chữ ký
		return JWKToPublicKey(clientJWK)
	})

	if err != nil {
		return "", "", fmt.Errorf("invalid DPoP proof signature: %w", err)
	}

	claims, ok := token.Claims.(*DPoPClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid DPoP claims")
	}

	// 1. Kiểm tra HTTP Method (htm)
	if !strings.EqualFold(claims.HTTPMethod, expectedMethod) {
		return "", "", fmt.Errorf("htm claim mismatch: expected %s, got %s", expectedMethod, claims.HTTPMethod)
	}

	// 2. Kiểm tra HTTP URI (htu) - Loại bỏ query parameters nếu có
	cleanURI := strings.Split(expectedURI, "?")[0]
	cleanClaimURI := strings.Split(claims.HTTPURI, "?")[0]
	if !strings.EqualFold(cleanClaimURI, cleanURI) {
		return "", "", fmt.Errorf("htu claim mismatch: expected %s, got %s", cleanURI, cleanClaimURI)
	}

	// 3. Kiểm tra thời gian phát hành (iat) - Cửa sổ cho phép +- 60 giây
	if claims.IssuedAt == nil {
		return "", "", errors.New("missing iat claim in DPoP proof")
	}
	timeDiff := time.Since(claims.IssuedAt.Time)
	if timeDiff < -60*time.Second || timeDiff > 60*time.Second {
		return "", "", fmt.Errorf("dpop proof expired or issued in the future: iat=%v", claims.IssuedAt.Time)
	}

	// 4. Kiểm tra mã hash của Access Token (ath) nếu được yêu cầu
	if accessToken != "" {
		if claims.AccessTokenHash == "" {
			return "", "", errors.New("missing ath claim in DPoP proof while validating access token")
		}
		hash := sha256.Sum256([]byte(accessToken))
		expectedAth := base64.RawURLEncoding.EncodeToString(hash[:])
		if claims.AccessTokenHash != expectedAth {
			return "", "", errors.New("ath claim mismatch: proof is not bound to this access token")
		}
	}

	// Tính toán JWK Thumbprint để định danh Client thiết bị
	jkt, err := ComputeThumbprint(clientJWK)
	if err != nil {
		return "", "", fmt.Errorf("failed to compute client key thumbprint: %w", err)
	}

	return jkt, claims.ID, nil
}

