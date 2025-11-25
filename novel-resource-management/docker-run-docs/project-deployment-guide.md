# Novel Resource Management 项目部署指南

## 概述

本文档详细记录了novel-resource-management项目从手动部署到完全自动化部署的完整过程，包括遇到的问题、解决方案和最终实现的一键部署方案。

## 项目背景

**项目类型**: Hyperledger Fabric区块链应用
**技术栈**: Go语言、Gin框架、MongoDB集群、Fabric测试网络
**部署目标**: 容器化部署，连接本地MongoDB和Fabric网络，实现一键自动化部署

## 初始问题分析

### 原始需求
用户希望为novel-resource-management项目创建Dockerfile，将项目打包并在Docker容器中运行。

### 技术挑战
1. **Docker容器化** - 将Go应用容器化
2. **MongoDB连接** - 容器需要连接宿主机的MongoDB副本集
3. **Fabric网络集成** - 容器需要连接到Fabric测试网络
4. **自动化部署** - 消除手动配置步骤

## 第一阶段：基础Docker化

### 1.1 创建Dockerfile

**多阶段构建配置**
```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o novel-api .

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/novel-api .
COPY --from=builder /app/.env .
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["./novel-api"]
```

**设计要点**
- 多阶段构建减少最终镜像大小
- 非root用户运行提升安全性
- 内置健康检查
- 复制环境变量文件

### 1.2 创建docker-compose.yml

**初始配置**
```yaml
version: '3.8'

services:
  novel-api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: novel-api
    restart: unless-stopped
    environment:
      - SERVER_PORT=8080
      - MONGODB_URI=mongodb://admin:passward@host.docker.internal:27017/novel?replicaSet=rs0&authSource=admin
      - FABRIC_PEER_HOST=peer0.org1.example.com
      - FABRIC_PEER_PORT=7051
    ports:
      - "8080:8080"
    volumes:
      - ../test-network:/app/test-network:ro
```

## 第二阶段：MongoDB连接问题解决

### 2.1 问题描述

**错误信息**
```
failed to connect to MongoDB: context deadline exceeded
```

**根本原因**
1. MongoDB绑定IP限制 - 默认只绑定localhost
2. 副本集配置错误 - 使用127.0.0.1导致容器无法访问

### 2.2 解决方案

**步骤1: 修改MongoDB配置**
```yaml
# /opt/homebrew/etc/mongod.conf
net:
  bindIp: 0.0.0.0  # 允许外部连接
  ipv6: true
```

**步骤2: 重新配置副本集**
```javascript
// 连接到MongoDB
mongo mongodb://admin:passward@127.0.0.1:27017/novel?authSource=admin

// 重新配置副本集成员地址
rs.reconfig({
  "_id": "rs0",
  "members": [
    { "_id": 0, "host": "172.16.181.101:27017" }
  ]
})

// 验证配置
rs.status()
```

**步骤3: 更新环境变量**
```yaml
environment:
  - MONGODB_URI=mongodb://admin:passward@172.16.181.101:27017/novel?replicaSet=rs0&authSource=admin
```

## 第三阶段：Fabric网络连接问题解决

### 3.1 问题描述

**错误信息**
```
dns resolver: missing address
failed to start event listening: failed to exit idle mode: dns resolver: missing address
```

**根本原因**
- novel-api容器与Fabric peer容器位于不同Docker网络
- 无法解析peer0.org1.example.com

### 3.2 解决方案

**方案分析**
1. **问题识别**: 容器网络隔离
2. **网络连接**: 手动连接容器到fabric_test网络
3. **代码修改**: 移除dns://前缀

**解决步骤**
```bash
# 将novel-api容器连接到fabric_test网络
docker network connect fabric_test novel-api

# 重启容器
docker-compose restart novel-api
```

**代码修改**
```go
// network/connection.go
func NewGrpcConnection() (*grpc.ClientConn, error) {
    // ... 代码 ...
    peerAddress := fmt.Sprintf("%s:%s", peerHost, peerPort)
    // 移除 dns:// 前缀，直接使用容器名
    return grpc.NewClient(peerAddress, grpc.WithTransportCredentials(transportCredentials))
}
```

