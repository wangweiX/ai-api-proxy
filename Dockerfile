# 构建阶段
FROM golang:1.23.2-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy ./cmd/proxy

# 运行阶段
FROM alpine:latest

# 安装证书
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /app

# 创建日志目录并设置权限
RUN mkdir -p /app/logs && chmod 755 /app/logs

# 复制可执行文件
COPY --from=builder /app/proxy .
COPY --from=builder /app/config.yaml .

# 设置环境变量，指定默认配置文件路径
ENV CONFIG_PATH=/app/config.yaml

# 复制入口脚本
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 设置入口点
ENTRYPOINT ["/app/entrypoint.sh"]