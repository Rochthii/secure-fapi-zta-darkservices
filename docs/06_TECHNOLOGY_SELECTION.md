# PART 6 — TECHNOLOGY SELECTION MATRIX

## 6.1 Identity Provider Selection

| Tiêu chí (Weight) | Keycloak | Authentik | Zitadel | Auth0 | **Custom Go IdP** |
|---|---|---|---|---|---|
| FAPI 2.0 Support (25%) | ⭐⭐⭐ Partial | ⭐⭐ Limited | ⭐⭐⭐ Good | ⭐⭐⭐⭐ Full | ⭐⭐⭐⭐⭐ Full control |
| DPoP Native Support (25%) | ⭐⭐ Plugin | ⭐ None | ⭐⭐ Partial | ⭐⭐⭐⭐ Yes | ⭐⭐⭐⭐⭐ Full control |
| Resource Footprint (15%) | ⭐⭐ Heavy (1GB+) | ⭐⭐⭐ Medium | ⭐⭐⭐ Medium | ⭐⭐⭐⭐ Cloud | ⭐⭐⭐⭐⭐ Minimal |
| Learning Value (20%) | ⭐⭐ Config-heavy | ⭐⭐ Config-heavy | ⭐⭐⭐ Moderate | ⭐ Zero (SaaS) | ⭐⭐⭐⭐⭐ Maximum |
| OpenZiti Integration (15%) | ⭐⭐ Manual | ⭐⭐ Manual | ⭐⭐ Manual | ⭐ Cloud lock-in | ⭐⭐⭐⭐⭐ Native SDK |
| **Weighted Score** | **2.45** | **1.90** | **2.45** | **2.80** | **⭐ 5.00** |

### Quyết định: **Custom Go IdP** ✅

**Lý do chính:**
1. **Kiểm soát 100%** logic DPoP + PKCE — không phụ thuộc plugin/cấu hình bên thứ 3.
2. **Tích hợp OpenZiti SDK native** — IdP có thể chạy như Dark Service.
3. **Giá trị nghiên cứu tối đa** — xây từ zero chứng minh hiểu biết sâu về FAPI 2.0.
4. **Footprint cực nhỏ** — single Go binary, không cần JVM hay Docker image nặng.

**Lý do loại bỏ:**
- Keycloak/Authentik: Quá nặng, cấu hình FAPI phức tạp, không có SDK Ziti.
- Auth0: Cloud-only, không phù hợp lab offline, giá trị nghiên cứu thấp.
- Zitadel: Tốt nhưng DPoP support chưa đầy đủ.

---

## 6.2 API Gateway Selection

| Tiêu chí (Weight) | Kong | Envoy | Traefik | APISIX | **Custom Go Gateway** |
|---|---|---|---|---|---|
| OpenZiti SDK Integration (30%) | ⭐ None | ⭐ None | ⭐ None | ⭐ None | ⭐⭐⭐⭐⭐ Native |
| DPoP Middleware (25%) | ⭐⭐ Plugin | ⭐ Custom filter | ⭐ Plugin | ⭐⭐ Plugin | ⭐⭐⭐⭐⭐ Full control |
| Dark Service Capable (25%) | ❌ No | ❌ No | ❌ No | ❌ No | ✅ Yes |
| Resource Footprint (10%) | ⭐⭐ Heavy | ⭐⭐⭐ Medium | ⭐⭐⭐⭐ Light | ⭐⭐⭐ Medium | ⭐⭐⭐⭐⭐ Minimal |
| Learning Value (10%) | ⭐⭐ Config | ⭐⭐ Config | ⭐⭐ Config | ⭐⭐ Config | ⭐⭐⭐⭐⭐ Maximum |
| **Weighted Score** | **1.35** | **1.20** | **1.25** | **1.35** | **⭐ 5.00** |

### Quyết định: **Custom Go Gateway** ✅

**Lý do quyết định:**
- **Không có gateway thương mại nào hỗ trợ OpenZiti SDK.** Đây là lý do quyết định. Để API Gateway trở thành Dark Service, nó PHẢI nhúng OpenZiti SDK và gọi `ctx.Listen()` thay vì `net.Listen()`. Không có Kong, Envoy, Traefik, hay APISIX nào hỗ trợ điều này.
- Go là ngôn ngữ chính thức của OpenZiti SDK → tương thích tối đa.

---

## 6.3 Overlay Network Selection

