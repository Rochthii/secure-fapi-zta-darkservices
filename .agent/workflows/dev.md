---
description: Chạy môi trường development local (Go + Docker Compose)
---

1. Kiểm tra xem Docker Desktop đã chạy chưa.
2. Kiểm tra xem file `docker/.env` có tồn tại không. Nếu không, hướng dẫn tạo từ template.
3. Chạy lệnh sau để khởi động cụm hạ tầng (OpenZiti + Postgres + Monitoring):
```powershell
docker compose -f docker/docker-compose.yml up -d
```
4. Kiểm tra trạng thái các container bằng `docker compose -f docker/docker-compose.yml ps`.
5. Đợi cụm OpenZiti khởi động hoàn toàn (khoảng 30 giây đến 1 phút).
6. Hướng dẫn user và chạy các ứng dụng Go (`idp`, `gateway`, `client`) ở chế độ phát triển local hoặc kiểm thử.

