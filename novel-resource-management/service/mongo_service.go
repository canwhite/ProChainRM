package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"novel-resource-management/database"
)

type MongoService struct {
	db *database.MongoDBInstance
}

func NewMongoService() *MongoService {
	return &MongoService{
		db: database.GetMongoInstance(),
	}
}

// Novel相关的MongoDB操作

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// CreateNovelInMongo 在MongoDB中创建Novel记录
func (ms *MongoService) CreateNovelInMongo(novel map[string]interface{}) error {
	collection := ms.db.GetCollection("novels")

	// 获取或生成ID
	id := getString(novel, "id")
	if id == "" {
		id = generateID() // 生成唯一ID
	}

	// 将map转换为Novel结构
	novelData := &database.Novel{
		ID:           id, // 使用传入的ID或生成的ID
		Author:       getString(novel, "author"),
		StoryOutline: getString(novel, "storyOutline"),
		Subsections:  getString(novel, "subsections"),
		Characters:   getString(novel, "characters"),
		Items:        getString(novel, "items"),
		TotalScenes:  getString(novel, "totalScenes"),
		CreatedAt:    getString(novel, "createdAt"),
		UpdatedAt:    getString(novel, "updatedAt"),
	}

	// 检查是否已存在相同的novel（根据storyOutline唯一索引）
	// 因为我们为storyOutline创建了唯一索引，所以只需要检查storyOutline是否重复
	filter := bson.M{"storyOutline": novelData.StoryOutline}
	var existingNovel database.Novel
	//将结果写入existingNovel，这个好方便呀
	err := collection.FindOne(context.Background(), filter).Decode(&existingNovel)
	if err == nil {
		log.Printf("Novel already exists in MongoDB, storyOutline: %s", novelData.StoryOutline)
		return nil // 已存在，不重复创建
	}

	// 插入新记录
	_, err = collection.InsertOne(context.Background(), novelData)
	if err != nil {
		return fmt.Errorf("failed to create novel in MongoDB: %v", err)
	}

	log.Printf("✅ Created novel in MongoDB: author=%s", novelData.Author)
	return nil
}

// UpdateNovelInMongo 在MongoDB中更新Novel记录
func (ms *MongoService) UpdateNovelInMongo(novel map[string]interface{}) error {
	collection := ms.db.GetCollection("novels")

	// 构建更新数据
	updateData := bson.M{
		//set
		"$set": bson.M{
			"author":       getString(novel, "author"),
			"storyOutline": getString(novel, "storyOutline"),
			"subsections":  getString(novel, "subsections"),
			"characters":   getString(novel, "characters"),
			"items":        getString(novel, "items"),
			"totalScenes":  getString(novel, "totalScenes"),
			"updatedAt":    getString(novel, "updatedAt"),
		},
	}

	// 根据storyOutline查找并更新（因为storyOutline是唯一索引）
	filter := bson.M{"storyOutline": getString(novel, "storyOutline")}
	result, err := collection.UpdateOne(context.Background(), filter, updateData)
	if err != nil {
		return fmt.Errorf("failed to update novel in MongoDB: %v", err)
	}

	if result.MatchedCount == 0 {
		// 如果没有找到记录，则创建新记录
		return ms.CreateNovelInMongo(novel)
	}

	log.Printf("✅ Updated novel in MongoDB: storyOutline=%s", getString(novel, "storyOutline"))
	return nil
}

// UserCredit相关的MongoDB操作

// CreateUserCreditInMongo 在MongoDB中创建UserCredit记录
func (ms *MongoService) CreateUserCreditInMongo(userCredit map[string]interface{}) error {
	collection := ms.db.GetCollection("user_credits")

	// 获取或生成ID
	id := getString(userCredit, "id")
	if id == "" {
		id = generateID() // 生成唯一ID
	}

	// 将map转换为UserCredit结构
	userCreditData := &database.UserCredit{
		ID:            id, // 使用传入的ID或生成的ID
		UserID:        getString(userCredit, "userId"),
		Credit:        getInt(userCredit, "credit"),
		TotalUsed:     getInt(userCredit, "totalUsed"),
		TotalRecharge: getInt(userCredit, "totalRecharge"),
		CreatedAt:     getString(userCredit, "createdAt"),
		UpdatedAt:     getString(userCredit, "updatedAt"),
	}

	// 检查是否已存在相同的用户积分记录
	filter := bson.M{"userId": userCreditData.UserID}
	var existingUserCredit database.UserCredit
	err := collection.FindOne(context.Background(), filter).Decode(&existingUserCredit)
	if err == nil {
		log.Printf("UserCredit already exists in MongoDB, userId: %s", userCreditData.UserID)
		return nil // 已存在，不重复创建
	}

	// 插入新记录
	_, err = collection.InsertOne(context.Background(), userCreditData)
	if err != nil {
		return fmt.Errorf("failed to create user credit in MongoDB: %v", err)
	}

	log.Printf("✅ Created user credit in MongoDB: userId=%s, credit=%d", userCreditData.UserID, userCreditData.Credit)
	return nil
}

