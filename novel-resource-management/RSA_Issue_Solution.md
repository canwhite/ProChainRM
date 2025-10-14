# RSA加密问题解决文档

## 问题背景

在实现Next.js和Go项目之间的RSA加密通信时，遇到了以下问题：
```
panic: runtime error: invalid memory address or nil pointer dereference
```

## 问题解决过程

### 1. 初始问题分析
- 项目需要实现Next.js和Go之间的RSA加密通信
- 测试RSA加密功能时出现空指针异常
- 错误发生在 `utils/rsa.go` 的 `Encrypt` 方法中

### 2. 错误的调试方向
一开始我以为是以下问题：
- 公钥文件格式错误
- 公钥解析失败
- 内存初始化问题
- 密钥文件路径问题

### 3. 真正的原因发现
通过深度调试程序 `debug_rsa_deep.go`，发现：
- **公钥本身完全正常**：PEM解码成功，PKIX解析成功，N和E都正确
- **问题出在函数调用上**：`rsa.EncryptOAEP(sha1.New(), nil, r.publicKey, []byte(data), nil)` 中的第二个参数 `random` 传入了 `nil`

### 4. 根本原因
`rsa.EncryptOAEP` 函数的签名：
```go
func EncryptOAEP(hash hash.Hash, random io.Reader, pub *PublicKey, msg []byte, label []byte) ([]byte, error)
```

- 第二个参数 `random` 必须是一个有效的随机数生成器
- 传入 `nil` 会导致函数内部尝试读取空指针，引发崩溃
- 必须使用 `crypto/rand.Reader` 作为随机数源

## 主要修改内容

### 修改文件：`utils/rsa.go`

**修改前：**
```go
// 使用RSA-OAEP加密，使用SHA-1哈希与Next.js保持一致
encrypted, err := rsa.EncryptOAEP(sha1.New(), nil, r.publicKey, []byte(data), nil)
```

**修改后：**
```go
// 使用RSA-OAEP加密，使用SHA-1哈希与Next.js保持一致
// nil 参数会导致空指针异常，必须使用 crypto/rand.Reader
encrypted, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, r.publicKey, []byte(data), nil)
```

**添加的导入：**
```go
import (
    "crypto/rand"  // 新增导入
    "crypto/rsa"
    "crypto/sha1"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "fmt"
    "log"
    "os"
)
```

## RSA在项目中的作用

### 1. 系统架构
```
Next.js项目 [有公钥]  <-- 加密数据 -->  Go项目 [有私钥]
      ^                             |
      |                             |
      +-------- 解密数据 <----------+
```

### 2. 关键组件

#### A. RSA密钥对
- **私钥**：`security/rsa_private_key.pem` - Go项目持有，用于解密
- **公钥**：`security/rsa_public_key.pem` - Next.js项目持有，用于加密

#### B. RSA工具类 (`utils/rsa.go`)
```go
type RSACrypto struct {
    privateKey *rsa.PrivateKey  // 私钥，用于解密
    publicKey  *rsa.PublicKey   // 公钥，用于加密
}
```

#### C. 中间件 (`middleware/rsa.go`)
- 检查请求头 `X-Encrypted-Request: true`
- 如果是加密请求，自动解密数据
- 将解密后的数据传递给业务逻辑

### 3. 工作流程

#### 加密过程（Next.js → Go）：
1. Next.js将数据序列化为JSON
2. 使用公钥加密JSON数据
3. 发送加密数据到Go接口，请求头包含 `X-Encrypted-Request: true`
4. Go中间件检测到加密请求
5. 使用私钥解密数据
6. 将解密后的数据传给业务逻辑

#### 加解密代码：
```go
// 加密
encrypted, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, data, nil)

// 解密
decrypted, err := rsa.DecryptOAEP(sha1.New(), nil, privateKey, encryptedData, nil)
```

### 4. 为什么用RSA-OAEP

#### OAEP (Optimal Asymmetric Encryption Padding) 的优点：
- **安全性高**：比原始的RSA加密更安全
- **抗攻击**：可以抵抗已知的攻击方式
- **标准化**：广泛使用的加密标准

#### 为什么用SHA-1：
- 为了与Next.js项目保持一致
- 确保两边的加密/解密算法匹配

### 5. 关键要点

1. **密钥安全**：私钥只存在于Go项目，不会暴露给外部
2. **透明性**：业务代码不需要关心加密/解密，中间件自动处理
3. **兼容性**：使用标准的RSA-OAEP算法，确保与Next.js的兼容性
4. **随机性**：加密时使用随机数生成器，确保同样的明文加密后结果不同

### 6. 通信示例

**Next.js发送的加密请求：**
```json
{
  "encryptedData": "FiMVKeKLQy26eWl6qm1L0uNozS+r1udHfF49KMyxQ8FEdpnB2Z..."
}
```

**Go中间件解密后的数据：**
```json
{
  "userId": "test_rsa_user_001",
  "credit": 100,
  "totalUsed": 0,
  "totalRecharge": 100
}
```

## 测试验证

### 测试RSA工具类
```bash
go run debug_rsa_v2.go
```

### 测试完整功能（需要先启动服务器）
```bash
go run test_rsa_simple.go
```

## 经验总结

1. **不要忽视函数参数要求**：即使是看似可选的参数也要仔细阅读文档
2. **深度调试很重要**：有时候问题不是表面看到的那个
3. **理解函数签名**：每个参数都有其特定的作用和限制
4. **随机数在加密中的重要性**：RSA加密需要随机数来确保安全性

通过这次修复，RSA加密通信系统现在完全正常工作！🎉