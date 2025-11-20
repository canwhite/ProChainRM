package main

import (
	"context"
	"fmt"
	"time"

	"novel-resource-management/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("=== MongoDB 简化测试 ===")

	// 1. 获取实例（自动加载配置和连接）
	fmt.Println("\n1. 获取MongoDB实例...")
	mongoInstance := database.GetMongoInstance()
	fmt.Printf("✅ 获取实例成功: %p\n", mongoInstance)

	// 2. 直接获取数据库（无需手动初始化）
	fmt.Println("\n2. 直接获取数据库...")
	db := mongoInstance.GetDatabase()
	fmt.Printf("✅ 数据库名称: %s\n", db.Name())

	// 3. 测试连接状态
	fmt.Println("\n3. 测试连接状态")
	if mongoInstance.IsConnected() {
		fmt.Println("✅ MongoDB连接正常")
	} else {
		fmt.Println("❌ MongoDB连接异常")
		return
	}

	// 4. 测试获取集合
	fmt.Println("\n4. 测试获取集合")
	userCreditCollection := mongoInstance.GetCollection("user_credits")
	novelCollection := mongoInstance.GetCollection("novels")
	fmt.Printf("用户积分集合: %s\n", userCreditCollection.Name())
	fmt.Printf("小说集合: %s\n", novelCollection.Name())

	// 5. 测试插入数据
	fmt.Println("\n5. 测试插入用户积分数据")
	testUserCredit := database.UserCredit{
		UserID:        "test_user_001",
		Credit:        100,
		TotalUsed:     5,
		TotalRecharge: 100,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 先删除可能存在的测试数据
	_, err := userCreditCollection.DeleteOne(context.Background(), bson.M{"user_id": "test_user_001"})
	if err != nil {
		fmt.Printf("清理测试数据失败: %v\n", err)
	}

	// 插入新数据
	result, err := userCreditCollection.InsertOne(context.Background(), testUserCredit)
	if err != nil {
		fmt.Printf("❌ 插入用户积分失败: %v\n", err)
	} else {
		fmt.Printf("✅ 插入用户积分成功，ID: %s\n", result.InsertedID)
	}

	// 6. 测试插入小说数据
	fmt.Println("\n6. 测试插入小说数据")
	testNovel := database.Novel{
		Title:       "测试小说",
		Author:      "测试作者",
		Category:    "玄幻",
		Description: "这是一本测试小说",
		Tags:        []string{"玄幻", "测试", "小说"},
		Price:       9.99,
		IsPublished: true,
		ViewCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 先删除可能存在的测试数据
	_, err = novelCollection.DeleteOne(context.Background(), bson.M{"title": "测试小说"})
	if err != nil {
		fmt.Printf("清理测试小说失败: %v\n", err)
	}

	// 插入新数据
	result, err = novelCollection.InsertOne(context.Background(), testNovel)
	if err != nil {
		fmt.Printf("❌ 插入小说失败: %v\n", err)
	} else {
		fmt.Printf("✅ 插入小说成功，ID: %s\n", result.InsertedID)
	}

	// 7. 测试查询数据
	fmt.Println("\n7. 测试查询用户积分数据")
	var foundUserCredit database.UserCredit
	err = userCreditCollection.FindOne(context.Background(), bson.M{"user_id": "test_user_001"}).Decode(&foundUserCredit)
	if err != nil {
		fmt.Printf("❌ 查询用户积分失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询用户积分成功:\n")
		fmt.Printf("   用户ID: %s\n", foundUserCredit.UserID)
		fmt.Printf("   积分: %d (类型: %T)\n", foundUserCredit.Credit, foundUserCredit.Credit) // 展示类型信息
		fmt.Printf("   已使用: %d (类型: %T)\n", foundUserCredit.TotalUsed, foundUserCredit.TotalUsed)
		fmt.Printf("   总充值: %d (类型: %T)\n", foundUserCredit.TotalRecharge, foundUserCredit.TotalRecharge)
		fmt.Printf("   创建时间: %s\n", foundUserCredit.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   ✅ 无需类型转换！直接使用int类型！\n")
	}

	// 8. 测试查询小说数据
	fmt.Println("\n8. 测试查询小说数据")
	var foundNovel database.Novel
	err = novelCollection.FindOne(context.Background(), bson.M{"title": "测试小说"}).Decode(&foundNovel)
	if err != nil {
		fmt.Printf("❌ 查询小说失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询小说成功:\n")
		fmt.Printf("   标题: %s\n", foundNovel.Title)
		fmt.Printf("   作者: %s\n", foundNovel.Author)
		fmt.Printf("   分类: %s\n", foundNovel.Category)
		fmt.Printf("   价格: %.2f (类型: %T)\n", foundNovel.Price, foundNovel.Price)
		fmt.Printf("   标签: %v\n", foundNovel.Tags)
		fmt.Printf("   是否发布: %t (类型: %T)\n", foundNovel.IsPublished, foundNovel.IsPublished)
	}

	// 9. 测试更新数据（模拟消费token）
	fmt.Println("\n9. 测试更新数据（模拟消费token）")
	if foundUserCredit.Credit > 0 {
		// 使用BSON直接更新，无需类型转换！
		filter := bson.M{"user_id": "test_user_001"}
		update := bson.M{
			"$inc": bson.M{
				"credit":     -1,        // 直接使用int
				"total_used": 1,         // 直接使用int
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}

		updateResult, err := userCreditCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Printf("❌ 更新用户积分失败: %v\n", err)
		} else {
			fmt.Printf("✅ 更新用户积分成功，匹配记录: %d，修改记录: %d\n",
				updateResult.MatchedCount, updateResult.ModifiedCount)

			// 重新查询验证结果
			var updatedUserCredit database.UserCredit
			err = userCreditCollection.FindOne(context.Background(), filter).Decode(&updatedUserCredit)
			if err == nil {
				fmt.Printf("   更新后积分: %d -> %d\n", foundUserCredit.Credit, updatedUserCredit.Credit)
				fmt.Printf("   更新后已使用: %d -> %d\n", foundUserCredit.TotalUsed, updatedUserCredit.TotalUsed)
				fmt.Printf("   ✅ BSON操作无需类型转换！\n")
			}
		}
	}

	// 10. 测试条件查询
	fmt.Println("\n10. 测试条件查询")
	cursor, err := userCreditCollection.Find(context.Background(), bson.M{
		"credit": bson.M{"$gte": 50},
	}, options.Find().SetLimit(10))
	if err != nil {
		fmt.Printf("❌ 条件查询失败: %v\n", err)
	} else {
		defer cursor.Close(context.Background())

		var richUsers []database.UserCredit
		err = cursor.All(context.Background(), &richUsers)
		if err != nil {
			fmt.Printf("❌ 解析查询结果失败: %v\n", err)
		} else {
			fmt.Printf("✅ 查询积分>=50的用户，找到 %d 个:\n", len(richUsers))
			for _, user := range richUsers {
				fmt.Printf("   用户: %s, 积分: %d\n", user.UserID, user.Credit)
			}
		}
	}

	// 11. 测试获取连接统计信息
	fmt.Println("\n11. 测试获取连接统计信息")
	stats := mongoInstance.GetStats()
	fmt.Printf("✅ 连接统计信息:\n")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// 12. 测试单例在不同地方的使用
	fmt.Println("\n12. 测试单例在不同地方的使用")
	testSingletonInDifferentFunction()

	// 13. 清理测试数据
	fmt.Println("\n13. 清理测试数据")
	_, err = userCreditCollection.DeleteOne(context.Background(), bson.M{"user_id": "test_user_001"})
	if err != nil {
		fmt.Printf("❌ 清理用户积分测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理用户积分测试数据成功")
	}

	_, err = novelCollection.DeleteOne(context.Background(), bson.M{"title": "测试小说"})
	if err != nil {
		fmt.Printf("❌ 清理小说测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理小说测试数据成功")
	}

	// 14. 断开连接（可选，程序结束时会自动断开）
	fmt.Println("\n14. 测试断开连接")
	err = mongoInstance.Disconnect()
	if err != nil {
		fmt.Printf("❌ 断开连接失败: %v\n", err)
	} else {
		fmt.Println("✅ 断开连接成功")
	}

	fmt.Println("\n=== MongoDB单例测试完成 ===")
}

// 测试单例在不同函数中的使用
func testSingletonInDifferentFunction() {
	mongoInstance := database.GetMongoInstance()
	fmt.Printf("在另一个函数中获取单例，地址: %p\n", mongoInstance)

	if mongoInstance.IsConnected() {
		fmt.Println("✅ 在另一个函数中，MongoDB连接仍然可用")

		// 测试获取集合
		collection := mongoInstance.GetCollection("test_collection")
		fmt.Printf("✅ 成功获取测试集合: %s\n", collection.Name())
	} else {
		fmt.Println("❌ 在另一个函数中，MongoDB连接不可用")
	}
}

// 演示如何在实际服务中使用
func demonstrateRealUsage() {
	fmt.Println("\n=== 实际使用演示 ===")

	// 在服务中获取MongoDB实例
	mongoInstance := database.GetMongoInstance()

	// 获取集合
	userCreditCollection := mongoInstance.GetCollection("user_credits")

	// 示例：创建或更新用户积分
	userID := "user_123"

	// 检查用户是否存在
	var userCredit database.UserCredit
	err := userCreditCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&userCredit)

	if err != nil {
		// 用户不存在，创建新用户
		newUserCredit := database.UserCredit{
			UserID:        userID,
			Credit:        50,
			TotalUsed:     0,
			TotalRecharge: 50,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err = userCreditCollection.InsertOne(context.Background(), newUserCredit)
		if err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
			return
		}
		fmt.Printf("✅ 创建新用户 %s，初始积分: %d\n", userID, newUserCredit.Credit)
	} else {
		// 用户存在，直接使用，无需类型转换！
		fmt.Printf("✅ 用户 %s 存在，当前积分: %d (直接使用int，无需转换！)\n",
			userID, userCredit.Credit)
	}
}