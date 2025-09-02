# ClashLink v0.1.0.5 Docker Build and Push Script
# Execute when network is stable

$VERSION = "v0.1.0.5"
$IMAGE_NAME = "uttogg/clashlink"
$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$GIT_COMMIT = "final-release"

Write-Host "🚀 Building ClashLink $VERSION Docker Image" -ForegroundColor Green
Write-Host "Image Name: $IMAGE_NAME"
Write-Host "Build Time: $BUILD_TIME"
Write-Host

# Check Docker
try {
    docker info | Out-Null
    Write-Host "✅ Docker is running" -ForegroundColor Green
}
catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop" -ForegroundColor Red
    exit 1
}

Write-Host "🔨 Building image..." -ForegroundColor Blue
try {
    docker build `
        -t "$IMAGE_NAME`:$VERSION" `
        -t "$IMAGE_NAME`:latest" `
        --build-arg VERSION="$VERSION" `
        --build-arg BUILD_TIME="$BUILD_TIME" `
        --build-arg GIT_COMMIT="$GIT_COMMIT" `
        .
    
    Write-Host "✅ Image build successful" -ForegroundColor Green
    
    Write-Host "📤 Pushing to Docker Hub..." -ForegroundColor Blue
    docker push "$IMAGE_NAME`:$VERSION"
    docker push "$IMAGE_NAME`:latest"
    
    Write-Host "✅ Image push successful" -ForegroundColor Green
    Write-Host
    Write-Host "🎉 ClashLink $VERSION released to Docker Hub!" -ForegroundColor Green
    Write-Host "🐳 Image URL: https://hub.docker.com/r/$IMAGE_NAME" -ForegroundColor Cyan
    Write-Host
    Write-Host "📋 Users can deploy with:" -ForegroundColor Yellow
    Write-Host "  docker pull $IMAGE_NAME`:latest"
    Write-Host "  docker-compose up -d"
}
catch {
    Write-Host "❌ Build or push failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
