package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Novel 结构体用于存储小说资源信息
type Novel struct {
	ID            string `json:"id"`
	Author        string `json:"author,omitempty"`
	StoryOutline  string `json:"storyOutline,omitempty"`
	Subsections   string `json:"subsections,omitempty"`
	Characters    string `json:"characters,omitempty"`
	Items         string `json:"items,omitempty"`
	TotalScenes   string `json:"totalScenes,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
}


type UserCredit struct {
	UserID      string `json:"userId"`
	Credit      int    `json:"credit"`
	TotalUsed   int    `json:"totalUsed"`
	TotalRecharge int `json:"totalRecharge"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// CreditHistory 结构体用于存储积分变更历史
type CreditHistory struct {
	UserID      string `json:"userId"`
	Amount      int    `json:"amount"` //积分变动的数额
	Type        string `json:"type"` // "consume", "recharge", "reward"
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
	NovelID     string `json:"novelId,omitempty"`
}

// CreateNovel creates a new novel in the world state
func (s *SmartContract) CreateNovel(ctx contractapi.TransactionContextInterface, id string, author string, storyOutline string, 
    subsections string, characters string, items string, totalScenes string) error {
	//judge whether novel is existed
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if novel exists: %v", err)
	}
	if exists {
		return fmt.Errorf("novel with ID %s already exists", id)
	}

	novel := Novel{
		ID:           id,
		Author:       author,
		StoryOutline: storyOutline,
		Subsections:  subsections,
		Characters:   characters,
		Items:        items,
		TotalScenes:  totalScenes,
		CreatedAt:    "",
	}

	novelJSON, err := json.Marshal(novel)
	if err != nil {
		return fmt.Errorf("failed to marshal novel: %v", err)
	}

	return ctx.GetStub().PutState(id, novelJSON)
}

//read
func (s *SmartContract) ReadNovel(ctx contractapi.TransactionContextInterface, id string)(*Novel ,error){
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read the novel failed:%v",err)
	}
	if(!exists){
		return nil, fmt.Errorf("I can not find the novel")
	}

	novelJSON, err := ctx.GetStub().GetState(id)

	if err != nil{
		return nil ,fmt.Errorf("the novel is not found:%v",err)
	}

	var novel Novel 
	//we can firstly fullfil a statement, get resource ,then we judge the
	//para1: the target need to be unmarshal 
	//para2: the variable that accept  the return data
	if err = json.Unmarshal(novelJSON,&novel); err != nil{
		return nil, fmt.Errorf("反序列化小说失败: %v", err)

	}

	return &novel, nil
}

//get all novels
func (s *SmartContract)GetAllNovels(ctx contractapi.TransactionContextInterface)([]*Novel,error){
   resultsIterator,err := ctx.GetStub().GetStateByRange("","")
   if err != nil{
	return nil, fmt.Errorf("获取数据失败:%v",err)
   }
   //iterate all,
   defer resultsIterator.Close()

   //注意已经声明过了
   var novels []*Novel
   
   for resultsIterator.HasNext(){
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var novel Novel
		
		//上边已经声明过了error，这里是复用
		err = json.Unmarshal(queryResponse.Value,&novel)

		if err != nil{
			return nil,err
		}

		novels =  append(novels, &novel)
   }
   return novels, nil
} 

// UpdateNovel updates an existing novel in the world state
func (s *SmartContract) UpdateNovel(ctx contractapi.TransactionContextInterface, id string, author string, storyOutline string, 
	subsections string, characters string, items string, totalScenes string) error {
	
	// Check if novel exists
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if novel exists: %v", err)
	}
	if !exists {
		return fmt.Errorf("novel with ID %s does not exist", id)
	}

	// Get existing novel to preserve CreatedAt
	existingNovel, err := s.ReadNovel(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read existing novel: %v", err)
	}

	// Create updated novel with preserved CreatedAt
	updatedNovel := Novel{
		ID:           id,
		Author:       author,
		StoryOutline: storyOutline,
		Subsections:  subsections,
		Characters:   characters,
		Items:        items,
		TotalScenes:  totalScenes,
		CreatedAt:    existingNovel.CreatedAt,
	}

	// Convert to JSON
	novelJSON, err := json.Marshal(updatedNovel)
	if err != nil {
		return fmt.Errorf("failed to marshal novel: %v", err)
	}

	// Save to world state，这个是需要key-value
	return ctx.GetStub().PutState(id, novelJSON)
}

//delete novel
func (s *SmartContract)DeleteNovel(ctx contractapi.TransactionContextInterface , id string) error{
	isExisting,err := s.NovelExists(ctx, id)
	if err != nil{
		return fmt.Errorf("failed to get novel:%v",err)
	}
	if(!isExisting){
		//todo, this object 
		return fmt.Errorf("the novel is not existed")
	}

	return ctx.GetStub().DelState(id)
}


