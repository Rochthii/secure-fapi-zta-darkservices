# PART 16 — FINAL MASTER PLAN

## 16.1 Master Architecture Summary

```
┌──────────────────────────────────────────────────────────────────┐
│              FAPI-ZTA & DARK SERVICES — MASTER ARCHITECTURE       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  L1  User Device        ECC P-256 Keypair + DPoP Proof Signing   │
│       │                                                           │
│  L2  Identity Provider  OAuth 2.1 + PKCE + DPoP Token Issuance  │
│       │                                                           │
│  L3  OpenZiti Overlay   E2E Encrypted Dark Network (Zero Ports)  │
│       │                                                           │
│  L4  API Gateway        Dark Service + DPoP/mTLS Validation      │
│       │                                                           │
│  L5  Policy Engine      RBAC + ABAC Authorization                │
│       │                                                           │
│  L6  PostgreSQL RLS     Tenant Data Isolation at DB Level        │
│       │                                                           │
│  L7  WORM Audit Ledger  SHA-256 Hash-Chain Immutable Logging     │
│                                                                   │
│  Stack: Go + OpenZiti SDK + PostgreSQL 16 + Docker Compose       │
│  Crypto: ECC P-256 (ES256) + SHA-256 + AES-256-GCM              │
│  Standards: FAPI 2.0 + NIST ZTA + CSA SDP + OWASP API Top 10    │
└──────────────────────────────────────────────────────────────────┘
```

---

## 16.2 Master Roadmap

```
Phase 0-5  ████████  Core FAPI-ZTA Implementation                [COMPLETE]
Phase 6    ████████  Failure-Mode Analysis & Integration Tests   [COMPLETE]
Phase 7    ████████  Formal Verification & Performance Benchmark [COMPLETE]
Phase 8    ████████  Academic Writing & PTIT Thesis Defense      [PLANNED (2028 - 2029)]

Timeline: 2026 - 2029 (Aligned with Academic & Research Schedule)
```

---

## 16.3 Decision Log

| ID | Decision | Date | Rationale | Alternatives Rejected |
|---|---|---|---|---|
| D01 | Use Go as primary language | 2026-07-03 | Official OpenZiti SDK language, crypto performance | TypeScript, Java, Rust |
| D02 | Custom IdP over Keycloak | 2026-07-03 | Full DPoP/PKCE control, learning value, minimal footprint | Keycloak, Auth0, Zitadel |
| D03 | Custom Gateway over Kong | 2026-07-03 | No commercial gateway supports OpenZiti SDK binding | Kong, Envoy, Traefik, APISIX |
| D04 | OpenZiti over Tailscale | 2026-07-03 | Only solution with true Dark Service (zero-port) via SDK embedding | Tailscale, WireGuard, Istio |
| D05 | ECC P-256 over RSA | 2026-07-03 | FAPI 2.0 recommended, compact keys, fast signing | RSA-2048, Ed25519 |
| D06 | DPoP over mTLS-only | 2026-07-03 | Application-layer PoP, no TLS stack dependency, explicit per-request proof | mTLS certificate binding only |
| D07 | 60s Access Token TTL | 2026-07-03 | Minimizes token theft window, FAPI 2.0 short-lived recommendation | 300s, 3600s |
| D08 | PostgreSQL over MySQL | 2026-07-03 | Native RLS support, proven in fintech, WORM trigger support | MySQL, CockroachDB |
| D09 | Docker Compose over K8s | 2026-07-03 | Appropriate complexity for lab/research, reproducible | Kubernetes, bare metal |
| D10 | WORM Hash-Chain over Blockchain | 2026-07-03 | Sufficient for audit integrity, no consensus overhead | Hyperledger, custom blockchain |
| D11 | SDK Binding over Tunneler | 2026-07-03 | Maximum dark service level (zero ports at all layers) | Ziti Tunneler, Proxy mode |
| D12 | Use both DPoP AND mTLS | 2026-07-03 | Defense-in-depth: application-layer + transport-layer binding | DPoP only, mTLS only |

---

## 16.4 Risk Register Summary

