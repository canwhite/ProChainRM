//go:build test

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"novel-resource-management/utils"
	"os"
	"strings"
)

// RSA测试专用的结构体
type EncryptedRequest struct {
	EncryptedData string `json:"encryptedData"`
}

type UserCredit struct {
	UserID        string `json:"userId"`
	Credit        int    `json:"credit"`
	TotalUsed     int    `json:"totalUsed"`
	TotalRecharge int    `json:"totalRecharge"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

const (
	BASE_URL = "http://localhost:8080"
	USER_URL = BASE_URL + "/api/v1/users"
)

func main() {
	fmt.Println("🔐 开始RSA加密中间件测试...")
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
	testNormalAPICall()

	// 测试3：加密API调用
	testEncryptedAPICall()

	// 测试4：错误处理
	testErrorHandling()

	fmt.Println("============================")
	fmt.Println("🏁 RSA加密中间件测试完成")
}

// testRSAUtils 测试RSA工具类
func testRSAUtils() {
	fmt.Println("1️⃣  测试RSA工具类...")

	// 准备测试数据
	testData := UserCredit{
		UserID:        "test_rsa_user_001",
		Credit:        100,
		TotalUsed:     0,
		TotalRecharge: 100,
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

// testNormalAPICall 测试普通API调用（不加密）
func testNormalAPICall() {
	fmt.Println("2️⃣  测试普通API调用（不加密）...")

	// 创建用户积分（不加密）
	userCredit := UserCredit{
		UserID:        "test_normal_user_001",
		Credit:        50,
		TotalUsed:     10,
		TotalRecharge: 60,
	}

	jsonData, err := json.Marshal(userCredit)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	resp, err := http.Post(USER_URL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("❌ 普通API调用失败: %v\n", err)
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

	if resp.StatusCode == 200 {
		fmt.Println("✅ 普通API调用成功")
	} else {
		fmt.Println("❌ 普通API调用失败")
	}
	fmt.Println("")
}

// testEncryptedAPICall 测试加密API调用
func testEncryptedAPICall() {
	fmt.Println("3️⃣  测试加密API调用...")

	// 准备测试数据
	userCredit := UserCredit{
		UserID:        "test_encrypted_user_001",
		Credit:        200,
		TotalUsed:     20,
		TotalRecharge: 220,
	}

	jsonData, err := json.Marshal(userCredit)
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
	req, err := http.NewRequest("POST", USER_URL, strings.NewReader(string(requestJSON)))
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
		fmt.Printf("❌ 加密API调用失败: %v\n", err)
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

	if resp.StatusCode == 200 {
		fmt.Println("✅ 加密API调用成功")
	} else {
		fmt.Println("❌ 加密API调用失败")
	}
	fmt.Println("")
}

// testErrorHandling 测试错误处理
func testErrorHandling() {
	fmt.Println("4️⃣  测试错误处理...")

	// 测试1：无效的加密数据
	invalidEncryptedRequest := EncryptedRequest{
		EncryptedData: "invalid_base64_data_12345",
	}

	requestJSON, err := json.Marshal(invalidEncryptedRequest)
	if err != nil {
		fmt.Printf("❌ 序列化失败: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", USER_URL, strings.NewReader(string(requestJSON)))
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

	// 测试2：空加密数据
	emptyRequest := EncryptedRequest{
		EncryptedData: "",
	}

	requestJSON, _ = json.Marshal(emptyRequest)
	req, _ = http.NewRequest("POST", USER_URL, strings.NewReader(string(requestJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted-Request", "true")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("空加密数据测试 - HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))

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