#!/bin/bash

# ClashLink Docker 一键安装脚本
# 自动安装 Docker 并部署 ClashLink 应用

set -e

# ===========================================
# 配置变量
# ===========================================
GITHUB_REPO="your-username/clashlink"
VERSION="latest"
INSTALL_DIR="/opt/clashlink"
DATA_DIR="$INSTALL_DIR/data"
SERVICE_PORT="8080"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🐳 ClashLink Docker 一键安装脚本${NC}"
echo

# ===========================================
# 检查系统
# ===========================================
check_system() {
    echo -e "${BLUE}🔍 检查系统环境...${NC}"
    
    # 检查操作系统
    if [[ ! -f /etc/os-release ]]; then
        echo -e "${RED}❌ 不支持的操作系统${NC}"
        exit 1
    fi
    
    source /etc/os-release
    echo "操作系统: $PRETTY_NAME"
    
    # 检查架构
    ARCH=$(uname -m)
    echo "系统架构: $ARCH"
    
    if [[ "$ARCH" != "x86_64" && "$ARCH" != "aarch64" ]]; then
        echo -e "${RED}❌ 不支持的系统架构: $ARCH${NC}"
        exit 1
    fi
    
    # 检查权限
    if [[ $EUID -eq 0 ]]; then
        echo -e "${YELLOW}⚠️ 检测到 root 用户，建议使用普通用户安装${NC}"
    fi
    
    echo "✅ 系统检查通过"
}

# ===========================================
# 安装 Docker
# ===========================================
install_docker() {
    echo -e "${BLUE}🐳 安装 Docker...${NC}"
    
    # 检查 Docker 是否已安装
    if command -v docker &> /dev/null; then
        echo "✅ Docker 已安装"
        return 0
    fi
    
    # 安装 Docker
    echo "下载并安装 Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    
    # 启动 Docker 服务
    sudo systemctl enable docker
    sudo systemctl start docker
    
    # 将当前用户添加到 docker 组
    sudo usermod -aG docker $USER
    
    echo "✅ Docker 安装完成"
    echo -e "${YELLOW}⚠️ 请重新登录或执行 'newgrp docker' 以应用组权限${NC}"
}

# ===========================================
# 安装 Docker Compose
# ===========================================
install_docker_compose() {
    echo -e "${BLUE}🔧 安装 Docker Compose...${NC}"
    
    # 检查是否已安装
    if command -v docker-compose &> /dev/null; then
        echo "✅ Docker Compose 已安装"
        return 0
    fi
    
    # 检查 Docker Compose V2
    if docker compose version &> /dev/null; then
        echo "✅ Docker Compose V2 已安装"
        return 0
    fi
    
    # 安装 Docker Compose
    COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep -oP '"tag_name": "\K(.*)(?=")')
    
    sudo curl -L "https://github.com/docker/compose/releases/download/$COMPOSE_VERSION/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    
    sudo chmod +x /usr/local/bin/docker-compose
    
    echo "✅ Docker Compose 安装完成"
}

