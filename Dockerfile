# 基础镜像
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod tidy

# 复制代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o data-engine ./cmd/data-engine

# 最终镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/data-engine /app/

# 复制.env.example文件
COPY .env.example /app/.env.example

# 安装必要的工具
RUN apk add --no-cache ca-certificates

# 设置环境变量
ENV GIN_MODE=release

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./data-engine"]