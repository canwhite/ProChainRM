# Fabric 资产服务指南

本指南详细介绍了如何使用 `AssetService` 与 Hyperledger Fabric 链码进行交互，实现资产的创建、查询、更新和删除操作。

## 资产服务架构概览

### 1. 服务组件

- **AssetService**: 封装与链码交互的核心服务
- **client.Contract**: Fabric 网关提供的合约接口
- **链码函数**: 链码中定义的智能合约方法
- **交易提交**: 向区块链网络提交交易
- **状态查询**: 从区块链账本查询数据

### 2. 服务结构详解

#### 2.1 服务定义 (`AssetService`)

```go
type AssetService struct {
    contract *client.Contract
}
```

**关键组件**:
- `contract`: Fabric 网关合约对象，提供链码调用能力
- 通过网关连接到特定的通道和链码

#### 2.2 服务初始化 (`NewAssetService`)

```go
func NewAssetService(gateway *client.Gateway) *AssetService {
    network := gateway.GetNetwork("mychannel")
    contract := network.GetContract("basic")
    
    return &AssetService{
        contract: contract,
    }
}
```

**初始化流程**:
1. **获取网络**: `gateway.GetNetwork("mychannel")` - 指定 Fabric 通道
2. **获取合约**: `network.GetContract("basic")` - 指定链码名称
3. **返回服务**: 创建包含合约的服务实例

### 3. 核心功能详解

#### 3.1 初始化账本 (`InitLedger`)

**目的**: 使用示例数据初始化区块链账本

**实现**:
```go
func (s *AssetService) InitLedger() error {
    fmt.Println("Submitting InitLedger transaction...")
    _, err := s.contract.SubmitTransaction("InitLedger")
    if err != nil {
        return fmt.Errorf("failed to init ledger: %w", err)
    }
    fmt.Println("✓ Ledger initialized successfully")
    return nil
}
```

**操作说明**:
- **交易类型**: SubmitTransaction - 会修改账本状态的交易
- **链码函数**: "InitLedger" - 链码中的初始化函数
- **无参数**: 此函数不接受额外参数
- **结果**: 在账本中创建预设的示例资产

#### 3.2 创建资产 (`CreateAsset`)

**目的**: 在区块链上创建新的数字资产

**实现**:
```go
func (s *AssetService) CreateAsset(id, color, size, owner, value string) error {
    fmt.Printf("Creating asset %s...\n", id)
    _, err := s.contract.SubmitTransaction("CreateAsset", id, color, size, owner, value)
    if err != nil {
        return fmt.Errorf("failed to create asset %s: %w", id, err)
    }
    fmt.Printf("✓ Asset %s created successfully\n", id)
    return nil
}
```

**参数说明**:
- `id`: 资产唯一标识符
- `color`: 资产颜色属性
- `size`: 资产尺寸属性  
- `owner`: 资产所有者
- `value`: 资产价值

**交易流程**:
1. 客户端提交交易提案
2. 节点背书并返回背书结果
3. 客户端收集足够的背书
4. 客户端向排序节点提交交易
5. 排序节点打包交易并分发
6. 节点验证并提交到账本

#### 3.3 查询所有资产 (`GetAllAssets`)

**目的**: 从区块链账本查询所有资产记录

**实现**:
```go
func (s *AssetService) GetAllAssets() (string, error) {
    fmt.Println("Querying all assets...")
    result, err := s.contract.EvaluateTransaction("GetAllAssets")
    if err != nil {
        return "", fmt.Errorf("failed to get all assets: %w", err)
    }
    return string(result), nil
}
```

**查询特点**:
- **交易类型**: EvaluateTransaction - 只读查询，不修改账本
- **即时响应**: 直接从节点本地账本查询，无需共识
- **无费用**: 查询操作不产生交易费用
- **结果格式**: 返回 JSON 格式的资产列表

#### 3.4 读取特定资产 (`ReadAsset`)

**目的**: 根据资产ID查询特定资产的详细信息

**实现**:
```go
func (s *AssetService) ReadAsset(id string) (string, error) {
    fmt.Printf("Reading asset %s...\n", id)
    result, err := s.contract.EvaluateTransaction("ReadAsset", id)
    if err != nil {
        return "", fmt.Errorf("failed to read asset %s: %w", id, err)
    }
    return string(result), nil
}
```

**参数说明**:
- `id`: 要查询的资产唯一标识符
- **返回**: 指定资产的详细信息（JSON格式）

#### 3.5 更新资产 (`UpdateAsset`)

**目的**: 更新区块链上现有资产的属性

**实现**:
```go
func (s *AssetService) UpdateAsset(id, color, size, owner, value string) error {
    fmt.Printf("Updating asset %s...\n", id)
    _, err := s.contract.SubmitTransaction("UpdateAsset", id, color, size, owner, value)
    if err != nil {
        return fmt.Errorf("failed to update asset %s: %w", id, err)
    }
    fmt.Printf("✓ Asset %s updated successfully\n", id)
    return nil
}
```

