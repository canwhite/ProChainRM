# Fabric证书路径问题排查与解决方案

本文档记录了在Docker容器中遇到Fabric证书路径问题的排查过程和解决方案。

## 🚨 问题现象

### 初始症状
```
ERROR: Container restarts (1)
日志显示: "Failed to create gRPC connection: failed to read TLS certificate:
          open ../test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem: no such file or directory"
```

### 前端表现
```json
{"error":"Proxy request failed"}
```

## 🔍 问题分析

### 根本原因
Docker容器内使用相对路径 `../test-network/` 无法找到Fabric证书文件，因为：

1. **路径结构不同**: 容器内工作目录是 `/app`，相对路径计算不同
2. **挂载路径不匹配**: 代码期望的路径 ≠ 实际挂载路径

### 文件结构对比

**宿主机路径结构**:
```
ProChainRM/
├── test-network/                    # Fabric网络配置
│   └── organizations/
│       └── peerOrganizations/
│           └── org1.example.com/
│               ├── tlsca/
│               └── users/
└── novel-resource-management/        # 应用代码
    ├── main.go
    ├── network/
    │   └── connection.go           # 使用相对路径
    └── Dockerfile
```

**容器内预期路径**:
```
/app/                                # 工作目录
├── novel-api                        # 编译后的应用
└── test-network/                    # 挂载的证书目录
    └── organizations/
        └── peerOrganizations/
            └── org1.example.com/
                ├── tlsca/
                └── users/
```

**问题**: 代码中的 `../test-network/` 路径在容器内解析为 `/app/../test-network/` = `/test-network/`，而不是 `/app/test-network/`

## 💡 解决方案

### 方案对比

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| 修改挂载路径匹配代码 | 代码无需修改 | 路径结构不够直观 | ⭐⭐ |
| 修改代码使用绝对路径 | 灵活性高 | 需要修改多处代码 | ⭐⭐⭐⭐ |
| 使用环境变量配置 | 最佳实践，配置灵活 | 需要额外配置 | ⭐⭐⭐⭐⭐ |

### 最终采用方案: 环境变量配置

#### 1. 添加环境变量配置

**docker-compose.yml**:
```yaml
services:
  novel-api:
    environment:
      # 添加Fabric证书路径环境变量
      - FABRIC_CERT_PATH=/app/test-network/organizations/peerOrganizations/org1.example.com
    volumes:
      # 确保证书正确挂载
      - ../test-network:/app/test-network:ro
```

#### 2. 修改网络连接代码

**network/connection.go**:
```go
func NewGrpcConnection() (*grpc.ClientConn, error) {
    // 获取Fabric证书路径
    certPath := os.Getenv("FABRIC_CERT_PATH")
    if certPath == "" {
        certPath = "../test-network/organizations/peerOrganizations/org1.example.com" // 默认路径
    }

    // 使用环境变量路径
    tlsCertificatePEM, err := os.ReadFile(fmt.Sprintf("%s/tlsca/tlsca.org1.example.com-cert.pem", certPath))
    // ...
}
```

#### 3. 统一修改所有证书路径

需要修改的文件和函数:
- `NewGrpcConnection()` - TLS证书
- `NewIdentity()` - 用户证书
- `NewSign()` - 私钥文件

## 🔧 实施步骤

### Step 1: 更新环境变量
```yaml
# docker-compose.yml
environment:
  - FABRIC_CERT_PATH=/app/test-network/organizations/peerOrganizations/org1.example.com
```

### Step 2: 修改网络连接代码
```go
// network/connection.go
// 修改三个函数: NewGrpcConnection, NewIdentity, NewSign
```

### Step 3: 强制重新构建镜像
```bash
docker-compose down
docker-compose up -d --build
```

### Step 4: 验证修复
```bash
docker-compose ps          # 检查容器状态
docker-compose logs novel-api  # 查看日志
```

## ✅ 验证结果

### 修复前
```
STATUS: Restarting (1)
错误: "TLS certificate file not found"
前端: {"error":"Proxy request failed"}
```

### 修复后
```
STATUS: Up 25 seconds (health: starting)
日志: "MongoDB connection timeout"  # 新问题，表示证书问题已解决
```

## 📋 监控命令

### 容器状态监控
```bash
# 实时查看容器状态
docker-compose ps

# 查看容器资源使用
docker stats novel-api

# 实时查看日志
docker-compose logs -f novel-api
```

### 证书文件验证
```bash
# 进入容器检查证书文件
docker-compose exec novel-api ls -la /app/test-network/organizations/peerOrganizations/org1.example.com/

# 检查特定证书文件
docker-compose exec novel-api ls -la /app/test-network/organizations/peerOrganizations/org1.example.com/tlsca/
```

### API健康检查
```bash
# 测试API是否可用
curl -s http://localhost:8080/health

# 测试完整状态
curl -s http://localhost:8080/api/v1/novels
```

## 🎯 关键经验教训

### 1. 路径管理最佳实践
- **Docker容器内避免使用相对路径**
- **使用环境变量配置外部依赖路径**
- **明确挂载路径和容器内路径的映射关系**

### 2. 开发工作流优化
```bash
# 开发时的调试流程
docker-compose down
docker-compose up -d --build    # 强制重新构建
docker-compose logs -f novel-api # 实时查看日志
```

### 3. 错误排查思路
1. **检查容器状态** - 是否在重启
2. **查看详细日志** - 具体错误信息
3. **验证挂载路径** - 文件是否存在
4. **测试环境变量** - 配置是否生效
5. **逐步验证** - 从简单到复杂

## 🔄 相关问题

### 后续遇到的MongoDB连接问题

**现象**:
```
MongoDB自动连接失败: server selection error: server selection timeout
```

**原因**: 容器内无法访问宿主机的127.0.0.1:27017

**解决方案**: 使用 `host.docker.internal` 替代 `127.0.0.1`

## 📚 相关文档

- [Docker部署指南](DOCKER_DEPLOYMENT.md)
- [Docker核心概念](DOCKER_CONCEPTS.md)

---

**问题解决时间**: 2025-11-24
**版本**: v1.0
**状态**: Fabric证书问题已解决 ✅