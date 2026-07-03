# FAPI-ZTA & Dark Services Architecture

> **Financial-grade API Zero Trust Architecture with OpenZiti Dark Services**

## Overview

Enterprise-grade security architecture combining:
- **FAPI 2.0** (Financial-grade API Security Profile)
- **Zero Trust Architecture** (NIST SP 800-207)
- **OpenZiti Dark Services** (Invisible API — Zero Network Attack Surface)
- **DPoP** (RFC 9449 — Demonstrating Proof-of-Possession)
- **mTLS** (RFC 8705 — Mutual TLS Client Authentication)
- **PKCE** (RFC 7636 — Proof Key for Code Exchange)

## Project Status

**Phase 0: Architecture & Design — COMPLETE**

No source code has been written. The entire design must be approved before implementation begins.

## Documentation

The complete architecture documentation set is in [`docs/`](./docs/00_MASTER_INDEX.md):

| Part | Document | Content |
|---|---|---|
| 1 | Executive Summary | Project goals, novel contributions, scalability |
| 2 | Problem Statement | Why JWT+RLS is insufficient, attack vector analysis |
| 3 | Literature Review | 11 standards/technologies reviewed with fitness ratings |
| 4-5 | Requirements & Threats | Functional/Non-functional reqs, STRIDE, Kill Chain |
| 6 | Technology Selection | Weighted scoring matrix for all technology choices |
| 7 | Target Architecture | C4 diagrams, Trust Boundaries, Data Flow |
| 8 | Identity & Access | PKI, mTLS, DPoP, OAuth 2.1, Triple Identity Binding |
| 9 | Dark Services | OpenZiti deep dive, comparisons with VPN/Mesh/Proxy |
| 10-12 | Security & Compliance | Zero Trust policies, NIST/OWASP/PCI DSS/FAPI mapping |
| 13 | Implementation Roadmap | 9 phases with deliverables and quality gates |
| 14 | Project Structure | Repository layout, infrastructure, Makefile targets |
| 15 | Validation & Benchmark | Pen test plan, attack simulations, performance benchmarks |
| 16 | Final Master Plan | Decision log, risk register, development backlog |

## Technology Stack

| Component | Technology | Standard |
|---|---|---|
| Language | Go (Golang) | — |
| Overlay Network | OpenZiti | CSA SDP v3 |
| Identity Provider | Custom Go IdP | FAPI 2.0 |
| API Gateway | Custom Go Dark Service | NIST ZTA |
| Database | PostgreSQL 16 | ISO 27017 |
| Cryptography | ECC P-256 (ES256) | RFC 9449 |
| Infrastructure | Docker Compose | — |
| Monitoring | Prometheus + Grafana + Loki | — |

## Getting Started

> ⚠️ **Implementation has not started.** Review the [Master Index](./docs/00_MASTER_INDEX.md) first.

## License

MIT