**更新逻辑**:
- **全量更新**: 提供资产的所有属性值
- **交易验证**: 验证更新者是否有权限修改该资产
- **版本控制**: 确保更新基于最新状态

#### 3.6 删除资产 (`DeleteAsset`)

**目的**: 从区块链账本中删除指定资产

**实现**:
```go
func (s *AssetService) DeleteAsset(id string) error {
    fmt.Printf("Deleting asset %s...\n", id)
    _, err := s.contract.SubmitTransaction("DeleteAsset", id)
    if err != nil {
        return fmt.Errorf("failed to delete asset %s: %w", id, err)
    }
    fmt.Printf("✓ Asset %s deleted successfully\n", id)
    return nil
}
```

**删除注意事项**:
- **逻辑删除 vs 物理删除**: 通常标记为已删除而非真正移除
- **权限验证**: 验证删除者是否有权限
- **审计追踪**: 删除操作会记录在区块链历史中

### 4. 交易类型对比

| 操作类型 | 函数名称 | 交易类型 | 账本修改 | 共识要求 | 响应时间 |
|----------|----------|----------|----------|----------|----------|
| 初始化 | `InitLedger` | SubmitTransaction | ✅ | ✅ | 慢 |
| 创建 | `CreateAsset` | SubmitTransaction | ✅ | ✅ | 慢 |
| 查询所有 | `GetAllAssets` | EvaluateTransaction | ❌ | ❌ | 快 |
| 查询单个 | `ReadAsset` | EvaluateTransaction | ❌ | ❌ | 快 |
| 更新 | `UpdateAsset` | SubmitTransaction | ✅ | ✅ | 慢 |
| 删除 | `DeleteAsset` | SubmitTransaction | ✅ | ✅ | 慢 |

### 5. 使用示例

```go
// 完整的资产操作流程示例
func demoAssetOperations() error {
    // 1. 建立网关连接
    connection, err := network.NewGrpcConnection()
    if err != nil {
        return err
    }
    defer connection.Close()
    
    // 2. 创建身份和签名
    id := network.NewIdentity()
    sign := network.NewSign()
    
    // 3. 连接网关
    gateway, err := client.Connect(
        id,
        client.WithSign(sign),
        client.WithClientConnection(connection),
    )
    if err != nil {
        return err
    }
    defer gateway.Close()
    
    // 4. 创建资产服务
    assetService := service.NewAssetService(gateway)
    
    // 5. 初始化账本
    if err := assetService.InitLedger(); err != nil {
        return err
    }
    
    // 6. 创建新资产
    if err := assetService.CreateAsset("asset1", "blue", "5", "Alice", "100"); err != nil {
        return err
    }
    
    // 7. 查询资产
    assets, err := assetService.GetAllAssets()
    if err != nil {
        return err
    }
    fmt.Println("所有资产:", assets)
    
    // 8. 读取特定资产
    asset, err := assetService.ReadAsset("asset1")
    if err != nil {
        return err
    }
    fmt.Println("资产详情:", asset)
    
    // 9. 更新资产
    if err := assetService.UpdateAsset("asset1", "red", "7", "Bob", "150"); err != nil {
        return err
    }
    
    // 10. 删除资产
    if err := assetService.DeleteAsset("asset1"); err != nil {
        return err
    }
    
    return nil
}
```

### 6. 错误处理机制

#### 6.1 常见错误类型

1. **链码错误**: 链码函数不存在或参数错误
2. **背书错误**: 背书节点拒绝交易
3. **提交错误**: 交易提交失败
4. **状态错误**: 资产不存在或已被删除
5. **权限错误**: 无权限执行操作

#### 6.2 错误处理策略

```go
// 错误处理示例
result, err := assetService.ReadAsset("nonexistent")
if err != nil {
    if strings.Contains(err.Error(), "does not exist") {
        fmt.Println("资产不存在")
    } else {
        fmt.Printf("查询失败: %v", err)
    }
}
```

### 7. 性能优化建议

#### 7.1 查询优化
- **缓存策略**: 对频繁查询的数据实施客户端缓存
- **分页查询**: 大量数据时使用分页查询
- **索引使用**: 确保链码中正确使用索引

#### 7.2 交易优化
- **批量操作**: 合并多个操作为单个交易
- **异步处理**: 非关键操作使用异步提交
- **错误重试**: 实现指数退避重试机制

### 8. 安全最佳实践

1. **参数验证**: 客户端验证所有输入参数
2. **权限检查**: 确保用户有执行操作的权限
3. **审计日志**: 记录所有关键操作
4. **异常监控**: 监控异常交易模式
5. **数据加密**: 敏感数据在传输中加密

### 9. 测试策略

#### 9.1 单元测试
```go
func TestCreateAsset(t *testing.T) {
    // 测试资产创建逻辑
}
```

#### 9.2 集成测试
```go
func TestAssetServiceIntegration(t *testing.T) {
    // 测试完整的服务流程
}
```

#### 9.3 性能测试
- **并发测试**: 测试高并发场景下的表现
- **负载测试**: 测试大量资产时的性能
- **压力测试**: 测试系统极限承载能力