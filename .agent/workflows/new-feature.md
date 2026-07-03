---
description: Quy trình phát triển một tính năng mới từ đầu đến cuối
---

Đây là quy trình chuẩn để phát triển một tính năng mới. Hãy đọc kỹ yêu cầu của tôi trước khi bắt đầu.

1. **PLANNING MODE**: Phân tích yêu cầu và tạo `implementation_plan.md` bao gồm:
   - Mô tả tính năng
   - Danh sách file cần tạo/sửa
   - Các bước thực hiện theo thứ tự
   - Kế hoạch kiểm tra

2. **REVIEW**: Trình bày plan cho tôi và chờ phê duyệt. KHÔNG bắt đầu code trước khi được đồng ý.

3. **EXECUTION MODE**: Sau khi được phê duyệt, thực hiện theo đúng plan:
   - Tạo hoặc sửa từng file theo thứ tự đã đề ra
   - Tuân thủ các quy tắc trong `GLOBAL_RULES`
   - Không tạo mock/demo data, chỉ dùng logic thật

4. **VERIFICATION**: Sau khi code xong:
   - Chạy TypeScript check
   - Kiểm tra các edge case
   - Đảm bảo RLS đã được enforce ở server-side nếu liên quan đến dữ liệu multi-tenant

5. **REPORT**: Tạo `walkthrough.md` tóm tắt những gì đã làm.

6. Gọi workflow `/push` để commit và push.
