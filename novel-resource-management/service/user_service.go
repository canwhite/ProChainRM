package service

import (
	"context"
	"crypto/hmac" //有专门的hmac包
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"novel-resource-management/database"
)

type UserCreditService struct {
	contract *client.Contract
}

// 新建一个service
func NewUserCreditService(gateway *client.Gateway) (*UserCreditService, error) {
	network := gateway.GetNetwork("mychannel")
	if network == nil {
		return nil, fmt.Errorf("userCredit network does not exist")
	}

	contract := network.GetContract("novel-basic")
	if contract == nil {
		return nil, fmt.Errorf("userCredit contract does not exist")
	}

	return &UserCreditService{
		contract: contract,
	}, nil
}

// create
func (us *UserCreditService) CreateUserCredit(userId string, credit int, totalUsed int, totalRecharge int) error {
	// 注意：链码层面已经包含了存在性检查，不需要在服务层重复检查
	// 移除服务层的ReadUserCredit调用，避免与链码的检查产生MVCC冲突

	// Gateway要求所有参数都是string类型，需要手动转换int参数
	_, err := us.contract.SubmitTransaction("CreateUserCredit", userId, strconv.Itoa(credit), strconv.Itoa(totalUsed), strconv.Itoa(totalRecharge))
	if err != nil {
		return fmt.Errorf("create user credit failed:%v", err)
	}
	return nil
}

// delete
func (us *UserCreditService) DeleteUserCredit(userId string) error {
	_, err := us.contract.SubmitTransaction("DeleteUserCredit", userId)
	if err != nil {
		return fmt.Errorf("delete user credit failed:%v", err)
	}
	return nil
}

// update
func (us *UserCreditService) UpdateUserCredit(userId string, credit int, totalUsed int, totalRecharge int) error {
	// Gateway要求所有参数都是string类型，需要手动转换int参数
	_, err := us.contract.SubmitTransaction("UpdateUserCredit", userId, strconv.Itoa(credit), strconv.Itoa(totalUsed), strconv.Itoa(totalRecharge))
	if err != nil {
		return fmt.Errorf("updateUserCreditFailed:%v", err)
	}
	return nil
}

