# GORM 配合 MySQL 使用指南

## 1. 安装依赖

### 安装 GORM 和 MySQL 驱动

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

## 2. 连接 MySQL 数据库

### 基本连接方式

```go
package main

import (
    "fmt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // 数据库连接字符串格式
    dsn := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

    // 连接数据库
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    fmt.Println("数据库连接成功!")
}
```

### 带连接池配置的连接

```go
package main

import (
    "time"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func connectDB() (*gorm.DB, error) {
    dsn := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // 获取底层的 *sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 设置连接池
    sqlDB.SetMaxIdleConns(10)           // 空闲连接池中的最大连接数
    sqlDB.SetMaxOpenConns(100)          // 数据库的最大打开连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接可重用的最长时间

    return db, nil
}
```

## 3. 模型定义

### 基本模型结构

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
    Age      int    `gorm:"default:18" json:"age"`
    Active   bool   `gorm:"default:true" json:"active"`
}

// Novel 小说模型
type Novel struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Title       string `gorm:"not null;size:200" json:"title"`
    Author      string `gorm:"size:100" json:"author"`
    Description string `gorm:"type:text" json:"description"`
    UserID      uint   `gorm:"not null" json:"user_id"`  // 外键

    // 关联关系
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

### 常用的 GORM 标签

```go
type Example struct {
    ID      uint   `gorm:"primaryKey"`                    // 主键
    Name    string `gorm:"size:100;not null"`           // 长度限制，非空
    Email   string `gorm:"uniqueIndex"`                 // 唯一索引
    Code    string `gorm:"index:idx_code"`              // 普通索引，指定索引名

    CreatedAt time.Time `gorm:"autoCreateTime"`         // 自动设置创建时间
    UpdatedAt time.Time `gorm:"autoUpdateTime"`         // 自动更新时间

    Status string `gorm:"default:'active'"`            // 默认值
    Count  int    `gorm:"default:0;comment:'计数'"`     // 默认值和注释

    // 外键关联
    UserID uint `gorm:"foreignKey:UserID;references:ID"`
}
```

## 4. 数据库操作

### 自动迁移

```go
// 自动迁移数据库表
err := db.AutoMigrate(&User{}, &Novel{})
if err != nil {
    panic("failed to migrate database")
}
```

### CRUD 操作

#### 创建数据

```go
// 创建单个用户
user := User{
    Username: "zhangsan",
    Email:    "zhangsan@example.com",
    Age:      25,
}
result := db.Create(&user)
fmt.Println("创建的用户ID:", user.ID)

// 批量创建
users := []User{
    {Username: "lisi", Email: "lisi@example.com"},
    {Username: "wangwu", Email: "wangwu@example.com"},
}
result := db.Create(&users)
fmt.Println("影响的行数:", result.RowsAffected)
```

#### 查询数据

```go
// 查询所有用户
var users []User
result := db.Find(&users)

// 按条件查询
var user User
result := db.Where("username = ?", "zhangsan").First(&user)

// 多条件查询
db.Where("age > ? AND active = ?", 18, true).Find(&users)

// 使用 struct 条件查询
db.Where(&User{Username: "zhangsan", Active: true}).First(&user)

// 使用 map 条件查询
db.Where(map[string]interface{}{
    "age >=": 18,
    "active":  true,
}).Find(&users)

// 分页查询
offset := 0
limit := 10
db.Offset(offset).Limit(limit).Find(&users)

// 排序
db.Order("age desc, created_at asc").Find(&users)
```

#### 更新数据

```go
// 更新整个结构体
db.Model(&user).Updates(User{
    Age:    26,
    Active: false,
})

// 更新指定字段
db.Model(&user).Update("age", 27)

// 使用 map 更新多个字段
db.Model(&user).Updates(map[string]interface{}{
    "age":    28,
    "active": true,
})

// 更新符合条件的所有记录
db.Model(&User{}).Where("age < ?", 18).Update("active", false)
```

#### 删除数据

```go
// 软删除（如果模型有 DeletedAt 字段）
db.Delete(&user, 1) // 根据 ID 删除

// 硬删除
db.Unscoped().Delete(&user, 1)

// 批量删除
db.Where("age > ?", 100).Delete(&User{})
```

## 5. 关联关系

### 一对多关系

```go
type User struct {
    ID       uint    `gorm:"primaryKey"`
    Username string
    Novels   []Novel `gorm:"foreignKey:UserID"` // 用户有多个小说
}

type Novel struct {
    ID     uint `gorm:"primaryKey"`
    Title  string
    UserID uint // 外键
    User   User `gorm:"foreignKey:UserID"` // 小说属于一个用户
}

// 预加载关联数据
var user User
db.Preload("Novels").First(&user, 1)

// 查询用户及其小说
var users []User
db.Preload("Novels", "active = ?", true).Find(&users)
```

### 多对多关系

```go
type User struct {
    ID       uint      `gorm:"primaryKey"`
    Username string
    Roles    []Role    `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 创建多对多关联
