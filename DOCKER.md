# 🐳 ClashLink Docker 部署指南

本文档提供 ClashLink 的 Docker 容器化部署完整指南。

## 📋 目录

- [快速开始](#-快速开始)
- [详细配置](#-详细配置)
- [数据持久化](#-数据持久化)
- [管理运维](#-管理运维)
- [故障排除](#-故障排除)

## 🚀 快速开始

### 一键安装
```bash
# 下载并执行一键安装脚本
curl -fsSL https://raw.githubusercontent.com/your-username/clashlink/main/docker-install.sh | bash
```

### 手动安装
```bash
# 1. 克隆项目
git clone https://github.com/your-username/clashlink.git
cd clashlink

# 2. 构建镜像
chmod +x docker-build.sh
./docker-build.sh

# 3. 启动服务
docker-compose up -d

# 4. 访问应用
open http://localhost:8080
```

## ⚙️ 详细配置

### Dockerfile 说明

我们的 Dockerfile 使用多阶段构建来优化镜像大小：

```dockerfile
# 构建阶段：使用完整的 Go 环境编译应用
FROM golang:1.21-alpine AS builder
# ... 编译过程

# 运行阶段：使用轻量级 Alpine 镜像
FROM alpine:latest
# ... 只包含运行时必需的文件
```

**优势**:
- 🔸 **小体积**: 最终镜像只有 ~20MB
- 🔸 **安全性**: 不包含构建工具和源码
- 🔸 **快速**: 启动速度快，资源占用少

### 环境变量配置

创建 `.env` 文件自定义配置：

```bash
# 复制配置模板
cp env.example .env

# 编辑配置
nano .env
```

**重要配置项**:
```bash
# 安全配置
JWT_SECRET=your-very-secure-jwt-secret-here

# 服务配置
HOST_PORT=8080
LOG_LEVEL=info

# GitHub 配置
GITHUB_REPO=your-username/clashlink
```

### 网络配置

#### 端口映射
```bash
# 默认端口
-p 8080:8080

# 自定义端口
-p 9090:8080

# 仅本地访问
-p 127.0.0.1:8080:8080
```

#### 反向代理
```nginx
# Nginx 配置示例
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 💾 数据持久化

### 数据卷挂载

ClashLink 需要持久化以下数据：

```bash
# 数据库文件
-v /opt/clashlink/data/database:/app/backend

# 订阅配置文件
-v /opt/clashlink/data/subscriptions:/app/subscriptions

# 应用日志
-v /opt/clashlink/data/logs:/app/logs
```

### 备份策略

#### 自动备份
```bash
# 创建备份脚本
cat > /opt/clashlink/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backup/clashlink"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"
docker exec clashlink-app tar -czf - /app/backend /app/subscriptions | \
    cat > "$BACKUP_DIR/clashlink_$DATE.tar.gz"

# 清理旧备份 (保留7天)
find "$BACKUP_DIR" -name "clashlink_*.tar.gz" -mtime +7 -delete
EOF

chmod +x /opt/clashlink/backup.sh

# 添加到 crontab (每天凌晨3点备份)
(crontab -l 2>/dev/null; echo "0 3 * * * /opt/clashlink/backup.sh") | crontab -
```

#### 手动备份
```bash
# 使用管理脚本
./docker-run.sh backup

# 或直接使用 Docker
docker exec clashlink-app tar -czf - /app/backend /app/subscriptions > backup.tar.gz
```

## 🔧 管理运维

### 容器管理

```bash
# 启动服务
./docker-run.sh start

# 停止服务
./docker-run.sh stop

# 重启服务
./docker-run.sh restart

# 查看状态
./docker-run.sh status

# 查看日志
./docker-run.sh logs -f
```

### 应用更新

```bash
# 方法1: 使用管理脚本
./docker-run.sh update

# 方法2: 手动更新
docker-compose pull
docker-compose up -d

# 方法3: 重新构建
./docker-build.sh v1.1.0
docker-compose up -d
```

### 监控配置

#### 健康检查
```bash
# 查看健康状态
docker inspect --format='{{.State.Health.Status}}' clashlink-app

# 健康检查日志
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' clashlink-app
```

#### 资源监控
```bash
# 实时资源使用
docker stats clashlink-app

# 容器详细信息
docker inspect clashlink-app
```

### 日志管理

#### 日志轮转
```bash
# 配置 Docker 日志轮转
cat > /etc/docker/daemon.json << EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

sudo systemctl restart docker
```

#### 日志查看
```bash
# 查看最近日志
docker logs --tail 100 clashlink-app

# 跟踪实时日志
docker logs -f clashlink-app

# 查看特定时间段日志
docker logs --since="2024-01-01T00:00:00" --until="2024-01-02T00:00:00" clashlink-app
```

## 🚨 故障排除

### 常见问题

#### 1. 容器无法启动
```bash
# 查看容器日志
docker logs clashlink-app

# 检查镜像是否存在
docker images | grep clashlink

# 检查端口占用
sudo netstat -tlnp | grep 8080
```

#### 2. 数据丢失
```bash
# 检查数据卷挂载
docker inspect clashlink-app | grep -A 10 "Mounts"

# 检查数据目录权限
ls -la data/

# 恢复备份
./docker-run.sh restore
```

#### 3. 性能问题
```bash
# 查看资源使用
docker stats clashlink-app

# 调整资源限制
docker update --memory=1g --cpus=1.0 clashlink-app
```

#### 4. 网络连接问题
```bash
# 测试容器网络
docker exec clashlink-app curl -I https://api.github.com

# 检查防火墙
sudo ufw status
sudo iptables -L
```

### 调试技巧

#### 进入容器调试
```bash
# 进入运行中的容器
docker exec -it clashlink-app /bin/sh

# 以 root 用户进入
docker exec -it -u root clashlink-app /bin/sh

# 运行临时调试容器
docker run -it --rm clashlink:latest /bin/sh
```

#### 查看构建过程
```bash
# 详细构建日志
docker build --no-cache --progress=plain -t clashlink:debug .

# 查看镜像层
docker history clashlink:latest
```

## 📈 性能优化

### 镜像优化
- ✅ 多阶段构建减少镜像大小
- ✅ 使用 Alpine Linux 基础镜像
- ✅ 移除调试信息和符号表
- ✅ 使用非特权用户运行

### 运行时优化
```bash
# 资源限制
docker run -d \
  --name clashlink-app \
  --memory=512m \
  --cpus=0.5 \
  --restart unless-stopped \
  -p 8080:8080 \
  clashlink:latest
```

### 存储优化
```bash
# 使用命名卷而非绑定挂载（可选）
docker volume create clashlink-data
docker run -d \
  --name clashlink-app \
  -v clashlink-data:/app/backend \
  clashlink:latest
```

## 🔐 安全建议

### 容器安全
- ✅ 使用非特权用户运行应用
- ✅ 最小化镜像包含的组件
- ✅ 定期更新基础镜像
- ✅ 配置资源限制

### 网络安全
```bash
# 创建自定义网络
docker network create clashlink-net

# 在自定义网络中运行
docker run -d \
  --name clashlink-app \
  --network clashlink-net \
  -p 8080:8080 \
  clashlink:latest
```

### 数据安全
- 🔒 定期备份数据
- 🔒 加密敏感配置
- 🔒 限制容器文件系统访问
- 🔒 使用 secrets 管理敏感信息

---

**提示**: 生产环境部署建议配合 Nginx、SSL 证书和防火墙使用。