| Tiêu chí (Weight) | **OpenZiti** | Tailscale | WireGuard | Istio |
|---|---|---|---|---|
| Dark Service Support (30%) | ⭐⭐⭐⭐⭐ Native | ⭐⭐ Limited | ❌ No | ❌ No |
| SDK Embedding (25%) | ⭐⭐⭐⭐⭐ Go SDK | ⭐⭐ API only | ❌ Kernel | ⭐⭐ Sidecar |
| Identity-first (20%) | ⭐⭐⭐⭐⭐ Yes | ⭐⭐⭐⭐ Yes | ⭐⭐ Key-only | ⭐⭐⭐ SPIFFE |
| Self-hosted (15%) | ⭐⭐⭐⭐⭐ Fully OSS | ⭐⭐⭐ Coordination server | ⭐⭐⭐⭐⭐ Fully OSS | ⭐⭐⭐⭐ OSS |
| Simplicity (10%) | ⭐⭐⭐ Docker Compose | ⭐⭐⭐⭐ Simple | ⭐⭐⭐⭐⭐ Minimal | ⭐⭐ Complex K8s |
| **Weighted Score** | **⭐ 4.75** | **2.95** | **2.00** | **2.00** |

### Quyết định: **OpenZiti** ✅

**Lý do:**
- **Duy nhất** hỗ trợ Dark Service thực sự (zero listening ports) với SDK embedding.
- Tailscale: Không hỗ trợ SDK embedding, phụ thuộc coordination server.
- WireGuard: Chỉ là VPN tunnel, không có identity/policy framework.
- Istio: Yêu cầu Kubernetes, sidecar proxy vẫn mở port nội bộ.

---

## 6.4 Observability Stack Selection

| Tiêu chí (Weight) | Prometheus + Grafana | Datadog | ELK Stack | **Prometheus + Grafana + Loki** |
|---|---|---|---|---|
| Self-hosted (25%) | ⭐⭐⭐⭐⭐ | ❌ Cloud | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Go Integration (25%) | ⭐⭐⭐⭐⭐ Native | ⭐⭐⭐⭐ Agent | ⭐⭐⭐ Beats | ⭐⭐⭐⭐⭐ Native |
| Resource Footprint (20%) | ⭐⭐⭐⭐ Light | ⭐⭐⭐ Agent | ⭐⭐ Heavy | ⭐⭐⭐⭐ Light |
| Log + Metrics + Trace (20%) | ⭐⭐⭐ Metrics only | ⭐⭐⭐⭐⭐ All | ⭐⭐⭐ Logs focus | ⭐⭐⭐⭐ Metrics+Logs |
| Cost (10%) | ⭐⭐⭐⭐⭐ Free | ⭐⭐ Expensive | ⭐⭐⭐⭐ Free | ⭐⭐⭐⭐⭐ Free |
| **Weighted Score** | **4.25** | **2.60** | **3.00** | **⭐ 4.50** |

### Quyết định: **Prometheus + Grafana + Loki** ✅

---

## 6.5 Database Selection

| Tiêu chí | PostgreSQL | MySQL | CockroachDB |
|---|---|---|---|
| RLS Native | ✅ Yes | ❌ No | ✅ Yes |
| Proven in Fintech | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| WORM Triggers | ✅ Full | ⚠️ Limited | ✅ Full |
| Ecosystem | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

### Quyết định: **PostgreSQL 16** ✅

---

## 6.6 Cryptography Selection

| Tiêu chí | ECC P-256 (ES256) | RSA-2048 | EdDSA (Ed25519) |
|---|---|---|---|
| FAPI 2.0 Recommended | ✅ Yes | ✅ Yes | ⚠️ Not in FAPI |
| Key Size | 256-bit (compact) | 2048-bit (large) | 256-bit (compact) |
| Sign/Verify Speed | ⭐⭐⭐⭐⭐ Fast | ⭐⭐⭐ Moderate | ⭐⭐⭐⭐⭐ Fast |
| JWT Library Support | ⭐⭐⭐⭐⭐ Universal | ⭐⭐⭐⭐⭐ Universal | ⭐⭐⭐⭐ Good |
| Browser Web Crypto | ✅ Yes | ✅ Yes | ⚠️ Limited |

### Quyết định: **ECC P-256 (ES256)** ✅

**Lý do:** FAPI 2.0 khuyến nghị, kích thước nhỏ, hiệu năng cao, hỗ trợ rộng rãi.

---

## 6.7 Final Technology Stack

```
┌─────────────────────────────────────────────────────────┐
│                   SELECTED STACK                         │
├─────────────────────────────────────────────────────────┤
│  Language:        Go (Golang)                           │
│  Overlay:         OpenZiti (Controller + Router + SDK)   │
│  Identity:        Custom Go IdP (FAPI 2.0 compliant)    │
│  Gateway:         Custom Go Dark Service                 │
│  Database:        PostgreSQL 16                          │
│  Crypto:          ECC P-256 / ES256                      │
│  Auth Protocol:   OAuth 2.1 + PKCE + DPoP + mTLS        │
│  Observability:   Prometheus + Grafana + Loki            │
│  Infrastructure:  Docker Compose                         │
│  Audit:           WORM Vault (SHA-256 Hash Chain)        │
└─────────────────────────────────────────────────────────┘
```

---

> **Next:** [PART 7 — Target Architecture](./07_TARGET_ARCHITECTURE.md)
