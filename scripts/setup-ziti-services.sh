#!/bin/bash
# Script tự động hóa cấu hình dịch vụ, identities và policies trên OpenZiti Controller
# Chạy trên máy host nhưng thực thi các câu lệnh CLI trong Container Docker

set -e

# Di chuyển về thư mục gốc của dự án
cd "$(dirname "$0")/.."

# Load biến cấu hình từ file docker/.env
if [ -f docker/.env ]; then
    echo "Tải cấu hình từ docker/.env..."
    # Đọc file .env, lọc bỏ chú thích và dòng trống, xuất ra biến môi trường tạm thời
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

echo "=== [1/4] XÓA CẤU HÌNH CŨ NẾU TỒN TẠI ==="
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  ziti edge delete service-policy gateway-bind-policy --ignore-missing
  ziti edge delete service-policy client-dial-policy --ignore-missing
  ziti edge delete service financial-ledger-service --ignore-missing
  ziti edge delete identity gateway-prod --ignore-missing
  ziti edge delete identity gateway-dev --ignore-missing
  ziti edge delete identity client-alice --ignore-missing
  ziti edge delete identity client-bob --ignore-missing
  ziti edge delete identity client-evil --ignore-missing
  ziti edge delete edge-router-policy all-routers-all-identities --ignore-missing
  ziti edge delete service-edge-router-policy all-services-all-routers --ignore-missing
"

echo "=== [2/4] THIẾT LẬP ROUTER POLICIES (MESH NETWORKING) ==="
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  # Cho phép tất cả các Identity kết nối thông qua tất cả Edge Routers
  ziti edge create edge-router-policy all-routers-all-identities --edge-router-roles '#all' --identity-roles '#all'
  
  # Cho phép tất cả dịch vụ chuyển tiếp qua tất cả Edge Routers
  ziti edge create service-edge-router-policy all-services-all-routers --service-roles '#all' --edge-router-roles '#all'
"

echo "=== [3/4] ĐĂNG KÝ DỊCH VỤ & KHỞI TẠO CÁC DANH TÍNH (IDENTITIES) ==="
# Đăng ký dịch vụ logic đại diện cho API Gateway tàng hình
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  ziti edge create service financial-ledger-service
"

# Khởi tạo các danh tính mạng và xuất mã Token ghi danh dùng 1 lần (.jwt)
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  mkdir -p /persistent/tokens
  
  # API Gateways (Vai trò: Host/Bind dịch vụ)
  ziti edge create identity gateway-prod -a gateways,gateway-prod -o /persistent/tokens/gateway-prod.jwt
  ziti edge create identity gateway-dev -a gateways,gateway-dev -o /persistent/tokens/gateway-dev.jwt
  
  # Các Client thiết bị hợp lệ (Vai trò: Dial/Connect dịch vụ)
  ziti edge create identity client-alice -a clients,alice -o /persistent/tokens/client-alice.jwt
  ziti edge create identity client-bob -a clients,bob -o /persistent/tokens/client-bob.jwt
  
  # Client giả lập hacker / không có quyền (Dùng để kiểm thử chặn truy cập ở tầng mạng)
  ziti edge create identity client-evil -a evil -o /persistent/tokens/client-evil.jwt
"

echo "=== [4/4] THIẾT LẬP CHÍNH SÁCH BẢO MẬT (BIND & DIAL POLICIES) ==="
docker exec -t $ZITI_CONTAINER /bin/bash -c "source /persistent/ziti.env &&
  # Cấp quyền Bind (Hosting) cho nhóm Gateway (#gateways)
  ziti edge create service-policy gateway-bind-policy Bind --service-roles '@financial-ledger-service' --identity-roles '#gateways'
  
  # Cấp quyền Dial (Connecting) cho nhóm Client hợp lệ (#clients), loại trừ client-evil (#evil)
  ziti edge create service-policy client-dial-policy Dial --service-roles '@financial-ledger-service' --identity-roles '#clients'
"

# Sao chép các tệp JWT ghi danh từ Volume Docker ra ngoài Host để các ứng dụng Client/Gateway nạp
echo "=== SAO CHÉP TOKEN JWT RA MÁY HOST ==="
mkdir -p docker/tokens
docker cp $ZITI_CONTAINER:/persistent/tokens/. docker/tokens/

echo "=== HOÀN THÀNH CẤU HÌNH OVERLAY NETWORK ==="
echo "Các mã thông báo ghi danh (.jwt) đã được lưu tại: docker/tokens/"
