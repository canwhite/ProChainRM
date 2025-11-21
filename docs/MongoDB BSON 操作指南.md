# MongoDB BSON 操作指南

## 📖 代码解析

### 基础示例：用户积分消费

```go
update := bson.M{
    "$inc": bson.M{
        "credit":     -1,        // 直接使用int
        "total_used": 1,         // 直接使用int
    },
    "$set": bson.M{
        "updated_at": time.Now(),
    },
}
```

## 🔍 逐层分析

### 1. `bson.M` 是什么？

```go
bson.M  // 等价于 map[string]interface{}
```

```go
// bson.M 的定义
type M map[string]interface{}
```

所以上面的代码等价于：

```go
update := map[string]interface{}{
    "$inc": map[string]interface{}{
        "credit":     -1,
        "total_used": 1,
    },
    "$set": map[string]interface{}{
        "updated_at": time.Now(),
    },
}
```

### 2. MongoDB 操作符

#### `$inc` - 增加或减少数值
```go
"$inc": bson.M{
    "credit":     -1,        // credit 字段减1
    "total_used": 1,         // total_used 字段加1
}
```

**效果：**
- 如果原来是 `credit: 100`，操作后变成 `credit: 99`
- 如果原来是 `total_used: 50`，操作后变成 `total_used: 51`

#### `$set` - 设置字段值
```go
"$set": bson.M{
    "updated_at": time.Now(),  // 设置更新时间
}
```

**效果：**
- 将 `updated_at` 字段设置为当前时间
- 如果字段不存在，会自动创建

## 🎯 完整的业务含义

这段代码模拟的是**用户消费积分**的场景：

```go
// 原始数据可能是：
{
    "user_id": "test_user_001",
    "credit": 100,           // 当前积分
    "total_used": 50,        // 已使用积分
    "total_recharge": 100,   // 总充值积分
    "updated_at": "2024-01-15 10:00:00"
}

// 执行更新后变成：
{
    "user_id": "test_user_001",
    "credit": 99,            // 100 - 1 = 99
    "total_used": 51,        // 50 + 1 = 51
    "total_recharge": 100,   // 不变
    "updated_at": "2024-01-15 14:30:00"  // 更新为当前时间
}
```

## 💡 为什么用这种结构？

### 1. **原子操作**
```go
// ✅ 原子操作，不会被并发修改干扰
update := bson.M{
    "$inc": bson.M{"credit": -1},
}

// ❌ 非原子操作，可能有并发问题
user := get_user()
user.credit -= 1
save_user(user)  // 在这期间可能有其他修改
```

### 2. **MongoDB 原生支持**
```go
// MongoDB 的 update 操作就是这种格式
db.user_credits.updateOne(
    {"user_id": "test_user_001"},
    {
        "$inc": {"credit": -1},
        "$set": {"updated_at": new Date()}
    }
)
```

## 🔧 实际使用示例

### 1. 消费积分函数

```go
func ConsumeCredit(userID string, amount int) error {
    collection := database.GetMongoInstance().GetCollection("user_credits")

    filter := bson.M{"user_id": userID}
    update := bson.M{
        "$inc": bson.M{
            "credit":     -amount,
            "total_used": amount,
        },
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("更新积分失败: %v", err)
    }

    if result.MatchedCount == 0 {
        return errors.New("用户不存在")
    }

    return nil
}
```

### 2. 复杂的更新操作

```go
func UpdateUserActivity(userID string, action string) error {
    collection := database.GetMongoInstance().GetCollection("user_activities")

    activity := bson.M{
        "user_id":    userID,
        "action":     action,
        "created_at": time.Now(),
    }

    // 同时更新用户积分和添加活动记录
    update := bson.M{
        "$inc": bson.M{"credit": -1},
        "$set": bson.M{
            "last_activity": time.Now(),
            "updated_at":   time.Now(),
        },
        "$push": bson.M{  // $push 向数组添加元素
            "activities": bson.M{
                "$each": []bson.M{activity},
                "$position": 0,  // 添加到数组开头
                "$slice": 10,    // 最多保留10个活动
            },
        },
    }

    return collection.UpdateOne(context.Background(),
        bson.M{"user_id": userID}, update)
}
```

