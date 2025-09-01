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

### 运行步骤

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
