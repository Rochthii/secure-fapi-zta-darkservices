package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GeneratePKCE generates a code_verifier and code_challenge (S256)
func GeneratePKCE() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

// GenerateDPoPKey generates a new ECDSA P-256 private key
func GenerateDPoPKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// CalculateJWKThumbprint computes the RFC 7638 SHA-256 thumbprint of an EC public key
func CalculateJWKThumbprint(pubKey *ecdsa.PublicKey) (string, error) {
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	// Ensure they are padded to 32 bytes (256 bits)
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

// GenerateDPoPProof creates and signs a DPoP Proof JWT
func GenerateDPoPProof(key *ecdsa.PrivateKey, method, uri, accessToken string) (string, error) {
	now := time.Now()
	
	// Create JWT ID (jti)
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", err
	}
	jti := hex.EncodeToString(jtiBytes)

	// Prepare claims
	claims := jwt.MapClaims{
		"htm": method,
		"htu": uri,
		"iat": now.Unix(),
		"jti": jti,
	}

	// Calculate ath (AccessToken Hash) if accessToken is provided
	if accessToken != "" {
		hash := sha256.Sum256([]byte(accessToken))
		claims["ath"] = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	// Populate JWT Header with public key JWK representation
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

	// Sign the token
	signedProof, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign DPoP proof: %w", err)
	}

	return signedProof, nil
}
