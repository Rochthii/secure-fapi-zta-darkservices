# Secure FAPI-ZTA & Dark Services - AI Coding Rules

## 🎯 Project Context
High-security financial transaction system compliant with Zero Trust Architecture (NIST SP 800-207), FAPI 2.0, and OpenZiti overlay.
- **Network Layer**: Pure Go SDK embedding (`ziti.Context.Listen`/`Dial`) — zero open inbound TCP ports.
- **Application Layer**: FAPI 2.0 + DPoP (RFC 9449) + mTLS. Cross-layer binding matches Ziti source ID with token `sub`.
- **Decoupled Auth Plane**: Gateway (PEP) queries Standalone Policy Engine (PDP) via OpenZiti gRPC virtual channel.
- **Data Layer**: Tenant isolation via Postgres RLS (`set_config`). Immutable audit ledger (WORM) via triggers + HMAC-SHA-256.

---

## 🛠️ AI Coding Rules

### 1. Zero-Mock & Vibe Coding Principle
- **ABSOLUTE RULE**: No dummy responses, mock data, fake success UI notifications, or placeholder paths.
- All actions must execute real network, database, or cryptographic operations. Keep existing comments intact.

### 2. Postgres Row-Level Security (RLS) & WORM
- **No App-level Filtering**: Security must never rely on application-level `WHERE tenant_id = ?`. PostgreSQL RLS is mandatory.
- **Context Injection**: Every transaction must invoke `SELECT set_config('app.current_tenant', $1, true)` and `app.audit_secret` before query execution. Clean up context on defer.
- **WORM Integrity**: `audit_logs` must be immutable. Reject all `UPDATE` and `DELETE` queries via database triggers.

### 3. FAPI 2.0, DPoP & OpenZiti
- **Ziti-only Tunnels**: Never dial/listen using OS network sockets; use embedded Ziti Go SDK.
- **Cross-Layer Verification**: Match OpenZiti identity (`SourceIdentifier()`) with access token `sub` claim.
- **DPoP Verification**: Enforce ES256 signatures, validate `htm`/`htu`/`ath` claims, verify `exp` (max 60s), and track `jti` to block replay attacks.

### 4. PEP-PDP gRPC Integration & Codec Rules
- **JSON gRPC Codec**: Both PEP and PDP must register and force the `"json"` codec using `grpc.ForceCodec(JSONCodec())`. Do not use default protobuf binary codec.
- **No JSON omitempty on Enums**: The `decision` field in gRPC check access structures must **never** contain `omitempty` tags. This ensures `DENY (0)` values are serialized.
- **Fail-Closed Mode**: Fail-closed is mandatory. If the PDP times out (80ms threshold) or is unreachable, the Gateway must reject with HTTP 503/403.
- **Enriched ABAC Context**: Forward IP, time, DPoP key thumbprint (`cnf.jkt`), and Ziti identity in gRPC request contexts.

---

## 📂 Key File Map
- **DB Init & WORM:** [init.sql](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docker/postgres/init.sql)
- **Identity Provider:** [idp/main.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/idp/main.go)
- **Gateway Main:** [gateway/main.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/main.go)
- **PDP Client Package:** [pdpclient/client.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/internal/pdpclient/client.go)
- **Security Middleware:** [middleware/auth.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/internal/middleware/auth.go)
