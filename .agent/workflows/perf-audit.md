---
description: Phân tích hiệu năng — tìm điểm nghẽn và tối ưu
---

Dùng khi trang web chậm hoặc cần tối ưu hoá định kỳ.

### Phase 1: Đo Lường

1. **Server PERFORMANCE**:
   - Kiểm tra các hàm nào chưa dùng caching (`unstable_cache`, React `cache()`)
   - Có query nào chạy N+1 không? (vòng lặp gọi query từng phần tử)
   - Middleware có đang làm quá nhiều không? (mục tiêu < 5ms)

2. **BUNDLE SIZE**:
// turbo
   ```
   npx next build
   ```
   Xem kết quả bundle size. Flag các chunk > 100KB.

3. **DATABASE**:
   - Các query thường xuyên có index chưa?
   - Các bảng lớn có được query với filter hiệu quả không?

### Phase 2: Tối ưu

4. **CACHING**: Xác định và thêm cache vào các hàm fetch data phù hợp.
5. **LAZY LOADING**: Các component nặng (bản đồ, YouTube...) có được lazy load không?
6. **IMAGE OPTIMIZATION**: Ảnh có dùng `next/image` với `sizes` đúng không?

### Phase 3: Báo cáo

7. Tổng hợp:
   - Những điểm nghẽn đã tìm thấy
   - Những tối ưu đã thực hiện
   - Ước lượng cải thiện hiệu năng (%)
