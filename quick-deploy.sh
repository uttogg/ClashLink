#!/bin/bash

# ClashLink 快速部署脚本
# 从 Docker Hub 拉取镜像并快速启动

set -e

# 配置
VERSION="v0.1.0.1"
DOCKER_IMAGE="uttogg/clashlink"
CONTAINER_NAME="clashlink-app"
HOST_PORT="8080"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 ClashLink 快速部署脚本${NC}"
echo -e "${BLUE}从 Docker Hub 拉取镜像: ${DOCKER_IMAGE}:${VERSION}${NC}"
echo

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️ Docker 未安装，正在安装...${NC}"
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker $USER
    echo -e "${YELLOW}请重新登录或执行 'newgrp docker' 后再运行此脚本${NC}"
    exit 1
fi

# 检查 Docker 服务
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}⚠️ Docker 服务未运行，正在启动...${NC}"
    sudo systemctl start docker
fi

# 停止现有容器（如果存在）
if docker ps -a | grep -q "$CONTAINER_NAME"; then
    echo -e "${BLUE}🛑 停止现有容器...${NC}"
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

# 拉取最新镜像
echo -e "${BLUE}📥 拉取 Docker 镜像...${NC}"
docker pull "$DOCKER_IMAGE:$VERSION"
docker tag "$DOCKER_IMAGE:$VERSION" "$DOCKER_IMAGE:latest"

# 创建数据目录
echo -e "${BLUE}📁 创建数据目录...${NC}"
mkdir -p data/{database,subscriptions,logs}
chmod 755 data/{database,subscriptions,logs}

# 生成随机 JWT 密钥
JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -d '\n')

# 启动容器
echo -e "${BLUE}🚀 启动 ClashLink 容器...${NC}"
docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    -p "$HOST_PORT:8080" \
    -v "$(pwd)/data/database:/app/backend" \
    -v "$(pwd)/data/subscriptions:/app/subscriptions" \
    -v "$(pwd)/data/logs:/app/logs" \
    -e TZ=Asia/Shanghai \
    -e JWT_SECRET="$JWT_SECRET" \
    -e LOG_LEVEL=info \
    --health-cmd="curl -f http://localhost:8080/api/version || exit 1" \
    --health-interval=30s \
    --health-timeout=10s \
    --health-retries=3 \
    "$DOCKER_IMAGE:$VERSION"

# 等待容器启动
echo -e "${BLUE}⏳ 等待服务启动...${NC}"
sleep 5

# 检查容器状态
if docker ps | grep -q "$CONTAINER_NAME"; then
    echo -e "${GREEN}✅ ClashLink 启动成功！${NC}"
    echo
    echo "🌐 访问地址: http://localhost:$HOST_PORT"
    echo "🔑 JWT 密钥: $JWT_SECRET"
    echo
    echo "📊 管理命令:"
    echo "  docker logs $CONTAINER_NAME -f     # 查看日志"
    echo "  docker restart $CONTAINER_NAME     # 重启服务"
    echo "  docker stop $CONTAINER_NAME        # 停止服务"
    echo
    echo "📁 数据目录: $(pwd)/data/"
    echo "💾 数据库: $(pwd)/data/database/data.db"
    echo "📋 订阅: $(pwd)/data/subscriptions/"
    echo
    echo -e "${YELLOW}🔒 安全提醒:${NC}"
    echo "  - 首次访问需要设置管理员账号"
    echo "  - 请保存好 JWT 密钥"
    echo "  - 生产环境建议配置 HTTPS"
    
    # 显示健康检查状态
    echo
    echo "🏥 健康检查:"
    sleep 5
    HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "unknown")
    echo "  状态: $HEALTH_STATUS"
    
else
    echo -e "${RED}❌ 容器启动失败${NC}"
    echo "查看错误日志:"
    docker logs "$CONTAINER_NAME"
    exit 1
fi
