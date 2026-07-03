---
description: Quy trình phân tích và sửa một bug bất kỳ một cách có hệ thống
---

Đây là quy trình sửa bug chuyên nghiệp. Hãy đọc kỹ mô tả bug mà tôi cung cấp trước khi bắt đầu.

1. **PHÂN TÍCH**: Đọc mô tả bug. Xác định:
   - Component / file nào liên quan?
   - Đây là lỗi UI, logic, hay dữ liệu?
   - Điều kiện nào gây ra bug (steps to reproduce)?

2. **TÌM NGUYÊN NHÂN GỐC RỄ**: Đọc code liên quan. KHÔNG sửa ngẫu nhiên. Hãy xác định RÕ RÀNG dòng code nào gây ra bug trước khi sửa.

3. **LẬP KẾ HOẠCH SỬA**: Mô tả ngắn gọn cách sửa (1-2 câu). Nếu có nhiều cách, hãy chọn cách ít ảnh hưởng nhất đến các phần khác.

4. **THỰC HIỆN SỬA**: Sửa bug theo kế hoạch đã lập. Chỉ thay đổi những gì cần thiết.

5. **KIỂM TRA**: Kiểm tra lại logic xung quanh vùng code vừa sửa để đảm bảo không gây ra bug mới (regression).

6. **BÁO CÁO**: Trình bày:
   - Nguyên nhân gốc rễ
   - Cách đã sửa
   - Những file đã thay đổi
