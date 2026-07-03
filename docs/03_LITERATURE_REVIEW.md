# PART 3 — LITERATURE REVIEW

## 3.1 NIST Zero Trust Architecture (SP 800-207)

**Nguồn:** NIST Special Publication 800-207, August 2020.

**Triết lý cốt lõi:** "Never Trust, Always Verify" — Không tin tưởng bất kỳ thực thể nào dựa trên vị trí mạng. Mọi truy cập đều phải xác thực, ủy quyền, và mã hóa.

**7 Nguyên tắc (Tenets):**
1. Mọi tài nguyên đều là đối tượng bảo vệ.
2. Mọi giao tiếp phải được mã hóa, bất kể vị trí mạng.
3. Quyền truy cập cấp theo từng phiên (per-session).
4. Chính sách truy cập dựa trên thuộc tính động (identity, device posture, context).
5. Giám sát liên tục tư thế an toàn của tất cả tài sản.
6. Xác thực và ủy quyền nghiêm ngặt trước mọi phiên.
7. Thu thập dữ liệu liên tục để cải tiến chính sách.

**Áp dụng trong dự án:** Nguyên tắc 1-7 là nền tảng thiết kế toàn bộ kiến trúc. OpenZiti thực thi Tenet 2 (mã hóa E2E), DPoP thực thi Tenet 3 (per-request verification).

---

## 3.2 Google BeyondCorp

**Nguồn:** Google Research Papers, 2014-2017.

**Khái niệm:** Mô hình Zero Trust đầu tiên triển khai quy mô lớn tại Google. Loại bỏ hoàn toàn VPN, thay bằng xác thực thiết bị + người dùng ở mọi điểm truy cập.

**Nguyên lý:**
- Truy cập không phụ thuộc vị trí mạng.
- Mỗi thiết bị có danh tính mật mã học (device certificate).
- Access Proxy xác thực mọi request.
- Liên tục đánh giá trust level của thiết bị.

**Áp dụng trong dự án:** Mô hình device trust của BeyondCorp → mTLS client certificate. Access Proxy → API Gateway Dark Service.

---

## 3.3 FAPI 2.0 (Financial-grade API)

**Nguồn:** OpenID Foundation — FAPI 2.0 Security Profile (Final, 19/02/2025).

**Bản chất:** Hardened security profile xây dựng trên OAuth 2.1, thiết kế cho API có giá trị cao (ngân hàng, y tế, chính phủ).

**Yêu cầu bắt buộc:**
- **Sender-Constrained Tokens**: Bearer token thuần bị cấm. Token phải ràng buộc với Client qua DPoP hoặc mTLS.
- **PKCE bắt buộc**: Mọi Authorization Code flow phải dùng PKCE (S256).
- **Exact Redirect URI Matching**: Không cho phép wildcard redirect.
- **Formal Attacker Model**: Thiết kế dựa trên mô hình kẻ tấn công chính thức.

**Áp dụng trong dự án:** FAPI 2.0 là chuẩn target cho toàn bộ Identity & Access Architecture.

---

## 3.4 OAuth 2.1

**Nguồn:** IETF draft-ietf-oauth-v2-1.

**Thay đổi so với OAuth 2.0:**
- PKCE bắt buộc cho tất cả Authorization Code flows.
- Loại bỏ Implicit Grant.
- Loại bỏ Resource Owner Password Credentials.
- Refresh Token phải sender-constrained hoặc one-time-use.
- Exact redirect URI matching.

**Áp dụng trong dự án:** OAuth 2.1 là authorization framework cơ sở.

---

## 3.5 OpenID Connect (OIDC)

**Nguồn:** OpenID Foundation — OpenID Connect Core 1.0.

**Vai trò:** Identity layer trên OAuth 2.0. Cung cấp ID Token chứa thông tin xác thực người dùng.

**Áp dụng trong dự án:** OIDC discovery endpoint (`.well-known/openid-configuration`) cho phép Client tự động phát hiện cấu hình IdP.

---

## 3.6 SPIFFE/SPIRE

**Nguồn:** CNCF SPIFFE Specification v1.0.

**Khái niệm:**
- **SPIFFE** (Secure Production Identity Framework For Everyone): Tiêu chuẩn định danh workload.
- **SPIRE**: Runtime implementation của SPIFFE.
- SPIFFE ID format: `spiffe://trust-domain/workload-identifier`.

**So sánh với OpenZiti:**
- SPIFFE/SPIRE tập trung vào workload identity trong service mesh.
- OpenZiti cung cấp full overlay network + identity + dark services.
- Dự án này chọn OpenZiti vì nó bao gồm networking layer mà SPIFFE/SPIRE không có.

---

## 3.7 OpenZiti

**Nguồn:** OpenZiti Documentation (openziti.io), NetFoundry.

**Định nghĩa:** Nền tảng zero trust networking mã nguồn mở. Tạo overlay network cho phép kết nối ứng dụng mà không cần mở cổng inbound.

