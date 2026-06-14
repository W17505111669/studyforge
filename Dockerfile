# 多阶段构建：编译阶段用完整 Go 环境，运行阶段用最小镜像
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译（CGO_ENABLED=0 生成静态二进制）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o studyforge .

# ========== 运行阶段 ==========
FROM alpine:3.19

# 安装 CA 证书（HTTPS 请求需要）和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 从编译阶段复制二进制
COPY --from=builder /app/studyforge /usr/local/bin/studyforge

# 复制前端静态文件（Vue 构建产物）
COPY --from=builder /app/web /usr/local/share/web

# 创建数据目录
RUN mkdir -p /data

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

# 启动
CMD ["studyforge"]
