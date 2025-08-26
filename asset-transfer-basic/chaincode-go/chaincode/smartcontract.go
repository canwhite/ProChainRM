package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

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

// InitLedger adds a base set of assets to the ledger
// ctx（contractapi.TransactionContextInterface）是Fabric链码中用于与账本交互的上下文对象，常用方法有：


// 1. GetStub()：获取ChaincodeStubInterface对象，进一步操作账本（如PutState、GetState、DelState等）。
// 2. GetClientIdentity()：获取ClientIdentity对象，可用于获取调用者的身份信息（如ID、MSPID、证书等）。
// 3. GetStub().GetState(key string)：根据key从账本中读取数据。
// 4. GetStub().PutState(key string, value []byte)：将数据写入账本。
// 5. GetStub().DelState(key string)：根据key删除账本中的数据。

// 6. GetStub().GetTxID()：获取当前交易的ID。
// 7. GetStub().GetChannelID()：获取当前通道ID。

// 8. GetStub().GetCreator()：获取交易发起者的证书信息。
// 9. GetStub().GetFunctionAndParameters()：获取调用的函数名和参数。


// 10. GetStub().SetEvent(name string, payload []byte)：设置链码事件。


// 11. GetStub().GetHistoryForKey(key string)：获取某个key的历史变更记录。
// 12. GetStub().GetStateByRange(startKey, endKey string)：按范围查询账本数据。
// 13. GetStub().GetQueryResult(query string)：执行富查询（CouchDB）。
// 14. GetStub().InvokeChaincode(chaincodeName string, args [][]byte, channel string)：调用其他链码。
// 15. GetStub().GetPrivateData(collection, key string)：获取私有数据。
// 16. GetStub().PutPrivateData(collection, key string, value []byte)：写入私有数据。
// 17. GetStub().DelPrivateData(collection, key string)：删除私有数据。
// 18. GetStub().GetTransient()：获取transient数据。
// 19. GetStub().GetBinding()：获取交易绑定信息。
// 20. GetStub().GetSignedProposal()：获取签名的提案。
// 这些方法可以满足大部分链码开发需求，具体可参考Fabric官方文档和API说明。

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	for _, asset := range assets {
		// json.Marshal的作用是将Go语言中的结构体（如Asset）序列化为JSON格式的字节切片（[]byte），
		// 这样可以方便地将数据存储到区块链的世界状态（World State）中，或者进行网络传输。
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
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

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset

	// unmarshal获得的是Asset结构体的数据
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	// 这样返回外部会自动解引用，岂不是很方便
	// Go 语言中返回结构体指针，外部调用时会自动解引用，非常方便
	// 例如：asset, err := contract.ReadAsset(ctx, "asset1")，可以直接使用 asset.ID 等字段
	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
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

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
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

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
