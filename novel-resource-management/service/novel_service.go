package service

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)


type NovelService struct {
	contract *client.Contract
}

func NewNovelService(gateway *client.Gateway) (*NovelService ,error){
	network := gateway.GetNetwork("mychannel")
	if network == nil{
		return nil, fmt.Errorf("无法获取network对象")

	}
	//先有network，再有contract
	contract := network.GetContract("basic")
	if contract == nil{
		return nil, fmt.Errorf("无法获取contract")
	}
	return &NovelService{contract: contract},nil 
}

//create novel
func (s *NovelService) CreateNovel(id, author, storyOutline, 
	subsections, characters, items, totalScenes string) error {

	fmt.Printf("Creating novel %s...\n", id)
	// 增删改操作需要使用SubmitTransaction，这里已经正确调用了SubmitTransaction方法
	_, err := s.contract.SubmitTransaction("CreateNovel",
		id, author, storyOutline, subsections, characters, items, totalScenes)
	if err != nil {
		return fmt.Errorf("failed to create novel %s: %w", id, err)
	}
	return nil
}

//update
func (s* NovelService) UpdateNovel(id, author, storyOutline, subsections, characters, items, totalScenes string) error {
	_,err := s.contract.SubmitTransaction("UpdateNovel",id, author, storyOutline, subsections, characters, items, totalScenes)
	if err != nil{
		return fmt.Errorf("failed to update novel %s: %w", id, err)
	}
	return nil
}      


//del 
func (s* NovelService)DeleteNovel(id string)error{
	_,err := s.contract.SubmitTransaction("DeleteNovel",id)
	if err != nil{
		return fmt.Errorf("failed to delete novel %s: %w",id, err)
	}
	return nil
}


//TODO，返回成map[string]interface{}
//ReadNovel 读取小说信息
func (s *NovelService)ReadNovel(id string)(string, error){
	fmt.Printf("Reading novel %s...\n", id)
	
	result, err := s.contract.EvaluateTransaction("ReadNovel", id)
	if err != nil {
		return "", fmt.Errorf("failed to read novel %s: %w", id, err)
	}
	
	return string(result), nil
}

//get all novels
func (s *NovelService) GetAllNovels() (string, error) {
	fmt.Println("Getting all novels...")
	
	result, err := s.contract.EvaluateTransaction("GetAllNovels")
	if err != nil {
		return "", fmt.Errorf("failed to get all novels: %w", err)
	}
	return string(result), nil
}

