---
description: Quy trình dọn dẹp ngữ cảnh (Context Compaction) để tối ưu token trong các phiên làm việc dài theo phong cách ECC.
---

# Quy trình /compact (ECC Optimized)

Khi ngữ cảnh bắt đầu trở nên quá tải (phản hồi chậm, tốn token), hãy thực hiện các bước sau:

1. **Tổng kết trạng thái (Final Sync):** 
   - Đọc kỹ `CLAUDE.md` và cập nhật các "Context Memory" gần nhất.
   - Đảm bảo các `implementation_plan.md` hoặc `task.md` đang ở trạng thái mới nhất.

2. **Ghi nhớ các giá trị then chốt:**
   - Liệt kê các biến môi trường, đường dẫn file, hoặc ID quan trọng trong câu trả lời cuối cùng trước khi compact.

3. **Thực hiện Compact:**
   - Tạo một bản tóm tắt ngắn gọn (Dưới 300 từ) về: 
     - [ ] Những gì đã làm được.
     - [ ] Những lỗi/góc khuất đã gặp.
     - [ ] Bước tiếp theo cần làm ngay lập tức.
   - Sau đó yêu cầu người dùng gõ lệnh `/compact` (nếu có công cụ) hoặc thông báo: "Hệ thống đã sẵn sàng cho một chu kỳ ngữ cảnh mới".

4. **Tái khởi động (Post-Compact):**
   - Đọc lại `CLAUDE.md` ngay lập tức để tái lập "Hành lang bộ nhớ" (Memory Corridor).
   - Tiếp tục công việc từ mục "Next Step" trong `CLAUDE.md`.

// turbo
// lệnh này chỉ là chỉ dẫn, không chạy terminal.
