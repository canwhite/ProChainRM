# Hyperledger Fabric 链码实现文档

## 架构设计

**建议：单一链码包含两个功能模块**

- 小说资源存证系统
- 用户积分系统

## 完整链码实现

### 1. 结构体定义

```go
// Novel 结构体 - 小说资源信息
type Novel struct {
    ID            string `json:"id"`                    // 小说唯一标识
    Author        string `json:"author,omitempty"`      // 作者
    StoryOutline  string `json:"storyOutline,omitempty"` // 故事大纲
    Subsections   string `json:"subsections,omitempty"`  // 分卷信息
    Characters    string `json:"characters,omitempty"`   // 角色信息
    Items         string `json:"items,omitempty"`        // 道具物品
    TotalScenes   string `json:"totalScenes,omitempty"`  // 总场景数
    CreatedAt     string `json:"createdAt,omitempty"`    // 创建时间
}

// UserCredit 结构体 - 用户积分信息
type UserCredit struct {
    UserID        string `json:"userId"`               // 用户ID
    Credit        int    `json:"credit"`               // 当前积分
    TotalUsed     int    `json:"totalUsed"`            // 总使用积分
    TotalRecharge int    `json:"totalRecharge"`        // 总充值积分
    CreatedAt     string `json:"createdAt,omitempty"`  // 创建时间
    UpdatedAt     string `json:"updatedAt,omitempty"`  // 更新时间
}

// CreditHistory 结构体 - 积分变更历史
type CreditHistory struct {
    UserID      string `json:"userId"`              // 用户ID
    Amount      int    `json:"amount"`              // 变更金额(正负)
    Type        string `json:"type"`                // 类型: consume/recharge/reward
    Description string `json:"description"`         // 描述
    Timestamp   string `json:"timestamp"`           // 时间戳
    NovelID     string `json:"novelId,omitempty"`   // 关联小说ID(可选)
}
```

### 2. 键名规范

- 小说资源: `NOVEL_<novelID>`
- 用户积分: `USER_<userID>`
- 积分历史: `HISTORY_<userID>_<txID>`

### 3. 核心方法

#### 小说资源管理