# ===========================================
# 下载项目文件
# ===========================================
download_project() {
    echo -e "${BLUE}📥 下载 ClashLink 项目...${NC}"
    
    # 创建安装目录
    sudo mkdir -p "$INSTALL_DIR"
    sudo chown $USER:$USER "$INSTALL_DIR"
    
    cd "$INSTALL_DIR"
    
    # 下载项目文件
    if command -v git &> /dev/null; then
        echo "使用 Git 克隆项目..."
        git clone "https://github.com/$GITHUB_REPO.git" .
    else
        echo "下载项目压缩包..."
        wget "https://github.com/$GITHUB_REPO/archive/main.zip" -O clashlink.zip
        unzip clashlink.zip
        mv clashlink-main/* .
        rm -rf clashlink-main clashlink.zip
    fi
    
    echo "✅ 项目文件下载完成"
}

# ===========================================
# 配置应用
# ===========================================
configure_app() {
    echo -e "${BLUE}⚙️ 配置应用...${NC}"
    
    cd "$INSTALL_DIR"
    
    # 复制环境变量配置
    if [[ -f "env.example" ]]; then
        cp env.example .env
        echo "✅ 环境配置文件已创建"
    fi
    
    # 生成随机 JWT 密钥
    JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | base64)
    
    # 更新环境变量
    if [[ -f ".env" ]]; then
        sed -i "s/change-this-secret-in-production-environment/$JWT_SECRET/" .env
        sed -i "s/your-username\/clashlink/$GITHUB_REPO/" .env
        echo "✅ JWT 密钥已生成"
    fi
    
    # 设置脚本权限
    chmod +x docker-build.sh docker-run.sh upgrade.sh build.sh
    
    echo "✅ 应用配置完成"
}

# ===========================================
# 部署应用
# ===========================================
deploy_app() {
    echo -e "${BLUE}🚀 部署 ClashLink 应用...${NC}"
    
    cd "$INSTALL_DIR"
    
    # 构建 Docker 镜像
    echo "构建 Docker 镜像..."
    ./docker-build.sh "$VERSION"
    
    # 创建数据目录
    echo "创建数据目录..."
    mkdir -p "$DATA_DIR"/{database,subscriptions,logs}
    chmod 755 "$DATA_DIR"/{database,subscriptions,logs}
    
    # 启动应用
    echo "启动应用容器..."
    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
        # 使用 Docker Compose
        docker-compose up -d
    else
        # 使用运行脚本
        ./docker-run.sh start
    fi
    
    echo "✅ 应用部署完成"
}

# ===========================================
# 验证部署
# ===========================================
verify_deployment() {
    echo -e "${BLUE}✅ 验证部署...${NC}"
    
    # 等待服务启动
    echo "等待服务启动..."
    sleep 10
    
    # 检查容器状态
    if docker ps | grep -q clashlink; then
        echo "✅ 容器运行正常"
    else
        echo -e "${RED}❌ 容器未运行${NC}"
        echo "查看容器日志:"
        docker logs clashlink-app
        return 1
    fi
    
    # 检查服务响应
    for i in {1..10}; do
        if curl -s -f "http://localhost:$SERVICE_PORT/api/version" > /dev/null; then
            echo "✅ 服务响应正常"
            break
        else
            echo "等待服务启动... ($i/10)"
            sleep 3
        fi
        
        if [[ $i -eq 10 ]]; then
            echo -e "${RED}❌ 服务响应超时${NC}"
            return 1
        fi
    done
    
    echo "✅ 部署验证完成"
}

# ===========================================
# 显示部署信息
# ===========================================
show_deployment_info() {
    echo
    echo -e "${GREEN}🎉 ClashLink 部署成功！${NC}"
    echo
    echo "访问信息:"
    echo "  🌐 Web 界面: http://localhost:$SERVICE_PORT"
    echo "  🌐 外网访问: http://$(curl -s ifconfig.me):$SERVICE_PORT"
    echo
    echo "管理命令:"
    echo "  📊 查看状态: cd $INSTALL_DIR && ./docker-run.sh status"
    echo "  📋 查看日志: cd $INSTALL_DIR && ./docker-run.sh logs -f"
    echo "  🔄 重启服务: cd $INSTALL_DIR && ./docker-run.sh restart"
    echo "  🛑 停止服务: cd $INSTALL_DIR && ./docker-run.sh stop"
    echo
    echo "数据位置:"
    echo "  📁 应用目录: $INSTALL_DIR"
    echo "  💾 数据目录: $DATA_DIR"
    echo "  📄 配置文件: $INSTALL_DIR/.env"
    echo
    echo "安全提醒:"
    echo "  🔒 请修改 JWT 密钥: $INSTALL_DIR/.env"
    echo "  🛡️ 建议配置防火墙规则"
    echo "  🔐 如需公网访问，建议配置 HTTPS"
    echo
}

# ===========================================
# 主函数
# ===========================================
main() {
    echo -e "${BLUE}开始安装 ClashLink...${NC}"
    echo
    
    # 检查系统
    check_system
    
    # 安装 Docker
    install_docker
    
    # 安装 Docker Compose
    install_docker_compose
    
    # 下载项目
    download_project
    
    # 配置应用
    configure_app
    
    # 部署应用
    deploy_app
    
    # 验证部署
    if verify_deployment; then
        show_deployment_info
    else
        echo -e "${RED}❌ 部署验证失败${NC}"
        echo "请检查日志并重试安装"
        exit 1
    fi
}

# 运行主函数
main "$@"
