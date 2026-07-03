# PART 5 — SECURITY THREAT MODELING

## 5.1 Overview
Tài liệu này đặc tả mô hình phân tích mối đe dọa (Threat Modeling) cho hệ thống **secure-fapi-zta-darkservices** sử dụng phương pháp luận STRIDE để nhận diện, đánh giá và đề xuất phương án phòng ngự chi tiết cho từng cấu phần hạ tầng.

---

## 5.2 Ranh giới tin cậy (Trust Boundaries)

Hệ thống được phân vùng thành 4 Ranh giới tin cậy độc lập, phân tách bởi các rào cản mật mã học:

```
[ Internet (Vùng Không Tin Cậy - Zone 0) ]
                  X  <--- Không có đường dẫn IP trực tiếp (Dark Service)
================= boundary: OpenZiti Enrollment & Identity =================
[ OpenZiti Overlay Network (Vùng Xác Thực Mạng - Zone 1) ]
                  │  <--- Chỉ Ziti Identity được Dial
================= boundary: mTLS Client Certificate =======================
[ Internal Application Zone (Vùng Xác Thực Ứng Dụng - Zone 2) ]
                  │  <--- DPoP Token + RBAC validation
================= boundary: Database Connection & RLS context =============
[ Database Isolation Zone (Vùng Dữ Liệu Cô Lập - Zone 3) ]
```

---

## 5.3 Phân tích STRIDE cho từng cấu phần (Component-level STRIDE)

### 5.3.1 Client App
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Kẻ tấn công giả mạo thiết bị Client hợp lệ để gửi request. | Yêu cầu Ziti Identity hợp lệ + Chứng chỉ mTLS do CA nội bộ cấp + DPoP Key. |
| **T** — Tampering | Có | Kẻ tấn công sửa đổi cấu hình Client (file config Ziti `.json` hoặc Private Key) trên thiết bị. | Đề xuất lưu trữ Private Key trong phần cứng an toàn (TPM, Keychain, Secure Enclave). |
| **R** — Repudiation | Không | Client phủ nhận giao dịch đã thực hiện. | Mỗi giao dịch đều được ký bằng DPoP Proof chứa timestamp, URI và token hash độc bản. |
| **I** — Info Disclosure | Có | Rò rỉ Access Token hoặc Refresh Token lưu trên thiết bị. | Access Token chỉ có hạn dùng 60 giây. Refresh Token áp dụng cơ chế quay vòng một lần (Rotation). |
| **D** — Denial of Service | Không | Client bị lợi dụng để spam request làm sập hệ thống. | Triển khai Rate Limiting tại lớp API Gateway (nội bộ mạng overlay). |
| **E** — EoP (Elevation of Privilege) | Có | Client viewer cố gắng thực hiện hành động chuyển tiền (operator). | Phân quyền RBAC nghiêm ngặt bằng JWT claims được ký bởi IdP, xác thực ở phía Gateway. |

---

### 5.3.2 Identity Provider (IdP)
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Kẻ tấn công dựng một IdP giả mạo để cấp token. | API Gateway chỉ tin tưởng JWT được ký bởi khóa ES256 khớp với danh sách khóa tại JWKS của IdP thật. |
| **T** — Tampering | Có | Đánh cắp khóa ký ES256 của IdP (`idp-private.key`). | Khóa riêng tư của IdP được phân quyền ở mức hệ điều hành tối giản (`0600`), chỉ tiến trình IdP có quyền đọc. |
| **R** — Repudiation | Không | IdP phủ nhận việc cấp phát token. | IdP ghi log chi tiết các đợt cấp phát token vào syslog hệ thống kèm mã JTI. |
| **I** — Info Disclosure | Có | Rò rỉ mã Auth Code hoặc PKCE Verifier trong bộ nhớ. | RAMStore xóa ngay lập tức Auth Code sau khi đổi token thành công (One-time-use). |
| **D** — Denial of Service | Có | Kẻ tấn công spam API `/token` hoặc `/authorize` để làm tràn RAM. | RAMStore tự động giải phóng (cleanup ticker) các bản ghi hết hạn mỗi 1 phút. |
| **E** — EoP | Có | Kẻ tấn công bypass xác thực PKCE để đổi token trái phép. | Bắt buộc kiểm tra PKCE phương thức mã hóa một chiều `S256` trên mọi request đổi token. |

---

### 5.3.3 API Gateway (Dark Service)
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Kẻ tấn công giả mạo làm Gateway để đón traffic của Client. | Chỉ duy nhất danh tính mạng được phân quyền chính sách `Bind` trên Controller mới có thể đại diện dịch vụ. |
| **T** — Tampering | Có | Sửa đổi mã hash audit log hoặc dữ liệu giao dịch trong quá trình định tuyến. | Toàn bộ đường truyền qua OpenZiti Fabric được mã hóa E2E bằng thuật toán AES-256-GCM. |
| **R** — Repudiation | Có | Kẻ tấn công cố gắng xóa lịch sử giao dịch trên Gateway. | Chuyển tiếp log ngay lập tức về WORM Database, Gateway không lưu trữ trạng thái. |
| **I** — Info Disclosure | Có | Rò rỉ thông tin giao dịch của Tenant này sang Tenant khác tại Gateway. | Triển khai RLS Context Middleware, trích xuất Tenant ID từ JWT claims và bắt buộc `SET LOCAL app.tenant_id`. |
| **D** — Denial of Service | Không | Tấn công DDoS làm sập Gateway từ Internet. | **Triệt tiêu hoàn toàn bề mặt tấn công:** Gateway không mở bất kỳ cổng TCP/UDP inbound nào trên Internet. |
| **E** — EoP | Có | bypass xác thực mTLS hoặc DPoP để chiếm quyền quản trị. | Bộ lọc chuỗi xác thực liên tục (Continuous Verification) yêu cầu đồng thời cả Ziti, mTLS cert và DPoP key. |

