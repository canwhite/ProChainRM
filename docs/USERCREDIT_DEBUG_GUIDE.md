# UserCredit 更新问题调试指南

## 问题描述
UserCredit 数据在 MongoDB 中没有更新，可能是没有监听到事件或者同步过程中出现问题。

## 🔍 调试步骤

### 1. 检查事件监听器是否正常工作

启动你的事件监听服务，观察控制台输出：

```bash
# 运行你的服务
go run main.go
```

**期望看到的输出**：
```
🎧 Starting event listener...
✅ MongoDB自动连接成功! 数据库: novel
🔍 开始为数据库创建索引...
📚 为 novels 集合创建 storyOutline 索引...
✅ novels 集合的 storyOutline 索引创建成功
💰 为 user_credits 集合创建 userId 索引...
✅ user_credits 集合的 userId 索引创建成功
📜 为 credit_histories 集合创建 userId + timestamp 复合索引...
✅ credit_histories 集合的 userId-timestamp 索引创建成功
🎉 所有数据库索引创建完成！查询速度将会大幅提升
```

### 2. 执行 UserCredit 更新操作

通过你的 API 或直接调用链码方法来更新 UserCredit。

### 3. 观察调试输出

现在代码中已经添加了详细的调试信息，你应该能看到类似这样的输出：

**如果没有看到任何输出**：
- ❌ 事件没有被监听到
- ❌ 事件名称不匹配

**如果看到了输出**：
```
🔍 [DEBUG] 接收到事件: UpdateUserCredit
📦 [DEBUG] 事件载荷长度: 123 字节
📋 [DEBUG] 解析后的事件数据:
   userId: user123 (类型: string)
   credit: 95 (类型: float64)
   totalUsed: 5 (类型: float64)
   totalRecharge: 100 (类型: float64)
   updatedAt: 2024-01-15T10:35:00Z (类型: string)
💰 [DEBUG] 处理 UpdateUserCredit 事件...
🔍 [DEBUG] UserCredit 数据 - userId: user123, credit: 95, totalUsed: 5, totalRecharge: 100
✅ [DEBUG] UpdateUserCredit 同步到 MongoDB 成功!
```

## 🚨 常见问题排查

### 问题1：没有看到任何事件输出

**可能原因**：
1. 事件监听器没有启动
2. 事件监听器连接的链码名称不对
3. 智能合约没有发出事件

**排查方法**：
1. 确认链码名称是否正确
```go
// 检查 event_service.go 中的链码名称
events, err := es.network.ChaincodeEvents(ctx, "novel-basic") // 确认这个名称正确
```

2. 检查智能合约是否真的发出了事件
3. 确认事件监听器是否在正确的通道上

### 问题2：事件名称不匹配

**可能原因**：
智能合约发出的事件名称与代码中处理的事件名称不一致。

**排查方法**：
查看调试输出中的事件名称，对比代码中的 switch case：

```go
switch eventName {
case "UpdateUserCredit":  // 确认这个名称与链码中的事件名称一致
    es.handleUpdateUserCreditEvent(eventData)
// ...
}
```

**常见的事件名称不匹配**：
- `UpdateUserCredit` vs `UpdateUserCreditData`
- `updateUserCredit` vs `UpdateUserCredit`
- `userCreditUpdated` vs `UpdateUserCredit`

### 问题3：事件数据格式不正确

**可能原因**：
事件载荷中的字段名与代码期望的不一致。

**排查方法**：
查看调试输出中的 `解析后的事件数据` 部分，确认字段名是否正确：

```
📋 [DEBUG] 解析后的事件数据:
   userId: user123 (类型: string)  ← 应该是 userId
   credit: 95 (类型: float64)      ← 应该是 credit
   // 确认字段名完全匹配
```

### 问题4：MongoDB 更新失败

**可能原因**：
1. MongoDB 连接问题
2. 数据格式转换问题
3. 字段匹配问题

**排查方法**：
检查是否有错误输出：
```
❌ Failed to sync UpdateUserCredit to MongoDB: 具体错误信息
```

## 🛠️ 手动测试方法

### 测试1：直接查询 MongoDB

```javascript
// 连接到 MongoDB
mongo novel

// 查看当前 user_credits 集合
db.user_credits.find().pretty()

// 查看 user_credits 集合的索引
db.user_credits.getIndexes()
```

### 测试2：手动插入测试数据

```javascript
// 手动插入一条 UserCredit 记录
db.user_credits.insertOne({
    userId: "test_user_123",
    credit: 100,
    totalUsed: 0,
    totalRecharge: 100,
    createdAt: "2024-01-15T10:30:00Z",
    updatedAt: "2024-01-15T10:30:00Z"
})

// 查看插入结果
db.user_credits.find({userId: "test_user_123"})
```

### 测试3：检查链码事件

如果你能直接调用链码，可以这样测试：

```go
// 在链码中确认事件发出
func (s *SmartContract) UpdateUserCredit(ctx contractapi.TransactionContextInterface, userId string, credit int, totalUsed int, totalRecharge int) error {
    // ... 更新逻辑 ...

    // 确保事件发出
    eventPayload := map[string]interface{}{
        "userId": userId,
        "credit": credit,
        "totalUsed": totalUsed,
        "totalRecharge": totalRecharge,
        "updatedAt": time.Now().Format(time.RFC3339),
    }
    payloadBytes, _ := json.Marshal(eventPayload)
    ctx.GetStub().SetEvent("UpdateUserCredit", payloadBytes) // 确认事件名称正确

    return nil
}
```

## 📋 调试检查清单

在排查问题时，请检查以下每一项：

- [ ] 事件监听服务是否正常启动？
- [ ] MongoDB 连接是否成功？
- [ ] 索引是否创建成功？
- [ ] 智能合约是否真的调用了 UpdateUserCredit？
- [ ] 链码是否发出了 `UpdateUserCredit` 事件？
- [ ] 事件名称是否完全匹配（大小写敏感）？
- [ ] 事件载荷中是否包含需要的字段？
- [ ] 字段名是否完全匹配？
- [ ] 数据类型是否正确？
- [ ] MongoDB 更新操作是否执行？
- [ ] 是否有任何错误信息输出？

## 🎯 最可能的问题

根据经验，最常见的问题是：

1. **事件名称不匹配**：链码发出的事件名称与代码中处理的不一致
2. **字段名不匹配**：事件载荷中的字段名与代码期望的不一致
3. **事件监听器没有启动**：服务没有正常运行

## 📞 如果问题仍然存在

如果按照以上步骤仍然无法解决问题，请提供以下信息：

1. 完整的控制台输出
2. 智能合约中的事件发送代码
3. 你调用 UpdateUserCredit 的具体方法
4. MongoDB 中的当前数据状态

这样我可以更准确地帮你定位问题！

---

*记住：调试就像侦探工作，需要一步步排除可能性，最终找到问题的根源。*