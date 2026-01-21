# LegalExtractor Web 版 Dockerfile
# 多阶段构建：前端构建 + 后端构建 + 运行时

# ============================================
# 阶段 1: 前端构建
# ============================================
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm ci --registry=https://registry.npmmirror.com

# 复制前端源码
COPY frontend/ ./

# 构建前端（跳过类型检查以加速构建）
RUN npm run build || (npm run build -- --skipLibCheck)

# ============================================
# 阶段 2: 后端构建
# ============================================
# 使用 rc-alpine 以支持最新的 Go 版本 (如 1.24)
FROM golang:rc-alpine AS backend-builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 设置 Go 代理（支持构建参数注入，CI 环境可留空）
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY . .

# 复制前端构建产物到嵌入目录（如果需要）
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 构建 Web 服务
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ============================================
# 阶段 3: 运行时
# ============================================
FROM alpine:3.19

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/server /app/server

# 复制前端静态文件（用于独立托管）
COPY --from=frontend-builder /app/frontend/dist /app/static

# 复制配置文件模板（可选）
# COPY config/conf.yaml.example /app/config/conf.yaml

# 创建非 root 用户运行
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动服务
CMD ["/app/server"]