---

### 5.3.4 OpenZiti Controller & Routers
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Giả mạo Router để tham gia mạng overlay. | Mọi Router tham gia fabric phải được enroll qua Controller bằng cơ chế mật mã học mTLS PKI riêng của Ziti. |
| **T** — Tampering | Có | Sửa đổi chính sách mạng (Service Policies) trên Controller. | Đóng cổng quản trị Controller ra Internet, chỉ cho phép quản trị qua ZAC Console nội bộ hoặc VPN bảo mật. |
| **R** — Repudiation | Không | Không ghi vết hoạt động quản trị mạng. | Bật tính năng Audit Logging mặc định của Ziti Controller. |
| **I** — Info Disclosure | Có | Đánh cắp các tệp tin danh tính `.json` lưu trên máy chủ. | Sử dụng biến môi trường Docker bảo mật và phân quyền file hệ thống tối đa. |
| **D** — Denial of Service | Có | Tấn công làm tràn băng thông của Edge Routers. | Sử dụng cơ chế mesh routing, phân tán tải sang nhiều Router khác nhau. |
| **E** — EoP | Có | bypass phân quyền chính sách để kết nối chéo dịch vụ. | Triển khai chính sách Ziti Zero Trust cụ thể đến từng cặp Identity-to-Service (không dùng wildcard). |

---

### 5.3.5 PostgreSQL Database
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Ứng dụng kết nối bằng tài khoản `postgres` (Superuser) bypass RLS. | Bắt buộc kết nối bằng role hạn chế quyền `app_user`. Không sử dụng tài khoản superuser trong ứng dụng. |
| **T** — Tampering | Có | Quản trị viên cơ sở dữ liệu thay đổi số dư hoặc xóa nhật ký kiểm toán. | Ràng buộc trigger WORM chặn đứng lệnh `UPDATE` và `DELETE` trên bảng `audit_logs` ở tầng nhân DB. |
| **R** — Repudiation | Có | Thay đổi thông tin log mà không bị phát hiện. | Hàm trigger tự động tính toán **SHA-256 Hash-chain** liên kết chặt chẽ dòng sau với dòng trước. |
| **I** — Info Disclosure | Có | Lỗi SQL Injection làm rò rỉ dữ liệu chéo Tenant. | Bật RLS cứng trên DB làm chốt chặn cuối cùng. SQL Injection ở app layer vẫn bị DB chặn nếu không có Tenant Context đúng. |
| **D** — Denial of Service | Có | Làm nghẽn hàng đợi kết nối DB. | Cấu hình Connection Pool giới hạn tối đa kết nối và tài nguyên sử dụng cho `app_user`. |
| **E** — EoP | Có | Chiếm quyền quản trị cơ sở dữ liệu. | Đóng cổng DB `5432` ra Internet công cộng, chỉ lắng nghe trên `127.0.0.1` phục vụ local dev/test. |

---

### 5.3.6 Internal PKI (Certificate Authority)
| Mối đe dọa (STRIDE) | Khả thi | Chi tiết mối đe dọa | Phương án giảm thiểu (Mitigations) |
|---|---|---|---|
| **S** — Spoofing | Có | Kẻ tấn công tự cấp chứng chỉ giả từ CA. | Khóa riêng tư của Root CA (`root-ca.key`) phải được lưu trữ hoàn toàn offline. |
| **T** — Tampering | Có | Sửa đổi tệp CRL (Certificate Revocation List) để hợp lệ hóa cert đã thu hồi. | CRL phải được ký số bởi CA và lưu trữ trên vùng nhớ chỉ đọc của Gateway. |
| **R** — Repudiation | Không | CA không ghi nhận lịch sử cấp phát. | Sử dụng tệp serial `.srl` tăng dần để lưu vết số lượng cert đã cấp. |
| **I** — Info Disclosure | Có | Rò rỉ Intermediate CA key (`intermediate-ca.key`). | Lưu trữ Intermediate CA trong môi trường cô lập, phân quyền hệ thống tối đa. |
| **D** — Denial of Service | Không | Tấn công làm sập CA. | CA hoạt động theo cơ chế offline/on-demand, không cần chạy online liên tục để xác thực kết nối. |
| **E** — EoP | Có | bypass quyền cấp chứng chỉ để chiếm quyền hệ thống. | Giới hạn quyền sử dụng script cấp phát chứng chỉ. |

---

## 5.4 Ma trận Mối đe dọa & Biện pháp phòng ngự tổng hợp

```
                    ┌─────────────────────────┐
                    │     Client Device       │
                    │                         │
                    │  [S] -> mTLS + DPoP     │
                    │  [I] -> Token Rotation  │
                    └──────────┬──────────────┘
                               │
               ================┴================ boundary: OpenZiti (Zero Ports)
                               │
                    ┌──────────▼──────────┐
                    │    Ziti Routers     │
                    │                     │
                    │  [D] -> Dark Service│
                    │  [T] -> AES-256 E2E │
                    └──────────┬──────────┘
                               │
               ================┴================ boundary: mTLS Handshake Validation
                               │
                    ┌──────────▼──────────┐
                    │   API Gateway Go    │
                    │                     │
                    │  [S] -> ES256 DPoP  │
                    │  [E] -> RBAC Check  │
                    └──────────┬──────────┘
                               │
               ================┴================ boundary: DB Role & Tenant Context
                               │
                    ┌──────────▼──────────┐
                    │     PostgreSQL      │
                    │                     │
                    │  [I] -> Row-Sec RLS  │
                    │  [T] -> WORM Trigger │
                    │  [R] -> SHA-256 Chain│
                    └─────────────────────┘
```
