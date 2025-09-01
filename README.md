# ClashLink - VLESS/VMess 到 Clash 订阅转换工具

一个带用户认证的 VLESS/VMess 到 Clash 订阅的在线转换与测试工具。

## 功能特性

- 🔐 用户注册和登录系统（基于JWT认证）
- 🔄 支持 VLESS/VMess 节点链接解析
- ✅ 节点连通性检测
- 📝 自动生成 Clash YAML 配置文件
- 🌐 提供订阅链接服务
- 📱 响应式Web界面

## 技术栈

### 后端
- Go 1.21+
- SQLite 数据库
- JWT 认证
- bcrypt 密码加密
~
### 前端
- 原生 HTML/CSS/JavaScript
- 现代响应式设计

## 项目结构

```
/clash-converter-app
├── frontend/
│   ├── index.html      # 主应用页面
│   ├── login.html      # 登录/注册页面
│   ├── style.css       # 全局样式
│   ├── script.js       # 主应用逻辑
│   └── auth.js         # 认证逻辑
├── backend/
│   ├── main.go         # 服务器主程序
│   ├── auth.go         # 认证处理
│   ├── middleware.go   # JWT中间件
│   ├── database.go     # 数据库操作
│   ├── user.go         # 用户模型
│   ├── parser.go       # 节点解析
│   ├── checker.go      # 连通性检测
│   ├── generator.go    # 配置生成
│   └── go.mod          # Go模块文件
├── subscriptions/      # 订阅文件存储目录
└── README.md           # 项目说明
```

## 安装和运行

### 前提条件

