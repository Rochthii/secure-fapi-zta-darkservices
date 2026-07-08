# QA & Verification Checklist

## 1. Safety Guidelines
- **Disable Go Test Cache**: Always run tests with `-count=1` (e.g., `go test -v -count=1 ./...`).
- **Restart Servers**: Always terminate and restart background servers (IdP/Gateway) after modifying code before running tests.
- **Unused imports/variables**: Always clean up unused imports and variables to avoid Go compiler errors.

## 2. Test Scenarios
- **Valid Flow (Test1)**: OAuth 2.1 PKCE + DPoP token exchange and request succeeds.
- **Client Spoofing (Test2)**: IdP rejects invalid client ID/secret.
- **DPoP Replay (Test3)**: Gateway rejects reused JTI.
- **Ziti Fail-Closed (Test4)**: Gateway rejects direct TCP connection when `ENFORCE_ZITI=true`.
- **Tenant Isolation (Test5)**: Postgres RLS isolates tenant query results.
- **WORM Ledger (Test6)**: Postgres trigger rejects `UPDATE`/`DELETE` on `audit_logs` even for superusers.
- **PEP-PDP Dynamic Auth**: Gateway intercepts request, forwards context to PDP over Ziti overlay, PDP checks AST under 1ms, Gateway enforces (Fail-Closed default).

## 3. Benchmarks
- **Latency Breakdown**:
  ```powershell
  go test -v -count=1 -run=TestLatencyBreakdown ./...
  ```
- **Throughput / Allocations**:
  ```powershell
  go test -bench=BenchmarkEndToEndFlow -benchmem -run=^$ ./...
  ```
- **PEP-PDP Channel Benchmark**:
  ```powershell
  go test -bench=BenchmarkCheckAccess_Local -benchmem -run=^$ ./internal/pdpclient
  ```