user := User{Username: "admin"}
role1 := Role{Name: "admin"}
role2 := Role{Name: "user"}

db.Create(&user)
db.Create(&role1)
db.Create(&role2)

// 添加关联
db.Model(&user).Association("Roles").Append(&role1, &role2)

// 查询带关联的数据
var users []User
db.Preload("Roles").Find(&users)
```

## 6. 高级查询

### 原生 SQL

```go
// 执行原生查询
var results []map[string]interface{}
db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&results)

// 执行原生更新
db.Exec("UPDATE users SET active = ? WHERE age < ?", false, 18)
```

### 事务处理

```go
// 开始事务
tx := db.Begin()

// 执行操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&novel).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交事务
tx.Commit()
```

### 钩子函数

```go
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string
    Email    string
}

// 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 数据验证或处理
    if u.Username == "" {
        return errors.New("username cannot be empty")
    }
    return nil
}

// 更新前的钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 数据处理逻辑
    return nil
}
```

## 7. 配置优化

### 数据库配置

```go
package config

import (
    "time"
    "gorm.io/gorm"
)

type DatabaseConfig struct {
    Host            string
    Port            string
    User            string
    Password        string
    DBName          string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime time.Duration
}

func GetDatabaseConfig() *DatabaseConfig {
    return &DatabaseConfig{
        Host:            "127.0.0.1",
        Port:            "3306",
        User:            "root",
        Password:        "password",
        DBName:          "test_db",
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    }
}

func ConnectDB(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User,
        config.Password,
        config.Host,
        config.Port,
        config.DBName,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // 日志级别
    })
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

    return db, nil
}
```

## 8. 常见问题解决

### 1. 时区问题

```go
// 在连接字符串中指定时区
dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

// 或者在模型中设置时区
type Model struct {
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

### 2. 连接超时

```go
// 在连接字符串中添加超时参数
dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s"
```

### 3. 字符编码

```go
// 使用 utf8mb4 编码支持完整 Unicode
dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&collation=utf8mb4_unicode_ci"
```

## 9. 性能优化建议

### 1. 索引优化

```go
// 为常用查询字段添加索引
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string `gorm:"index:idx_username"` // 单字段索引
    Email    string `gorm:"uniqueIndex"`        // 唯一索引
    Age      int
    Status   string
}

// 复合索引
type Novel struct {
    ID     uint `gorm:"primaryKey"`
    Title  string
    UserID uint `gorm:"index:idx_user_status"` // 复合索引的一部分
    Status string `gorm:"index:idx_user_status"`
}
```

### 2. 预加载优化

```go
// 避免N+1查询问题
var users []User
db.Preload("Novels").Find(&users) // 一次查询获取所有关联数据

// 或者使用Joins
db.Select("users.*, novels.title").
    Joins("LEFT JOIN novels ON novels.user_id = users.id").
    Find(&users)
```

### 3. 批量操作

```go
// 批量插入
var users []User
// ... 填充数据
db.CreateInBatches(users, 100) // 每100条一批

// 批量更新
db.Model(&User{}).Where("active = ?", false).Update("status", "inactive")
```

## 10. 完整示例

```go
package main

import (
    "fmt"
    "time"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
    Age      int    `gorm:"default:18" json:"age"`
    Active   bool   `gorm:"default:true" json:"active"`

    Novels []Novel `gorm:"foreignKey:UserID" json:"novels,omitempty"`
}

type Novel struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Title       string `gorm:"not null;size:200" json:"title"`
    Author      string `gorm:"size:100" json:"author"`
    Description string `gorm:"type:text" json:"description"`
    UserID      uint   `gorm:"not null" json:"user_id"`

    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func main() {
    // 连接数据库
    dsn := "root:password@tcp(127.0.0.1:3306)/novel_db?charset=utf8mb4&parseTime=True&loc=Local"

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移
    if err := db.AutoMigrate(&User{}, &Novel{}); err != nil {
        panic("failed to migrate database")
    }

    // 创建示例数据
    user := User{
        Username: "example_user",
        Email:    "user@example.com",
        Age:      25,
        Active:   true,
        Novels: []Novel{
            {Title: "第一个小说", Author: "作者名", Description: "小说描述"},
            {Title: "第二个小说", Author: "另一个作者", Description: "另一个描述"},
        },
    }

    // 创建数据
    if err := db.Create(&user).Error; err != nil {
        panic("failed to create user")
    }

    // 查询数据
    var users []User
    db.Preload("Novels").Find(&users)

    for _, user := range users {
        fmt.Printf("用户: %s, 小说数量: %d\n", user.Username, len(user.Novels))
        for _, novel := range user.Novels {
            fmt.Printf("  - %s\n", novel.Title)
        }
    }

    fmt.Println("GORM + MySQL 示例运行完成!")
}
```

这个指南涵盖了 GORM 配合 MySQL 使用的主要方面，从基本连接到高级操作，帮助你快速上手和深入使用。