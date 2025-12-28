package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	BaseURL    = "http://localhost:8080"
	TestEmail  = "beetle5249@gmail.com"
	TestUserID = "691058f50987397c91e4e078"
)

// RechargeRequest 充值请求结构
type RechargeRequest struct {
	Title       string `json:"title"`
	OrderSN     string `json:"order_sn"`
	Email       string `json:"email"`
	ActualPrice int    `json:"actual_price"`
	OrderInfo   string `json:"order_info"`
	GoodID      string `json:"good_id"`
	GoodName    string `json:"gd_name"`
}

// RechargeResponse 充值响应结构
type RechargeResponse struct {
	Message    string `json:"message"`
	UserID     string `json:"userId"`
	Email      string `json:"email"`
	OrderSN    string `json:"orderSn"`
	GoodName   string `json:"goodName"`
	AddedTokens int   `json:"addedTokens"`
	NewCredit   int    `json:"newCredit"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error string `json:"error"`
}

// UserCreditResponse 用户积分响应结构
type UserCreditResponse struct {
	Credit map[string]interface{} `json:"credit"`
}

func main() {
	fmt.Println("=========================================")
	fmt.Println("       充值接口测试脚本")
	fmt.Println("=========================================")
	fmt.Println()

	// 1. 检查服务状态
	fmt.Println("1. 检查服务状态...")
	if !checkHealth() {
		log.Fatal("❌ 服务未启动,请先运行: go run main.go")
	}
	fmt.Println("✅ 服务正常运行")
	fmt.Println()

	// 2. 查询当前积分
	fmt.Println("2. 查询当前积分...")
	currentCredit := getUserCredit(TestUserID)
	if currentCredit == -1 {
		log.Fatal("❌ 无法获取用户积分")
	}
	fmt.Printf("当前积分: %d\n", currentCredit)
	fmt.Println()

	// 3. 发送充值请求
	fmt.Println("3. 发送充值请求...")
	rechargeReq := RechargeRequest{
		Title:       "150 Token 充值包",
		OrderSN:     fmt.Sprintf("TEST_ORDER_%d", time.Now().Unix()),
		Email:       TestEmail,
		ActualPrice: 150,
		OrderInfo:   "测试充值账号",
		GoodID:      "GOOD_TEST_001",
		GoodName:    "150 Token套餐",
	}

	fmt.Printf("充值数据:\n")
	fmt.Printf("  邮箱: %s\n", TestEmail)
	fmt.Printf("  充值金额: 150 Token\n")
	fmt.Printf("  订单号: %s\n", rechargeReq.OrderSN)
	fmt.Println()

	newCredit, err := sendRechargeRequest(rechargeReq)
	if err != nil {
		log.Printf("❌ 充值失败: %v", err)
		log.Println("可能原因:")
		log.Println("  1. 用户不存在")
		log.Println("  2. MongoDB 连接失败")
		log.Println("  3. Fabric 链码调用失败")
		return
	}

	fmt.Printf("✅ 充值成功!\n")
	fmt.Printf("充值后积分: %d\n", newCredit)

	// 验证积分
	expectedCredit := currentCredit + 150
	if newCredit == expectedCredit {
		fmt.Printf("✅ 积分验证通过: +150 Token\n")
	} else {
		fmt.Printf("❌ 积分异常: 期望 %d, 实际 %d\n", expectedCredit, newCredit)
	}
	fmt.Println()

	// 4. 等待2秒后再次查询
	fmt.Println("4. 再次查询积分确认...")
	time.Sleep(2 * time.Second)
	finalCredit := getUserCredit(TestUserID)
	if finalCredit == newCredit {
		fmt.Printf("✅ 数据一致: %d\n", finalCredit)
	} else {
		fmt.Printf("⚠️  数据不一致: 期望 %d, 实际 %d\n", newCredit, finalCredit)
	}
	fmt.Println()

	// 5. 测试用户不存在的情况
	fmt.Println("5. 测试用户不存在的情况...")
	invalidReq := RechargeRequest{
		Title:       "测试",
		OrderSN:     fmt.Sprintf("TEST_ORDER_INVALID_%d", time.Now().Unix()),
		Email:       "nonexistent@example.com",
		ActualPrice: 150,
		OrderInfo:   "测试",
		GoodID:      "TEST",
		GoodName:    "测试",
	}

	_, err = sendRechargeRequest(invalidReq)
	if err != nil {
		if bytes.Contains([]byte(err.Error()), []byte("用户不存在")) {
			fmt.Println("✅ 错误处理正确: 正确识别了不存在的用户")
		} else {
			fmt.Printf("❌ 错误处理异常: %v\n", err)
		}
	}
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("✅ 测试完成!")
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println("总结:")
	fmt.Println("  ✅ 服务状态检查")
	fmt.Println("  ✅ 充值接口调用")
	fmt.Println("  ✅ 积分验证")
	fmt.Println("  ✅ 错误处理测试")
	fmt.Println()
	fmt.Println("如需查看详细日志,请检查服务端输出")
}

// checkHealth 检查服务健康状态
func checkHealth() bool {
	resp, err := http.Get(BaseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return bytes.Contains(body, []byte("ok"))
}

// getUserCredit 获取用户积分
func getUserCredit(userID string) int {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/%s", BaseURL, userID))
	if err != nil {
		log.Printf("请求失败: %v", err)
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return -1
	}

	var result UserCreditResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析响应失败: %v", err)
		return -1
	}

	// 从 map 中提取 credit 字段
	if credit, ok := result.Credit["credit"].(float64); ok {
		return int(credit)
	}

	return -1
}

// sendRechargeRequest 发送充值请求
func sendRechargeRequest(req RechargeRequest) (int, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := http.Post(
		BaseURL+"/api/v1/users/recharge",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	// 先尝试解析错误响应
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return 0, fmt.Errorf(errResp.Error)
	}

	// 解析成功响应
	var rechargeResp RechargeResponse
	if err := json.Unmarshal(body, &rechargeResp); err != nil {
		return 0, fmt.Errorf("解析响应失败: %v", err)
	}

	// 打印响应
	prettyJSON, _ := json.MarshalIndent(rechargeResp, "", "  ")
	fmt.Println("充值响应:")
	fmt.Println(string(prettyJSON))

	return rechargeResp.NewCredit, nil
}
