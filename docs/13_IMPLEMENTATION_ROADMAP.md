# PART 13 — IMPLEMENTATION ROADMAP

## Phase 0: Research & Foundation (Week 1)

### Mục tiêu
Hoàn thiện toàn bộ tài liệu thiết kế. Không viết code.

### Deliverables
| # | Deliverable | Status |
|---|---|---|
| 0.1 | Master Documentation Set (16 Parts) | ✅ |
| 0.2 | Technology Selection Matrix with justification | ✅ |
| 0.3 | Threat Model (STRIDE + Kill Chain) | ✅ |
| 0.4 | Compliance Mapping (NIST + OWASP + PCI DSS + FAPI) | ✅ |
| 0.5 | Architecture Diagrams (C4, Trust Boundary, Data Flow) | ✅ |

### Risk
- Scope creep — adding features beyond core architecture.
- Analysis paralysis — over-researching without moving forward.

### Success Criteria
- All 16 documentation parts reviewed and approved.
- No open architectural questions remaining.

---

## Phase 1: Infrastructure Foundation (Week 2)

### Mục tiêu
Xây dựng hạ tầng Docker Compose: OpenZiti cluster + PostgreSQL.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 1.1 | Docker Compose stack | Ziti Controller + Edge Router + ZAC Console |
| 1.2 | PostgreSQL setup | Database with RLS policies + WORM audit triggers |
| 1.3 | Internal PKI | Root CA + Intermediate CA + cert generation scripts |
| 1.4 | Environment configuration | `.env` files, network configuration |

### Risk
- Docker Desktop version incompatibility with OpenZiti.
- OpenZiti container image changes between versions.

### Success Criteria
- `docker compose up` starts all services successfully.
- ZAC Console accessible and shows controller status.
- PostgreSQL running with RLS policies active.
- CA hierarchy generates valid certificates.

---

## Phase 2: Identity & Authentication (Week 3)

### Mục tiêu
Triển khai Custom Go Identity Provider tuân thủ FAPI 2.0.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 2.1 | OIDC Discovery endpoint | `/.well-known/openid-configuration` |
| 2.2 | JWKS endpoint | `/.well-known/jwks.json` with ES256 public keys |
| 2.3 | Authorization endpoint | `/authorize` with PKCE validation |
| 2.4 | Token endpoint | `/token` with DPoP + PKCE verification |
| 2.5 | PKCE module | `code_verifier` / `code_challenge` verification |
| 2.6 | DPoP validation module | Proof parsing, signature verification, JTI tracking |
| 2.7 | Unit tests | Coverage for PKCE + DPoP logic |

### Risk
- ECC key generation/signing edge cases.
- DPoP nonce handling complexity.

### Success Criteria
- IdP issues DPoP-bound access tokens.
- Tokens contain correct `cnf.jkt` claim.
- PKCE validation rejects incorrect `code_verifier`.
- DPoP replay detection works (duplicate JTI rejected).

---

## Phase 3: Network Overlay (Week 3-4)

### Mục tiêu
Cấu hình OpenZiti overlay network với Dark Service topology.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 3.1 | Ziti Service definition | `financial-ledger-service` registered |
| 3.2 | Ziti Identities | Gateway, IdP, Client identities created + enrolled |
| 3.3 | Service Policies | Bind + Dial policies configured |
| 3.4 | Automation script | `setup-ziti-services.sh` for reproducibility |
| 3.5 | Connectivity test | Client → Router → Service verified |

### Risk
- Ziti enrollment JWT expiration if not used promptly.
- Edge Router connectivity issues in Docker networking.

### Success Criteria
- Gateway identity successfully binds to service.
- Client identity successfully dials to service.
- `nmap` scan on gateway container shows 0 open ports.

---

## Phase 4: Dark Services API Gateway (Week 4-5)

