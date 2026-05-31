# Cross-compilation script for PostgreSQL Client (Windows/PowerShell)
# This script builds the application for multiple platforms

$APP_NAME = "postgresql-client"
$VERSION = if ($env:VERSION) { $env:VERSION } else { "1.0.0" }
$OUTPUT_DIR = "dist"

Write-Host "Building $APP_NAME v$VERSION..." -ForegroundColor Cyan
Write-Host ""

# Create output directory
New-Item -ItemType Directory -Path $OUTPUT_DIR -Force | Out-Null

# Define target platforms and architectures
$TARGETS = @{
    "darwin/amd64"   = "macOS_x86_64"
    "darwin/arm64"   = "macOS_ARM64"
    "linux/amd64"    = "Linux_x86_64"
    "linux/386"      = "Linux_i386"
    "linux/arm64"    = "Linux_ARM64"
    "windows/amd64"  = "Windows_x86_64.exe"
    "windows/386"    = "Windows_i386.exe"
}

# Build for each target
foreach ($target in $TARGETS.Keys) {
    $GOOS_GOARCH = $target
    $OUTPUT_NAME = $TARGETS[$target]
    
    $GOOS = $GOOS_GOARCH.Split('/')[0]
    $GOARCH = $GOOS_GOARCH.Split('/')[1]
    
    Write-Host "Building for $GOOS/$GOARCH -> $OUTPUT_NAME..." -ForegroundColor Yellow
    
    # Output file name with app name prefix
    if ($GOOS -eq "windows") {
        $OUTPUT_FILE = Join-Path $OUTPUT_DIR "${APP_NAME}-${OUTPUT_NAME}"
    } else {
        $OUTPUT_FILE = Join-Path $OUTPUT_DIR "${APP_NAME}-${OUTPUT_NAME}"
    }
    
    Write-Host "  Output: $OUTPUT_FILE"
    
    $env:CGO_ENABLED = "0"
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    go build -o $OUTPUT_FILE -ldflags="-s -w" . 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Built: $OUTPUT_NAME" -ForegroundColor Green
    } else {
        Write-Host "  Failed: $OUTPUT_NAME" -ForegroundColor Red
    }
    Write-Host ""
}

Write-Host "Build completed! Output directory: $OUTPUT_DIR" -ForegroundColor Cyan
Write-Host ""
Write-Host "Files built:"
Get-ChildItem $OUTPUT_DIR | Select-Object Name, Length