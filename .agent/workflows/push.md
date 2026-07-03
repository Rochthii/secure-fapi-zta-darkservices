---
description: Kiểm tra lint, build và push code lên GitHub (Go Project)
---

1. Định dạng code Go:
```powershell
go fmt ./...
```
2. Kiểm tra lỗi tĩnh bằng go vet:
```powershell
go vet ./...
```
3. Chạy build thử các thành phần hoặc chạy Makefile build để đảm bảo không lỗi:
```powershell
go build -o build/idp ./idp/...
go build -o build/gateway ./gateway/...
go build -o build/client ./client/...
```
4. Nếu có lỗi nghiêm trọng ở các bước trên, hãy dừng lại và sửa lỗi.
5. Xem lại toàn bộ những thay đổi đã thực hiện (diff) và tóm tắt ngắn gọn.
6. Đề xuất commit message theo chuẩn Conventional Commits.
7. Thực hiện add, commit và push (Trong PowerShell dùng `;` để ngăn cách câu lệnh thay vì `&&`):
```powershell
git add .; git commit -m "feat/fix: descriptive message"; git push origin main
```

