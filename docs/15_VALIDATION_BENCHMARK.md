# PART 15 — VALIDATION & BENCHMARK

## 15.1 Penetration Testing Plan

### Test Category: Dark Service Invisibility (M1)

| Test ID | Test Name | Tool | Command | Expected Result |
|---|---|---|---|---|
| PEN-01 | TCP Full Port Scan | nmap | `nmap -p 1-65535 -sS <gateway-ip>` | All 65535 ports: filtered/closed |
| PEN-02 | UDP Port Scan | nmap | `nmap -p 1-65535 -sU <gateway-ip>` | All ports: filtered/closed |
| PEN-03 | Service Detection | nmap | `nmap -sV -p 1-65535 <gateway-ip>` | No services detected |
| PEN-04 | OS Detection | nmap | `nmap -O <gateway-ip>` | OS detection impossible |
| PEN-05 | Aggressive Scan | nmap | `nmap -A <gateway-ip>` | No information revealed |
| PEN-06 | HTTP Probe | curl | `curl https://<gateway-ip>:443` | Connection refused |
| PEN-07 | DNS Resolution | dig | `dig financial-ledger.api` | NXDOMAIN (no DNS record) |

### Test Category: API Security Testing (M2-M3)

| Test ID | Test Name | Method | Expected Result |
|---|---|---|---|
| API-01 | Valid DPoP request | Full auth flow + DPoP proof | 200 OK |
| API-02 | Bearer token only (no DPoP) | `Authorization: Bearer <token>` | 401 Unauthorized |
| API-03 | Stolen token on different device | Copy token, sign DPoP with different key | 401 — cnf.jkt mismatch |
| API-04 | Replay DPoP proof | Resend same request with same jti | 401 — jti already used |
| API-05 | Expired token | Wait 61s, use expired token | 401 — token expired |
| API-06 | Modified DPoP htm | Sign proof with htm=GET, send POST | 401 — htm mismatch |
| API-07 | Modified DPoP htu | Sign proof for wrong URL | 401 — htu mismatch |
| API-08 | No mTLS cert | Connect without client certificate | Connection refused |
| API-09 | Expired mTLS cert | Use certificate past validity | Connection refused |
| API-10 | Wrong CA cert | Use certificate from untrusted CA | Connection refused |

---

## 15.2 mTLS Validation Test Suite

| Test ID | Scenario | Setup | Expected |
|---|---|---|---|
| MTLS-01 | Valid cert + valid Ziti identity | Enrolled client, valid cert | ✅ Connection established |
| MTLS-02 | Valid cert + no Ziti identity | Valid cert, not enrolled | ❌ No network path |
| MTLS-03 | No cert + valid Ziti identity | Enrolled, no cert | ❌ TLS handshake fails |
| MTLS-04 | Self-signed cert | Cert not from internal CA | ❌ Trust chain validation fails |
| MTLS-05 | Revoked cert | Cert in CRL | ❌ Revocation check fails |

---

## 15.3 DPoP Validation Test Suite

| Test ID | Scenario | DPoP Proof Content | Expected |
|---|---|---|---|
| DPOP-01 | Valid proof | Correct htm, htu, jti, iat, ath | ✅ 200 OK |
| DPOP-02 | Missing DPoP header | No DPoP header at all | ❌ 401 |
| DPOP-03 | Wrong algorithm | `alg: "RS256"` instead of `ES256` | ❌ 401 |
| DPOP-04 | Future iat | `iat` = now + 3600 | ❌ 401 |
| DPOP-05 | Old iat | `iat` = now - 3600 | ❌ 401 |
| DPOP-06 | Duplicate jti | Reuse jti from previous request | ❌ 401 |
| DPOP-07 | Wrong ath | Incorrect access token hash | ❌ 401 |
| DPOP-08 | Different keypair | Sign with key not matching cnf.jkt | ❌ 401 |
| DPOP-09 | Malformed JWT | Invalid JSON in payload | ❌ 401 |
| DPOP-10 | Missing jwk in header | No public key embedded | ❌ 401 |

---

## 15.4 Dark Service Validation

### Verification Protocol

