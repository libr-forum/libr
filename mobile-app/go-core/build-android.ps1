# Configuration
$PackagePath = "./bridge"
$OutputName = "libr-core.aar"
$AndroidApi = 21

# Setup output directory
$OutputDir = "../android/app/libs"
if (-not (Test-Path $OutputDir)) {
    $OutputDir = "./libs"
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force
    }
}

$OutputPath = "$OutputDir/$OutputName"

# Setup Go module
$env:GO111MODULE = "on"

# Build
Write-Host "🚀 Building Go mobile bridge for Android..." -ForegroundColor Cyan
# Added -ldflags "-checklinkname=0" for Go 1.23+ compatibility with wlynxg/anet
gomobile bind -v -target=android -androidapi $AndroidApi -ldflags "-checklinkname=0" -o $OutputPath $PackagePath

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful! Output: $OutputPath" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
