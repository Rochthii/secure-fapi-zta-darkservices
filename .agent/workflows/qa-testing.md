---
description: Kiểm thử toàn diện trước khi release — QA checklist
---

Thực hiện QA checklist đầy đủ cho dự án Secure FAPI-ZTA & Dark Services. Báo cáo kết quả rõ ràng bằng tiếng Việt.

### 1. Quy Trình Kiểm Thử An Toàn (Go & Server Guidelines)
Để tránh lỗi và kết quả kiểm thử sai lệch, bắt buộc tuân thủ:
1. **Khắc phục Cache của Go Test:** Go tự động cache kết quả test nếu không phát hiện thay đổi trong mã kiểm thử. Bắt buộc phải sử dụng flag `-count=1` (ví dụ: `go test -v -count=1 ./...`) để chạy thực tế.
2. **Chu kỳ sống của Server Task:** Khi thay đổi mã nguồn của IdP hoặc Gateway, phải chấm dứt (kill) tiến trình máy chủ cũ đang chạy ngầm và khởi động lại tiến trình mới để áp dụng các thay đổi trước khi chạy bộ test.
3. **Tránh lỗi Biên dịch Go:** Trình biên dịch Go cấm tuyệt đối các thư viện import hoặc biến được khai báo nhưng không sử dụng. Luôn chạy `go build` kiểm tra trước khi chạy test.
4. **Cấu hình môi trường mạng Ziti:** Khi chạy test fail-closed, thiết lập `ENFORCE_ZITI=true` để giả lập chặn kết nối thô không đi qua overlay OpenZiti.

---

### 2. Kịch Bản Kiểm Thử & Xác Minh (Testing Scenarios)

#### A. Kiểm thử chức năng & Bảo mật
1. **Valid Flow (Test1)**: Chạy OAuth 2.1 PKCE + DPoP token exchange và gửi request hợp lệ lên Gateway thành công.
2. **Client Spoofing (Test2)**: Xác thực IdP chặn đứng client ID lạ hoặc sai Client Secret.
3. **DPoP Replay (Test3)**: Xác thực Gateway chặn đứng các cuộc tấn công phát lại DPoP Proof (JTI reuse).
4. **Ziti Fail-Closed (Test4)**: Xác thực Gateway từ chối kết nối trực tiếp qua TCP thông thường khi bật `ENFORCE_ZITI=true`.
5. **Tenant Isolation (Test5)**: Xác thực Postgres RLS cô lập 100% dữ liệu và logs giữa các tenants (ví dụ: Alice vs Bob).
6. **WORM Ledger Immutability (Test6)**: Xác thực database block 100% hành vi `UPDATE`/`DELETE` trên bảng `audit_logs` kể cả bởi superuser.

#### B. Kiểm thử hiệu năng & Độ trễ (Benchmarking)
- Chạy đo đạc Latency Breakdown:
  ```powershell
  go test -v -count=1 -run=TestLatencyBreakdown ./...
  ```
- Chạy đo đạc Throughput / Allocations:
  ```powershell
  go test -bench=BenchmarkEndToEndFlow -benchmem -run=^$ ./...
  ```

---

### 3. Báo cáo kết quả
- ✅ PASS: Mô tả chi tiết thời gian và kết quả.
- ❌ FAIL: Nguyên nhân và nhật ký lỗi.