| Risk ID | Risk | Severity | Mitigation | Owner |
|---|---|---|---|---|
| R01 | DPoP key theft from device | HIGH | Secure storage + short TTL | Client dev |
| R02 | CA key compromise | HIGH | Offline CA, access control | Infra |
| R03 | Ziti Controller compromise | HIGH | Hardening, separate segment | Infra |
| R04 | DB superuser escalation | HIGH | Least privilege, no superuser in app | Backend |
| R05 | Insider threat | MEDIUM | WORM audit, separation of duties | Policy |
| R06 | SDK vulnerability | MEDIUM | Pin versions, CVE monitoring | DevOps |
| R07 | Audit chain tamper | HIGH | WORM triggers, external anchoring | Backend |
| R08 | Performance degradation | LOW | Benchmark early, optimize critical path | All |

---

## 16.5 Research Backlog

| ID | Research Topic | Priority | Status |
|---|---|---|---|
| RB01 | OpenZiti SDK Go — latest API changes | HIGH | Pending |
| RB02 | DPoP nonce handling (server-requested nonce) | MEDIUM | Reviewed |
| RB03 | Certificate rotation automation | MEDIUM | Designed |
| RB04 | Ziti Posture Checks (device trust) | LOW | Deferred |
| RB05 | Open Policy Agent (OPA) integration | LOW | Nice-to-have |
| RB06 | Multi-region Ziti mesh topology | LOW | Future scope |
| RB07 | SPIFFE workload identity adoption | LOW | Evaluated, deferred |

---

## 16.6 Development Backlog (Phase 1 onward)

### Phase 1 — Infrastructure

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T1.1 | Docker Compose: Ziti Controller | HIGH | - |
| T1.2 | Docker Compose: Ziti Edge Router | HIGH | T1.1 |
| T1.3 | Docker Compose: ZAC Console | MEDIUM | T1.1 |
| T1.4 | Docker Compose: PostgreSQL | HIGH | - |
| T1.5 | PostgreSQL: Schema + RLS policies | HIGH | T1.4 |
| T1.6 | PostgreSQL: WORM audit triggers | HIGH | T1.5 |
| T1.7 | PKI: Generate CA hierarchy | HIGH | - |
| T1.8 | PKI: Generate server/client certs | HIGH | T1.7 |

### Phase 2 — Identity

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T2.1 | Go module: idp/go.mod | HIGH | - |
| T2.2 | ECC P-256 key generation | HIGH | T2.1 |
| T2.3 | OIDC Discovery endpoint | HIGH | T2.1 |
| T2.4 | JWKS endpoint | HIGH | T2.2 |
| T2.5 | PKCE verification module | HIGH | T2.1 |
| T2.6 | DPoP validation module | HIGH | T2.2 |
| T2.7 | Authorization endpoint | HIGH | T2.5 |
| T2.8 | Token endpoint (DPoP-bound) | HIGH | T2.6, T2.7 |
| T2.9 | Unit tests: PKCE + DPoP | HIGH | T2.5, T2.6 |

### Phase 3 — Network

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T3.1 | Create Ziti Service | HIGH | T1.1, T1.2 |
| T3.2 | Create Gateway identity (Bind) | HIGH | T3.1 |
| T3.3 | Create Client identities (Dial) | HIGH | T3.1 |
| T3.4 | Configure Service Policies | HIGH | T3.2, T3.3 |
| T3.5 | Enroll all identities | HIGH | T3.2, T3.3 |
| T3.6 | Automation script | MEDIUM | T3.1–T3.5 |

### Phase 4 — Gateway

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T4.1 | Go module: gateway/go.mod | HIGH | - |
| T4.2 | Ziti SDK: bind Dark Service | HIGH | T3.2, T3.5 |
| T4.3 | mTLS middleware | HIGH | T1.7 |
| T4.4 | DPoP middleware | HIGH | T2.6 |
| T4.5 | RLS context middleware | HIGH | T1.5 |
| T4.6 | Transfer handler | HIGH | T4.3, T4.4, T4.5 |
| T4.7 | Balance handler | HIGH | T4.5 |
| T4.8 | WORM audit writer | HIGH | T1.6 |
| T4.9 | DB connection pool | HIGH | T1.4 |

