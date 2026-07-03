---
description: Quy trình deploy lên staging hoặc production trên Vercel
---

⚠️ KHÔNG deploy production nếu chưa chạy `/qa-testing` và `/review-code` trước.

1. **PRE-DEPLOY CHECKLIST**:
   - Kiểm tra tất cả environment variables trong `.env.local` đã được set trên Vercel chưa?
   - Có migration database nào cần chạy trước không? (Nếu có → chạy `/db-migration` trước)
   - Branch hiện tại đã được push lên GitHub chưa?

2. **CHỌN MÔI TRƯỜNG**:
   - Staging (preview): Deploy từ nhánh `develop` hoặc feature branch
   - Production: Chỉ deploy từ nhánh `main`

3. **KIỂM TRA BUILD CỤC BỘ**:
// turbo
   ```
   npm run build
   ```
   Nếu build lỗi → DỪNG LẠI, không deploy.

4. **DEPLOY**:
   - Staging: `vercel` (không flag `--prod`)
   - Production: `vercel --prod`

5. **POST-DEPLOY VERIFICATION**:
   - Kiểm tra URL đã deploy, trang chủ load được không?
   - Kiểm tra 1 tenant thật (ví dụ: chùa Kleang) xem hoạt động không?
   - Kiểm tra Vercel Function Logs có lỗi bất thường không?

6. **BÁO CÁO**: Báo cáo kết quả deploy (URL, thời gian, trạng thái).
