FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

RUN go install github.com/air-verse/air@latest

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖项
RUN go mod download

CMD ["air", "-c", ".air.toml"]

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o webauthn-server ./cmd/main.go

# 使用轻量级基础镜像
FROM alpine:latest

# 安装必要的CA证书和curl（用于健康检查）
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/webauthn-server .

# 创建配置目录
RUN mkdir -p /app/config

# 设置环境变量
ENV PORT=8080

# 暴露应用端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 运行应用
CMD ["./webauthn-server"]