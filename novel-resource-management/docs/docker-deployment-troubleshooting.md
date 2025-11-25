# Docker部署故障排除指南

## 概述

本文档记录了novel-resource-management项目Docker化部署过程中遇到的问题及解决方案，涵盖MongoDB连接、Fabric网络集成、容器网络配置等方面的故障排除经验。

## 项目背景

- **项目类型**: Go语言Hyperledger Fabric区块链应用
- **技术栈**: Gin框架、MongoDB集群、Fabric测试网络
- **部署环境**: macOS宿主机 + Docker容器
- **目标**: 将API服务容器化，同时连接本地MongoDB和Fabric网络

## 问题分析与解决方案

### 问题1: MongoDB连接失败

#### 症状
```
failed to connect to MongoDB: context deadline exceeded
```

#### 根本原因
1. **副本集配置问题**: MongoDB副本集成员地址配置为`127.0.0.1:27017`，Docker容器无法访问
2. **网络绑定问题**: MongoDB默认只绑定localhost，容器无法连接

#### 解决方案

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
mongo mongodb://admin:715705%40Qc123@127.0.0.1:27017/novel?authSource=admin

// 重新配置副本集成员地址
rs.reconfig({
  "_id": "rs0",
  "members": [
    { "_id": 0, "host": "172.16.181.101:27017" }  // 使用宿主机IP
  ]
})

// 验证配置
rs.status()
```

**步骤3: 更新环境变量**
```yaml
# docker-compose.yml
environment:
  - MONGODB_URI=mongodb://admin:715705%40Qc123@172.16.181.101:27017/novel?replicaSet=rs0&authSource=admin
```

### 问题2: Docker网络隔离导致Fabric连接失败

#### 症状
```
dns resolver: missing address
failed to start event listening: failed to exit idle mode: dns resolver: missing address
```

#### 根本原因
novel-api容器与Fabric peer容器位于不同的Docker网络中，无法直接通信。

#### 解决方案

**步骤1: 修改网络连接代码**
```go
// network/connection.go
func NewGrpcConnection() (*grpc.ClientConn, error) {
    // ... 其他代码 ...

    peerAddress := fmt.Sprintf("%s:%s", peerHost, peerPort)
    // 移除 dns:// 前缀，直接使用容器名
    return grpc.NewClient(peerAddress, grpc.WithTransportCredentials(transportCredentials))
}
```

**步骤2: 连接容器到Fabric网络**
```bash
# 将novel-api容器连接到fabric_test网络
docker network connect fabric_test novel-api

# 重启容器应用网络更改
docker-compose restart novel-api
```

**步骤3: 验证网络连接**
```bash
# 检查网络连接状态
docker network inspect fabric_test | grep novel-api
docker network inspect fabric_test | grep peer0.org1.example.com
```

### 问题3: 架构不匹配导致容器启动失败

#### 症状
```
exec ./novel-api: exec format error
```

#### 根本原因
在macOS上构建的二进制文件是Mach-O格式，而Docker容器需要Linux ELF格式。

#### 解决方案

**方案1: Docker多阶段构建（推荐）**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o novel-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/novel-api .
EXPOSE 8080
CMD ["./novel-api"]
```

**方案2: 本地交叉编译**
```bash
# Intel/AMD macOS
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o novel-api .

# Apple Silicon macOS
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o novel-api .
```

## 最终配置

### docker-compose.yml
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
      - MONGODB_URI=mongodb://admin:715705%40Qc123@172.16.181.101:27017/novel?replicaSet=rs0&authSource=admin
      - MONGODB_DATABASE=novel
      - MONGODB_TIMEOUT=30s
      - MONGODB_MAX_POOL_SIZE=10
      - MONGODB_MIN_POOL_SIZE=2
      - FABRIC_CERT_PATH=/app/test-network/organizations/peerOrganizations/org1.example.com
      - FABRIC_PEER_HOST=peer0.org1.example.com
      - FABRIC_PEER_PORT=7051
    ports:
      - "8080:8080"
    volumes:
      # 挂载Fabric证书文件
      - ../test-network:/app/test-network:ro
    extra_hosts:
      # 允许容器访问宿主机服务
      - "host.docker.internal:host-gateway"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### 网络架构
```
novel-resource-management_default (网络)
└── novel-api容器 (192.168.224.2)

fabric_test (网络)
├── peer0.org1.example.com容器 (172.22.0.3)
├── peer0.org2.example.com容器 (172.22.0.4)
└── novel-api容器 (同时连接)
```

## 验证清单

### MongoDB连接验证
- [x] MongoDB服务正常运行
- [x] 副本集配置正确
- [x] 容器能访问宿主机MongoDB
- [x] 数据库索引创建成功
- [x] 读写操作正常

### Fabric网络验证
- [x] Fabric peer容器运行正常
- [x] 容器间网络连通性正常
- [x] TLS证书验证通过
- [x] 链码初始化成功
- [x] 事件监听器启动成功

### API服务验证
- [x] 健康检查端点正常
- [x] 小说数据API正常
- [x] 用户积分API正常
- [x] 数据库同步正常
- [x] 区块链集成正常

## 常见故障排除命令

### 容器管理
```bash
# 重新构建镜像
docker-compose build --no-cache

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs novel-api

# 进入容器调试
docker-compose exec novel-api sh

# 检查网络连接
docker network ls
docker network inspect <network_name>
```

### MongoDB调试
```bash
# 连接MongoDB
mongo mongodb://admin:password@host:port/database?authSource=admin

# 检查副本集状态
rs.status()
rs.conf()

# 测试网络连通性
telnet 172.16.181.101 27017
```

### Fabric网络调试
```bash
# 检查Fabric容器状态
docker ps | grep example.com

# 测试gRPC连接
docker-compose exec novel-api telnet peer0.org1.example.com 7051

# 查看网络配置
docker network inspect fabric_test
```

## 性能优化建议

### 1. 网络优化
- 使用Docker自定义网络提高性能
- 配置合理的MTU值
- 优化DNS解析

### 2. MongoDB优化
- 配置连接池参数
- 启用压缩传输
- 优化索引策略

### 3. 容器资源限制
```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 512M
    reservations:
      cpus: '0.5'
      memory: 256M
```

## 安全考虑

### 1. 网络安全
- 使用TLS加密通信
- 限制容器网络访问权限
- 配置防火墙规则

### 2. 数据安全
- 加密敏感配置信息
- 定期备份数据库
- 实施访问控制

### 3. 容器安全
- 使用非root用户运行
- 定期更新基础镜像
- 扫描安全漏洞

## 总结

通过本次Docker化部署，我们成功解决了以下关键技术挑战：

1. **跨网络数据库连接**: 通过修改MongoDB配置和副本集设置，实现了容器与宿主机MongoDB的稳定连接
2. **容器间网络通信**: 通过Docker网络配置和代码修改，解决了Fabric网络的连接问题
3. **多架构兼容性**: 通过多阶段构建和交叉编译，确保了二进制文件的架构兼容性
4. **服务集成**: 完整实现了MongoDB、Fabric网络和API服务的无缝集成

这些解决方案不仅解决了当前项目的部署问题，也为类似的企业级区块链应用Docker化提供了有价值的参考经验。

## 相关文件

- `Dockerfile` - 多阶段构建配置
- `docker-compose.yml` - 服务编排配置
- `network/connection.go` - Fabric网络连接逻辑
- `docs/deployment-guide.md` - 部署指南
- `.env` - 环境变量配置