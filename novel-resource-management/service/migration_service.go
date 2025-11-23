package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"novel-resource-management/database"
)

// MigrationService 处理数据迁移服务
type MigrationService struct {
	mongoService *MongoService
}

// NewMigrationService 创建数据迁移服务
func NewMigrationService() *MigrationService {
	return &MigrationService{
		mongoService: NewMongoService(),
	}
}

// MongoDBData 从 MongoDB 读取的所有数据
type MongoDBData struct {
	Novels      []*database.Novel      `json:"novels"`
	UserCredits []*database.UserCredit `json:"userCredits"`
}

// GetAllDataFromMongoDB 从 MongoDB 读取所有 novels 和 userCredits 数据
func (ms *MigrationService) GetAllDataFromMongoDB() (*MongoDBData, error) {
	log.Println("🔍 开始从 MongoDB 读取数据...")

	result := &MongoDBData{
		Novels:      make([]*database.Novel, 0),
		UserCredits: make([]*database.UserCredit, 0),
	}

	// 读取所有 novels
	if err := ms.getAllNovels(result); err != nil {
		return nil, fmt.Errorf("读取 novels 数据失败: %v", err)
	}

	// 读取所有 userCredits
	if err := ms.getAllUserCredits(result); err != nil {
		return nil, fmt.Errorf("读取 userCredits 数据失败: %v", err)
	}

	log.Printf("✅ 从 MongoDB 读取完成: novels=%d, userCredits=%d",
		len(result.Novels), len(result.UserCredits))

	return result, nil
}

// getAllNovels 从 MongoDB 读取所有小说数据
func (ms *MigrationService) getAllNovels(result *MongoDBData) error {
	collection := ms.mongoService.db.GetCollection("novels")

	cursor, err := collection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("查询 novels 失败: %v", err)
	}
	defer cursor.Close(context.Background())

	//对查询内容进行循环，用的是next
	for cursor.Next(context.Background()) {
		var novel database.Novel
		if err := cursor.Decode(&novel); err != nil {
			log.Printf("⚠️ 解析 novel 数据失败: %v", err)
			continue
		}
		result.Novels = append(result.Novels, &novel)
	}

	return nil
}

// getAllUserCredits 从 MongoDB 读取所有用户积分数据
func (ms *MigrationService) getAllUserCredits(result *MongoDBData) error {
	collection := ms.mongoService.db.GetCollection("user_credits")

	cursor, err := collection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("查询 user_credits 失败: %v", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var userCredit database.UserCredit
		//反序列化的方式很简单
		if err := cursor.Decode(&userCredit); err != nil {
			log.Printf("⚠️ 解析 userCredit 数据失败: %v", err)
			continue
		}
		result.UserCredits = append(result.UserCredits, &userCredit)
	}

	return nil
}

// ToJSON 将数据转换为 JSON 字符串，用于传递给链码
func (md *MongoDBData) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(md)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %v", err)
	}
	return string(jsonBytes), nil
}

// GetStats 获取数据统计信息
func (md *MongoDBData) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"totalNovels":       len(md.Novels),
		"totalUserCredits":  len(md.UserCredits),
		"totalCreditSum":    md.calculateTotalCredit(),
		"averageCredit":     md.calculateAverageCredit(),
	}
}

// calculateTotalCredit 计算总积分
func (md *MongoDBData) calculateTotalCredit() int {
	total := 0
	for _, uc := range md.UserCredits {
		total += uc.Credit
	}
	return total
}

// calculateAverageCredit 计算平均积分
func (md *MongoDBData) calculateAverageCredit() float64 {
	if len(md.UserCredits) == 0 {
		return 0
	}
	return float64(md.calculateTotalCredit()) / float64(len(md.UserCredits))
}