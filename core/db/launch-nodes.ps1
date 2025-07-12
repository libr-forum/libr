$DB_PORT_START = 5433
$NODE_PORT_START = 8000
$NODE_COUNT = 50

for ($i = 0; $i -lt $NODE_COUNT; $i++) {
    $DB_PORT = $DB_PORT_START + $i
    $NODE_PORT = $NODE_PORT_START + $i

    Write-Host "🖥️ Launching node $i | DB_PORT=$DB_PORT | PORT=$NODE_PORT"

    # Create .env file for reference/debugging (optional)
    $envFile = ".env.$NODE_PORT"
    @"
DB_HOST=localhost
DB_PORT=$DB_PORT
DB_USER=user
DB_PASS=password
DB_NAME=database$i
BOOTSTRAP=127.0.0.1:8000,127.0.0.1:8010,127.0.0.1:8039
PORT=$NODE_PORT
"@ | Set-Content $envFile

    # Read env vars into a single string like: $env:KEY='val'; $env:KEY2='val2';
    $envSetup = (Get-Content $envFile | ForEach-Object { "`$env:" + ($_ -replace "=", "='") + "'" }) -join "; "

    # Start new PowerShell window with env vars and run go
    Start-Process "powershell.exe" -ArgumentList @("-NoExit", "-Command", "$envSetup; go run main.go")

    Start-Sleep -Seconds 1
}
