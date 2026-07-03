# ADR-004: Sử dụng PostgreSQL Row-Level Security & Trigger WORM

*   **Status:** APPROVED
*   **Date:** 2026-07-03
*   **Deciders:** Principal Security Architect / Zero Trust Architect

---

## 1. Bối cảnh (Context)
Cô lập dữ liệu đa khách thuê (Multi-tenant Isolation) và ghi vết nhật ký bất biến (Immutable Auditing) là hai yêu cầu bắt buộc của hệ thống Fintech ZTA:
- Nếu cô lập dữ liệu chỉ thực hiện ở tầng ứng dụng (Application Layer), một lỗi lập trình nhỏ (quên thêm mệnh đề `WHERE tenant_id = ...`) sẽ dẫn đến rò rỉ dữ liệu chéo diện rộng.
- Nếu nhật ký kiểm toán (Audit logs) có thể bị sửa đổi bởi quản trị viên cơ sở dữ liệu (DBA) hoặc hacker chiếm quyền, hệ thống sẽ mất khả năng quy trách nhiệm (Non-repudiation).

Các giải pháp thay thế:
1.  **Sử dụng MongoDB / CSDL NoSQL:** RLS rất khó cấu hình chặt chẽ ở mức cơ sở dữ liệu, không có cơ chế trigger WORM mạnh mẽ như SQL.
2.  **Lưu trữ Audit log lên Blockchain:** Tuyệt đối an toàn và bất biến, nhưng gây ra độ trễ cực lớn, tiêu tốn tài nguyên và tăng độ phức tạp vận hành lab không đáng có.

---

## 2. Quyết định (Decision)
Lựa chọn **PostgreSQL 16** làm hệ cơ sở dữ liệu chính, áp dụng hai lớp phòng thủ cứng tại tầng DB:
- **Row-Level Security (RLS):** Bật RLS trên bảng nghiệp vụ. Mọi truy vấn từ Gateway kết nối bằng tài khoản `app_user` đều phải đi qua chính sách lọc tự động dựa trên biến context `app.tenant_id`.
- **WORM Trigger & Hash-chaining:** 
  - Tạo trigger chặn hoàn toàn các lệnh `UPDATE` và `DELETE` trên bảng `audit_logs` ở tầng database.
  - Tạo trigger tự động tính toán hash SHA-256 xâu chuỗi liên kết các bản ghi audit log (`block_hash` dòng N phụ thuộc trực tiếp vào `block_hash` dòng N-1).

---

## 3. Hệ quả (Consequences)

### Điểm tốt (Pros):
- Đảm bảo an toàn dữ liệu chiều sâu (Defense-in-Depth). Kể cả khi Gateway bị tấn công SQL Injection, RLS của Postgres vẫn chặn đứng hành vi đọc chéo dữ liệu của Tenant khác.
- Lịch sử kiểm toán được bảo vệ tuyệt đối: Ngay cả admin cũng không thể chỉnh sửa log giao dịch mà không làm đứt gãy chuỗi liên kết Hash (làm đứt chuỗi sẽ bị phát hiện ngay lập tức qua script kiểm định).
- PostgreSQL là CSDL tiêu chuẩn, hiệu năng cao, dễ cấu hình và chạy ổn định trong container Docker Alpine.

### Điểm xấu (Cons):
- API Gateway phải gánh thêm chi phí chạy lệnh `SET LOCAL app.tenant_id = ...` cho mỗi phiên transaction, tăng nhẹ thời gian truy vấn DB.
- Quản lý khóa mật mã để kiểm tra tính toàn vẹn của chuỗi hash cần được thực hiện qua các script tự động hóa.
