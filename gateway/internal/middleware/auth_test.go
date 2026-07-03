package middleware

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock connection that implements SourceIdentifier
type mockZitiConn struct {
	net.Conn
	identity string
}

func (m *mockZitiConn) SourceIdentifier() string {
	return m.identity
}

// Mock connection that does NOT implement SourceIdentifier
type mockPlainConn struct {
	net.Conn
}

func TestGetZitiIdentity(t *testing.T) {
	// Test case 1: Connection is nil
	if id := GetZitiIdentity(nil); id != "" {
		t.Errorf("Expected empty string for nil connection, got %q", id)
	}

	// Test case 2: Connection does not implement SourceIdentifier
	plainConn := &mockPlainConn{}
	if id := GetZitiIdentity(plainConn); id != "" {
		t.Errorf("Expected empty string for plain connection, got %q", id)
	}

	// Test case 3: Connection implements SourceIdentifier
	zitiConn := &mockZitiConn{identity: "test-client-alice"}
	if id := GetZitiIdentity(zitiConn); id != "test-client-alice" {
		t.Errorf("Expected identity 'test-client-alice', got %q", id)
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userClaims     *TokenClaims // nil if no claims injected
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Unauthorized: No claims in context",
			userClaims:     nil,
			allowedRoles:   []string{"operator"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Authorized: Single role matches",
			userClaims: &TokenClaims{
				Role: "operator",
			},
			allowedRoles:   []string{"operator"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Authorized: Matches one of multiple roles",
			userClaims: &TokenClaims{
				Role: "viewer",
			},
			allowedRoles:   []string{"operator", "viewer"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Forbidden: Role mismatch",
			userClaims: &TokenClaims{
				Role: "viewer",
			},
			allowedRoles:   []string{"operator"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock final handler that returns 200 OK
			okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap in RequireRole middleware
			middleware := RequireRole(tt.allowedRoles...)(okHandler)

			// Setup request & recorder
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.userClaims != nil {
				ctx := context.WithValue(req.Context(), ClaimsKey, *tt.userClaims)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			// Run
			middleware.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	ctx := context.Background()

	// Test 1: Extract when not present
	_, ok := GetClaimsFromContext(ctx)
	if ok {
		t.Error("Expected GetClaimsFromContext to return ok=false when claims not in context")
	}

	// Test 2: Extract when present
	expectedClaims := TokenClaims{
		Sub:      "alice",
		TenantID: "tenant-a",
		Role:     "operator",
		Scope:    "read write",
	}
	ctxWithClaims := context.WithValue(ctx, ClaimsKey, expectedClaims)
	claims, ok := GetClaimsFromContext(ctxWithClaims)
	if !ok {
		t.Fatal("Expected GetClaimsFromContext to succeed")
	}
	if claims.Sub != expectedClaims.Sub || claims.Role != expectedClaims.Role {
		t.Errorf("Extracted claims did not match: got %+v, expected %+v", claims, expectedClaims)
	}
}