// UpdateUserCreditInMongo 在MongoDB中更新UserCredit记录
func (ms *MongoService) UpdateUserCreditInMongo(userCredit map[string]interface{}) error {
	collection := ms.db.GetCollection("user_credits")

	// 构建更新数据
	updateData := bson.M{
		"$set": bson.M{
			"credit":         getInt(userCredit, "credit"),
			"totalUsed":      getInt(userCredit, "totalUsed"),
			"totalRecharge":  getInt(userCredit, "totalRecharge"),
			"updatedAt":      getString(userCredit, "updatedAt"),
		},
	}

	// 根据userId查找并更新
	filter := bson.M{"userId": getString(userCredit, "userId")}
	result, err := collection.UpdateOne(context.Background(), filter, updateData)
	if err != nil {
		return fmt.Errorf("failed to update user credit in MongoDB: %v", err)
	}

	if result.MatchedCount == 0 {
		// 如果没有找到记录，则创建新记录
		return ms.CreateUserCreditInMongo(userCredit)
	}

	log.Printf("✅ Updated user credit in MongoDB: userId=%s, credit=%d",
		getString(userCredit, "userId"), getInt(userCredit, "credit"))
	return nil
}

// CreateCreditHistoryInMongo 在MongoDB中创建CreditHistory记录
func (ms *MongoService) CreateCreditHistoryInMongo(creditHistory map[string]interface{}) error {
	collection := ms.db.GetCollection("credit_histories")

	// 获取或生成ID
	id := getString(creditHistory, "id")
	if id == "" {
		id = generateID() // 生成唯一ID
	}

	// 将map转换为CreditHistory结构
	creditHistoryData := &database.CreditHistory{
		ID:          id, // 使用传入的ID或生成的ID
		UserID:      getString(creditHistory, "userId"),
		Amount:      getInt(creditHistory, "amount"),
		Type:        getString(creditHistory, "type"),
		Description: getString(creditHistory, "description"),
		Timestamp:   getString(creditHistory, "timestamp"),
		NovelID:     getString(creditHistory, "novelId"),
	}

	// 插入新记录
	_, err := collection.InsertOne(context.Background(), creditHistoryData)
	if err != nil {
		return fmt.Errorf("failed to create credit history in MongoDB: %v", err)
	}

	log.Printf("✅ Created credit history in MongoDB: userId=%s, amount=%d, type=%s",
		creditHistoryData.UserID, creditHistoryData.Amount, creditHistoryData.Type)
	return nil
}

// 辅助函数

// getString 从map中安全获取string值
func getString(data map[string]interface{}, key string) string {
	//comma ok 语法
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// getInt 从map中安全获取int值
func getInt(data map[string]interface{}, key string) int {
	if value, exists := data[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64: // JSON数字默认解析为float64
			return int(v)
		case string:
			// 如果是字符串形式的数字，尝试解析
			if num, err := strconv.Atoi(v); err == nil {
				return num
			}
		}
	}
	return 0
}

// CreateIndexes 创建必要的索引 - 数据库查询加速器
// 小白解释：索引就像书的目录，有了目录就能快速找到想要的内容，不用一页一页翻
func (ms *MongoService) CreateIndexes() error {
	// 创建上下文，告诉MongoDB这是一个完整的操作，不要中途打断
	ctx := context.Background()

	log.Println("🔍 开始为数据库创建索引...")

	// 第一步：为小说集合创建故事大纲索引
	// 使用 storyOutline 作为唯一索引，确保每个故事都是独一无二的
	log.Println("📚 为 novels 集合创建 storyOutline 索引...")
	novelsCollection := ms.db.GetCollection("novels")

	// 首先删除可能存在的错误索引
	indexes, err := novelsCollection.Indexes().List(ctx)
	if err == nil {
		for indexes.Next(ctx) {
			var index bson.M
			indexes.Decode(&index)
			if name, ok := index["name"]; ok && name == "novels_userId_novelId_key" {
				log.Println("🗑️ 删除错误的 userId+novelId 索引...")
				_, dropErr := novelsCollection.Indexes().DropOne(ctx, "novels_userId_novelId_key")
				if dropErr != nil {
					log.Printf("⚠️ 删除错误索引失败: %v", dropErr)
				} else {
					log.Println("✅ 成功删除错误的 userId+novelId 索引")
				}
			}
		}
	}

	// 创建正确的 storyOutline 索引
	// {"storyOutline": 1} 表示按故事大纲升序排列
	// SetUnique(true) 表示每个故事大纲必须是唯一的
	_, err = novelsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"storyOutline": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("❌ 创建 novels 集合的 storyOutline 索引失败: %v", err)
	}
	log.Println("✅ novels 集合的 storyOutline 索引创建成功")

	// 第二步：为用户积分集合创建用户ID索引
	// 为什么要用userId？因为查询用户积分信息时，总是根据用户ID来查
	log.Println("💰 为 user_credits 集合创建 userId 索引...")
	userCreditsCollection := ms.db.GetCollection("user_credits")

	// SetUnique(true) 确保每个用户只能有一个积分记录
	_, err = userCreditsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"userId": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("❌ 创建 user_credits 集合的 userId 索引失败: %v", err)
	}
	log.Println("✅ user_credits 集合的 userId 索引创建成功")

	// 第三步：为积分历史集合创建复合索引
	// 为什么要用userId + timestamp？因为查看积分历史时，通常按用户和时间排序
	log.Println("📜 为 credit_histories 集合创建 userId + timestamp 复合索引...")
	creditHistoriesCollection := ms.db.GetCollection("credit_histories")

	// 复合索引：{"userId": 1, "timestamp": -1}
	// 1 表示升序（A-Z, 0-9），-1 表示降序（Z-A, 9-0）
	// 这样可以快速找到某个用户的所有积分历史，并按时间从新到旧排序
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "timestamp", Value: -1},
		},
	}

	_, err = creditHistoriesCollection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("❌ 创建 credit_histories 集合的 userId-timestamp 索引失败: %v", err)
	}
	log.Println("✅ credit_histories 集合的 userId-timestamp 索引创建成功")

	log.Println("🎉 所有数据库索引创建完成！查询速度将会大幅提升")
	return nil
}