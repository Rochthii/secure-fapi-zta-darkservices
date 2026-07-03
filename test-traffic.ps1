# Script to generate continuous real traffic for Secure FAPI-ZTA SOC Dashboard demonstration

Write-Host "Generating FAPI-ZTA Client Transactions Traffic..." -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop traffic generation." -ForegroundColor Yellow

# Add local Go path to env path
$env:Path = "e:\Projects\Project_TN\secure-fapi-zta-darkservices\go-local\go\bin;" + $env:Path

# Change to client directory
cd client

$count = 1
while ($true) {
    Write-Host "Sending Request #$count..." -ForegroundColor Gray
    
    # 1. Send balance request
    & go run main.go -identity client-alice -cmd balance -ziti=false
    
    # 2. Randomly send transfer request to generate new WORM logs
    if ($count % 5 -eq 0) {
        $amount = Get-Random -Minimum 100 -Maximum 5000
        Write-Host "Executing Transaction: Alice transfers $amount..." -ForegroundColor Yellow
        & go run main.go -identity client-alice -cmd transfer -amount $amount -desc "Demo Auto Transfer $amount" -ziti=false
    }
    
    $count++
    Start-Sleep -Seconds 2
}
