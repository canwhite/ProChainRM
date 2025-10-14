package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"novel-resource-management/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// RSARequestMiddleware RSA请求解密中间件
func RSARequestMiddleware() gin.HandlerFunc {
	// 这里返回一个闭包（即返回 gin.HandlerFunc，而不是直接执行逻辑），
	// 是因为 Gin 的中间件机制要求每个中间件都返回一个 HandlerFunc（类型为 func(*gin.Context)）。
	// 这样 Gin 框架才能在收到请求时依次调用所有中间件，每个中间件都接收当前请求的上下文（*gin.Context）。
	// 通过闭包，我们能捕获一些外部变量或配置，并返回一个具体处理请求的函数，符合 Gin 的调用方式。
	// 通常用法如下：
	//   func MyMiddleware() gin.HandlerFunc {
	//       return func(c *gin.Context) {
	//           // 具体的逻辑
	//           c.Next()
	//       }
	//   }
	return func(c *gin.Context) {
		// 检查是否为加密请求
		if !isEncryptedRequest(c) {
			c.Next()
			return
		}

		// 读取并解密请求体
		decryptedBody, err := decryptRequestBody(c)
		if err != nil {
			c.JSON(400, gin.H{"error": "解密请求失败: " + err.Error()})
			// Abort的作用是立即终止后续handler链的执行，即后面的中间件和业务处理函数不会再执行。
			// 在这里，当解密失败时需要立刻返回错误响应，不可继续流转到后续的业务逻辑，否则会收到未解密的body导致业务异常。
			c.Abort()
			return
		}

		// 将解密后的数据重新设置到请求体中
		c.Request.Body = io.NopCloser(strings.NewReader(decryptedBody))
		
		// 继续处理请求
		c.Next()
	}
}

// RSAResponseMiddleware RSA响应加密中间件
func RSAResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 拦截响应
		writer := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = writer
		
		c.Next()

		// 检查是否需要加密响应
		if shouldEncryptResponse(c) {
			// 加密响应数据
			encrypted, err := encryptResponseData(writer.body.String())
			if err != nil {
				c.JSON(500, gin.H{"error": "加密响应失败: " + err.Error()})
				return
			}
			
			// 设置加密响应
			c.Header("Content-Type", "application/json")
			c.String(200, fmt.Sprintf(`{"encrypted":true,"data":"%s"}`, encrypted))
		} else {
			// 直接返回原始响应
			c.Data(200, c.ContentType(), writer.body.Bytes())
		}
	}
}

// isEncryptedRequest 检查是否为加密请求
func isEncryptedRequest(c *gin.Context) bool {
	// 检查请求头
	encrypted := c.GetHeader("X-Encrypted-Request")
	return encrypted == "true"
}

// shouldEncryptResponse 检查是否需要加密响应
func shouldEncryptResponse(c *gin.Context) bool {
	// 如果请求是加密的，则响应也加密
	return isEncryptedRequest(c)
}

// decryptRequestBody 解密请求体
func decryptRequestBody(c *gin.Context) (string, error) {
	// 读取原始请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", fmt.Errorf("读取请求体失败: %v", err)
	}

	// 解析加密请求结构
	// 是的，前端请求时若要走加密流程，必须在HTTP请求头中加上： X-Encrypted-Request: true
	// 例如（使用fetch举例）:
	// fetch('/api/xxx', {
	//   method: 'POST',
	//   headers: { 'X-Encrypted-Request': 'true', 'Content-Type': 'application/json' },
	//   body: JSON.stringify({ encryptedData: xxx })
	// })
	// 这样后端才会自动识别并按加密流程处理该请求和响应。
	var encryptedRequest struct {
		EncryptedData string `json:"encryptedData"`
		Signature     string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(body, &encryptedRequest); err != nil {
		return "", fmt.Errorf("解析加密请求失败: %v", err)
	}

	// 解密数据
	decrypted, err := utils.DecryptWithRSA(encryptedRequest.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}

	return decrypted, nil
}

// encryptResponseData 加密响应数据
func encryptResponseData(data string) (string, error) {
	return utils.EncryptWithRSA(data)
}

// responseBodyWriter 用于拦截响应的Writer
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// RSACombinedMiddleware 组合的RSA中间件（解密请求+加密响应）
func RSACombinedMiddleware() gin.HandlerFunc {
	decryptMiddleware := RSARequestMiddleware()
	
	return func(c *gin.Context) {
		// 先解密请求
		decryptMiddleware(c)
		
		if c.IsAborted() {
			return
		}
		
		// 继续处理请求
		c.Next()
		
		// 注意：响应加密需要在gin的其他地方处理
		// 这里可以注册一个after hook或者使用自定义的响应拦截器
	}
}