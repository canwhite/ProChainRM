# Database 模块详解

## 📁 目录结构

`novel-resource-management/database` 目录包含3个核心文件：
- `mongodb.go` - MongoDB 连接和管理
- `init.go` - 数据库初始化
- `models.go` - 数据模型定义

## 🗄️ 什么是 MongoDB？

MongoDB 是一个 NoSQL 数据库，与传统的关系型数据库（如 MySQL）不同：

### MySQL vs MongoDB 对比

| 特性 | MySQL | MongoDB |
|------|-------|---------|
| 数据结构 | 表格（固定行列） | 文档集合（JSON格式） |
| 模式 | 严格模式 | 灵活模式 |
| 扩展性 | 垂直扩展 | 水平扩展 |
| 查询语言 | SQL | MongoDB查询语言 |
| 适合场景 | 关系复杂、事务性强 | 文档存储、大数据量 |

## 🔧 核心组件解析

### 1. 配置管理 (`mongodb.go:14-34`)

```go
type MongoDBConfig struct {
    URI            string        // 数据库连接地址
    Database       string        // 数据库名称
    Timeout        time.Duration // 连接超时时间
    MaxPoolSize    uint64        // 最大连接池大小
    MinPoolSize    uint64        // 最小连接池大小
    MaxConnIdleTTL time.Duration // 连接空闲时间
}
```

**小白理解：** 这就像是配置数据库的"拨号设置"，告诉程序如何连接到数据库。

### 2. 默认配置 (`mongodb.go:25-34`)

```go
 func DefaultMongoDBConfig() *MongoDBConfig {
      return &MongoDBConfig{
          // 格式: mongodb://用户名:密码@主机:端口/?authSource=认证数据库，注意这样不安全，最好还是用.env
          URI:            "mongodb://myuser:mypassword@localhost:27017/?authSource=admin",
          Database:       "novel",
          Timeout:        10 * time.Second,
          MaxPoolSize:    10,
          MinPoolSize:    2,
          MaxConnIdleTTL: 30 * time.Minute,
      }
  }
```

**配置参数说明：**
- `URI`: MongoDB 服务器地址，默认本地 27017 端口
- `Database`: 数据库名称，这里是 `novel`
- `Timeout`: 连接超时时间，10秒
- `MaxPoolSize`: 最大连接数，防止连接过多
- `MinPoolSize`: 最小连接数，保证基本性能
- `MaxConnIdleTTL`: 连接空闲时间，超过30分钟自动关闭

### 3. 单例模式 (`mongodb.go:44-57`)

```go
var (
    mongoInstance *MongoDBInstance
    mongoOnce     sync.Once
)

func GetMongoInstance() *MongoDBInstance {
    mongoOnce.Do(func() {
        mongoInstance = &MongoDBInstance{
            config: DefaultMongoDBConfig(),
        }
    })
    return mongoInstance
}
```

**小白理解：** 单例模式确保整个程序只有一个数据库连接实例，避免重复创建连接浪费资源。就像一个家里只有一个路由器，大家都用同一个。

**优势：**
- 节省内存和资源
- 避免连接冲突
- 保证数据一致性

### 4. 连接管理 (`mongodb.go:68-105`)

```go
func (m *MongoDBInstance) Connect() error {
    // 设置客户端选项
    clientOptions := options.Client().ApplyURI(m.config.URI)
    clientOptions.SetMaxPoolSize(m.config.MaxPoolSize)
    clientOptions.SetMinPoolSize(m.config.MinPoolSize)
    // ... 其他配置

    // 连接并测试
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return fmt.Errorf("连接MongoDB失败: %v", err)
    }

    // 检查连接
    err = client.Ping(ctx, nil)
    if err != nil {
        return fmt.Errorf("MongoDB连接测试失败: %v", err)
    }

    m.client = client
    m.database = client.Database(m.config.Database)

    log.Printf("MongoDB连接成功! 数据库: %s", m.config.Database)
    return nil
}
```

**连接流程：**
1. 配置连接参数
2. 建立连接
3. 测试连接（Ping）
4. 保存连接实例

## 📊 数据模型详解 (`models.go`)

### 1. UserCredit (用户积分)

```go
type UserCredit struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID        string             `bson:"user_id" json:"user_id"`
    Credit        int                `bson:"credit" json:"credit"`
    TotalUsed     int                `bson:"total_used" json:"total_used"`
    TotalRecharge int                `bson:"total_recharge" json:"total_recharge"`
    IsActive      bool               `bson:"is_active" json:"is_active"`
    CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}
```

