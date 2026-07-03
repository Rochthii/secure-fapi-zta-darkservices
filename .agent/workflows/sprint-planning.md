---
description: Lên kế hoạch sprint mới — phân tích backlog và phân công công việc
---

// turbo-all

Dùng workflow này khi bắt đầu một sprint mới (thường 1-2 tuần/sprint).

1. **ĐỌC BACKLOG**: Kiểm tra `docs/` và tìm các tính năng/bug còn tồn đọng. Liệt kê chúng.

2. **ƯU TIÊN HOÁ**: Phân loại từng item theo mức độ ưu tiên:
   - 🔴 P0 — Khẩn cấp (ảnh hưởng sản xuất, security)
   - 🟠 P1 — Quan trọng (tính năng cốt lõi chưa có)
   - 🟡 P2 — Trung bình (cải thiện UX, tối ưu hoá)
   - 🟢 P3 — Mong muốn (nice-to-have)

3. **ƯỚC LƯỢNG**: Với mỗi item, ước lượng độ phức tạp:
   - S (Small): < 1 buổi
   - M (Medium): 1-2 buổi
   - L (Large): > 2 buổi → cần chia nhỏ hơn

4. **TẠO SPRINT PLAN**: Tạo file `docs/sprint/SPRINT_[số]_[ngày].md` với:
   - Mục tiêu sprint
   - Danh sách task theo thứ tự ưu tiên
   - Tiêu chí hoàn thành (Definition of Done)

5. **BÁO CÁO**: Trình bày sprint plan và hỏi tôi có muốn điều chỉnh gì không.
