---
description: Báo cáo nhanh trạng thái dự án mỗi ngày
---

Thực hiện kiểm tra nhanh và báo cáo tình trạng hiện tại của dự án. Báo cáo bằng tiếng Việt.

1. **KIỂM TRA GIT**: Xem các commit gần nhất (5 commit), có gì thay đổi?

2. **KIỂM TRA LỖI BUILD**:
   - Đọc `build-log.txt` nếu có
   - Kiểm tra các file TODO/FIXME còn tồn đọng

3. **KIỂM TRA TASKS**: Đọc `task.md` (nếu có) trong thư mục artifacts, liệt kê:
   - Công việc đang làm dở
   - Công việc đã hoàn thành gần đây

4. **BÁO CÁO NHANH** theo format:
   ```
   📊 DAILY CHECK - [Ngày hôm nay]

   ✅ Hoàn thành gần đây:
   - ...

   🔄 Đang thực hiện:
   - ...

   ⚠️ Cần chú ý:
   - ...

   🎯 Đề xuất việc tiếp theo:
   - ...
   ```
