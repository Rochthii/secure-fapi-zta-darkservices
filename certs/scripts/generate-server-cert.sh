#!/bin/bash
# Script cấp phát chứng chỉ Server phục vụ mTLS (ECC P-256)

set -e

# Di chuyển về thư mục gốc certs/
cd "$(dirname "$0")/.."
mkdir -p server

# Tên miền hoặc danh tính server, mặc định là localhost
CN=${1:-localhost}

echo "=== [1/3] Khởi tạo Server Key (ECC P-256) ==="
openssl ecparam -name prime256v1 -genkey -noout -out server/server.key

echo "=== [2/3] Tạo CSR cho Server (CN=$CN) ==="
openssl req -new -sha256 \
    -key server/server.key \
    -subj "/CN=$CN/O=PTIT Thesis/C=VN" \
    -out server/server.csr

echo "=== [3/3] Ký duyệt chứng chỉ Server bằng Intermediate CA ==="
# Cấu hình thuộc tính mở rộng (serverAuth, clientAuth) và SAN (Subject Alternative Names)
cat <<EOF > server/server.ext
[ server_ext ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
DNS.2 = gateway.internal
DNS.3 = idp.internal
IP.1 = 127.0.0.1
EOF

openssl x509 -req -sha256 -days 365 \
    -in server/server.csr \
    -CA ca/intermediate-ca.crt \
    -CAkey ca/intermediate-ca.key \
    -CAcreateserial \
    -extfile server/server.ext \
    -extensions server_ext \
    -out server/server.crt

# Dọn dẹp tệp phụ CSR và file cấu hình tạm
rm -f server/server.csr server/server.ext

echo "=== HOÀN THÀNH ==="
echo "Đã cấp chứng chỉ Server (server/server.crt) thành công!"
