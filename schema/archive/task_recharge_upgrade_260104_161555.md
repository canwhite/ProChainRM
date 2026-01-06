woz# Task: Recharge 接口升级方案设计

**任务ID**: task_recharge_upgrade_260104_161555
**创建时间**: 2026-01-04
**状态**: 进行中
**目标**: 分析现有 recharge 接口的问题，设计并输出完整的升级方案

---

## 最终目标

产出一套完整的 recharge 接口升级方案，包括：
1. 现有问题深度分析
2. 升级方案设计（安全、可靠、可扩展）
3. 实施步骤和代码改进
4. 测试验证方案

---

## 拆解步骤

### 1. 深度分析现有实现
- [ ] 1.1 分析 API 层调用思路
- [ ] 1.2 分析 Service 层业务逻辑
- [ ] 1.3 识别关键问题和风险点
- [ ] 1.4 评估当前设计的优缺点

### 2. 设计升级方案
- [ ] 2.1 安全性改进（签名验证、幂等性）
- [ ] 2.2 业务逻辑改进（动态充值金额）
- [ ] 2.3 数据一致性改进（事务性、补偿机制）
- [ ] 2.4 可观测性改进（日志、监控、审计）

### 3. 制定实施计划
- [ ] 3.1 优先级排序
- [ ] 3.2 具体改进措施
- [ ] 3.3 代码实现方案
- [ ] 3.4 测试和验证方案

### 4. 输出完整方案文档
- [ ] 4.1 问题清单
- [ ] 4.2 升级方案详细设计
- [ ] 4.3 实施路线图
- [ ] 4.4 风险评估和回滚方案

---

## 当前进度

### 正在进行: 识别关键问题和风险点

已完成 API 层和 Service 层的深度分析，正在整理问题清单和风险评估。

---

## 调研笔记

### 现有实现分析

#### 1. API 层调用思路 (server.go:554-599)
```
调用流程:
第三方支付平台回调
    ↓
POST /api/v1/users/recharge
    ↓
rechargeUserTokens Handler
    ├─ 解析请求参数 (title, order_sn, email, actual_price, etc.)
    ├─ 固定充值金额 = 150 (忽略 actual_price)
    └─ 调用 Service 层 AddTokensByEmail(email, 150)
```

**关键发现**:
- ✅ 接收了完整的第三方支付回调参数
- ❌ 忽略了 `actual_price`，固定充值 150 token
- ❌ 接收了 `order_sn` 但未使用
- ❌ 无任何安全验证机制

#### 2. Service 层业务逻辑 (user_service.go:136-211)
```
调用流程:
AddTokensByEmail(email, amount)
    ↓
1. MongoDB 查询用户 (users 集合) → userId
    ↓
2. 读取链上积分 (ReadUserCredit) → userCredit
    ↓
3. 计算新积分:
   - newCredit = credit + amount
   - newTotalRecharge = totalRecharge + amount
    ↓
4. 更新链码 (UpdateUserCredit)
    ↓
5. 同步 MongoDB (user_credits 集合)
```

**关键发现**:
- ✅ 链上链下双重更新
- ⚠️ MongoDB 更新失败只记录警告，无补偿机制
- ❌ 无幂等性保护
- ❌ 无事务保证

#### 3. 数据结构
- **users**: 用户基本信息 (email → userId)
- **user_credits**: 用户积分信息
- **UserCredit**: Credit, TotalUsed, TotalRecharge, UpdatedAt

---

## 问题清单

### 🔴 严重问题 (P0 - 必须修复)

#### 1. 安全漏洞 - 无签名验证
**问题描述**:
- 任何人都可以调用充值接口
- 无请求来源验证
- 无防重放攻击机制

**风险**:
- 恶意用户可以伪造充值请求
- 可能导致大量资金损失

**影响**: 生产环境致命风险

---

#### 2. 幂等性问题 - 重复充值风险
**问题描述**:
- 接收了 `order_sn` 但未使用
- 相同订单重复回调会导致重复充值
- 无充值记录表记录已处理订单

**风险**:
- 第三方支付平台可能重复回调
- 网络重试可能导致重复处理
- 用户可以多次获得相同订单的 token

