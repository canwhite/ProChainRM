package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
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
	Timestamp   string `json:"timestamp"`   // 新增：时间戳
	Signature   string `json:"signature"`   // 新增：HMAC 签名
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

// getRechargeSecretKey 从环境变量获取充值接口的 HMAC 密钥
func getRechargeSecretKey() string {
	key := os.Getenv("RECHARGE_SECRET_KEY")
	if key == "" {
		log.Printf("⚠️ 警告: RECHARGE_SECRET_KEY 环境变量未设置，使用默认值")
		key = "your-secret-key-change-in-production"
	}
	return key
}

// computeHMACSignature 计算 HMAC-SHA256 签名
func computeHMACSignature(params map[string]string, secretKey string) string {
	// 步骤1: 按字母序排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 步骤2: 拼接参数
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramStr := strings.Join(parts, "&")

	// 步骤3: 计算 HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(paramStr))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
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

	// 6. 幂等性测试（相同订单号重复请求）
	fmt.Println("6. 测试幂等性（相同订单号重复请求）...")
	idempotentOrderSN := fmt.Sprintf("TEST_IDEMPOTENT_%d", time.Now().Unix())

	// 第一次请求
	idempotentReq1 := RechargeRequest{
		Title:       "幂等性测试",
		OrderSN:     idempotentOrderSN,
		Email:       TestEmail,
		ActualPrice: 150,
		OrderInfo:   "幂等性测试第一次",
		GoodID:      "TEST_IDEMPOTENT",
		GoodName:    "幂等性测试套餐",
	}

	credit1, err1 := sendRechargeRequest(idempotentReq1)
	if err1 != nil {
		fmt.Printf("❌ 第一次请求失败: %v\n", err1)
	} else {
		fmt.Printf("✅ 第一次请求成功: 积分=%d\n", credit1)

		// 第二次请求（相同订单号）
		idempotentReq2 := RechargeRequest{
			Title:       "幂等性测试",
			OrderSN:     idempotentOrderSN, // 相同订单号
			Email:       TestEmail,
			ActualPrice: 150,
			OrderInfo:   "幂等性测试第二次",
			GoodID:      "TEST_IDEMPOTENT",
			GoodName:    "幂等性测试套餐",
		}

		credit2, err2 := sendRechargeRequest(idempotentReq2)
		if err2 != nil {
			fmt.Printf("❌ 第二次请求失败（不应该失败）: %v\n", err2)
		} else if credit2 == credit1 {
			fmt.Printf("✅ 幂等性验证通过: 两次返回相同积分=%d\n", credit2)
		} else {
			fmt.Printf("❌ 幂等性验证失败: 第一次积分=%d, 第二次积分=%d\n", credit1, credit2)
		}
	}
	fmt.Println()

	// 7. 签名错误测试
	fmt.Println("7. 测试签名错误...")
	wrongSigReq := RechargeRequest{
		Title:       "签名错误测试",
		OrderSN:     fmt.Sprintf("TEST_WRONG_SIG_%d", time.Now().Unix()),
		Email:       TestEmail,
		ActualPrice: 150,
		OrderInfo:   "签名错误测试",
		GoodID:      "TEST_WRONG_SIG",
		GoodName:    "签名错误测试套餐",
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
		Signature:   "this_is_a_wrong_signature_1234567890abcdef", // 错误签名
	}

	_, err = sendRechargeRequest(wrongSigReq)
	if err != nil {
		if bytes.Contains([]byte(err.Error()), []byte("签名验证失败")) {
			fmt.Println("✅ 签名错误测试通过: 正确拒绝了错误签名")
		} else {
			fmt.Printf("❌ 签名错误测试异常: %v\n", err)
		}
	} else {
		fmt.Println("❌ 签名错误测试失败: 错误签名应该被拒绝")
	}
	fmt.Println()

	// 8. 时间戳过期测试（5分钟前）
	fmt.Println("8. 测试时间戳过期（5分钟前）...")
	expiredTimestamp := time.Now().Unix() - (5 * 60 + 10) // 5分10秒前
	expiredReq := RechargeRequest{
		Title:       "时间戳过期测试",
		OrderSN:     fmt.Sprintf("TEST_EXPIRED_%d", time.Now().Unix()),
		Email:       TestEmail,
		ActualPrice: 150,
		OrderInfo:   "时间戳过期测试",
		GoodID:      "TEST_EXPIRED",
		GoodName:    "时间戳过期测试套餐",
		Timestamp:   strconv.FormatInt(expiredTimestamp, 10),
		// 签名由 sendRechargeRequest 计算
	}

	_, err = sendRechargeRequest(expiredReq)
	if err != nil {
		if bytes.Contains([]byte(err.Error()), []byte("请求过期")) {
			fmt.Println("✅ 时间戳过期测试通过: 正确拒绝了过期请求")
		} else {
			fmt.Printf("❌ 时间戳过期测试异常: %v\n", err)
		}
	} else {
		fmt.Println("❌ 时间戳过期测试失败: 过期请求应该被拒绝")
	}
	fmt.Println()

	// 9. 时间戳未来测试
	fmt.Println("9. 测试时间戳未来...")
	futureTimestamp := time.Now().Unix() + 300 // 5分钟后
	futureReq := RechargeRequest{
		Title:       "时间戳未来测试",
		OrderSN:     fmt.Sprintf("TEST_FUTURE_%d", time.Now().Unix()),
		Email:       TestEmail,
		ActualPrice: 150,
		OrderInfo:   "时间戳未来测试",
		GoodID:      "TEST_FUTURE",
		GoodName:    "时间戳未来测试套餐",
		Timestamp:   strconv.FormatInt(futureTimestamp, 10),
		// 签名由 sendRechargeRequest 计算
	}

	_, err = sendRechargeRequest(futureReq)
	if err != nil {
		if bytes.Contains([]byte(err.Error()), []byte("请求时间戳来自未来")) {
			fmt.Println("✅ 时间戳未来测试通过: 正确拒绝了未来时间戳")
		} else {
			fmt.Printf("❌ 时间戳未来测试异常: %v\n", err)
		}
	} else {
		fmt.Println("❌ 时间戳未来测试失败: 未来时间戳应该被拒绝")
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
	fmt.Println("  ✅ 错误处理测试（用户不存在）")
	fmt.Println("  ✅ 幂等性测试")
	fmt.Println("  ✅ 签名错误测试")
	fmt.Println("  ✅ 时间戳过期测试")
	fmt.Println("  ✅ 时间戳未来测试")
	fmt.Println()
	fmt.Println("安全验证功能测试完成!")
	fmt.Println("所有安全机制（HMAC签名、时间戳验证、幂等性）均已覆盖")
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
	// 如果时间戳为空，生成当前时间戳
	if req.Timestamp == "" {
		req.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// 计算 HMAC 签名（仅在签名未提供时）
	if req.Signature == "" {
		params := map[string]string{
			"actual_price": strconv.Itoa(req.ActualPrice),
			"email":        req.Email,
			"order_sn":     req.OrderSN,
			"timestamp":    req.Timestamp,
		}

		secretKey := getRechargeSecretKey()
		req.Signature = computeHMACSignature(params, secretKey)
	}

	// 安全地显示签名（前16字符）
	signaturePreview := req.Signature
	if len(signaturePreview) > 16 {
		signaturePreview = signaturePreview[:16] + "..."
	}
	log.Printf("📤 发送充值请求: orderSN=%s, timestamp=%s, signature=%s",
		req.OrderSN, req.Timestamp, signaturePreview)

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
