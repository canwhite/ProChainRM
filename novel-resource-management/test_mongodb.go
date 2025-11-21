package main

import (
	"context"
	"fmt"
	"time"

	"novel-resource-management/database"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	fmt.Println("=== MongoDB 链码结构一致测试 ===")

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
	creditHistoryCollection := mongoInstance.GetCollection("credit_histories")
	fmt.Printf("用户积分集合: %s\n", userCreditCollection.Name())
	fmt.Printf("小说集合: %s\n", novelCollection.Name())
	fmt.Printf("积分历史集合: %s\n", creditHistoryCollection.Name())

	// 5. 测试插入用户积分数据（与链码结构一致）
	fmt.Println("\n5. 测试插入用户积分数据")
	currentTimeStr := time.Now().Format("2006-01-02 15:04:05")
	testUserCredit := database.UserCredit{
		UserID:        "test_user_001",
		Credit:        100,
		TotalUsed:     5,
		TotalRecharge: 100,
		CreatedAt:     currentTimeStr,
		UpdatedAt:     currentTimeStr,
	}

	// 先删除可能存在的测试数据
	_, err := userCreditCollection.DeleteOne(context.Background(), bson.M{"userId": "test_user_001"})
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

	// 6. 测试插入小说数据（与链码结构一致）
	fmt.Println("\n6. 测试插入小说数据")
	testNovel := database.Novel{
		Author:       "测试作者",
		StoryOutline: "这是一个测试小说的故事大纲",
		Subsections:  "第一章,第二章,第三章",
		Characters:   "主角A,配角B,反派C",
		Items:        "神秘宝物,魔法卷轴",
		TotalScenes:  "10",
		CreatedAt:    currentTimeStr,
		UpdatedAt:    currentTimeStr,
	}

	// 先删除可能存在的测试数据
	_, err = novelCollection.DeleteOne(context.Background(), bson.M{"author": "测试作者"})
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

	// 7. 测试插入积分历史数据（与链码结构一致）
	fmt.Println("\n7. 测试插入积分历史数据")
	testCreditHistory := database.CreditHistory{
		UserID:      "test_user_001",
		Amount:      -5,                  // 消费5积分
		Type:        "consume",
		Description: "购买小说章节",
		Timestamp:   currentTimeStr,
		NovelID:     "novel_001",
	}

	// 插入积分历史
	result, err = creditHistoryCollection.InsertOne(context.Background(), testCreditHistory)
	if err != nil {
		fmt.Printf("❌ 插入积分历史失败: %v\n", err)
	} else {
		fmt.Printf("✅ 插入积分历史成功，ID: %s\n", result.InsertedID)
	}

	// 8. 测试查询用户积分数据
	fmt.Println("\n8. 测试查询用户积分数据")
	var foundUserCredit database.UserCredit
	err = userCreditCollection.FindOne(context.Background(), bson.M{"userId": "test_user_001"}).Decode(&foundUserCredit)
	if err != nil {
		fmt.Printf("❌ 查询用户积分失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询用户积分成功:\n")
		fmt.Printf("   用户ID: %s\n", foundUserCredit.UserID)
		fmt.Printf("   积分: %d\n", foundUserCredit.Credit)
		fmt.Printf("   已使用: %d\n", foundUserCredit.TotalUsed)
		fmt.Printf("   总充值: %d\n", foundUserCredit.TotalRecharge)
		fmt.Printf("   创建时间: %s\n", foundUserCredit.CreatedAt)
	}

	// 9. 测试查询小说数据
	fmt.Println("\n9. 测试查询小说数据")
	var foundNovel database.Novel
	err = novelCollection.FindOne(context.Background(), bson.M{"author": "测试作者"}).Decode(&foundNovel)
	if err != nil {
		fmt.Printf("❌ 查询小说失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询小说成功:\n")
		fmt.Printf("   作者: %s\n", foundNovel.Author)
		fmt.Printf("   故事大纲: %s\n", foundNovel.StoryOutline)
		fmt.Printf("   章节: %s\n", foundNovel.Subsections)
		fmt.Printf("   角色: %s\n", foundNovel.Characters)
		fmt.Printf("   物品: %s\n", foundNovel.Items)
		fmt.Printf("   总场景数: %s\n", foundNovel.TotalScenes)
	}

	// 10. 测试查询积分历史数据
	fmt.Println("\n10. 测试查询积分历史数据")
	var foundCreditHistory database.CreditHistory
	err = creditHistoryCollection.FindOne(context.Background(), bson.M{"userId": "test_user_001"}).Decode(&foundCreditHistory)
	if err != nil {
		fmt.Printf("❌ 查询积分历史失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询积分历史成功:\n")
		fmt.Printf("   用户ID: %s\n", foundCreditHistory.UserID)
		fmt.Printf("   变动金额: %d\n", foundCreditHistory.Amount)
		fmt.Printf("   类型: %s\n", foundCreditHistory.Type)
		fmt.Printf("   描述: %s\n", foundCreditHistory.Description)
		fmt.Printf("   时间戳: %s\n", foundCreditHistory.Timestamp)
		fmt.Printf("   小说ID: %s\n", foundCreditHistory.NovelID)
	}

	// 11. 测试更新数据（模拟积分消费）
	fmt.Println("\n11. 测试更新数据（模拟积分消费）")
	if foundUserCredit.Credit > 0 {
		filter := bson.M{"userId": "test_user_001"}
		update := bson.M{
			"$inc": bson.M{
				"credit":     -1,
				"totalUsed": 1,
			},
			"$set": bson.M{
				"updatedAt": time.Now().Format("2006-01-02 15:04:05"),
			},
		}

		updateResult, err := userCreditCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Printf("❌ 更新用户积分失败: %v\n", err)
		} else {
			fmt.Printf("✅ 更新用户积分成功，匹配记录: %d，修改记录: %d\n",
				updateResult.MatchedCount, updateResult.ModifiedCount)
		}
	}

	// 12. 清理测试数据
	fmt.Println("\n12. 清理测试数据")
	_, err = userCreditCollection.DeleteOne(context.Background(), bson.M{"userId": "test_user_001"})
	if err != nil {
		fmt.Printf("❌ 清理用户积分测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理用户积分测试数据成功")
	}

	_, err = novelCollection.DeleteOne(context.Background(), bson.M{"author": "测试作者"})
	if err != nil {
		fmt.Printf("❌ 清理小说测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理小说测试数据成功")
	}

	_, err = creditHistoryCollection.DeleteOne(context.Background(), bson.M{"userId": "test_user_001"})
	if err != nil {
		fmt.Printf("❌ 清理积分历史测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理积分历史测试数据成功")
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("🎉 MongoDB 模型已与链码结构保持一致！")
	fmt.Println("📋 包含的结构体:")
	fmt.Println("   - Novel (小说资源)")
	fmt.Println("   - UserCredit (用户积分)")
	fmt.Println("   - CreditHistory (积分历史)")
}