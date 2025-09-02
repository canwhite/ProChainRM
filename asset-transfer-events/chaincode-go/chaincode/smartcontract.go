package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

/*
总结：在链码开发中，SetEvent 主要适合在“增、删、改”操作时使用。

1. 增（Create/Insert）：如 CreateAsset、AddXXX 等函数。因为有新数据写入账本，适合发出事件通知监听方有新资源创建。
2. 删（Delete/Remove）：如 DeleteAsset、RemoveXXX 等函数。删除数据时，发出事件便于外部系统同步删除操作。
3. 改（Update/Modify）：如 UpdateAsset、TransferAsset 等函数。数据被修改时，发出事件便于监听方感知数据变更。

而“查（Read/Query）”操作一般不需要 SetEvent，因为它们只是读取数据，没有对账本状态产生变更。

最佳实践：在每个增删改函数成功写入账本（PutState/DelState）前后，调用 ctx.GetStub().SetEvent(eventName, payload) 发出事件，eventName 可用操作类型（如 "CreateAsset"、"DeleteAsset"、"UpdateAsset"），payload 可为相关数据的 JSON。

*/

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	AppraisedValue int    `json:"AppraisedValue"`
	Color          string `json:"Color"`
	ID             string `json:"ID"`
	Owner          string `json:"Owner"`
	Size           int    `json:"Size"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	//已经在判断是否存在了，我们把这部分内容提取出来
	existing, err := s.readState(ctx, id)
	if err == nil && existing != nil {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	//注意叫做SetEvent
	ctx.GetStub().SetEvent("CreateAsset", assetJSON)
	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) readState(ctx contractapi.TransactionContextInterface, id string) ([]byte, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	return assetJSON, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := s.readState(ctx, id)
	if err != nil {
		return nil, err
	}

	//Unmarshal需要将解析的值给到一个地址
	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	//可能会失败的地方都应该有err
	_, err := s.readState(ctx, id)
	if err != nil {
		return err
	}
	//短变量声明
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	//转化为json
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	ctx.GetStub().SetEvent("UpdateAsset", assetJSON)
	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	assetJSON, err := s.readState(ctx, id)
	if err != nil {
		return err
	}

	ctx.GetStub().SetEvent("DeleteAsset", assetJSON)
	return ctx.GetStub().DelState(id)
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := asset.Owner
	asset.Owner = newOwner

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	ctx.GetStub().SetEvent("TransferAsset", assetJSON)
	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

func (s *SmartContract) IsAssetExisting(ctx contractapi.TransactionContextInterface,id string)(bool,error){
	assetJSON , err = ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}


func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface)([]*Asset,error){
	
	resultsIterator,err := ctx.GetStub().GetStateByRange("","")
	if err != nil{
		return nil,err
	}
	defer resultsIterator.Close()	

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil{
			return nil ,err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value,&asset)
		if err != nil{
			return nil,err
		}
		//因为这里是野指针数组，记得填入指针
		assets = append(assets,&asset)

	}
	return assets,nil
}