---
description: Cập nhật và đồng bộ tài liệu kỹ thuật sau khi có thay đổi lớn
---

Dùng workflow này sau khi hoàn thành một tính năng lớn hoặc thay đổi kiến trúc.

1. **XÁC ĐỊNH PHẠM VI**: Liệt kê những gì đã thay đổi kể từ lần cập nhật tài liệu cuối.

2. **CẬP NHẬT DOCS THEO LOẠI**:

   **Thay đổi kiến trúc / routing:**
   → Cập nhật `docs/ARCHITECTURE.md`

   **Thay đổi database / migration:**
   → Cập nhật `docs/07_DATABASE_MIGRATIONS_RLS.md`

   **Tính năng admin mới:**
   → Cập nhật `docs/05_ADMIN_FEATURES.md`

   **API / Server Actions mới:**
   → Tạo hoặc cập nhật docs tương ứng trong `docs/`

3. **KIỂM TRA CONSISTENCY**: Đảm bảo không có tài liệu nào mô tả tính năng đã bị xoá hoặc thay đổi mà chưa được update.

4. **CHANGELOG**: Thêm entry mới vào `CHANGELOG.md` (nếu có) theo format:
   ```
   ## [Ngày] - Tên thay đổi
   ### Added / Changed / Fixed / Removed
   - Mô tả ngắn gọn
   ```

5. **COMMIT DOCS**: Commit riêng cho docs với message `docs: ...`