**场景示例**:
```
用户支付 10 元 → order_sn=ORDER123
    ↓
第1次回调: 充值 150 token ✅
    ↓
网络抖动，第2次回调: 又充值 150 token ✅❌ (错误!)
    ↓
结果: 用户支付 10 元，获得 300 token
```

**影响**: 资金损失，业务逻辑错误

---

### 🟡 重要问题 (P1 - 强烈建议修复)

#### 3. 业务逻辑缺陷 - 忽略实际支付金额
**问题描述**:
- 固定充值 150 token，忽略 `actual_price`
- 无法支持多种充值套餐
- 业务扩展性差

**当前代码** (server.go:579):
```go
// 固定充值 150 token (忽略 actual_price)
const rechargeAmount = 150
```

**影响**:
- 无法实现差异化定价
- 用户支付不同金额获得相同 token
- 商业逻辑受限

---

#### 4. 数据一致性问题 - 无补偿机制
**问题描述**:
- 链码更新成功，MongoDB 更新失败时只记录警告
- 无回滚机制
- 可能导致链上链下数据不一致

**当前代码** (user_service.go:201-206):
```go
if err != nil {
    log.Printf("⚠️ MongoDB 更新失败: %v", err)
    // 不返回错误,因为链码已经更新成功
} else {
    log.Printf("✅ MongoDB 同步更新成功")
}
```

**风险**:
- 链上积分已增加，MongoDB 未同步
- 用户查询积分时可能看到不一致的数据
- 难以排查和修复

---

### 🟢 改进建议 (P2 - 可选优化)

#### 5. 可观测性问题 - 缺少审计和监控
**问题描述**:
- 无充值记录表
- 无审计日志
- 无业务指标监控

**影响**:
- 难以追踪充值历史
- 难以排查问题
- 无法进行数据分析

---

#### 6. 用户体验问题 - 错误提示不友好
**问题描述**:
- 错误消息过于技术化
- 无具体的错误代码
- 难以进行国际化

**当前代码** (server.go:584-587):
```go
c.JSON(http.StatusInternalServerError, gin.H{
    "error": err.Error(),
})
```

---

## 风险评估

### 当前系统风险矩阵

| 问题 | 严重程度 | 发生概率 | 影响范围 | 优先级 |
|------|---------|---------|---------|--------|
| 无签名验证 | 🔴 高 | 🔴 高 | 资金安全 | P0 |
| 重复充值 | 🔴 高 | 🟡 中 | 资金安全 | P0 |
| 忽略支付金额 | 🟡 中 | 🟢 低 | 业务扩展 | P1 |
| 数据一致性 | 🟡 中 | 🟢 低 | 数据质量 | P1 |
| 缺少审计 | 🟢 低 | 🟢 低 | 运维效率 | P2 |

---

## 下一步行动

1. ✅ 完成问题识别
2. 设计升级方案
3. 制定实施计划
4. 输出完整方案文档

---

## 升级方案设计

### 总体设计原则

1. **安全性第一**: 所有外部调用必须验证签名
2. **幂等性保证**: 相同订单多次处理结果一致
3. **数据一致性**: 链上链下数据保持同步
4. **可扩展性**: 支持多种充值套餐
5. **可观测性**: 完整的审计和监控

---

### 方案 1: 安全性改进 (P0)

#### 1.1 HMAC 签名验证机制

**目标**: 验证请求来源的合法性，防止伪造请求

**设计方案**:

```go
// 配置密钥（从环境变量或配置中心读取）
const RECHARGE_SECRET_KEY = "your-secret-key-here"

// 1. 第三方平台在发送回调时计算签名
// signature = HMAC-SHA256(secret_key, sorted_params_string)

// 2. API 层验证签名
func validateSignature(params RechargeRequest, receivedSignature string) bool {
    // 按字母序排序参数
    sortedParams := sortParams(params)

    // 计算 HMAC-SHA256
    computedSignature := computeHMAC(sortedParams, RECHARGE_SECRET_KEY)

    // 对比签名
    return hmac.Equal([]byte(computedSignature), []byte(receivedSignature))
}
```

**请求格式**:
```json
{
  "title": "充值10元",
  "order_sn": "ORDER20260104123456",
  "email": "user@example.com",
  "actual_price": 1000,
  "order_info": "...",
  "good_id": "PACKAGE_10",
  "gd_name": "10元套餐",
  "timestamp": "1704356796",
  "signature": "abc123..."  // 新增字段
}
```

