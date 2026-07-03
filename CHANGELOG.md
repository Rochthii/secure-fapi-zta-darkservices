# Changelog

Tài liệu này ghi nhận toàn bộ lịch sử thay đổi, nâng cấp và tiến trình triển khai của dự án **Secure FAPI-ZTA Dark Services Gateway**.

---

## [Hoàn thành Phase 3, 4 & 5] — 2026-07-03

### Added
- **Mạng Overlay OpenZiti (Phase 3)**:
  - Khởi tạo script [setup-ziti-services.sh](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/scripts/setup-ziti-services.sh) để tự động cấu hình dịch vụ `financial-ledger-service`, tạo 5 identities và phân quyền Bind/Dial.
  - Khởi tạo script [enroll-identities.sh](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/scripts/enroll-identities.sh) để enroll mã `.jwt` lấy chứng chỉ X.509 mạng ảo và sinh file config JSON.
  - Các tệp cấu hình JSON danh tính được lưu cục bộ tại [docker/identities/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docker/identities/).
- **Dark Services API Gateway (Phase 4)**:
  - Thiết lập modular Go module tại [gateway/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/gateway/).
  - Tạo `gateway/internal/ziti/ziti.go` lắng nghe ẩn qua OpenZiti SDK.
  - Tạo `gateway/internal/auth/crypto.go` cache JWKS và xác thực DPoP Proof.
  - Tạo `gateway/internal/auth/jti.go` cache jti chống Replay.
  - Tạo `gateway/internal/middleware/conn.go` trích xuất Ziti Identity bằng reflection.
  - Tạo `gateway/internal/middleware/auth.go` chạy chuỗi xác thực liên tục `SecureAPI` và phân quyền RBAC `RequireRole`.
  - Tạo `gateway/internal/audit/db.go` kết nối Postgres và tiêm context RLS transaction.
  - Tạo `gateway/internal/api/handlers.go` các endpoint `/api/balance`, `/api/transfer` và `/api/audit-logs`.
- **Client Application (Phase 5)**:
  - Thiết lập modular Go module tại [client/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/client/).
  - Tạo `client/crypto/crypto.go` sinh PKCE và tự động ký DPoP Proof.
  - Tạo `client/ziti/ziti.go` nhúng bộ quay số dialer Ziti vào HTTP Client.
  - Tạo `client/main.go` giao diện CLI thực hiện Authorization PKCE, DPoP exchange và gửi request tàng hình qua Ziti.

### Changed
- **Bảo mật**: Cập nhật [.gitignore](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/.gitignore) để phớt lờ thư mục `docker/tokens/` và `docker/identities/` chứa khóa mật mã học.
- **Tài liệu**: Cập nhật [docs/14_PROJECT_STRUCTURE.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/14_PROJECT_STRUCTURE.md) khớp với cấu trúc mã nguồn Go thực tế.
- **Quy trình Lộ trình**: Đẩy Giai đoạn 6 (Validation & Testing) lên trước Giai đoạn 7 (Observability) trong [13_IMPLEMENTATION_ROADMAP.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/13_IMPLEMENTATION_ROADMAP.md) và [16_FINAL_MASTER_PLAN.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/16_FINAL_MASTER_PLAN.md).
- **Sanitization**: Làm sạch file `.agent/skills/skill-api-design.md` loại bỏ key ví dụ `sk_live_abc123` để triệt tiêu cảnh báo bảo mật từ GitHub.

### Security Validation
- Chạy thử nghiệm **Ziti Policy Advisor** kiểm thử quyền mạng ảo. Kết quả: `client-alice` và `client-bob` được thông qua (`Dial: Y`), `client-evil` bị từ chối kết nối ngay ở lớp mạng ảo overlay (`Dial: N`).
- Kiểm thử biên dịch thành công cả hai module `gateway` và `client` trên compiler Go 1.25.0.

---

## [Hoàn thành Phase 2.5] — 2026-07-03

### Added
- **Đặc tả Thiết kế Kiến trúc**:
  - Tạo tài liệu Threat Modeling [docs/security/threat-model.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/security/threat-model.md) phân tích mô hình STRIDE của từng thành phần qua các ranh giới tin cậy.
  - Tạo chỉ mục và các tài liệu ADR trong thư mục [docs/adr/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/adr/):
    - `ADR-001`: Quyết định chọn OpenZiti SDK binding làm hạ tầng mạng tàng hình.
    - `ADR-002`: Quyết định sử dụng FAPI 2.0 Security Profile và chữ ký ES256 (ECC P-256).
    - `ADR-003`: Quyết định tự viết IdP và Gateway bằng Go.
    - `ADR-004`: Quyết định chọn Postgres RLS cô lập Tenant và trigger WORM liên kết hash-chain SHA-256 bảo vệ audit logs.
  - Tạo tệp tin [docs/diagrams/sequence_flows.md](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docs/diagrams/sequence_flows.md) chứa các sơ đồ luồng Mermaid (Login, Service Access, Ziti Enrollment).

---

## [Hoàn thành Phase 1 & 2] — 2026-07-03

### Added
- **Hạ tầng nền tảng (Phase 1)**:
  - Khởi tạo repo git và remote trên GitHub.
  - Thiết lập cụm Docker Compose [docker/docker-compose.yml](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docker/docker-compose.yml) chứa OpenZiti Controller, Edge Routers, ZAC Console và PostgreSQL.
  - Viết file khởi tạo schema database [docker/postgres/init.sql](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/docker/postgres/init.sql) hỗ trợ RLS và trigger WORM hash-chain.
  - Tạo các script OpenSSL ECC P-256 CA tại thư mục [certs/scripts/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/certs/scripts/).
- **Identity Provider (Phase 2)**:
  - Khởi tạo Go module tại [idp/](file:///e:/Projects/Project_TN/secure-fapi-zta-darkservices/idp/).
  - Triển khai config sinh khóa ký ES256, logic verify PKCE SHA-256, verify DPoP proof, RAMStore lưu mã code và jti cache.
  - Viết các Handler OIDC Discovery, JWKS, Authorize, Token phát hành DPoP-bound Access Token (exp 60s).
  - Viết Unit Tests đầy đủ cho PKCE và DPoP.
