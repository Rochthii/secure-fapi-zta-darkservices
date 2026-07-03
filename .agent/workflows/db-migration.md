---
description: Quy trình an toàn để tạo và áp dụng Supabase database migration
---

⚠️ ĐÂY LÀ QUY TRÌNH NGUY HIỂM. Thao tác sai có thể mất dữ liệu. Tuân thủ nghiêm ngặt từng bước.

1. **PHÂN TÍCH YÊU CẦU**: Mô tả rõ ràng thay đổi schema cần thực hiện (table mới, cột mới, index, RLS policy...).

2. **TẠO FILE MIGRATION**: Tạo file mới trong `supabase/migrations/` theo format:
   ```
   YYYYMMDDHHMMSS_ten_migration_ro_rang.sql
   ```
   Nội dung phải bao gồm:
   - SQL thực hiện thay đổi
   - RLS policy tương ứng nếu tạo table mới
   - Comment giải thích mục đích của migration

3. **KIỂM TRA LOGIC SQL**:
   - Có transaction (`BEGIN/COMMIT`) cho các thay đổi phức tạp không?
   - Có xử lý rollback nếu thất bại không?
   - Không có `DROP TABLE` mà không có `IF EXISTS` kiểm tra trước

4. **REVIEW**: Trình bày file migration cho tôi kiểm tra trước khi áp dụng.

5. **ÁP DỤNG**: Chỉ sau khi tôi xác nhận, mới apply migration vào Supabase (local hoặc remote).

6. **XÁC NHẬN**: Kiểm tra migration đã được apply thành công, không có lỗi.
