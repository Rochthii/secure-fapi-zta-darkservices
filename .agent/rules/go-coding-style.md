# Quy chuẩn viết code Go (Golang Coding Style)

Tài liệu này định nghĩa quy chuẩn viết code Go cho dự án **secure-fapi-zta-darkservices**.

---

## 1. Nguyên tắc cơ bản

- **Go Idioms**: Tuân thủ các nguyên tắc thiết kế của Go (Simple, Explicit, Readable). Tránh viết code quá phức tạp hoặc lạm dụng OOP.
- **Formatting**: Bắt buộc định dạng code bằng `go fmt` hoặc `goimports` trước khi commit.
- **Explicit Returns**: Không dùng named return values trừ khi thực sự cần thiết để cải thiện readability của các hàm phức tạp.

---

## 2. Quản lý lỗi (Error Handling)

- **Không bỏ qua lỗi**: Tuyệt đối không dùng `_` để bỏ qua lỗi trả về, trừ các trường hợp ghi log/close connection phụ.
- **Error Wrapping**: Sử dụng `%w` để wrap error kèm bối cảnh:
  ```go
  if err != nil {
      return fmt.Errorf("failed to validate DPoP proof: %w", err)
  }
  ```
- **Sentinel Errors**: Định nghĩa lỗi tĩnh ở cấp độ package cho các lỗi nghiệp vụ phổ biến:
  ```go
  var ErrInvalidDPoPProof = errors.New("dpop proof is invalid")
  ```

---

## 3. Quản lý Concurrency

- **Goroutine Lifetime**: Mỗi khi start một Goroutine mới, phải xác định rõ khi nào nó kết thúc. Tránh rò rỉ Goroutine (goroutine leak).
- **Context Propagation**: Truyền `context.Context` xuyên suốt các lớp xử lý (giao tiếp mạng, gọi DB) để quản lý timeout và hủy bỏ (cancellation).
- **Mutex**: Lock mutex ngay trước dòng cần bảo vệ dữ liệu và `defer mutex.Unlock()` ở dòng tiếp theo.

---

## 4. Đặt tên (Naming Conventions)

- **Package Name**: Đặt tên package ngắn gọn, viết thường toàn bộ (lowercase), không dùng snake_case hay camelCase (ví dụ: `middleware`, `crypto`, `audit`).
- **Interfaces**: Tên interface nên kết thúc bằng hậu tố `er` nếu chỉ có 1 method (ví dụ: `Writer`, `Validator`).
- **Receiver Names**: Sử dụng 1-2 ký tự đại diện viết thường (ví dụ: `func (g *Gateway) Serve(...)`). Không dùng `this` hay `self`.

---

## 5. Struct và Memory Allocation

- **Pointers vs Values**:
  - Dùng pointer receiver (`*T`) khi struct cần được sửa đổi dữ liệu hoặc có kích thước lớn.
  - Dùng value receiver (`T`) cho các struct nhỏ, immutable.
- **Slices & Maps Initialisation**: Ưu tiên khởi tạo slice/map với dung lượng định trước (nếu biết trước kích thước) để tối ưu memory allocation:
  ```go
  users := make([]User, 0, len(rawUsers))
  ```
