# 部署脚本使用说明（Linux优化版）

## 概述

`deploy.go` 是一个自动化的Go部署脚本，专门针对Linux环境进行了优化，同时也支持其他操作系统。

## 主要改进

### 1. 智能操作系统检测
- 自动检测Linux环境并应用优化配置
- 针对不同操作系统使用不同的IP选择策略

### 2. Linux优化的IP选择策略
```bash
# Linux环境优先级：
1. 192.168.x.x  (最常见的Linux内网)
2. 10.x.x.x     (企业网络)
3. 172.16.x.x   (特定偏好网段)
4. 172.x.x.x    (其他私有网络)
```

### 3. 增强的Docker网络过滤
- 覆盖Docker默认网桥网络：172.17.x.x
- 包含Docker自定义网络：172.18-31.x.x
- 过滤Docker Desktop（Mac）：192.168.65.x
- 排除回环地址：127.x.x.x

### 4. 详细的调试日志
- 网络接口状态信息
- IP选择过程日志
- Docker网络过滤详情

## 使用方法

### 基本使用
```bash
# 编译并运行
go run scripts/deploy.go
```

### 启用调试模式
```bash
# 显示详细的网络接口信息
DEBUG_NETWORK=true go run scripts/deploy.go
```

### 自定义环境变量
```bash
# 指定.env文件路径
ENV_PATH=/custom/path/.env go run scripts/deploy.go

# 设置MongoDB认证信息
export MONGO_USER=admin
export MONGO_PASS=password
go run scripts/deploy.go
```

## 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DEBUG_NETWORK` | 启用详细网络调试信息 | `false` |
| `ENV_PATH` | 自定义.env文件路径 | `../.env` |
| `MONGO_USER` | MongoDB用户名 | `admin` |
| `MONGO_PASS` | MongoDB密码 | `password` |

## Linux部署特殊处理

### 网络接口支持
脚本支持常见的Linux网络接口：
- `eth0` - 传统以太网接口
- `ens33` - 新式命名规范
- `enp0s3` - 新式命名规范
- 其他标准Linux网络接口

### 备用IP策略
- **Linux环境**: `192.168.1.100` (常见网段)
- **其他环境**: `172.16.181.101` (原有设置)

## 故障排除

### 1. 获取不到正确的IP
```bash
# 启用调试模式查看详细信息
DEBUG_NETWORK=true go run scripts/deploy.go
```

### 2. MongoDB连接失败
- 检查MongoDB服务状态：`systemctl status mongod`
- 验证认证信息：检查环境变量 `MONGO_USER` 和 `MONGO_PASS`
- 检查防火墙设置

### 3. Docker部署问题
- 确保Docker服务运行：`systemctl status docker`
- 检查docker-compose.yml配置
- 查看容器日志：`docker-compose logs`

## 部署流程

1. **环境检测** - 自动识别操作系统
2. **IP获取** - 智能选择最佳IP地址
3. **MongoDB配置** - 自动配置副本集
4. **Docker部署** - 启动容器服务
5. **健康检查** - 验证服务可用性

## 示例输出

```bash
🐧 检测到Linux环境，应用Linux优化配置
🔍 开始获取宿主机IP地址...
📋 找到 2 个网络接口
  - 检查接口: eth0
    ✅ 发现有效IP: 192.168.1.50 (来自接口: eth0)
🎯 Linux环境优先选择192.168网段IP: 192.168.1.50
✅ 使用优先选择的IP: 192.168.1.50

🚀 开始自动化部署novel-resource-management...
✅ 宿主机IP: 192.168.1.50
✅ MongoDB副本集配置完成
✅ Docker部署完成
🎉 自动化部署完成!
```