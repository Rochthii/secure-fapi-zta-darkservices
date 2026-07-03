# Ponytail Ruleset (Tailored for PTIT Thesis SaaS)

Triết lý của DietrichGebert/ponytail là tối giản hóa mã nguồn, chống "Over-engineering" và thúc đẩy tái sử dụng. File này định nghĩa cách áp dụng Ponytail một cách hài hòa trong dự án PTIT Thesis SaaS.

## 1. Thang quyết định (Decision Ladder)

Trước khi viết bất kỳ file mới, component mới hoặc logic nghiệp vụ mới, hãy đi qua 7 bước:

1. **YAGNI (You Ain't Gonna Need It):** Tính năng này có thực sự cần thiết hay chỉ là phỏng đoán tương lai?
2. **Tái sử dụng:** Codebase đã có helper, utility, SQL functions hay custom hooks nào tương tự chưa? (Hãy reuse trước khi viết mới).
3. **Thư viện chuẩn:** Có thể dùng JavaScript/TypeScript Standard Library không?
4. **Native Platform Features:** Có thể giải quyết bằng các tính năng gốc (như HTML5 inputs, CSS native, PostgreSQL RLS Policies, triggers) thay vì cài thêm thư viện JS không?
5. **Dependency hiện tại:** Có thư viện nào đã cài trong `package.json` làm được việc này chưa?
6. **Đơn giản hóa:** Có thể viết gọn thành một dòng hoặc cấu trúc đơn giản không?
7. **Viết code tối thiểu:** Chỉ viết lượng mã tối thiểu để chạy được, bảo đảm xử lý lỗi đầy đủ và bảo mật tuyệt đối.

## 2. Áp dụng cho Backend & Core Security (Nghiêm ngặt)

- **Less Code = Less Attack Surface:** Mã nguồn backend/API càng ít thì lỗ hổng an ninh càng thấp. Giữ logic API gọn gàng.
- **Tối ưu hóa Database:** Tận dụng Row-Level Security (RLS) Custom Claims và chỉ mục B-Tree thay vì viết code lọc phức tạp ở server-side.
- **Tránh Wrapper thừa:** Không tạo các interface, wrapper, hoặc class phức tạp chỉ cho 1-2 hàm đơn giản.

## 3. Ngoại lệ cho Frontend & UI/UX (Premium Aesthetics)

- **Aesthetics Bypass:** Không được áp dụng Ponytail để đơn giản hóa giao diện đến mức thô sơ. Dự án yêu cầu giao diện cực kỳ premium, sống động (Framer Motion, SVG animations, hover effects, dark modes).
- **YAGNI cho UI (Production-Real):** Giao diện phải đẹp nhưng không được chứa dữ liệu giả lập (mock data) hay các nút bấm ảo không có chức năng thực tế. Mọi thành phần UI tương tác phải gọi xuống API thật và ghi nhận side-effect thật.

## 4. Xử lý lỗi và Bảo mật (Không khoan nhượng)

- Ponytail khuyến khích viết ít code, nhưng **tuyệt đối không được bỏ qua**:
  - Xác thực ngữ cảnh (ABAC).
  - Kiểm tra quyền hạn Server-side (không chỉ ẩn UI).
  - Ghi Audit Logs bất biến (WORM Vault).
  - Xử lý lỗi chi tiết và tường minh.