## 📊 常用的MongoDB操作符

### 数值操作

#### `$inc` - 增减数值
```go
// 增加积分
update := bson.M{
    "$inc": bson.M{
        "credit": 10,       // 加10分
        "login_count": 1,   // 登录次数加1
    },
}

// 减少积分
update := bson.M{
    "$inc": bson.M{
        "credit": -5,       // 减5分
        "lives": -1,        // 生命减1
    },
}

// 负数操作
update := bson.M{
    "$inc": bson.M{
        "health": -20,      // 生命值减20
        "mana": -50,        // 魔法值减50
    },
}
```

#### `$mul` - 乘法操作
```go
// 积分翻倍
update := bson.M{
    "$mul": bson.M{
        "credit": 2,        // 积分乘以2
        "bonus": 1.5,       // 奖励乘以1.5
    },
}
```

#### `$min` 和 `$max` - 设置最小/最大值
```go
// 设置最小值
update := bson.M{
    "$min": bson.M{
        "health": 0,        // 确保生命值不低于0
        "level": 1,         // 确保等级不低于1
    },
}

// 设置最大值
update := bson.M{
    "$max": bson.M{
        "health": 100,      // 生命值不超过100
        "experience": 9999, // 经验值上限
    },
}
```

### 字段操作

#### `$set` - 设置字段值
```go
// 简单设置
update := bson.M{
    "$set": bson.M{
        "name": "新用户名",
        "status": "active",
        "avatar": "new_avatar.jpg",
    },
}

// 嵌套对象设置
update := bson.M{
    "$set": bson.M{
        "profile.bio": "这是我的个人简介",
        "settings.theme": "dark",
        "config.language": "zh-CN",
    },
}

// 数组元素设置
update := bson.M{
    "$set": bson.M{
        "scores.0": 100,    // 设置数组的第一个元素
        "tags.$": "热门",    // 设置匹配的数组元素
    },
}
```

#### `$unset` - 删除字段
```go
// 删除单个字段
update := bson.M{
    "$unset": bson.M{
        "old_field": 1,      // 删除字段
        "temp_data": "",     // 删除字段
    },
}

// 删除多个字段
update := bson.M{
    "$unset": bson.M{
        "deleted_field": 1,
        "obsolete_data": 1,
        "tmp_cache": 1,
    },
}
```

#### `$rename` - 重命名字段
```go
// 重命名字段
update := bson.M{
    "$rename": bson.M{
        "old_name": "new_name",
        "user_name": "username",
        "login_time": "last_login",
    },
}
```

### 数组操作

#### `$push` - 添加数组元素
```go
// 简单添加
update := bson.M{
    "$push": bson.M{
        "tags": "新标签",
        "friends": "新朋友",
    },
}

// 添加多个元素
update := bson.M{
    "$push": bson.M{
        "tags": bson.M{
            "$each": []string{"标签1", "标签2", "标签3"},
        },
    },
}

// 添加到指定位置并限制数量
update := bson.M{
    "$push": bson.M{
        "recent_activities": bson.M{
            "$each": []bson.M{activity1, activity2},
            "$position": 0,      // 添加到数组开头
            "$slice": 10,        // 最多保留10个元素
        },
    },
}
```

#### `$pull` - 删除数组元素
```go
// 删除匹配的元素
update := bson.M{
    "$pull": bson.M{
        "tags": "要删除的标签",
        "blocked_users": "要解除拉黑的用户",
    },
}

// 删除满足条件的元素
update := bson.M{
    "$pull": bson.M{
        "orders": bson.M{
            "status": "cancelled",  // 删除所有状态为cancelled的订单
        },
    },
}
```

#### `$addToSet` - 去重添加
```go
// 添加到集合（自动去重）
update := bson.M{
    "$addToSet": bson.M{
        "tags": "新标签",       // 如果已存在则不添加
        "friends": "新朋友",    // 如果已是好友则不添加
    },
}

// 添加多个元素到集合
update := bson.M{
    "$addToSet": bson.M{
        "$each": []string{"标签1", "标签2", "标签3"},
    },
}
```

