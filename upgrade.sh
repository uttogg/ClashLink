#!/bin/bash

# ClashLink 自动升级脚本
# 用法: ./upgrade.sh [选项]
# 选项:
#   --check-only    仅检查更新，不执行升级
#   --force         强制升级，即使没有检测到新版本
#   --backup-only   仅创建备份
#   --rollback      回滚到上一个版本

set -e  # 遇到错误立即退出

# ===========================================
# 配置变量
# ===========================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="/opt/clashlink"
BACKUP_DIR="/opt/clashlink_backups"
SERVICE_NAME="clashlink"
GITHUB_REPO="your-username/clashlink"  # 替换为你的GitHub仓库
CURRENT_VERSION="1.0.0"
LOG_FILE="/var/log/clashlink_upgrade.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===========================================
# 日志函数
# ===========================================
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        INFO)
            echo -e "${GREEN}[INFO]${NC} $message"
            echo "[$timestamp] [INFO] $message" >> "$LOG_FILE"
            ;;
        WARN)
            echo -e "${YELLOW}[WARN]${NC} $message"
            echo "[$timestamp] [WARN] $message" >> "$LOG_FILE"
            ;;
        ERROR)
            echo -e "${RED}[ERROR]${NC} $message"
            echo "[$timestamp] [ERROR] $message" >> "$LOG_FILE"
            ;;
        DEBUG)
            echo -e "${BLUE}[DEBUG]${NC} $message"
            echo "[$timestamp] [DEBUG] $message" >> "$LOG_FILE"
            ;;
    esac
}

# ===========================================
# 工具函数
# ===========================================
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log ERROR "此脚本需要root权限运行"
        echo "请使用: sudo $0 $*"
        exit 1
    fi
}

check_dependencies() {
    local deps=("curl" "jq" "systemctl" "git")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log ERROR "缺少依赖: ${missing[*]}"
        log INFO "请安装缺少的依赖："
        echo "  Ubuntu/Debian: sudo apt install ${missing[*]}"
        echo "  CentOS/RHEL: sudo yum install ${missing[*]}"
        exit 1
    fi
}

get_latest_version() {
    local api_url="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    
    log DEBUG "检查最新版本: $api_url"
    
    local response=$(curl -s -H "Accept: application/vnd.github.v3+json" "$api_url")
    
    if [[ -z "$response" ]]; then
        log ERROR "无法获取版本信息"
        return 1
    fi
    
    local latest_version=$(echo "$response" | jq -r '.tag_name')
    local release_url=$(echo "$response" | jq -r '.html_url')
    local release_notes=$(echo "$response" | jq -r '.body')
    
    if [[ "$latest_version" == "null" ]]; then
        log ERROR "解析版本信息失败"
        return 1
    fi
    
    echo "$latest_version|$release_url|$release_notes"
}

