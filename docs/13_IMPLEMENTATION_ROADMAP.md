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

## Phase 6: Failure-Mode Analysis & Security Validation (2026 - 2027)

### Mục tiêu
Kiểm thử toàn diện các lỗ hổng bảo mật tiềm ẩn, tự động hóa kịch bản tấn công và phân tích các trường hợp sai lỗi cấu hình (Failure-modes).

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 6.1 | Attack Scenario Tests | Bộ integration test tự động kiểm thử các lỗi bảo mật cấu hình (như bypass Ziti, spoof client secret). **(COMPLETE)** |
| 6.2 | Failure-Mode Analysis | Nghiên cứu sâu về các điểm sập (fail-open) và xây dựng cơ chế fail-closed chặt chẽ cho hệ thống. **(COMPLETE)** |
| 6.3 | Penetration Test simulation | Giả lập tấn công đánh cắp token, tấn công phát lại (replay), và kiểm soát máy chủ cơ sở dữ liệu. **(COMPLETE)** |
| 6.4 | Security Audit Report | Báo cáo chi tiết các kịch bản lỗi, mã kiểm thử và phương án phòng vệ chủ động. **(COMPLETE)** |

### Success Criteria
- Hệ thống tự động từ chối và cảnh báo chính xác 100% các cuộc tấn công giả lập.
- Xây dựng thành công tài liệu phân tích lỗi hệ thống làm cơ sở cho Chương 4 của đồ án.

---

## Phase 7: Formal Verification & Empirical Benchmarking (2027 - 2028)

### Mục tiêu
Chứng minh tính an toàn toán học của mô hình liên kết chéo (Cross-layer binding) và đo lường định lượng chi phí hiệu năng (Overhead benchmark).

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 7.1 | Protocol Modeling | Khai báo và mô hình hóa giao thức (Ziti + DPoP + mTLS) sử dụng ProVerif tại [protocol.pv](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/security/protocol.pv). **(COMPLETE)** |
| 7.2 | Formal Verification | Kiểm chứng chính thức các thuộc tính an toàn chống lại Cuckoo's Token Attack và Insider Threats. |
| 7.3 | Latency Breakdown | Đo kiểm thời gian xử lý chi tiết của từng lớp (Network, DPoP verify, RLS context switch, Hash-chain) tại [performance_test.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/tests/performance_test.go). **(COMPLETE)** |
| 7.4 | Trade-off Analysis | Báo cáo định lượng về mối quan hệ đánh đổi giữa bảo mật gia cường và độ trễ giao dịch. **(COMPLETE)** |

### Success Criteria
- Hoàn thành mô hình kiểm chứng không chứa lỗi logic giao thức.
- Xuất bản số liệu benchmark p50/p95/p99 trực quan so sánh với baseline.

---

## Phase 8: Academic Writing & PTIT Thesis Defense (2028 - 2029)

### Mục tiêu
Tổng hợp kết quả nghiên cứu, công bố bài báo khoa học và hoàn thiện hồ sơ bảo vệ đồ án tốt nghiệp PTIT.

### Deliverables
| # | Deliverable | Description |
|---|---|---|
| 8.1 | Academic Paper | Viết và gửi bài báo khoa học tham dự các hội thảo uy tín trong nước/quốc tế (FAIR, NICS, v.v.). |
| 8.2 | Complete Thesis | Soạn thảo cuốn đồ án tốt nghiệp PTIT cấu trúc chuẩn khoa học (gồm cả 3 chương nghiên cứu sâu). |
| 8.3 | Defense Materials | Thiết kế slide báo cáo, chuẩn bị kịch bản demo và mô phỏng tấn công thời gian thực. |

### Success Criteria
- Đồ án được hội đồng nghiệm thu đánh giá xuất sắc.
- Có ít nhất một bài báo khoa học được chấp nhận đăng hoặc báo cáo chuyên đề cấp học viện.

---

## Master Timeline (2026 - 2029)

```
2026          ████████████████  Phase 1-5: Hiện thực hóa core system (COMPLETE)
2026 - 2027   ████████████████  Phase 6: Failure-Mode Analysis & Integration Tests
2027 - 2028   ████████████████  Phase 7: Formal Verification & Performance Benchmarking
2028 - 2029   ████████          Phase 8: Công bố khoa học & Bảo vệ đồ án PTIT
```

---

> **Next:** [PART 14 — Project Structure](./14_PROJECT_STRUCTURE.md)

