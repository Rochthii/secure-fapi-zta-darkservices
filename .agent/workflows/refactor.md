---
description: Tái cấu trúc code — dọn dẹp, tối ưu, không thay đổi chức năng
---

Dùng khi cần refactor mà KHÔNG thay đổi behavior của hệ thống.

**QUY TẮC VÀNG**: Sau khi refactor, hệ thống phải hoạt động y hệt như trước.

1. **XÁC ĐỊNH PHẠM VI REFACTOR**:
   - File / component nào cần refactor?
   - Lý do: code trùng lặp / quá phức tạp / sai pattern / quá chậm?
   - Mục tiêu cụ thể sau khi refactor là gì?

2. **LẬP KẾ HOẠCH**:
   - Liệt kê các thay đổi sẽ thực hiện
   - Xác định các component KHÁC phụ thuộc vào code sẽ thay đổi
   - Đánh giá rủi ro (thấp / trung / cao)

3. **THỰC HIỆN TỪNG BƯỚC NHỎ**:
   - Mỗi bước chỉ thay đổi một việc (single responsibility)
   - Sau mỗi bước nhỏ kiểm tra TypeScript không báo lỗi mới

4. **KIỂM TRA SAU REFACTOR**:
// turbo
   ```
   npx tsc --noEmit
   ```
   Đảm bảo không có lỗi TypeScript mới nào phát sinh.

5. **COMMIT**: Dùng prefix `refactor:` hoặc `chore:` trong commit message.
   Ví dụ: `refactor: extract tenant query logic into shared util`
