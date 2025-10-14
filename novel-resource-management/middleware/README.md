# RSA中间件实现详解

## 1. 中间件是什么？

中间件就像是一个"门口保安"，在请求到达真正的处理函数之前，先进行检查和处理。

**生活中的例子：**
- 你要进入一个大楼（API调用）
- 保安先检查你的身份证（中间件检查请求头）
- 如果是加密的，就先解密（中间件解密）
- 然后让你正常进入大楼（继续处理请求）

## 2. 我写的中间件详细解释

### 文件位置：`middleware/rsa.go`

```go
// RSARequestMiddleware RSA请求解密中间件
func RSARequestMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 第1步：检查是否为加密请求
        if !isEncryptedRequest(c) {
            c.Next()  // 不是加密请求，直接放行
            return
        }

        // 第2步：读取并解密请求体
        decryptedBody, err := decryptRequestBody(c)
        if err != nil {
            c.JSON(400, gin.H{"error": "解密失败"})
            c.Abort()  // 停止处理
            return
        }

        // 第3步：将解密后的数据放回请求体
        c.Request.Body = io.NopCloser(strings.NewReader(decryptedBody))
        
        // 第4步：继续处理请求
        c.Next()
    }
}
```

**关键函数解释：**

1. `isEncryptedRequest(c)` - 检查请求头
```go
func isEncryptedRequest(c *gin.Context) bool {
    encrypted := c.GetHeader("X-Encrypted-Request")
    return encrypted == "true"  // 如果请求头有 X-Encrypted-Request: true
}
```

2. `decryptRequestBody(c)` - 解密请求体
```go
func decryptRequestBody(c *gin.Context) (string, error) {
    // 读取原始请求数据
    body, _ := io.ReadAll(c.Request.Body)
    
    // 解析JSON格式：{"encryptedData": "xxxx"}
    var encryptedRequest struct {
        EncryptedData string `json:"encryptedData"`
    }
    json.Unmarshal(body, &encryptedRequest)
    
    // 调用RSA解密函数
    decrypted, err := utils.DecryptWithRSA(encryptedRequest.EncryptedData)
    
    return decrypted, err
}
```

## 3. 中间件如何起作用？

### 在`api/server.go`中的配置：

```go
func (s *Server) setupRoutes() {
    // 创建一个路由组，应用RSA中间件
    encryptedRoutes := s.router.Group("/")
    encryptedRoutes.Use(middleware.RSARequestMiddleware())  // 应用中间件
    {
        // 这些路由都会经过RSA中间件处理
        encryptedRoutes.POST("/api/v1/users", s.createUserCredit)
        encryptedRoutes.PUT("/api/v1/users/:id", s.updateUserCredit)
    }
    
    // 这些路由不会经过RSA中间件
    s.router.GET("/health", s.healthCheck)
    novels := s.router.Group("/api/v1/novels")
    // ...
}
```

**执行流程：**
1. Next.js发送请求到 `POST /api/v1/users`
2. Gin框架发现这个路由有中间件
3. 先执行 `RSARequestMiddleware()`
4. 中间件解密数据后，调用 `c.Next()`
5. 继续执行 `s.createUserCredit` 函数
6. `createUserCredit` 函数拿到的是已经解密的数据

## 4. `utils/rsa.go` 实现的内容

### 核心结构体：
```go
// RSACrypto RSA加密解密器
type RSACrypto struct {
    privateKey *rsa.PrivateKey  // 私钥，用于解密
    publicKey  *rsa.PublicKey   // 公钥，用于加密
}
```

### 主要函数：

1. `NewRSACrypto()` - 创建RSA工具实例
```go
func NewRSACrypto() (*RSACrypto, error) {
    // 从文件读取私钥和公钥
    privateKeyPEM := getPrivateKey()  // 读取 security/rsa_private_key.pem
    publicKeyPEM := getPublicKey()    // 读取 security/rsa_public_key.pem
    
    // 解析PEM格式的密钥
    privateKey := parsePrivateKey(privateKeyPEM)
    publicKey := parsePublicKey(publicKeyPEM)
    
    return &RSACrypto{privateKey, publicKey}, nil
}
```