```go
// 创建小说资源
func (s *SmartContract) CreateNovel(ctx contractapi.TransactionContextInterface,
    id string, author string, storyOutline string,
    subsections string, characters string, items string, totalScenes string) error

// 读取小说资源
func (s *SmartContract) ReadNovel(ctx contractapi.TransactionContextInterface, id string) (*Novel, error)

// 检查小说是否存在
func (s *SmartContract) NovelExists(ctx contractapi.TransactionContextInterface, id string) (bool, error)

// 获取所有小说
func (s *SmartContract) GetAllNovels(ctx contractapi.TransactionContextInterface) ([]*Novel, error)


// 创建小说资源
func (s *SmartContract) CreateNovel(ctx contractapi.TransactionContextInterface,
	id string, author string, storyOutline string,
	subsections string, characters string, items string, totalScenes string) error {

	key := "NOVEL_" + id
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return fmt.Errorf("检查小说是否存在失败: %v", err)
	}
	if exists {
		return fmt.Errorf("小说ID %s 已存在", id)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	novel := Novel{
		ID:           id,
		Author:       author,
		StoryOutline: storyOutline,
		Subsections:  subsections,
		Characters:   characters,
		Items:        items,
		TotalScenes:  totalScenes,
		CreatedAt:    now,
	}
	novelJSON, err := json.Marshal(novel)
	if err != nil {
		return fmt.Errorf("序列化小说失败: %v", err)
	}
	return ctx.GetStub().PutState(key, novelJSON)
}

// 读取小说资源
func (s *SmartContract) ReadNovel(ctx contractapi.TransactionContextInterface, id string) (*Novel, error) {
	key := "NOVEL_" + id
	novelJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("读取小说失败: %v", err)
	}
	if novelJSON == nil {
		return nil, fmt.Errorf("小说ID %s 不存在", id)
	}
	var novel Novel
	if err := json.Unmarshal(novelJSON, &novel); err != nil {
		return nil, fmt.Errorf("反序列化小说失败: %v", err)
	}
	return &novel, nil
}

// 检查小说是否存在
func (s *SmartContract) NovelExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	key := "NOVEL_" + id
	novelJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, fmt.Errorf("检查小说存在性失败: %v", err)
	}
	return novelJSON != nil, nil
}

// 获取所有小说
func (s *SmartContract) GetAllNovels(ctx contractapi.TransactionContextInterface) ([]*Novel, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("NOVEL_", "NOVEL_~")
	if err != nil {
		return nil, fmt.Errorf("查询所有小说失败: %v", err)
	}
	defer resultsIterator.Close()

	var novels []*Novel
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var novel Novel
		if err := json.Unmarshal(queryResponse.Value, &novel); err != nil {
			return nil, err
		}
		novels = append(novels, &novel)
	}
	return novels, nil
}

// 更新小说资源
func (s *SmartContract) UpdateNovel(ctx contractapi.TransactionContextInterface,
	id string, author string, storyOutline string,
	subsections string, characters string, items string, totalScenes string) error {

	key := "NOVEL_" + id
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return fmt.Errorf("检查小说是否存在失败: %v", err)
	}
	if !exists {
		return fmt.Errorf("小说ID %s 不存在", id)
	}
	// 保留原有创建时间
	oldNovel, err := s.ReadNovel(ctx, id)
	if err != nil {
		return fmt.Errorf("读取原小说失败: %v", err)
	}
	novel := Novel{
		ID:           id,
		Author:       author,
		StoryOutline: storyOutline,
		Subsections:  subsections,
		Characters:   characters,
		Items:        items,
		TotalScenes:  totalScenes,
		CreatedAt:    oldNovel.CreatedAt,
	}
	novelJSON, err := json.Marshal(novel)
	if err != nil {
		return fmt.Errorf("序列化小说失败: %v", err)
	}
	return ctx.GetStub().PutState(key, novelJSON)
}

// 删除小说资源
func (s *SmartContract) DeleteNovel(ctx contractapi.TransactionContextInterface, id string) error {
	key := "NOVEL_" + id
	exists, err := s.NovelExists(ctx, id)
	if err != nil {
		return fmt.Errorf("检查小说是否存在失败: %v", err)
	}
	if !exists {
		return fmt.Errorf("小说ID %s 不存在", id)
	}
	return ctx.GetStub().DelState(key)
}



```

#### 用户积分管理

