package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig MongoDB连接配置
type MongoDBConfig struct {
	URI            string        `json:"uri" yaml:"uri"`
	Database       string        `json:"database" yaml:"database"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	MaxPoolSize    uint64        `json:"max_pool_size" yaml:"max_pool_size"`
	MinPoolSize    uint64        `json:"min_pool_size" yaml:"min_pool_size"`
	MaxConnIdleTTL time.Duration `json:"max_conn_idle_ttl" yaml:"max_conn_idle_ttl"`
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

var (
	mongoInstance *MongoDBInstance
	mongoOnce     sync.Once
)

// GetMongoInstance 获取MongoDB单例实例
func GetMongoInstance() *MongoDBInstance {
	mongoOnce.Do(func() {
		mongoInstance = &MongoDBInstance{
			config: DefaultMongoDBConfig(),
		}
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
	m.mu.RLock()
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.client.Ping(ctx, nil)
	return err == nil
}

// GetStats 获取连接统计信息
func (m *MongoDBInstance) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.client == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取服务器状态
	serverStatus, err := m.client.Database("admin").RunCommand(ctx, map[string]interface{}{
		"serverStatus": 1,
	}).DecodeBytes()

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