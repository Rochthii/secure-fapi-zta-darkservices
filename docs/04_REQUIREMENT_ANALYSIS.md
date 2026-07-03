# PART 4 — REQUIREMENT ANALYSIS

## 4.1 Functional Requirements

### FR-01: Zero Trust Authentication
- Hệ thống phải xác thực mọi request bằng DPoP Proof JWT (RFC 9449).
- Access Token phải là sender-constrained (ràng buộc với DPoP keypair của Client).
- Không chấp nhận Bearer token thuần.

### FR-02: Mutual TLS Client Authentication
- Mọi kết nối tới API Gateway phải qua mTLS (RFC 8705).
- Client phải trình diện X.509 certificate đã được CA nội bộ ký.
- Gateway phải xác thực certificate chain và kiểm tra revocation status.

### FR-03: PKCE Authorization Flow
- Authorization Code flow phải sử dụng PKCE (RFC 7636) với method S256.
- Không chấp nhận method `plain`.
- `code_verifier` phải có entropy tối thiểu 256 bits (43 ký tự Base64URL).

### FR-04: Dark Service API Gateway
- API Gateway phải hoạt động như một OpenZiti Dark Service.
- Gateway không được mở bất kỳ cổng TCP/UDP inbound nào trên host.
- Chỉ các enrolled Ziti identity có Dial policy mới có thể kết nối.

### FR-05: Tenant Data Isolation
- Dữ liệu phải được cô lập ở tầng database bằng PostgreSQL Row-Level Security.
- RLS policy phải sử dụng `current_setting('app.tenant_id')` từ JWT claims.
- Middleware phải inject context bằng `SET LOCAL` cho mỗi transaction.

### FR-06: Immutable Audit Logging
- Mọi hoạt động quan trọng phải được ghi vào Audit Ledger.
- Audit records phải tạo thành hash-chain SHA-256 (record N hash phụ thuộc record N-1).
- Cấm UPDATE và DELETE trên bảng audit_logs (WORM semantics).

### FR-07: Token Lifecycle Management
- Access Token TTL: 60 giây (cực ngắn).
- Refresh Token: sender-constrained hoặc one-time-use (rotation).
- Token revocation phải có hiệu lực ngay lập tức.

### FR-08: Identity Provider
- Triển khai custom IdP hỗ trợ FAPI 2.0 profile.
- Cung cấp OIDC Discovery endpoint.
- Cung cấp JWKS endpoint cho public key distribution.

---

## 4.2 Non-Functional Requirements

### NFR-01: Security
| Yêu cầu | Chỉ số |
|---|---|
| Zero network attack surface | `nmap -p 1-65535` → 0 ports open |
| Token theft protection | Stolen token unusable on different device |
| Replay attack protection | JTI uniqueness check, window < 60s |
| E2E encryption | All traffic through Ziti overlay (AES-256-GCM) |
| Certificate validation | Full chain validation + revocation check |

### NFR-02: Scalability
| Yêu cầu | Chỉ số |
|---|---|
| Horizontal scaling | Gateway instances behind Ziti load balancing |
| Database scaling | Connection pooling, read replicas supported |
| Identity scaling | OpenZiti supports 10,000+ identities per controller |

### NFR-03: Availability
| Yêu cầu | Chỉ số |
|---|---|
| Gateway uptime | Target 99.9% (lab environment) |
| IdP uptime | Single instance acceptable for lab |
| Database uptime | PostgreSQL with WAL archiving |

### NFR-04: Reliability
| Yêu cầu | Chỉ số |
|---|---|
| Audit chain integrity | SHA-256 verification passes 100% |
| RLS enforcement | Zero cross-tenant data leaks |
| Token validation | Zero false-positive authentications |

### NFR-05: Maintainability
| Yêu cầu | Chỉ số |
|---|---|
| Modular architecture | Each layer independently replaceable |
| Infrastructure as Code | Docker Compose for full environment |
| Documentation | Complete architecture docs before any code |

### NFR-06: Compliance
| Framework | Requirement |
|---|---|
| NIST SP 800-207 | Full ZTA tenet adherence |
| OWASP API Top 10 | All 10 categories addressed |
| FAPI 2.0 | Sender-constrained tokens, PKCE mandatory |
| PCI DSS | Encryption in transit, access control, audit logging |
| ISO 27001 | A.9 Access Control, A.10 Cryptography, A.12 Operations |

---

# PART 5 — THREAT MODELING

## 5.1 STRIDE Analysis

### Assets Under Protection
| Asset | Classification | Location |
|---|---|---|
| Financial transaction data | **Confidential** | PostgreSQL |
| User credentials | **Secret** | Identity Provider |
| Private keys (DPoP, mTLS) | **Secret** | Client device, Gateway |
| Access/Refresh tokens | **Confidential** | In-transit, client memory |
| Audit logs | **Integrity-Critical** | PostgreSQL (WORM) |
| Ziti Identity files | **Secret** | Enrolled devices |
| CA private key | **Top Secret** | Offline storage |

