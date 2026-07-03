# PART 18 — ARCHITECTURE DECISION RECORDS (ADR) INDEX

Tài liệu này quản lý danh sách các quyết định thiết kế kiến trúc quan trọng cho dự án **secure-fapi-zta-darkservices**. Mỗi quyết định được ghi nhận chi tiết theo mẫu chuẩn ADR để lưu lại bối cảnh, lý do lựa chọn và hệ quả lâu dài của thiết kế.

---

## Danh sách quyết định kiến trúc (ADR List)

| Mã ADR | Tiêu đề quyết định | Trạng thái | Ngày quyết định |
|---|---|---|---|
| **[ADR-001](./ADR-001-use-openziti.md)** | Lựa chọn OpenZiti làm hạ tầng mạng Zero Trust Overlay | **APPROVED** | 2026-07-03 |
| **[ADR-002](./ADR-002-use-fapi2-es256.md)** | Áp dụng FAPI 2.0 Security Profile & Mã hóa ES256 | **APPROVED** | 2026-07-03 |
| **[ADR-003](./ADR-003-use-custom-go-idp-gateway.md)** | Tự phát triển Identity Provider và API Gateway bằng Go | **APPROVED** | 2026-07-03 |
| **[ADR-004](./ADR-004-use-postgres-worm.md)** | Sử dụng PostgreSQL Row-Level Security & Trigger WORM | **APPROVED** | 2026-07-03 |

---

## Cấu trúc chuẩn của một ADR
Mỗi tài liệu ADR được cấu trúc theo 3 phần chính:
1.  **Bối cảnh (Context):** Mô tả vấn đề cần giải quyết, các ràng buộc kỹ thuật và các giải pháp thay thế khả thi.
2.  **Quyết định (Decision):** Giải pháp kiến trúc được lựa chọn và cơ sở lập luận kỹ thuật cho lựa chọn đó.
3.  **Hệ quả (Consequences):** Những lợi ích thu được (Pros) và những thách thức/nợ kỹ thuật (Cons) mà quyết định này mang lại cho hệ thống.