**小白理解：** 这就像用户的"钱包"记录

| 字段 | 说明 | 示例 |
|------|------|------|
| `UserID` | 用户唯一标识 | "user123" |
| `Credit` | 当前积分余额 | 500 |
| `TotalUsed` | 总消费积分 | 200 |
| `TotalRecharge` | 总充值积分 | 700 |
| `IsActive` | 账户是否激活 | true |

### 2. Novel (小说)

```go
type Novel struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title       string             `bson:"title" json:"title"`
    Author      string             `bson:"author" json:"author"`
    Category    string             `bson:"category" json:"category"`
    Content     string             `bson:"content" json:"content"`
    Description string             `bson:"description" json:"description"`
    Tags        []string           `bson:"tags" json:"tags"`
    Price       float64            `bson:"price" json:"price"`
    IsPublished bool               `bson:"is_published" json:"is_published"`
    ViewCount   int                `bson:"view_count" json:"view_count"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
```

**小白理解：** 这是小说的"基本信息卡"

| 字段 | 说明 | 示例 |
|------|------|------|
| `Title` | 小说标题 | "三体" |
| `Author` | 作者 | "刘慈欣" |
| `Category` | 分类 | "科幻" |
| `Tags` | 标签数组 | ["科幻", "硬科幻", "获奖作品"] |
| `Price` | 价格 | 29.9 |
| `ViewCount` | 浏览次数 | 1250 |

### 3. UserNovelPurchase (购买记录)

```go
type UserNovelPurchase struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID   string             `bson:"user_id" json:"user_id"`
    NovelID  string             `bson:"novel_id" json:"novel_id"`
    Price    float64            `bson:"price" json:"price"`
    PaidAt   time.Time          `bson:"paid_at" json:"paid_at"`
    Status   string             `bson:"status" json:"status"` // "completed", "pending", "failed"
}
```

**小白理解：** 这是"购物小票"

| 字段 | 说明 | 示例 |
|------|------|------|
| `UserID` | 购买用户 | "user123" |
| `NovelID` | 购买的小说ID | "novel456" |
| `Price` | 购买价格 | 19.9 |
| `PaidAt` | 购买时间 | 2024-01-15 14:30:00 |
| `Status` | 购买状态 | "completed" |

**状态说明：**
- `completed`: 购买完成
- `pending`: 待支付
- `failed`: 支付失败

### 4. UserActivity (用户活动日志)

```go
type UserActivity struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID    string             `bson:"user_id" json:"user_id"`
    Action    string             `bson:"action" json:"action"` // "login", "purchase", "read"
    TargetID  string             `bson:"target_id" json:"target_id"`
    TargetType string            `bson:"target_type" json:"target_type"` // "novel", "user"
    Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
```

**小白理解：** 这是用户的"行为日记"

| 字段 | 说明 | 示例 |
|------|------|------|
| `Action` | 用户行为 | "purchase" |
| `TargetID` | 操作对象ID | "novel456" |
| `TargetType` | 对象类型 | "novel" |
| `Metadata` | 额外信息 | `{"price": 19.9, "payment_method": "alipay"}` |

**常见行为类型：**
- `login`: 用户登录
- `purchase`: 购买小说
- `read`: 阅读小说
- `review`: 发表评论

## 🔍 标签解释

你可能注意到代码中有 `bson:"..."` 和 `json:"..."` 标签：

### BSON 标签
```go
UserID string `bson:"user_id" json:"user_id"`
```

- `bson:"field_name"`: 告诉 MongoDB 数据库中这个字段叫什么名字
- `json:"field_name"`: 告诉 JSON 序列化时这个字段叫什么名字

**为什么要用不同命名？**
- Go 语言习惯用驼峰命名：`UserID`
- 数据库和 JSON 习惯用下划线：`user_id`

**标签说明：**
- `_id,omitempty`: MongoDB 的主键，omitempty 表示如果为空则不序列化
- `user_id`: 在数据库中的字段名

## 🚀 使用流程

### 1. 初始化连接 (`init.go`)

```go
// 从环境变量初始化
func InitMongoDBFromEnv() error {
    config := DefaultMongoDBConfig()

    // 读取环境变量
    if uri := os.Getenv("MONGODB_URI"); uri != "" {
        config.URI = uri
    }

    if database := os.Getenv("MONGODB_DATABASE"); database != "" {
        config.Database = database
    }

    return GetMongoInstance().WithConfig(config).Connect()
}