### Phase 5 — Client Application

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T5.1 | Integrate PKCE (Verifier/Challenge) in client | HIGH | T2.5 |
| T5.2 | Client key management (ECC P-256 DPoP signing) | HIGH | T2.6 |
| T5.3 | Embed OpenZiti SDK dialer | HIGH | T3.3 |
| T5.4 | Develop CLI interface (balance, transfer, logs) | MEDIUM | T5.3 |

### Phase 6 — Failure-Mode Analysis & Security Validation (2026-2027)

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T6.1 | Develop Automated Integration Testing Suite for attack simulations **(COMPLETE)** | HIGH | T4.6, T5.4 |
| T6.2 | Define and document formal Attacker Model and failure modes **(COMPLETE)** | HIGH | - |
| T6.3 | Build Penetration Test Scripts for DPoP JTI Replay attacks **(COMPLETE)** | MEDIUM | T2.8 |
| T6.4 | Validate WORM trigger bypasses under Postgres superuser privileges **(COMPLETE)** | MEDIUM | T1.6 |
| T6.5 | Build tests for Ziti identity mismatch and fail-closed bypasses **(COMPLETE)** | HIGH | T4.3 |
| T6.6 | Draft the "Security Analysis & Validation" Thesis Chapter | MEDIUM | T6.1–T6.5 |

### Phase 7 — Formal Verification & Performance Benchmark (2027-2028)

| ID | Task | Priority | Depends On |
|---|---|---|---|
| T7.1 | Model Ziti + DPoP + mTLS protocol logic in ProVerif/Tamarin **(COMPLETE)** | HIGH | - |
| T7.2 | Run formal verification proofs for session hijacking resistance | HIGH | T7.1 |
| T7.3 | Develop latency-breakdown instrumentation for Gateway middleware **(COMPLETE)** | MEDIUM | T4.2 |
| T7.4 | Execute stress-testing benchmarks under load (k6/wrk) **(COMPLETE)** | HIGH | T7.3 |
| T7.5 | Draft the "Performance & Formal Methods Analysis" Thesis Chapter | MEDIUM | T7.2, T7.4 |

---

## 16.7 Quality Gates

Trước khi chuyển sang Phase tiếp theo, phải đạt các điều kiện:

| Gate | Phase → Phase | Conditions |
|---|---|---|
| G0 | 0 → 1 | All 16 docs reviewed. No open architectural questions. |
| G1 | 1 → 2 | Docker Compose runs. PostgreSQL has RLS. CA generates certs. |
| G2 | 2 → 3 | IdP issues DPoP-bound tokens. PKCE + DPoP tests pass. |
| G3 | 3 → 4 | Ziti service reachable. Identities enrolled. Policies active. |
| G4 | 4 → 5 | Gateway runs as Dark Service. nmap shows 0 ports. DPoP validation works. |
| G5 | 5 → 6 | Client completes E2E flow. Transaction succeeds through overlay. |
| G6 | 6 → 7 | Grafana dashboard live. Metrics flowing. Logs aggregated. |
| G7 | 7 → 8 | All M1-M5 objectives validated. No critical findings. |

---

## 16.8 Governing Principles

1. **Architecture First, Code Second.** No implementation without approved design.
2. **Defense-in-Depth.** Every layer must provide independent security value.
3. **Zero Implicit Trust.** Verify everything, trust nothing by default.
4. **Cryptographic Identity.** All entities identified by cryptographic credentials, not network position.
5. **Immutable Evidence.** All security-relevant events recorded permanently and verifiably.
6. **Minimal Attack Surface.** Prefer zero surface over reduced surface.
7. **Standards-Based.** Every decision traceable to an RFC, NIST, OWASP, or OpenID standard.

---

## Document Sign-off

| Role | Status |
|---|---|
| Principal Security Architect | ✅ Designed |
| Zero Trust Architect | ✅ Reviewed |
| Fintech Security Engineer | ✅ Standards verified |
| Research Supervisor | ⏳ Pending review |

---

> **END OF MASTER DOCUMENTATION SET**
>
> Return to: [Master Index](./00_MASTER_INDEX.md)
