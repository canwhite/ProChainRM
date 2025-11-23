# ProChainRM 项目 dotenv 配置

## 1. 安装

```bash
go get github.com/joho/godotenv
```

## 2. 创建 .env 文件

在项目根目录创建 `.env` 文件：

```env
# MongoDB 配置
MONGODB_URI=mongodb://myuser:mypassword@localhost:27017
MONGODB_DATABASE=novel

# 数据库连接池配置
MONGODB_TIMEOUT=30s
MONGODB_MAX_POOL_SIZE=10
MONGODB_MIN_POOL_SIZE=2

# 服务器配置
SERVER_PORT=8080
```

## 3. 修改 init.go

在 `novel-resource-management/database/init.go` 中添加 dotenv 支持：

```go
package database

import (
    "log"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

// LoadEnv 加载环境变量
func LoadEnv() {
    if err := godotenv.Load(); err != nil {
        log.Printf("未找到 .env 文件，使用系统环境变量")
    }
}

// InitMongoDBFromEnv 从环境变量初始化MongoDB连接
func InitMongoDBFromEnv() error {
    // 先加载 .env 文件
    LoadEnv()

    config := DefaultMongoDBConfig()

    // 直接使用 MONGODB_URI
    if uri := os.Getenv("MONGODB_URI"); uri != "" {
        config.URI = uri
    }

    if database := os.Getenv("MONGODB_DATABASE"); database != "" {
        config.Database = database
    }

    if timeout := os.Getenv("MONGODB_TIMEOUT"); timeout != "" {
        if duration, err := time.ParseDuration(timeout); err == nil {
            config.Timeout = duration
        }
    }

    if maxPool := os.Getenv("MONGODB_MAX_POOL_SIZE"); maxPool != "" {
        if size, err := strconv.ParseUint(maxPool, 10, 64); err == nil {
            config.MaxPoolSize = size
        }
    }

    if minPool := os.Getenv("MONGODB_MIN_POOL_SIZE"); minPool != "" {
        if size, err := strconv.ParseUint(minPool, 10, 64); err == nil {
            config.MinPoolSize = size
        }
    }

    return GetMongoInstance().WithConfig(config).Connect()
}

// AutoInitMongoDB 自动初始化MongoDB
func AutoInitMongoDB() {
    err := InitMongoDBFromEnv()
    if err != nil {
        log.Printf("MongoDB初始化失败: %v", err)
    } else {
        log.Println("MongoDB初始化成功")
    }
}
```

## 4. 修改 main.go

```go
package main

import (
    "log"
    "your-project/novel-resource-management/database"
)

func main() {
    // 初始化数据库（会自动加载 .env 文件）
    database.AutoInitMongoDB()

    // 检查连接状态
    dbInstance := database.GetMongoInstance()
    if dbInstance.IsConnected() {
        log.Println("数据库连接成功")
    } else {
        log.Fatal("数据库连接失败")
    }

    // 你的业务逻辑...
}
```

## 5. 创建 .env.example

```env
# MongoDB 配置示例
MONGODB_URI=mongodb://username:password@localhost:27017
MONGODB_DATABASE=novel

# 连接池配置
MONGODB_TIMEOUT=30s
MONGODB_MAX_POOL_SIZE=10
MONGODB_MIN_POOL_SIZE=2

# 服务器配置
SERVER_PORT=8080
```

## 6. 添加到 .gitignore

```gitignore
.env
```

## 7. 完成！

现在你的项目支持：
- 从 `.env` 文件读取配置
- 从环境变量读取配置
- 灵活的 MongoDB 连接配置
- 连接池配置

运行项目时，会自动加载 `.env` 文件中的配置。