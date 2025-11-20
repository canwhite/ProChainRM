package database

import (
	"log"
	"os"
	"strconv"
	"time"
)

// InitMongoDBFromEnv 从环境变量初始化MongoDB连接
func InitMongoDBFromEnv() error {
	config := DefaultMongoDBConfig()

	// 从环境变量读取配置
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

	if idleTTL := os.Getenv("MONGODB_MAX_CONN_IDLE_TTL"); idleTTL != "" {
		if duration, err := time.ParseDuration(idleTTL); err == nil {
			config.MaxConnIdleTTL = duration
		}
	}

	return GetMongoInstance().WithConfig(config).Connect()
}

// AutoInitMongoDB 自动初始化MongoDB（带默认配置）
func AutoInitMongoDB() {
	err := InitMongoDBFromEnv()
	if err != nil {
		log.Printf("MongoDB初始化失败，使用默认配置: %v", err)
		// 使用默认配置重试
		err = GetMongoInstance().Connect()
		if err != nil {
			log.Fatalf("MongoDB连接失败: %v", err)
		}
	} else {
		log.Println("MongoDB初始化成功")
	}
}