**优势**:
- ✅ 防止请求伪造
- ✅ 防止中间人攻击
- ✅ 验证数据完整性

**注意事项**:
- 密钥需要安全存储（建议使用环境变量或密钥管理服务）
- 定期轮换密钥
- 记录所有签名验证失败的请求

---

#### 1.2 时间戳验证（防重放攻击）

**设计方案**:
```go
const MAX_REQUEST_AGE = 5 * 60 // 5分钟

func validateTimestamp(timestamp int64) error {
    now := time.Now().Unix()
    age := now - timestamp

    if age < 0 {
        return fmt.Errorf("请求时间戳来自未来")
    }

    if age > MAX_REQUEST_AGE {
        return fmt.Errorf("请求过期，超过 %d 秒", MAX_REQUEST_AGE)
    }

    return nil
}
```

**优势**:
- ✅ 防止重放攻击
- ✅ 限制请求有效期

---

### 方案 2: 幂等性保证 (P0)

#### 2.1 充值记录表设计

**目标**: 记录所有充值请求，防止重复处理

**数据结构**:
```go
type RechargeRecord struct {
    ID           string    `bson:"_id" json:"id"`
    OrderSN      string    `bson:"orderSn" json:"orderSn"`      // 唯一索引
    UserID       string    `bson:"userId" json:"userId"`
    Email        string    `bson:"email" json:"email"`
    Amount       int       `bson:"amount" json:"amount"`         // 实际充值 token 数量
    ActualPrice  int       `bson:"actualPrice" json:"actualPrice"` // 支付金额（分）
    PackageID    string    `bson:"packageId" json:"packageId"`
    Status       string    `bson:"status" json:"status"`         // pending, success, failed
    ProcessedAt  time.Time `bson:"processedAt" json:"processedAt"`
    CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
    UpdatedAt    time.Time `bson:"updatedAt" json:"updatedAt"`
}
```

**数据库索引**:
```javascript
// MongoDB 索引
db.recharge_records.createIndex({ "orderSn": 1 }, { unique: true })
db.recharge_records.createIndex({ "userId": 1, "createdAt": -1 })
db.recharge_records.createIndex({ "status": 1 })
```

---

#### 2.2 幂等性处理逻辑

**改进后的 Service 层逻辑**:
```go
func (us *UserCreditService) AddTokensByEmailWithIdempotency(
    email string,
    orderSN string,
    actualPrice int,
    packageName string,
) (string, int, error) {

    // 1. 检查订单是否已处理
    existingRecord, err := us.findRechargeRecordByOrderSN(orderSN)
    if err == nil && existingRecord != nil {
        // 订单已处理，返回之前的结果
        if existingRecord.Status == "success" {
            log.Printf("⚠️ 订单已处理: orderSN=%s", orderSN)
            return existingRecord.UserID, existingRecord.Amount, nil
        }
        if existingRecord.Status == "failed" {
            return "", 0, fmt.Errorf("订单之前处理失败，请人工介入: %s", orderSN)
        }
    }

    // 2. 查询用户
    userId, err := us.findUserByEmail(email)
    if err != nil {
        us.createRechargeRecord(orderSN, "", email, 0, actualPrice, packageName, "failed")
        return "", 0, err
    }

    // 3. 创建充值记录（状态：pending）
    us.createRechargeRecord(orderSN, userId, email, 0, actualPrice, packageName, "pending")

    // 4. 计算充值金额（根据套餐配置）
    amount := us.calculateRechargeAmount(actualPrice, packageName)

    // 5. 读取当前积分
    userCredit, err := us.ReadUserCredit(userId)
    if err != nil {
        us.updateRechargeRecordStatus(orderSN, "failed")
        return userId, 0, fmt.Errorf("读取用户积分失败: %v", err)
    }

    // 6. 更新链码
    newCredit, totalRecharge := us.calculateNewCredit(userCredit, amount)
    err = us.UpdateUserCredit(userId, newCredit, 0, totalRecharge)
    if err != nil {
        us.updateRechargeRecordStatus(orderSN, "failed")
        return userId, 0, fmt.Errorf("更新链码失败: %v", err)
    }

    // 7. 同步 MongoDB
    err = us.syncMongoDB(userId, newCredit, totalRecharge)
    if err != nil {
        // 记录失败，但不回滚链码（需要补偿机制）
        log.Printf("⚠️ MongoDB 同步失败: %v", err)
        // 标记需要补偿任务
        us.createCompensationTask(userId, newCredit, totalRecharge)
    }

    // 8. 更新充值记录为成功
    us.updateRechargeRecord(orderSN, userId, email, amount, actualPrice, packageName, "success")

    log.Printf("✅ 充值成功: userId=%s, orderSN=%s, amount=%d", userId, orderSN, amount)
    return userId, newCredit, nil
}
```