2. `Decrypt(encryptedBase64 string)` - 解密数据
```go
func (r *RSACrypto) Decrypt(encryptedBase64 string) (string, error) {
    // 第1步：Base64解码
    encryptedData, _ := base64.StdEncoding.DecodeString(encryptedBase64)
    
    // 第2步：RSA-OAEP解密（使用SHA-1哈希）
    decrypted, _ := rsa.DecryptOAEP(sha1.New(), nil, r.privateKey, encryptedData, nil)
    
    // 第3步：返回字符串
    return string(decrypted), nil
}
```

3. `Encrypt(data string)` - 加密数据
```go
func (r *RSACrypto) Encrypt(data string) (string, error) {
    // 第1步：RSA-OAEP加密（使用SHA-1哈希）
    encrypted, _ := rsa.EncryptOAEP(sha1.New(), nil, r.publicKey, []byte(data), nil)
    
    // 第2步：Base64编码
    result := base64.StdEncoding.EncodeToString(encrypted)
    
    return result, nil
}
```

4. 全局实例和便捷函数
```go
var globalRSACrypto *RSACrypto  // 全局RSA实例

// 初始化函数
func InitRSACrypto() error {
    globalRSACrypto, _ = NewRSACrypto()
    return nil
}

// 便捷的解密函数
func DecryptWithRSA(encryptedBase64 string) (string, error) {
    return globalRSACrypto.Decrypt(encryptedBase64)
}

// 便捷的加密函数
func EncryptWithRSA(data string) (string, error) {
    return globalRSACrypto.Encrypt(data)
}
```

## 5. 完整的数据流程

### Next.js发送加密请求：
```json
{
  "encryptedData": "xxxxxBase64编码的加密数据xxxxx"
}
```

### 中间件处理过程：
1. 检测到 `X-Encrypted-Request: true` 头
2. 读取请求体，提取 `encryptedData` 字段
3. 调用 `utils.DecryptWithRSA()` 解密
4. 将解密后的JSON数据重新设置到请求体中
5. 调用 `c.Next()` 继续处理

### API处理函数接收到：
```json
{
  "userId": "user123",
  "credit": 100,
  "totalUsed": 0,
  "totalRecharge": 100
}
```

**关键点：** API处理函数完全不知道数据被加密过，它接收到的就是正常的JSON数据！

## 6. 为什么这样设计？

1. **透明性**：现有的API代码不需要任何修改
2. **复用性**：同一个中间件可以应用到多个路由
3. **安全性**：密钥管理集中在一个地方
4. **灵活性**：可以选择性应用中间件到特定路由

---

## 7. 关于响应加密的说明

### 当前实现状态：

#### 请求处理：
- ✅ Next.js发送加密请求 → Go中间件自动解密 → API函数处理
- ✅ 中间件会检查 `X-Encrypted-Request: true` 头
- ✅ 自动解密 `encryptedData` 字段

#### 响应处理：
- ❌ Go返回明文响应 → Next.js直接接收
- ❌ 没有对响应进行加密

### 如果需要加密响应，可以实现的方案：

#### 方案1：简单实现 - 总是加密响应
```go
// 在中间件中加密所有响应
func RSAResponseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 拦截响应
        writer := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
        c.Writer = writer
        
        c.Next()  // 处理请求
        
        // 加密响应数据
        encrypted, _ := utils.EncryptWithRSA(writer.body.String())
        c.JSON(200, gin.H{"encryptedData": encrypted})
    }
}
```

#### 方案2：条件实现 - 根据请求决定是否加密响应
```go
// 只有加密请求才返回加密响应
func RSAResponseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        writer := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
        c.Writer = writer
        
        c.Next()
        
        // 检查原请求是否为加密请求
        if c.GetHeader("X-Encrypted-Request") == "true" {
            encrypted, _ := utils.EncryptWithRSA(writer.body.String())
            c.JSON(200, gin.H{"encryptedData": encrypted})
        } else {
            c.Data(200, c.ContentType(), writer.body.Bytes())
        }
    }
}
```

**考虑因素：**
- 如果响应需要加密，Next.js端需要相应的解密逻辑
- 当前实现已满足基本需求：解密Next.js的加密请求
- 可以根据实际需求决定是否实现响应加密