#### `$pop` - 移除数组首尾元素
```go
// 移除最后一个元素
update := bson.M{
    "$pop": bson.M{
        "recent_activities": 1,   // 1表示移除最后一个，-1表示移除第一个
    },
}

// 移除第一个元素
update := bson.M{
    "$pop": bson.M{
        "queue": -1,              // 移除队列的第一个元素
    },
}
```

### 条件更新

#### `$setOnInsert` - 仅在插入时设置
```go
// 只在文档不存在时设置字段
update := bson.M{
    "$setOnInsert": bson.M{
        "created_at": time.Now(),
        "initial_level": 1,
        "welcome_bonus": 100,
    },
    "$inc": bson.M{
        "login_count": 1,
    },
}

// 使用 upsert 选项
collection.UpdateOne(
    context.Background(),
    bson.M{"user_id": userID},
    update,
    options.Update().SetUpsert(true),  // 如果不存在则插入
)
```

#### `$currentDate` - 设置当前时间
```go
// 设置当前时间
update := bson.M{
    "$currentDate": bson.M{
        "last_modified": true,           // 设置为当前时间
        "last_login": bson.M{"$type": "timestamp"},  // 设置为时间戳
    },
}
```

### 数组元素操作

#### `$` - 更新匹配的数组元素
```go
// 更新查询条件匹配的数组元素
filter := bson.M{
    "user_id": userID,
    "orders.product_id": productID,  // 查找特定的订单项
}

update := bson.M{
    "$set": bson.M{
        "orders.$.status": "shipped",     // 更新匹配的订单状态
        "orders.$.ship_date": time.Now(),
    },
}
```

#### `$[]` - 更新所有数组元素
```go
// 更新所有数组元素
update := bson.M{
    "$set": bson.M{
        "scores.$[]": bson.M{
            "verified": true,      // 给所有分数添加验证标记
            "updated_at": time.Now(),
        },
    },
}
```

## 🚀 组合操作示例

### 1. 用户登录更新
```go
func UpdateUserLogin(userID string, clientInfo map[string]interface{}) error {
    collection := database.GetMongoInstance().GetCollection("users")

    filter := bson.M{"user_id": userID}

    update := bson.M{
        "$inc": bson.M{
            "login_count": 1,
            "total_login_days": 1,
        },
        "$set": bson.M{
            "last_login": time.Now(),
            "last_login_ip": clientInfo["ip"],
            "last_login_device": clientInfo["device"],
            "status": "online",
        },
        "$setOnInsert": bson.M{
            "created_at": time.Now(),
            "initial_login_device": clientInfo["device"],
        },
        "$currentDate": bson.M{
            "last_seen": true,
        },
    }

    opts := options.Update().SetUpsert(true)
    _, err := collection.UpdateOne(context.Background(), filter, update, opts)
    return err
}
```

### 2. 小说阅读更新
```go
func UpdateNovelReading(userID string, novelID string, chapter int) error {
    collection := database.GetMongoInstance().GetCollection("reading_history")

    filter := bson.M{"user_id": userID, "novel_id": novelID}

    update := bson.M{
        "$inc": bson.M{
            "read_chapters": 1,
            "reading_time": time.Minute,  // 假设每章阅读1分钟
        },
        "$set": bson.M{
            "current_chapter": chapter,
            "last_read_time": time.Now(),
        },
        "$setOnInsert": bson.M{
            "started_reading": time.Now(),
        },
        "$push": bson.M{
            "chapter_history": bson.M{
                "$each": []bson.M{{
                    "chapter": chapter,
                    "read_time": time.Now(),
                }},
                "$position": 0,
                "$slice": 50,  // 保留最近50章的阅读记录
            },
        },
    }

    opts := options.Update().SetUpsert(true)
    _, err := collection.UpdateOne(context.Background(), filter, update, opts)
    return err
}
```

