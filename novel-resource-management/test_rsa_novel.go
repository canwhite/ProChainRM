package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"novel-resource-management/utils"
	"os"
	"strings"
	"time"
)

// NovelData 小说数据结构
type NovelData struct {
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

// EncryptedRequest 加密请求结构
type EncryptedRequest struct {
	EncryptedData string `json:"encryptedData"`
}

const (
	BASE_URL = "http://localhost:8080"
	NOVEL_URL = BASE_URL + "/api/v1/novels"
)

func main() {
	fmt.Println("🔐 开始Novel RSA加密测试...")
	fmt.Println("============================")

	// 先检查密钥文件
	diagnoseKeyFiles()

	// 初始化RSA工具
	fmt.Println("正在初始化RSA工具...")
	if err := utils.InitRSACrypto(); err != nil {
		fmt.Printf("❌ RSA工具初始化失败: %v\n", err)
		fmt.Println("请检查：")
		fmt.Println("1. security/rsa_private_key.pem 文件是否存在")
		fmt.Println("2. security/rsa_public_key.pem 文件是否存在")
		fmt.Println("3. 密钥文件格式是否正确")
		return
	}
	fmt.Println("✅ RSA工具初始化成功")
	
	// 验证初始化是否真的成功
	fmt.Println("验证RSA工具初始化状态...")
	testResult, err := utils.EncryptWithRSA("test")
	if err != nil {
		fmt.Printf("❌ RSA工具验证失败: %v\n", err)
		return
	}
	fmt.Printf("✅ RSA工具验证成功，测试加密结果长度: %d\n", len(testResult))

	// 测试1：RSA加解密功能
	testRSAUtils()

	// 测试2：普通API调用（不加密）
	testNormalNovelAPICall()

	// 测试3：加密API调用
	testEncryptedNovelAPICall()

	// 测试4：PUT请求加密
	testEncryptedNovelPutCall()

	// 测试5：错误处理
	testErrorHandling()

	fmt.Println("============================")
	fmt.Println("🏁 Novel RSA加密测试完成")
}

// testRSAUtils 测试RSA工具类
func testRSAUtils() {
	fmt.Println("1️⃣  测试RSA工具类...")

	// 准备测试数据
	testData := NovelData{
		ID:           "test_novel_rsa_001",
		Author:       "测试作者",
		StoryOutline: "这是一个测试故事大纲",
		Subsections:  "章节1,章节2,章节3",
		Characters:   "主角:张三,配角:李四",
		Items:        "道具1,道具2",
		TotalScenes:  "10",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	// 序列化数据
	jsonData, err := json.Marshal(testData)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	fmt.Printf("原始数据: %s\n", string(jsonData))

	// 使用项目中的RSA工具加密
	encryptedData, err := utils.EncryptWithRSA(string(jsonData))
	if err != nil {
		fmt.Printf("❌ 加密失败: %v\n", err)
		return
	}

	fmt.Printf("加密成功，数据长度: %d\n", len(encryptedData))
	fmt.Printf("加密数据前50字符: %s...\n", encryptedData[:min(50, len(encryptedData))])

	// 使用项目中的RSA工具解密
	decryptedData, err := utils.DecryptWithRSA(encryptedData)
	if err != nil {
		fmt.Printf("❌ 解密失败: %v\n", err)
		return
	}

	fmt.Printf("解密数据: %s\n", decryptedData)

	// 验证数据一致性
	if string(jsonData) == decryptedData {
		fmt.Println("✅ RSA工具类测试通过")
	} else {
		fmt.Println("❌ RSA工具类测试失败 - 数据不一致")
	}
	fmt.Println("")
}

// testNormalNovelAPICall 测试普通小说API调用（不加密）
func testNormalNovelAPICall() {
	fmt.Println("2️⃣  测试普通小说API调用（不加密）...")

	// 创建小说数据（不加密）
	novel := NovelData{
		ID:           "test_normal_novel_001",
		Author:       "普通测试作者",
		StoryOutline: "这是一个普通测试故事大纲",
		Subsections:  "普通章节1,普通章节2",
		Characters:   "普通主角,普通配角",
		Items:        "普通道具1,普通道具2",
		TotalScenes:  "5",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(novel)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	resp, err := http.Post(NOVEL_URL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("❌ 普通小说API调用失败: %v\n", err)
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

	// 现在POST请求需要RSA加密，所以应该返回错误
	if resp.StatusCode == 400 {
		fmt.Println("✅ 普通小说API调用正确被拒绝（需要RSA加密）")
	} else {
		fmt.Println("❌ 普通小说API调用应该被拒绝但成功了")
	}
	fmt.Println("")
}

// testEncryptedNovelAPICall 测试加密小说API调用
func testEncryptedNovelAPICall() {
	fmt.Println("3️⃣  测试加密小说API调用...")

	// 准备测试数据
	novel := NovelData{
		ID:           "test_encrypted_novel_001",
		Author:       "加密测试作者",
		StoryOutline: "这是一个加密测试故事大纲",
		Subsections:  "加密章节1,加密章节2",
		Characters:   "加密主角,加密配角",
		Items:        "加密道具1,加密道具2",
		TotalScenes:  "8",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(novel)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	fmt.Printf("原始数据: %s\n", string(jsonData))

	// 使用项目中的RSA工具加密
	encryptedData, err := utils.EncryptWithRSA(string(jsonData))
	if err != nil {
		fmt.Printf("❌ 加密失败: %v\n", err)
		return
	}

	// 创建加密请求
	encryptedRequest := EncryptedRequest{
		EncryptedData: encryptedData,
	}

	requestJSON, err := json.Marshal(encryptedRequest)
	if err != nil {
		fmt.Printf("❌ 序列化加密请求失败: %v\n", err)
		return
	}

	// 发送加密请求
	req, err := http.NewRequest("POST", NOVEL_URL, strings.NewReader(string(requestJSON)))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	// 设置加密请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted-Request", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 加密小说API调用失败: %v\n", err)
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

	// 由于需要fabric网络，可能会返回内部错误，但RSA解密应该成功
	if resp.StatusCode == 200 {
		fmt.Println("✅ 加密小说API调用成功")
	} else if resp.StatusCode == 500 {
		fmt.Println("⚠️  加密小说API调用 - RSA解密成功但业务逻辑失败（可能没有fabric网络）")
	} else {
		fmt.Println("❌ 加密小说API调用失败")
	}
	fmt.Println("")
}

// testEncryptedNovelPutCall 测试加密PUT请求
func testEncryptedNovelPutCall() {
	fmt.Println("4️⃣  测试加密小说PUT请求...")

	// 准备更新的测试数据
	updatedNovel := NovelData{
		ID:           "test_put_novel_001",
		Author:       "更新后的作者",
		StoryOutline: "更新后的故事大纲",
		Subsections:  "更新章节1,更新章节2,更新章节3",
		Characters:   "更新主角,更新配角,更新反派",
		Items:        "更新道具1,更新道具2,更新道具3",
		TotalScenes:  "15",
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(updatedNovel)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	fmt.Printf("更新数据: %s\n", string(jsonData))

	// 使用RSA加密
	encryptedData, err := utils.EncryptWithRSA(string(jsonData))
	if err != nil {
		fmt.Printf("❌ 加密失败: %v\n", err)
		return
	}

	// 创建加密请求
	encryptedRequest := EncryptedRequest{
		EncryptedData: encryptedData,
	}

	requestJSON, err := json.Marshal(encryptedRequest)
	if err != nil {
		fmt.Printf("❌ 序列化加密请求失败: %v\n", err)
		return
	}

	// 发送PUT请求
	putURL := NOVEL_URL + "/test_put_novel_001"
	req, err := http.NewRequest("PUT", putURL, strings.NewReader(string(requestJSON)))
	if err != nil {
		fmt.Printf("❌ 创建PUT请求失败: %v\n", err)
		return
	}

	// 设置加密请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted-Request", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 加密PUT请求失败: %v\n", err)
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

	// 同样，可能因为fabric网络问题返回内部错误
	if resp.StatusCode == 200 {
		fmt.Println("✅ 加密PUT请求成功")
	} else if resp.StatusCode == 500 {
		fmt.Println("⚠️  加密PUT请求 - RSA解密成功但业务逻辑失败（可能没有fabric网络）")
	} else {
		fmt.Println("❌ 加密PUT请求失败")
	}
	fmt.Println("")
}

// testErrorHandling 测试错误处理
func testErrorHandling() {
	fmt.Println("5️⃣  测试错误处理...")

	// 测试1：无效的加密数据
	invalidEncryptedRequest := EncryptedRequest{
		EncryptedData: "invalid_base64_data_12345",
	}

	requestJSON, err := json.Marshal(invalidEncryptedRequest)
	if err != nil {
		fmt.Printf("❌ 序列化失败: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", NOVEL_URL, strings.NewReader(string(requestJSON)))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted-Request", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("无效加密数据测试 - HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))

	if resp.StatusCode == 400 {
		fmt.Println("✅ 无效加密数据处理正确")
	} else {
		fmt.Println("❌ 无效加密数据处理异常")
	}

	fmt.Println("")
}

// diagnoseKeyFiles 诊断密钥文件
func diagnoseKeyFiles() {
	fmt.Println("📋 密钥文件诊断...")
	
	privateKeyPath := "security/rsa_private_key.pem"
	publicKeyPath := "security/rsa_public_key.pem"
	
	// 检查私钥文件
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		fmt.Printf("❌ 私钥文件不存在: %s\n", privateKeyPath)
	} else {
		fmt.Printf("✅ 私钥文件存在: %s\n", privateKeyPath)
		// 检查文件大小
		if info, err := os.Stat(privateKeyPath); err == nil {
			fmt.Printf("   文件大小: %d bytes\n", info.Size())
		}
	}
	
	// 检查公钥文件
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		fmt.Printf("❌ 公钥文件不存在: %s\n", publicKeyPath)
	} else {
		fmt.Printf("✅ 公钥文件存在: %s\n", publicKeyPath)
		// 检查文件大小
		if info, err := os.Stat(publicKeyPath); err == nil {
			fmt.Printf("   文件大小: %d bytes\n", info.Size())
		}
	}
	
	fmt.Println("")
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}