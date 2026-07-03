# Secure FAPI-ZTA & Mạng Dark Services (Tàng hình API)

[![Go Version](https://img.shields.io/badge/Go-1.25.0-blue.svg)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://www.docker.com/)
[![Security Profile](https://img.shields.io/badge/Security-FAPI%202.0%20%2F%20ZTA-red.svg)](https://openid.net/sg/fapi/)

> **Kiến trúc bảo mật giao dịch tài chính cấp độ ngân hàng (Financial-grade API) kết hợp mạng ảo tàng hình OpenZiti, phân tách đa khách thuê Row-Level Security, và nhật ký bất biến mật mã học WORM Ledger.**

---

## Tổng quan Hệ thống

Dự án này hiện thực hóa mô hình bảo mật Zero Trust Architecture (NIST SP 800-207) thông qua sự kết hợp của 3 tầng bảo vệ chặt chẽ:
1. **Lớp Mạng (Network Layer - Dark Services):** API Gateway được giấu kín hoàn toàn khỏi Internet công cộng thông qua OpenZiti Overlay. Gateway không mở bất kỳ cổng TCP inbound nào ra ngoài (0 open ports), ngăn chặn tuyệt đối mọi hành vi rà quét mạng (port scanning).
2. **Lớp Ứng dụng (Application Layer - FAPI 2.0):** Xác thực kép bằng mTLS X.509 ở lớp truyền dẫn và cơ chế sinh khóa ký số DPoP (RFC 9449) ràng buộc Token với thiết bị ở lớp ứng dụng. Triển khai cơ chế Ràng buộc chéo (Cross-Layer Binding): API Gateway kiểm tra đối chiếu danh tính mTLS mạng ảo (SourceIdentifier) khớp chính xác với sub claim trong DPoP Token để chống tấn công đánh cắp Token chéo thiết bị.
3. **Lớp Dữ liệu (Data Layer - Database Security):** Sử dụng cơ chế phân tách đa thuê PostgreSQL Row-Level Security (RLS) ở mức cứng database bằng context set_config. Nhật ký kiểm toán bảo mật WORM (Write Once, Read Many) được bảo vệ bằng trigger ngăn chặn tuyệt đối lệnh UPDATE/DELETE và tính toán băm liên kết chuỗi khối SHA-256 Hash-chaining bảo đảm tính bất biến toàn vẹn.

---

## Đặc tính Kiến trúc và Điểm Độc Bản (USPs)

* **Ràng Buộc Chéo Lớp Mạng & Ứng Dụng (Cross-Layer Binding):** Gateway liên kết cứng chứng chỉ client mTLS của OpenZiti với chữ ký DPoP JWT để triệt tiêu lỗ hổng đánh cắp token của thiết bị hợp pháp.
* **Sổ cái WORM Hash-chain Đa Thuê:** RLS không chỉ cô lập dữ liệu giao dịch mà còn cô lập cả chuỗi liên kết log kiểm toán SHA-256. Mỗi khách hàng (Tenant) sở hữu một chuỗi log mã hóa độc lập và bất biến.
* **Không phụ thuộc API Gateway thương mại:** Toàn bộ lõi xác thực DPoP, PKCE, mTLS, và liên kết Ziti SDK được viết trực tiếp bằng Go thuần, tối ưu hóa hiệu năng và thu hẹp bề mặt tấn công (attack surface).

---

## Cấu trúc Thư mục Dự án

```
secure-fapi-zta-darkservices/
│
├── docker/                             # Cấu hình hạ tầng
│   ├── docker-compose.yml              # Ziti Controller + Routers + PostgreSQL
│   ├── .env                            # Biến môi trường Ziti
│   └── postgres/
│       ├── Dockerfile                  # Custom PostgreSQL Image
│       └── init.sql                    # Schema DB + RLS Policies + WORM Triggers
│
├── certs/                              # Quản lý PKI nội bộ
│   └── scripts/                        # Script tự động sinh CA & Certs (ECC P-256)
│
├── idp/                                # Identity Provider (Go Module)
│   ├── main.go                         # Điểm chạy IdP phục vụ cấp token
│   ├── handler/                        # OIDC Discovery, JWKS, Authorize, Token
│   └── crypto/                         # Xác thực PKCE & DPoP Proof
│
├── gateway/                            # API Gateway tàng hình (Go Module)
│   ├── main.go                         # Lắng nghe ẩn qua OpenZiti SDK
│   └── internal/
│       ├── api/                        # Handlers (balance, transfer, audit-logs)
│       ├── audit/                      # DB Client + RLS Context Injection
│       ├── auth/                       # JWKS Cache + DPoP Proof Verify
│       └── middleware/                 # Cross-layer Auth + RBAC Middleware
│
├── client/                             # Client CLI Application (Go Module)
│   ├── main.go                         # Giao diện gọi API qua mạng tàng hình
│   ├── crypto/                         # Sinh khóa DPoP & ký proof
│   └── ziti/                           # OpenZiti Client Dialer
│
├── scripts/                            # Kịch bản tự động hóa cấu hình Ziti
│   ├── setup-ziti-services.sh          # Tạo Ziti Services, Identities & Policies
│   └── enroll-identities.sh            # Ghi danh lấy cấu hình kết nối JSON
│
└── CHANGELOG.md                        # Nhật ký cập nhật tiến độ dự án
```

---

## Hướng dẫn Cài đặt & Khởi chạy (Chế độ Local Debug)

### Cách 1: Khởi chạy nhanh toàn bộ dự án bằng 1 câu lệnh duy nhất
Dự án cung cấp sẵn tập lệnh PowerShell để khởi chạy đồng thời cụm Docker, Identity Provider (IdP), API Gateway và Dashboard Next.js chỉ bằng 1 câu lệnh:
```powershell
./run-all.ps1
```
Tập lệnh sẽ tự động khởi tạo các tài nguyên và mở các cửa sổ PowerShell riêng biệt chạy nền cho từng tiến trình, giúp dễ dàng theo dõi logs và quản trị.

### Cách 2: Khởi chạy thủ công từng cấu phần
Nếu muốn chạy thủ công từng thành phần, hãy làm theo các bước dưới đây:

#### Bước 1: Khởi chạy hạ tầng Docker
Yêu cầu máy cài sẵn Docker và Docker Compose. Chạy lệnh sau để bật Ziti và PostgreSQL:
```bash
docker compose -f docker/docker-compose.yml up -d
```
Đảm bảo tất cả các container đều chuyển sang trạng thái healthy.

### Bước 2: Chạy Identity Provider (IdP)
Mở terminal mới và khởi chạy máy chủ cấp Token (lắng nghe trên cổng 8081):
```bash
$env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path
cd idp
go run main.go
```

### Bước 3: Chạy API Gateway (Chế độ Debug)
Mở terminal thứ hai, tắt chế độ Ziti Overlay để chạy thử nghiệm kết nối TCP local (lắng nghe trên cổng 8080):
```bash
$env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path
cd gateway
$env:USE_ZITI = "false"
go run main.go
```

### Bước 4: Chạy Client thực thi nghiệp vụ
Mở terminal thứ ba và sử dụng Client CLI để tương tác:

* **Truy vấn số dư của Alice (Tenant A):**
    ```bash
    $env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path
    cd client
    go run main.go -identity client-alice -cmd balance -ziti=false
    ```

* **Thực hiện giao dịch chuyển tiền cho Alice:**
    ```bash
    go run main.go -identity client-alice -cmd transfer -amount 1500 -desc "Alice nop hoc phi PTIT" -ziti=false
    ```

* **Xem lịch sử Log mã hóa WORM của Alice:**
    ```bash
    go run main.go -identity client-alice -cmd logs -ziti=false
    ```

---

## Kịch bản Thử nghiệm và Xác minh Bảo mật

### 1. Kiểm thử Phân tách Dữ liệu (Row-Level Security)
Đóng vai Bob (Tenant B) truy vấn số dư và xem logs giao dịch:
```bash
go run main.go -identity client-bob -cmd balance -ziti=false
```
Kết quả: Số dư của Bob hiển thị là 0. Bob hoàn toàn không nhìn thấy số tiền 1500 của Alice và không đọc được bất kỳ bản ghi log nào của Alice. Cơ chế RLS hoạt động hoàn hảo.

### 2. Kiểm thử Tính Bất Biến của Log (Database WORM Test)
Thử thực thi lệnh SQL sửa đổi dữ liệu trực tiếp trong Postgres container để xóa dấu vết:
```bash
docker exec -t docker-postgresql-1 psql -U postgres -d fapi_db -c "UPDATE audit_logs SET action = 'HACKED' WHERE id = 2;"
```
Kết quả: Trình quản lý database ném lỗi:
`ERROR: Audit logs are immutable (WORM)`

---

## Giao diện Giám sát An ninh (SOC & Telemetry Dashboard)

Dự án hỗ trợ 2 mô hình giám sát an ninh mạng thời gian thực:

### 1. Dashboard Next.js Tự Xây (Custom-Built Cyber SOC)
Một giao diện trạm điều hành an ninh mạng độc lập, được thiết kế theo phong cách tối giản kính mờ (Glassmorphism), tự động kết nối trực tiếp đến Go Gateway và cơ sở dữ liệu Postgres để hiển thị:
- **Zero Trust Topology Map:** Sơ đồ mạng tàng hình OpenZiti động, hiển thị trạng thái kết nối các nodes và chuyển màu đỏ cảnh báo nếu phát hiện tấn công.
- **WORM Ledger Log Stream:** Terminal hiển thị các log bất biến và chuỗi băm Hash-chain (block_hash & prev_hash) trực tiếp từ database.
- **Crypto Latency Radar:** Đồ thị Recharts biểu diễn độ trễ xử lý mật mã của DPoP (signature verify) và database context set.

**Cách chạy:**
1. Di chuyển vào thư mục dashboard:
   ```bash
   cd dashboard
   ```
2. Khởi chạy máy chủ phát triển ở cổng 3001 (tránh đụng độ cổng Grafana):
   ```bash
   npm run dev -- -p 3001
   ```
3. Truy cập trình duyệt tại: http://localhost:3001

### 2. Cụm Telemetry Tiêu Chuẩn Công Nghiệp (Prometheus + Grafana + Loki)
Đã được tích hợp vào cụm Docker Compose để tự động nạp sẵn (provisioned) nguồn dữ liệu và SOC dashboard:
- **Bảng 1 (Zero Trust Network Map):** Trạng thái kết nối các mTLS tunnels.
- **Bảng 2 (DPoP & Loki Security Logs):** Biểu đồ phát hiện tấn công Replay và mismatch token.
- **Bảng 3 (Crypto Performance):** Biểu đồ Prometheus ghi nhận Gateway requests và DB Latency.

**Cách chạy:**
1. Khởi chạy toàn bộ hạ tầng (Ziti + Postgres + Prometheus + Grafana + Loki + Promtail):
   ```bash
   cd docker
   docker compose up -d
   ```
2. Truy cập Grafana tại: http://localhost:3000 (Tài khoản: admin / admin).
3. Vào mục Dashboards -> Chọn Secure FAPI-ZTA SOC & Telemetry Dashboard.

---

## Tài liệu Đặc tả Kỹ thuật

Hồ sơ thiết kế chi tiết nằm tại thư mục docs/00_MASTER_INDEX.md:
* [docs/security/threat-model.md](./docs/security/threat-model.md): Phân tích STRIDE Threat Modeling chi tiết.
* [docs/adr/](./docs/adr/): Các quyết định kiến trúc cốt lõi (OpenZiti, DPoP, Go, Postgres WORM).
* [docs/diagrams/sequence_flows.md](./docs/diagrams/sequence_flows.md): Sơ đồ tuần tự các luồng đăng ký thiết bị và giao dịch.
* [docs/15_VALIDATION_BENCHMARK.md](./docs/15_VALIDATION_BENCHMARK.md): Các ca kiểm thử bảo mật & kế hoạch đo lường hiệu năng.

---

## Giấy phép

Mã nguồn phát hành dưới giấy phép MIT License.
