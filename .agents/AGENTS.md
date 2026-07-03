# Secure FAPI-ZTA & Dark Services - AI Coding Rules

## 🎯 Bối Cảnh Dự Án (Project Context)
Dự án này là hệ thống giao dịch tài chính bảo mật cao tuân thủ Zero Trust Architecture (NIST SP 800-207), ngân hàng cấp độ FAPI 2.0 và mạng tàng hình OpenZiti.
- **Lớp Mạng (OpenZiti Overlay):** Gateway ẩn hoàn toàn, không mở cổng TCP inbound. Mọi luồng mạng client-server sử dụng SDK OpenZiti binding.
- **Lớp Ứng Dụng (FAPI 2.0 + DPoP + mTLS):** Xác thực kép mTLS X.509 mạng ảo + DPoP JWT (RFC 9449) ràng buộc thiết bị. Ràng buộc chéo (Cross-Layer Binding): Gateway khớp `SourceIdentifier` (Ziti Identity) với `sub` claim của token.
- **Lớp Dữ Liệu (Postgres RLS + WORM):** Cách ly Tenant bằng Row-Level Security (`set_config`). Logs giao dịch bất biến (WORM) dùng trigger cấm UPDATE/DELETE và chuỗi băm SHA-256 Hash-chaining.

---

## 🛠️ Quy Tắc Lập Trình Cho AI (AI Coding Rules)

### 1. Nguyên Tắc Vibe Coding & Code Sạch (Clean Code & No Mock)
- **TUÂN THỦ TUYỆT ĐỐI:** Không viết code demo, dữ liệu giả lập (mock data), nút bấm giả, hoặc các hàm giả lập thành công trong các luồng chính. 
- Mọi API, nút bấm, và hành động trên UI/CLI phải tác động thật tới Database/Hạ tầng.
- Giữ nguyên các chú thích (comments) và tài liệu mã nguồn hiện tại nếu không liên quan trực tiếp đến chỉnh sửa.
- Ưu tiên hiệu quả, chính xác và tập trung vào mục tiêu, tránh viết code dư thừa hoặc các hàm helper không sử dụng.

### 2. An Toàn Database & Postgres Row-Level Security (RLS)
- **Không lọc thủ công:** Tuyệt đối không dựa vào điều kiện `WHERE tenant_id = ?` ở tầng ứng dụng Go để bảo mật. Mọi truy cập bắt buộc phải được bảo vệ bởi PostgreSQL RLS.
- **Tiêm Context Transaction:** Mọi thao tác truy vấn trong Transaction phải tiêm context tenant qua biến cấu hình: `SELECT set_config('app.current_tenant', $1, true)` trước khi thực hiện bất kỳ câu lệnh SELECT/INSERT/UPDATE nào.
- **Tránh rò rỉ Context Pool:** Khi sử dụng connection pooling (`pgxpool`), hãy đảm bảo context session được thiết lập/reset chính xác theo từng request, tránh rò rỉ dữ liệu giữa các connection của client khác nhau.
- **SECURITY INVOKER:** Sử dụng `SECURITY INVOKER` theo mặc định cho các stored procedure và database functions. Chỉ sử dụng `SECURITY DEFINER` khi thực sự cần bỏ qua RLS và phải có comment giải trình lý do rõ ràng.
- **Logs WORM Bất Biến:** Bảng `audit_logs` là bảng WORM bất biến. Không bao giờ được phép dùng câu lệnh `UPDATE` hoặc `DELETE` trên bảng này. 
- **Hash-chaining:** Khi chèn bản ghi mới vào `audit_logs`, hàm băm `hash` phải được tính toán tự động bằng trigger băm liên kết `SHA-256` dựa trên dòng dữ liệu trước đó để đảm bảo toàn vẹn chuỗi log.

### 3. Quy Tắc Bảo Mật Go & API Gateway (FAPI 2.0 & OpenZiti)
- **Xác thực Phía Server:** Mọi kiểm tra quyền hạn (RBAC) và xác thực mTLS/DPoP phải được kiểm tra từ phía Server/Gateway, không tin cậy Client.
- **DPoP Verification:** Token DPoP JWT phải chứa `jti` chống replay attack (phải cache và kiểm tra jti), kiểm tra thời gian hết hạn (`exp`) chặt chẽ (mặc định 60 giây), và chữ ký phải khớp với khóa công khai (ES256).
- **OpenZiti Binding:** Khi triển khai client dialer hoặc server listener, luôn sử dụng SDK thuần của Go (`ziti.Context.Dial` hoặc `ziti.Context.Listen`), không gọi trực tiếp thông qua network socket của hệ điều hành.

### 4. Quy Trình Kiểm Thử & Xác Minh (Testing & Verification)
- Bắt buộc tuân thủ các kịch bản kiểm thử bảo mật nâng cao và các chỉ dẫn kỹ thuật tránh lỗi (Go cache, server task restart) được quy định tại workflow [qa-testing.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/.agent/workflows/qa-testing.md).

### 5. Quy Tắc Đồng Bộ Tài Liệu, Lộ Trình & Nhật Ký (Documentation, Roadmap & Changelog Sync)
- Sau khi hoàn thành code, bắt buộc thực hiện đồng bộ tiến độ, nhật ký thay đổi và tài liệu thiết kế theo chỉ dẫn tại workflow [update-docs.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/.agent/workflows/update-docs.md).

---

## 📂 Sơ Đồ File Quan Trọng (Key File Map)
- **Database Init & RLS/WORM:** [init.sql](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docker/postgres/init.sql)
- **Identity Provider (IdP):** [idp/main.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/idp/main.go)
- **API Gateway Entry:** [gateway/main.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/main.go)
- **Gateway Middlewares (Cross-layer & Auth):** [gateway/internal/middleware/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/internal/middleware/)
- **Client Application Entry:** [client/main.go](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/client/main.go)
