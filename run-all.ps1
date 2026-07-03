# Script to run all components of the Secure FAPI-ZTA project (Docker, IdP, Gateway, Dashboard)

Write-Host "Starting Secure FAPI-ZTA & Dark Services System..." -ForegroundColor Cyan

# Append Go path locally to be safe
$env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path

# Step 1: Start Docker infrastructure
Write-Host "Step 1: Starting Docker containers (Ziti Overlay + PostgreSQL + Telemetry)..." -ForegroundColor Yellow
docker compose -f docker/docker-compose.yml up -d

# Step 2: Start IdP in a new window
Write-Host "Step 2: Starting Identity Provider (IdP) on port 8081..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path = 'e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;' + `$env:Path; cd idp; go run main.go"

# Step 3: Start API Gateway in a new window
Write-Host "Step 3: Starting API Gateway on port 8080 (Debug Mode)..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path = 'e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;' + `$env:Path; cd gateway; `$env:USE_ZITI = 'false'; go run main.go"

# Step 4: Start Next.js Dashboard in a new window
Write-Host "Step 4: Starting Cyber SOC Dashboard on port 3001..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd dashboard; npm run dev -- -p 3001"

Write-Host "All components successfully launched in separate windows!" -ForegroundColor Green
Write-Host "API Gateway: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Identity Provider: http://localhost:8081" -ForegroundColor Cyan
Write-Host "SOC Dashboard: http://localhost:3001" -ForegroundColor Cyan
