package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gateway/internal/auth"
	"gateway/internal/pdpclient"
	"gateway/internal/telemetry"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwksCache   *auth.JWKSCache
	enforceZiti bool
	// pdpClient là gRPC client tới standalone-policy-engine (PDP).
	// nil chỉ trong test — khi nil EnforcePolicy sẽ fail-closed.
	pdpClient *pdpclient.PDPClient
}

type TokenClaims struct {
	Sub      string `json:"sub"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	Scope    string `json:"scope"`
	JKT      string `json:"jkt"` // DPoP Key Thumbprint — device fingerprint cho ABAC
}

// NewAuthMiddleware tạo middleware xác thực với gRPC PDP client.
// pdpClient phải được khởi tạo trước từ gateway/main.go.
func NewAuthMiddleware(jwksURL string, enforceZiti bool, pdpClient *pdpclient.PDPClient) *AuthMiddleware {
	return &AuthMiddleware{
		jwksCache:   auth.NewJWKSCache(jwksURL, 5*time.Minute),
		enforceZiti: enforceZiti,
		pdpClient:   pdpClient,
	}
}

// SecureAPI is the core continuous verification middleware chain
func (m *AuthMiddleware) SecureAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Kiểm tra tiêu đề Authorization: DPoP <token>
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "invalid_token: missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "DPoP") {
			http.Error(w, "invalid_token: Authorization type must be DPoP", http.StatusUnauthorized)
			return
		}
		accessToken := parts[1]

		// 2. Kiểm tra tiêu đề DPoP Proof
		dpopHeader := r.Header.Get("DPoP")
		if dpopHeader == "" {
			http.Error(w, "invalid_dpop_proof: missing DPoP header proof", http.StatusUnauthorized)
			return
		}

		dpopStart := time.Now()

		// Xác định URI gọi thực tế
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		fullURI := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)

		// 3. Xác thực DPoP Proof (Chữ ký, claims htm/htu/iat/ath)
		jkt, jti, err := auth.VerifyDPoPProof(dpopHeader, r.Method, fullURI, accessToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid_dpop_proof: %v", err), http.StatusUnauthorized)
			return
		}

		// Chống Replay: Check JTI trùng lặp
		if auth.GetJTICache().IsJTIUsedAndSave(jti, 2*time.Minute) {
			http.Error(w, "invalid_dpop_proof: replay attack detected (jti already used)", http.StatusUnauthorized)
			return
		}
		dpopTime := time.Since(dpopStart).Microseconds()

		tokenStart := time.Now()

		// 4. Xác thực Access Token JWT từ IdP
		token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected token signing method: %v", t.Header["alg"])
			}
			kid, _ := t.Header["kid"].(string)
			if kid == "" {
				return nil, fmt.Errorf("missing kid in token header")
			}
			return m.jwksCache.GetPublicKey(kid)
		})

		if err != nil || !token.Valid {
			http.Error(w, fmt.Sprintf("invalid_token: signature verification failed: %v", err), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid_token: invalid claims format", http.StatusUnauthorized)
			return
		}

		// 5. Kiểm tra ràng buộc DPoP JKT (cnf.jkt == jkt)
		cnfVal, ok := claims["cnf"].(map[string]interface{})
		if !ok {
			http.Error(w, "invalid_token: missing cnf claim", http.StatusUnauthorized)
			return
		}
		tokenJkt, _ := cnfVal["jkt"].(string)
		if tokenJkt != jkt {
			http.Error(w, "invalid_token: token jkt binding mismatch (sender-constraining failed)", http.StatusUnauthorized)
			return
		}
		tokenTime := time.Since(tokenStart).Microseconds()

		zitiStart := time.Now()

		// 6. XÁC THỰC LIÊN TỤC mTLS - Đối chiếu danh tính mạng với Access Token
		conn := GetConnFromContext(r.Context())
		zitiIdentity := GetZitiIdentity(conn)
		tokenSub, _ := claims["sub"].(string)

		// Nếu sử dụng Ziti, bắt buộc identity mạng trùng khớp với token sub
		// Nếu zitiIdentity rỗng khi yêu cầu Ziti, từ chối luôn (fail-closed)
		if m.enforceZiti {
			if zitiIdentity == "" {
				http.Error(w, "forbidden: missing OpenZiti network identity in a secure context", http.StatusForbidden)
				return
			}
			if tokenSub != zitiIdentity {
				http.Error(w, fmt.Sprintf("forbidden: network identity '%s' does not match token subject '%s'", zitiIdentity, tokenSub), http.StatusForbidden)
				return
			}
		} else {
			// Chế độ debug/fallback: Nếu có zitiIdentity thì vẫn kiểm tra trùng khớp
			if zitiIdentity != "" && tokenSub != zitiIdentity {
				http.Error(w, fmt.Sprintf("forbidden: network identity '%s' does not match token subject '%s'", zitiIdentity, tokenSub), http.StatusForbidden)
				return
			}
		}
		zitiTime := time.Since(zitiStart).Microseconds()

		// Gắn các chỉ số hiệu năng vào HTTP Response Headers
		w.Header().Set("X-Perf-Dpop-Verify-Us", fmt.Sprintf("%d", dpopTime))
		w.Header().Set("X-Perf-Token-Verify-Us", fmt.Sprintf("%d", tokenTime))
		w.Header().Set("X-Perf-Ziti-Verify-Us", fmt.Sprintf("%d", zitiTime))

		// Ghi nhận vào Prometheus Exporter
		telemetry.RecordSecurityOverhead(dpopTime, tokenTime, zitiTime)

		// Trích xuất claims hợp lệ và đưa vào context của request
		tenantID, _ := claims["tenant_id"].(string)
		role, _ := claims["role"].(string)
		scope, _ := claims["scope"].(string)

		// Trích xuất DPoP JKT (device fingerprint) từ cnf claim để dùng trong ABAC
		// Phải dùng tên khác jkt — biến jkt đã được khai báo từ auth.VerifyDPoPProof() phía trên
		var dpopJKT string
		if cnf, ok := claims["cnf"].(map[string]interface{}); ok {
			dpopJKT, _ = cnf["jkt"].(string)
		}

		tClaims := TokenClaims{
			Sub:      tokenSub,
			TenantID: tenantID,
			Role:     role,
			Scope:    scope,
			JKT:      dpopJKT,
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, tClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if the authenticated user has the necessary role
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(TokenClaims)
			if !ok {
				http.Error(w, "unauthorized: missing claims context", http.StatusUnauthorized)
				return
			}

			roleAllowed := false
			for _, role := range allowedRoles {
				if strings.EqualFold(claims.Role, role) {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				http.Error(w, fmt.Sprintf("forbidden: role '%s' is not authorized to access this resource", claims.Role), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaimsFromContext extracts token claims from request context
func GetClaimsFromContext(ctx context.Context) (TokenClaims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(TokenClaims)
	return claims, ok
}

// EnforcePolicy là PEP (Policy Enforcement Point) — gọi gRPC PDP để quyết định phân quyền.
//
// Luồng xử lý:
//  1. Trích xuất resource và action từ HTTP request
//  2. Xây dựng enriched ABAC context (IP, time, DPoP JKT, Ziti identity, scope)
//  3. Gọi gRPC PDP.CheckAccess() với timeout 80ms
//  4. Nếu PDP trả DENY → 403 Forbidden
//  5. Nếu PDP không phản hồi → 503 Service Unavailable (fail-closed)
//  6. Nếu PDP trả ALLOW → gắn matched_policy_id vào response header và cho phép đi tiếp
func (m *AuthMiddleware) EnforcePolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized: missing claims context", http.StatusUnauthorized)
			return
		}

		// Kiểm tra pdpClient — fail-closed nếu chưa được khởi tạo
		if m.pdpClient == nil {
			http.Error(w, "forbidden: policy enforcement point not initialized", http.StatusServiceUnavailable)
			return
		}

		// 1. Map URL path → resource name
		resource := "unknown"
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/balance"):
			resource = "resource:balance"
		case strings.HasPrefix(r.URL.Path, "/api/transfer"):
			resource = "resource:transfer"
		case strings.HasPrefix(r.URL.Path, "/api/audit-logs"):
			resource = "resource:audit-logs"
		}

		// 2. Map HTTP method → action
		action := "READ"
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			action = "CREATE"
		}

		// 3. Trích xuất network identity của client (từ Ziti conn hoặc RemoteAddr)
		conn := GetConnFromContext(r.Context())
		zitiIdentity := GetZitiIdentity(conn)
		clientIP := r.RemoteAddr
		if zitiIdentity != "" {
			// Khi qua Ziti overlay, RemoteAddr là địa chỉ virtual — dùng Ziti identity thay thế
			clientIP = zitiIdentity
		}

		// 4. Xây dựng enriched ABAC context — tất cả attributes này được đánh giá bởi AST evaluator của PDP
		abacCtx := map[string]string{
			"ip_address":     clientIP,
			"request_time":   time.Now().UTC().Format("15:04:05Z"),
			"http_method":    r.Method,
			"dpop_jkt":       claims.JKT,      // Device fingerprint — dùng để whitelist thiết bị
			"ziti_identity":  zitiIdentity,    // Network identity từ OpenZiti
			"tenant_id":      claims.TenantID,
			"role":           claims.Role,
			"scope":          claims.Scope,
			"token_subject":  claims.Sub,
		}

		// 5. Gọi gRPC PDP — subject là dạng "role:<role>" để PDP tra cứu trong Trie
		// Dùng role là subject chính để PDP có thể tra cứu theo Role-based policy
		subject := fmt.Sprintf("role:%s", claims.Role)

		pdpStart := time.Now()
		allow, matchedPolicyID, httpStatus, err := m.pdpClient.CheckAccess(
			r.Context(),
			claims.TenantID,
			subject,
			action,
			resource,
			abacCtx,
		)
		pdpLatency := time.Since(pdpStart).Microseconds()

		// Ghi nhận latency PDP vào response header để monitoring
		w.Header().Set("X-Perf-PDP-Verify-Us", fmt.Sprintf("%d", pdpLatency))

		// Ghi nhận vào Prometheus
		telemetry.RecordPDPOverhead(pdpLatency)

		if err != nil {
			// Lỗi hạ tầng PDP (timeout, connection refused) — fail-closed
			http.Error(w, fmt.Sprintf("service_unavailable: policy decision point error: %v", err), httpStatus)
			return
		}

		if !allow {
			// PDP trả DENY
			msg := fmt.Sprintf("forbidden: PDP denied access to '%s' for subject '%s'", resource, subject)
			if matchedPolicyID != "" {
				msg = fmt.Sprintf("%s (matched policy: %s)", msg, matchedPolicyID)
			}
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		// PDP trả ALLOW — gắn policy ID vào header để audit/debug
		if matchedPolicyID != "" {
			w.Header().Set("X-Matched-Policy-Id", matchedPolicyID)
		}

		next.ServeHTTP(w, r)
	})
}
