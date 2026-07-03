# PART 2 — PROBLEM STATEMENT

## 2.1 Vì sao JWT + RLS là chưa đủ

### Hạn chế cố hữu của JWT Bearer Tokens

JWT (JSON Web Token) theo RFC 7519 là tiêu chuẩn xác thực phổ biến nhất hiện nay. Tuy nhiên, JWT Bearer Token có **lỗ hổng thiết kế cơ bản**:

| Vấn đề | Mô tả | Hậu quả |
|---|---|---|
| **Bearer = Người mang = Người sở hữu** | Bất kỳ ai "mang" (possess) token đều được coi là chủ sở hữu hợp lệ. Không có cơ chế chứng minh quyền sở hữu. | Token bị đánh cắp → Kẻ tấn công có toàn quyền truy cập. |
| **Stateless = Không thể thu hồi** | JWT stateless không thể bị thu hồi trước khi hết hạn (trừ khi dùng blacklist — phá vỡ tính stateless). | Token bị lộ → Phải chờ hết hạn hoặc xây dựng hạ tầng thu hồi phức tạp. |
| **Payload minh bạch** | JWT payload chỉ Base64-encoded, không mã hóa. Ai cũng đọc được nội dung. | Rò rỉ thông tin nhạy cảm (user ID, tenant ID, roles) nếu token bị chặn bắt. |
| **Không ràng buộc thiết bị** | Token không gắn với thiết bị hay phiên cụ thể. Có thể sử dụng trên bất kỳ máy nào. | Copy token từ máy A → dùng trên máy B → hoạt động bình thường. |

### Hạn chế của RLS (Row-Level Security)

RLS cô lập dữ liệu ở tầng database — đây là tuyến phòng thủ cuối cùng. Nhưng RLS **chỉ là một lớp duy nhất** trong kiến trúc phòng thủ:

| Lỗ hổng | Mô tả |
|---|---|
| **Phụ thuộc context injection** | RLS hoạt động dựa trên `current_setting()`. Nếu middleware quên SET LOCAL → RLS bất lực. |
| **Không bảo vệ tầng mạng** | RLS không ngăn DDoS, port scan, hay token theft. Nó chỉ kiểm soát dữ liệu đã lọt vào database. |
| **Không có Continuous Verification** | Một khi context được set → RLS tin tưởng suốt transaction. Không có khả năng re-verify giữa chừng. |
| **Single Point of Failure** | Nếu kẻ tấn công chiếm quyền superuser PostgreSQL → bypass toàn bộ RLS. |

### Kết luận

> **JWT + RLS = Necessary but Insufficient.**
> Cần một kiến trúc phòng thủ chiều sâu (Defense-in-Depth) với nhiều lớp bảo mật độc lập, bổ trợ lẫn nhau.

---

## 2.2 Những hạn chế của API Public hiện nay

### Bề mặt tấn công (Attack Surface) của API truyền thống

```
┌─────────────────────────────────────────────────────────────────┐
│                    INTERNET (UNTRUSTED ZONE)                     │
│                                                                  │
│    Attacker ──────► Port Scan ──────► ✅ Port 443 OPEN          │
│    Attacker ──────► API Discovery ──► ✅ Endpoints visible      │
│    Attacker ──────► DDoS ──────────► ✅ Service reachable       │
│    Attacker ──────► Brute Force ───► ✅ Login endpoint exists   │
│                                                                  │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                    ┌─────▼─────┐
                    │ Firewall  │  ← Chỉ lọc IP/Port, không hiểu context
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │ API GW    │  ← Mở port 443/8080 công khai
                    │ (Public)  │  ← Bất kỳ ai trên Internet đều kết nối được
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │ Services  │  ← Trust boundary chỉ dựa trên JWT
                    └───────────┘
```

**Vấn đề cốt lõi:** API phải **mở cổng lắng nghe** (listening port) trên Internet → Kẻ tấn công biết **ở đâu** để tấn công.

### So sánh: API Public vs Dark Service

| Tiêu chí | API Public (Hiện tại) | Dark Service (Giải pháp) |
|---|---|---|
| **Port mở** | Port 443 luôn mở | 0 port mở |
| **Khả năng phát hiện** | `nmap` phát hiện ngay | Không thể phát hiện |
| **DDoS** | Dễ bị tấn công | Không có target để tấn công |
| **API Discovery** | Endpoint path có thể brute-force | Không có endpoint trên Internet |
| **Kết nối** | Bất kỳ ai | Chỉ enrolled identity |
| **Xác thực** | Sau khi kết nối (connect-then-auth) | Trước khi kết nối (auth-then-connect) |

---

## 2.3 Các kiểu tấn công hiện đại nhắm vào API

### 2.3.1 API Enumeration & Discovery

**Mô tả:** Kẻ tấn công rà quét (enumerate) các API endpoint bằng cách thử các path phổ biến (`/api/v1/users`, `/api/admin`, `/graphql`, ...) hoặc sử dụng các công cụ tự động (Burp Suite Intruder, ffuf, Gobuster).

