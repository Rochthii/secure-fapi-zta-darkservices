package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWKKey represents a public key in JWKS
type JWKKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
}

type JWKS struct {
	Keys []JWKKey `json:"keys"`
}

// JWKSCache handles fetching and caching of IdP's JWKS
type JWKSCache struct {
	jwksURL    string
	keys       map[string]*ecdsa.PublicKey
	mutex      sync.RWMutex
	lastUpdate time.Time
	ttl        time.Duration
}

func NewJWKSCache(jwksURL string, ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		jwksURL: jwksURL,
		keys:    make(map[string]*ecdsa.PublicKey),
		ttl:     ttl,
	}
}

// GetPublicKey retrieves the ECDSA public key for the given kid (key ID)
func (c *JWKSCache) GetPublicKey(kid string) (*ecdsa.PublicKey, error) {
	c.mutex.RLock()
	key, ok := c.keys[kid]
	cacheAge := time.Since(c.lastUpdate)
	c.mutex.RUnlock()

	if ok && cacheAge < c.ttl {
		return key, nil
	}

	// Update cache
	if err := c.FetchKeys(); err != nil {
		if ok {
			// Fallback to expired cache if fetch fails
			return key, nil
		}
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()
	key, ok = c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key ID '%s' not found in JWKS", kid)
	}
	return key, nil
}

// FetchKeys makes an HTTP request to fetch keys from the IdP JWKS endpoint
func (c *JWKSCache) FetchKeys() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(c.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code from JWKS: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	newKeys := make(map[string]*ecdsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty == "EC" && k.Crv == "P-256" {
			pubKey, err := JWKToECDSA(k.X, k.Y)
			if err != nil {
				continue
			}
			newKeys[k.Kid] = pubKey
		}
	}

	c.keys = newKeys
	c.lastUpdate = time.Now()
	return nil
}

// JWKToECDSA parses X and Y base64url coordinates into an ecdsa.PublicKey
func JWKToECDSA(xStr, yStr string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return pubKey, nil
}

// CalculateJWKThumbprint computes the RFC 7638 SHA-256 thumbprint of an EC JWK
func CalculateJWKThumbprint(xStr, yStr string) (string, error) {
	// Fields MUST be alphabetically sorted: crv, kty, x, y
	jsonStr := fmt.Sprintf(`{"crv":"P-256","kty":"EC","x":"%s","y":"%s"}`, xStr, yStr)
	hash := sha256.Sum256([]byte(jsonStr))
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

// VerifyDPoPProof verifies the DPoP Proof JWT
// Returns (thumbprint, jti, error)
func VerifyDPoPProof(proofJWT, method, targetURI, accessToken string) (string, string, error) {
	var jwkMap map[string]interface{}

	// Parse proof JWT and extract client public key from header
	token, err := jwt.Parse(proofJWT, func(t *jwt.Token) (interface{}, error) {
		// Verify signature method
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		// Extract jwk from header
		jwkVal, ok := t.Header["jwk"]
		if !ok {
			return nil, fmt.Errorf("missing jwk claim in DPoP header")
		}

		var okCast bool
		jwkMap, okCast = jwkVal.(map[string]interface{})
		if !okCast {
			return nil, fmt.Errorf("invalid jwk format in DPoP header")
		}

		// Verify EC P-256
		if jwkMap["kty"] != "EC" || jwkMap["crv"] != "P-256" {
			return nil, fmt.Errorf("DPoP key must be EC P-256")
		}

		xStr, _ := jwkMap["x"].(string)
		yStr, _ := jwkMap["y"].(string)
		if xStr == "" || yStr == "" {
			return nil, fmt.Errorf("missing X/Y coordinates in DPoP key")
		}

		return JWKToECDSA(xStr, yStr)
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid DPoP proof signature: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid DPoP claims")
	}

	// Validate claims: htm, htu, iat, jti
	htm, _ := claims["htm"].(string)
	if !strings.EqualFold(htm, method) {
		return "", "", fmt.Errorf("invalid htm claim: expected %s, got %s", method, htm)
	}

	htu, _ := claims["htu"].(string)
	// Check if targetURI ends with htu path to allow flexible protocol/host matching in Ziti network proxying
	if !verifyHTU(htu, targetURI) {
		return "", "", fmt.Errorf("invalid htu claim: expected match for %s, got %s", targetURI, htu)
	}

	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return "", "", fmt.Errorf("missing or invalid iat claim")
	}
	iat := time.Unix(int64(iatFloat), 0)
	if time.Since(iat) > 60*time.Second || time.Until(iat) > 60*time.Second {
		return "", "", fmt.Errorf("DPoP proof expired or issued in the future (iat: %v)", iat)
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return "", "", fmt.Errorf("missing jti claim in DPoP proof")
	}

	// Verify Access Token Binding (ath claim) if accessToken is provided
	if accessToken != "" {
		ath, _ := claims["ath"].(string)
		if ath == "" {
			return "", "", fmt.Errorf("missing ath claim in DPoP proof (required for request validation)")
		}

		expectedAth := sha256.Sum256([]byte(accessToken))
		encodedExpectedAth := base64.RawURLEncoding.EncodeToString(expectedAth[:])
		if ath != encodedExpectedAth {
			return "", "", fmt.Errorf("ath claim mismatch: proof is not bound to this access token")
		}
	}

	// Calculate thumbprint of client public key (cnf.jkt)
	xStr, _ := jwkMap["x"].(string)
	yStr, _ := jwkMap["y"].(string)
	jkt, err := CalculateJWKThumbprint(xStr, yStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to calculate JWK thumbprint: %w", err)
	}

	return jkt, jti, nil
}

// verifyHTU verifies that targetURI matches htu path flexibly
func verifyHTU(htu, targetURI string) bool {
	if htu == targetURI {
		return true
	}
	// Strip scheme and host if comparing path-only
	cleanHTU := stripHostAndScheme(htu)
	cleanTarget := stripHostAndScheme(targetURI)
	return cleanHTU == cleanTarget
}

func stripHostAndScheme(uri string) string {
	if idx := strings.Index(uri, "://"); idx != -1 {
		uri = uri[idx+3:]
	}
	if idx := strings.Index(uri, "/"); idx != -1 {
		return uri[idx:]
	}
	return "/"
}