```
Step 1: Start Docker Compose stack
Step 2: Identify Gateway container IP
        $ docker inspect <container> | grep IPAddress
Step 3: Run nmap from outside Docker network
        $ nmap -p 1-65535 -sS <gateway-ip>
Step 4: Record output
        Expected: "All 65535 scanned ports are filtered"
Step 5: Attempt direct HTTP connection
        $ curl -v https://<gateway-ip>:443 --connect-timeout 5
        Expected: "Connection refused" or timeout
Step 6: Connect via Ziti overlay (enrolled client)
        $ ./client dial financial-ledger-service
        Expected: Connection successful, API responds
Step 7: Document contrast
        "External scan: invisible. Ziti connection: functional."
```

---

## 15.5 RLS Isolation Testing

### Test Protocol

```
Step 1: Create 2 tenants (A and B) with seed data
Step 2: Authenticate as Tenant A
Step 3: Query transactions → Should see only Tenant A data
Step 4: Attempt SQL injection to bypass RLS
        → Should still see only Tenant A data
Step 5: Direct SQL query as app_user (bypass API)
        → SET LOCAL 'app.tenant_id' = 'tenant-a-id'
        → SELECT * FROM transactions
        → Should see only Tenant A data
Step 6: Without SET LOCAL → SELECT * FROM transactions
        → Should return 0 rows
```

---

## 15.6 WORM Audit Chain Verification

### Integrity Test Protocol

```
Step 1: Insert 100 audit records via normal API flow
Step 2: Verify chain integrity
        FOR each record N (from 2 to 100):
          computed = SHA256(N.id + N.timestamp + N.actor + N.action 
                         + N.resource + N.details + N.prev_hash)
          ASSERT computed == N.block_hash
Step 3: Attempt UPDATE on record #50
        → Should fail: "Audit logs are immutable"
Step 4: Attempt DELETE on record #50
        → Should fail: "Audit logs are immutable"
Step 5: Manually corrupt record #50 hash (via superuser)
        → Rerun verification
        → Chain should break at record #51 (prev_hash mismatch)
```

---

## 15.7 Performance Benchmark Plan

### Benchmark Scenarios

| Scenario | Description | Metric |
|---|---|---|
| B1 | Traditional API (direct TCP, Bearer token) | Baseline latency |
| B2 | Dark Service API (Ziti overlay, DPoP + mTLS) | Full-stack latency |
| B3 | DPoP proof generation (client-side) | Signing time |
| B4 | DPoP proof validation (server-side) | Validation time |
| B5 | RLS query performance | Query with vs without RLS |
| B6 | WORM audit write | Hash computation + INSERT time |

### Expected Results

| Metric | Traditional API | Dark Service API | Overhead |
|---|---|---|---|
| P50 Latency | ~5ms | ~15-30ms | +10-25ms (Ziti overlay) |
| P99 Latency | ~20ms | ~50-80ms | Acceptable for Fintech |
| DPoP Sign | N/A | ~1-2ms | Negligible |
| DPoP Validate | N/A | ~1-2ms | Negligible |
| Throughput | ~1000 RPS | ~200-500 RPS | Acceptable for lab |

### Benchmark Tooling
- Go `testing.B` benchmark framework.
- Custom load generator for Ziti overlay requests.
- `time` measurements in middleware chain.

---

## 15.8 Attack Simulation Scenarios

| Simulation | Attacker Action | System Response | Verification |
|---|---|---|---|
| **S1: External Recon** | nmap + Shodan scan | Zero results | Screenshot nmap output |
| **S2: Token Theft** | Capture token from logs, use on attacker machine | 401 — DPoP mismatch | Show HTTP response |
| **S3: Replay Attack** | Resend captured request | 401 — jti already used | Show JTI cache hit |
| **S4: Cert Theft** | Use stolen cert without Ziti identity | No network path | Show connection failure |
| **S5: Data Exfil** | Tenant A queries Tenant B data | 0 rows returned | Show query result |
| **S6: Audit Tamper** | Admin tries to delete audit record | Trigger blocks DELETE | Show SQL error |

---

> **Next:** [PART 16 — Final Master Plan](./16_FINAL_MASTER_PLAN.md)
