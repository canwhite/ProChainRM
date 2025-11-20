package service

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
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
b

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
