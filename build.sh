#!/bin/bash

# ClashLink 构建脚本
# 用于构建发布版本并更新版本信息

set -e

# ===========================================
# 配置变量
# ===========================================
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

BUILD_DIR="build"
DIST_DIR="dist"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}🚀 ClashLink 构建脚本${NC}"
echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo "Git 提交: $GIT_COMMIT"
echo "Go 版本: $GO_VERSION"
echo

# ===========================================
# 更新版本配置文件
# ===========================================
echo -e "${BLUE}📝 更新版本配置...${NC}"

# 读取现有配置
if [[ -f "version.json" ]]; then
    # 使用 jq 更新版本信息
    if command -v jq >/dev/null 2>&1; then
        jq --arg version "$VERSION" \
           --arg build_time "$BUILD_TIME" \
           --arg git_commit "$GIT_COMMIT" \
           --arg go_version "$GO_VERSION" \
           '.version = $version | .build_time = $build_time | .git_commit = $git_commit | .go_version = $go_version' \
           version.json > version.json.tmp && mv version.json.tmp version.json
    else
        echo "警告: jq 未安装，无法自动更新版本配置"
    fi
else
    # 创建新的版本配置文件
    cat > version.json << EOF
{
    "name": "ClashLink",
    "version": "$VERSION",
    "build_time": "$BUILD_TIME",
    "git_commit": "$GIT_COMMIT",
    "go_version": "$GO_VERSION",
    "github_repo": "your-username/clashlink",
    "homepage": "https://github.com/your-username/clashlink",
    "description": "VLESS/VMess 到 Clash 订阅转换工具",
    "features": [
        "用户认证系统",
        "节点连通性检测",
        "自动订阅生成",
        "毛玻璃界面设计",
        "自动更新检测"
    ]
}
EOF
fi

echo "✅ 版本配置已更新"

# ===========================================
# 清理和准备构建目录
# ===========================================
echo -e "${BLUE}🧹 清理构建目录...${NC}"
rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# ===========================================
# 构建后端
# ===========================================
echo -e "${BLUE}🔨 构建后端...${NC}"
cd backend

# 下载依赖
go mod tidy

# 构建 Linux 版本
echo "构建 Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" -o "../$BUILD_DIR/clashlink-linux-amd64" .

echo "构建 Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" -o "../$BUILD_DIR/clashlink-linux-arm64" .

cd ..
echo "✅ 后端构建完成"

# ===========================================
# 打包发布文件
# ===========================================
echo -e "${BLUE}📦 打包发布文件...${NC}"

# Linux 版本
echo "打包 Linux 版本..."
mkdir -p "$BUILD_DIR/clashlink-linux"
cp -r frontend "$BUILD_DIR/clashlink-linux/"
mkdir -p "$BUILD_DIR/clashlink-linux/backend"
cp "$BUILD_DIR/clashlink-linux-amd64" "$BUILD_DIR/clashlink-linux/backend/clashlink"
cp version.json "$BUILD_DIR/clashlink-linux/"
cp upgrade.sh "$BUILD_DIR/clashlink-linux/"
chmod +x "$BUILD_DIR/clashlink-linux/backend/clashlink"
chmod +x "$BUILD_DIR/clashlink-linux/upgrade.sh"

cd "$BUILD_DIR"
tar -czf "../$DIST_DIR/clashlink-linux.tar.gz" clashlink-linux/
cd ..

# Linux ARM64 版本
echo "打包 Linux ARM64 版本..."
mkdir -p "$BUILD_DIR/clashlink-linux-arm64"
cp -r frontend "$BUILD_DIR/clashlink-linux-arm64/"
mkdir -p "$BUILD_DIR/clashlink-linux-arm64/backend"
cp "$BUILD_DIR/clashlink-linux-arm64" "$BUILD_DIR/clashlink-linux-arm64/backend/clashlink"
cp version.json "$BUILD_DIR/clashlink-linux-arm64/"
cp upgrade.sh "$BUILD_DIR/clashlink-linux-arm64/"
chmod +x "$BUILD_DIR/clashlink-linux-arm64/backend/clashlink"
chmod +x "$BUILD_DIR/clashlink-linux-arm64/upgrade.sh"

cd "$BUILD_DIR"
tar -czf "../$DIST_DIR/clashlink-linux-arm64.tar.gz" clashlink-linux-arm64/
cd ..



echo "✅ 打包完成"

# ===========================================
# 生成校验和
# ===========================================
echo -e "${BLUE}🔐 生成校验和...${NC}"
cd "$DIST_DIR"

if command -v sha256sum >/dev/null 2>&1; then
    sha256sum * > checksums.txt
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 * > checksums.txt
else
    echo "警告: 无法生成校验和文件（缺少 sha256sum 或 shasum）"
fi

cd ..

# ===========================================
# 显示构建结果
# ===========================================
echo
echo -e "${GREEN}🎉 构建完成！${NC}"
echo
echo "构建产物："
ls -la "$DIST_DIR/"
echo
echo "版本信息："
echo "  版本: $VERSION"
echo "  构建时间: $BUILD_TIME"
echo "  Git 提交: $GIT_COMMIT"
echo "  Go 版本: $GO_VERSION"
echo
echo "支持的平台："
echo "  - Linux x86_64 (Intel/AMD 64位)"
echo "  - Linux ARM64 (ARM 64位，适用于树莓派等ARM设备)"
echo
echo "发布说明："
echo "  1. 将 $DIST_DIR/ 目录下的文件上传到 GitHub Releases"
echo "  2. 创建新的 Release 标签: $VERSION"
echo "  3. 更新版本说明和更新日志"
echo "  4. 发布 Release 以触发自动更新检测"
echo
echo "部署说明："
echo "  - Linux 服务器部署请参考 DEPLOY.md"
echo "  - 支持 x86_64 和 ARM64 架构的 Linux 系统"
echo "  - 升级脚本会自动检测系统架构并下载对应版本"
