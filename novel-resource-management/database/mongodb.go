package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

// MongoDBConfig MongoDB连接配置
type MongoDBConfig struct {
	URI            string
	Database       string
	Timeout        time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxConnIdleTTL time.Duration
}

// DefaultMongoDBConfig 默认MongoDB配置
func DefaultMongoDBConfig() *MongoDBConfig {
	return &MongoDBConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "novel_rm",
		Timeout:        10 * time.Second,
		MaxPoolSize:    10,
		MinPoolSize:    2,
		MaxConnIdleTTL: 30 * time.Minute,
	}
}

// MongoDBInstance MongoDB单例实例
type MongoDBInstance struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoDBConfig
	mu       sync.RWMutex
}


/**
var 是 Go 中声明变量的关键字。这里的写法叫分组变量声明：
var (
	mongoInstance *MongoDBInstance
	mongoOnce     sync.Once
)

等价于分开写：
var mongoInstance *MongoDBInstance
var mongoOnce sync.Once
*/
var (
	mongoInstance *MongoDBInstance
	mongoOnce     sync.Once
)

// loadMongoConfig 从环境变量加载配置
func loadMongoConfig() *MongoDBConfig {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("未找到 .env 文件，使用系统环境变量: %v", err)
	}

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

	return config
}

// GetMongoInstance 获取MongoDB单例实例（自动初始化）
func GetMongoInstance() *MongoDBInstance {
	//sync.Once的读方法，内置一个匿名函数
	mongoOnce.Do(func() {
		// 加载配置
		config := loadMongoConfig()

		// 创建实例
		mongoInstance = &MongoDBInstance{
			config: config,
		}

		// 自动连接
		if err := mongoInstance.Connect(); err != nil {
			//抛出错误的一种方式
			log.Fatalf("MongoDB自动连接失败: %v", err)
		}

		log.Printf("✅ MongoDB自动连接成功! 数据库: %s", config.Database)
	})
	return mongoInstance
}

// WithConfig 设置配置
func (m *MongoDBInstance) WithConfig(config *MongoDBConfig) *MongoDBInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return m
}

// Connect 连接到MongoDB
func (m *MongoDBInstance) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		return nil // 已连接
	}

	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(m.config.URI)
	clientOptions.SetMaxPoolSize(m.config.MaxPoolSize)
	clientOptions.SetMinPoolSize(m.config.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(m.config.MaxConnIdleTTL)
	clientOptions.SetConnectTimeout(m.config.Timeout)
	clientOptions.SetServerSelectionTimeout(m.config.Timeout)

	// 连接到MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

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

// Disconnect 断开MongoDB连接
func (m *MongoDBInstance) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	err := m.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("断开MongoDB连接失败: %v", err)
	}

	m.client = nil
	m.database = nil
	log.Println("MongoDB连接已断开")
	return nil
}

// GetClient 获取MongoDB客户端
func (m *MongoDBInstance) GetClient() *mongo.Client {
	//读锁
	m.mu.RLock()
	//延迟执行，test在最前面
	defer m.mu.RUnlock()

	if m.client == nil {
		log.Fatal("MongoDB未初始化，请先调用Connect()")
	}
	return m.client
}

// GetDatabase 获取数据库
func (m *MongoDBInstance) GetDatabase() *mongo.Database {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.database == nil {
		log.Fatal("MongoDB数据库未初始化，请先调用Connect()")
	}
	return m.database
}

// GetCollection 获取集合
func (m *MongoDBInstance) GetCollection(name string) *mongo.Collection {
	return m.GetDatabase().Collection(name)
}

// IsConnected 检查是否已连接
func (m *MongoDBInstance) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.client == nil {
		return false
	}

	//设置timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.client.Ping(ctx, nil)
	return err == nil
}

// GetStats 获取连接统计信息，这个是一个很有意思的使用
func (m *MongoDBInstance) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.client == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	//context解决的是泄漏的问题
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取服务器状态
	serverStatus, err := m.client.Database("admin").RunCommand(ctx, map[string]interface{}{
		"serverStatus": 1,
	}).DecodeBytes()

	//直接定义和返回map
	stats := map[string]interface{}{
		"connected":   m.IsConnected(),
		"database":    m.config.Database,
		"uri":         m.config.URI,
		"max_pool_size": m.config.MaxPoolSize,
		"min_pool_size": m.config.MinPoolSize,
	}

	if err == nil {
		stats["server_info"] = serverStatus
	}

	return stats
}

// Close 关闭连接（程序退出时调用）
func (m *MongoDBInstance) Close() error {
	return m.Disconnect()
}