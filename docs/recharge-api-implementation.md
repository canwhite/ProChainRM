# 用户充值接口实现方案

## 需求描述

实现一个充值回调接口,接收第三方平台的充值通知,根据用户邮箱增加对应的 token。

## 接口设计

### 接口路径
```
POST /api/v1/users/recharge
```

### 请求体
```json
{
  "title": "商品标题",
  "order_sn": "订单号20250101",
  "email": "user@example.com",           // ← 用这个字段识别用户
  "actual_price": 150,                   // ← 充值金额(暂时忽略)
  "order_info": "用户填写的充值账号",
  "good_id": "商品ID",
  "gd_name": "商品名称"
}
```

### 响应体
**成功:**
```json
{
  "message": "充值成功",
  "userId": "691058f50987397c91e4e078",
  "email": "user@example.com",
  "addedTokens": 150,
  "newCredit": 194
}
```

**失败:**
```json
{
  "error": "用户不存在"
}
```

## 数据结构分析

### MongoDB 集合关系

#### 1. `users` 集合
```json
{
  "_id": "691058f50987397c91e4e078",
  "email": "beetle5249@gmail.com",
  "username": "admin",
  "passwordHash": "$2b$12$...",
  "deviceFingerprint": "1c8d99c3",
  "isActive": true,
  "role": "USER",
  "novelIds": [...],
  "createdAt": "2025-11-09T09:03:49.107Z",
  "updatedAt": "2025-12-25T04:40:58.097Z"
}
```

#### 2. `user_credits` 集合
```json
{
  "_id": "1763794965367909000",
  "userId": "691058f50987397c91e4e078",  // ← 对应 users._id
  "credit": 44,
  "totalUsed": 56,
  "totalRecharge": 0,
  "createdAt": "2025-11-22 07:02:43",
  "updatedAt": "2025-12-16 02:55:01"
}
```

### 关键映射关系
- `users.email` → 用户邮箱 (唯一标识)
- `users._id` → MongoDB ObjectId
- `user_credits.userId` → 对应 `users._id`

## 实现逻辑

```
1. 接收完整请求体 (包含所有字段但只用 email)
   ↓
2. 提取 email 字段
   ↓
3. 从 MongoDB users 集合查询: {email: "xxx"}
   ↓
4. 如果用户不存在,返回错误
   ↓
5. 获取 users._id 作为 userId
   ↓
6. 从 user_credits 读取当前积分信息
   ↓
7. 计算新的积分:
   - credit = credit + 150
   - totalRecharge = totalRecharge + 150
   ↓
8. 调用链码 UpdateUserCredit 更新区块链
   ↓
9. 同步更新 MongoDB user_credits 集合
   ↓
10. 返回成功响应
```

## 实现步骤

### 步骤 1: 添加 User 结构体定义
**文件:** `database/models.go`

添加 `User` 结构体,用于映射 MongoDB users 集合。

### 步骤 2: 实现 Service 层方法
**文件:** `service/user_service.go`

添加 `AddTokensByEmail(email string, amount int)` 方法:
- 从 MongoDB users 查询用户
- 验证用户是否存在
- 读取 user_credits 当前积分
- 更新链码和 MongoDB

### 步骤 3: 实现 API 层接口
**文件:** `api/server.go`

添加 `POST /api/v1/users/recharge` 接口:
- 解析请求体 (接收所有字段)
- 调用 Service 层方法
- 返回响应

### 步骤 4: 测试
使用测试工具验证接口功能。

## 注意事项

1. **固定充值金额**: 当前固定增加 150 token,忽略 `actual_price` 字段
2. **用户不存在**: 返回明确错误信息,不自动创建用户
3. **并发安全**: 链码层面已经处理并发冲突
4. **数据同步**: 确保链码和 MongoDB 数据一致性

## 文件修改清单

- [x] `database/models.go` - 添加 User 结构体
- [ ] `service/user_service.go` - 添加 AddTokensByEmail 方法
- [ ] `api/server.go` - 添加 /recharge 接口
- [ ] 测试验证

## 实现时间

- 预计完成时间: 30分钟
- 测试时间: 15分钟