### 3.3 自动化网络连接

**Docker Compose网络配置**
```yaml
networks:
  novel-network:
    driver: bridge
    ipam:
      config:
        - subnet: 192.168.200.0/24
  fabric_test:
    external: true

services:
  novel-api:
    networks:
      - novel-network
      - fabric_test
```

## 第四阶段：自动化部署脚本开发

### 4.1 自动化需求

**目标**
1. 自动获取宿主机真实IP
2. 自动配置MongoDB副本集
3. 自动执行Docker部署
4. 提供完整的错误处理和进度反馈

### 4.2 Go脚本实现

**文件结构**
```
scripts/
├── deploy.go          # 主部署脚本
├── deploy             # 编译后的可执行文件
└── config-host-mongodb.sh  # MongoDB配置脚本（未使用）
```

**核心功能实现**

**IP获取逻辑**
```go
func getHostIP() (string, error) {
    interfaces, err := net.Interfaces()
    if err != nil {
        return "", fmt.Errorf("获取网络接口失败: %v", err)
    }

    for _, inter := range interfaces {
        // 优先选择172.16网段（局域网）
        if strings.HasPrefix(ip.String(), "172.16.") {
            return ip.String(), nil
        }
    }
    // 备用方案
    return "172.16.181.101", nil
}
```

**MongoDB配置逻辑**
```go
func configureMongoDBReplicaSet(hostIP string) error {
    mongoURI := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%s/%s?authSource=admin",
        MongoUser, MongoPass, MongoPort, MongoDatabase)

    // 检查连接
    if err := checkMongoConnection(mongoURI); err != nil {
        return err
    }

    // 配置副本集
    return configureReplicaSet(mongoURI, hostIP)
}
```

**Docker部署逻辑**
```go
func runDockerDeploy() error {
    // 1. 停止现有容器
    fmt.Println("🔄 停止现有容器...")
    exec.Command("docker-compose", "down").Run()

    // 2. 启动新服务
    fmt.Println("🔨 构建并启动服务...")
    cmd := exec.Command("docker-compose", "up", "-d")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Docker Compose启动失败: %v", err)
    }

    // 3. 健康检查
    return performHealthCheck()
}
```

## 第五阶段：测试和验证

### 5.1 功能验证

**MongoDB连接测试**
```bash
mongosh mongodb://admin:passward@127.0.0.1:27017/admin?authSource=admin --eval "db.adminCommand('ping')"
```

**API健康检查**
```bash
curl -s http://localhost:8080/health
```

**网络连接验证**
```bash
docker network inspect fabric_test | grep novel-api
```

### 5.2 完整部署测试

**自动化脚本执行**
```bash
./scripts/deploy
```

**执行流程**
```
🚀 开始自动化部署novel-resource-management...
🔍 找到172.16网段IP: 172.16.181.101
✅ 宿主机IP: 172.16.181.101
🔧 开始配置MongoDB副本集...
✅ MongoDB连接成功
✅ 副本集已配置，检查IP配置...
✅ 副本集配置已正确
🐳 开始Docker部署...
✅ Docker服务正常
🔄 停止现有容器...
🔨 构建并启动服务...
⏳ 等待服务启动...
🏥 执行健康检查...
✅ 服务健康检查通过
🎉 自动化部署完成!
```

## 最终解决方案

### 完整的自动化部署方案

**一键部署命令**
```bash
./scripts/deploy
```

**部署流程**
1. **智能IP获取** - 自动识别宿主机真实局域网IP
2. **MongoDB自动配置** - 检查并配置副本集使用宿主机IP
3. **Docker自动部署** - 停止旧容器，构建并启动新服务
4. **网络自动连接** - 容器自动连接到fabric_test网络
5. **健康检查验证** - 自动等待并验证服务可用性

### 技术亮点

**网络解决方案**
- 自动获取宿主机真实IP (172.16.181.101)
- 避免Docker网络地址冲突
- 自动连接到外部Fabric网络