// look up
func (us *UserCreditService) ReadUserCredit(userId string) (map[string]interface{}, error) {
	result, err := us.contract.EvaluateTransaction("ReadUserCredit", userId)
	if err != nil {
		return nil, fmt.Errorf("read user credit failed: %v", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	return data, nil
}

func (us *UserCreditService) GetAllUserCredits() ([]map[string]interface{}, error) {
	result, err := us.contract.EvaluateTransaction("GetAllUserCredits")
	if err != nil {
		return nil, fmt.Errorf("get all user credits failed: %v", err)
	}

	var data []map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	return data, nil
}

// ConsumeUserToken 消费用户token，每次调用减少一个token，直到减少到0
func (us *UserCreditService) ConsumeUserToken(userId string) error {
	// 先读取当前用户积分信息
	userCredit, err := us.ReadUserCredit(userId)
	if err != nil {
		return fmt.Errorf("读取用户积分失败: %v", err)
	}

	// 解析当前积分信息
	// ReadUserCredit方法返回的是map[string]interface{}类型，其中数值类型在JSON解析后会变成float64类型
	// 所以需要使用类型断言.(float64)先转换为float64，再转换为int类型
	// userCredit["credit"] 从map中获取credit字段的值
	credit := int(userCredit["credit"].(float64))
	totalUsed := int(userCredit["totalUsed"].(float64))
	totalRecharge := int(userCredit["totalRecharge"].(float64))

	// 检查token是否足够
	if credit <= 0 {
		return fmt.Errorf("用户 %s 的token不足，当前剩余: %d", userId, credit)
	}

	// 更新积分信息：减少1个token，增加已使用数量
	updatedCredit := credit - 1
	updatedTotalUsed := totalUsed + 1

	// 调用现有的UpdateUserCredit方法更新链上数据
	err = us.UpdateUserCredit(userId, updatedCredit, updatedTotalUsed, totalRecharge)
	if err != nil {
		return fmt.Errorf("更新用户积分失败: %v", err)
	}

	return nil
}

// AddTokensByEmail 通过邮箱给用户增加token
func (us *UserCreditService) AddTokensByEmail(email string, amount int) (string, int, error) {

	// 1. 从 MongoDB users 集合查询用户,获取 userId (即 users._id)
	mongoInstance := database.GetMongoInstance()
	usersCollection := mongoInstance.GetCollection("users")

	// 使用投影排除日期字段，避免类型转换问题
	// 这里的投影类似于子视图，只是为了查询需要的字段
	opts := options.FindOne().SetProjection(bson.M{
		"_id":       1,
		"email":     1,
		"username":  1,
		"novelIds":  1,
	})

	var user database.User
	err := usersCollection.FindOne(context.Background(), bson.M{"email": email}, opts).Decode(&user)
	if err != nil {
		return "", 0, fmt.Errorf("用户不存在: %s", email)
	}

	userId := user.ID
	log.Printf("✅ 找到用户: email=%s, userId=%s", email, userId)


	// 2. 读取当前用户积分信息
	userCredit, err := us.ReadUserCredit(userId)
	if err != nil {
		return userId, 0, fmt.Errorf("读取用户积分失败: %v", err)
	}

	// 3. 解析当前积分
	credit := int(userCredit["credit"].(float64))
	totalUsed := int(userCredit["totalUsed"].(float64))
	totalRecharge := int(userCredit["totalRecharge"].(float64))

	// 4. 计算新的积分
	newCredit := credit + amount
	newTotalRecharge := totalRecharge + amount

	// 5. 更新链码
	err = us.UpdateUserCredit(userId, newCredit, totalUsed, newTotalRecharge)
	if err != nil {
		return userId, 0, fmt.Errorf("更新链码失败: %v", err)
	}

	// 6. 同步更新 MongoDB user_credits 集合
	userCreditsCollection := mongoInstance.GetCollection("user_credits")

	//更新操作
	update := bson.M{
		"$set": bson.M{
			"credit":         newCredit,
			"totalRecharge":  newTotalRecharge,
			"updatedAt":      time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	//查询，这个有一个上下文，先filter，然后再有一个update方法
	_, err = userCreditsCollection.UpdateOne(
		context.Background(),
		bson.M{"userId": userId},
		update,
	)
	if err != nil {
		log.Printf("⚠️ MongoDB 更新失败: %v", err)
		// 不返回错误,因为链码已经更新成功
	} else {
		log.Printf("✅ MongoDB 同步更新成功")
	}

	log.Printf("✅ 充值成功: userId=%s, 增加token=%d, 新积分=%d", userId, amount, newCredit)

	return userId, newCredit, nil
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// P0-2: 幂等性支持 - 充值记录管理
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// RechargeRecord 充值记录（用于幂等性保证）
type RechargeRecord struct {
	ID          string    `bson:"_id" json:"id"`
	OrderSN     string    `bson:"orderSn" json:"orderSn"`       // 唯一索引
	UserID      string    `bson:"userId" json:"userId"`
	Email       string    `bson:"email" json:"email"`
	Amount      int       `bson:"amount" json:"amount"`         // 实际充值 token 数量
	ActualPrice int       `bson:"actualPrice" json:"actualPrice"` // 支付金额（分）
	Status      string    `bson:"status" json:"status"`         // pending, success, failed
	//time.Time
	ProcessedAt time.Time `bson:"processedAt" json:"processedAt"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

// findRechargeRecordByOrderSN 根据订单号查找充值记录
func (us *UserCreditService) findRechargeRecordByOrderSN(orderSN string) (*RechargeRecord, error) {
	mongoInstance := database.GetMongoInstance()
	collection := mongoInstance.GetCollection("recharge_records")

	//先定义，再赋值
	var record RechargeRecord
	//M是map的意思
	err := collection.FindOne(context.Background(), bson.M{"orderSn": orderSN}).Decode(&record)
	if err != nil {
		// 记录不存在
		return nil, nil
	}

	return &record, nil
}

// createRechargeRecord 创建充值记录
func (us *UserCreditService) createRechargeRecord(
	orderSN string,
	userID string,
	email string,
	amount int,
	actualPrice int,
	status string,
) error {
	mongoInstance := database.GetMongoInstance()
	collection := mongoInstance.GetCollection("recharge_records")

	now := time.Now()
	record := RechargeRecord{
		ID:          primitive.NewObjectID().Hex(),
		OrderSN:     orderSN,
		UserID:      userID,
		Email:       email,
		Amount:      amount,
		ActualPrice: actualPrice,
		Status:      status,
		ProcessedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := collection.InsertOne(context.Background(), record)
	if err != nil {
		return fmt.Errorf("创建充值记录失败: %v", err)
	}

	log.Printf("✅ 创建充值记录: orderSN=%s, status=%s", orderSN, status)
	return nil
}

// updateRechargeRecord 更新充值记录
func (us *UserCreditService) updateRechargeRecord(
	orderSN string,
	userID string,
	email string,
	amount int,
	actualPrice int,
	status string,
) error {
	mongoInstance := database.GetMongoInstance()
	collection := mongoInstance.GetCollection("recharge_records")

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"userId":      userID,
			"email":       email,
			"amount":      amount,
			"actualPrice": actualPrice,
			"status":      status,
			"processedAt": now,
			"updatedAt":   now,
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"orderSn": orderSN},
		update,
	)
	if err != nil {
		return fmt.Errorf("更新充值记录失败: %v", err)
	}

	log.Printf("✅ 更新充值记录: orderSN=%s, status=%s, amount=%d", orderSN, status, amount)
	return nil
}

// updateRechargeRecordStatus 仅更新充值记录状态
func (us *UserCreditService) updateRechargeRecordStatus(orderSN string, status string) error {
	mongoInstance := database.GetMongoInstance()
	collection := mongoInstance.GetCollection("recharge_records")

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": now,
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"orderSn": orderSN},
		update,
	)
	if err != nil {
		return fmt.Errorf("更新充值记录状态失败: %v", err)
	}

	log.Printf("✅ 更新充值记录状态: orderSN=%s, status=%s", orderSN, status)
	return nil
}

// AddTokensByEmailWithIdempotency 带幂等性保证的充值方法
func (us *UserCreditService) AddTokensByEmailWithIdempotency(
	email string,
	orderSN string,
	actualPrice int,
) (string, int, error) { //多值返回

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第1步：检查订单是否已处理（幂等性检查）
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	existingRecord, err := us.findRechargeRecordByOrderSN(orderSN)

	// 
	if err == nil && existingRecord != nil {
		
		if existingRecord.Status == "success" {
			// 幂等性保证：返回之前的结果
			return existingRecord.UserID, existingRecord.Amount, nil
		}

		if existingRecord.Status == "failed" {
			return "", 0, fmt.Errorf("订单之前处理失败，请人工介入: %s", orderSN)
		}

		if existingRecord.Status == "pending" {
			return "", 0, fmt.Errorf("订单正在处理中: %s", orderSN)
		}
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第2步：查询用户
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	mongoInstance := database.GetMongoInstance()
	usersCollection := mongoInstance.GetCollection("users")

	opts := options.FindOne().SetProjection(bson.M{
		"_id":      1,
		"email":    1,
		"username": 1,
		"novelIds": 1,
	})

	var user database.User
	err = usersCollection.FindOne(context.Background(), bson.M{"email": email}, opts).Decode(&user)
	if err != nil {
		// 创建失败记录
		us.createRechargeRecord(orderSN, "", email, 0, actualPrice, "failed")
		return "", 0, fmt.Errorf("用户不存在: %s", email)
	}

	userId := user.ID
	log.Printf("✅ 找到用户: email=%s, userId=%s", email, userId)

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第3步：创建充值记录（状态：pending）
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	err = us.createRechargeRecord(orderSN, userId, email, 0, actualPrice, "pending")
	if err != nil {
		// 可能是并发插入导致的重复订单
		existingRecord, _ := us.findRechargeRecordByOrderSN(orderSN)
		if existingRecord != nil && existingRecord.Status == "success" {
			log.Printf("⚠️ 并发处理：订单已被其他请求处理: orderSN=%s", orderSN)
			return existingRecord.UserID, existingRecord.Amount, nil
		}
		return "", 0, fmt.Errorf("创建充值记录失败: %v", err)
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第4步：计算充值金额（目前固定150，后续可配置套餐）
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	const rechargeAmount = 150

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第5步：读取当前用户积分
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	userCredit, err := us.ReadUserCredit(userId)
	if err != nil {
		us.updateRechargeRecordStatus(orderSN, "failed")
		return userId, 0, fmt.Errorf("读取用户积分失败: %v", err)
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第6步：计算新积分
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	credit := int(userCredit["credit"].(float64))
	totalUsed := int(userCredit["totalUsed"].(float64))
	totalRecharge := int(userCredit["totalRecharge"].(float64))

	newCredit := credit + rechargeAmount
	newTotalRecharge := totalRecharge + rechargeAmount

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第7步：更新链码
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	err = us.UpdateUserCredit(userId, newCredit, totalUsed, newTotalRecharge)
	if err != nil {
		us.updateRechargeRecordStatus(orderSN, "failed")
		return userId, 0, fmt.Errorf("更新链码失败: %v", err)
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第8步：同步更新 MongoDB
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	userCreditsCollection := mongoInstance.GetCollection("user_credits")
	update := bson.M{
		"$set": bson.M{
			"credit":        newCredit,
			"totalRecharge": newTotalRecharge,
			"updatedAt":     time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	_, err = userCreditsCollection.UpdateOne(
		context.Background(),
		bson.M{"userId": userId},
		update,
	)
	if err != nil {
		log.Printf("⚠️ MongoDB 同步失败: %v", err)
		// 不返回错误,因为链码已经更新成功
	} else {
		log.Printf("✅ MongoDB 同步更新成功")
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 第9步：更新充值记录为成功
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	us.updateRechargeRecord(orderSN, userId, email, rechargeAmount, actualPrice, "success")

	log.Printf("✅ 充值成功: userId=%s, orderSN=%s, amount=%d, newCredit=%d",
		userId, orderSN, rechargeAmount, newCredit)

	return userId, newCredit, nil
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// P0-1: HMAC 签名验证
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	// MAX_REQUEST_AGE 请求最大有效时间（5分钟）
	MAX_REQUEST_AGE = 5 * 60
)

// GetRechargeSecretKey 从环境变量获取充值接口的 HMAC 密钥
func GetRechargeSecretKey() string {
	key := os.Getenv("RECHARGE_SECRET_KEY")
	if key == "" {
		log.Printf("⚠️ 警告: RECHARGE_SECRET_KEY 环境变量未设置，使用默认值")
		key = "your-secret-key-change-in-production"
	}
	return key
}

// ComputeHMACSignature 计算 HMAC-SHA256 签名（导出函数）
func ComputeHMACSignature(params map[string]string, secretKey string) string {


	// 显示密钥摘要（不显示完整密钥）
	var keySummary string
	if len(secretKey) > 8 {
		keySummary = secretKey[:4] + "..." + secretKey[len(secretKey)-4:]
	} else {
		keySummary = "***"
	}


	// 步骤1: 按字母序排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 步骤2: 拼接参数
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	// strings主要是形状改变，而不是类型改变
	paramStr := strings.Join(parts, "&")

	// 步骤3: 计算 HMAC-SHA256
	// 创建带密钥的哈希计算器
	h := hmac.New(sha256.New, []byte(secretKey))
	// 将参数输入哈希器
	h.Write([]byte(paramStr))
	
	// 将哈希值转可传输字符串
	//  |------------|------------------------------------|
     // | h.Sum() 作用 | 哈希计算的"最终结算"，返回计算结果                 |
	 // | 参数 nil     | 表示"只返回哈希值，不追加任何数据"  
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// ValidateHMACSignature 验证 HMAC 签名（导出函数）
func ValidateHMACSignature(params map[string]string, receivedSignature string, secretKey string) bool {

	// 用同样的密钥和参数计算签名
	computedSignature := ComputeHMACSignature(params, secretKey)

	// 对比签名（使用 hmac.Equal 防止时序攻击）
	isValid := hmac.Equal([]byte(computedSignature), []byte(receivedSignature))

	if isValid {
		log.Printf("🔐 [ValidateHMACSignature] ✅ 签名验证通过")
	} else {
		// 详细对比签名
		if len(computedSignature) != len(receivedSignature) {
			log.Printf("🔐 [ValidateHMACSignature] ❌ 签名长度不匹配: 计算=%d, 接收=%d",
				len(computedSignature), len(receivedSignature))
		} else {
			// 逐个字符对比（仅显示前几个字符）
			maxChars := 10
			if len(computedSignature) > maxChars {
				log.Printf("🔐 [ValidateHMACSignature] ❌ 签名内容不匹配 (前%d个字符):", maxChars)
				log.Printf("🔐 [ValidateHMACSignature] ❌ 计算: %s", computedSignature[:maxChars])
				log.Printf("🔐 [ValidateHMACSignature] ❌ 接收: %s", receivedSignature[:maxChars])
			}
		}
	}

	return isValid
}

// ValidateTimestamp 验证时间戳（防重放攻击）（导出函数）
func ValidateTimestamp(timestamp int64) error {
	now := time.Now().Unix()
	age := now - timestamp


	if age < 0 {
		errMsg := fmt.Sprintf("请求时间戳来自未来，时间差=%d秒", age)
		log.Printf("❌ [ValidateTimestamp] %s", errMsg)
		return fmt.Errorf("请求时间戳来自未来")
	}

	if age > MAX_REQUEST_AGE {
		errMsg := fmt.Sprintf("请求过期，时间差=%d秒，超过阈值%d秒", age, MAX_REQUEST_AGE)
		log.Printf("❌ [ValidateTimestamp] %s", errMsg)
		return fmt.Errorf("请求过期，超过 %d 秒", MAX_REQUEST_AGE)
	}

	log.Printf("✅ [ValidateTimestamp] 时间戳验证通过: 时间差=%d秒 (在阈值%d秒内)", age, MAX_REQUEST_AGE)
	return nil
}
