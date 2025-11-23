package chaincode

import (
	"encoding/json"
	"fmt"
	"log" //主要
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Novel 结构体用于存储小说资源信息
type Novel struct {
	ID           string `json:"id"`
	Author       string `json:"author,omitempty"`
	StoryOutline string `json:"storyOutline,omitempty"`
	Subsections  string `json:"subsections,omitempty"`
	Characters   string `json:"characters,omitempty"`
	Items        string `json:"items,omitempty"`
	TotalScenes  string `json:"totalScenes,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

type UserCredit struct {
	UserID        string `json:"userId"`
	Credit        int    `json:"credit"`
	TotalUsed     int    `json:"totalUsed"`
	TotalRecharge int    `json:"totalRecharge"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// CreditHistory 结构体用于存储积分变更历史
type CreditHistory struct {
	UserID      string `json:"userId"`
	Amount      int    `json:"amount"` //积分变动的数额
	Type        string `json:"type"`   // "consume", "recharge", "reward"
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
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	novelJSON, err := json.Marshal(novel)
	if err != nil {
		return fmt.Errorf("failed to marshal novel: %v", err)
	}

	//setEvent
	ctx.GetStub().SetEvent("CreateNovel", novelJSON)
	return ctx.GetStub().PutState(id, novelJSON)
}

// read
func (s *SmartContract) ReadNovel(ctx contractapi.TransactionContextInterface, id string) (*Novel, error) {

	novelJSON, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("the novel is not found:%v", err)
	}

	if novelJSON == nil {
		return nil, fmt.Errorf("the novel is not found")
	}

	var novel Novel
	//we can firstly fullfil a statement, get resource ,then we judge the
	//para1: the target need to be unmarshal
	//para2: the variable that accept  the return data
	if err = json.Unmarshal(novelJSON, &novel); err != nil {
		return nil, fmt.Errorf("反序列化小说失败: %v", err)

	}

	return &novel, nil
}

// GetAllNovels returns all novels from the world state
func (s *SmartContract) GetAllNovels(ctx contractapi.TransactionContextInterface) ([]*Novel, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var novels []*Novel

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next: %v", err)
		}

		var novel Novel
		err = json.Unmarshal(queryResponse.Value, &novel)
		if err != nil {
			// Skip non-novel data
			continue
		}

		// Check if this is actually a novel by validating required fields
		if novel.ID != "" {
			// Ensure UpdatedAt is not empty for schema compliance
			if novel.UpdatedAt == "" {
				novel.UpdatedAt = novel.CreatedAt
			}
			novels = append(novels, &novel)
		}
	}
	return novels, nil
}


// UpdateNovel updates an existing novel in the world state
func (s *SmartContract) UpdateNovel(ctx contractapi.TransactionContextInterface, id string, author string, storyOutline string,
	subsections string, characters string, items string, totalScenes string) error {

	// 直接读取现有小说，一次性检查存在性和获取数据
	// 避免先检查存在性再读取的双重操作导致的MVCC冲突
	existingNovelJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read novel state: %v", err)
	}
	if existingNovelJSON == nil {
		return fmt.Errorf("novel with ID %s does not exist", id)
	}

	// 解析现有小说数据以保留 CreatedAt
	var existingNovel Novel
	err = json.Unmarshal(existingNovelJSON, &existingNovel)
	if err != nil {
		return fmt.Errorf("failed to unmarshal existing novel: %v", err)
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
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	// Convert to JSON
	novelJSON, err := json.Marshal(updatedNovel)
	if err != nil {
		return fmt.Errorf("failed to marshal novel: %v", err)
	}
	//setEvent
	ctx.GetStub().SetEvent("UpdateNovel", novelJSON)
	// Save to world state，这个是需要key-value
	return ctx.GetStub().PutState(id, novelJSON)
}

