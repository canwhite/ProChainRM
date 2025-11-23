package service

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// ChaincodeMigrationService 链码迁移服务
type ChaincodeMigrationService struct {
	contract *client.Contract
}

// NewChaincodeMigrationService 创建链码迁移服务
func NewChaincodeMigrationService(gateway *client.Gateway) (*ChaincodeMigrationService, error) {
	//先找到channel
	network := gateway.GetNetwork("mychannel")
	if network == nil {
		return nil, fmt.Errorf("无法获取network对象")
	}

	//再拿到合约
	contract := network.GetContract("novel-basic")
	if contract == nil {
		return nil, fmt.Errorf("无法获取contract")
	}

	return &ChaincodeMigrationService{
		contract: contract,
	}, nil
}

// InitChaincodeFromMongoDB 从 MongoDB 数据初始化链码
func (cms *ChaincodeMigrationService) InitChaincodeFromMongoDB(ctx context.Context) (string, error) {
	log.Println("🚀 开始从 MongoDB 初始化链码...")

	// 创建数据迁移服务
	migrationService := NewMigrationService()

	// 从 MongoDB 读取所有数据
	mongoData, err := migrationService.GetAllDataFromMongoDB()
	if err != nil {
		return "", fmt.Errorf("从 MongoDB 读取数据失败: %v", err)
	}

	// 显示数据统计
	stats := mongoData.GetStats()
	log.Printf("📊 MongoDB 数据统计: %+v", stats)

	// 检查是否有数据需要导入
	if len(mongoData.Novels) == 0 && len(mongoData.UserCredits) == 0 {
		log.Println("⚠️ MongoDB 中没有数据需要导入")
		return "MongoDB 中没有数据需要导入", nil
	}

	// 将数据转换为 JSON
	jsonData, err := mongoData.ToJSON()
	if err != nil {
		return "", fmt.Errorf("序列化 MongoDB 数据失败: %v", err)
	}

	log.Printf("📦 准备导入链码的数据大小: %d 字符", len(jsonData))

	// 调用链码的 InitFromMongoDB 方法
	result, err := cms.contract.SubmitTransaction("InitFromMongoDB", jsonData)
	if err != nil {
		return "", fmt.Errorf("调用链码 InitFromMongoDB 失败: %v", err)
	}

	log.Println("✅ 链码初始化完成!")
	return string(result), nil
}

// GetChaincodeStatus 获取链码状态
func (cms *ChaincodeMigrationService) GetChaincodeStatus(ctx context.Context) (map[string]interface{}, error) {
	log.Println("🔍 检查链码状态...")

	// 获取所有 novels
	novelsResult, err := cms.contract.EvaluateTransaction("GetAllNovels")
	if err != nil {
		return nil, fmt.Errorf("获取 novels 失败: %v", err)
	}

	// 获取所有 userCredits
	userCreditsResult, err := cms.contract.EvaluateTransaction("GetAllUserCredits")
	if err != nil {
		return nil, fmt.Errorf("获取 userCredits 失败: %v", err)
	}

	status := map[string]interface{}{
		"chaincodeConnected": true,
		"novelsCount":        len(string(novelsResult)),
		"userCreditsCount":   len(string(userCreditsResult)),
		"novelsDataSize":     len(novelsResult),
		"userCreditsDataSize": len(userCreditsResult),
	}

	log.Printf("📊 链码状态: %+v", status)
	return status, nil
}

// ValidateDataConsistency 验证链上链下数据一致性
func (cms *ChaincodeMigrationService) ValidateDataConsistency(ctx context.Context) (map[string]interface{}, error) {
	log.Println("🔍 验证链上链下数据一致性...")

	// 获取链上数据
	chaincodeStatus, err := cms.GetChaincodeStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取链码状态失败: %v", err)
	}

	// 获取链下数据
	migrationService := NewMigrationService()
	mongoData, err := migrationService.GetAllDataFromMongoDB()
	if err != nil {
		return nil, fmt.Errorf("获取 MongoDB 数据失败: %v", err)
	}

	report := map[string]interface{}{
		"consistent":    true,
		"discrepancies": []string{},
	}

	// 2025/11/22 15:13:07 🔍 验证链上链下数据一致性...
	// 2025/11/22 15:13:07 🔍 检查链码状态...
	// 2025/11/22 15:13:07 📊 链码状态: map[chaincodeConnected:true novelsCount:23527 novelsDataSize:23527 userCreditsCount:151 userCreditsDataSize:151]
	// 2025/11/22 15:13:07 🔍 开始从 MongoDB 读取数据...
	// 2025/11/22 15:13:07 ✅ 从 MongoDB 读取完成: novels=1, userCredits=1
	// 2025/11/22 15:13:07 ❌ 链上链下数据不一致: [Novels 数量不一致: 链上 23527, 链下 1 UserCredits 数量不一致: 链上 151, 链下 1]
	// 2025/11/22 15:13:07 ⚠️ 数据一致性验证发现问题: [Novels 数量不一致: 链上 23527, 链下 1 UserCredits 数量不一致: 链上 151, 链下 1]
	// 2025/11/22 15:13:26 🔍 [DEBUG] getUserCredit 请求，用户ID: 691058f50987397c91e4e078
	// 2025/11/22 15:13:26 📡 [DEBUG] 调用 creditService.ReadUserCredit(691058f50987397c91e4e078)

	// 比较 novels 数量
	chaincodeNovelsCount := chaincodeStatus["novelsCount"].(int)
	mongoNovelsCount := len(mongoData.Novels)
	if chaincodeNovelsCount != mongoNovelsCount {
		report["consistent"] = false
		report["discrepancies"] = append(report["discrepancies"].([]string),
			fmt.Sprintf("Novels 数量不一致: 链上 %d, 链下 %d", chaincodeNovelsCount, mongoNovelsCount))
	}

	// 比较 userCredits 数量
	chaincodeCreditsCount := chaincodeStatus["userCreditsCount"].(int)
	mongoCreditsCount := len(mongoData.UserCredits)
	if chaincodeCreditsCount != mongoCreditsCount {
		report["consistent"] = false
		report["discrepancies"] = append(report["discrepancies"].([]string),
			fmt.Sprintf("UserCredits 数量不一致: 链上 %d, 链下 %d", chaincodeCreditsCount, mongoCreditsCount))
	}

	if report["consistent"].(bool) {
		log.Println("✅ 链上链下数据一致")
	} else {
		log.Printf("❌ 链上链下数据不一致: %+v", report["discrepancies"])
	}

	return report, nil
}