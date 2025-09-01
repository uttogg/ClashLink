# 🐧 ClashLink Linux 部署指南

本文档专门针对 Linux 服务器环境的部署说明。

## 📋 系统要求

- **操作系统**: Ubuntu 18.04+ / Debian 10+ / CentOS 8+ / RHEL 8+ / 其他Linux发行版
- **架构**: x86_64 (amd64) 或 ARM64 (aarch64)
- **内存**: 至少 512MB RAM
- **存储**: 至少 1GB 可用空间
- **网络**: 需要访问 GitHub API (用于更新检测)

## 🚀 快速部署

### 🐳 Docker 部署 (推荐)

Docker 是最推荐的部署方式，提供环境隔离、数据持久化和简化的管理。

#### 1. Docker 环境准备

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sudo sh

# 将用户添加到 docker 组
sudo usermod -aG docker $USER

# 重新登录或执行以下命令
newgrp docker

# 验证 Docker 安装
docker --version
docker info
```

#### 2. 使用 Docker Compose 部署

```bash
# 下载项目
git clone https://github.com/your-username/clashlink.git
cd clashlink

# 设置环境变量（可选）
export JWT_SECRET="your-very-secure-jwt-secret-here"
export VERSION="v1.0.0"

# 构建并启动
docker-compose up -d

# 查看状态
docker-compose ps
docker-compose logs -f
```

#### 3. 手动 Docker 部署

```bash
# 构建镜像
chmod +x docker-build.sh
./docker-build.sh v1.0.0

# 创建数据目录
mkdir -p data/{database,subscriptions,logs}
chmod 755 data/{database,subscriptions,logs}

# 启动容器
docker run -d \
  --name clashlink-app \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/data/database:/app/backend \
  -v $(pwd)/data/subscriptions:/app/subscriptions \
  -v $(pwd)/data/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e JWT_SECRET="your-secure-jwt-secret" \
  clashlink:v1.0.0

# 验证部署
docker ps | grep clashlink
curl -I http://localhost:8080
```

#### 4. Docker 管理

```bash
# 使用管理脚本
chmod +x docker-run.sh

# 查看状态
./docker-run.sh status

# 查看日志
./docker-run.sh logs -f

# 重启服务
./docker-run.sh restart

# 备份数据
./docker-run.sh backup

# 更新应用
./docker-run.sh update
```

### 2. 传统安装脚本

```bash
# 下载并运行安装脚本
curl -fsSL https://raw.githubusercontent.com/your-username/clashlink/main/install.sh | bash

# 或者手动下载后执行
wget https://raw.githubusercontent.com/your-username/clashlink/main/install.sh
chmod +x install.sh
sudo ./install.sh
```

### 2. 手动部署

#### 安装依赖
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y curl wget git unzip

# CentOS/RHEL
sudo yum install -y curl wget git unzip
# 或者 (CentOS 8+)
sudo dnf install -y curl wget git unzip
```

#### 安装 Go
```bash
# 下载 Go
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 安装 Go
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 部署应用
```bash
# 创建应用目录
sudo mkdir -p /opt/clashlink
sudo chown $USER:$USER /opt/clashlink

# 下载最新版本
cd /opt/clashlink
wget https://github.com/your-username/clashlink/releases/latest/download/clashlink-linux.tar.gz
tar -xzf clashlink-linux.tar.gz --strip-components=1
rm clashlink-linux.tar.gz

# 设置权限
chmod +x backend/clashlink
chmod +x upgrade.sh
```

#### 创建系统服务
```bash
# 创建用户
sudo useradd --system --no-create-home --shell /bin/false clashlink

# 设置权限
sudo chown -R clashlink:clashlink /opt/clashlink

# 创建服务文件
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

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable clashlink
sudo systemctl start clashlink
```

## 🔧 配置管理

### 环境变量配置
```bash
# 编辑服务文件添加环境变量
sudo systemctl edit clashlink

