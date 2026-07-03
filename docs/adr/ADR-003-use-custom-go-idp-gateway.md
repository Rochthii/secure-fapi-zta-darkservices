# ADR-003: Tự phát triển Identity Provider và API Gateway bằng Go

*   **Status:** APPROVED
*   **Date:** 2026-07-03
*   **Deciders:** Principal Security Architect / Zero Trust Architect

---

## 1. Bối cảnh (Context)
Để hiện thực hóa kiến trúc FAPI-ZTA & Dark Services, chúng ta cần một máy chủ Identity Provider (IdP) cấp token và một API Gateway tiếp nhận traffic.

Các giải pháp thay thế đã được đánh giá:
1.  **Dùng Keycloak / Authentik (Cho IdP):** Quá nặng (ngốn >1GB RAM), cấu hình FAPI/DPoP rất phức tạp qua giao diện, không tích hợp được SDK OpenZiti để biến IdP thành Dark Service.
2.  **Dùng Kong / APISIX (Cho API Gateway):** Kong/APISIX không hỗ trợ cơ chế nhúng SDK OpenZiti để lắng nghe dịch vụ ẩn. Chúng bắt buộc phải chạy ở chế độ proxy mạng truyền thống (Connect-then-Auth).
3.  **Dùng Supabase Auth (Cho IdP):** Supabase Auth không hỗ trợ đặc tả FAPI 2.0 và cơ chế ràng buộc DPoP, chỉ phát hành Bearer Token thông thường.

---

## 2. Quyết định (Decision)
Tự phát triển cả **Identity Provider (IdP)** và **API Gateway** bằng ngôn ngữ **Go (Golang)**:
- Go là ngôn ngữ chính thức được OpenZiti hỗ trợ SDK native tốt nhất.
- Go biên dịch ra file binary tĩnh duy nhất, không cần cài đặt VM/Runtime nặng nề, footprint tài nguyên cực nhỏ (<20MB RAM).
- Việc tự viết code giúp kiểm soát 100% logic xác thực PKCE, giải mã header JWK, tính toán Thumbprint và xây dựng middleware chuỗi xác thực liên tục.

---

## 3. Hệ quả (Consequences)

### Điểm tốt (Pros):
- Đồ án đạt tính học thuật và thực tiễn tối đa (không sử dụng giải pháp "kéo thả" cấu hình có sẵn).
- Dung lượng triển khai siêu nhẹ, có thể dễ dàng đóng gói vào Docker container tối giản (distroless/alpine).
- Dễ dàng tích hợp OpenZiti SDK native thông qua thư viện Go của Ziti.

### Điểm xấu (Cons):
- Khối lượng viết mã nguồn ban đầu lớn (phải tự viết các API đặc tả như JWKS, OIDC Discovery, DPoP signature verification).
- Không có sẵn giao diện quản trị người dùng đẹp mắt như Keycloak (phải giả lập hoặc quản lý database qua SQL).
- Phải tự chịu trách nhiệm về các lỗ hổng bảo mật trong mã nguồn tự viết (được khắc phục qua việc áp dụng bộ kỹ năng go-security và go-reviewer của đại lý).