**优势**:
- ✅ 完全防止重复充值
- ✅ 提供完整的充值历史
- ✅ 支持审计和排查

---

### 方案 3: 业务逻辑改进 (P1)

#### 3.1 充值套餐配置

**数据结构**:
```go
type RechargePackage struct {
    ID          string `bson:"_id" json:"id"`           // PACKAGE_10, PACKAGE_50, etc.
    Name        string `bson:"name" json:"name"`         // "10元套餐"
    Price       int    `bson:"price" json:"price"`       // 价格（分）
    TokenAmount int    `bson:"tokenAmount" json:"tokenAmount"` // 获得的 token 数量
    Bonus       int    `bson:"bonus" json:"bonus"`       // 赠送 token 数量
    IsActive    bool   `bson:"isActive" json:"isActive"`
    SortOrder   int    `bson:"sortOrder" json:"sortOrder"`
}
```

**套餐配置示例**:
```javascript
// recharge_packages 集合
[
  {
    _id: "PACKAGE_10",
    name: "10元套餐",
    price: 1000,
    tokenAmount: 150,
    bonus: 0,
    isActive: true,
    sortOrder: 1
  },
  {
    _id: "PACKAGE_50",
    name: "50元套餐",
    price: 5000,
    tokenAmount: 750,
    bonus: 50,  // 赠送 50 token
    isActive: true,
    sortOrder: 2
  },
  {
    _id: "PACKAGE_100",
    name: "100元套餐",
    price: 10000,
    tokenAmount: 1500,
    bonus: 200,  // 赠送 200 token
    isActive: true,
    sortOrder: 3
  }
]
```

**计算逻辑**:
```go
func (us *UserCreditService) calculateRechargeAmount(actualPrice int, packageID string) int {
    // 从数据库查询套餐配置
    pkg, err := us.findPackageByID(packageID)
    if err != nil {
        log.Printf("⚠️ 套餐不存在，使用默认计算: 1元=15token")
        return actualPrice / 100 * 15  // 默认：1元=15token
    }

    // 验证价格是否匹配
    if pkg.Price != actualPrice {
        log.Printf("⚠️ 价格不匹配: expected=%d, actual=%d", pkg.Price, actualPrice)
        // 可以选择：
        // 1. 使用 actual_price 重新计算
        // 2. 返回错误
        // 3. 使用套餐配置的 tokenAmount
    }

    return pkg.TokenAmount + pkg.Bonus
}
```

**优势**:
- ✅ 支持多种充值套餐
- ✅ 支持促销活动
- ✅ 业务逻辑清晰

---

### 方案 4: 数据一致性改进 (P1)

#### 4.1 补偿机制

**数据结构**:
```go
type CompensationTask struct {
    ID          string    `bson:"_id" json:"id"`
    UserID      string    `bson:"userId" json:"userId"`
    Credit      int       `bson:"credit" json:"credit"`
    TotalUsed   int       `bson:"totalUsed" json:"totalUsed"`
    TotalRecharge int     `bson:"totalRecharge" json:"totalRecharge"`
    Status      string    `bson:"status" json:"status"`  // pending, processing, completed, failed
    RetryCount  int       `bson:"retryCount" json:"retryCount"`
    LastError   string    `bson:"lastError" json:"lastError"`
    CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
    UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}
```

**补偿任务处理**:
```go
// 后台定时任务（每分钟执行一次）
func (us *UserCreditService) processCompensationTasks() {
    tasks := us.findPendingCompensationTasks()

    for _, task := range tasks {
        err := us.retryMongoDBSync(task)
        if err != nil {
            task.RetryCount++
            task.LastError = err.Error()
            task.Status = "failed"

            if task.RetryCount >= 5 {
                log.Printf("❌ 补偿任务失败，超过最大重试次数: taskId=%s", task.ID)
                // 发送告警
                us.sendAlert(task)
            }
        } else {
            task.Status = "completed"
            log.Printf("✅ 补偿任务成功: taskId=%s", task.ID)
        }

        us.updateCompensationTask(task)
    }
}
```

