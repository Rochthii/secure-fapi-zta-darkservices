# PART 9 — DARK SERVICES ARCHITECTURE

## 9.1 OpenZiti Fabric — Deep Analysis

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenZiti Architecture                      │
│                                                              │
│  ┌─────────────┐                                            │
│  │ Controller   │ ← Control Plane                           │
│  │              │   • Identity registry                     │
│  │              │   • Service definitions                   │
│  │              │   • Policy engine                         │
│  │              │   • PKI / CA                              │
│  └──────┬──────┘                                            │
│         │ Control channel (mTLS)                             │
│  ┌──────▼──────┐      ┌──────────────┐                      │
│  │ Edge Router │──────│ Edge Router  │ ← Data Plane         │
│  │ (Region A)  │ mesh │ (Region B)   │   • Traffic relay    │
│  └──────┬──────┘      └──────┬───────┘   • Smart routing    │
│         │                     │           • E2E encryption   │
│    ┌────▼─────┐         ┌────▼─────┐                        │
│    │ SDK App  │         │ SDK App  │ ← Application Plane    │
│    │ (Client) │         │ (Server) │   • Bind/Dial          │
│    │ Dial     │         │ Bind     │   • Dark Service       │
│    └──────────┘         └──────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### Controller
- **Vai trò:** "Bộ não" quản lý toàn bộ mạng overlay.
- **Chức năng:**
  - Quản lý Identity: tạo, enroll, disable, delete.
  - Quản lý Service: định nghĩa service name, thuộc tính.
  - Quản lý Policy: Bind policy (ai được host), Dial policy (ai được connect).
  - PKI: CA nội bộ, ký certificate cho identity khi enroll.
- **Không nằm trên data path:** Controller không xử lý traffic ứng dụng.

### Edge Router
- **Vai trò:** Chuyển tiếp traffic giữa các endpoint.
- **Chức năng:**
  - Nhận kết nối từ SDK (Client Dial, Server Bind).
  - Tạo mesh links với các Router khác.
  - Smart routing: chọn đường đi tối ưu.
  - E2E encryption: traffic mã hóa từ nguồn đến đích.

### Identity & Enrollment
- **Identity:** Thực thể mật mã học duy nhất (user, device, app, router).
- **Enrollment process:**
  1. Admin tạo identity trên Controller → nhận enrollment JWT (one-time-use).
  2. Client nhận JWT → trình diện cho Controller.
  3. Controller xác thực JWT → Client sinh keypair + CSR.
  4. Controller ký certificate → Client nhận identity config (JSON).
  5. Client sử dụng identity config cho mọi kết nối sau này.

---

## 9.2 Dark Service Mechanism — Technical Deep Dive

### Traditional API vs Dark Service

```
═══════════ TRADITIONAL API ═══════════

  Internet ──► Port 443 OPEN ──► API Server
                 │
                 ├── nmap: "443/tcp open https"
                 ├── curl: responds with data
                 └── DDoS: target reachable

═══════════ DARK SERVICE ═══════════

  Internet ──► ??? ──► NOTHING
                │
                ├── nmap: "All 65535 ports filtered"
                ├── curl: "Connection refused" (no endpoint)
                └── DDoS: no target exists

  Enrolled   ──► Ziti Router ──► Ziti Fabric ──► Dark Service
  Client          (outbound)      (encrypted)     (no inbound port)
```

### How It Works

1. **Server side (API Gateway):**
   - Nhúng OpenZiti SDK vào ứng dụng Go.
   - Gọi `ctx.Listen("financial-ledger-service")` — **KHÔNG** gọi `net.Listen("tcp", ":8080")`.
   - SDK tạo kết nối **outbound** tới Ziti Router gần nhất.
   - Đăng ký service name "financial-ledger-service" trên mạng overlay.
   - **Kết quả:** Không có socket TCP/UDP nào mở trên host.

2. **Client side:**
   - Nhúng OpenZiti SDK vào ứng dụng Client Go.
   - Gọi `ctx.Dial("financial-ledger-service")`.
   - SDK tạo kết nối **outbound** tới Ziti Router gần nhất.
   - Router kiểm tra Dial policy → nếu hợp lệ, tạo circuit tới Server.
   - **Kết quả:** Client kết nối tới service mà không cần biết IP/port của server.

3. **Key insight: ALL connections are OUTBOUND.**
   - Server → outbound tới Router (Bind).
   - Client → outbound tới Router (Dial).
   - Không bao giờ có kết nối inbound tới application.
   - Firewall rule: chỉ cần cho phép outbound → zero inbound ports.

---

## 9.3 Service Intercept vs Service Binding

