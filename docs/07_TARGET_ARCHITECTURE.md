# PART 7 — TARGET ARCHITECTURE

## 7.1 Context Diagram (C4 Level 1)

Tổng quan hệ thống từ góc nhìn bên ngoài: ai tương tác với hệ thống, qua đường nào.

```mermaid
C4Context
    title System Context Diagram — FAPI-ZTA Dark Services

    Person(user, "Financial User", "End user performing transactions via enrolled device")
    Person(admin, "System Admin", "Manages identities, policies, monitors system")

    System(fapi_system, "FAPI-ZTA Platform", "Financial-grade API with Dark Service invisibility")

    System_Ext(ziti_network, "OpenZiti Overlay", "Zero Trust overlay network — no public endpoints")
    System_Ext(pki, "Internal PKI", "Certificate Authority for mTLS certificates")

    Rel(user, ziti_network, "Connects via enrolled Ziti identity")
    Rel(ziti_network, fapi_system, "Routes traffic through encrypted overlay")
    Rel(admin, fapi_system, "Manages via Ziti Admin Console (ZAC)")
    Rel(fapi_system, pki, "Validates client certificates")
```

---

## 7.2 Container Diagram (C4 Level 2)

Các container (process/service) bên trong hệ thống và mối quan hệ.

```mermaid
C4Container
    title Container Diagram — FAPI-ZTA Dark Services

    Person(client, "Client App", "Go CLI with embedded Ziti SDK + DPoP")

    System_Boundary(platform, "FAPI-ZTA Platform") {
        Container(idp, "Identity Provider", "Go", "OAuth 2.1 + PKCE + DPoP token issuance")
        Container(gateway, "API Gateway", "Go + Ziti SDK", "Dark Service — zero inbound ports")
        Container(postgres, "PostgreSQL 16", "Database", "RLS tenant isolation + WORM audit")
        Container(ziti_ctrl, "Ziti Controller", "OpenZiti", "Identity, service, policy management")
        Container(ziti_router, "Ziti Edge Router", "OpenZiti", "Data plane — traffic relay")
    }

    Rel(client, ziti_router, "Dial via Ziti overlay", "mTLS + Ziti Identity")
    Rel(ziti_router, gateway, "Route to Dark Service", "Ziti fabric")
    Rel(client, idp, "OAuth 2.1 + PKCE + DPoP", "HTTPS or Ziti")
    Rel(gateway, postgres, "SQL + SET LOCAL", "mTLS")
    Rel(gateway, idp, "Validate tokens via JWKS", "Internal")
    Rel(ziti_router, ziti_ctrl, "Control plane", "mTLS")
```

---

## 7.3 Component Diagram (C4 Level 3) — API Gateway

```mermaid
C4Component
    title Component Diagram — API Gateway (Dark Service)

    Container_Boundary(gw, "API Gateway") {
        Component(ziti_bind, "Ziti Service Binder", "Binds to financial-ledger-service on overlay")
        Component(mw_mtls, "mTLS Middleware", "Validates X.509 client certificate from Ziti conn")
        Component(mw_dpop, "DPoP Middleware", "Validates DPoP proof JWT, checks ath, jti cache")
        Component(mw_rls, "RLS Context Middleware", "Extracts JWT claims, SET LOCAL tenant_id")
        Component(handler_tx, "Transfer Handler", "Processes financial transactions")
        Component(handler_bal, "Balance Handler", "Queries account balance")
        Component(audit_worm, "WORM Audit Writer", "Writes hash-chained audit records")
        Component(db_pool, "DB Connection Pool", "PostgreSQL connection with RLS context")
    }

    Rel(ziti_bind, mw_mtls, "Incoming Ziti connection")
    Rel(mw_mtls, mw_dpop, "Cert-validated request")
    Rel(mw_dpop, mw_rls, "DPoP-validated request")
    Rel(mw_rls, handler_tx, "Context-injected request")
    Rel(mw_rls, handler_bal, "Context-injected request")
    Rel(handler_tx, db_pool, "SQL query")
    Rel(handler_tx, audit_worm, "Log transaction")
```

---

## 7.4 Trust Boundary Diagram