1. **安装 Go 1.21 或更高版本**
   - 从 [Go官网](https://golang.org/dl/) 下载并安装
   - 确保 `go` 命令在系统 PATH 中

2. **安装 Git**（可选，用于克隆项目）

### 本地开发运行

1. **进入后端目录**
   ```bash
   cd backend
   ```

2. **下载依赖**
   ```bash
   go mod tidy
   ```

3. **运行服务器**
   ```bash
   go run .
   ```

4. **访问应用**
   - 打开浏览器访问：http://localhost:8080
   - 首次访问会看到登录/注册页面

## 🐧 Linux 服务器部署

### 🐳 Docker 部署 (推荐)

Docker 部署是最简单和可靠的方式，提供了完整的环境隔离和数据持久化。

#### 快速开始

1. **安装 Docker**
   ```bash
   # Ubuntu/Debian
   curl -fsSL https://get.docker.com | sudo sh
   sudo usermod -aG docker $USER
   
   # 重新登录或执行
   newgrp docker
   ```

2. **构建镜像**
   ```bash
   # 构建 ClashLink Docker 镜像
   chmod +x docker-build.sh
   ./docker-build.sh v1.0.0
   ```

3. **启动服务**
   ```bash
   # 使用 Docker Compose (推荐)
   docker-compose up -d
   
   # 或使用运行脚本
   chmod +x docker-run.sh
   ./docker-run.sh start
   ```

4. **访问应用**
   - 浏览器访问：http://localhost:8080
   - 首次访问进行系统初始化

#### Docker Compose 部署

```yaml
# 创建 docker-compose.yml 或使用项目自带的配置
version: '3.8'
services:
  clashlink:
    build: .
    container_name: clashlink-app
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data/database:/app/backend
      - ./data/subscriptions:/app/subscriptions
      - ./data/logs:/app/logs
    environment:
      - TZ=Asia/Shanghai
      - JWT_SECRET=your-secure-jwt-secret
```

```bash
# 启动服务
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 手动 Docker 运行

```bash
# 创建数据目录
mkdir -p data/{database,subscriptions,logs}

# 运行容器
docker run -d \
  --name clashlink-app \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/data/database:/app/backend \
  -v $(pwd)/data/subscriptions:/app/subscriptions \
  -v $(pwd)/data/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e JWT_SECRET=your-secure-jwt-secret \
  clashlink:latest
```

#### Docker 管理命令

```bash
# 查看容器状态
./docker-run.sh status

# 查看实时日志
./docker-run.sh logs -f

# 进入容器 shell
./docker-run.sh shell

# 备份数据
./docker-run.sh backup

# 更新容器
./docker-run.sh update

# 重启容器
./docker-run.sh restart
```

### 传统 Linux 部署

如果您不想使用 Docker，也可以直接在 Linux 系统上部署：

#### 📋 系统要求
- Linux 系统 (Ubuntu 18.04+, Debian 10+, CentOS 8+, RHEL 8+ 等)
- 支持架构：x86_64 (Intel/AMD) 或 ARM64 (ARM处理器)
- 至少 512MB RAM
- 至少 1GB 可用磁盘空间

#### 🔧 步骤 1: 更新系统
```bash
# 更新软件包列表
sudo apt update && sudo apt upgrade -y

# 安装基本工具
# Ubuntu/Debian:
sudo apt install -y curl wget git unzip tar

# CentOS/RHEL 8+:
# sudo dnf install -y curl wget git unzip tar

# 旧版本 CentOS/RHEL:
# sudo yum install -y curl wget git unzip tar
```

#### 📦 步骤 2: 安装 Go 语言环境
```bash
# 下载 Go 1.21.5 (请检查最新版本)
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 删除旧版本并安装新版本
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc

# 重新加载环境变量
source ~/.bashrc

# 验证安装
go version
```

#### 🚀 步骤 3: 部署应用
```bash
# 创建应用目录
sudo mkdir -p /opt/clashlink
sudo chown $USER:$USER /opt/clashlink
cd /opt/clashlink

# 克隆或上传项目文件
# 方法1: 如果使用Git
git clone <your-repo-url> .

# 方法2: 手动上传文件
# 将项目文件上传到 /opt/clashlink 目录

# 设置正确的权限
sudo chown -R $USER:$USER /opt/clashlink
chmod +x backend/*.go
```

#### 📝 步骤 4: 配置应用
```bash
# 进入后端目录
cd /opt/clashlink/backend

# 下载 Go 依赖
go mod tidy

# 创建必要的目录
mkdir -p /opt/clashlink/subscriptions

# 设置目录权限
chmod 755 /opt/clashlink/subscriptions
```

#### 🔥 步骤 5: 编译应用
```bash
# 编译应用
cd /opt/clashlink/backend
go build -o clashlink .

# 验证编译结果
./clashlink --help 2>/dev/null || echo "编译成功，准备启动"
```

#### 🛡️ 步骤 6: 创建系统服务
```bash
# 创建系统用户
sudo useradd --system --no-create-home --shell /bin/false clashlink

# 设置文件权限
sudo chown -R clashlink:clashlink /opt/clashlink
sudo chmod +x /opt/clashlink/backend/clashlink

# 创建 systemd 服务文件
sudo tee /etc/systemd/system/clashlink.service > /dev/null <<EOF
[Unit]
Description=ClashLink - VLESS/VMess to Clash Converter
After=network.target
Wants=network.target

[Service]
Type=simple
User=clashlink
Group=clashlink
WorkingDirectory=/opt/clashlink/backend
ExecStart=/opt/clashlink/backend/clashlink
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全设置
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/clashlink

# 环境变量
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
```

#### 🚀 步骤 7: 启动服务
```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start clashlink

# 设置开机自启
sudo systemctl enable clashlink

# 检查服务状态
sudo systemctl status clashlink

# 查看服务日志
sudo journalctl -u clashlink -f
```

#### 🌐 步骤 8: 配置防火墙 (可选)
```bash
# 如果使用 UFW 防火墙
sudo ufw allow 8080/tcp
sudo ufw reload

# 如果使用 iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

#### 🔒 步骤 9: 配置 Nginx 反向代理 (推荐)
```bash
# 安装 Nginx
sudo apt install -y nginx

# 创建站点配置
sudo tee /etc/nginx/sites-available/clashlink > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;  # 替换为你的域名
    
    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    
    # 限制请求大小
    client_max_body_size 1M;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 静态文件缓存
    location /static/ {
        proxy_pass http://127.0.0.1:8080;
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
    
    # 订阅文件
    location /subscriptions/ {
        proxy_pass http://127.0.0.1:8080;
        expires 1h;
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/clashlink /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
sudo systemctl enable nginx
```

#### 🔐 步骤 10: 配置 SSL (推荐)
```bash
# 安装 Certbot
sudo apt install -y certbot python3-certbot-nginx

# 获取 SSL 证书
sudo certbot --nginx -d your-domain.com

# 设置自动续期
sudo crontab -e
# 添加以下行：
# 0 12 * * * /usr/bin/certbot renew --quiet
```

#### 📊 步骤 11: 监控和维护
```bash
# 查看服务状态
sudo systemctl status clashlink

# 查看实时日志
sudo journalctl -u clashlink -f

# 重启服务
sudo systemctl restart clashlink

# 查看端口占用
sudo netstat -tlnp | grep :8080

# 检查磁盘使用
du -sh /opt/clashlink/subscriptions/

# 清理旧的订阅文件（建议定期执行）
find /opt/clashlink/subscriptions/ -name "*.yaml" -mtime +7 -delete
```

#### 🔧 故障排除

**服务无法启动**
```bash
# 检查日志
sudo journalctl -u clashlink --no-pager

# 检查端口占用
sudo lsof -i :8080

# 手动测试
cd /opt/clashlink/backend
sudo -u clashlink ./clashlink
```

**权限问题**
```bash
# 重新设置权限
sudo chown -R clashlink:clashlink /opt/clashlink
sudo chmod +x /opt/clashlink/backend/clashlink
```

**数据库问题**
```bash
# 检查数据库文件
ls -la /opt/clashlink/backend/data.db
sudo -u clashlink touch /opt/clashlink/backend/data.db
```

#### 🎯 访问应用
- **本地访问**: http://localhost:8080
- **域名访问**: http://your-domain.com (如果配置了 Nginx)
- **HTTPS 访问**: https://your-domain.com (如果配置了 SSL)

#### 📈 性能优化建议
- 定期清理过期的订阅文件
- 使用 Nginx 缓存静态资源
- 监控服务器资源使用情况
- 设置日志轮转避免日志文件过大

## 🔄 升级和更新

### 📋 升级前准备

#### 🔍 检查当前版本
```bash
# 查看当前运行的版本信息
cd /opt/clashlink/backend
./clashlink --version 2>/dev/null || echo "当前版本未知"

# 查看服务状态
sudo systemctl status clashlink

# 备份当前配置
sudo cp -r /opt/clashlink /opt/clashlink.backup.$(date +%Y%m%d_%H%M%S)
```

#### 💾 数据备份
```bash
# 备份数据库
sudo cp /opt/clashlink/backend/data.db /opt/clashlink/backend/data.db.backup.$(date +%Y%m%d_%H%M%S)

# 备份重要配置文件
sudo tar -czf /opt/clashlink_backup_$(date +%Y%m%d_%H%M%S).tar.gz \
    /opt/clashlink/backend/data.db \
    /etc/systemd/system/clashlink.service \
    /etc/nginx/sites-available/clashlink

# 验证备份
ls -la /opt/clashlink_backup_*.tar.gz
```

### 🚀 升级方式

#### 方式一: Git 更新 (推荐)
```bash
# 停止服务
sudo systemctl stop clashlink

# 进入项目目录
cd /opt/clashlink

# 拉取最新代码
git fetch origin
git pull origin main

# 查看更新日志
git log --oneline -10

# 更新依赖
cd backend
go mod tidy

# 重新编译
go build -o clashlink .

# 设置权限
sudo chown clashlink:clashlink clashlink
sudo chmod +x clashlink

# 重启服务
sudo systemctl start clashlink
sudo systemctl status clashlink
```

#### 方式二: 手动文件替换
```bash
# 停止服务
sudo systemctl stop clashlink

# 下载新版本文件到临时目录
cd /tmp
wget https://github.com/your-repo/releases/download/v1.x.x/clashlink-linux.tar.gz
tar -xzf clashlink-linux.tar.gz

# 备份当前版本
sudo mv /opt/clashlink/backend/clashlink /opt/clashlink/backend/clashlink.old

# 替换可执行文件
sudo cp clashlink-linux/backend/clashlink /opt/clashlink/backend/
sudo chown clashlink:clashlink /opt/clashlink/backend/clashlink
sudo chmod +x /opt/clashlink/backend/clashlink

# 更新前端文件（如果有变化）
sudo cp -r clashlink-linux/frontend/* /opt/clashlink/frontend/
sudo chown -R clashlink:clashlink /opt/clashlink/frontend/

# 重启服务
sudo systemctl start clashlink
```

#### 方式三: Docker 升级 (可选)
```bash
# 拉取最新镜像
docker pull your-registry/clashlink:latest

# 停止并删除旧容器
docker stop clashlink
docker rm clashlink

# 启动新容器
docker run -d --name clashlink \
    -p 8080:8080 \
    -v /opt/clashlink/data:/app/data \
    -v /opt/clashlink/subscriptions:/app/subscriptions \
    --restart unless-stopped \
    your-registry/clashlink:latest
```

### 🔧 升级后验证

#### ✅ 基础功能测试
```bash
# 检查服务状态
sudo systemctl status clashlink

# 测试端口连通性
curl -I http://localhost:8080

# 查看服务日志
sudo journalctl -u clashlink -n 20

# 测试数据库连接
sudo -u clashlink /opt/clashlink/backend/clashlink --test-db 2>/dev/null || echo "需要手动测试数据库"
```

#### 🌐 Web 功能测试
```bash
# 测试登录页面
curl -s http://localhost:8080/ | grep -q "ClashLink" && echo "✅ 首页正常" || echo "❌ 首页异常"

# 测试 API 端点
curl -s http://localhost:8080/api/init | grep -q "message" && echo "✅ API 正常" || echo "❌ API 异常"

# 测试静态文件
curl -I http://localhost:8080/static/style.css | grep -q "200 OK" && echo "✅ 静态文件正常" || echo "❌ 静态文件异常"
```

### 🚨 升级失败回滚

#### 🔄 快速回滚
```bash
# 停止服务
sudo systemctl stop clashlink

# 恢复可执行文件
sudo mv /opt/clashlink/backend/clashlink.old /opt/clashlink/backend/clashlink

# 恢复数据库（如果需要）
sudo cp /opt/clashlink/backend/data.db.backup.* /opt/clashlink/backend/data.db

# 重启服务
sudo systemctl start clashlink

# 验证回滚
sudo systemctl status clashlink
```

#### 🗂️ 完整回滚
```bash
# 停止服务
sudo systemctl stop clashlink

# 删除当前版本
sudo rm -rf /opt/clashlink

# 恢复备份
sudo tar -xzf /opt/clashlink_backup_*.tar.gz -C /

# 或者从备份目录恢复
sudo mv /opt/clashlink.backup.* /opt/clashlink

# 重启服务
sudo systemctl start clashlink
```

### 📊 升级监控脚本

#### 创建升级脚本
```bash
# 创建升级脚本
sudo tee /opt/clashlink/upgrade.sh > /dev/null <<'EOF'
#!/bin/bash

# ClashLink 升级脚本
set -e

BACKUP_DIR="/opt/clashlink_backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="/var/log/clashlink_upgrade.log"

# 创建日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 创建备份目录
mkdir -p "$BACKUP_DIR"

log "开始升级 ClashLink..."

# 备份当前版本
log "备份当前版本..."
sudo systemctl stop clashlink
cp -r /opt/clashlink "$BACKUP_DIR/clashlink_$TIMESTAMP"
cp /opt/clashlink/backend/data.db "$BACKUP_DIR/data_$TIMESTAMP.db"

# 拉取更新
log "拉取最新代码..."
cd /opt/clashlink
git pull origin main

# 更新依赖和编译
log "编译新版本..."
cd backend
go mod tidy
go build -o clashlink .

# 设置权限
sudo chown clashlink:clashlink clashlink
sudo chmod +x clashlink

# 启动服务
log "启动服务..."
sudo systemctl start clashlink

# 验证升级
sleep 5
if sudo systemctl is-active --quiet clashlink; then
    log "✅ 升级成功！服务正常运行"
    
    # 清理旧备份（保留最近3个）
    cd "$BACKUP_DIR"
    ls -t clashlink_* | tail -n +4 | xargs -r rm -rf
    ls -t data_*.db | tail -n +4 | xargs -r rm -f
    
else
    log "❌ 升级失败，开始回滚..."
    sudo systemctl stop clashlink
    cp -r "$BACKUP_DIR/clashlink_$TIMESTAMP"/* /opt/clashlink/
    sudo systemctl start clashlink
    log "回滚完成"
    exit 1
fi

log "升级完成！"
EOF

# 设置执行权限
sudo chmod +x /opt/clashlink/upgrade.sh
```

#### 自动化升级检查
```bash
# 创建升级检查脚本
sudo tee /opt/clashlink/check_updates.sh > /dev/null <<'EOF'
#!/bin/bash

# 检查是否有新版本
cd /opt/clashlink

# 获取远程最新提交
git fetch origin

# 比较本地和远程版本
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)

if [ "$LOCAL" != "$REMOTE" ]; then
    echo "发现新版本！"
    echo "本地版本: $LOCAL"
    echo "远程版本: $REMOTE"
    
    # 可以选择自动升级或发送通知
    # /opt/clashlink/upgrade.sh
    
    # 或者发送邮件通知
    # echo "ClashLink 有新版本可用" | mail -s "ClashLink Update Available" admin@example.com
else
    echo "当前已是最新版本"
fi
EOF

sudo chmod +x /opt/clashlink/check_updates.sh

# 添加到定时任务（每天检查一次）
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/clashlink/check_updates.sh") | crontab -
```

### 🔧 数据库迁移

#### 数据库结构更新
```bash
# 如果新版本需要数据库结构更新
sudo tee /opt/clashlink/migrate_db.sh > /dev/null <<'EOF'
#!/bin/bash

# 数据库迁移脚本
DB_FILE="/opt/clashlink/backend/data.db"
BACKUP_FILE="/opt/clashlink/backend/data.db.pre_migration.$(date +%Y%m%d_%H%M%S)"

# 备份数据库
cp "$DB_FILE" "$BACKUP_FILE"

# 执行迁移 SQL
sqlite3 "$DB_FILE" <<SQL
-- 示例：添加新字段
-- ALTER TABLE users ADD COLUMN last_login DATETIME;
-- 
-- 示例：创建新表
-- CREATE TABLE IF NOT EXISTS user_sessions (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     user_id INTEGER,
--     session_token TEXT,
--     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );

-- 更新版本信息
INSERT OR REPLACE INTO system_settings (setting_key, setting_value) 
VALUES ('db_version', '1.1.0');
SQL

echo "数据库迁移完成"
EOF

sudo chmod +x /opt/clashlink/migrate_db.sh
```

### 📋 升级最佳实践

#### 🎯 升级策略
1. **测试环境先行**: 在测试环境先验证升级
2. **分阶段升级**: 如果有多个实例，分批升级
3. **备份验证**: 确保备份可用且完整
4. **回滚准备**: 准备快速回滚方案
5. **监控告警**: 升级后密切监控系统状态

#### ⏰ 升级时机
- **低峰时段**: 选择用户访问量较少的时间
- **维护窗口**: 预先通知用户维护时间
- **紧急修复**: 安全漏洞需要立即升级

#### 📝 升级记录
```bash
# 创建升级日志模板
sudo tee /opt/clashlink/UPGRADE_LOG.md > /dev/null <<'EOF'
# ClashLink 升级日志

## 升级记录

### v1.1.0 -> v1.2.0 (2024-01-15)
- **升级时间**: 2024-01-15 02:00 AM
- **升级人员**: admin
- **升级内容**: 
  - 新增用户管理功能
  - 优化节点检测算法
  - 修复已知安全问题
- **数据库变更**: 添加 user_sessions 表
- **配置变更**: 无
- **回滚**: 无需回滚
- **问题**: 无

### 升级检查清单
- [ ] 备份数据库
- [ ] 备份配置文件
- [ ] 停止服务
- [ ] 更新代码
- [ ] 运行数据库迁移
- [ ] 重新编译
- [ ] 启动服务
- [ ] 验证功能
- [ ] 监控日志
EOF
```

现在您有了完整的升级和更新策略，包括自动化脚本、回滚方案和最佳实践！🚀

## 使用说明

### 用户认证

1. **注册账号**
   - 用户名：至少3个字符
   - 密码：至少6个字符
   - 注册成功后会自动跳转到登录页面

2. **登录系统**
   - 输入注册的用户名和密码
   - 登录成功后会跳转到主应用页面

### 节点转换

1. **输入节点链接**
   - 支持 VMess 和 VLESS 协议
   - 每行一个链接
   - 支持的格式：
     - `vmess://base64编码的JSON配置`
     - `vless://uuid@server:port?参数`

2. **配置选项**
   - **检测节点连通性**：测试节点是否可用
   - **仅包含在线节点**：只在配置中包含测试通过的节点
   - **配置文件名称**：自定义生成的配置文件名

3. **生成订阅**
   - 点击"生成订阅"按钮
   - 等待处理完成
   - 获取订阅链接和配置文件

### 使用生成的订阅

1. **复制订阅链接**
   - 点击"复制"按钮复制订阅链接
   - 在 Clash 客户端中添加订阅

2. **下载配置文件**
   - 点击"下载配置"按钮
   - 直接导入到 Clash 客户端

## API 接口

### 公开接口

- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录

### 认证接口（需要JWT）

- `POST /api/generate` - 生成订阅配置

### 静态文件

- `/static/` - 前端静态文件
- `/subscriptions/` - 订阅配置文件

## 安全说明

- 密码使用 bcrypt 加密存储
- JWT 令牌有效期为24小时
- 建议在生产环境中修改 JWT 密钥
- 定期清理生成的订阅文件

## 开发说明

### 后端开发

```bash
# 运行开发服务器
cd backend
go run .

# 构建生产版本
go build -o clashlink
```

### 前端开发

前端使用原生技术栈，直接修改 HTML/CSS/JS 文件即可。

### 数据库

项目使用 SQLite 数据库，数据文件为 `backend/data.db`，首次运行时会自动创建。

## 故障排除

1. **Go 命令不存在**
   - 确保已安装 Go 并添加到系统 PATH

2. **端口被占用**
   - 使用 `lsof -i :8080` 查看端口占用
   - 修改 `main.go` 中的端口号（默认8080）

3. **数据库错误**
   - 删除 `data.db` 文件重新初始化数据库
   - 检查文件权限：`chmod 644 data.db`

4. **前端页面无法访问**
   - 检查静态文件路径是否正确
   - 确保前端文件存在于 `frontend/` 目录
   - 检查文件权限：`chmod -R 644 frontend/`

5. **服务启动失败**
   - 查看日志：`journalctl -u clashlink -f`
   - 检查配置文件权限和路径

## 许可证

本项目仅供学习和个人使用。

## 贡献

欢迎提交 Issue 和 Pull Request。

