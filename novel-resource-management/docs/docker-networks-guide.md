# Docker网络配置简单指南

## 🤔 什么是Docker网络？

想象每个容器是一个独立的房间：
- **默认情况**：房间之间互相看不见，不能通信
- **有了网络**：像建了一座桥，让房间之间可以互相访问

## 📖 你的配置解析

### 网络定义（建房子）
```yaml
networks:
  # 内部网络（自己家的Wi-Fi）
  novel-network:
    driver: bridge           # 网络类型：桥接模式
    ipam:
      config:
        - subnet: 192.168.200.0/24  # IP地址范围

  # 外部网络（别人的Wi-Fi）
  fabric_test:
    external: true           # 已经存在的网络
```

### 服务使用（住进去）
```yaml
services:
  novel-api:
    networks:
      - novel-network        # 连接到自家网络
      - fabric_test          # 同时连接到外部网络
```

## 🏠 通俗比喻

### 就像你有两个Wi-Fi：
1. **家里Wi-Fi**（novel-network）
   - 自己建的
   - 连接家里设备
   - IP地址：192.168.200.x

2. **公司Wi-Fi**（fabric_test）
   - 公司建的
   - 连接公司设备
   - 加入已有网络

### 容器就像你的手机：
- 可以同时连家里Wi-Fi和公司Wi-Fi
- 在不同网络里用不同的服务

## 🎯 实际效果

### 网络连通图：
```
            ┌─────────────────┐
            │   novel-network │  ← 自家网络
            │  192.168.200.0/24│
            └─────────┬───────┘
                      │
          ┌──────────┴──────────┐
          │                     │
┌─────────▼─────────┐   ┌────────▼────────┐
│   novel-api       │   │  Fabric网络      │
│   容器            │◄──►│  (peer容器们)   │
│  有两个网络接口    │   │                 │
└───────────────────┘   └─────────────────┘
```

### 容器能力：
```bash
# 在novel-api容器内
ping mongodb              # 通过novel-network访问
ping peer0.org1.example.com  # 通过fabric_test访问
```

## 🔧 为什么要这样配置？

### 1. 网络隔离
- **自家网络**：只放自己的服务，更安全
- **外部网络**：只连接必要的区块链服务

### 2. IP管理
- **固定网段**：192.168.200.0/24，避免IP冲突
- **可控范围**：你知道每个服务的IP范围

### 3. 灵活连接
- 一个容器可以连接多个网络
- 像有多个网络接口一样

## 📋 检查你的网络

### 查看项目网络：
```bash
docker-compose ps

# 查看网络详情
docker network ls
docker network inspect novel-resource-management_novel-network
```

### 查看容器网络：
```bash
docker inspect novel-api | grep Networks -A 10

# 输出示例：
# "Networks": {
#   "fabric_test": {...},
#   "novel-resource-management_novel-network": {...}
# }
```

## 💡 常见问题

### Q1: 为什么需要两个网络？
**A**:
- novel-network：连接你自己的服务
- fabric_test：连接区块链网络
- 分开管理，更安全清晰

### Q2: external: true 是什么意思？
**A**: 这个网络已经存在，不需要创建，直接加入即可

### Q3: 如果不指定subnet会怎样？
**A**: Docker会自动分配，但IP可能不固定，指定subnet更可控

## 🎉 总结

简单说就是：
- **建两个网络世界**：一个自建，一个外部
- **让容器多栖生存**：同时连接两个网络
- **实现跨网络通信**：既能访问自家服务，又能访问区块链

这就是让容器"既能在家办公，又能去公司开会"的配置！