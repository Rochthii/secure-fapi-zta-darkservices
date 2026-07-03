---
description: Kiểm tra toàn diện chất lượng code trước khi deploy
---

Thực hiện code review toàn diện cho dự án. Hãy báo cáo bằng tiếng Việt, có cấu trúc rõ ràng.

1. **BẢO MẬT**:
   - Kiểm tra tất cả Server Actions: có enforce `tenantId` đúng cách không?
   - Kiểm tra RLS policies có được gọi từ server-side không (không chỉ UI hiding)?
   - Có key bí mật nào bị expose ra client-side không?

2. **HIỆU NĂNG**:
   - Xác định các query N+1 tiềm ẩn
   - Kiểm tra các hàm nặng có được cache không (unstable_cache, React cache)?
   - Middleware có nhẹ không (< 5ms)?

3. **CHẤT LƯỢNG CODE**:
   - Có TODO/FIXME nào còn sót lại không?
   - Có logic mock/fake/demo nào không (vi phạm GLOBAL RULE)?
   - TypeScript có đang bị bypass bằng `any` không hợp lý không?

4. **MULTI-TENANT**:
   - Mọi query đến Supabase có filter theo `tenant_id` không?
   - Không có chỗ nào dùng hardcode tenant ID không?

5. **BÁO CÁO**: Tổng hợp kết quả theo dạng:
   - 🔴 Vấn đề nghiêm trọng (cần sửa ngay)
   - 🟡 Cần cải thiện (nên sửa)
   - 🟢 Tốt (đang hoạt động đúng)
