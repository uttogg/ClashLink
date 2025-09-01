#!/bin/bash

# ClashLink Docker 构建脚本
# 用于构建和管理 Docker 镜像

set -e

# ===========================================
# 配置变量
# ===========================================
IMAGE_NAME="clashlink"
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
REGISTRY=${DOCKER_REGISTRY:-""}

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🐳 ClashLink Docker 构建脚本${NC}"
echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo "Git 提交: $GIT_COMMIT"
echo

# ===========================================
# 检查环境
# ===========================================
echo -e "${BLUE}🔍 检查环境...${NC}"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装或不在 PATH 中"
    echo "请安装 Docker: https://docs.docker.com/engine/install/"
    exit 1
fi

# 检查 Docker 服务是否运行
if ! docker info &> /dev/null; then
    echo "❌ Docker 服务未运行"
    echo "请启动 Docker 服务: sudo systemctl start docker"
    exit 1
fi

echo "✅ Docker 环境检查通过"

# ===========================================
# 构建镜像
# ===========================================
echo -e "${BLUE}🔨 构建 Docker 镜像...${NC}"

# 构建参数
BUILD_ARGS="--build-arg VERSION=$VERSION --build-arg BUILD_TIME=$BUILD_TIME --build-arg GIT_COMMIT=$GIT_COMMIT"

# 构建多架构镜像标签
TAGS=""
if [[ -n "$REGISTRY" ]]; then
    TAGS="$TAGS -t $REGISTRY/$IMAGE_NAME:$VERSION -t $REGISTRY/$IMAGE_NAME:latest"
else
    TAGS="$TAGS -t $IMAGE_NAME:$VERSION -t $IMAGE_NAME:latest"
fi

# 执行构建
echo "构建命令: docker build $BUILD_ARGS $TAGS ."
docker build $BUILD_ARGS $TAGS .

echo "✅ Docker 镜像构建完成"

# ===========================================
# 显示构建结果
# ===========================================
echo -e "${BLUE}📊 构建结果:${NC}"

# 显示镜像信息
if [[ -n "$REGISTRY" ]]; then
    docker images "$REGISTRY/$IMAGE_NAME"
else
    docker images "$IMAGE_NAME"
fi

# ===========================================
# 镜像测试
# ===========================================
echo -e "${BLUE}🧪 测试镜像...${NC}"

# 创建临时容器进行测试
TEST_CONTAINER="clashlink-test-$$"

echo "启动测试容器..."
docker run -d --name "$TEST_CONTAINER" \
    -p 8081:8080 \
    --health-interval=10s \
    --health-timeout=5s \
    --health-retries=3 \
    $IMAGE_NAME:$VERSION

# 等待容器启动
echo "等待容器启动..."
sleep 10

# 检查容器状态
if docker ps | grep -q "$TEST_CONTAINER"; then
    echo "✅ 容器启动成功"
    
    # 测试 API 端点
    if curl -s -f http://localhost:8081/api/version > /dev/null; then
        echo "✅ API 端点响应正常"
    else
        echo "⚠️ API 端点可能有问题"
    fi
    
    # 显示容器日志
    echo -e "${YELLOW}📋 容器日志 (最后10行):${NC}"
    docker logs --tail 10 "$TEST_CONTAINER"
    
else
    echo "❌ 容器启动失败"
    docker logs "$TEST_CONTAINER"
fi

# 清理测试容器
echo "清理测试容器..."
docker stop "$TEST_CONTAINER" >/dev/null 2>&1 || true
docker rm "$TEST_CONTAINER" >/dev/null 2>&1 || true

# ===========================================
# 推送镜像（如果指定了仓库）
# ===========================================
if [[ -n "$REGISTRY" ]]; then
    echo -e "${BLUE}📤 推送镜像到仓库...${NC}"
    
    read -p "是否推送镜像到 $REGISTRY? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker push "$REGISTRY/$IMAGE_NAME:$VERSION"
        docker push "$REGISTRY/$IMAGE_NAME:latest"
        echo "✅ 镜像推送完成"
    else
        echo "跳过镜像推送"
    fi
fi

# ===========================================
# 构建总结
# ===========================================
echo
echo -e "${GREEN}🎉 Docker 构建完成！${NC}"
echo
echo "镜像标签："
if [[ -n "$REGISTRY" ]]; then
    echo "  - $REGISTRY/$IMAGE_NAME:$VERSION"
    echo "  - $REGISTRY/$IMAGE_NAME:latest"
else
    echo "  - $IMAGE_NAME:$VERSION"
    echo "  - $IMAGE_NAME:latest"
fi
echo
echo "运行命令："
echo "  # 简单运行"
echo "  docker run -d -p 8080:8080 --name clashlink $IMAGE_NAME:$VERSION"
echo
echo "  # 使用数据持久化"
echo "  docker run -d -p 8080:8080 --name clashlink \\"
echo "    -v \$(pwd)/data/database:/app/backend \\"
echo "    -v \$(pwd)/data/subscriptions:/app/subscriptions \\"
echo "    $IMAGE_NAME:$VERSION"
echo
echo "  # 使用 Docker Compose"
echo "  docker-compose up -d"
echo
