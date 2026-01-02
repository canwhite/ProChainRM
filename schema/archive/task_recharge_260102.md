# Task: 调研 novel-resource-management 下的 recharge 接口

**任务ID**: task_recharge_260102
**创建时间**: 2026-01-02
**状态**: 已完成
**目标**: 全面理解 recharge 接口的实现机制、调用流程和测试方法

---

## 最终目标

产出一份完整的 recharge 接口调研报告，包括：
1. 接口定义和参数
2. 实现逻辑分析
3. 链码调用流程
4. 测试方法和示例
5. 潜在问题和建议

---

## 拆解步骤

### 1. 定位接口定义
- [x] 1.1 在 API 层找到 recharge 路由定义
- [x] 1.2 查看请求参数和响应格式
- [x] 1.3 确认接口路径和方法

### 2. 分析实现逻辑
- [x] 2.1 阅读 Handler 函数实现
- [x] 2.2 查看调用链码的方法
- [x] 2.3 理解业务逻辑流程

### 3. 查看链码实现
- [x] 3.1 找到链码中的充值逻辑
- [x] 3.2 分析链码参数验证
- [x] 3.3 理解链上数据结构

### 4. 检查测试文件
- [x] 4.1 阅读 test_recharge.go
- [x] 4.2 查看测试用例
- [x] 4.3 提取调用示例

### 5. 生成调研报告
- [x] 5.1 整合所有信息
- [x] 5.2 编写完整调研报告
- [x] 5.3 提供使用建议

---

## 当前进度

### ✅ 任务已完成
所有调研步骤已全部完成，已生成完整调研报告。

---

## 下一步行动

无（任务已完成）

---

## 任务总结

### 完成时间
2026-01-02

### 产出成果
1. ✅ 完整的接口定义文档
2. ✅ 三层架构实现分析（API → Service → Chaincode）
3. ✅ 完整调用链路图
4. ✅ 测试方法和示例
5. ✅ 潜在问题分析和改进建议

### 核心发现
- 接口路径: `POST /api/v1/users/recharge`
- 固定充值 150 Token（忽略 actual_price）
- 通过邮箱识别用户
- 链上链下双重更新机制
- 事件驱动架构

### 关键文件位置
- API 层: `api/server.go:554-599`
- Service 层: `service/user_service.go:136-211`
- Chaincode 层: `novel-resource-events/chaincode/smartcontract.go:372-405`
- 测试文件: `test_recharge.go`

### 改进建议
1. 根据 actual_price 动态计算充值金额
2. 添加幂等性保护（使用 order_sn）
3. 实现 HMAC 签名验证
4. 增加充值记录表
5. 完善错误处理和重试机制

---

## 调研笔记

### ✅ 已完成调研内容

#### 1. API 层 (server.go)
- 路由定义: `POST /api/v1/users/recharge` (line:108)
- Handler: `rechargeUserTokens` (line:554-599)
- 请求参数:
  - title (string)
  - order_sn (string)
  - email (string, required)
  - actual_price (int)
  - order_info (string)
  - good_id (string)
  - gd_name (string)
- 固定充值: 150 token (忽略 actual_price)

#### 2. Service 层 (user_service.go)
- 方法: `AddTokensByEmail` (line:136-211)
- 业务逻辑:
  1. 从 MongoDB users 集合查询用户获取 userId
  2. 读取当前用户积分信息
  3. 计算新积分 (credit + amount)
  4. 调用链码 UpdateUserCredit 更新链上数据
  5. 同步更新 MongoDB user_credits 集合

#### 3. Chaincode 层 (smartcontract.go)
- 方法: `UpdateUserCredit` (line:372-405)
- 功能:
  1. 读取现有用户积分数据
  2. 更新 credit, totalUsed, totalRecharge 字段
  3. 设置 UpdatedAt 时间戳
  4. 触发 UpdateUserCredit 事件
  5. 保存到区块链状态

#### 4. 测试文件 (test_recharge.go)
- 测试邮箱: beetle5249@gmail.com
- 测试用户ID: 691058f50987397c91e4e078
- 测试流程:
  1. 健康检查
  2. 查询当前积分
  3. 发送充值请求
  4. 验证积分增加
  5. 测试错误处理

---

## 完整调用链路

```
第三方支付平台回调
    ↓
POST /api/v1/users/recharge
    ↓
rechargeUserTokens (Handler)
    ↓
AddTokensByEmail (Service)
    ↓
1. MongoDB查询用户 (users 集合)
2. 读取链上积分 (ReadUserCredit)
3. 更新链上积分 (UpdateUserCredit)
4. 同步MongoDB (user_credits 集合)
    ↓
UpdateUserCredit (Chaincode)
    ↓
更新区块链状态 + 触发事件
```

---