**Thành phần:**
- **Controller**: Quản lý control plane — identities, services, policies.
- **Router**: Tạo data fabric mesh, chuyển tiếp traffic mã hóa.
- **SDK**: Nhúng trực tiếp vào ứng dụng (Go, Java, C#, ...).
- **Tunneler**: Cho phép ứng dụng không-Ziti kết nối qua overlay.

**Cơ chế Dark Service:**
- Ứng dụng dùng Ziti SDK gọi `ctx.Listen("service-name")` thay vì `net.Listen("tcp", ":port")`.
- Không mở socket TCP/UDP trên host.
- Traffic chỉ đi qua Ziti Router → E2E encrypted.

**Áp dụng trong dự án:** Lớp L3 (Overlay Network) và L4 (API Gateway) sử dụng OpenZiti làm nền tảng.

---

## 3.8 Service Mesh (Istio, Linkerd)

**Nguồn:** Istio Documentation, Linkerd Documentation.

**Khái niệm:** Infrastructure layer xử lý service-to-service communication (mTLS, load balancing, observability).

**So sánh với OpenZiti:**
- Service Mesh hoạt động bên trong cluster (east-west traffic).
- OpenZiti hoạt động cross-network (bao gồm Internet).
- Service Mesh vẫn cần public ingress → không thực sự "dark".
- OpenZiti triệt tiêu hoàn toàn inbound ports.

**Quyết định:** Không chọn Service Mesh cho dự án này. OpenZiti đáp ứng tốt hơn yêu cầu Dark Service.

---

## 3.9 mTLS (Mutual TLS — RFC 8705)

**Nguồn:** IETF RFC 8705, "OAuth 2.0 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens".

**Hai cơ chế:**
1. **mTLS Client Authentication**: Client trình diện X.509 certificate trong TLS handshake → thay thế `client_secret`.
2. **Certificate-Bound Access Tokens**: Token chứa fingerprint SHA-256 của client cert (`cnf.x5t#S256`) → token bị đánh cắp không dùng được nếu không có cert tương ứng.

**Áp dụng trong dự án:** mTLS là lớp xác thực tại tầng transport, bổ sung cho DPoP ở tầng application.

---

## 3.10 DPoP (RFC 9449)

**Nguồn:** IETF RFC 9449, "OAuth 2.0 Demonstrating Proof-of-Possession at the Application Layer".

**Cơ chế:** Client sinh cặp khóa ECC, ký DPoP Proof JWT cho mỗi request. Server xác thực proof trước khi chấp nhận token.

**Cấu trúc DPoP Proof JWT:**
```
Header: { "typ": "dpop+jwt", "alg": "ES256", "jwk": {public_key} }
Payload: { "htm": "POST", "htu": "https://api/transfer", "jti": "unique-id", "iat": 1719990000, "ath": "hash-of-access-token" }
Signature: ECDSA(header.payload, private_key)
```

**Áp dụng trong dự án:** DPoP là cơ chế sender-constraining chính, hoạt động ở tầng application layer (L1 + L4).

---

## 3.11 Token Binding (RFC 8471)

**Nguồn:** IETF RFC 8471, "The Token Binding Protocol Version 1.0".

**Khái niệm:** Ràng buộc token với TLS connection. Bị hạn chế bởi thiếu hỗ trợ trình duyệt (Chrome đã loại bỏ hỗ trợ).

**Quyết định:** Không sử dụng Token Binding. DPoP (RFC 9449) là giải pháp thay thế hiện đại hơn, hoạt động ở application layer, không phụ thuộc TLS stack.

---

## 3.12 Bảng tổng hợp Literature Review

| Công nghệ / Tiêu chuẩn | Vai trò | Ưu điểm | Nhược điểm | Độ phù hợp |
|---|---|---|---|---|
| **NIST SP 800-207** | Khung kiến trúc ZTA | Chuẩn quốc tế, toàn diện, được chính phủ Mỹ áp dụng | Lý thuyết, không quy định triển khai cụ thể | ⭐⭐⭐⭐⭐ |
| **Google BeyondCorp** | Mô hình tham chiếu ZTA | Đã chứng minh quy mô lớn tại Google | Phụ thuộc hạ tầng Google | ⭐⭐⭐⭐ |
| **FAPI 2.0** | Chuẩn API tài chính | Bắt buộc sender-constrained tokens, attacker model chính thức | Phức tạp triển khai | ⭐⭐⭐⭐⭐ |
| **OAuth 2.1** | Authorization framework | Cleanup OAuth 2.0, PKCE bắt buộc | Chưa ra bản Final chính thức | ⭐⭐⭐⭐⭐ |
| **OpenID Connect** | Identity layer | Chuẩn hoá, phổ biến rộng rãi | Overkill cho machine-to-machine | ⭐⭐⭐⭐ |
| **SPIFFE/SPIRE** | Workload identity | CNCF standard, vendor-neutral | Không có networking layer | ⭐⭐⭐ |
| **OpenZiti** | Zero trust overlay + Dark Service | Mã nguồn mở, SDK Go chính thức, Dark Service native | Cộng đồng nhỏ hơn Istio | ⭐⭐⭐⭐⭐ |
| **Istio** | Service mesh | Ecosystem lớn, mTLS tự động | Phức tạp, cần K8s, không thực sự dark | ⭐⭐⭐ |
| **mTLS (RFC 8705)** | Transport-layer auth | Chứng minh danh tính PKI, certificate-bound tokens | Quản lý certificate lifecycle phức tạp | ⭐⭐⭐⭐⭐ |
| **DPoP (RFC 9449)** | Application-layer PoP | Không phụ thuộc TLS stack, chống token theft hiệu quả | Cần thêm logic ở cả client và server | ⭐⭐⭐⭐⭐ |
| **Token Binding (RFC 8471)** | TLS-layer token binding | Ràng buộc mạnh với TLS | Chrome đã bỏ hỗ trợ, ít được áp dụng | ⭐⭐ |

---

> **Next:** [PART 4 — Requirement Analysis](./04_REQUIREMENT_ANALYSIS.md)
