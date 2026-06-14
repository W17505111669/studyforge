# ========== 阶段 1：构建 Vue 前端 ==========
FROM node:20-alpine AS frontend

WORKDIR /app/web

# 先复制依赖文件，利用 Docker 缓存层
COPY web/package.json web/package-lock.json ./
RUN npm ci --silent

# 复制前端源码并构建
COPY web/ ./
RUN npm run build

# ========== 阶段 2：编译 Go 后端 ==========
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制 Go 源代码
COPY . .

# 从前端构建阶段复制产物（使 Go 阶段能打包 web/dist）
COPY --from=frontend /app/web/dist ./web/dist

# 编译（CGO_ENABLED=0 生成静态二进制）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o studyforge .

# ========== 阶段 3：运行时 ==========
FROM alpine:3.19

# 安装 CA 证书（HTTPS 请求需要）和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从 Go 编译阶段复制二进制
COPY --from=builder /app/studyforge ./studyforge

# 复制前端静态文件（Vite 构建产物 + PWA 资源）
COPY --from=frontend /app/web/dist ./web/dist
COPY --from=builder /app/web/public ./web/public

# 创建数据目录（SQLite + 上传文件）
RUN mkdir -p /app/data

# 暴露端口（Railway 通过 PORT 环境变量指定实际端口）
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

# 启动
CMD ["./studyforge"]