// delete novel
func (s *SmartContract) DeleteNovel(ctx contractapi.TransactionContextInterface, id string) error {
	// isExisting,err := s.NovelExists(ctx, id)
	novelJSON, err := s.ReadNovel(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get novel:%v", err)
	}
	if novelJSON == nil {
		return fmt.Errorf("the novel is not found")
	}
	//setEvent
	novelJSONBytes, err := json.Marshal(novelJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal novel for event: %v", err)
	}
	ctx.GetStub().SetEvent("DeleteNovel", novelJSONBytes)
	//只返回了error
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
	//设置前缀
	novels := []Novel{
		{
			ID:           "novel_001",
			Author:       "测试作者1",
			StoryOutline: "这是第一个初始测试小说的大纲。",
			Subsections:  "第一章,第二章",
			Characters:   "主角A,配角B",
			Items:        "神秘宝物",
			TotalScenes:  "2",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:           "novel_002",
			Author:       "测试作者2",
			StoryOutline: "这是第二个初始测试小说的大纲。",
			Subsections:  "序章,终章",
			Characters:   "主角C,配角D",
			Items:        "古老卷轴",
			TotalScenes:  "2",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:           "novel_003",
			Author:       "测试作者3",
			StoryOutline: "这是第三个初始测试小说的大纲。",
			Subsections:  "开篇,高潮,结尾",
			Characters:   "主角E,配角F",
			Items:        "魔法石",
			TotalScenes:  "3",
			CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
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

	//设置前缀
	usercredits := []UserCredit{
		{
			UserID:        "usercredit_001",
			Credit:        100,
			TotalUsed:     0,
			TotalRecharge: 0,
			CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:     time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			UserID:        "usercredit_002",
			Credit:        200,
			TotalUsed:     0,
			TotalRecharge: 0,
			CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:     time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	for _, userCredit := range usercredits {
		//marshal
		userCreditJSON, err := json.Marshal(userCredit)
		if err != nil {
			return "", fmt.Errorf("marshal 测试用户信用 %s 失败: %v", userCredit.UserID, err)
		}
		err = ctx.GetStub().PutState(userCredit.UserID, userCreditJSON)
	}

	return "多个初始测试小说已成功写入区块链", nil
}

// 增
func (s *SmartContract) CreateUserCredit(ctx contractapi.TransactionContextInterface, userId string, credit int, totalUsed int, totalRecharge int) error {

	exists, err := s.UserCreditExists(ctx, userId)
	if err != nil {
		//我采用最小错误包装
		return fmt.Errorf("judge exists failed:%v", err)
	}
	if exists {
		return fmt.Errorf("user credit with ID %s already exists", userId)
	}

	//获取当前时间
	currentTime := time.Now()
	//这里设置为这样，主要是因为时间戳格式
	currentTimeStr := currentTime.Format("2006-01-02 15:04:05")
	// timestamp := currentTime.Unix()      // 秒级时间戳
	// currentTimestamp := currentTime.UnixMilli() // 毫秒级时间戳

	userCredit := &UserCredit{
		UserID:        userId,
		Credit:        credit,
		TotalUsed:     totalUsed,
		TotalRecharge: totalRecharge,
		CreatedAt:     currentTimeStr,
		UpdatedAt:     currentTimeStr, // Set UpdatedAt same as CreatedAt for new records
	}

	//这里默认取地址了,如果只有err可以直接=，然后重复利用声明的这个err
	userCreditJSON, err := json.Marshal(userCredit)
	if err != nil {
		return fmt.Errorf("marshal failed:%v", err)
	}
	// 是的，PutState 只会返回 error，如果没有错误就是存储成功，不需要返回其他内容。
	err = ctx.GetStub().PutState(userId, userCreditJSON)

	if err != nil {
		return fmt.Errorf("put state failed:%v", err)
	}
	//setEvent
	ctx.GetStub().SetEvent("CreateUserCredit", userCreditJSON)

	return nil
}

// 删,
func (s *SmartContract) DeleteUserCredit(ctx contractapi.TransactionContextInterface, userId string) error {
	//先验证是否存在
	// 先通过ReadUserCredit方法读取，再判断
	userCreditJSON, err := s.ReadUserCredit(ctx, userId)
	if err != nil {
		return fmt.Errorf("读取用户积分信息失败: %v", err)
	}
	if userCreditJSON == nil {
		return fmt.Errorf("用户 %s 不存在", userId)
	}

	//最后我们去删除
	err = ctx.GetStub().DelState(userId)
	if err != nil {
		return fmt.Errorf("del failed:%v", err)
	}

	//setEvent
	userCreditJSONBytes, err := json.Marshal(userCreditJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal user credit for event: %v", err)
	}
	ctx.GetStub().SetEvent("DeleteUserCredit", userCreditJSONBytes)
	return nil
}

// 改,
func (s *SmartContract) UpdateUserCredit(ctx contractapi.TransactionContextInterface, userId string, credit int, totalUsed int, totalRecharge int) error {
	existingUserCredit, err := s.ReadUserCredit(ctx, userId)
	if err != nil {
		return fmt.Errorf("read failed:%v", err)
	}
	if existingUserCredit == nil {
		return fmt.Errorf("%s is not existed", userId)
	}

	// 是的，这里相当于声明并初始化了一个UserCredit指针，updatedUserCredit 指向了一个新的 UserCredit 结构体实例，并且字段已经被赋值。
	updatedUserCredit := &UserCredit{
		//用原来的UserId，UserID不变
		UserID:        existingUserCredit.UserID,
		Credit:        credit,
		TotalUsed:     totalUsed,
		TotalRecharge: totalRecharge,
		CreatedAt:     existingUserCredit.CreatedAt,
		UpdatedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}

	//更新，还是需要和create的时候保持一致，marshal转化为json，再putState
	updatedUserCreditJSON, err := json.Marshal(updatedUserCredit)
	if err != nil {
		return fmt.Errorf("marshal failed:%v", err)
	}

	//setEvent
	ctx.GetStub().SetEvent("UpdateUserCredit", updatedUserCreditJSON)
	err = ctx.GetStub().PutState(userId, updatedUserCreditJSON)
	if err != nil {
		return fmt.Errorf("put state failed:%v", err)
	}
	return nil
}

// 查,
func (s *SmartContract) ReadUserCredit(ctx contractapi.TransactionContextInterface, userId string) (*UserCredit, error) {
	//直接获取
	userCreditJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return nil, fmt.Errorf("read failed:%v", err)
	}
	if userCreditJSON == nil {
		return nil, fmt.Errorf("%s is not existed", userId)
	}
	var userCredit UserCredit
	//用指针做操作最重要的作用是为了写
	err = json.Unmarshal(userCreditJSON, &userCredit)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed:%v", err)
	}
	//因为返回值定义的是指针，所以可以直接返回指针，使用的时候也很方便，可以直接用，因为可以自动解引用
	return &userCredit, nil
}

// 多个查
func (s *SmartContract) GetAllUserCredits(ctx contractapi.TransactionContextInterface) ([]*UserCredit, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("get state by range failed:%v", err)
	}

	defer resultsIterator.Close()

	var userCredits []*UserCredit

	//因为先判断了HasNext，所以我们可以直接从Next中取值
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("get next failed:%v", err)
		}

		// 过滤出UserCredit类型的数据
		var userCredit UserCredit
		err = json.Unmarshal(queryResponse.Value, &userCredit)
		if err != nil {
			// 如果不是UserCredit类型，跳过
			continue
		}

		// 验证UserCredit结构体字段
		if userCredit.UserID == "" {
			// 如果不是有效的UserCredit，跳过
			continue
		}

		// Ensure UpdatedAt is not empty for schema compliance
		if userCredit.UpdatedAt == "" {
			userCredit.UpdatedAt = userCredit.CreatedAt
		}

		userCredits = append(userCredits, &userCredit)
	}

	//确保没有nil
	return userCredits, nil
}

