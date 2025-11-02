# Fabric 背书策略问题分析与解决方案

## 问题概述

在调用 PUT `/api/v1/novels/:id` 接口更新小说时，出现 500 错误：

```
{
    "error": "failed to update novel novel_1759829953718_ojrasoe1q: rpc error: code = Aborted desc = failed to collect enough transaction endorsements, see attached details for more info"
}
```

## 根本原因

### 背书策略配置分析

#### 1. **默认背书策略位置**
- **配置文件**: `/test-network/network.config`
- **配置项**: `CC_END_POLICY="NA"` (第40-41行)

#### 2. **脚本中的默认策略说明**
在 `/test-network/scripts/utils.sh` 第78行明确说明：
```bash
println "    -ccep <policy>  - (Optional) Chaincode endorsement policy using signature policy syntax. The default policy requires an endorsement from Org1 and Org2"
```

#### 3. **链码部署流程分析**
`/test-network/scripts/deployCC.sh` 部署脚本显示：
- 第78行：在 Org1 安装链码
- 第80行：在 Org2 安装链码
- 第88行：Org1 批准链码定义
- 第96行：Org2 批准链码定义
- 第100行：两个组织都批准后才提交链码定义

#### 4. **实际部署情况**
根据 `novel-basic.tar.gz` 文件的存在，推测 novel-basic 链码是使用类似以下命令部署的：

```bash
./network.sh deployCC \
  -ccn novel-basic \
  -ccp ../novel-resource-events/chaincode \
  -ccl go \
  -ccv 1.0
```

**关键问题**: 没有指定 `-ccep` 参数，因此使用了默认的 **"Org1 AND Org2"** 背书策略。

### 应用配置分析

#### 1. **应用连接配置**
`/novel-resource-management/main.go` 中的应用只连接到单一网关：
```go
network := gateway.GetNetwork("mychannel")
```

#### 2. **交易处理流程**
当应用调用 `s.contract.SubmitTransaction("UpdateNovel", ...)` 时：
1. Fabric SDK 会根据链码的背书策略收集背书
2. 默认策略要求 Org1 和 Org2 都背书
3. 但应用只连接了一个组织的节点
4. 导致无法收集到足够的背书，交易失败

## 解决方案

### 方案1：修改链码背书策略（推荐）

#### 1.1 **重新部署链码为单组织背书**

```bash
cd /Users/zack/Desktop/ProChainRM/test-network

# 使用单个组织背书策略重新部署链码
./network.sh deployCC \
  -ccn novel-basic \
  -ccp ../novel-resource-events/chaincode \
  -ccl go \
  -ccv 1.0 \
  -ccep "AND('Org1MSP.member')"  # 只需要 Org1 背书
```

#### 1.2 **或者使用多数策略**

```bash
./network.sh deployCC \
  -ccn novel-basic \
  -ccp ../novel-resource-events/chaincode \
  -ccl go \
  -ccv 2.0 \
  -ccep "OR('Org1MSP.member','Org2MSP.member')"  # 任一组织背书即可
```

### 方案2：修改应用支持多组织连接

#### 2.1 **修改应用配置**

在 `/novel-resource-management/main.go` 中添加多组织支持：

```go
// 为每个组织创建网关连接
gatewayOrg1, err := client.Connect(idOrg1, client.WithSign(sign), client.WithClientConnection(connOrg1))
gatewayOrg2, err := client.Connect(idOrg2, client.WithSign(sign), client.WithClientConnection(connOrg2))

// 创建支持多组织的网络客户端
networkOrg1 := gatewayOrg1.GetNetwork("mychannel")
networkOrg2 := gatewayOrg2.GetNetwork("mychannel")
```

#### 2.2 **修改服务层**

在 `novel_service.go` 中实现多组织交易：

```go
func (s *NovelService) UpdateNovelMultiOrg(id, author, storyOutline, subsections, characters, items, totalScenes string) error {
    // 获取多个组织的合约
    contractOrg1 := s.networkOrg1.GetContract("novel-basic")
    contractOrg2 := s.networkOrg2.GetContract("novel-basic")

    // 提交交易到两个组织
    _, err := contractOrg1.SubmitTransaction("UpdateNovel", id, author, storyOutline, subsections, characters, items, totalScenes)
    if err != nil {
        return fmt.Errorf("failed to update novel from Org1: %w", err)
    }

    // 这里需要协调两个组织的交易...
    return nil
}
```

