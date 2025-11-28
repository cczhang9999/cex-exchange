# 多阶段构建 - 后端Go服务
FROM golang:1.21-alpine AS backend-builder

# 设置工作目录
WORKDIR /app

# 使用阿里云apk源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev

# 复制go mod文件
COPY cex-exchange/go.mod cex-exchange/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY cex-exchange/ .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 使用阿里云apk源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 复制配置文件
COPY cex-exchange/config.yaml .

# 创建数据目录
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/klines || exit 1

# 启动命令
CMD ["./main"] 