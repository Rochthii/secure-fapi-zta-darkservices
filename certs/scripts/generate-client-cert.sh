#!/bin/bash
# Script cấp phát chứng chỉ Client phục vụ mTLS (ECC P-256)

set -e

# Di chuyển về thư mục gốc certs/
cd "$(dirname "$0")/.."
mkdir -p clients

CLIENT_NAME=$1

if [ -z "$CLIENT_NAME" ]; then
    echo "LỖI: Vui lòng truyền tên Client làm tham số thứ nhất."
    echo "Ví dụ: ./generate-client-cert.sh client-alice"
    exit 1
fi

echo "=== [1/3] Khởi tạo Client Key cho $CLIENT_NAME (ECC P-256) ==="
openssl ecparam -name prime256v1 -genkey -noout -out clients/"$CLIENT_NAME".key

echo "=== [2/3] Tạo CSR cho $CLIENT_NAME ==="
openssl req -new -sha256 \
    -key clients/"$CLIENT_NAME".key \
    -subj "/CN=$CLIENT_NAME/O=PTIT Thesis/C=VN" \
    -out clients/"$CLIENT_NAME".csr

echo "=== [3/3] Ký duyệt chứng chỉ Client bằng Intermediate CA ==="
# Cấu hình thuộc tính mở rộng (chỉ cho phép clientAuth)
cat <<EOF > clients/"$CLIENT_NAME".ext
[ client_ext ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature
extendedKeyUsage = clientAuth
EOF

openssl x509 -req -sha256 -days 90 \
    -in clients/"$CLIENT_NAME".csr \
    -CA ca/intermediate-ca.crt \
    -CAkey ca/intermediate-ca.key \
    -CAcreateserial \
    -extfile clients/"$CLIENT_NAME".ext \
    -extensions client_ext \
    -out clients/"$CLIENT_NAME".crt

# Dọn dẹp tệp phụ CSR và file cấu hình tạm
rm -f clients/"$CLIENT_NAME".csr clients/"$CLIENT_NAME".ext

echo "=== HOÀN THÀNH ==="
echo "Đã cấp chứng chỉ Client (clients/$CLIENT_NAME.crt) thành công!"