compare_versions() {
    local current=$1
    local latest=$2
    
    # 去除 v 前缀
    current=${current#v}
    latest=${latest#v}
    
    if [[ "$current" == "$latest" ]]; then
        return 1  # 版本相同
    else
        return 0  # 版本不同（假设有更新）
    fi
}

create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/clashlink_$timestamp"
    
    log INFO "创建备份: $backup_path"
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR"
    
    # 停止服务
    log INFO "停止服务..."
    systemctl stop "$SERVICE_NAME" || log WARN "服务停止失败"
    
    # 备份应用目录
    if [[ -d "$APP_DIR" ]]; then
        cp -r "$APP_DIR" "$backup_path"
        log INFO "应用目录备份完成"
    fi
    
    # 备份数据库
    if [[ -f "$APP_DIR/backend/data.db" ]]; then
        cp "$APP_DIR/backend/data.db" "$BACKUP_DIR/data_$timestamp.db"
        log INFO "数据库备份完成"
    fi
    
    # 备份系统配置
    if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
        cp "/etc/systemd/system/$SERVICE_NAME.service" "$BACKUP_DIR/service_$timestamp.service"
        log INFO "系统服务配置备份完成"
    fi
    
    # 清理旧备份（保留最近5个）
    cd "$BACKUP_DIR"
    ls -t clashlink_* 2>/dev/null | tail -n +6 | xargs -r rm -rf
    ls -t data_*.db 2>/dev/null | tail -n +6 | xargs -r rm -f
    ls -t service_*.service 2>/dev/null | tail -n +6 | xargs -r rm -f
    
    echo "$backup_path"
}

download_and_upgrade() {
    local latest_version=$1
    local temp_dir=$(mktemp -d)
    
    log INFO "下载新版本 $latest_version..."
    
    # 检测系统架构
    local arch=$(uname -m)
    local download_file=""
    
    case $arch in
        x86_64)
            download_file="clashlink-linux.tar.gz"
            ;;
        aarch64|arm64)
            download_file="clashlink-linux-arm64.tar.gz"
            ;;
        *)
            log WARN "不支持的架构: $arch，尝试使用 x86_64 版本"
            download_file="clashlink-linux.tar.gz"
            ;;
    esac
    
    # 构建下载URL
    local download_url="https://github.com/$GITHUB_REPO/releases/download/$latest_version/$download_file"
    
    # 下载新版本
    if ! curl -L -o "$temp_dir/clashlink.tar.gz" "$download_url"; then
        log WARN "下载预编译包失败，尝试从源码编译..."
        compile_from_source "$latest_version" "$temp_dir"
    else
        log INFO "解压下载的包..."
        tar -xzf "$temp_dir/clashlink.tar.gz" -C "$temp_dir"
    fi
    
    # 更新应用文件
    log INFO "更新应用文件..."
    
    # 更新可执行文件
    if [[ -f "$temp_dir/backend/clashlink" ]]; then
        cp "$temp_dir/backend/clashlink" "$APP_DIR/backend/"
        chown clashlink:clashlink "$APP_DIR/backend/clashlink"
        chmod +x "$APP_DIR/backend/clashlink"
    fi
    
    # 更新前端文件
    if [[ -d "$temp_dir/frontend" ]]; then
        cp -r "$temp_dir/frontend/"* "$APP_DIR/frontend/"
        chown -R clashlink:clashlink "$APP_DIR/frontend/"
    fi
    
    # 清理临时文件
    rm -rf "$temp_dir"
}

