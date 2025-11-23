package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type NovelService struct {
	contract *client.Contract
}

func NewNovelService(gateway *client.Gateway) (*NovelService, error) {
	network := gateway.GetNetwork("mychannel")
	if network == nil {
		return nil, fmt.Errorf("无法获取network对象")

	}
	//先有network，再有contract
	contract := network.GetContract("novel-basic")
	if contract == nil {
		return nil, fmt.Errorf("无法获取contract")
	}
	return &NovelService{contract: contract}, nil
}

// create novel
func (s *NovelService) CreateNovel(id, author, storyOutline,
	subsections, characters, items, totalScenes string) error {

	fmt.Printf("Creating novel %s...\n", id)

	// 增删改操作需要使用SubmitTransaction，这里已经正确调用了SubmitTransaction方法
	// 注意：链码层面已经包含了存在性检查，不需要在服务层重复检查
	_, err := s.contract.SubmitTransaction("CreateNovel",
		id, author, storyOutline, subsections, characters, items, totalScenes)
	if err != nil {
		return fmt.Errorf("failed to create novel %s: %w", id, err)
	}
	return nil
}

// update
func (s *NovelService) UpdateNovel(id, author, storyOutline, subsections, characters, items, totalScenes string) error {
	_, err := s.contract.SubmitTransaction("UpdateNovel", id, author, storyOutline, subsections, characters, items, totalScenes)
	if err != nil {
		return fmt.Errorf("failed to update novel %s: %w", id, err)
	}
	return nil
}

// del
func (s *NovelService) DeleteNovel(id string) error {
	_, err := s.contract.SubmitTransaction("DeleteNovel", id)
	if err != nil {
		return fmt.Errorf("failed to delete novel %s: %w", id, err)
	}
	return nil
}

// ReadNovel 读取小说信息
func (s *NovelService) ReadNovel(id string) (map[string]interface{}, error) {
	fmt.Printf("Reading novel %s...\n", id)

	result, err := s.contract.EvaluateTransaction("ReadNovel", id)

	if err != nil {
		return nil, fmt.Errorf("failed to read novel %s: %w", id, err)
	}

	//map[string]interface{}
	var novelData map[string]interface{}

	err = json.Unmarshal(result, &novelData)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed:%v", err)
	}

	return novelData, nil
}

// get all novels
func (s *NovelService) GetAllNovels() ([]map[string]interface{}, error) {
	fmt.Println("Getting all novels...")

	result, err := s.contract.EvaluateTransaction("GetAllNovels")

	if err != nil {
		return nil, fmt.Errorf("failed to get all novels: %w", err)
	}

	// 🔍 添加调试信息
	fmt.Printf("🔍 [DEBUG] 链码返回原始数据长度: %d\n", len(result))
	fmt.Printf("🔍 [DEBUG] 链码返回原始数据内容: %q\n", string(result))

	// 检查是否为空数据
	if len(result) == 0 {
		fmt.Printf("⚠️ [WARNING] 链码返回空数据\n")
		return []map[string]interface{}{}, nil
	}

	// 检查是否是有效的JSON开头
	trimmedResult := strings.TrimSpace(string(result))
	if !strings.HasPrefix(trimmedResult, "[") {
		fmt.Printf("❌ [ERROR] 链码返回的不是有效的JSON数组格式，开头是: %q\n", trimmedResult[:10])
		return nil, fmt.Errorf("invalid JSON format, expected array but got: %s", trimmedResult[:min(50, len(trimmedResult))])
	}

	var novels []map[string]interface{}

	err = json.Unmarshal(result, &novels)

	if err != nil {
		fmt.Printf("❌ [ERROR] JSON 解析失败: %v\n", err)
		fmt.Printf("❌ [ERROR] 尝试解析的数据内容: %q\n", string(result))
		return nil, fmt.Errorf("unmarshal failed:%w", err)
	}

	fmt.Printf("✅ [SUCCESS] 解析成功，获取到 %d 个小说\n", len(novels))
	return novels, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