# 添加以下内容：
[Service]
Environment="PORT=8080"
Environment="LOG_LEVEL=info"
Environment="JWT_SECRET=your-jwt-secret-here"
```

### 防火墙配置
```bash
# UFW (Ubuntu)
sudo ufw allow 8080/tcp
sudo ufw reload

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### Nginx 反向代理
```bash
# 安装 Nginx
sudo apt install -y nginx  # Ubuntu/Debian
# sudo yum install -y nginx  # CentOS/RHEL

# 创建配置文件
sudo tee /etc/nginx/sites-available/clashlink > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;
    
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
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/clashlink /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## 🔄 更新升级

### 自动更新
```bash
# 检查更新
sudo /opt/clashlink/upgrade.sh --check-only

# 执行更新
sudo /opt/clashlink/upgrade.sh

# 强制更新
sudo /opt/clashlink/upgrade.sh --force
```

### 定时检查更新
```bash
# 添加到 crontab
sudo crontab -e

# 添加以下行（每天凌晨2点检查更新）
0 2 * * * /opt/clashlink/upgrade.sh --check-only
```

## 📊 监控运维

### 服务管理
```bash
# 查看状态
sudo systemctl status clashlink

# 查看日志
sudo journalctl -u clashlink -f

# 重启服务
sudo systemctl restart clashlink

# 停止服务
sudo systemctl stop clashlink
```

### 性能监控
```bash
# 查看进程资源使用
ps aux | grep clashlink

# 查看端口占用
sudo netstat -tlnp | grep :8080

# 查看磁盘使用
du -sh /opt/clashlink/subscriptions/

# 清理旧订阅文件
find /opt/clashlink/subscriptions/ -name "*.yaml" -mtime +7 -delete
```

### 日志管理
```bash
# 查看应用日志
sudo journalctl -u clashlink --since "1 hour ago"

# 设置日志轮转
sudo tee /etc/logrotate.d/clashlink > /dev/null <<EOF
/var/log/clashlink/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    sharedscripts
    postrotate
        systemctl reload clashlink
    endscript
}
EOF
```

## 🔒 安全加固

### SSL/TLS 配置
```bash
# 安装 Certbot
sudo apt install -y certbot python3-certbot-nginx

# 获取 SSL 证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 访问控制
```bash
# 限制访问 IP (在 Nginx 配置中)
location / {
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://127.0.0.1:8080;
    # ... 其他配置
}
```

## 🚨 故障排除

### 常见问题

1. **服务无法启动**
   ```bash
   # 检查日志
   sudo journalctl -u clashlink --no-pager
   
   # 检查端口占用
   sudo lsof -i :8080
   
   # 手动测试
   cd /opt/clashlink/backend
   sudo -u clashlink ./clashlink
   ```

2. **权限问题**
   ```bash
   sudo chown -R clashlink:clashlink /opt/clashlink
   sudo chmod +x /opt/clashlink/backend/clashlink
   ```

3. **数据库问题**
   ```bash
   ls -la /opt/clashlink/backend/data.db
   sudo -u clashlink touch /opt/clashlink/backend/data.db
   ```

4. **网络连接问题**
   ```bash
   # 测试外网连接
   curl -I https://api.github.com
   
   # 测试本地服务
   curl -I http://localhost:8080
   ```

### 备份恢复
```bash
# 创建备份
sudo tar -czf /backup/clashlink-$(date +%Y%m%d).tar.gz /opt/clashlink

# 恢复备份
sudo systemctl stop clashlink
sudo tar -xzf /backup/clashlink-20240101.tar.gz -C /
sudo systemctl start clashlink
```

## 📞 技术支持

如果遇到部署问题，请：

1. 查看 [README.md](./README.md) 中的详细文档
2. 检查 [Issues](https://github.com/your-username/clashlink/issues) 中的已知问题
3. 提交新的 Issue 并提供详细的错误日志

---

**注意**: 请将配置中的 `your-domain.com` 和 `your-username` 替换为实际的域名和GitHub用户名。