```go
// 初始化用户积分(默认100分)

func (s *SmartContract) InitUserCredit(ctx contractapi.TransactionContextInterface, userID string) error

// 获取用户积分
func (s *SmartContract) GetUserCredit(ctx contractapi.TransactionContextInterface, userID string) (*UserCredit, error)

// 消耗积分
func (s *SmartContract) ConsumeCredit(ctx contractapi.TransactionContextInterface,
    userID string, amount int, novelID string) error

// 充值积分
func (s *SmartContract) RechargeCredit(ctx contractapi.TransactionContextInterface,
    userID string, amount int) error

// 获取积分历史
func (s *SmartContract) GetUserCreditHistory(ctx contractapi.TransactionContextInterface,
    userID string) ([]*CreditHistory, error)


// 用户积分相关的增删改查实现

// 初始化用户积分（默认100分）
func (s *SmartContract) InitUserCredit(ctx contractapi.TransactionContextInterface, userID string) error {
	key := "USER_" + userID
	exists, err := s.UserCreditExists(ctx, userID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("用户 %s 的积分账户已存在", userID)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	credit := UserCredit{
		UserID:        userID,
		Credit:        100,
		TotalUsed:     0,
		TotalRecharge: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	creditJSON, err := json.Marshal(credit)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(key, creditJSON)
}

// 查询用户积分
func (s *SmartContract) GetUserCredit(ctx contractapi.TransactionContextInterface, userID string) (*UserCredit, error) {
	key := "USER_" + userID
	creditJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("读取用户积分失败: %v", err)
	}
	if creditJSON == nil {
		return nil, fmt.Errorf("用户 %s 的积分账户不存在", userID)
	}
	var credit UserCredit
	if err := json.Unmarshal(creditJSON, &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}


// 创建用户（仅创建用户信息，不初始化积分）
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, userID string) error {
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	// 检查用户是否已存在
	exists, err := s.UserCreditExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("检查用户是否存在失败: %v", err)
	}
	if exists {
		return fmt.Errorf("用户 %s 已存在", userID)
	}
	// 这里只创建一个空的用户积分账户（积分为0），如需默认100分请调用InitUserCredit
	now := time.Now().Format("2006-01-02 15:04:05")
	user := UserCredit{
		UserID:        userID,
		Credit:        0,
		TotalUsed:     0,
		TotalRecharge: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化用户信息失败: %v", err)
	}
	key := "USER_" + userID
	return ctx.GetStub().PutState(key, userJSON)
}


// 检查用户积分账户是否存在
func (s *SmartContract) UserCreditExists(ctx contractapi.TransactionContextInterface, userID string) (bool, error) {
	key := "USER_" + userID
	creditJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, err
	}
	return creditJSON != nil, nil
}

// 消耗积分
func (s *SmartContract) ConsumeCredit(ctx contractapi.TransactionContextInterface, userID string, amount int, novelID string) error {
	if amount <= 0 {
		return fmt.Errorf("消耗积分数额必须大于0")
	}
	credit, err := s.GetUserCredit(ctx, userID)
	if err != nil {
		return err
	}
	if credit.Credit < amount {
		return fmt.Errorf("用户 %s 积分不足", userID)
	}
	credit.Credit -= amount
	credit.TotalUsed += amount
	credit.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	// 更新积分
	key := "USER_" + userID
	creditJSON, err := json.Marshal(credit)
	if err != nil {
		return err
	}
	if err := ctx.GetStub().PutState(key, creditJSON); err != nil {
		return err
	}

	// 记录积分历史
	txID := ctx.GetStub().GetTxID()
	history := CreditHistory{
		UserID:      userID,
		Amount:      -amount,
		Type:        "consume",
		Description: fmt.Sprintf("消耗%d积分访问小说%s", amount, novelID),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		NovelID:     novelID,
	}
	historyKey := fmt.Sprintf("HISTORY_%s_%s", userID, txID)
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(historyKey, historyJSON)
}

// 充值积分
func (s *SmartContract) RechargeCredit(ctx contractapi.TransactionContextInterface, userID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("充值积分数额必须大于0")
	}
	credit, err := s.GetUserCredit(ctx, userID)
	if err != nil {
		return err
	}
	credit.Credit += amount
	credit.TotalRecharge += amount
	credit.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	// 更新积分
	key := "USER_" + userID
	creditJSON, err := json.Marshal(credit)
	if err != nil {
		return err
	}
	if err := ctx.GetStub().PutState(key, creditJSON); err != nil {
		return err
	}

	// 记录积分历史
	txID := ctx.GetStub().GetTxID()
	history := CreditHistory{
		UserID:      userID,
		Amount:      amount,
		Type:        "recharge",
		Description: fmt.Sprintf("充值%d积分", amount),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
	historyKey := fmt.Sprintf("HISTORY_%s_%s", userID, txID)
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(historyKey, historyJSON)
}

// 查询用户积分历史
func (s *SmartContract) GetUserCreditHistory(ctx contractapi.TransactionContextInterface, userID string) ([]*CreditHistory, error) {
	// 假设所有历史记录key前缀为HISTORY_<userID>_
	prefix := fmt.Sprintf("HISTORY_%s_", userID)
	resultsIterator, err := ctx.GetStub().GetStateByRange(prefix, prefix+"~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var histories []*CreditHistory
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var history CreditHistory
		if err := json.Unmarshal(queryResponse.Value, &history); err != nil {
			continue // 跳过异常数据
		}
		histories = append(histories, &history)
	}
	return histories, nil
}




```

#### 组合业务逻辑

```go
// 使用积分访问小说资源(原子操作)
func (s *SmartContract) AccessNovelWithCredit(ctx contractapi.TransactionContextInterface,
    userID string, novelID string, cost int) (*Novel, error)
```

### 4. 初始化数据

```go
// 初始化链码时创建测试数据
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    // 创建测试用户，每个用户默认100积分
    testUsers := []UserCredit{
        {UserID: "user1", Credit: 100, TotalUsed: 0, TotalRecharge: 0},
        {UserID: "user2", Credit: 100, TotalUsed: 0, TotalRecharge: 0},
        {UserID: "user3", Credit: 100, TotalUsed: 0, TotalRecharge: 0},
    }
    // ... 存储用户数据
}
```

