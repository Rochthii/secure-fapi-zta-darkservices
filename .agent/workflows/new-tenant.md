---
description: Quy trình onboard một đơn vị thành viên (tenant) mới vào hệ thống
---

Dùng workflow này khi cần thêm một tổ chức/đơn vị thành viên (tenant) mới vào hệ thống multi-tenant FAPI-ZTA.

1. **THU THẬP THÔNG TIN**: Yêu cầu cung cấp:
   - Tên tổ chức/đơn vị thành viên (tiếng Việt, tiếng Anh)
   - Tên viết tắt định danh (ví dụ: `tenant-abc`)
   - Domain kết nối mạng ảo (ví dụ: `tenant-abc.api`)
   - Địa chỉ liên hệ, email, số điện thoại người đại diện

2. **CẤU HÌNH DATABASE**:
   - Chèn bản ghi mới vào bảng `tenants` trong PostgreSQL:
     - `id`: Sinh mã UUID mới
     - `name`: Tên đơn vị
     - `slug`: Tên viết tắt định danh
   - Tạo tài khoản người dùng/nhân viên ban đầu cho tenant đó trong bảng liên quan.

3. **CẤU HÌNH PHÂN QUYỀN MẠNG ẢO (OPENZITI)**:
   - Tạo danh tính mạng ảo (Ziti Identity) mới cho thiết bị của tenant:
     `ziti edge create identity device <tenant-slug>-client`
   - Ghi danh (enroll) danh tính để sinh tệp `.json` kết nối.
   - Gán chính sách truy cập (Dial Policy) cho danh tính mới kết nối tới dịch vụ `financial-ledger-service`.

4. **KIỂM TRA XÁC MINH**:
   - Chạy lệnh test gọi API balance với tenant context mới để kiểm tra:
     - Cơ chế RLS cô lập dữ liệu hoạt động chính xác (không thấy chéo dữ liệu của tenant khác).
     - Token DPoP được cấp phát và verify thành công từ IdP.

5. **BÁO CÁO**: Tổng hợp thông tin credentials mạng ảo và token config vừa tạo bàn giao cho tenant mới.
