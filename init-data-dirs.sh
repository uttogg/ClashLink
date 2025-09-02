#!/bin/bash

# ClashLink 数据目录初始化脚本
# 确保数据目录有正确的权限

set -e

# 数据目录
DATA_DIR="./data"
DIRS=("database" "subscriptions" "logs")

echo "🔧 初始化 ClashLink 数据目录..."

# 创建主数据目录
mkdir -p "$DATA_DIR"
chmod 755 "$DATA_DIR"

# 创建子目录并设置权限
for dir in "${DIRS[@]}"; do
    FULL_PATH="$DATA_DIR/$dir"
    echo "创建目录: $FULL_PATH"
    mkdir -p "$FULL_PATH"
    
    # 设置权限 - 确保容器中的用户(1001)可以写入
    chmod 755 "$FULL_PATH"
    
    # 如果运行在支持的系统上，设置所有者
    if command -v chown >/dev/null 2>&1; then
        # 设置为当前用户，但允许其他用户访问
        chown -R $USER:$USER "$FULL_PATH" 2>/dev/null || true
    fi
done

echo "✅ 数据目录初始化完成"
echo
echo "📁 创建的目录:"
ls -la "$DATA_DIR"
echo
echo "🔒 权限设置:"
ls -ld "$DATA_DIR"/*
echo
echo "💡 提示:"
echo "  - 如果仍有权限问题，请运行: sudo chown -R 1001:1001 $DATA_DIR"
echo "  - 或设置更宽松的权限: chmod -R 777 $DATA_DIR"
