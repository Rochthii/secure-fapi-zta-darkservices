# ADR-001: Lựa chọn OpenZiti làm hạ tầng mạng Zero Trust Overlay

*   **Status:** APPROVED
*   **Date:** 2026-07-03
*   **Deciders:** Principal Security Architect / Zero Trust Architect

---

## 1. Bối cảnh (Context)
Các API Gateway truyền thống yêu cầu mở cổng inbound công khai (như `443` cho HTTPS) để nhận traffic từ Client. Điều này dẫn đến các rủi ro bảo mật:
- Kẻ tấn công có thể rà quét cổng (port scan), dò tìm endpoint ẩn (API discovery).
- Dễ bị tấn công từ chối dịch vụ (DDoS) do các cổng kết nối luôn mở rộng.
- Nguy cơ khai thác lỗ hổng Zero-day trên máy chủ API trước khi xác thực.

Các giải pháp thay thế đã được đánh giá:
1.  **VPN (Virtual Private Network):** Truy cập theo kiểu all-or-nothing (sau khi vào VPN sẽ thấy toàn bộ mạng nội bộ), IP vẫn bị lộ, quản lý certificate thủ công phức tạp.
2.  **Istio Service Mesh:** Bắt buộc phải chạy trong Kubernetes, không hỗ trợ che giấu dịch vụ ra ngoài Internet công cộng, vẫn cần public ingress gateway.
3.  **Tailscale / WireGuard:** Phụ thuộc vào server điều phối bên ngoài (SaaS), không hỗ trợ nhúng SDK native trực tiếp vào code ứng dụng (SDK-bound).

---

## 2. Quyết định (Decision)
Lựa chọn **OpenZiti** làm hạ tầng mạng Zero Trust Overlay cho dự án.
- Tận dụng cơ chế **Dark Services** bằng cách sử dụng Ziti SDK nhúng trực tiếp vào API Gateway và Client App.
- Thay vì gọi `net.Listen("tcp", ":port")`, Gateway sẽ gọi `ctx.Listen("service-name")` qua Ziti SDK để bind dịch vụ.
- Đóng hoàn toàn tất cả các cổng TCP/UDP inbound trên tường lửa của máy chủ Gateway (Zero Inbound Ports). Mọi kết nối chuyển tiếp qua Ziti Edge Router đều là kết nối outbound.

---

## 3. Hệ quả (Consequences)

### Điểm tốt (Pros):
- Triệt tiêu hoàn toàn bề mặt tấn công mạng (Network Attack Surface = 0). Quét `nmap` Gateway trả về 0 cổng mở.
- Ngăn chặn triệt để tấn công DDoS và API Scanning từ Internet.
- Phân quyền kết nối chi tiết đến từng danh tính (Identity-first) thay vì phân quyền theo địa chỉ IP/Subnet.
- Mã hóa E2E tự động toàn bộ traffic qua mạng overlay bằng thuật toán AES-256-GCM.

### Điểm xấu (Cons):
- Tăng độ phức tạp của mã nguồn do phải tích hợp SDK OpenZiti vào cả Client LẪN Gateway.
- Phát sinh độ trễ mạng (Network Latency Overhead) khoảng 10-25ms do traffic phải chuyển tiếp qua Ziti Edge Routers.
- Cộng đồng phát triển OpenZiti nhỏ hơn so với các giải pháp VPN truyền thống.
