# PART 1 — EXECUTIVE SUMMARY

## 1.1 Project Title

**FAPI-ZTA & Dark Services Architecture**
Financial-grade API Zero Trust Architecture with OpenZiti Dark Services for Invisible API Infrastructure

## 1.2 Mục tiêu dự án

Thiết kế và triển khai một nền tảng API đạt chuẩn bảo mật **Financial-grade** (cấp Ngân hàng/Fintech) với khả năng **tàng hình hoàn toàn trên Internet**, kết hợp:

1. **Zero Trust Architecture (ZTA)** theo chuẩn NIST SP 800-207 — triệt tiêu niềm tin ngầm định, xác thực liên tục mọi phiên.
2. **FAPI 2.0 Security Profile** — chuẩn API cấp cao nhất của OpenID Foundation, bắt buộc sender-constrained tokens.
3. **OpenZiti Dark Services** — biến API Gateway thành dịch vụ tàng hình, không thể phát hiện bằng bất kỳ kỹ thuật quét mạng nào.

## 1.3 Giá trị nghiên cứu

### Giá trị học thuật
- **Tích hợp liên ngành**: Kết hợp lý thuyết Mật mã học ứng dụng (DPoP, mTLS, ECC), Kiến trúc mạng Zero Trust, và Software-Defined Perimeter (SDP) vào một hệ thống thống nhất.
- **Mô hình mới**: Đề xuất mô hình "Defense-in-Depth 7 Layer" — chuỗi phòng thủ chiều sâu 7 lớp từ thiết bị người dùng tới cơ sở dữ liệu.
- **Thực chứng**: Chứng minh bằng thực nghiệm rằng kỹ thuật Dark Service có thể triệt tiêu hoàn toàn bề mặt tấn công mạng của API.

### Giá trị thực tiễn
- Áp dụng trực tiếp cho nền tảng Ngân hàng số, Ví điện tử, Open Banking.
- Tuân thủ chuẩn quốc tế: NIST, OWASP, PCI DSS, ISO 27001, FAPI 2.0.
- Mã nguồn mở, tái sử dụng cho doanh nghiệp.

## 1.4 Điểm mới (Novel Contributions)

| # | Điểm mới | Mô tả |
|---|---|---|
| 1 | **Invisible Financial API** | API Gateway không mở bất kỳ cổng TCP/UDP nào trên Internet. Kẻ tấn công chạy `nmap` quét 65535 cổng → kết quả: 0 cổng mở. |
| 2 | **Triple Token Binding** | Access Token đồng thời ràng buộc bởi 3 yếu tố: DPoP keypair (thiết bị), mTLS certificate (danh tính PKI), và Ziti Identity (danh tính mạng overlay). |
| 3 | **Cryptographic Defense Chain** | Chuỗi phòng thủ mật mã học 7 lớp liên hoàn — mỗi lớp độc lập nhưng tạo hiệu ứng tích lũy bảo mật. Phá vỡ 1 lớp không ảnh hưởng các lớp còn lại. |
| 4 | **Zero Attack Surface Paradigm** | Thay vì giảm bề mặt tấn công (reduce attack surface), hệ thống **triệt tiêu** bề mặt tấn công mạng (eliminate network attack surface). |
| 5 | **FAPI 2.0 + SDP Convergence** | Lần đầu tích hợp chuẩn FAPI 2.0 (Financial-grade API) với CSA Software-Defined Perimeter v3 trong cùng một kiến trúc. |

## 1.5 Tính ứng dụng thực tế

### Ngành Fintech / Ngân hàng số
- Open Banking Platform (PSD2/PSD3 compliance).
- Core Banking API Gateway.
- Payment API cho ví điện tử.

### Ngành Y tế
- Hệ thống FHIR API cho trao đổi hồ sơ bệnh án.
- Yêu cầu bảo mật HIPAA.

### Chính phủ số
- API liên thông dữ liệu giữa các cơ quan.
- Yêu cầu bảo mật cấp quốc gia.

## 1.6 Khả năng mở rộng

```
                    ┌─────────────────────┐
                    │  Current Scope      │
                    │  (This Project)     │
                    │                     │
                    │  • Single Region    │
                    │  • Lab Environment  │
                    │  • 2-3 Services     │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Phase 2 (Future)   │
                    │                     │
                    │  • Multi-Region     │
                    │  • Ziti Mesh        │
                    │  • 10+ Services     │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Phase 3 (Vision)   │
                    │                     │
                    │  • Multi-Cloud      │
                    │  • Edge Computing   │
                    │  • 100+ Services    │
                    │  • AI Threat Detect │
                    └─────────────────────┘
```

Kiến trúc được thiết kế theo nguyên lý **modular** và **loosely-coupled**:
- Mỗi lớp phòng thủ hoạt động độc lập → có thể thay thế hoặc nâng cấp riêng.
- OpenZiti Fabric hỗ trợ mở rộng từ single-node đến multi-region mesh.
- Tách biệt Identity Provider, Gateway, Database → scale theo từng thành phần.

---

> **Next:** [PART 2 — Problem Statement](./02_PROBLEM_STATEMENT.md)
