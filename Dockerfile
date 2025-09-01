# ClashLink Dockerfile - 多阶段构建
# 此 Dockerfile 使用多阶段构建来创建轻量级的生产镜像

# ==========================================
# 构建阶段 (BUILD STAGE)
# ==========================================
# 使用官方 Go 镜像作为构建环境，包含完整的 Go 工具链
FROM golang:1.21-alpine AS builder

# 设置构建参数，可在构建时通过 --build-arg 传递
ARG VERSION=1.0.0
ARG BUILD_TIME
ARG GIT_COMMIT

# 在 Alpine 中安装必要的构建工具
# git: 用于获取依赖包
# ca-certificates: 用于 HTTPS 连接
# tzdata: 时区数据
RUN apk add --no-cache git ca-certificates tzdata

# 设置构建环境的工作目录
WORKDIR /app

# 首先复制 go.mod 和 go.sum 文件
# 这样做可以利用 Docker 的层缓存机制：
# 如果依赖没有变化，Docker 会复用缓存的依赖下载层
COPY backend/go.mod backend/go.sum ./

# 下载 Go 模块依赖
# go mod download 会下载所有在 go.mod 中声明的依赖包
RUN go mod download

# 复制所有后端源代码文件
# 将 backend/ 目录下的所有 .go 文件复制到容器的 /app 目录
COPY backend/*.go ./

# 复制前端静态文件目录
# Web 应用需要这些静态文件来提供前端界面
COPY frontend/ ./frontend/

# 创建订阅文件目录并设置权限
# subscriptions/ 目录用于存储生成的 Clash 配置文件
# 设置 777 权限确保应用可以写入文件
RUN mkdir -p ./subscriptions && chmod 777 ./subscriptions

# 复制版本配置文件
# version.json 包含应用的版本信息和配置
COPY version.json ./

# 编译 Go 应用程序
# CGO_ENABLED=0: 禁用 CGO，生成静态链接的二进制文件，不依赖系统库
# GOOS=linux: 交叉编译为 Linux 平台可执行文件（如果宿主机不是 Linux，这很重要）
# -ldflags: 链接器标志
#   -s: 移除符号表，减小文件大小
#   -w: 移除调试信息，进一步减小文件大小
#   -X: 在编译时注入变量值，用于版本信息
# -o clashlink: 指定编译后的二进制文件名为 'clashlink'
# .: 编译当前目录（/app）下的所有 Go 源文件
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o clashlink .

# ==========================================
# 运行阶段 (RUNTIME STAGE)
# ==========================================
# 使用一个更小的、只包含运行环境的 Alpine Linux 镜像作为最终镜像的基础
# 这就是多阶段构建的关键：最终镜像不包含 Go 编译器等不必要的工具
FROM alpine:latest

# 安装运行时依赖
# ca-certificates: 用于 HTTPS 请求（检查 GitHub 更新时需要）
# tzdata: 时区数据，确保时间显示正确
# curl: 用于健康检查和调试
RUN apk update && apk add --no-cache ca-certificates tzdata curl && rm -rf /var/cache/apk/*

# 设置时区为中国标准时间
ENV TZ=Asia/Shanghai

# 创建非特权用户来运行应用（安全最佳实践）
# 不使用 root 用户运行应用程序
RUN addgroup -g 1001 -S clashlink && \
    adduser -u 1001 -S clashlink -G clashlink

# 设置容器内的工作目录
WORKDIR /app

# 从构建阶段 (builder) 复制编译好的二进制文件 'clashlink' 到最终镜像
COPY --from=builder /app/clashlink ./clashlink

# 复制前端静态文件目录 'frontend/'
# 这些是 Web 应用需要的 HTML, CSS, JS 文件
COPY --from=builder /app/frontend ./frontend

# 复制版本配置文件
COPY --from=builder /app/version.json ./version.json

# 创建数据持久化目录
# 这些目录将在运行时挂载为数据卷
RUN mkdir -p /app/backend /app/subscriptions /app/logs && \
    chown -R clashlink:clashlink /app

# 设置可执行文件权限
RUN chmod +x ./clashlink

# 切换到非特权用户
USER clashlink

# 暴露应用程序监听的端口
# 这告诉 Docker 容器监听 8080 端口
EXPOSE 8080

# 设置健康检查
# Docker 会定期执行此命令来检查容器是否健康
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/version || exit 1

# 定义容器启动时默认执行的命令
# 运行我们编译好的 Go 应用程序
CMD ["./clashlink"]
