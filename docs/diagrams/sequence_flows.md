# PART 17 — SEQUENCE DIAGRAMS

Tài liệu này cung cấp các sơ đồ tuần tự (Sequence Diagrams) trực quan mô tả chi tiết các luồng nghiệp vụ bảo mật cốt lõi của hệ thống **secure-fapi-zta-darkservices**.

---

## 17.1 Luồng đăng nhập và cấp phát Token ràng buộc (Login & Token Exchange Flow)

Sơ đồ này mô tả chi tiết các bước từ khi Client khám phá cấu hình IdP, thực hiện luồng PKCE và nhận về DPoP-bound Access Token (đáp ứng tiêu chuẩn FAPI 2.0).

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client Device
    participant IdP as Identity Provider

    Note over Client, IdP: Khám phá dịch vụ & Cài đặt ban đầu
    Client->>IdP: GET /.well-known/openid-configuration
    IdP-->>Client: OIDC Discovery Document (ES256, S256, DPoP supported)
    Client->>IdP: GET /.well-known/jwks.json
    IdP-->>Client: JWKS Document (Public Key của IdP)

    Note over Client: Chuẩn bị PKCE (RFC 7636)
    Client->>Client: Sinh ngẫu nhiên code_verifier
    Client->>Client: Tính code_challenge = Base64URL(SHA256(code_verifier))

    Note over Client, IdP: Luồng Ủy quyền (Authorization Request)
    Client->>IdP: GET /authorize?response_type=code&client_id=...&code_challenge=...&code_challenge_method=S256
    IdP->>IdP: Kiểm tra tham số PKCE (Chỉ chấp nhận S256)
    IdP->>IdP: Lưu code_challenge tương ứng với code vào RAMStore (Hạn dùng 5 phút)
    IdP-->>Client: Trả về authorization_code (One-time-use)

    Note over Client: Chuẩn bị DPoP (RFC 9449)
    Client->>Client: Sinh cặp khóa thiết bị ECC P-256 (DPoP Keypair)
    Client->>Client: Sinh DPoP Proof JWT chứa: htm=POST, htu=/token, jti=ngẫu_nhiên
    Client->>Client: Ký DPoP Proof bằng Private Key của Client

    Note over Client, IdP: Luồng đổi Token (Token Request)
    Client->>IdP: POST /token (grant_type=authorization_code, code, code_verifier)<br/>Header DPoP: <DPoP Proof JWT>
    
    IdP->>IdP: Xác thực DPoP: Giải mã JWK ở header, verify chữ ký Proof JWT
    IdP->>IdP: Xác thực claim: htm == "POST", htu == "/token", iat trong hạn 60s
    IdP->>IdP: Chống Replay: Kiểm tra jti có bị trùng trong RAM Cache?
    IdP->>IdP: Xác thực PKCE: So sánh SHA256(code_verifier) == code_challenge trong RAM
    
    IdP->>IdP: Tính toán JWK Thumbprint (RFC 7638) của Client Key = cnf.jkt
    IdP->>IdP: Sinh Access Token chứa cnf.jkt, ký bằng Private Key của IdP (ES256)
    
    IdP-->>Client: Trả về JSON: access_token (TTL 60s, token_type=DPoP), refresh_token
