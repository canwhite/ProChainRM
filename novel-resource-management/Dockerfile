# 构建阶段，整个过程是按顺序执行的
FROM golang:1.23-alpine AS builder

# 设置工作目录,进入容器中的/app，等价于mkdir -p /app，then cd /app
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o novel-api .

# 运行阶段，将build和run区分开
FROM alpine:latest

# 安装ca-certificates用于HTTPS连接、时区数据和wget
RUN apk --no-cache add ca-certificates tzdata wget

# 设置时区为Asia/Shanghai
ENV TZ=Asia/Shanghai

# 创建非root用户，在Docker容器里创建一个普通用户，而不是用root用户运行应用。
# 创建一个名为"appgroup"的用户组
# 创建一个名为"appuser"的用户，并加入appgroup组
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件，  "从厨房（builder）把做好的菜（novel-api）端到餐厅（alpine）"
COPY --from=builder /app/novel-api .

# 复制.env文件（如果存在），同上
COPY --from=builder /app/.env .

# 创建证书目录
#  "把宿主机的 ../test-network 目录挂载到容器的 /app/test-network 目录"，主要是为了避免证书复制
# 搭配的是volume挂载
#  Docker volume挂载，ro是只读挂载
#   volumes:
#     - ../test-network:/app/test-network:ro
RUN mkdir -p /app/test-network/organizations/peerOrganizations/org1.example.com

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./novel-api"]