| Phương thức | Cơ chế | Ứng dụng cần sửa? | Mức độ tàng hình |
|---|---|---|---|
| **SDK Binding** (chọn) | Nhúng SDK vào app, gọi `ctx.Listen()` | ✅ Có — thay `net.Listen` | ⭐⭐⭐⭐⭐ Tuyệt đối |
| **Tunneler Intercept** | Ziti Tunneler chặn traffic ở OS level | ❌ Không cần sửa | ⭐⭐⭐⭐ Rất cao |
| **Proxy (ziti-edge-tunnel)** | Reverse proxy qua Ziti tunnel | ❌ Không cần sửa | ⭐⭐⭐ Cao (vẫn có local port) |

**Quyết định:** Sử dụng **SDK Binding** để đạt mức tàng hình tuyệt đối. App không mở port ở bất kỳ tầng nào.

---

## 9.4 So sánh Dark Service với các giải pháp khác

### vs VPN

| Tiêu chí | VPN | Dark Service (OpenZiti) |
|---|---|---|
| Mô hình | Network-level tunnel | Application-level overlay |
| Granularity | All-or-nothing network access | Per-service access control |
| Identity | IP-based, shared credentials | Cryptographic identity per entity |
| Lateral Movement | ⚠️ Once inside VPN = access everything | ✅ Each service isolated |
| Port visibility | Server ports still visible inside VPN | Zero ports visible anywhere |
| Performance | Bottleneck at VPN concentrator | Distributed mesh routing |

### vs Reverse Proxy (nginx, Caddy)

| Tiêu chí | Reverse Proxy | Dark Service |
|---|---|---|
| Public exposure | ⚠️ Proxy port open (443) | ✅ Zero ports open |
| DDoS target | ⚠️ Proxy is the target | ✅ No target exists |
| Auth model | Connect-then-Authenticate | Authenticate-then-Connect |
| API Discovery | ⚠️ Endpoints discoverable | ✅ Nothing to discover |

### vs Bastion Host / Jump Server

| Tiêu chí | Bastion Host | Dark Service |
|---|---|---|
| SSH port | ⚠️ Port 22 open | ✅ Zero ports |
| Single point of failure | ⚠️ Yes | ✅ Distributed mesh |
| Audit | ⚠️ SSH session logging | ✅ Per-request cryptographic audit |
| Scalability | ⚠️ Manual management | ✅ Identity-based automation |

### vs Service Mesh (Istio/Linkerd)

| Tiêu chí | Service Mesh | Dark Service |
|---|---|---|
| Scope | East-West (inside cluster) | North-South + East-West + Cross-network |
| Kubernetes required | ✅ Yes (practically) | ❌ No |
| Public ingress | ⚠️ Still needs ingress gateway | ✅ Zero ingress |
| Sidecar overhead | ⚠️ Per-pod sidecar | ✅ SDK embedded, no sidecar |
| Dark capability | ❌ Services still discoverable inside cluster | ✅ Completely dark |

---

## 9.5 OpenZiti Policy Architecture

### Service Policy Types

```
┌─────────────────────────────────────────────────────────────┐
│                    POLICY ARCHITECTURE                        │
│                                                              │
│  ┌────────────────────┐                                     │
│  │  Bind Policy        │                                    │
│  │  "Who can HOST      │                                    │
│  │   this service?"    │                                    │
│  │                     │                                    │
│  │  Identity: gateway  │ ──► Service: financial-ledger      │
│  │  Action: Bind       │                                    │
│  └────────────────────┘                                     │
│                                                              │
│  ┌────────────────────┐                                     │
│  │  Dial Policy        │                                    │
│  │  "Who can CONNECT   │                                    │
│  │   to this service?" │                                    │
│  │                     │                                    │
│  │  Identity: client-* │ ──► Service: financial-ledger      │
│  │  Action: Dial       │                                    │
│  └────────────────────┘                                     │
│                                                              │
│  ┌────────────────────┐                                     │
│  │  Edge Router Policy │                                    │
│  │  "Which routers can │                                    │
│  │   this identity     │                                    │
│  │   connect through?" │                                    │
│  │                     │                                    │
│  │  Identity: #all     │ ──► Router: #all                   │
│  └────────────────────┘                                     │
└─────────────────────────────────────────────────────────────┘
```

### Identity Provisioning for This Project

| Identity Name | Type | Permissions | Purpose |
|---|---|---|---|
| `gateway-identity` | Device | Bind `financial-ledger-service` | API Gateway Dark Service |
| `idp-identity` | Device | Bind `identity-provider-service` | IdP Dark Service |
| `client-alice` | User | Dial `financial-ledger-service`, Dial `identity-provider-service` | End user Alice |
| `client-bob` | User | Dial `financial-ledger-service`, Dial `identity-provider-service` | End user Bob |
| `admin-identity` | Admin | Dial all, Manage via ZAC | System administrator |

---

> **Next:** [PART 10 — Security Architecture](./10_SECURITY_ARCHITECTURE.md)
