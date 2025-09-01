#!/bin/bash

# ClashLink Docker 运行脚本
# 用于快速启动、停止和管理 ClashLink Docker 容器

set -e

# ===========================================
# 配置变量
# ===========================================
IMAGE_NAME="clashlink"
CONTAINER_NAME="clashlink-app"
VERSION=${VERSION:-"latest"}
HOST_PORT=${HOST_PORT:-8080}
DATA_DIR=${DATA_DIR:-"$(pwd)/data"}

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# ===========================================
# 工具函数
# ===========================================
show_help() {
    cat << EOF
ClashLink Docker 运行脚本

用法: $0 <命令> [选项]

命令:
    start       启动 ClashLink 容器
    stop        停止 ClashLink 容器
    restart     重启 ClashLink 容器
    status      查看容器状态
    logs        查看容器日志
    shell       进入容器 shell
    clean       清理容器和数据
    update      更新并重启容器
    backup      备份数据
    restore     恢复数据

选项:
    -p, --port <port>     指定宿主机端口 (默认: 8080)
    -d, --data <dir>      指定数据目录 (默认: ./data)
    -v, --version <ver>   指定镜像版本 (默认: latest)
    -h, --help            显示此帮助信息

环境变量:
    HOST_PORT            宿主机端口
    DATA_DIR             数据目录路径
    VERSION              镜像版本
    JWT_SECRET           JWT 密钥

示例:
    $0 start                    # 启动容器
    $0 start -p 9090           # 在端口 9090 启动
    $0 logs -f                 # 跟踪日志
    $0 backup                  # 备份数据

EOF
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker 未安装${NC}"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo -e "${RED}❌ Docker 服务未运行${NC}"
        exit 1
    fi
}

create_data_dirs() {
    echo -e "${BLUE}📁 创建数据目录...${NC}"
    mkdir -p "$DATA_DIR/database"
    mkdir -p "$DATA_DIR/subscriptions"
    mkdir -p "$DATA_DIR/logs"
    
    # 设置权限
    chmod 755 "$DATA_DIR"
    chmod 755 "$DATA_DIR/database"
    chmod 755 "$DATA_DIR/subscriptions"
    chmod 755 "$DATA_DIR/logs"
    
    echo "✅ 数据目录已创建: $DATA_DIR"
}

container_exists() {
    docker ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"
}

container_running() {
    docker ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"
}

start_container() {
    echo -e "${BLUE}🚀 启动 ClashLink 容器...${NC}"
    
    # 检查容器是否已存在
    if container_exists; then
        if container_running; then
            echo -e "${YELLOW}⚠️ 容器已在运行${NC}"
            return 0
        else
            echo "启动现有容器..."
            docker start "$CONTAINER_NAME"
        fi
    else
        # 创建数据目录
        create_data_dirs
        
        # 启动新容器
        echo "创建并启动新容器..."
        docker run -d \
            --name "$CONTAINER_NAME" \
            --restart unless-stopped \
            -p "$HOST_PORT:8080" \
            -v "$DATA_DIR/database:/app/backend" \
            -v "$DATA_DIR/subscriptions:/app/subscriptions" \
            -v "$DATA_DIR/logs:/app/logs" \
            -e TZ=Asia/Shanghai \
            -e JWT_SECRET="${JWT_SECRET:-change-this-secret-in-production}" \
            -e LOG_LEVEL="${LOG_LEVEL:-info}" \
            --health-cmd="curl -f http://localhost:8080/api/version || exit 1" \
            --health-interval=30s \
            --health-timeout=10s \
            --health-retries=3 \
            "$IMAGE_NAME:$VERSION"
    fi
    
    # 等待容器启动
    echo "等待容器启动..."
    sleep 5
    
    # 检查容器状态
    if container_running; then
        echo -e "${GREEN}✅ 容器启动成功${NC}"
        echo "访问地址: http://localhost:$HOST_PORT"
        
        # 显示健康检查状态
        echo "健康检查状态:"
        docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "未知"
    else
        echo -e "${RED}❌ 容器启动失败${NC}"
        echo "查看日志:"
        docker logs "$CONTAINER_NAME"
        exit 1
    fi
}

stop_container() {
    echo -e "${BLUE}🛑 停止 ClashLink 容器...${NC}"
    
    if container_running; then
        docker stop "$CONTAINER_NAME"
        echo -e "${GREEN}✅ 容器已停止${NC}"
    else
        echo -e "${YELLOW}⚠️ 容器未在运行${NC}"
    fi
}

