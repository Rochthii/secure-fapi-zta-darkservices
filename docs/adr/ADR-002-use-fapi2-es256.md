# ADR-002: Áp dụng FAPI 2.0 Security Profile & Mã hóa ES256

*   **Status:** APPROVED
*   **Date:** 2026-07-03
*   **Deciders:** Principal Security Architect / Zero Trust Architect

---

## 1. Bối cảnh (Context)
OAuth 2.0 và JWT Bearer Token truyền thống có các kẽ hở bảo mật nghiêm trọng trong môi trường Tài chính/Fintech:
- **Token Theft:** Token bị đánh cắp khỏi bộ nhớ trình duyệt hoặc log mạng có thể được sử dụng lại trên bất kỳ thiết bị nào khác (vì Bearer token không kiểm chứng người mang khóa).
- **Replay Attack:** Client có thể bị lừa gửi lại cùng một token nhiều lần để thực thi giao dịch trùng lặp.
- **Authorization Code Interception:** Authorization Code có thể bị chặn bắt trên thiết bị công cộng nếu không có PKCE bảo vệ.

Các phương án thay thế:
1.  **mTLS-bound Access Tokens (RFC 8705):** Ràng buộc token với chứng chỉ mTLS tầng Transport. Tốt, nhưng phụ thuộc hoàn toàn vào TLS stack, khó triển khai nếu đi qua nhiều tầng proxy trung gian.
2.  **DPoP (RFC 9449):** Ràng buộc token ở tầng ứng dụng (Application Layer) bằng chữ ký số bất đối xứng per-request. Rất linh hoạt, đi qua được mọi proxy.

---

## 2. Quyết định (Decision)
Áp dụng **FAPI 2.0 Security Profile (Final, 2025)** kết hợp cả **DPoP (RFC 9449)** và **mTLS (RFC 8705)** để tạo cơ chế phòng vệ kép (Dual Token Binding):
- Bắt buộc sử dụng thuật toán mã hóa **ES256 (ECDSA sử dụng Elliptic Curve P-256)** cho cả khóa ký DPoP của Client và khóa ký JWT của IdP.
- Loại bỏ hoàn toàn các thuật toán RSA lỗi thời (như RS256) và các thuật toán symmetric key kém an toàn (như HS256).
- Mọi giao dịch tài chính phải đi kèm DPoP Proof JWT chứa `ath` (mã hash của access token) và mã định danh độc bản `jti` chống trùng lặp.

---

## 3. Hệ quả (Consequences)

### Điểm tốt (Pros):
- Đạt chuẩn bảo mật an toàn cao nhất của OpenID Foundation dành cho Fintech/Ngân hàng số.
- Vô hiệu hóa hoàn toàn rủi ro Token Theft: Kể cả khi hacker lấy trộm được Access Token, họ cũng không dùng được nếu không sở hữu Private Key của Client để tạo chữ ký DPoP tương ứng.
- Khóa ECC P-256 có kích thước cực nhỏ (256-bit) nhưng hiệu năng ký/verify nhanh hơn và an toàn tương đương khóa RSA 3078-bit, tiết kiệm băng thông truyền tải Header.

### Điểm xấu (Cons):
- Client phải liên tục sinh chữ ký số cho mỗi request gọi API → Gây hao tổn nhẹ CPU trên thiết bị Client.
- IdP và Gateway phải duy trì bộ nhớ đệm `jti` để kiểm tra trùng lặp chữ ký số của DPoP Proof, đòi hỏi cấu hình dọn dẹp RAM định kỳ.