### 3. 库存管理更新
```go
func UpdateInventory(productID string, quantity int, operation string) error {
    collection := database.GetMongoInstance().GetCollection("inventory")

    filter := bson.M{"product_id": productID}

    var update bson.M

    switch operation {
    case "purchase":
        update = bson.M{
            "$inc": bson.M{
                "stock": -quantity,
                "sold_quantity": quantity,
                "total_revenue": quantity * 29.9,  // 假设单价29.9
            },
            "$set": bson.M{
                "last_purchase_time": time.Now(),
            },
            "$max": bson.M{
                "peak_daily_sales": bson.M{"$add": []interface{}{"$peak_daily_sales", quantity}},
            },
        }

    case "restock":
        update = bson.M{
            "$inc": bson.M{
                "stock": quantity,
                "restock_count": 1,
                "total_restocked": quantity,
            },
            "$set": bson.M{
                "last_restock_time": time.Now(),
            },
        }

    case "adjust":
        update = bson.M{
            "$inc": bson.M{
                "stock": quantity,
                "adjustment_count": 1,
            },
            "$set": bson.M{
                "last_adjustment_time": time.Now(),
                "adjustment_reason": "inventory_check",
            },
        }
    }

    result, err := collection.UpdateOne(context.Background(), filter, update)
    if result.MatchedCount == 0 {
        return errors.New("商品不存在")
    }
    return err
}
```

## 🎯 最佳实践

### 1. 性能考虑
```go
// ✅ 使用索引字段进行查询
filter := bson.M{"user_id": userID}  // 确保 user_id 有索引

// ✅ 批量更新
collection.UpdateMany(
    context.Background(),
    bson.M{"status": "pending"},
    bson.M{"$set": bson.M{"status": "processed"}},
)

// ✅ 使用事务处理复杂更新
session, err := client.StartSession()
if err != nil {
    return err
}
defer session.EndSession(context.Background())

callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
    // 在事务中执行多个更新操作
    if err := updateUserCredits(sessCtx, userID, -10); err != nil {
        return nil, err
    }
    if err := addTransactionRecord(sessCtx, userID, -10); err != nil {
        return nil, err
    }
    return nil, nil
}

_, err = session.WithTransaction(context.Background(), callback)
```

### 2. 错误处理
```go
func SafeUpdate(userID string, update bson.M) error {
    collection := database.GetMongoInstance().GetCollection("users")

    result, err := collection.UpdateOne(context.Background(),
        bson.M{"user_id": userID}, update)

    if err != nil {
        return fmt.Errorf("数据库更新失败: %v", err)
    }

    if result.MatchedCount == 0 {
        return errors.New("用户不存在")
    }

    if result.ModifiedCount == 0 {
        return errors.New("数据没有实际变化")
    }

    return nil
}
```

### 3. 数据一致性
```go
// ✅ 使用乐观锁
version := getUserVersion(userID)  // 获取当前版本

filter := bson.M{
    "user_id": userID,
    "version": version,  // 确保版本号匹配
}

update := bson.M{
    "$inc": bson.M{
        "credit": -10,
        "version": 1,  // 版本号递增
    },
}

result, err := collection.UpdateOne(context.Background(), filter, update)
if result.MatchedCount == 0 {
    return errors.New("数据已被其他操作修改，请重试")
}
```

## 📝 总结

### 核心操作符记忆
- **`$inc`** - 数值增减（积分、计数器）
- **`$set`** - 设置字段值（状态、时间）
- **`$push`** - 添加数组元素（历史记录、标签）
- **`$pull`** - 删除数组元素（移除记录）
- **`$unset`** - 删除字段（清理数据）
- **`$setOnInsert`** - 仅插入时设置（初始化字段）

### 设计原则
1. **原子性优先** - 尽可能用MongoDB原子操作
2. **索引利用** - 查询条件使用索引字段
3. **事务处理** - 复杂操作使用事务
4. **错误处理** - 检查操作结果和错误
5. **性能优化** - 批量操作和避免全表扫描

通过掌握这些BSON操作符和最佳实践，你可以高效地进行MongoDB数据操作！