restart_container() {
    echo -e "${BLUE}🔄 重启 ClashLink 容器...${NC}"
    
    if container_exists; then
        docker restart "$CONTAINER_NAME"
        echo -e "${GREEN}✅ 容器已重启${NC}"
        
        # 等待启动
        sleep 5
        echo "访问地址: http://localhost:$HOST_PORT"
    else
        echo "容器不存在，创建新容器..."
        start_container
    fi
}

show_status() {
    echo -e "${BLUE}📊 容器状态:${NC}"
    
    if container_exists; then
        # 显示容器信息
        docker ps -a --filter "name=$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        
        echo
        echo "健康状态:"
        docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "未配置健康检查"
        
        echo
        echo "资源使用:"
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" "$CONTAINER_NAME"
    else
        echo "容器不存在"
    fi
}

show_logs() {
    echo -e "${BLUE}📋 查看容器日志:${NC}"
    
    if container_exists; then
        # 解析额外参数
        shift  # 移除 'logs' 命令
        docker logs "$@" "$CONTAINER_NAME"
    else
        echo "容器不存在"
        exit 1
    fi
}

enter_shell() {
    echo -e "${BLUE}🐚 进入容器 shell:${NC}"
    
    if container_running; then
        docker exec -it "$CONTAINER_NAME" /bin/sh
    else
        echo "容器未运行，无法进入 shell"
        exit 1
    fi
}

clean_container() {
    echo -e "${BLUE}🧹 清理容器和数据...${NC}"
    
    read -p "⚠️ 这将删除容器和所有数据，确定继续? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # 停止并删除容器
        if container_exists; then
            docker stop "$CONTAINER_NAME" 2>/dev/null || true
            docker rm "$CONTAINER_NAME"
        fi
        
        # 删除数据目录
        if [[ -d "$DATA_DIR" ]]; then
            rm -rf "$DATA_DIR"
            echo "✅ 数据目录已删除: $DATA_DIR"
        fi
        
        echo -e "${GREEN}✅ 清理完成${NC}"
    else
        echo "取消清理操作"
    fi
}

update_container() {
    echo -e "${BLUE}🔄 更新 ClashLink 容器...${NC}"
    
    # 拉取最新镜像
    echo "拉取最新镜像..."
    docker pull "$IMAGE_NAME:$VERSION"
    
    # 备份当前数据
    backup_data
    
    # 停止旧容器
    if container_running; then
        docker stop "$CONTAINER_NAME"
    fi
    
    # 删除旧容器
    if container_exists; then
        docker rm "$CONTAINER_NAME"
    fi
    
    # 启动新容器
    start_container
    
    echo -e "${GREEN}✅ 容器更新完成${NC}"
}

backup_data() {
    echo -e "${BLUE}💾 备份数据...${NC}"
    
    if [[ ! -d "$DATA_DIR" ]]; then
        echo "数据目录不存在，无需备份"
        return
    fi
    
    BACKUP_FILE="clashlink-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
    
    tar -czf "$BACKUP_FILE" -C "$DATA_DIR" .
    
    echo "✅ 数据备份完成: $BACKUP_FILE"
}

restore_data() {
    echo -e "${BLUE}📤 恢复数据...${NC}"
    
    # 列出可用的备份文件
    echo "可用的备份文件:"
    ls -la clashlink-backup-*.tar.gz 2>/dev/null || {
        echo "未找到备份文件"
        exit 1
    }
    
    read -p "请输入要恢复的备份文件名: " BACKUP_FILE
    
    if [[ ! -f "$BACKUP_FILE" ]]; then
        echo "备份文件不存在: $BACKUP_FILE"
        exit 1
    fi
    
    # 停止容器
    if container_running; then
        docker stop "$CONTAINER_NAME"
    fi
    
    # 恢复数据
    mkdir -p "$DATA_DIR"
    tar -xzf "$BACKUP_FILE" -C "$DATA_DIR"
    
    # 重启容器
    if container_exists; then
        docker start "$CONTAINER_NAME"
    fi
    
    echo "✅ 数据恢复完成"
}

# ===========================================
# 参数解析
# ===========================================
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--port)
                HOST_PORT="$2"
                shift 2
                ;;
            -d|--data)
                DATA_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
}

# ===========================================
# 主逻辑
# ===========================================
main() {
    local command=$1
    shift || true
    
    # 解析参数
    parse_args "$@"
    
    # 检查 Docker
    check_docker
    
    case $command in
        start)
            start_container
            ;;
        stop)
            stop_container
            ;;
        restart)
            restart_container
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$@"
            ;;
        shell)
            enter_shell
            ;;
        clean)
            clean_container
            ;;
        update)
            update_container
            ;;
        backup)
            backup_data
            ;;
        restore)
            restore_data
            ;;
        *)
            echo "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 检查是否提供了命令
if [[ $# -eq 0 ]]; then
    show_help
    exit 1
fi

# 运行主函数
main "$@"
