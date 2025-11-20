package main

import (
	"fmt"
	"log"

	"novel-resource-management/database"
)

func main() {
	fmt.Println("=== MongoDB单例简单测试 ===")

	// 1. 获取单例实例
	mongoInstance := database.GetMongoInstance()
	fmt.Printf("MongoDB实例地址: %p\n", mongoInstance)

	// 2. 使用自动初始化
	fmt.Println("正在初始化MongoDB连接...")
	database.AutoInitMongoDB()

	// 3. 检查连接状态
	if mongoInstance.IsConnected() {
		fmt.Println("✅ MongoDB连接成功！")
	} else {
		fmt.Println("❌ MongoDB连接失败")
		log.Fatal("无法连接到MongoDB")
	}

	// 4. 获取统计信息
	stats := mongoInstance.GetStats()
	fmt.Printf("连接信息: %+v\n", stats)

	fmt.Println("=== 测试完成 ===")
}