# 充值接口测试指南

## 接口信息

**接口路径:** `POST /api/v1/users/recharge`

**服务地址:** `http://localhost:8080`

## 测试方法

### 方法 1: 使用 curl

```bash
curl -X POST http://localhost:8080/api/v1/users/recharge \
  -H "Content-Type: application/json" \
  -d '{
    "title": "150 Token 充值包",
    "order_sn": "ORDER20250126001",
    "email": "beetle5249@gmail.com",
    "actual_price": 150,
    "order_info": "用户充值账号",
    "good_id": "GOOD_001",
    "gd_name": "150 Token套餐"
  }'
```

### 方法 2: 使用 Postman

1. **创建新请求**
   - Method: `POST`
   - URL: `http://localhost:8080/api/v1/users/recharge`

2. **设置 Headers**
   ```
   Content-Type: application/json
   ```

3. **设置 Body (选择 raw + JSON)**
   ```json
   {
     "title": "150 Token 充值包",
     "order_sn": "ORDER20250126001",
     "email": "beetle5249@gmail.com",
     "actual_price": 150,
     "order_info": "用户充值账号",
     "good_id": "GOOD_001",
     "gd_name": "150 Token套餐"
   }
   ```

4. **发送请求**

## 期望响应

### 成功响应 (200 OK)
```json
{
  "message": "充值成功",
  "userId": "691058f50987397c91e4e078",
  "email": "beetle5249@gmail.com",
  "orderSn": "ORDER20250126001",
  "goodName": "150 Token套餐",
  "addedTokens": 150,
  "newCredit": 194
}
```

### 失败响应 (用户不存在)
```json
{
  "error": "用户不存在: nonexist@example.com"
}
```

### 失败响应 (参数错误)
```json
{
  "error": "请求参数错误: Key: 'RechargeRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag"
}
```

## 验证步骤

### 1. 启动服务
```bash
cd novel-resource-management
go run main.go
```

确认看到以下日志:
```
🚀 Starting Fabric Gateway API Server...
📋 Available endpoints:
  POST   /api/v1/users/recharge       <- 充值接口
```

### 2. 查询当前积分
```bash
curl http://localhost:8080/api/v1/users/691058f50987397c91e4e078
```

记录当前 `credit` 值,例如: `44`

### 3. 执行充值
发送充值请求 (参考上面的 curl 或 Postman 方法)

### 4. 验证充值结果
再次查询积分:
```bash
curl http://localhost:8080/api/v1/users/691058f50987397c91e4e078
```

确认:
- `credit` 应该是原值 + 150 (例如: 44 + 150 = 194)
- `totalRecharge` 应该是原值 + 150

### 5. 检查日志
在服务端日志中应该看到:
```
📥 收到充值回调: email=beetle5249@gmail.com, order_sn=ORDER20250126001, actual_price=150, good_name=150 Token套餐
✅ 找到用户: email=beetle5249@gmail.com, userId=691058f50987397c91e4e078
✅ MongoDB 同步更新成功
✅ 充值成功: userId=691058f50987397c91e4e078, 增加token=150, 新积分=194
```

## 第三方平台集成示例

### PHP 调用示例
```php
<?php
$postdata = [
    'title' => $this->order->title,
    'order_sn' => $this->order->order_sn,
    'email' => $this->order->email,
    'actual_price' => $this->order->actual_price,
    'order_info' => $this->order->info,
    'good_id' => $goodInfo->id,
    'gd_name' => $goodInfo->gd_name
];

$ch = curl_init('http://localhost:8080/api/v1/users/recharge');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($postdata));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json'
]);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($httpCode === 200) {
    echo "充值成功: " . $response;
} else {
    echo "充值失败: " . $response;
}
?>
```

### Python 调用示例
```python
import requests
import json

url = 'http://localhost:8080/api/v1/users/recharge'
data = {
    'title': '150 Token 充值包',
    'order_sn': 'ORDER20250126001',
    'email': 'beetle5249@gmail.com',
    'actual_price': 150,
    'order_info': '用户充值账号',
    'good_id': 'GOOD_001',
    'gd_name': '150 Token套餐'
}

response = requests.post(url, json=data)

if response.status_code == 200:
    print("充值成功:", response.json())
else:
    print("充值失败:", response.text)
```

## 常见问题

### Q1: 提示"用户不存在"
**原因:** MongoDB users 集合中没有该邮箱的记录

**解决:**
1. 检查邮箱是否正确
2. 确认用户已在系统中注册

### Q2: 充值成功但 MongoDB 没有更新
**原因:** MongoDB 连接问题或同步失败

**解决:**
1. 检查 MongoDB 服务是否运行
2. 查看服务端日志中的错误信息
3. 确认环境变量 `MONGODB_URI` 配置正确

### Q3: 链码更新失败
**原因:** Fabric 网络连接问题或链码未部署

**解决:**
1. 检查 Fabric 网络是否运行
2. 确认链码已正确部署和实例化
3. 查看服务端日志中的详细错误信息

## 数据一致性检查

### 检查链上数据
使用 Fabric CLI 或区块链浏览器查询链上状态

### 检查 MongoDB 数据
```javascript
// MongoDB shell
use novel
db.user_credits.findOne({userId: "691058f50987397c91e4e078"})
```

确认:
- `credit` 字段已更新
- `totalRecharge` 字段已更新
- `updatedAt` 时间戳是最新的
