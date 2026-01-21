//go:build test

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Novel 结构体用于测试
type Novel struct {
	ID           string `json:"id"`
	Author       string `json:"author"`
	StoryOutline string `json:"storyOutline"`
	Subsections  string `json:"subsections"`
	Characters   string `json:"characters"`
	Items        string `json:"items"`
	TotalScenes  string `json:"totalScenes"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// UserCredit 结构体用于测试
type UserCredit struct {
	UserID        string `json:"userId"`
	Credit        int    `json:"credit"`
	TotalUsed     int    `json:"totalUsed"`
	TotalRecharge int    `json:"totalRecharge"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// APIResponse 通用响应结构
type APIResponse struct {
	Message string                 `json:"message"`
	ID      string                 `json:"id"`
	Novel   map[string]interface{} `json:"novel"`
	Credit  map[string]interface{} `json:"credit"`
	Error   string                 `json:"error"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

const (
	BASE_URL = "http://localhost:8080"
	NOVEL_URL = BASE_URL + "/api/v1/novels"
	USER_URL = BASE_URL + "/api/v1/users"
	HEALTH_URL = BASE_URL + "/health"
)

func main() {
	fmt.Println("🚀 开始Go API测试...")
	fmt.Println("==================")

	// 测试健康检查
	testHealthCheck()

	// 测试小说CRUD
	testNovelCRUD()

	// 测试用户积分CRUD
	testUserCreditCRUD()

	fmt.Println("==================")
	fmt.Println("🏁 Go API测试完成")
}

func testHealthCheck() {
	fmt.Println("1️⃣  健康检查...")
	
	//1）先get，拿到resp，
	resp, err := http.Get(HEALTH_URL)
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	//记得resp的close
	defer resp.Body.Close()

	//2）读
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var health HealthResponse
	//3）挂载数据
	if err := json.Unmarshal(body, &health); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("健康检查响应: %+v\n", health)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ 健康检查通过")
	} else {
		fmt.Println("❌ 健康检查失败")
	}
	fmt.Println("")
}

func testNovelCRUD() {
	fmt.Println("2️⃣  获取所有小说...")
	resp, err := http.Get(NOVEL_URL)
	if err != nil {
		fmt.Printf("❌ 获取小说列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("3️⃣  创建新小说...")
	novel := Novel{
		ID:           "test_novel_go_001",
		Author:       "Go测试作者",
		StoryOutline: "这是Go测试创建的小说大纲",
		Subsections:  "第一章,第二章,第三章",
		Characters:   "主角,配角,反派",
		Items:        "魔法剑,神秘护符",
		TotalScenes:  "3",
	}

	novelJSON, err := json.Marshal(novel)
	if err != nil {
		fmt.Printf("❌ 序列化小说数据失败: %v\n", err)
		return
	}

	resp, err = http.Post(NOVEL_URL, "application/json", bytes.NewBuffer(novelJSON))
	if err != nil {
		fmt.Printf("❌ 创建小说失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("4️⃣  获取单个小说...")
	resp, err = http.Get(NOVEL_URL + "/test_novel_go_001")
	if err != nil {
		fmt.Printf("❌ 获取单个小说失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("5️⃣  更新小说...")
	novel.Author = "更新的Go测试作者"
	novel.StoryOutline = "这是更新后的Go测试小说大纲"
	novel.Subsections = "第一章,第二章,第三章,第四章"
	novel.TotalScenes = "4"

	novelJSON, _ = json.Marshal(novel)
	req, _ := http.NewRequest("PUT", NOVEL_URL+"/test_novel_go_001", bytes.NewBuffer(novelJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 更新小说失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")
}

func testUserCreditCRUD() {
	fmt.Println("6️⃣  获取所有用户积分...")
	resp, err := http.Get(USER_URL)
	if err != nil {
		fmt.Printf("❌ 获取用户积分列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("7️⃣  创建用户积分...")
	userCredit := UserCredit{
		UserID:        "test_user_go_001",
		Credit:        150,
		TotalUsed:     25,
		TotalRecharge: 175,
	}

	userCreditJSON, err := json.Marshal(userCredit)
	if err != nil {
		fmt.Printf("❌ 序列化用户积分数据失败: %v\n", err)
		return
	}

	resp, err = http.Post(USER_URL, "application/json", bytes.NewBuffer(userCreditJSON))
	if err != nil {
		fmt.Printf("❌ 创建用户积分失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("8️⃣  获取单个用户积分...")
	resp, err = http.Get(USER_URL + "/test_user_go_001")
	if err != nil {
		fmt.Printf("❌ 获取单个用户积分失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("9️⃣  更新用户积分...")
	userCredit.Credit = 200
	userCredit.TotalUsed = 50
	userCredit.TotalRecharge = 250

	userCreditJSON, _ = json.Marshal(userCredit)
	req, _ := http.NewRequest("PUT", USER_URL+"/test_user_go_001", bytes.NewBuffer(userCreditJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 更新用户积分失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	fmt.Println("")

	fmt.Println("🔟  清理测试数据...")
	
	// 删除测试小说
	req, _ = http.NewRequest("DELETE", NOVEL_URL+"/test_novel_go_001", nil)
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 删除测试小说失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 已删除测试小说: test_novel_go_001")
	} else {
		fmt.Printf("❌ 删除测试小说失败，状态码: %d\n", resp.StatusCode)
	}

	// 删除测试用户
	req, _ = http.NewRequest("DELETE", USER_URL+"/test_user_go_001", nil)
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 删除测试用户失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 已删除测试用户: test_user_go_001")
	} else {
		fmt.Printf("❌ 删除测试用户失败，状态码: %d\n", resp.StatusCode)
	}

	fmt.Println("🏁 测试数据清理完成")
}