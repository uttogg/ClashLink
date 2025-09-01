# ClashLink Release Builder
param([string]$Version = "v0.1.0.1")

$BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$BuildDir = "build"
$DistDir = "dist"

Write-Host "Building ClashLink $Version" -ForegroundColor Green

# Clean build directories
if (Test-Path $BuildDir) { Remove-Item -Path $BuildDir -Recurse -Force }
if (Test-Path $DistDir) { Remove-Item -Path $DistDir -Recurse -Force }
New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
New-Item -ItemType Directory -Path $DistDir -Force | Out-Null

# Build Linux binaries
Write-Host "Building Linux binaries..." -ForegroundColor Blue
Set-Location backend

# Check Go installation
try {
    $GoVersion = go version
    Write-Host "Go Version: $GoVersion"
}
catch {
    Write-Host "Error: Go not found in PATH" -ForegroundColor Red
    exit 1
}

go mod tidy

# Linux amd64
Write-Host "Building Linux amd64..." -ForegroundColor Cyan
$env:GOOS = "linux"
$env:GOARCH = "amd64" 
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o "../$BuildDir/clashlink-linux-amd64" .

# Check amd64 binary size
$Amd64Size = (Get-Item "../$BuildDir/clashlink-linux-amd64").Length
Write-Host "AMD64 binary size: $([math]::Round($Amd64Size/1MB,2)) MB"

# Linux arm64
Write-Host "Building Linux arm64..." -ForegroundColor Cyan
$env:GOARCH = "arm64"
go build -ldflags "-s -w" -o "../$BuildDir/clashlink-linux-arm64" .

# Check arm64 binary size
$Arm64Size = (Get-Item "../$BuildDir/clashlink-linux-arm64").Length
Write-Host "ARM64 binary size: $([math]::Round($Arm64Size/1MB,2)) MB"

# Verify both binaries are reasonable size
if ($Amd64Size -lt 1MB) {
    Write-Host "Warning: AMD64 binary seems too small" -ForegroundColor Yellow
}
if ($Arm64Size -lt 1MB) {
    Write-Host "Warning: ARM64 binary seems too small" -ForegroundColor Yellow
}

# Reset env
Remove-Item env:GOOS -ErrorAction SilentlyContinue
Remove-Item env:GOARCH -ErrorAction SilentlyContinue  
Remove-Item env:CGO_ENABLED -ErrorAction SilentlyContinue

Set-Location ..

# Package Linux x86_64
Write-Host "Packaging Linux x86_64..." -ForegroundColor Blue
$LinuxDir = Join-Path $BuildDir "clashlink-linux-x64"
New-Item -ItemType Directory -Path $LinuxDir -Force | Out-Null

Copy-Item -Path "frontend" -Destination $LinuxDir -Recurse
Copy-Item -Path "subscriptions" -Destination $LinuxDir -Recurse
New-Item -ItemType Directory -Path (Join-Path $LinuxDir "backend") -Force | Out-Null
Copy-Item -Path (Join-Path $BuildDir "clashlink-linux-amd64") -Destination (Join-Path $LinuxDir "backend/clashlink")
Copy-Item -Path "version.json" -Destination $LinuxDir
Copy-Item -Path "upgrade.sh" -Destination $LinuxDir
Copy-Item -Path "docker-*.sh" -Destination $LinuxDir
Copy-Item -Path "Dockerfile" -Destination $LinuxDir
Copy-Item -Path "docker-compose.yml" -Destination $LinuxDir
Copy-Item -Path "README.md" -Destination $LinuxDir
Copy-Item -Path "DEPLOY.md" -Destination $LinuxDir
Copy-Item -Path "DOCKER.md" -Destination $LinuxDir
Copy-Item -Path "env.example" -Destination $LinuxDir

# Package Linux ARM64
Write-Host "Packaging Linux ARM64..." -ForegroundColor Blue
$LinuxArmDir = Join-Path $BuildDir "clashlink-linux-arm64-pkg"
New-Item -ItemType Directory -Path $LinuxArmDir -Force | Out-Null

Copy-Item -Path "frontend" -Destination $LinuxArmDir -Recurse
Copy-Item -Path "subscriptions" -Destination $LinuxArmDir -Recurse
New-Item -ItemType Directory -Path (Join-Path $LinuxArmDir "backend") -Force | Out-Null
Copy-Item -Path (Join-Path $BuildDir "clashlink-linux-arm64") -Destination (Join-Path $LinuxArmDir "backend/clashlink")
Copy-Item -Path "version.json" -Destination $LinuxArmDir
Copy-Item -Path "upgrade.sh" -Destination $LinuxArmDir
Copy-Item -Path "docker-*.sh" -Destination $LinuxArmDir
Copy-Item -Path "Dockerfile" -Destination $LinuxArmDir
Copy-Item -Path "docker-compose.yml" -Destination $LinuxArmDir
Copy-Item -Path "README.md" -Destination $LinuxArmDir
Copy-Item -Path "DEPLOY.md" -Destination $LinuxArmDir
Copy-Item -Path "DOCKER.md" -Destination $LinuxArmDir
Copy-Item -Path "env.example" -Destination $LinuxArmDir

# Verify package contents
Write-Host "Verifying package contents..." -ForegroundColor Blue
$X64BackendSize = (Get-Item (Join-Path $LinuxDir "backend/clashlink")).Length
$ArmBackendSize = (Get-Item (Join-Path $LinuxArmDir "backend/clashlink")).Length

Write-Host "x64 package backend size: $([math]::Round($X64BackendSize/1MB,2)) MB"
Write-Host "ARM64 package backend size: $([math]::Round($ArmBackendSize/1MB,2)) MB"

# Create archives
Write-Host "Creating archives..." -ForegroundColor Blue

Compress-Archive -Path (Join-Path $BuildDir "clashlink-linux-x64") -DestinationPath (Join-Path $DistDir "clashlink-linux-$Version.zip") -Force
Compress-Archive -Path (Join-Path $BuildDir "clashlink-linux-arm64-pkg") -DestinationPath (Join-Path $DistDir "clashlink-linux-arm64-$Version.zip") -Force

# Generate checksums
Write-Host "Generating checksums..." -ForegroundColor Blue
Set-Location $DistDir
$Files = Get-ChildItem -File -Name "*.zip"
$ChecksumContent = @()

foreach ($File in $Files) {
    $Hash = Get-FileHash -Path $File -Algorithm SHA256
    $ChecksumContent += "$($Hash.Hash.ToLower())  $File"
}

$ChecksumContent | Set-Content -Path "checksums.txt" -Encoding ASCII
Set-Location ..

Write-Host
Write-Host "Build completed!" -ForegroundColor Green
Write-Host "Files created in $DistDir/:"
Get-ChildItem -Path $DistDir | Format-Table Name, @{Label="Size(MB)";Expression={[math]::Round($_.Length/1MB,2)}}, LastWriteTime