# Go项目中集成MongoDB完整指南

## 1. 安装MongoDB驱动

```bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
go get go.mongodb.org/mongo-driver/bson/primitive
go get go.mongodb.org/mongo-driver/bson
```

## 2. 基本连接配置

### 基础连接示例
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() (*mongo.Client, error) {
    // 设置客户端选项
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    clientOptions.SetMaxPoolSize(10)  // 连接池大小
    clientOptions.SetMaxConnIdleTime(10 * time.Minute)  // 连接空闲时间

    // 连接到MongoDB
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return nil, fmt.Errorf("连接MongoDB失败: %v", err)
    }

    // 检查连接
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("无法连接到MongoDB: %v", err)
    }

    fmt.Println("MongoDB连接成功!")
    return client, nil
}

// 优雅关闭连接
func CloseMongoDB(client *mongo.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return client.Disconnect(ctx)
}
```

### 连接字符串配置
```go
// 本地连接
uri := "mongodb://localhost:27017"

// 带认证的连接
uri := "mongodb://username:password@localhost:27017"

// 连接到副本集
uri := "mongodb://host1:27017,host2:27017,host3:27017/?replicaSet=rs0"

// 带选项的连接
uri := "mongodb://localhost:27017/?maxPoolSize=20&w=majority"
```

## 3. 定义数据模型

```go
import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// 用户积分模型
type UserCredit struct {
    ID            primitive.ObjectID `bson:"_id,omitempty"`
    UserID        string             `bson:"user_id" json:"user_id"`
    Credit        int                `bson:"credit" json:"credit"`
    TotalUsed     int                `bson:"total_used" json:"total_used"`
    TotalRecharge int                `bson:"total_recharge" json:"total_recharge"`
    CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// 小说模型
type Novel struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Title       string             `bson:"title" json:"title"`
    Author      string             `bson:"author" json:"author"`
    Content     string             `bson:"content" json:"content"`
    Category    string             `bson:"category" json:"category"`
    Tags        []string           `bson:"tags" json:"tags"`
    PublishedAt time.Time          `bson:"published_at" json:"published_at"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
```

### BSON标签说明
- `_id`: MongoDB的主键，`omitempty`表示创建时为空则自动生成
- `user_id`: 字段名在MongoDB中的存储名称
- `json`: 序列化为JSON时的字段名

## 4. 基本CRUD操作

### 创建文档 (Create)
```go
func (us *UserCreditService) CreateUserCredit(userId string, credit, totalUsed, totalRecharge int) error {
    userCredit := UserCredit{
        UserID:        userId,
        Credit:        credit,
        TotalUsed:     totalUsed,
        TotalRecharge: totalRecharge,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    result, err := us.collection.InsertOne(context.Background(), userCredit)
    if err != nil {
        return fmt.Errorf("插入用户积分失败: %v", err)
    }

    fmt.Printf("插入成功，ID: %v\n", result.InsertedID)
    return nil
}

// 批量插入
func (us *UserCreditService) BatchCreateUserCredits(credits []UserCredit) error {
    var docs []interface{}
    for _, credit := range credits {
        credit.CreatedAt = time.Now()
        credit.UpdatedAt = time.Now()
        docs = append(docs, credit)
    }

    result, err := us.collection.InsertMany(context.Background(), docs)
    if err != nil {
        return fmt.Errorf("批量插入失败: %v", err)
    }

    fmt.Printf("批量插入成功，插入了 %d 个文档\n", len(result.InsertedIDs))
    return nil
}
```

### 查询文档 (Read)
```go
// 单个查询
func (us *UserCreditService) ReadUserCredit(userId string) (*UserCredit, error) {
    var userCredit UserCredit
    filter := bson.M{"user_id": userId}

    err := us.collection.FindOne(context.Background(), filter).Decode(&userCredit)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("用户 %s 不存在", userId)
        }
        return nil, fmt.Errorf("查询用户积分失败: %v", err)
    }

    return &userCredit, nil
}

// 查询所有
func (us *UserCreditService) GetAllUserCredits() ([]UserCredit, error) {
    cursor, err := us.collection.Find(context.Background(), bson.M{})
    if err != nil {
        return nil, fmt.Errorf("查询所有用户积分失败: %v", err)
    }
    defer cursor.Close(context.Background())

    var userCredits []UserCredit
    err = cursor.All(context.Background(), &userCredits)
    if err != nil {
        return nil, fmt.Errorf("解析查询结果失败: %v", err)
    }

    return userCredits, nil
}

// 条件查询
func (us *UserCreditService) FindUsersByCreditRange(minCredit, maxCredit int) ([]UserCredit, error) {
    filter := bson.M{
        "credit": bson.M{
            "$gte": minCredit,
            "$lte": maxCredit,
        },
    }

    cursor, err := us.collection.Find(context.Background(), filter)
    if err != nil {
        return nil, fmt.Errorf("条件查询失败: %v", err)
    }
    defer cursor.Close(context.Background())

    var userCredits []UserCredit
    err = cursor.All(context.Background(), &userCredits)
    return userCredits, err
}

// 分页查询
func (us *UserCreditService) GetUserCreditsWithPagination(page, pageSize int) ([]UserCredit, int64, error) {
    // 计算跳过的文档数
    skip := (page - 1) * pageSize

    // 查询选项
    findOptions := options.Find()
    findOptions.SetSkip(int64(skip))
    findOptions.SetLimit(int64(pageSize))
    findOptions.SetSort(bson.M{"created_at": -1}) // 按创建时间倒序

    cursor, err := us.collection.Find(context.Background(), bson.M{}, findOptions)
    if err != nil {
        return nil, 0, fmt.Errorf("分页查询失败: %v", err)
    }
    defer cursor.Close(context.Background())

    var userCredits []UserCredit
    err = cursor.All(context.Background(), &userCredits)
    if err != nil {
        return nil, 0, fmt.Errorf("解析分页结果失败: %v", err)
    }

    // 获取总数
    total, err := us.collection.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        return nil, 0, fmt.Errorf("统计总数失败: %v", err)
    }

    return userCredits, total, nil
}
```

### 更新文档 (Update)
```go
// 更新单个文档
func (us *UserCreditService) UpdateUserCredit(userId string, credit, totalUsed, totalRecharge int) error {
    filter := bson.M{"user_id": userId}
    update := bson.M{
        "$set": bson.M{
            "credit":         credit,
            "total_used":     totalUsed,
            "total_recharge": totalRecharge,
            "updated_at":     time.Now(),
        },
    }

    result, err := us.collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("更新用户积分失败: %v", err)
    }

    if result.MatchedCount == 0 {
        return fmt.Errorf("用户 %s 不存在", userId)
    }

    return nil
}

// 原子更新（增加积分）
func (us *UserCreditService) AddCredit(userId string, amount int) error {
    filter := bson.M{"user_id": userId}
    update := bson.M{
        "$inc": bson.M{
            "credit": amount,
        },
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    result, err := us.collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("增加积分失败: %v", err)
    }

    if result.MatchedCount == 0 {
        return fmt.Errorf("用户 %s 不存在", userId)
    }

    return nil
}

// 消费token（原子操作）
func (us *UserCreditService) ConsumeUserToken(userId string) error {
    filter := bson.M{
        "user_id": userId,
        "credit": bson.M{"$gt": 0}, // 积分大于0
    }

    update := bson.M{
        "$inc": bson.M{
            "credit":     -1,
            "total_used": 1,
        },
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    result, err := us.collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("消费token失败: %v", err)
    }

    if result.MatchedCount == 0 {
        return fmt.Errorf("用户 %s 不存在或积分不足", userId)
    }

    return nil
}

// 批量更新
func (us *UserCreditService) BatchUpdateCredits(updates []UserCredit) error {
    var models []mongo.WriteModel

    for _, update := range updates {
        filter := bson.M{"user_id": update.UserID}
        updateDoc := bson.M{
            "$set": bson.M{
                "credit":         update.Credit,
                "total_used":     update.TotalUsed,
                "total_recharge": update.TotalRecharge,
                "updated_at":     time.Now(),
            },
        }

        model := mongo.UpdateOneModel{
            Filter: filter,
            Update: updateDoc,
        }
        models = append(models, model)
    }

    result, err := us.collection.BulkWrite(context.Background(), models)
    if err != nil {
        return fmt.Errorf("批量更新失败: %v", err)
    }

    fmt.Printf("批量更新成功，修改了 %d 个文档\n", result.ModifiedCount)
    return nil
}
```

### 删除文档 (Delete)
```go
// 删除单个文档
func (us *UserCreditService) DeleteUserCredit(userId string) error {
    filter := bson.M{"user_id": userId}

    result, err := us.collection.DeleteOne(context.Background(), filter)
    if err != nil {
        return fmt.Errorf("删除用户积分失败: %v", err)
    }

    if result.DeletedCount == 0 {
        return fmt.Errorf("用户 %s 不存在", userId)
    }

    return nil
}

// 批量删除
func (us *UserCreditService) DeleteUsersByCreditRange(maxCredit int) error {
    filter := bson.M{"credit": bson.M{"$lte": maxCredit}}

    result, err := us.collection.DeleteMany(context.Background(), filter)
    if err != nil {
        return fmt.Errorf("批量删除失败: %v", err)
    }

    fmt.Printf("批量删除成功，删除了 %d 个文档\n", result.DeletedCount)
    return nil
}
```

## 5. 完整的服务类示例

```go
package service

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type UserCreditService struct {
    collection *mongo.Collection
}

func NewUserCreditService(client *mongo.Client, dbName string) *UserCreditService {
    return &UserCreditService{
        collection: client.Database(dbName).Collection("user_credits"),
    }
}

// 创建索引以提高查询性能
func (us *UserCreditService) CreateIndexes() error {
    // 为user_id创建唯一索引
    userIdIndex := mongo.IndexModel{
        Keys: bson.M{"user_id": 1},
        Options: options.Index().SetUnique(true),
    }

    // 为credit创建普通索引
    creditIndex := mongo.IndexModel{
        Keys: bson.M{"credit": 1},
    }

    // 为复合查询创建复合索引
    compoundIndex := mongo.IndexModel{
        Keys: bson.M{"credit": 1, "created_at": -1},
    }

    indexes := []mongo.IndexModel{userIdIndex, creditIndex, compoundIndex}

    _, err := us.collection.Indexes().CreateMany(context.Background(), indexes)
    if err != nil {
        return fmt.Errorf("创建索引失败: %v", err)
    }

    fmt.Println("索引创建成功")
    return nil
}

// 使用事务确保数据一致性
func (us *UserCreditService) TransferCredit(fromUserId, toUserId string, amount int) error {
    // 创建会话
    session, err := us.collection.Database().Client().StartSession()
    if err != nil {
        return fmt.Errorf("创建事务会话失败: %v", err)
    }
    defer session.EndSession(context.Background())

    // 在事务中执行操作
    _, err = session.WithTransaction(context.Background(), func(sessionCtx mongo.SessionContext) (interface{}, error) {
        // 扣除发送方积分
        fromFilter := bson.M{"user_id": fromUserId, "credit": bson.M{"$gte": amount}}
        fromUpdate := bson.M{
            "$inc": bson.M{"credit": -amount},
            "$set": bson.M{"updated_at": time.Now()},
        }

        fromResult, err := us.collection.UpdateOne(sessionCtx, fromFilter, fromUpdate)
        if err != nil {
            return nil, fmt.Errorf("扣除发送方积分失败: %v", err)
        }

        if fromResult.MatchedCount == 0 {
            return nil, fmt.Errorf("发送方积分不足或不存在")
        }

        // 增加接收方积分
        toFilter := bson.M{"user_id": toUserId}
        toUpdate := bson.M{
            "$inc": bson.M{"credit": amount},
            "$set": bson.M{"updated_at": time.Now()},
            "$setOnInsert": bson.M{
                "total_used": 0,
                "total_recharge": amount,
                "created_at": time.Now(),
            },
        }

        // 使用upsert确保接收方存在
        toResult, err := us.collection.UpdateOne(sessionCtx, toFilter, toUpdate, options.Update().SetUpsert(true))
        if err != nil {
            return nil, fmt.Errorf("增加接收方积分失败: %v", err)
        }

        return map[string]interface{}{
            "from_result": fromResult,
            "to_result":   toResult,
        }, nil
    })

    return err
}
```

## 6. 在项目中使用

```go
package main

import (
    "log"
    "your-project/service"
)

func main() {
    // 初始化MongoDB连接
    client, err := InitMongoDB()
    if err != nil {
        log.Fatal("数据库连接失败:", err)
    }
    defer CloseMongoDB(client)

    // 创建服务
    userCreditService := service.NewUserCreditService(client, "prochain_rm")

    // 创建索引
    err = userCreditService.CreateIndexes()
    if err != nil {
        log.Printf("创建索引失败: %v", err)
    }

    // 使用服务进行操作
    err = userCreditService.CreateUserCredit("user123", 100, 0, 100)
    if err != nil {
        log.Printf("创建用户积分失败: %v", err)
    }

    // 消费token
    err = userCreditService.ConsumeUserToken("user123")
    if err != nil {
        log.Printf("消费token失败: %v", err)
    }

    // 查询用户积分
    userCredit, err := userCreditService.ReadUserCredit("user123")
    if err != nil {
        log.Printf("查询用户积分失败: %v", err)
    } else {
        log.Printf("用户积分: %+v", userCredit)
    }
}
```

## 7. 性能优化建议

### 连接池配置
```go
clientOptions := options.Client().ApplyURI(uri)
clientOptions.SetMaxPoolSize(100)              // 最大连接数
clientOptions.SetMinPoolSize(5)                // 最小连接数
clientOptions.SetMaxConnIdleTime(10 * time.Minute)  // 连接空闲时间
clientOptions.SetConnectTimeout(10 * time.Second)   // 连接超时
clientOptions.SetServerSelectionTimeout(5 * time.Second)  // 服务器选择超时
```

### 查询优化
```go
// 使用投影减少数据传输
projection := bson.M{
    "user_id": 1,
    "credit": 1,
    "_id": 0,  // 不返回_id
}

findOptions := options.Find()
findOptions.SetProjection(projection)

// 使用限制减少结果集
findOptions.SetLimit(100)

// 使用排序
findOptions.SetSort(bson.M{"credit": -1})
```

### 批量操作
```go
// 使用批量写操作提高性能
var models []mongo.WriteModel

// 添加多个更新操作
models = append(models, mongo.UpdateOneModel{
    Filter: bson.M{"user_id": "user1"},
    Update: bson.M{"$inc": bson.M{"credit": 10}},
})

models = append(models, mongo.UpdateOneModel{
    Filter: bson.M{"user_id": "user2"},
    Update: bson.M{"$inc": bson.M{"credit": 20}},
})

result, err := collection.BulkWrite(context.Background(), models)
```

## 8. 错误处理和日志

```go
import (
    "log"
    "go.mongodb.org/mongo-driver/mongo"
)

func HandleMongoError(err error) error {
    if err == nil {
        return nil
    }

    switch err {
    case mongo.ErrNoDocuments:
        return fmt.Errorf("文档不存在")
    case mongo.ErrClientDisconnected:
        return fmt.Errorf("客户端已断开连接")
    case context.DeadlineExceeded:
        return fmt.Errorf("操作超时")
    default:
        return fmt.Errorf("MongoDB操作失败: %v", err)
    }
}

// 使用示例
userCredit, err := userCreditService.ReadUserCredit("user123")
if err != nil {
    log.Printf("查询失败: %v", HandleMongoError(err))
    return
}
```

## 总结

使用MongoDB的优势：
- **无需JSON解析和类型转换**：直接操作Go结构体
- **灵活的文档结构**：适合存储复杂嵌套数据
- **强大的查询能力**：支持丰富的查询条件
- **原子操作**：支持事务和原子更新
- **高性能**：支持索引和聚合查询

这个指南涵盖了Go项目中集成MongoDB的完整流程，你可以根据具体需求进行调整和扩展。