```
┌─ TRUST ZONE 0: UNTRUSTED ──────────────────────────────────────┐
│                                                                  │
│   [Internet]  ──X──  NO PATH TO ANY SERVICE                     │
│   [Attacker]  ──X──  Zero ports, zero endpoints                 │
│                                                                  │
├─ TRUST ZONE 1: ZITI OVERLAY (AUTHENTICATED) ───────────────────┤
│                                                                  │
│   ┌──────────┐      ┌──────────────┐      ┌──────────────┐     │
│   │ Enrolled │─mTLS─│ Ziti Edge    │─mTLS─│ Ziti         │     │
│   │ Client   │      │ Router       │      │ Controller   │     │
│   └──────────┘      └──────────────┘      └──────────────┘     │
│                                                                  │
├─ TRUST ZONE 2: APPLICATION (VERIFIED) ─────────────────────────┤
│                                                                  │
│   ┌──────────┐      ┌──────────────┐                            │
│   │ Identity │─────→│ API Gateway  │  ← Dark Service            │
│   │ Provider │      │ (Go+Ziti)    │  ← DPoP + mTLS validated   │
│   └──────────┘      └──────┬───────┘                            │
│                             │                                    │
├─ TRUST ZONE 3: DATA (ENFORCED) ────────────────────────────────┤
│                             │                                    │
│                      ┌──────▼───────┐                            │
│                      │ PostgreSQL   │  ← RLS enforced            │
│                      │ + WORM Audit │  ← Immutable logs          │
│                      └──────────────┘                            │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 7.5 Data Flow Diagram — Complete Transaction Flow

```mermaid
sequenceDiagram
    participant C as Client App
    participant ZR as Ziti Router
    participant IdP as Identity Provider
    participant GW as API Gateway (Dark)
    participant DB as PostgreSQL

    Note over C: Phase 1: Authentication
    C->>C: Generate ECC P-256 keypair (DPoP)
    C->>C: Generate PKCE code_verifier + code_challenge
    C->>IdP: GET /authorize?code_challenge=...&response_type=code
    IdP-->>C: authorization_code

    C->>C: Sign DPoP Proof JWT (htm=POST, htu=/token)
    C->>IdP: POST /token + code_verifier + DPoP header
    IdP->>IdP: Verify PKCE: SHA256(verifier) == challenge
    IdP->>IdP: Verify DPoP: signature, jti, iat
    IdP->>IdP: Issue Access Token with cnf.jkt = DPoP thumbprint
    IdP-->>C: access_token (60s TTL) + refresh_token

    Note over C: Phase 2: Connect via Dark Network
    C->>ZR: Dial "financial-ledger-service" via Ziti Identity
    ZR->>ZR: Authenticate Ziti Identity (mTLS)
    ZR->>GW: Route connection through overlay fabric

    Note over GW: Phase 3: Request Validation (7-Layer)
    C->>GW: POST /api/transfer + Authorization: DPoP <token> + DPoP: <proof>
    GW->>GW: L3: Verify Ziti connection identity
    GW->>GW: L4a: Verify mTLS client certificate
    GW->>GW: L4b: Verify DPoP proof (signature, htm, htu, jti, ath)
    GW->>GW: L4c: Verify Access Token (signature, exp, cnf.jkt)
    GW->>GW: L5: Check RBAC/ABAC policies

    Note over DB: Phase 4: Data Access with RLS
    GW->>DB: SET LOCAL 'app.tenant_id' = '<from_jwt>'
    GW->>DB: SET LOCAL 'app.role' = '<from_jwt>'
    GW->>DB: INSERT INTO transactions (tenant_id, amount, ...)
    DB->>DB: L6: RLS policy check (tenant_id matches)
    DB->>DB: L7: Trigger → INSERT audit_logs (hash-chained)
    DB-->>GW: Transaction result

    GW-->>C: 200 OK + transaction receipt
```

---

## 7.6 Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose Stack                       │
│                                                              │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │ ziti-controller   │  │ ziti-edge-router  │                │
│  │ :1280 (ctrl)      │  │ :3022 (edge)      │                │
│  │ :6262 (mgmt)      │  │                   │                │
│  └────────┬─────────┘  └────────┬──────────┘                │
│           │                      │                           │
│  ┌────────▼─────────┐           │                           │
│  │ ziti-console      │           │                           │
│  │ (ZAC) :8443       │           │                           │
│  └──────────────────┘           │                           │
│                                  │                           │
│  ┌──────────────────┐  ┌────────▼──────────┐                │
│  │ idp-server        │  │ api-gateway       │                │
│  │ (Go IdP)          │  │ (Go Dark Service) │                │
│  │ Ziti Dark Service │  │ Ziti Dark Service │                │
│  │ NO PUBLIC PORT    │  │ NO PUBLIC PORT    │                │
│  └──────────────────┘  └────────┬──────────┘                │
│                                  │                           │
│                         ┌────────▼──────────┐                │
│                         │ postgresql         │                │
│                         │ :5432 (internal)   │                │
│                         │ RLS + WORM Audit   │                │
│                         └───────────────────┘                │
│                                                              │
│  ┌──────────────────┐  ┌──────────────────┐                 │
│  │ prometheus        │  │ grafana           │                │
│  │ :9090             │  │ :3000             │                │
│  └──────────────────┘  └──────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

---

> **Next:** [PART 8 — Identity & Access Architecture](./08_IDENTITY_ACCESS.md)