func (s *SmartContract) NovelExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	novelJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return novelJSON != nil, nil
}


// 初始测试函数，一次性初始化多个小说对象
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) (string, error) {
	//能用值就用值
	novels := []Novel{
		{
			ID:           "test-novel-001",
			Author:       "测试作者1",
			StoryOutline: "这是第一个初始测试小说的大纲。",
			Subsections:  "第一章,第二章",
			Characters:   "主角A,配角B",
			Items:        "神秘宝物",
			TotalScenes:  "2",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:           "test-novel-002",
			Author:       "测试作者2",
			StoryOutline: "这是第二个初始测试小说的大纲。",
			Subsections:  "序章,终章",
			Characters:   "主角C,配角D",
			Items:        "古老卷轴",
			TotalScenes:  "2",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:           "test-novel-003",
			Author:       "测试作者3",
			StoryOutline: "这是第三个初始测试小说的大纲。",
			Subsections:  "开篇,高潮,结尾",
			Characters:   "主角E,配角F",
			Items:        "魔法石",
			TotalScenes:  "3",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	for _, novel := range novels {
		novelJSON, err := json.Marshal(novel)
		if err != nil {
			return "", fmt.Errorf("marshal 测试小说 %s 失败: %v", novel.ID, err)
		}
		err = ctx.GetStub().PutState(novel.ID, novelJSON)
		if err != nil {
			return "", fmt.Errorf("保存测试小说 %s 失败: %v", novel.ID, err)
		}
	}

	return "多个初始测试小说已成功写入区块链", nil
}


//增
func (s *SmartContract)CreateUserCredit(ctx contractapi.TransactionContextInterface, userId string , credit int, totalUsed int, totalRecharge int) error{

	exists, err := s.UserCreditExists(ctx, userId)
	if err != nil{
		//我采用最小错误包装
		return fmt.Errorf("judge exists failed:%v",err)
	}
	if exists{
		return fmt.Errorf("user credit with ID %s already exists", userId)
	}

	//获取当前时间
	currentTime := time.Now()
	//这里设置为这样，主要是因为时间戳格式
	currentTimeStr := currentTime.Format("2006-01-02 15:04:05")
	// timestamp := currentTime.Unix()      // 秒级时间戳
	// currentTimestamp := currentTime.UnixMilli() // 毫秒级时间戳

	userCredit := &UserCredit{
		UserID:userId,
		Credit:credit,
		TotalUsed:totalUsed,
		TotalRecharge:totalRecharge,
		CreatedAt:currentTimeStr,
		UpdatedAt:"",//如果要设定为nil，需要做字符串指针，因为指针类型 在 Go 中本质上是一个内存地址值，而 nil 表示空指针（即不指向任何内存地址）。
	}
	
	//这里默认取地址了,如果只有err可以直接=，然后重复利用声明的这个err
	userCreditJSON,err := json.Marshal(userCredit)
	if err != nil{
		return fmt.Errorf("marshal failed:%v",err)
	}
	// 是的，PutState 只会返回 error，如果没有错误就是存储成功，不需要返回其他内容。
	err =  ctx.GetStub().PutState(userId,userCreditJSON)

	if err != nil{
		return fmt.Errorf("put state failed:%v",err)
	}
	return nil
}

//删,
func (s *SmartContract)DeleteUserCredit(ctx contractapi.TransactionContextInterface, userId string)error{
	//先验证是否存在
	exists,err := s.UserCreditExists(ctx,userId)
	if err != nil{
		return fmt.Errorf("judge existing failed:%v",err)
	}
	if !exists{
		return fmt.Errorf("%s is not existed",userId)
	}

	//最后我们去删除	
	err = ctx.GetStub().DelState(userId)
	if err != nil{
		return fmt.Errorf("del failed:%v",err)
	}
	return nil
}

//改,
func (s *SmartContract)UpdateUserCredit(ctx contextapi.TransactionContextInterface,userId string , credit int, totalUsed int, totalRecharge string)(*UserCredit,error){
	//TODO,
}

//查,
func (s *SmartContract)ReadUserCredit(ctx contextapi.TransactionContextInterface,userId)(*UserCredit,error){
	//TODO,

}

//多个查
func (s *SmartContract)GetAllUserCredits(ctx contextapi.TransactionContextInterface)([]*UserCredit,error){
	//TODO,
}


//先添加辅助函数
func (s *SmartContract)UserCreditExists(ctx contractapi.TransactionContextInterface,userId string)(bool,error){
	userCreditJSON,err := ctx.GetStub().GetState(userId)
	if err != nil{
		return nil,err
	}
	return userCreditJSON != nil,nil
}

//TODO. implements some methods of token