### 方案3：修改网络配置为单组织模式

#### 3.1 **修改网络配置**

编辑 `/test-network/network.config`:

```bash
# 修改默认背书策略
CC_END_POLICY="AND('Org1MSP.member')"

# 或者修改为多数策略
CC_END_POLICY="OR('Org1MSP.member','Org2MSP.member')"
```

#### 3.2 **使用修改后的配置重新部署**

```bash
cd /Users/zack/Desktop/ProChainRM/test-network

# 停止网络
./network.sh down

# 启动网络
./network.sh up createChannel

# 重新部署链码（会使用新的默认策略）
./network.sh deployCC \
  -ccn novel-basic \
  -ccp ../novel-resource-events/chaincode \
  -ccl go
```

### 方案4：使用通道级别的默认策略

#### 4.1 **检查通道默认策略**

```bash
# 获取通道配置
docker exec peer0.org1.example.com peer channel fetch config /tmp/config_block.pb -o orderer.example.com:7050 -c mychannel --tls --cafile /etc/hyperledger/fabric/msp/config/tls/ca.crt

# 解码配置
docker exec peer0.org1.example.com peer channel decodeconfigblock /tmp/config_block.pb > /tmp/config.json
```

#### 4.2 **修改通道应用策略**

如果通道的 `/Channel/Application/Writers` 策略过于严格，可以考虑放宽。

## 推荐操作步骤

### 立即解决方案

1. **重新部署链码**（最简单有效）：
```bash
cd /Users/zack/Desktop/ProChainRM/test-network

# 停止现有链码（如果需要）
./network.sh cc remove -ccn novel-basic -c mychannel

# 重新部署为单组织背书
./network.sh deployCC \
  -ccn novel-basic \
  -ccp ../novel-resource-events/chaincode \
  -ccl go \
  -ccv 2.0 \
  -ccep "AND('Org1MSP.member')"
```

2. **验证链码状态**：
```bash
# 检查链码列表
./network.sh cc list -org 1

# 测试链码功能
docker exec peer0.org1.example.com peer chaincode query -C mychannel -n novel-basic -c '{"Args":["GetAllNovels"]}' --tls
```

### 长期解决方案

1. **修改网络配置文件**，将 `CC_END_POLICY` 设置为适合你业务需求的策略
2. **更新部署文档**，明确说明背书策略的选择和影响
3. **考虑使用环境变量**，根据不同环境（开发/测试/生产）使用不同的背书策略

## 验证方法

重新部署后，验证PUT请求：

```bash
# 先创建一个测试小说
curl -X POST http://localhost:8080/api/v1/novels \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_novel_endorsement",
    "author": "测试作者",
    "storyOutline": "测试大纲",
    "subsections": "章节1",
    "characters": "测试角色",
    "items": "测试道具",
    "totalScenes": "10"
  }'

# 然后测试更新
curl -X PUT http://localhost:8080/api/v1/novels/test_novel_endorsement \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_novel_endorsement",
    "author": "更新后的作者",
    "storyOutline": "更新后的大纲",
    "subsections": "更新章节1,更新章节2",
    "characters": "更新角色",
    "items": "更新道具",
    "totalScenes": "20"
  }'
```

如果返回200状态码，说明背书策略问题已解决！

## 相关文件位置

- **网络配置**: `/test-network/network.config`
- **部署脚本**: `/test-network/scripts/deployCC.sh`
- **工具脚本**: `/test-network/scripts/utils.sh`
- **应用配置**: `/novel-resource-management/main.go`
- **服务层**: `/novel-resource-management/service/novel_service.go`
- **链码代码**: `/novel-resource-events/chaincode/smartcontract.go`

## 总结

问题的根本原因是链码部署时使用了默认的双组织背书策略，而应用只连接了单个组织。最直接的解决方案是使用 `-ccep` 参数重新部署链码，指定单组织背书策略。