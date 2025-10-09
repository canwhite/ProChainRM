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
