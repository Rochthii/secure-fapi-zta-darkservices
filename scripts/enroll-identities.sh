#!/bin/bash
# Script tự động hóa ghi danh (Enrollment) các danh tính OpenZiti
# Chạy trên máy host nhưng thực thi việc enroll mật mã trong Container

set -e

# Di chuyển về thư mục gốc của dự án
cd "$(dirname "$0")/.."

# Load biến cấu hình từ file docker/.env
if [ -f docker/.env ]; then
    echo "Tải cấu hình từ docker/.env..."
    export $(grep -v '^#' docker/.env | xargs)
else
    echo "LỖI: Không tìm thấy tệp docker/.env"
    exit 1
fi

ZITI_USER=${ZITI_USER:-admin}
ZITI_PWD=${ZITI_PWD:-ptit_thesis_2026}
ZITI_CONTAINER=${ZITI_CONTAINER:-docker-ziti-controller-1}

echo "=== ĐĂNG NHẬP VÀO CONTROLLER ==="
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env && ziti edge login \"\${ZITI_CTRL_EDGE_ADVERTISED_ADDRESS}:\${ZITI_CTRL_EDGE_ADVERTISED_PORT}\" -u \"$ZITI_USER\" -p \"$ZITI_PWD\" -y"

echo "=== [1/2] THỰC HIỆN GHI DANH (ENROLLMENT) TRONG CONTAINER ==="
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  mkdir -p /persistent/identities
  
  # Enroll các danh tính Gateways
  ziti edge enroll -j /persistent/tokens/gateway-prod.jwt -o /persistent/identities/gateway-prod.json
  ziti edge enroll -j /persistent/tokens/gateway-dev.jwt -o /persistent/identities/gateway-dev.json
  
  # Enroll các danh tính Clients hợp lệ
  ziti edge enroll -j /persistent/tokens/client-alice.jwt -o /persistent/identities/client-alice.json
  ziti edge enroll -j /persistent/tokens/client-bob.jwt -o /persistent/identities/client-bob.json
  
  # Enroll danh tính Client không hợp lệ (Evil)
  ziti edge enroll -j /persistent/tokens/client-evil.jwt -o /persistent/identities/client-evil.json
"

# Sao chép các tệp cấu hình chứa Private Key/Cert X.509 ra ngoài máy Host
echo "=== [2/2] SAO CHÉP FILE CONFIG JSON RA MÁY HOST ==="
mkdir -p docker/identities
docker cp $ZITI_CONTAINER:/persistent/identities/. docker/identities/

echo "=== HOÀN THÀNH GHI DANH ==="
echo "Các tệp cấu hình danh tính JSON bảo mật đã được lưu tại: docker/identities/"