// 自动初始化
func AutoInitMongoDB() {
    err := InitMongoDBFromEnv()
    if err != nil {
        // 失败时使用默认配置
        err = GetMongoInstance().Connect()
    }
}
```

### 2. 获取数据库实例

```go
// 获取单例实例
dbInstance := database.GetMongoInstance()

// 检查连接状态
if !dbInstance.IsConnected() {
    log.Println("数据库未连接")
}
```

### 3. 获取集合（类似表格）

```go
// 获取用户积分集合
userCollection := dbInstance.GetCollection("user_credits")

// 获取小说集合
novelCollection := dbInstance.GetCollection("novels")

// 获取购买记录集合
purchaseCollection := dbInstance.GetCollection("user_novel_purchases")

// 获取活动日志集合
activityCollection := dbInstance.GetCollection("user_activities")
```

### 4. 基本操作示例

```go
// 插入用户积分记录
userCredit := UserCredit{
    UserID:        "user123",
    Credit:        100,
    TotalUsed:     0,
    TotalRecharge: 100,
    IsActive:      true,
    CreatedAt:     time.Now(),
    UpdatedAt:     time.Now(),
}

result, err := userCollection.InsertOne(context.Background(), userCredit)

// 查询小说
var novel Novel
err = novelCollection.FindOne(context.Background(), bson.M{
    "title": "三体",
}).Decode(&novel)

// 更新积分
update := bson.M{
    "$inc": bson.M{"credit": -20},        // 积分减20
    "$set": bson.M{"updated_at": time.Now()},
}
result, err := userCollection.UpdateOne(
    context.Background(),
    bson.M{"user_id": "user123"},
    update,
)
```

## 🛠️ 高级功能

### 1. 连接池管理

```go
// 获取连接统计信息
stats := dbInstance.GetStats()
fmt.Printf("连接状态: %v\n", stats["connected"])
fmt.Printf("最大连接数: %v\n", stats["max_pool_size"])
fmt.Printf("数据库名: %v\n", stats["database"])
```

### 2. 安全断开连接

```go
// 程序退出时安全断开
defer func() {
    if err := dbInstance.Close(); err != nil {
        log.Printf("断开数据库连接失败: %v", err)
    }
}()
```

### 3. 环境变量配置

可以通过设置环境变量来配置数据库：

```bash
# Linux/Mac
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DATABASE="novel"
export MONGODB_TIMEOUT="30s"
export MONGODB_MAX_POOL_SIZE="20"

# Windows
set MONGODB_URI=mongodb://localhost:27017
set MONGODB_DATABASE=novel
```

## 💡 实际应用场景

### 用户购买小说流程

1. **检查用户积分** (`UserCredit`)
   ```go
   userCredit := getUserCredit("user123")
   if userCredit.Credit < novel.Price {
       return "积分不足"
   }
   ```

2. **创建购买记录** (`UserNovelPurchase`)
   ```go
   purchase := UserNovelPurchase{
       UserID:  "user123",
       NovelID: "novel456",
       Price:   novel.Price,
       Status:  "pending",
       PaidAt:  time.Now(),
   }
   ```

3. **扣除积分** (`UserCredit`)
   ```go
   updateCredit("user123", -novel.Price)
   ```

4. **记录活动日志** (`UserActivity`)
   ```go
   activity := UserActivity{
       UserID:     "user123",
       Action:     "purchase",
       TargetID:   "novel456",
       TargetType: "novel",
       Metadata: map[string]interface{}{
           "price": novel.Price,
           "title": novel.Title,
       },
       CreatedAt: time.Now(),
   }
   ```

## 🎯 设计优势

### 1. **模块化设计**
- 连接管理与数据模型分离
- 每个文件职责单一明确

### 2. **配置灵活**
- 支持环境变量配置
- 提供合理默认值

### 3. **线程安全**
- 使用读写锁保护并发访问
- 单例模式避免重复连接

### 4. **错误处理完善**
- 连接失败自动重试
- 提供详细的错误信息

### 5. **易于扩展**
- 模型定义清晰
- 方便添加新的数据类型

## 📝 小白总结

这个数据库模块实现了：

1. **连接管理**：安全地连接到 MongoDB 数据库
2. **数据模型**：定义用户积分、小说、购买记录、活动日志等数据结构
3. **单例模式**：确保只有一个数据库连接，提高效率
4. **配置灵活性**：可以通过环境变量配置数据库连接

**业务价值：**
- 用户可以用积分购买小说
- 系统记录用户的购买历史
- 追踪用户的各种行为（登录、阅读、购买等）
- 管理小说的内容和状态
- 提供完整的用户行为分析数据

这是一个典型的小说资源管理系统的数据层设计，支持完整的业务流程！