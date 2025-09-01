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

### Windows 运行步骤

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

### Debian/Ubuntu 部署教程

#### 📋 系统要求
- Debian 10+ 或 Ubuntu 18.04+
- 至少 512MB RAM
- 至少 1GB 可用磁盘空间

#### 🔧 步骤 1: 更新系统
```bash
# 更新软件包列表
sudo apt update && sudo apt upgrade -y

# 安装基本工具
sudo apt install -y curl wget git unzip
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
go build -o clash-converter.exe
```

### 前端开发

前端使用原生技术栈，直接修改 HTML/CSS/JS 文件即可。

### 数据库

项目使用 SQLite 数据库，数据文件为 `backend/data.db`，首次运行时会自动创建。

## 故障排除

1. **Go 命令不存在**
   - 确保已安装 Go 并添加到系统 PATH

2. **端口被占用**
   - 修改 `main.go` 中的端口号（默认8080）

3. **数据库错误**
   - 删除 `data.db` 文件重新初始化数据库

4. **前端页面无法访问**
   - 检查静态文件路径是否正确
   - 确保前端文件存在于 `frontend/` 目录

## 许可证

本项目仅供学习和个人使用。

## 贡献

欢迎提交 Issue 和 Pull Request。

