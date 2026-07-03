# PART 14 — PROJECT STRUCTURE

## 14.1 Repository Structure

```
secure-fapi-zta-darkservices/
│
├── docs/                               # Architecture Documentation
│   ├── 00_MASTER_INDEX.md              # Document navigation index
│   ├── 01_EXECUTIVE_SUMMARY.md         # Part 1
│   ├── 02_PROBLEM_STATEMENT.md         # Part 2
│   ├── 03_LITERATURE_REVIEW.md         # Part 3
│   ├── 04_REQUIREMENT_ANALYSIS.md      # Parts 4-5
│   ├── 06_TECHNOLOGY_SELECTION.md      # Part 6
│   ├── 07_TARGET_ARCHITECTURE.md       # Part 7
│   ├── 08_IDENTITY_ACCESS.md           # Part 8
│   ├── 09_DARK_SERVICES.md             # Part 9
│   ├── 10_SECURITY_ARCHITECTURE.md     # Parts 10-12
│   ├── 13_IMPLEMENTATION_ROADMAP.md    # Part 13
│   ├── 14_PROJECT_STRUCTURE.md         # Part 14 (this file)
│   ├── 15_VALIDATION_BENCHMARK.md      # Part 15
│   └── 16_FINAL_MASTER_PLAN.md         # Part 16
│
├── docker/                             # Infrastructure as Code
│   ├── docker-compose.yml              # Full stack: Ziti + PostgreSQL + Monitoring
│   ├── .env                            # Environment variables
│   └── postgres/
│       ├── Dockerfile                  # Custom PostgreSQL image
│       └── init.sql                    # Schema + RLS + WORM triggers
│
├── certs/                              # PKI & Certificate Management
│   ├── scripts/
│   │   ├── generate-ca.sh             # Create Root + Intermediate CA
│   │   ├── generate-server-cert.sh    # Issue server certificates
│   │   └── generate-client-cert.sh    # Issue client certificates
│   ├── ca/                             # CA certificates (committed)
│   │   ├── root-ca.crt                # Root CA public cert
│   │   └── intermediate-ca.crt        # Intermediate CA public cert
│   └── .gitignore                      # Exclude private keys
│
├── idp/                                # Identity Provider (Go Module)
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                         # Entry point
│   ├── config/
│   │   └── config.go                   # IdP configuration
│   ├── handler/
│   │   ├── discovery.go                # OIDC Discovery
│   │   ├── jwks.go                     # JWKS endpoint
│   │   ├── authorize.go                # Authorization endpoint
│   │   └── token.go                    # Token endpoint
│   ├── crypto/
│   │   ├── dpop.go                     # DPoP validation
│   │   ├── dpop_test.go                # DPoP unit tests
│   │   ├── pkce.go                     # PKCE verification
│   │   └── pkce_test.go                # PKCE unit tests
│   └── store/
│       └── memory.go                   # In-memory auth code store
│
├── gateway/                            # API Gateway — Dark Service (Go Module)
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                         # Entry: Ziti listener + router setup
│   ├── internal/
│   │   ├── api/
│   │   │   └── handlers.go             # API Handlers (balance, transfer, audit-logs)
│   │   ├── audit/
│   │   │   └── db.go                   # DB connection + RLS context injection
│   │   ├── auth/
│   │   │   ├── crypto.go               # JWKS cache + DPoP proof verify
│   │   │   └── jti.go                  # JTI anti-replay cache
│   │   ├── middleware/
│   │   │   ├── auth.go                 # SecureAPI + RequireRole middlewares
│   │   │   └── conn.go                 # Ziti identity extraction helper
│   │   └── ziti/
│   │       └── ziti.go                 # OpenZiti Go SDK Listener binding
│   └── build/
│
├── client/                             # Client Application (Go Module)
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                         # Entry: CLI PKCE + DPoP exchange + Ziti requests
│   ├── crypto/
│   │   └── crypto.go                   # Client-side PKCE & DPoP proof generator
│   └── ziti/
│       └── ziti.go                     # Client-side OpenZiti SDK Dial transport

│
├── scripts/                            # Automation Scripts
│   ├── setup-ziti-services.sh          # Create Ziti services + policies
│   ├── enroll-identities.sh            # Enroll all identities
│   ├── seed-data.sql                   # Test data for transactions
│   ├── nmap-scan-test.sh               # Dark Service verification
│   └── run-all-tests.sh                # Full test suite runner
│
├── test/                               # Integration & Security Tests
│   ├── dark_service_test.go            # nmap verification
│   ├── dpop_bypass_test.go             # Token theft simulation
│   ├── mtls_bypass_test.go             # Cert-less connection test
│   ├── rls_isolation_test.go           # Cross-tenant access test
│   ├── worm_tamper_test.go             # Audit chain integrity test
│   └── benchmark_test.go              # Performance benchmark
│
├── results/                            # Test Results & Evidence
│   ├── nmap-scan-output.txt            # Port scan results
│   ├── benchmark-results.json          # Latency measurements
│   └── screenshots/                    # Visual evidence
│
├── README.md                           # Project overview
├── Makefile                            # Build & test automation
├── .gitignore                          # Git ignore rules
└── LICENSE                             # License file
```

---

## 14.2 Documentation Structure (Already Established)

```
docs/
├── 00_MASTER_INDEX.md          ← Navigation hub
├── 01-02                       ← WHY (Problem & Motivation)
├── 03                          ← WHAT (Literature Review)
├── 04-05                       ← WHAT (Requirements & Threats)
├── 06                          ← WITH WHAT (Technology Decisions)
├── 07-09                       ← HOW (Architecture Design)
├── 10-12                       ← HOW SECURE (Security & Compliance)
├── 13                          ← WHEN (Roadmap)
├── 14                          ← WHERE (Structure)
├── 15                          ← HOW TO VERIFY (Validation)
└── 16                          ← SUMMARY (Master Plan)
```

---

## 14.3 Infrastructure Structure

```
Docker Compose Services:
┌─────────────────────────────────────────────────────┐
│ Service              │ Image              │ Ports    │
├──────────────────────┼────────────────────┼──────────┤
│ ziti-controller      │ openziti/ziti-cli   │ Internal │
│ ziti-edge-router     │ openziti/ziti-cli   │ Internal │
│ ziti-console (ZAC)   │ openziti/zac        │ 8443     │
│ postgresql           │ postgres:16-alpine  │ 5432*    │
│ prometheus           │ prom/prometheus     │ 9090*    │
│ grafana              │ grafana/grafana     │ 3000*    │
│ loki                 │ grafana/loki        │ 3100*    │
└──────────────────────┴────────────────────┴──────────┘
* = Internal Docker network only, not exposed to host
  (except ZAC for admin access during development)
```

---

## 14.4 Makefile Targets

| Target | Description |
|---|---|
| `make infra-up` | Start Docker Compose stack |
| `make infra-down` | Stop and clean Docker stack |
| `make setup-ziti` | Run Ziti service/identity/policy setup |
| `make certs` | Generate CA and certificates |
| `make build-idp` | Build IdP binary |
| `make build-gateway` | Build Gateway binary |
| `make build-client` | Build Client binary |
| `make test-unit` | Run unit tests |
| `make test-security` | Run security test suite |
| `make test-dark` | Run nmap Dark Service verification |
| `make benchmark` | Run performance benchmark |
| `make test-all` | Run all tests |
| `make clean` | Clean build artifacts |

---

> **Next:** [PART 15 — Validation & Benchmark](./15_VALIDATION_BENCHMARK.md)
