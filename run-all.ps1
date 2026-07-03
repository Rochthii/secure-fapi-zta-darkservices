# Script to run all components of the Secure FAPI-ZTA project (Docker, IdP, Gateway, Dashboard)

Write-Host "Starting Secure FAPI-ZTA & Dark Services System..." -ForegroundColor Cyan

# Append Go path locally to be safe
$env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path

# Step 0: Ensure Docker Desktop is running
Write-Host "Step 0: Checking Docker Daemon status..." -ForegroundColor Yellow
$dockerCheck = & docker info 2>&1
if ($dockerCheck -like "*error during connect*" -or $dockerCheck -like "*is the docker daemon running*") {
    Write-Host "Docker is not running. Attempting to start Docker Desktop..." -ForegroundColor Yellow
    $dockerPath = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    if (Test-Path $dockerPath) {
        Start-Process $dockerPath
        Write-Host "Waiting for Docker Daemon to initialize (this may take up to 30 seconds)..." -ForegroundColor Yellow
        while ($true) {
            Start-Sleep -Seconds 5
            $check = & docker info 2>&1
            if ($check -notlike "*error during connect*" -and $check -notlike "*is the docker daemon running*") {
                Write-Host "Docker Daemon is active!" -ForegroundColor Green
                break
            }
            Write-Host "Still waiting for Docker Daemon to start..." -ForegroundColor Gray
        }
    } else {
        Write-Warning "Docker Desktop executable not found at default location: $dockerPath"
        Write-Warning "Please start Docker Desktop manually before running this script."
        Read-Host "Press ENTER after starting Docker Desktop to continue..."
    }
} else {
    Write-Host "Docker Daemon is active." -ForegroundColor Green
}

# Step 1: Start Docker infrastructure
Write-Host "Step 1: Starting Docker containers (Ziti Overlay + PostgreSQL + Telemetry)..." -ForegroundColor Yellow
docker compose -f docker/docker-compose.yml up -d

# Step 2: Start IdP in a new window (only if port 8081 is free)
$idpUsed = Get-NetTCPConnection -LocalPort 8081 -ErrorAction SilentlyContinue
if ($null -eq $idpUsed) {
    Write-Host "Step 2: Starting Identity Provider (IdP) on port 8081..." -ForegroundColor Yellow
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path = 'e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;' + `$env:Path; cd idp; go run main.go"
} else {
    Write-Host "Identity Provider (IdP) is already running on port 8081." -ForegroundColor Green
}

# Step 3: Start API Gateway in a new window (only if port 8080 is free)
$gwUsed = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
if ($null -eq $gwUsed) {
    Write-Host "Step 3: Starting API Gateway on port 8080 (Debug Mode)..." -ForegroundColor Yellow
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path = 'e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;' + `$env:Path; cd gateway; `$env:USE_ZITI = 'false'; go run main.go"
} else {
    Write-Host "API Gateway is already running on port 8080." -ForegroundColor Green
}

# Step 4: Start Next.js Dashboard in a new window (only if port 3001 is free)
$dbUsed = Get-NetTCPConnection -LocalPort 3001 -ErrorAction SilentlyContinue
if ($null -eq $dbUsed) {
    Write-Host "Step 4: Starting Cyber SOC Dashboard on port 3001..." -ForegroundColor Yellow
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd dashboard; npm run dev -- -p 3001"
} else {
    Write-Host "Cyber SOC Dashboard is already running on port 3001." -ForegroundColor Green
}

Write-Host "All components successfully launched in separate windows!" -ForegroundColor Green
Write-Host "API Gateway: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Identity Provider: http://localhost:8081" -ForegroundColor Cyan
Write-Host "SOC Dashboard: http://localhost:3001" -ForegroundColor Cyan
