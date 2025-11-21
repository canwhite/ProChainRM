package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type EventService struct {
	network      *client.Network
	mongoService *MongoService
}

func NewEventService(gateway *client.Gateway) *EventService {
	//先获取network再获取event
	network := gateway.GetNetwork("mychannel")

	// 创建MongoDB服务实例
	mongoService := NewMongoService()
	// 创建索引————开始就创建好index
	if err := mongoService.CreateIndexes(); err != nil {
		fmt.Printf("Warning: Failed to create MongoDB indexes: %v\n", err)
	}

	return &EventService{
		network:      network,
		mongoService: mongoService,
	}
}

func (es *EventService) StartEventListening(ctx context.Context) error {
	fmt.Println("🎧 Starting event listener...")
	
	events, err := es.network.ChaincodeEvents(ctx, "novel-basic")
	if err != nil {
		// 是的，%v是Go语言fmt包中最通用的格式化动词，几乎所有类型都可以用%v来输出其默认格式。
		// 例如：字符串、数字、结构体、切片、map、error等类型都可以用%v打印出来。
		// 但%w只能用于fmt.Errorf，并且只能用于error类型的包装，不能用于其他类型。
		return fmt.Errorf("failed to start event listening: %w", err)
	}
	//监听数据
	go func() {
		for event := range events {
			//多路复用器
			select {
			case <-ctx.Done():
				return
			default:
				novelOrUserCredit := formatJSON(event.Payload)
				fmt.Printf("\n<-- Chaincode event received: %s - %s\n", event.EventName, novelOrUserCredit)

				// 处理事件并同步到MongoDB
				es.processEventAndSyncToMongoDB(event.EventName, event.Payload)
			}
		}
	}()

	return nil
}

// 监听特定事件
func (es *EventService) ListenForSpecificEvents(ctx context.Context, eventNames []string) error {
	events, err := es.network.ChaincodeEvents(ctx, "novel-basic", client.WithStartBlock(0))
	if err != nil {
		return fmt.Errorf("failed to start specific event listening: %w", err)
	}
	go func() {
		fmt.Printf("🔍 Listening for specific events: %v\n", eventNames)
		for {
			select {
			case event, ok := <-events:
				if !ok {
					fmt.Println("All events processed, closing listener")
					return // channel关闭，所有事件处理完
				}

				// 检查是否是指定事件
				for _, name := range eventNames {
					if event.EventName == name {
						novelOrUserCredit := formatJSON(event.Payload)
						fmt.Printf("\n<-- Event from block %d: %s - %s\n",
							event.BlockNumber, event.EventName, novelOrUserCredit)

						// 处理事件并同步到MongoDB
						es.processEventAndSyncToMongoDB(event.EventName, event.Payload)
						break
					}
				}

			case <-ctx.Done():
				fmt.Println("Context cancelled, stopping listener")
				return
			}
		}
	}()
	return nil
}

func formatJSON(data []byte) string {
	//bytes.Buffer：可增长的缓冲区，性能更好
	var result bytes.Buffer
	//&result写入目标，data源数据，""前缀，"  "缩进2空格
	if err := json.Indent(&result, data, "", "  "); err != nil {
		//这个是复制字节数据到新字符串
		return string(data)
	}
	//这个是bytes.Buffer转换为字符串
	return result.String()
}

