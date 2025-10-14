# RSA 加密中间件使用指南

# todo，明天先看下代码，再试试调用 get all

## 实现概述

novel-resource-management 项目现在已经集成了 RSA 加密中间件，支持与 Next.js 项目进行安全的加密通信。

## 工作流程

1. **Next.js 前端** → 使用公钥加密请求数据 → 发送到 Go 接口
2. **Go 中间件** → 自动检测加密请求 → 使用私钥解密 → 将解密数据传递给原有处理函数
3. **Go 处理函数** → 正常处理业务逻辑 → 返回响应
4. **可选** → Go 中间件加密响应 → 返回给 Next.js

## 中间件特性

### RSARequestMiddleware

- 自动检测加密请求（通过`X-Encrypted-Request: true`头）
- 自动解密请求体中的`encryptedData`字段
- 将解密后的数据重新设置到请求体中
- 对原有 API 处理函数透明

### 支持的路由

当前已应用 RSA 中间件的路由：

- `POST /api/v1/users` - 创建用户积分
- `PUT /api/v1/users/:id` - 更新用户积分

## Next.js 调用示例

### 1. 加密请求数据

```javascript
// 准备要发送的数据
const userData = {
  userId: "user123",
  credit: 100,
  totalUsed: 0,
  totalRecharge: 100,
};

// 使用公钥加密数据
const encryptedData = await clientRsaUtils.encrypt(JSON.stringify(userData));

// 发送加密请求
const response = await fetch("http://localhost:8080/api/v1/users", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-Encrypted-Request": "true", // 标识这是加密请求
  },
  body: JSON.stringify({
    encryptedData: encryptedData,
  }),
});
```

### 2. 请求格式

加密请求的 JSON 格式：

```json
{
  "encryptedData": "Base64编码的加密数据"
}
```

### 3. 响应格式

目前返回原始响应格式，与普通 API 调用相同。

## 安全特性

1. **密钥管理**: 私钥安全存储在 Go 服务器端
2. **算法兼容**: 使用 RSA-OAEP with SHA-1，与 Next.js 项目保持一致
3. **透明处理**: 原有 API 代码无需修改，中间件自动处理加解密
4. **选择性应用**: 只有指定的路由应用 RSA 中间件

## 测试建议

1. 先测试普通 API 调用（不加密）
2. 测试加密 API 调用
3. 验证数据正确解密和处理
4. 检查错误处理机制

## 扩展更多路由

如需为更多路由添加 RSA 加密支持，只需将路由添加到`encryptedRoutes`组中：

```go
encryptedRoutes := s.router.Group("/")
encryptedRoutes.Use(middleware.RSARequestMiddleware())
{
    encryptedRoutes.POST("/api/v1/users", s.createUserCredit)
    encryptedRoutes.PUT("/api/v1/users/:id", s.updateUserCredit)
    encryptedRoutes.POST("/api/v1/novels", s.createNovel)  // 新增
    // 添加更多需要加密的路由
}
```
