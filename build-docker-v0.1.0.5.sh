#!/bin/bash

# ClashLink v0.1.0.5 Docker 构建和推送脚本
# 当网络稳定时执行此脚本

set -e

VERSION="v0.1.0.5"
IMAGE_NAME="uttogg/clashlink"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT="final-release"

echo "🚀 构建 ClashLink $VERSION Docker 镜像"
echo "镜像名称: $IMAGE_NAME"
echo "构建时间: $BUILD_TIME"
echo

# 检查 Docker
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker 未运行，请启动 Docker Desktop"
    exit 1
fi

echo "🔨 构建镜像..."
docker build \
    -t "$IMAGE_NAME:$VERSION" \
    -t "$IMAGE_NAME:latest" \
    --build-arg VERSION="$VERSION" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    .

if [ $? -eq 0 ]; then
    echo "✅ 镜像构建成功"
    
    echo "📤 推送到 Docker Hub..."
    docker push "$IMAGE_NAME:$VERSION"
    docker push "$IMAGE_NAME:latest"
    
    if [ $? -eq 0 ]; then
        echo "✅ 镜像推送成功"
        echo
        echo "🎉 ClashLink $VERSION 已发布到 Docker Hub!"
        echo "🐳 镜像地址: https://hub.docker.com/r/$IMAGE_NAME"
        echo
        echo "📋 用户可以使用以下命令部署:"
        echo "  docker pull $IMAGE_NAME:latest"
        echo "  docker-compose up -d"
        echo
    else
        echo "❌ 镜像推送失败"
        exit 1
    fi
else
    echo "❌ 镜像构建失败"
    exit 1
fi
