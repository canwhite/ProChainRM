# Docker 部署指南

本文档介绍如何使用Docker部署小说资源管理系统。

## 📋 目录

- [系统要求](#系统要求)
- [快速开始](#快速开始)
- [详细配置](#详细配置)
- [生产部署](#生产部署)
- [故障排除](#故障排除)

## 🔧 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 1GB 可用磁盘空间

## 🚀 快速开始

### 1. 前置条件

在部署应用之前，请确保：

1. **Fabric网络已启动**:
```bash
cd ../test-network
./network.sh up
./network.sh createChannel
./network.sh deployCC -ccn novel-basic -ccp ../novel-resource-events -ccl go -cci InitLedger
```

2. **本地MongoDB服务运行中**:
```bash
# 检查MongoDB是否运行
mongosh --eval "db.adminCommand('ping')" --host 127.0.0.1:27017 -u admin -p "715705@Qc123"
```

### 2. 启动应用服务

```bash
cd novel-resource-management
docker-compose up -d
```

### 3. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 检查API健康状态
curl http://localhost:8080/health

# 查看日志
docker-compose logs novel-api
```

## 📁 项目结构

```
novel-resource-management/
├── Dockerfile                 # 应用容器化配置
├── docker-compose.yml        # 应用服务编排（连接本地MongoDB）
├── .dockerignore            # Docker构建忽略文件
├── .env                     # 环境配置文件（包含本地MongoDB连接信息）
├── docs/
│   └── DOCKER_DEPLOYMENT.md # 本文档
└── ...
```

## ⚙️ 详细配置

### 环境变量配置

#### 本地MongoDB配置
```env
# 本地MongoDB连接信息（来自.env文件）
MONGODB_URI=mongodb://admin:715705%40Qc123@host.docker.internal:27017
MONGODB_DATABASE=novel
MONGODB_TIMEOUT=30s
MONGODB_MAX_POOL_SIZE=10
MONGODB_MIN_POOL_SIZE=2
```

#### 服务配置
```env
SERVER_PORT=8080
```

### Docker Compose 服务说明

#### Novel API服务（连接本地MongoDB）
```yaml
novel-api:
  build: .
  environment:
    - SERVER_PORT=8080
    - MONGODB_URI=mongodb://admin:715705%40Qc123@host.docker.internal:27017
    - MONGODB_DATABASE=novel
  ports:
    - "8080:8080"
  volumes:
    - ../test-network:/app/test-network:ro  # Fabric证书挂载
    - ./.env:/app/.env:ro                   # 环境配置文件
  extra_hosts:
    - "host.docker.internal:host-gateway"  # 允许访问宿主机服务
```

### 关键配置说明

#### host.docker.internal
这是一个特殊的DNS名称，允许Docker容器访问宿主机上的服务：
- `127.0.0.1:27017`（宿主机）→ `host.docker.internal:27017`（容器内）
- 这样容器就能连接你本地的MongoDB服务

#### 数据库连接说明
- **用户名**: `admin`
- **密码**: `715705@Qc123`
- **数据库名**: `novel`
- **连接地址**: `host.docker.internal:27017`
- **认证方式**: 用户名密码认证

### 自定义配置

#### 修改端口
编辑 `docker-compose.yml`：
```yaml
services:
  novel-api:
    ports:
      - "9090:8080"  # 将外部端口改为9090
```

#### 修改MongoDB配置
1. 编辑 `.env` 文件
2. 或者直接在 `docker-compose.yml` 中修改环境变量

## 🏭 生产部署

### 1. 安全配置

#### 使用生产环境变量
```bash
# 创建生产环境配置
cp .env .env.production

# 编辑生产环境配置
vim .env.production
```

#### 使用外部证书
```yaml
volumes:
  - /path/to/production/test-network:/app/test-network:ro
```

### 2. 性能优化

#### 资源限制
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

#### MongoDB优化
```yaml
mongodb:
  environment:
    MONGO_INITDB_CACHE_SIZE_GB: 0.25
    MONGO_WIRED_TIGER_CACHE_SIZE_GB: 0.25
```

### 3. 监控和日志

#### 日志配置
```yaml
services:
  novel-api:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### 健康检查增强
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

## 🔧 常用命令

### 服务管理
```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f novel-api
docker-compose logs -f mongodb

# 进入容器
docker-compose exec novel-api sh
docker-compose exec mongodb mongosh
```

### 数据管理
```bash
# 备份数据
docker-compose exec mongodb mongodump --out /backup

# 恢复数据
docker-compose exec mongodb mongorestore /backup

# 连接MongoDB
docker-compose exec mongodb mongosh -u admin -p 715705@Qc123
```

### 构建和镜像管理
```bash
# 重新构建镜像
docker-compose build --no-cache

# 查看镜像
docker images | grep novel

# 清理未使用的镜像
docker image prune
```

## 🐛 故障排除

### 常见问题

#### 1. 证书文件找不到
**错误**: `failed to read TLS certificate: no such file or directory`

**解决方案**:
```bash
# 确认test-network路径
ls -la ../test-network/organizations/

# 检查挂载路径
docker-compose exec novel-api ls -la /app/test-network/
```

#### 2. MongoDB连接失败
**错误**: `MongoDB自动连接失败`

**解决方案**:
```bash
# 检查MongoDB状态
docker-compose ps mongodb

# 查看MongoDB日志
docker-compose logs mongodb

# 手动连接测试
docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')"
```

#### 3. API服务无法启动
**错误**: `Failed to start server`

**解决方案**:
```bash
# 检查端口占用
netstat -tlnp | grep 8080

# 查看详细日志
docker-compose logs novel-api

# 进入容器调试
docker-compose exec novel-api sh
```

#### 4. 健康检查失败
**错误**: `Health check failed`

**解决方案**:
```bash
# 手动检查健康端点
curl http://localhost:8080/health

# 检查服务是否真正启动
docker-compose exec novel-api ps aux
```

### 调试技巧

#### 1. 查看详细日志
```bash
# 启用调试模式
docker-compose run --rm novel-api ./novel-api -debug

# 实时查看日志
docker-compose logs -f --tail=100 novel-api
```

#### 2. 网络调试
```bash
# 检查网络连接
docker-compose exec novel-api ping mongodb

# 检查端口连通性
docker-compose exec novel-api telnet mongodb 27017
```

#### 3. 证书调试
```bash
# 检查证书文件权限
docker-compose exec novel-api ls -la /app/test-network/organizations/

# 验证证书内容
docker-compose exec novel-api openssl x509 -in /app/test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem -text -noout
```

## 📊 监控指标

### 关键指标监控
- API响应时间
- MongoDB连接数
- 内存使用率
- CPU使用率
- 磁盘I/O

### 日志监控
- 应用错误日志
- 数据库连接错误
- Fabric网络连接状态

## 🔄 升级和维护

### 升级应用
```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose build --no-cache

# 重启服务
docker-compose up -d
```

### 数据迁移
```bash
# 备份当前数据
docker-compose exec mongodb mongodump --out /backup/$(date +%Y%m%d)

# 执行迁移脚本
docker-compose exec novel-api ./novel-api migrate
```

## 📞 支持

如果在部署过程中遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查应用日志和Docker日志
3. 确认所有前置条件已满足
4. 提交Issue并附上详细的错误信息和环境描述

---

**最后更新**: 2025-11-24
**版本**: 1.0.0