### Mục tiêu
Triển khai API Gateway chạy như OpenZiti Dark Service.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 4.1 | Ziti SDK integration | `ctx.Listen("financial-ledger-service")` |
| 4.2 | mTLS middleware | Extract and validate X.509 client cert |
| 4.3 | DPoP middleware | Validate DPoP proof JWT per-request |
| 4.4 | RLS context middleware | `SET LOCAL` tenant/role from JWT claims |
| 4.5 | Transfer handler | Financial transaction API |
| 4.6 | Balance handler | Account balance query API |
| 4.7 | WORM audit writer | Hash-chained audit record insertion |
| 4.8 | DB connection pool | PostgreSQL with RLS context injection |

### Risk
- OpenZiti SDK Go compatibility with latest Go version.
- Performance overhead of middleware chain.

### Success Criteria
- Gateway starts with zero TCP ports open.
- Client can transact through Ziti overlay.
- DPoP-invalid requests are rejected.
- Audit chain integrity verified.

---

## Phase 5: Client Application (Week 5)

### Mục tiêu
Triển khai Client Go application với full security flow.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 5.1 | PKCE flow | Generate `code_verifier` + `code_challenge` |
| 5.2 | DPoP key management | Generate ECC P-256 keypair, sign proofs |
| 5.3 | Ziti Dialer | Connect to Dark Service via overlay |
| 5.4 | E2E integration | Auth → Connect → Transact → Verify |

### Risk
- Key storage security on client device.
- Token refresh timing in short-TTL environment.

### Success Criteria
- Client completes full OAuth 2.1 + PKCE + DPoP flow.
- Client connects via Ziti overlay (not direct TCP).
- Transaction succeeds with proper authentication.

---

## Phase 6: Observability (Week 5-6)

### Mục tiêu
Triển khai monitoring và logging.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 6.1 | Prometheus metrics | Go `/metrics` endpoint |
| 6.2 | Grafana dashboards | Security metrics visualization |
| 6.3 | Loki log aggregation | Structured JSON log collection |
| 6.4 | Alert rules | DPoP failure rate, auth anomalies |

### Risk
- Prometheus scraping through Ziti overlay complexity.

### Success Criteria
- Grafana dashboard shows real-time security metrics.
- Loki aggregates logs from all services.
- Alerts fire on simulated attack scenarios.

---

## Phase 7: Validation & Testing (Week 6)

### Mục tiêu
Kiểm thử toàn diện mọi lớp phòng thủ.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 7.1 | Dark Service scan test | `nmap` verification |
| 7.2 | DPoP bypass attempts | Token theft simulation |
| 7.3 | mTLS bypass attempts | Connection without cert |
| 7.4 | RLS cross-tenant test | Data isolation verification |
| 7.5 | WORM tamper test | Audit chain integrity check |
| 7.6 | Performance benchmark | Latency comparison |
| 7.7 | Test report | Comprehensive results document |

### Success Criteria
- All 5 must-have objectives (M1-M5) validated with evidence.
- No critical or high-severity findings.

---

## Phase 8: Documentation & Finalization (Week 6-7)

### Mục tiêu
Hoàn thiện tài liệu, chuẩn bị báo cáo.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 8.1 | Final architecture docs | Updated with implementation learnings |
| 8.2 | Test results | Screenshots, nmap output, benchmark charts |
| 8.3 | README | Complete project setup guide |
| 8.4 | Git repository | Clean commit history, tagged releases |

---

## Master Timeline

```
Week 1  ████████████████  Phase 0: Research & Docs (COMPLETE)
Week 2  ████████████████  Phase 1: Infrastructure
Week 3  ████████████████  Phase 2: Identity + Phase 3: Network
Week 4  ████████████████  Phase 3: Network + Phase 4: Gateway
Week 5  ████████████████  Phase 4: Gateway + Phase 5: Client
Week 6  ████████████████  Phase 6: Observability + Phase 7: Testing
Week 7  ████████          Phase 8: Finalization
```

---

> **Next:** [PART 14 — Project Structure](./14_PROJECT_STRUCTURE.md)