// 先添加辅助函数
func (s *SmartContract) UserCreditExists(ctx contractapi.TransactionContextInterface, userId string) (bool, error) {
	userCreditJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return false, err
	}
	return userCreditJSON != nil, nil
}

// MongoImportData 从 MongoDB 导入的数据结构
type MongoImportData struct {
	Novels      []Novel      `json:"novels"`
	UserCredits []UserCredit `json:"userCredits"`
}

// InitFromMongoDB 从 MongoDB 数据初始化账本
// 参数：JSON字符串，包含从 MongoDB 读取的所有数据
func (s *SmartContract) InitFromMongoDB(ctx contractapi.TransactionContextInterface, jsonData string) (string, error) {

	// 解析 JSON 数据
	var importData MongoImportData
	// de stringify 
	if err := json.Unmarshal([]byte(jsonData), &importData); err != nil {
		return "", fmt.Errorf("解析 MongoDB 数据失败: %v", err)
	}

	// 导入 novels 数据
	novelSuccessCount := 0
	novelErrorCount := 0
	for _, novel := range importData.Novels {
		// 检查小说是否已存在，如果存在则跳过（MongoDB 数据优先）
		exists, err := s.NovelExists(ctx, novel.ID)
		if err != nil {
			log.Printf("⚠️ 检查小说 %s 存在性失败: %v", novel.ID, err)
			novelErrorCount++
			continue
		}

		if exists {
			log.Printf("⏭️ 小说 %s 已存在，跳过（MongoDB 数据优先）", novel.ID)
			continue
		}

		// 创建小说
		novelJSON, err := json.Marshal(novel)
		if err != nil {
			log.Printf("⚠️ 序列化小说 %s 失败: %v", novel.ID, err)
			novelErrorCount++
			continue
		}

		if err := ctx.GetStub().PutState(novel.ID, novelJSON); err != nil {
			log.Printf("⚠️ 保存小说 %s 失败: %v", novel.ID, err)
			novelErrorCount++
			continue
		}

		log.Printf("✅ 成功导入小说: %s - %s", novel.ID, novel.Author)
		novelSuccessCount++
	}

	// 导入 userCredits 数据
	creditSuccessCount := 0
	creditErrorCount := 0
	for _, userCredit := range importData.UserCredits {
		// 检查用户积分是否已存在，如果存在则跳过（MongoDB 数据优先）
		exists, err := s.UserCreditExists(ctx, userCredit.UserID)
		if err != nil {
			log.Printf("⚠️ 检查用户积分 %s 存在性失败: %v", userCredit.UserID, err)
			creditErrorCount++
			continue
		}

		if exists {
			log.Printf("⏭️ 用户积分 %s 已存在，跳过（MongoDB 数据优先）", userCredit.UserID)
			continue
		}

		// 创建用户积分
		userCreditJSON, err := json.Marshal(userCredit)
		if err != nil {
			log.Printf("⚠️ 序列化用户积分 %s 失败: %v", userCredit.UserID, err)
			creditErrorCount++
			continue
		}

		if err := ctx.GetStub().PutState(userCredit.UserID, userCreditJSON); err != nil {
			log.Printf("⚠️ 保存用户积分 %s 失败: %v", userCredit.UserID, err)
			creditErrorCount++
			continue
		}

		log.Printf("✅ 成功导入用户积分: %s - credit:%d", userCredit.UserID, userCredit.Credit)
		creditSuccessCount++
	}

	result := fmt.Sprintf("🎉 MongoDB 数据导入完成!\n"+
		"📚 小说数据: 成功 %d 个, 失败 %d 个\n"+
		"💰 用户积分数据: 成功 %d 个, 失败 %d 个\n"+
		"📊 总计: 成功 %d 个, 失败 %d 个",
		novelSuccessCount, novelErrorCount,
		creditSuccessCount, creditErrorCount,
		novelSuccessCount+creditSuccessCount, novelErrorCount+creditErrorCount)

	log.Println(result)
	return result, nil
}

//TODO. implements some methods of token
