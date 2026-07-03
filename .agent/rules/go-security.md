# Quy chuẩn bảo mật Go (Golang Security Guidelines)

Tài liệu này định nghĩa quy chuẩn bảo mật khi code Go cho dự án **secure-fapi-zta-darkservices**.

---

## 1. Mật mã học (Cryptography)

- **Crypto Packages**: Chỉ sử dụng package `crypto/` tiêu chuẩn của thư viện chuẩn Go (`crypto/ecdsa`, `crypto/elliptic`, `crypto/rand`, `crypto/sha256`). Tuyệt đối không tự viết thuật toán crypto.
- **Random Number Generation**: Luôn dùng `crypto/rand` (CSPRNG) cho các giá trị nhạy cảm (token, PKCE code, nonces). Không sử dụng `math/rand` (PRNG) cho mục đích bảo mật.
- **Key Generation**: Cặp khóa ECC dùng cho DPoP và mTLS bắt buộc sử dụng curve P-256 (đáp ứng FAPI 2.0).

---

## 2. Phòng tránh lỗi bảo mật bộ nhớ (Memory Security)

- **Unsafe Package**: Tuyệt đối cấm sử dụng package `unsafe` để bypass kiểm tra kiểu hoặc trực tiếp thao tác bộ nhớ.
- **Secrets in Memory**: 
  - Tránh in các giá trị nhạy cảm (private key, passwords, tokens) ra log. 
  - Cấu hình struct tag hoặc hàm `String()` tùy chỉnh để không in giá trị nhạy cảm khi dùng `fmt.Printf("%+v", struct)` hoặc log.
  - Xóa sạch khóa hoặc dữ liệu nhạy cảm ra khỏi RAM khi không còn sử dụng bằng cách ghi đè byte zero (`byte(0)`).

---

## 3. SQL Injection & Database Safety

- **Parameterized Queries**: Luôn dùng parameterized queries (`db.Query("SELECT ... WHERE id = $1", id)`) khi giao tiếp với PostgreSQL thông qua `database/sql`. Không cộng chuỗi SQL trực tiếp.
- **RLS Context Safety**: 
  - Khi thực hiện transaction, middleware bắt buộc phải chạy `SET LOCAL app.tenant_id = ...` đầu tiên.
  - Sử dụng transaction cục bộ (`tx`) và đảm bảo connection được trả về pool đúng hạn.

---

## 4. HTTP & Network Security

- **Server Timeout**: Luôn cấu hình timeout rõ ràng cho HTTP Server để tránh DDoS tấn công chậm (Slowloris):
  ```go
  server := &http.Server{
      ReadTimeout:  5 * time.Second,
      WriteTimeout: 10 * time.Second,
      IdleTimeout:  120 * time.Second,
  }
  ```
- **TLS Config**: Bắt buộc cấu hình TLS tối thiểu 1.3 đối với các giao tiếp mTLS/HTTPS công khai:
  ```go
  tlsConfig := &tls.Config{
      MinVersion: tls.VersionTLS13,
  }
  ```