### STRIDE Threat Matrix

| Threat Category | Target | Threat Description | Risk | Mitigation |
|---|---|---|---|---|
| **S** — Spoofing | Identity | Attacker impersonates legitimate user | HIGH | mTLS certificate + DPoP proof + Ziti Identity |
| **T** — Tampering | Transaction | Modify transfer amount in transit | HIGH | E2E encryption (Ziti) + request signing |
| **R** — Repudiation | Audit | User denies performing transaction | MEDIUM | WORM audit ledger with SHA-256 hash chain |
| **I** — Info Disclosure | Database | Unauthorized data access cross-tenant | HIGH | PostgreSQL RLS + middleware context injection |
| **D** — Denial of Service | Gateway | Flood API with requests | LOW | Dark Service = no public endpoint to flood |
| **E** — Elevation of Privilege | Authorization | User escalates from viewer to admin | HIGH | RBAC + ABAC policy engine + RLS |

---

## 5.2 Attack Tree

```
                    ┌──────────────────────┐
                    │ GOAL: Unauthorized   │
                    │ Financial Transaction│
                    └──────────┬───────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
    ┌─────────▼──────┐  ┌─────▼──────┐  ┌──────▼─────────┐
    │ Steal Valid     │  │ Bypass      │  │ Compromise     │
    │ Credentials    │  │ Auth Layer  │  │ Infrastructure │
    └────────┬───────┘  └─────┬──────┘  └──────┬─────────┘
             │                │                 │
      ┌──────┼──────┐    ┌───┼────┐      ┌─────┼──────┐
      │      │      │    │   │    │      │     │      │
   Token  mTLS  Ziti  PKCE DPoP RLS  Ziti  DB     CA
   Theft  Cert  ID    Bypass Skip  Bypass Ctrl  Access  Comp.
      │      │      │    │   │    │      │     │      │
   BLOCKED BLOCKED BLOCKED ── ALL BLOCKED ── ALL BLOCKED ──
   (DPoP)  (PKI)  (Enroll)     (7-Layer Defense Chain)
```

**Kết luận:** Mọi nhánh tấn công đều bị chặn bởi ít nhất 2 lớp phòng thủ độc lập.

---

## 5.3 Kill Chain Analysis (Cyber Kill Chain — Lockheed Martin)

| Kill Chain Phase | Tấn công truyền thống | FAPI-ZTA Response |
|---|---|---|
| **1. Reconnaissance** | Port scan, API discovery | ❌ BLOCKED — Zero ports, zero endpoints on Internet |
| **2. Weaponization** | Craft exploit for discovered API | ❌ BLOCKED — No API to discover |
| **3. Delivery** | Send exploit to API endpoint | ❌ BLOCKED — No network path without Ziti Identity |
| **4. Exploitation** | Execute exploit on server | ❌ BLOCKED — mTLS + DPoP validation rejects unsigned requests |
| **5. Installation** | Install backdoor | ❌ BLOCKED — No inbound port, no shell access via API |
| **6. C2 (Command & Control)** | Establish outbound C2 channel | ⚠️ MITIGATED — Egress filtering + network monitoring |
| **7. Actions on Objectives** | Steal/modify data | ❌ BLOCKED — RLS isolation + WORM audit detection |

**Kết quả:** Kill Chain bị phá vỡ ngay từ Phase 1 (Reconnaissance). Kẻ tấn công không thể phát hiện mục tiêu.

---

## 5.4 Risk Register

| ID | Risk | Likelihood | Impact | Severity | Mitigation | Residual Risk |
|---|---|---|---|---|---|---|
| R01 | DPoP private key theft from client device | Medium | High | HIGH | Secure key storage (TPM/Keychain) + short token TTL | Low |
| R02 | CA private key compromise | Very Low | Critical | HIGH | Offline CA, HSM recommended for production | Very Low |
| R03 | Ziti Controller compromise | Low | Critical | HIGH | Controller hardening, separate network segment | Low |
| R04 | PostgreSQL superuser escalation | Low | Critical | HIGH | Principle of least privilege, no superuser in app | Low |
| R05 | Insider threat — admin abuse | Medium | High | HIGH | WORM audit, separation of duties, 4-eyes principle | Medium |
| R06 | OpenZiti SDK vulnerability | Low | High | MEDIUM | Pin SDK versions, monitor CVE feeds | Low |
| R07 | Audit chain tampering | Very Low | Critical | HIGH | WORM triggers, external hash anchoring | Very Low |

---

> **Next:** [PART 6 — Technology Selection Matrix](./06_TECHNOLOGY_SELECTION.md)