**容器化最佳实践**
- 多阶段构建优化镜像大小
- 非root用户运行提升安全性
- 内置健康检查机制
- 环境变量配置灵活

**自动化特性**
- 防重复启动机制
- 完整错误处理
- 实时进度反馈
- 跨平台兼容

## 项目文件结构

```
novel-resource-management/
├── Dockerfile                          # 多阶段构建配置
├── docker-compose.yml                 # 服务编排配置
├── scripts/
│   ├── deploy.go                     # 自动化部署脚本
│   └── deploy                        # 编译后的可执行文件
├── docs/
│   ├── project-deployment-guide.md   # 项目部署指南（本文档）
│   └── docker-deployment-troubleshooting.md  # 故障排除指南
├── network/
│   └── connection.go                  # Fabric网络连接逻辑
└── .env                              # 环境变量配置
```

## 使用指南

### 快速开始

**1. 环境准备**
```bash
# 确保MongoDB已启动并配置副本集
# 确保Fabric测试网络已启动
# 确保Docker已安装并运行
```

**2. 一键部署**
```bash
# 进入项目根目录
cd novel-resource-management

# 执行自动化部署
./scripts/deploy
```

**3. 验证部署**
```bash
# 健康检查
curl http://localhost:8080/health

# API测试
curl http://localhost:8080/api/v1/novels
```

### 手动部署

如果需要手动部署，可以参考以下步骤：

**1. 构建Docker镜像**
```bash
docker-compose build
```

**2. 启动服务**
```bash
docker-compose up -d
```

**3. 检查状态**
```bash
docker-compose ps
docker-compose logs novel-api
```

## 故障排除

### 常见问题

**MongoDB连接失败**
```bash
# 检查MongoDB服务状态
brew services list | grep mongodb

# 检查副本集状态
mongosh mongodb://admin:passward@127.0.0.1:27017/admin?authSource=admin --eval "rs.status()"
```

**Fabric网络连接问题**
```bash
# 检查Fabric网络
docker network ls | grep fabric

# 检查容器网络连接
docker network inspect fabric_test
```

**端口冲突**
```bash
# 检查端口占用
lsof -i :8080

# 停止占用端口的进程
docker-compose down
```

### 调试命令

**容器调试**
```bash
# 进入容器
docker-compose exec novel-api sh

# 查看容器日志
docker-compose logs novel-api

# 实时跟踪日志
docker-compose logs -f novel-api
```

**网络调试**
```bash
# 测试容器间网络连接
docker-compose exec novel-api ping peer0.org1.example.com

# 检查DNS解析
docker-compose exec novel-api nslookup peer0.org1.example.com
```

## 性能优化

### 部署优化

**1. 镜像优化**
```dockerfile
# 使用Alpine基础镜像
FROM alpine:latest

# 多阶段构建减少镜像大小
FROM golang:1.23-alpine AS builder

# 清理不必要的包
RUN apk del .build-deps
```

**2. 资源限制**
```yaml
services:
  novel-api:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

**3. 网络优化**
```yaml
networks:
  novel-network:
    driver: bridge
    driver_opts:
      com.docker.network.bridge.name: novel-br
```

### 监控建议

**1. 健康检查**
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 60s
```

**2. 日志管理**
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 总结

本项目成功实现了从手动部署到完全自动化部署的转变，主要解决了以下关键问题：

1. **Docker容器化** - 通过多阶段构建实现了优化的容器镜像
2. **MongoDB集成** - 解决了宿主机MongoDB副本集的容器访问问题
3. **Fabric网络连接** - 实现了容器与外部Docker网络的自动连接
4. **自动化部署** - 开发了完整的Go自动化部署脚本

**最终成果**:
- 一键部署命令 `./scripts/deploy`
- 完整的错误处理和状态反馈
- 跨平台兼容的解决方案
- 生产就绪的容器化应用

这个解决方案不仅解决了当前项目的部署问题，也为类似的企业级区块链应用Docker化部署提供了有价值的参考经验。