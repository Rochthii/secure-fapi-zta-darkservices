#!/bin/bash
# Script khởi tạo Root CA và Intermediate CA cho hệ thống mTLS
# Sử dụng thuật toán Elliptic Curve ECC P-256

set -e

# Đi chuyển về thư mục gốc của chứng chỉ (certs/)
cd "$(dirname "$0")/.."
mkdir -p ca

echo "=== [1/5] Khởi tạo Root CA Key (ECC P-256) ==="
openssl ecparam -name prime256v1 -genkey -noout -out ca/root-ca.key

echo "=== [2/5] Tạo Chứng chỉ Root CA tự ký (Self-signed, Hạn dùng 10 năm) ==="
openssl req -new -x509 -sha256 \
    -key ca/root-ca.key \
    -subj "/CN=FAPI-ZTA Root CA/O=PTIT Thesis/C=VN" \
    -days 3650 \
    -out ca/root-ca.crt

echo "=== [3/5] Khởi tạo Intermediate CA Key (ECC P-256) ==="
openssl ecparam -name prime256v1 -genkey -noout -out ca/intermediate-ca.key

echo "=== [4/5] Tạo Certificate Signing Request (CSR) cho Intermediate CA ==="
openssl req -new -sha256 \
    -key ca/intermediate-ca.key \
    -subj "/CN=FAPI-ZTA Intermediate CA/O=PTIT Thesis/C=VN" \
    -out ca/intermediate-ca.csr

echo "=== [5/5] Ký duyệt chứng chỉ Intermediate CA sử dụng Root CA ==="
# Tạo file cấu hình phần mở rộng (Extensions) cho CA trung gian
cat <<EOF > ca/ca.ext
[ ca_ext ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
EOF

openssl x509 -req -sha256 -days 1095 \
    -in ca/intermediate-ca.csr \
    -CA ca/root-ca.crt \
    -CAkey ca/root-ca.key \
    -CAcreateserial \
    -extfile ca/ca.ext \
    -extensions ca_ext \
    -out ca/intermediate-ca.crt

# Dọn dẹp tệp phụ CSR và file cấu hình tạm
rm -f ca/intermediate-ca.csr ca/ca.ext

echo "=== HOÀN THÀNH ==="
echo "Đã tạo Root CA (ca/root-ca.crt) và Intermediate CA (ca/intermediate-ca.crt) thành công!"
