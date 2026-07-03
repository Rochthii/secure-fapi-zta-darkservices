package crypto

import (
	"strings"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if verifier == "" {
		t.Error("Expected non-empty verifier")
	}

	if challenge == "" {
		t.Error("Expected non-empty challenge")
	}

	// S256 challenge length is typically 43 chars
	if len(challenge) != 43 {
		t.Errorf("Expected challenge length to be 43, got %d", len(challenge))
	}
}

func TestCalculateJWKThumbprintAndDPoPProof(t *testing.T) {
	key, err := GenerateDPoPKey()
	if err != nil {
		t.Fatalf("Failed to generate DPoP EC key: %v", err)
	}

	jkt, err := CalculateJWKThumbprint(&key.PublicKey)
	if err != nil {
		t.Fatalf("Failed to compute thumbprint: %v", err)
	}

	if jkt == "" {
		t.Error("Expected non-empty JWK thumbprint")
	}

	// Generate proof
	method := "GET"
	uri := "https://gateway.ziti/api/balance"
	accessToken := "test-access-token"

	proof, err := GenerateDPoPProof(key, method, uri, accessToken)
	if err != nil {
		t.Fatalf("Failed to generate DPoP Proof: %v", err)
	}

	if proof == "" {
		t.Error("Expected non-empty signed proof JWT")
	}

	// The JWT should consist of 3 parts separated by dots
	parts := strings.Split(proof, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT to have 3 parts, got %d", len(parts))
	}
}
