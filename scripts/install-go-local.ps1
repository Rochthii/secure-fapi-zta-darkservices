$zipPath = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go.zip"
$destPath = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local"

if (-not (Test-Path $destPath)) {
    New-Item -ItemType Directory -Path $destPath -Force | Out-Null
}

Write-Host "=== BẮT ĐẦU TẢI GOLANG 1.22.5 PORTABLE ==="
Write-Host "Tải tệp zip từ go.dev..."
$ProgressPreference = 'SilentlyContinue' # Ẩn thanh tiến trình để tải nhanh hơn
Invoke-WebRequest -Uri "https://go.dev/dl/go1.22.5.windows-amd64.zip" -OutFile $zipPath

Write-Host "=== GIẢI NÉN GOLANG ==="
Write-Host "Đang giải nén vào $destPath..."
Expand-Archive -Path $zipPath -DestinationPath $destPath -Force

Write-Host "=== DỌN DẸP ==="
Remove-Item -Path $zipPath -Force

Write-Host "=== HOÀN THÀNH ==="
$goExe = Join-Path $destPath "go\bin\go.exe"
if (Test-Path $goExe) {
    Write-Host "Golang đã được cài đặt cục bộ tại: $goExe"
    & $goExe version
} else {
    Write-Host "LỖI: Không tìm thấy go.exe sau khi giải nén."
}