```

---

## 17.2 Luồng truy cập dịch vụ tàng hình (Service Access Flow)

Sơ đồ này mô tả chi tiết cách thức Client vượt qua lớp mạng ảo tàng hình OpenZiti, xác thực mTLS và gửi request kèm DPoP-bound Access Token để Gateway thực thi truy vấn Row-Level Security (RLS) an toàn.

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client Device
    participant Ziti as OpenZiti Router
    participant Gateway as API Gateway (Dark Service)
    participant PDP as Policy Engine (PDP)
    participant DB as PostgreSQL

    Note over Client, Ziti: Lớp 1: Thiết lập mạng ẩn (Zero Trust Network)
    Client->>Ziti: Dial "financial-ledger-service" (outbound connection)
    Ziti->>Ziti: Xác thực Ziti Identity bằng mTLS mạng
    Ziti->>Gateway: Thiết lập kênh kết nối an toàn (E2E Encrypted AES-256-GCM)

    Note over Client: Lớp 2: Chuẩn bị Request tầng ứng dụng
    Client->>Client: Tính toán ath = Base64URL(SHA256(access_token))
    Client->>Client: Sinh DPoP Proof JWT mới chứa: htm=POST, htu=/api/transfer, ath
    Client->>Client: Ký DPoP Proof bằng Private Key của Client

    Note over Client, Gateway: Lớp 3: Gọi API & Xác thực liên tục (Continuous Verification)
    Client->>Gateway: POST /api/transfer<br/>Header Authorization: DPoP <access_token><br/>Header DPoP: <DPoP Proof JWT>
    
    Gateway->>Gateway: mTLS check: Xác thực chứng chỉ client cert từ Ziti connection
    Gateway->>Gateway: DPoP check: Verify chữ ký Proof JWT bằng public key đính kèm
    Gateway->>Gateway: Token check: Verify chữ ký Token bằng JWKS của IdP, verify exp
    Gateway->>Gateway: Binding check: So sánh cnf.jkt trong token == thumbprint public key của proof
    Gateway->>Gateway: Proof matching: So sánh ath trong proof == hash của access token gửi lên
    Gateway->>Gateway: Replay check: Kiểm tra jti của proof có bị trùng?
    
    Note over Gateway, PDP: Gọi PDP qua Ziti Overlay (gRPC JSON Codec)
    Gateway->>PDP: CheckAccess(tenant_id, subject, action, resource, context)
    PDP->>PDP: Tra cứu Trie O(log N) + Đánh giá AST các thuộc tính ABAC
    PDP-->>Gateway: Trả về ALLOW / DENY (avg 0.038ms)

    Note over Gateway, DB: Lớp 4: Truy cập cơ sở dữ liệu & WORM Audit
    Gateway->>DB: Thiết lập kết nối bằng role app_user (Non-superuser)
    Gateway->>DB: Chạy câu lệnh: SET LOCAL app.tenant_id = 'tenant-xyz'
    Gateway->>DB: Chạy truy vấn: INSERT INTO transactions ...
    
    DB->>DB: RLS check: DB tự động lọc chỉ cho phép Tenant ID khớp với context
    DB->>DB: Trigger audit: Tự động ghi chép log vào audit_logs
    DB->>DB: Trigger WORM: Chặn tuyệt đối UPDATE/DELETE
    DB->>DB: Trigger Hash-chain: Tính SHA-256(data + prev_hash) làm block_hash
    
    DB-->>Gateway: Trả về kết quả giao dịch
    Gateway-->>Client: Trả về 200 OK + Giao dịch thành công

```

---

## 17.3 Luồng ghi danh thiết bị Client (Client Identity Enrollment Flow)

Sơ đồ này mô tả cách Ziti Controller cấp phát chứng chỉ số X.509 an toàn cho một danh tính mạng mới thông qua quy trình ghi danh một lần.

```mermaid
sequenceDiagram
    autonumber
    participant Admin as System Administrator
    participant Ctrl as Ziti Controller
    participant Client as Client Device

    Admin->>Ctrl: Đăng ký Identity "client-alice"
    Ctrl->>Ctrl: Sinh mã enrollment token dùng 1 lần (JWT)
    Ctrl-->>Admin: Trả về file client-alice.jwt
    Admin->>Client: Phân phối tệp client-alice.jwt an toàn tới thiết bị

    Note over Client: Thực hiện Ghi danh (Enrollment)
    Client->>Client: Khởi tạo tiến trình: ziti edge enroll -j client-alice.jwt
    Client->>Client: Tự động sinh cặp khóa riêng tư (Private Key) trên thiết bị
    Client->>Client: Tạo Certificate Signing Request (CSR) chứa Public Key

    Client->>Ctrl: Gửi CSR kèm enrollment JWT lên Controller
    Ctrl->>Ctrl: Xác thực mã JWT hợp lệ và chưa từng sử dụng
    Ctrl->>Ctrl: Ký duyệt CSR bằng Ziti CA nội bộ
    Ctrl-->>Client: Trả về Chứng chỉ số X.509 hợp lệ

    Note over Client: Hoàn thành thiết lập thiết bị
    Client->>Client: Lưu trữ Private Key & Cert X.509
    Client->>Client: Sinh file cấu hình mạng chính thức: client-alice.json
```
