# MongoDB 事件同步服务

## 概述

MongoDB 事件同步服务实现了智能合约事件与 MongoDB 数据库之间的实时数据同步。当智能合约发出特定事件时，系统会自动将相关数据同步到 MongoDB 中，确保链上链下数据的一致性。

## 架构设计

```
智能合约事件 → 事件监听器 → 事件处理器 → MongoDB操作 → 数据库同步
```

### 核心组件

1. **EventService** - 事件监听和处理
2. **MongoService** - MongoDB 数据库操作
3. **数据模型** - 与链码保持一致的数据结构

## 文件结构

```
service/
├── event_service.go      # 事件监听和同步逻辑
├── mongo_service.go      # MongoDB CRUD 操作
└── MONGODB_SYNC_README.md # 本文档
```

## 支持的事件类型

| 事件名称 | 描述 | MongoDB集合 | 操作类型 |
|---------|------|------------|----------|
| `CreateNovel` | 创建小说 | `novels` | INSERT |
| `UpdateNovel` | 更新小说 | `novels` | UPSERT |
| `CreateUserCredit` | 创建用户积分 | `user_credits` | INSERT |
| `UpdateUserCredit` | 更新用户积分 | `user_credits` | UPSERT |
| `CreateCreditHistory` | 创建积分历史 | `credit_histories` | INSERT |
| `ConsumeUserToken` | 消费用户代币 | `user_credits` | UPDATE |

## 数据模型

### Novel 小说模型
```go
type Novel struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Author       string             `bson:"author,omitempty" json:"author,omitempty"`
    StoryOutline string             `bson:"storyOutline,omitempty" json:"storyOutline,omitempty"`
    Subsections  string             `bson:"subsections,omitempty" json:"subsections,omitempty"`
    Characters   string             `bson:"characters,omitempty" json:"characters,omitempty"`
    Items        string             `bson:"items,omitempty" json:"items,omitempty"`
    TotalScenes  string             `bson:"totalScenes,omitempty" json:"totalScenes,omitempty"`
    CreatedAt    string             `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
    UpdatedAt    string             `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
```

### UserCredit 用户积分模型
```go
type UserCredit struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID        string             `bson:"userId" json:"userId"`
    Credit        int                `bson:"credit" json:"credit"`
    TotalUsed     int                `bson:"totalUsed" json:"totalUsed"`
    TotalRecharge int                `bson:"totalRecharge" json:"totalRecharge"`
    CreatedAt     string             `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
    UpdatedAt     string             `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
```

### CreditHistory 积分历史模型
```go
type CreditHistory struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID      string             `bson:"userId" json:"userId"`
    Amount      int                `bson:"amount" json:"amount"`
    Type        string             `bson:"type" json:"type"`
    Description string             `bson:"description" json:"description"`
    Timestamp   string             `bson:"timestamp" json:"timestamp"`
    NovelID     string             `bson:"novelId,omitempty" json:"novelId,omitempty"`
}
```

## 数据库索引

为确保查询性能，系统会自动创建以下索引：

### novels 集合
- `author`: 唯一索引，用于快速查找小说

### user_credits 集合
- `userId`: 唯一索引，用于快速查找用户积分

### credit_histories 集合
- `userId` + `timestamp`: 复合索引，用于按用户和时间排序查询历史记录

## 使用方法

### 1. 初始化服务

```go
// 创建网关连接
gateway, err := client.Connect(networkConfig)
if err != nil {
    log.Fatalf("Failed to connect to network: %v", err)
}
defer gateway.Close()

// 创建事件服务（会自动初始化MongoDB连接）
eventService := NewEventService(gateway)
```

### 2. 启动事件监听

```go
// 启动通用事件监听
ctx := context.Background()
err = eventService.StartEventListening(ctx)
if err != nil {
    log.Fatalf("Failed to start event listening: %v", err)
}

// 或者监听特定事件类型
eventNames := []string{"CreateNovel", "UpdateNovel", "CreateUserCredit", "UpdateUserCredit"}
err = eventService.ListenForSpecificEvents(ctx, eventNames)
if err != nil {
    log.Fatalf("Failed to start specific event listening: %v", err)
}
```

### 3. 环境配置

在 `.env` 文件中配置 MongoDB 连接参数：

```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=novel_rm
MONGODB_TIMEOUT=10s
MONGODB_MAX_POOL_SIZE=10
MONGODB_MIN_POOL_SIZE=2
MONGODB_MAX_CONN_IDLE_TTL=30m
```

## 日志输出示例

```
🎧 Starting event listener...
✅ MongoDB自动连接成功! 数据库: novel_rm
✅ MongoDB indexes created successfully

<-- Chaincode event received: CreateNovel - {
  "author": "张三",
  "storyOutline": "一个关于冒险的故事",
  "createdAt": "2024-01-15T10:30:00Z"
}
📝 Processing CreateNovel event...
✅ Created novel in MongoDB: author=张三

<-- Chaincode event received: CreateUserCredit - {
  "userId": "user123",
  "credit": 100,
  "totalUsed": 0,
  "totalRecharge": 100,
  "createdAt": "2024-01-15T10:31:00Z"
}
💰 Processing CreateUserCredit event...
✅ Created user credit in MongoDB: userId=user123, credit=100
```

## 错误处理

系统具备完善的错误处理机制：

1. **连接错误**: MongoDB 连接失败时会记录错误日志并重试
2. **数据解析错误**: 事件载荷解析失败时记录错误并跳过处理
3. **重复数据处理**: 检测并跳过已存在的记录，避免重复插入
4. **索引创建错误**: 索引创建失败时记录警告但不影响主要功能

## 性能考虑

1. **连接池**: 使用 MongoDB 连接池管理数据库连接
2. **索引优化**: 为常用查询字段创建索引
3. **异步处理**: 事件处理采用异步方式，不阻塞主流程
4. **批量操作**: 可以根据需要扩展为批量操作以提高性能

## 监控和维护

1. **日志监控**: 通过日志观察同步状态和错误情况
2. **数据一致性**: 定期检查链上链下数据一致性
3. **性能监控**: 监控 MongoDB 查询性能和同步延迟
4. **备份策略**: 制定 MongoDB 数据备份和恢复策略

## 扩展性

系统设计支持以下扩展：

1. **新事件类型**: 在 `processEventAndSyncToMongoDB` 方法中添加新的事件处理逻辑
2. **数据转换**: 在 MongoService 中添加自定义数据转换逻辑
3. **多数据库支持**: 扩展支持其他类型的数据库（如 MySQL、PostgreSQL）
4. **事件过滤**: 添加事件过滤和路由机制

## 故障排除

### 常见问题

1. **MongoDB 连接失败**
   - 检查 MongoDB 服务是否运行
   - 验证连接字符串和认证信息
   - 检查网络连接和防火墙设置

2. **事件解析失败**
   - 检查智能合约事件数据格式
   - 验证 JSON 解析逻辑
   - 查看事件载荷是否符合预期

3. **数据同步延迟**
   - 检查事件监听器是否正常运行
   - 验证 MongoDB 写入性能
   - 检查网络延迟情况

### 调试技巧

1. 启用详细日志输出
2. 使用 MongoDB 客户端工具直接查询数据
3. 检查智能合约事件是否正常发出
4. 验证数据模型字段匹配情况

## 最佳实践

1. **定期测试**: 定期测试事件同步功能的完整性
2. **备份重要数据**: 定期备份 MongoDB 中的重要数据
3. **监控告警**: 设置关键指标的监控和告警
4. **文档更新**: 及时更新代码文档和操作手册
5. **版本控制**: 对数据库模式变更进行版本控制