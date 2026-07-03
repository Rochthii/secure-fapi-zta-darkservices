---
description: Kiểm tra bảo mật định kỳ — phát hiện lỗ hổng trước khi bị khai thác
---

Thực hiện security audit toàn diện. Báo cáo theo thứ tự ưu tiên giảm dần.

### 1. Kiểm tra RLS & Tenant Isolation

- Mọi Supabase query có filter `tenant_id` chưa?
- Server Actions có validate `tenantId` từ session chứ không lấy từ client request?
- Có route nào bypass middleware tenant-check không?

### 2. Kiểm tra Authentication & Authorization

- Các route admin có được bảo vệ bởi `auth()` chưa?
- Super admin features có check role `super_admin` ở server-side không?
- Có chỗ nào chỉ check permission ở UI mà không check ở API?

### 3. Kiểm tra Input Validation

- Các form submit có validate và sanitize input không?
- SQL injection: có nơi nào dùng string interpolation trong SQL không?
- File upload: có kiểm tra loại file và kích thước không?

### 4. Kiểm tra Secret Exposure

- Không có `SUPABASE_SERVICE_ROLE_KEY` hay secret key nào trong code client-side?
- `.env.local` không có trong `.gitignore` chứ?
- Không có hardcode API key nào trong source code?

### 5. Kiểm tra Audit Logging

- Các hành động quan trọng (tạo/xoá tin tức, thay đổi settings...) có `audit_log` không?
- Log có ghi đủ: ai, làm gì, khi nào, từ đâu?

### 6. Báo cáo

Phân loại kết quả:
- 🔴 CRITICAL: Phải sửa ngay
- 🟠 HIGH: Sửa trong sprint này
- 🟡 MEDIUM: Lên backlog
- 🟢 OK: Không có vấn đề
