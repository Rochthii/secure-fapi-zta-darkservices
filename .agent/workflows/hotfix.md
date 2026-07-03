---
description: Quy trình hotfix khẩn cấp — sửa lỗi nghiêm trọng trên production
---

⚠️ WORKFLOW NÀY CHỈ DÙNG KHI CÓ BUG NGHIÊM TRỌNG TRÊN PRODUCTION.
Ưu tiên tốc độ nhưng KHÔNG bỏ qua bước kiểm tra an toàn.

1. **XÁC ĐỊNH MỨC ĐỘ**: Phân loại bug:
   - SEV1: Hệ thống sập hoàn toàn / dữ liệu bị mất / security breach
   - SEV2: Tính năng quan trọng không hoạt động, ảnh hưởng nhiều người dùng
   - SEV3: Tính năng phụ lỗi, có cách workaround

2. **PHÂN TÍCH NHANH**: Trong vòng 5 phút phải xác định:
   - Lỗi xảy ra ở file nào / function nào?
   - Nguyên nhân là gì?
   - Có thể fix nhanh không hay cần rollback?

3. **QUYẾT ĐỊNH**: 
   - Nếu có thể fix nhanh (< 30 phút) → tiến hành fix
   - Nếu phức tạp → đề xuất tôi rollback deployment trước

4. **FIX & TEST NHANH**: Sửa bug, kiểm tra logic không bị regression.

5. **COMMIT KHẨN CẤP**: Commit với message prefix `hotfix:` và push ngay.
   Ví dụ: `hotfix: fix missing tenant filter in news query`

6. **BÁO CÁO SỰ CỐ**: Tạo báo cáo ngắn gọn:
   - Thời gian phát hiện / thời gian fix
   - Nguyên nhân gốc rễ
   - Giải pháp đã áp dụng
   - Cách phòng tránh lần sau
