package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=== MongoDB 连接详细调试 ===")

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("未找到 .env 文件，使用系统环境变量: %v", err)
	}

	// 1. 检查环境变量
	fmt.Println("\n1. 检查环境变量:")
	uri := os.Getenv("MONGODB_URI")
	database := os.Getenv("MONGODB_DATABASE")

	fmt.Printf("   MONGODB_URI: %s\n", uri)
	fmt.Printf("   MONGODB_DATABASE: %s\n", database)

	// 2. 检查不同URI的连接情况
	testURIs := []string{
		"mongodb://localhost:27017",                    // 本地默认
		"mongodb://127.0.0.1:27017",                    // 本地IP
		"mongodb://host.docker.internal:27017",         // Docker host
		uri,                                            // 从环境变量读取的完整URI
	}

	for i, testURI := range testURIs {
		fmt.Printf("\n%d. 测试URI: %s\n", i+2, testURI)
		testConnection(testURI)
	}
}

func testConnection(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建客户端选项
	clientOptions := options.Client().ApplyURI(uri)

	// 设置连接超时
	clientOptions.SetConnectTimeout(5 * time.Second)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)

	fmt.Printf("   正在连接...")

	// 尝试连接
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}

	defer client.Disconnect(ctx)

	// 测试ping
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Printf("❌ Ping失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 连接成功\n")

	// 尝试获取数据库列表
	databases, err := client.ListDatabaseNames(ctx, map[string]interface{}{})
	if err != nil {
		fmt.Printf("   ⚠️ 获取数据库列表失败: %v\n", err)
	} else {
		fmt.Printf("   📋 可用数据库: %v\n", databases)
	}
}