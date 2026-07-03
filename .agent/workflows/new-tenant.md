---
description: Quy trình onboard một ngôi chùa (tenant) mới vào hệ thống
---

Dùng workflow này khi cần thêm một chùa mới vào hệ thống multi-tenant.

1. **THU THẬP THÔNG TIN**: Yêu cầu tôi cung cấp:
   - Tên chùa (tiếng Việt, tiếng Khmer, tiếng Anh)
   - Domain (ví dụ: `chua-abc.com`)
   - Địa chỉ, số điện thoại, email liên hệ
   - Theme màu (xem `lib/themes-config.ts` để chọn)
   - Logo URL (nếu có)

2. **TẠO TENANT**: Chèn record mới vào table `tenants` trong Supabase với:
   - `id` mới (UUID)
   - `domain`
   - `name`, `name_km`, `name_en`
   - `theme_id` tương ứng

3. **TẠO SETTINGS**: Chèn dữ liệu mặc định vào `site_settings` cho tenant mới:
   - address, contact_phone, contact_email
   - logo_url, favicon_url
   - opening_hours, facebook_url

4. **KIỂM TRA**: Xác nhận:
   - Middleware routing hoạt động đúng với domain mới
   - Theme được áp dụng đúng
   - Trang chủ và trang liên hệ hiển thị đúng thông tin

5. **BÁO CÁO**: Tổng hợp thông tin tenant vừa tạo và các bước tiếp theo (upload logo, nhập tin tức...).
