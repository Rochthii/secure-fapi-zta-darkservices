# FAPI-ZTA & Dark Services Architecture
## Master Documentation Index

> **Project:** secure-fapi-zta-darkservices
> **Classification:** Enterprise Architecture Design Document
> **Status:** Architecture Phase — NO IMPLEMENTATION
> **Created:** 2026-07-03
> **Author Role:** Principal Security Architect / Zero Trust Architect

---

## Document Set

| # | Document | Status |
|---|---|---|
| **PART 1** | [Executive Summary](./01_EXECUTIVE_SUMMARY.md) | ✅ |
| **PART 2** | [Problem Statement](./02_PROBLEM_STATEMENT.md) | ✅ |
| **PART 3** | [Literature Review](./03_LITERATURE_REVIEW.md) | ✅ |
| **PART 4** | [Requirement Analysis](./04_REQUIREMENT_ANALYSIS.md) | ✅ |
| **PART 5** | [Threat Modeling](./05_THREAT_MODELING.md) | ✅ |
| **PART 6** | [Technology Selection Matrix](./06_TECHNOLOGY_SELECTION.md) | ✅ |
| **PART 7** | [Target Architecture](./07_TARGET_ARCHITECTURE.md) | ✅ |
| **PART 8** | [Identity & Access Architecture](./08_IDENTITY_ACCESS.md) | ✅ |
| **PART 9** | [Dark Services Architecture](./09_DARK_SERVICES.md) | ✅ |
| **PART 10** | [Security Architecture](./10_SECURITY_ARCHITECTURE.md) | ✅ |
| **PART 11** | [Observability Architecture](./11_OBSERVABILITY.md) | ✅ |
| **PART 12** | [Compliance Mapping](./12_COMPLIANCE_MAPPING.md) | ✅ |
| **PART 13** | [Implementation Roadmap](./13_IMPLEMENTATION_ROADMAP.md) | ✅ |
| **PART 14** | [Project Structure](./14_PROJECT_STRUCTURE.md) | ✅ |
| **PART 15** | [Validation & Benchmark](./15_VALIDATION_BENCHMARK.md) | ✅ |
| **PART 16** | [Final Master Plan](./16_FINAL_MASTER_PLAN.md) | ✅ |

---

## Governing Standards

| Standard | ID | Role |
|---|---|---|
| Zero Trust Architecture | NIST SP 800-207 | Architectural Framework |
| Financial-grade API 2.0 | OpenID FAPI 2.0 Final | API Security Profile |
| DPoP | IETF RFC 9449 | Token Sender-Constraining |
| mTLS | IETF RFC 8705 | Transport-layer Client Auth |
| PKCE | IETF RFC 7636 | Authorization Code Protection |
| OAuth 2.1 | draft-ietf-oauth-v2-1 | Authorization Framework |
| SDP | CSA SDP v3 | Dark Network Specification |
| API Security | OWASP API Top 10 2023 | Threat Baseline |

---

> **RULE:** No source code shall be written until ALL 16 documents are reviewed and approved.
