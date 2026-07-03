---
description: Kiểm thử toàn diện trước khi release — QA checklist
---

Thực hiện QA checklist đầy đủ. Báo cáo kết quả rõ ràng bằng tiếng Việt.

### Phase 1: Kiểm tra Chức năng Cốt lõi

1. **AUTHENTICATION**:
   - Đăng nhập admin hoạt động?
   - Phân quyền (tenant admin vs super admin) đúng không?
   - Session expire xử lý đúng?

2. **MULTI-TENANT ROUTING**:
   - Mỗi domain trỏ đúng đến tenant của nó?
   - Tenant A không thể xem dữ liệu của tenant B?
   - Domain không tồn tại trả về 404 đúng cách?

3. **TRANG CÔNG KHAI**:
   - Trang chủ mỗi chùa load đúng theme?
   - Tin tức, Pháp thoại, Liên hệ hoạt động?
   - i18n (vi/km/en) chuyển ngôn ngữ đúng?

4. **ADMIN PANEL**:
   - CRUD tin tức hoạt động?
   - Upload ảnh hoạt động?
   - Form liên hệ gửi được email?

### Phase 2: Kiểm tra Phi chức năng

5. **HIỆU NĂNG**:
   - Trang chủ load < 3 giây (mạng bình thường)?
   - Middleware chạy < 5ms?

6. **BẢO MẬT**:
   - Không có API endpoint nào expose dữ liệu chéo tenant?
   - Các form có validate input không?

### Phase 3: Báo cáo

7. Tổng hợp kết quả:
   - ✅ PASS: Hoạt động đúng
   - ❌ FAIL: Có vấn đề (mô tả chi tiết)
   - ⚠️ WARNING: Cần theo dõi thêm