compile_from_source() {
    local version=$1
    local build_dir=$2
    
    log INFO "从源码编译版本 $version..."
    
    # 克隆仓库
    git clone "https://github.com/$GITHUB_REPO.git" "$build_dir/source"
    cd "$build_dir/source"
    
    # 切换到指定版本
    git checkout "$version"
    
    # 编译后端
    cd backend
    go mod tidy
    go build -o clashlink .
    
    # 复制文件到临时目录
    mkdir -p "$build_dir/backend" "$build_dir/frontend"
    cp clashlink "$build_dir/backend/"
    cp -r ../frontend/* "$build_dir/frontend/"
}

start_service() {
    log INFO "启动服务..."
    systemctl daemon-reload
    systemctl start "$SERVICE_NAME"
    
    # 等待服务启动
    sleep 3
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log INFO "✅ 服务启动成功"
        return 0
    else
        log ERROR "❌ 服务启动失败"
        return 1
    fi
}

verify_upgrade() {
    log INFO "验证升级结果..."
    
    # 检查服务状态
    if ! systemctl is-active --quiet "$SERVICE_NAME"; then
        log ERROR "服务未运行"
        return 1
    fi
    
    # 检查端口连通性
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200"; then
        log WARN "Web服务可能未正常响应"
    fi
    
    log INFO "✅ 升级验证完成"
    return 0
}

rollback() {
    log INFO "开始回滚..."
    
    # 找到最新的备份
    local latest_backup=$(ls -t "$BACKUP_DIR"/clashlink_* 2>/dev/null | head -n1)
    
    if [[ -z "$latest_backup" ]]; then
        log ERROR "未找到备份文件"
        return 1
    fi
    
    log INFO "从备份恢复: $latest_backup"
    
    # 停止服务
    systemctl stop "$SERVICE_NAME"
    
    # 恢复文件
    rm -rf "$APP_DIR"
    cp -r "$latest_backup" "$APP_DIR"
    
    # 启动服务
    start_service
    
    if [[ $? -eq 0 ]]; then
        log INFO "✅ 回滚成功"
    else
        log ERROR "❌ 回滚失败"
        return 1
    fi
}

show_help() {
    cat << EOF
ClashLink 自动升级脚本

用法: $0 [选项]

选项:
    --check-only    仅检查更新，不执行升级
    --force         强制升级，即使没有检测到新版本
    --backup-only   仅创建备份
    --rollback      回滚到上一个版本
    --help          显示此帮助信息

示例:
    $0                  # 检查并升级到最新版本
    $0 --check-only     # 仅检查是否有新版本
    $0 --force          # 强制升级
    $0 --rollback       # 回滚到上一个版本

EOF
}

# ===========================================
# 主逻辑
# ===========================================
main() {
    local check_only=false
    local force_upgrade=false
    local backup_only=false
    local do_rollback=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-only)
                check_only=true
                shift
                ;;
            --force)
                force_upgrade=true
                shift
                ;;
            --backup-only)
                backup_only=true
                shift
                ;;
            --rollback)
                do_rollback=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log ERROR "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log INFO "ClashLink 升级脚本启动"
    log INFO "当前版本: $CURRENT_VERSION"
    
    # 检查权限和依赖
    check_root
    check_dependencies
    
    # 创建日志目录
    mkdir -p "$(dirname "$LOG_FILE")"
    
    # 处理回滚
    if [[ "$do_rollback" == true ]]; then
        rollback
        exit $?
    fi
    
    # 处理仅备份
    if [[ "$backup_only" == true ]]; then
        create_backup
        log INFO "备份完成"
        exit 0
    fi
    
    # 获取最新版本信息
    log INFO "检查最新版本..."
    local version_info
    if ! version_info=$(get_latest_version); then
        log ERROR "获取版本信息失败"
        exit 1
    fi
    
    IFS='|' read -r latest_version release_url release_notes <<< "$version_info"
    
    log INFO "最新版本: $latest_version"
    log INFO "发布页面: $release_url"
    
    # 仅检查模式
    if [[ "$check_only" == true ]]; then
        if compare_versions "$CURRENT_VERSION" "$latest_version"; then
            log INFO "🎉 发现新版本: $latest_version"
            echo "当前版本: $CURRENT_VERSION"
            echo "最新版本: $latest_version"
            echo "发布页面: $release_url"
            exit 0
        else
            log INFO "✅ 当前已是最新版本"
            exit 0
        fi
    fi
    
    # 检查是否需要升级
    if [[ "$force_upgrade" != true ]] && ! compare_versions "$CURRENT_VERSION" "$latest_version"; then
        log INFO "✅ 当前已是最新版本，无需升级"
        exit 0
    fi
    
    log INFO "🚀 开始升级到版本 $latest_version"
    
    # 创建备份
    local backup_path
    backup_path=$(create_backup)
    
    # 执行升级
    if download_and_upgrade "$latest_version"; then
        log INFO "文件更新完成"
    else
        log ERROR "文件更新失败，开始回滚"
        rollback
        exit 1
    fi
    
    # 启动服务
    if start_service; then
        log INFO "服务启动成功"
    else
        log ERROR "服务启动失败，开始回滚"
        rollback
        exit 1
    fi
    
    # 验证升级
    if verify_upgrade; then
        log INFO "🎉 升级成功完成！"
        log INFO "当前版本: $CURRENT_VERSION → $latest_version"
        log INFO "备份位置: $backup_path"
    else
        log WARN "升级完成但验证失败，请检查服务状态"
    fi
}

# 运行主函数
main "$@"
