package service

import( 
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"strconv"
)

type UserCreditService struct {
	contract *client.Contract
}


//新建一个service
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
func CreateUserCredit(us *UserCreditService, userId string, credit int, totalUsed int, totalRecharge int) error {
	// 是的，Hyperledger Fabric Gateway 的 SubmitTransaction 方法会自动将参数转为字符串类型（即使你传入的是 int、float 等类型），
	// 它会调用 fmt.Sprint() 进行转换。所以你可以直接传 int、float、bool 等基础类型参数，SDK 会自动转为字符串传递给链码。
	_, err := us.contract.SubmitTransaction("CreateUserCredit", userId, strconv.Itoa(credit), strconv.Itoa(totalUsed), strconv.Itoa(totalRecharge))
	if err != nil{
		return fmt.Errorf("create user credit failed:%v",err)
	}	
	return nil
}

// delete
func DeleteUserCredit(us * UserCreditService,userId string)error{
	_, err := us.contract.SubmitTransaction("DeleteUserCredit",userId)
	if err != nil{
		return fmt.Errorf("delete user credit failed:%v",err)
	}
	return nil
}

// update 
func UpdateUserCredit(us *UserCreditService,userId string,credit int, totalUsed int, totalRecharge int)error{
	_,err := us.contract.EvaluateTransaction("UpdateUserCredit",userId, strconv.Itoa(credit), strconv.Itoa(totalUsed), strconv.Itoa(totalRecharge))
	if err != nil{
		return fmt.Errorf("updateUserCreditFailed:%v",err)
	}
	return nil
}

// look up
func ReadUserCredit(us *UserCreditService,userId string) (map[string]interface{},error){
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


func GetAllUserCredits(us *UserCreditService) ([]map[string]interface{}, error) {
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