// processEventAndSyncToMongoDB 处理事件并同步到MongoDB
func (es *EventService) processEventAndSyncToMongoDB(eventName string, payload []byte) {
	// 添加调试信息
	fmt.Printf("\n🔍 [DEBUG] 接收到事件: %s\n", eventName)
	fmt.Printf("📦 [DEBUG] 事件载荷长度: %d 字节\n", len(payload))

	// 解析事件载荷
	var eventData map[string]interface{}
	if err := json.Unmarshal(payload, &eventData); err != nil {
		fmt.Printf("❌ Failed to parse event payload: %v\n", err)
		return
	}

	// 打印解析后的数据（用于调试）
	fmt.Printf("📋 [DEBUG] 解析后的事件数据:\n")
	for key, value := range eventData {
		fmt.Printf("   %s: %v (类型: %T)\n", key, value, value)
	}

	// 根据事件类型进行相应的MongoDB操作
	switch eventName {
	case "CreateNovel":
		fmt.Println("📝 [DEBUG] 处理 CreateNovel 事件...")
		es.handleCreateNovelEvent(eventData)
	case "UpdateNovel":
		fmt.Println("📝 [DEBUG] 处理 UpdateNovel 事件...")
		es.handleUpdateNovelEvent(eventData)
	case "CreateUserCredit":
		fmt.Println("💰 [DEBUG] 处理 CreateUserCredit 事件...")
		es.handleCreateUserCreditEvent(eventData)
	case "UpdateUserCredit":
		fmt.Println("💰 [DEBUG] 处理 UpdateUserCredit 事件...")
		es.handleUpdateUserCreditEvent(eventData)
	case "CreateCreditHistory":
		fmt.Println("📜 [DEBUG] 处理 CreateCreditHistory 事件...")
		es.handleCreateCreditHistoryEvent(eventData)
	case "ConsumeUserToken":
		fmt.Println("🔥 [DEBUG] 处理 ConsumeUserToken 事件...")
		es.handleConsumeUserTokenEvent(eventData)
	default:
		fmt.Printf("ℹ️ [DEBUG] 未处理的事件类型: %s\n", eventName)
		fmt.Printf("🔍 [DEBUG] 已知的事件类型: CreateNovel, UpdateNovel, CreateUserCredit, UpdateUserCredit, CreateCreditHistory, ConsumeUserToken\n")
	}
}

// handleCreateNovelEvent 处理创建小说事件
func (es *EventService) handleCreateNovelEvent(eventData map[string]interface{}) {
	fmt.Println("📝 Processing CreateNovel event...")

	if err := es.mongoService.CreateNovelInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync CreateNovel to MongoDB: %v\n", err)
	}
}

// handleUpdateNovelEvent 处理更新小说事件
func (es *EventService) handleUpdateNovelEvent(eventData map[string]interface{}) {
	fmt.Println("📝 Processing UpdateNovel event...")

	if err := es.mongoService.UpdateNovelInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync UpdateNovel to MongoDB: %v\n", err)
	}
}

// handleCreateUserCreditEvent 处理创建用户积分事件
func (es *EventService) handleCreateUserCreditEvent(eventData map[string]interface{}) {
	fmt.Println("💰 Processing CreateUserCredit event...")

	if err := es.mongoService.CreateUserCreditInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync CreateUserCredit to MongoDB: %v\n", err)
	}
}

// handleUpdateUserCreditEvent 处理更新用户积分事件
func (es *EventService) handleUpdateUserCreditEvent(eventData map[string]interface{}) {
	fmt.Println("💰 [DEBUG] 开始处理 UpdateUserCredit 事件...")

	// 打印关键字段信息
	userId := getStringFromMap(eventData, "userId")
	credit := getIntFromMap(eventData, "credit")
	totalUsed := getIntFromMap(eventData, "totalUsed")
	totalRecharge := getIntFromMap(eventData, "totalRecharge")

	fmt.Printf("🔍 [DEBUG] UserCredit 数据 - userId: %s, credit: %d, totalUsed: %d, totalRecharge: %d\n",
		userId, credit, totalUsed, totalRecharge)

	if err := es.mongoService.UpdateUserCreditInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync UpdateUserCredit to MongoDB: %v\n", err)
	} else {
		fmt.Println("✅ [DEBUG] UpdateUserCredit 同步到 MongoDB 成功!")
	}
}

// handleCreateCreditHistoryEvent 处理创建积分历史事件
func (es *EventService) handleCreateCreditHistoryEvent(eventData map[string]interface{}) {
	fmt.Println("📜 Processing CreateCreditHistory event...")

	if err := es.mongoService.CreateCreditHistoryInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync CreateCreditHistory to MongoDB: %v\n", err)
	}
}

// handleConsumeUserTokenEvent 处理消费用户代币事件
func (es *EventService) handleConsumeUserTokenEvent(eventData map[string]interface{}) {
	fmt.Println("🔥 Processing ConsumeUserToken event...")

	// ConsumeUserToken事件会触发UserCredit的更新，所以这里主要是同步UserCredit
	if err := es.mongoService.UpdateUserCreditInMongo(eventData); err != nil {
		fmt.Printf("❌ Failed to sync ConsumeUserToken to MongoDB: %v\n", err)
	}
}

// 辅助函数：从map中安全获取字符串值
func getStringFromMap(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从map中安全获取整数值
func getIntFromMap(data map[string]interface{}, key string) int {
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