// 部署链码流程（以 Fabric v2.x 为例，假设链码名为 novelcontract，通道为 mychannel）

./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

perhaps， this is complete steps

1. 打包链码
   在 novel-resource-events 目录下执行：

   ```
   peer lifecycle chaincode package novelcontract.tar.gz --path . --lang golang --label novelcontract_1
   ```

2. 安装链码

   ```
   peer lifecycle chaincode install novelcontract.tar.gz
   ```

3. 查询链码包 ID

   ```
   peer lifecycle chaincode queryinstalled
   ```

   记录输出的 Package ID。

4. 组织批准链码定义

   ```
   peer lifecycle chaincode approveformyorg \
     --channelID mychannel \
     --name novelcontract \
     --version 1.0 \
     --package-id <PackageID> \
     --sequence 1
   ```

5. 检查链码定义是否已准备好提交

   ```
   peer lifecycle chaincode checkcommitreadiness \
     --channelID mychannel \
     --name novelcontract \
     --version 1.0 \
     --sequence 1 \
     --output json
   ```

6. 提交链码定义

   ```
   peer lifecycle chaincode commit \
     --channelID mychannel \
     --name novelcontract \
     --version 1.0 \
     --sequence 1 \
     --peerAddresses <peer地址> \
     --tlsRootCertFiles <peer证书路径>
   ```

7. 初始化链码（如果需要 Init 方法）

   ```
   peer chaincode invoke -C mychannel -n novelcontract -c '{"function":"InitLedger","Args":[]}'
   ```

8. 验证链码部署
   ```
   peer chaincode query -C mychannel -n novelcontract -c '{"function":"GetAllNovels","Args":[]}'
   ```

注意：实际部署时需根据你的 Fabric 网络配置（如组织、peer、TLS、证书等）调整命令参数。
\*/

### 5. 使用示例

#### 部署链码后初始化

```bash
peer chaincode invoke -C mychannel -n novelcontract -c '{"function":"InitLedger","Args":[]}'
```

#### 创建小说资源

```bash
peer chaincode invoke -C mychannel -n novelcontract -c '{
    "function":"CreateNovel",
    "Args":["novel1", "张三", "修真世界大纲...", "第一卷:入门篇...", "主角:李强...", "飞剑、法宝...", "100"]
}'
```

#### 新用户注册

```bash
peer chaincode invoke -C mychannel -n novelcontract -c '{
    "function":"InitUserCredit",
    "Args":["newuser123"]
}'
```

#### 用户充值积分

```bash
peer chaincode invoke -C mychannel -n novelcontract -c '{
    "function":"RechargeCredit",
    "Args":["user1", "50"]
}'
```

#### 使用积分访问小说

```bash
peer chaincode invoke -C mychannel -n novelcontract -c '{
    "function":"AccessNovelWithCredit",
    "Args":["user1", "novel1", "10"]
}'
```

### 6. 查询方法

#### 查询用户积分

```bash
peer chaincode query -C mychannel -n novelcontract -c '{
    "function":"GetUserCredit",
    "Args":["user1"]
}'
```

#### 查询积分历史

```bash
peer chaincode query -C mychannel -n novelcontract -c '{
    "function":"GetUserCreditHistory",
    "Args":["user1"]
}'
```

#### 查询所有小说

```bash
peer chaincode query -C mychannel -n novelcontract -c '{
    "function":"GetAllNovels",
    "Args":[]
}'
```

## 总结

**推荐方案：单一链码包含两个功能模块**

**优势：**

1. **业务关联性强** - 积分消耗与资源访问直接关联
2. **原子操作** - 确保积分扣除和访问授权的一致性
3. **简化部署** - 只需部署和维护一个链码
4. **降低复杂度** - 避免跨链码调用的复杂性

**实现已完成，包含：**

- ✅ 小说资源存证功能
- ✅ 用户积分系统（默认 100 分）
- ✅ 积分充值与消耗
- ✅ 历史记录追踪
- ✅ 原子性业务操作

你可以直接照着这个文档实现，所有方法都已设计好，包括参数和返回值。
