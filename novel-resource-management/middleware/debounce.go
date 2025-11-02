package middleware

import (
	"sync"
	"time"
	"github.com/gin-gonic/gin"
)

// DebounceMiddleware 防抖中间件，防止短时间内重复请求
type DebounceMiddleware struct {

	requests sync.Map // 存储请求时间戳
	mu       sync.RWMutex
}

// NewDebounceMiddleware 创建新的防抖中间件
func NewDebounceMiddleware() *DebounceMiddleware {
	return &DebounceMiddleware{}
}

// Debounce 返回防抖中间件函数
// 方法接收器，关键词在于接收
func (dm *DebounceMiddleware) Debounce() gin.HandlerFunc {

	//中间件都是返回闭包
	return func(c *gin.Context) {
		// 只对修改操作进行防抖，也就是修改操作就是PUT、POST和DELETE
		if c.Request.Method != "PUT" && c.Request.Method != "POST" && c.Request.Method != "DELETE" {
			//继续向下进行，然后结束
			c.Next()
			return
		}
		
		// 获取请求唯一标识
		clientIP := c.ClientIP()
		path := c.FullPath()
		userAgent := c.Request.UserAgent()

		// 生成唯一请求键
		requestKey := clientIP + ":" + path + ":" + userAgent

		// 检查是否有相同的请求正在处理
		if timestamp, exists := dm.requests.Load(requestKey); exists {
			// 解释：attempt to load 这个 requestKey 上次请求处理的时间戳（如果有）；
			// 这实际上是一种类型转化
			lastTime := timestamp.(time.Time)
			// 如果距离上次请求不到500ms，认为是重复请求
			if time.Since(lastTime) < 500*time.Millisecond {
				c.JSON(429, gin.H{
					"error": "请求过于频繁，请稍后重试",
					"code":  "RATE_LIMITED",
				})
				//这里就相当于暂停了
				c.Abort()
				return
			}
		}

		// 记录请求时间
		dm.requests.Store(requestKey, time.Now())

		// 请求完成后延迟清理记录
		// 这里会到c.Next之后再执行，
		defer func() {
			go func(key string) {
				time.Sleep(1 * time.Second) // 1秒后清理
				//目前看到的一个Load，一个Save，一个Delete
				dm.requests.Delete(key)
			}(requestKey)
		}()

		c.Next()
	}
}