**Tác động:** Phát hiện endpoint ẩn, admin API, debug endpoint, hoặc API version cũ chưa bị vô hiệu hóa.

**Dark Service mitigation:** Không có public endpoint → Không có gì để enumerate.

---

### 2.3.2 Credential Stuffing

**Mô tả:** Sử dụng danh sách username/password bị rò rỉ từ các vụ data breach khác để thử đăng nhập hàng loạt.

**Tác động:** Chiếm tài khoản hợp lệ, đặc biệt với người dùng tái sử dụng mật khẩu.

**FAPI-ZTA mitigation:**
- DPoP: Kể cả đăng nhập thành công, token chỉ hoạt động trên thiết bị gốc.
- mTLS: Cần chứng chỉ X.509 → không thể brute-force từ xa.
- Ziti Identity: Cần enrolled identity → không thể kết nối mạng.

---

### 2.3.3 Token Replay Attack

**Mô tả:** Kẻ tấn công chặn bắt (intercept) access token từ network traffic, logs, hoặc browser storage, sau đó sử dụng lại token đó.

**Tác động:** Giả mạo danh tính, truy cập trái phép.

**FAPI-ZTA mitigation:**
- DPoP proof có `jti` (unique ID) và `iat` (timestamp) → mỗi proof chỉ dùng 1 lần.
- Access Token Hash (`ath`) ràng buộc proof với token cụ thể.
- Server lưu JTI cache → phát hiện replay ngay lập tức.

---

### 2.3.4 Man-in-the-Middle (MITM)

**Mô tả:** Kẻ tấn công đứng giữa Client và Server, chặn bắt và/hoặc sửa đổi traffic.

**Tác động:** Đánh cắp token, sửa đổi giao dịch, tiêm mã độc.

**FAPI-ZTA mitigation:**
- mTLS: Cả Client LẪN Server phải trình diện chứng chỉ → MITM không có cert → bị phát hiện.
- OpenZiti: E2E encryption qua overlay → traffic không đi qua Internet công cộng.

---

### 2.3.5 Lateral Movement

**Mô tả:** Sau khi xâm nhập một service, kẻ tấn công di chuyển ngang (lateral) sang các service khác trong cùng mạng nội bộ.

**Tác động:** Từ một điểm xâm nhập → chiếm toàn bộ hệ thống.

**FAPI-ZTA mitigation:**
- Zero Trust: Không tin tưởng nội bộ. Mỗi service-to-service call đều phải xác thực mTLS.
- OpenZiti: Mỗi service là một "dark island" độc lập. Compromise service A ≠ truy cập service B.
- Ziti Service Policies: Kiểm soát chính xác identity nào được Dial tới service nào.

---

### 2.3.6 Session Hijacking

**Mô tả:** Chiếm phiên (session) đang hoạt động của người dùng hợp lệ.

**Tác động:** Toàn quyền truy cập với danh tính nạn nhân.

**FAPI-ZTA mitigation:**
- DPoP: Token ràng buộc với private key → không thể hijack session mà không có private key.
- Short-lived token (60s): Cửa sổ tấn công cực ngắn.
- Continuous Verification: Mỗi request đều phải ký DPoP proof mới.

---

### 2.3.7 Service Enumeration

**Mô tả:** Rà quét mạng nội bộ để phát hiện các service đang chạy (port scan nội bộ).

**Tác động:** Lập bản đồ hạ tầng, tìm service yếu.

**FAPI-ZTA mitigation:**
- Dark Services: Service không mở port trên bất kỳ network interface nào (kể cả localhost trong một số cấu hình).
- OpenZiti Overlay: Traffic chỉ đi qua Ziti fabric, không qua TCP/IP stack thông thường.

---

## 2.4 Tổng hợp: Ma trận Tấn công vs Giải pháp

| # | Kiểu tấn công | JWT+RLS | FAPI-ZTA Dark Service | Lớp phòng thủ |
|---|---|---|---|---|
| 1 | API Enumeration | ❌ Không chặn | ✅ Không có endpoint | L3 (Ziti Overlay) |
| 2 | API Discovery | ❌ Endpoint lộ | ✅ Service tàng hình | L3 (Ziti Overlay) |
| 3 | Credential Stuffing | ⚠️ Rate limit | ✅ mTLS + Ziti Identity | L2 (IdP) + L3 |
| 4 | Token Replay | ❌ Bearer reusable | ✅ DPoP jti + ath | L1 (DPoP) + L4 |
| 5 | MITM | ⚠️ TLS một chiều | ✅ mTLS + E2E overlay | L3 (mTLS) |
| 6 | Lateral Movement | ❌ Flat network | ✅ Micro-segmentation | L3 (Ziti) + L5 |
| 7 | Session Hijacking | ❌ Token theft | ✅ DPoP key binding | L1 (DPoP) |
| 8 | Service Enumeration | ❌ Port scan | ✅ Zero listening ports | L3 + L4 |

---

> **Next:** [PART 3 — Literature Review](./03_LITERATURE_REVIEW.md)
