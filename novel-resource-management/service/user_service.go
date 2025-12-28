package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"novel-resource-management/database"
	"strconv"
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