**优势**:
- ✅ 自动修复不一致数据
- ✅ 提供重试机制
- ✅ 支持告警通知

---

#### 4.2 数据一致性检查

**定期对账任务**:
```go
func (us *UserCreditService) reconcileUserData() {
    // 1. 从链码读取所有用户积分
    chaincodeCredits := us.getAllCreditsFromChaincode()

    // 2. 从 MongoDB 读取所有用户积分
    mongoCredits := us.getAllCreditsFromMongo()

    // 3. 对比差异
    for userId, chaincodeCredit := range chaincodeCredits {
        mongoCredit := mongoCredits[userId]

        if chaincodeCredit.Credit != mongoCredit.Credit {
            log.Printf("⚠️ 数据不一致: userId=%s, chaincode=%d, mongo=%d",
                userId, chaincodeCredit.Credit, mongoCredit.Credit)

            // 创建修复任务
            us.createRepairTask(userId, chaincodeCredit)
        }
    }
}
```

---

### 方案 5: 可观测性改进 (P2)

#### 5.1 审计日志

**数据结构**:
```go
type AuditLog struct {
    ID        string    `bson:"_id" json:"id"`
    UserID    string    `bson:"userId" json:"userId"`
    Action    string    `bson:"action" json:"action"`  // recharge, consume, etc.
    OrderSN   string    `bson:"orderSn" json:"orderSn"`
    Amount    int       `bson:"amount" json:"amount"`
    Before    int       `bson:"before" json:"before"`  // 变动前积分
    After     int       `bson:"after" json:"after"`    // 变动后积分
    IP        string    `bson:"ip" json:"ip"`
    UserAgent string    `bson:"userAgent" json:"userAgent"`
    CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
```

---

#### 5.2 业务指标监控

**关键指标**:
```go
type Metrics struct {
    TotalRechargeCount    int64   // 总充值次数
    TotalRechargeAmount   int64   // 总充值金额（分）
    TotalTokenIssued      int64   // 总发放 token 数量
    AverageRechargeAmount float64 // 平均充值金额
    FailureRate           float64 // 失败率
    DuplicateOrderRate    float64 // 重复订单率
}
```

**监控查询**:
```javascript
// 今日充值统计
db.recharge_records.aggregate([
  { $match: { createdAt: { $gte: ISODate("2026-01-04") } } },
  { $group: {
      _id: null,
      totalAmount: { $sum: "$actualPrice" },
      totalTokens: { $sum: "$amount" },
      count: { $sum: 1 }
  }}
])
```

---

## 架构改进对比

### 改进前
```
第三方支付回调
    ↓
API: 解析参数
    ↓
Service: 固定充值 150 token
    ↓
Chaincode: 更新积分
    ↓
MongoDB: 同步（失败仅警告）
    ↓
返回结果

问题：
❌ 无签名验证
❌ 无幂等性保证
❌ 业务逻辑固化
❌ 数据一致性风险
❌ 无审计日志
```

### 改进后
```
第三方支付回调（带签名）
    ↓
API:
  ├─ 验证签名 ✅
  ├─ 验证时间戳 ✅
  └─ 解析参数
    ↓
Service:
  ├─ 检查订单幂等性 ✅
  ├─ 查询套餐配置 ✅
  ├─ 创建充值记录 ✅
  ├─ 计算充值金额 ✅
  └─ 调用链码更新
      ↓
Chaincode: 更新积分
    ↓
Service:
  ├─ 同步 MongoDB ✅
  ├─ 更新充值记录 ✅
  ├─ 记录审计日志 ✅
  └─ 创建补偿任务（如果需要）✅
      ↓
后台任务:
  ├─ 处理补偿任务 ✅
  └─ 定期对账 ✅
    ↓
返回结果

优势：
✅ 完整的安全验证
✅ 幂等性保证
✅ 灵活的业务逻辑
✅ 数据一致性保证
✅ 完整的审计追踪
```
