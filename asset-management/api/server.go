package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin" //用gin
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"sdk-go/service"
)

// Server represents the HTTP server
type Server struct {
	router       *gin.Engine
	assetService *service.AssetService
	eventService *service.EventService
	network      *client.Network
}

// NewServer creates a new HTTP server
func NewServer(gateway *client.Gateway) *Server {
	network := gateway.GetNetwork("mychannel")

	server := &Server{
		router:       gin.Default(),
		assetService: service.NewAssetService(gateway),
		eventService: service.NewEventService(gateway),
		network:      network,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	//GET 方法
	s.router.GET("/health", s.healthCheck)

	// Assets API
	assets := s.router.Group("/api/v1/assets")
	{
		// 在 RESTful API 设计中，通常通过 HTTP 方法来区分创建、更新和删除操作：
		// - 创建(Create)：使用 POST 方法。例如 POST /api/v1/assets 表示创建一个新资产。
		// - 更新(Update)：使用 PUT 方法（整体更新）或 PATCH 方法（部分更新）。如 PUT /api/v1/assets/:id。
		// - 删除(Delete)：使用 DELETE 方法。例如 DELETE /api/v1/assets/:id。
		// Gin 路由中已经通过 assets.POST、assets.PUT、assets.PATCH、assets.DELETE 进行了区分。
		// 具体的业务逻辑在对应的 handler（如 s.createAsset, s.updateAsset, s.deleteAsset）中实现。
		assets.GET("", s.getAllAssets)
		assets.GET("/:id", s.getAsset)
		assets.POST("", s.createAsset)
		assets.PUT("/:id", s.updateAsset)
		// PATCH 方法通常用于“部分更新”资源。与 PUT（整体替换资源）不同，PATCH 只需要提交需要修改的字段即可。
		// 例如 PATCH /api/v1/assets/:id/transfer 可以用于资产的转移操作，只需提供新的拥有者信息，而不必提交整个资产对象。
		// 在 Gin 中，assets.PATCH("/:id/transfer", s.transferAsset) 就是注册了一个 PATCH 路由，用于处理资产转移的业务逻辑。
		assets.PATCH("/:id/transfer", s.transferAsset)
		assets.DELETE("/:id", s.deleteAsset)
	}

	// Events API
	events := s.router.Group("/api/v1/events")
	{
		events.GET("/listen", s.streamEvents)
	}
}

// healthCheck returns server status
func (s *Server) healthCheck(c *gin.Context) {
	// 这里的 c.JSON 是 Gin 框架中用于返回 JSON 格式响应的方法。
	// 具体来说，c 是 *gin.Context 类型，代表当前的请求上下文。
	// c.JSON(statusCode, data) 会设置 HTTP 状态码（如 http.StatusOK），
	// 并将 data（可以是 map、结构体等）序列化为 JSON 格式返回给客户端。
	// 例如：
	// 
	// c.JSON(http.StatusOK, gin.H{
	//     "status":  "ok",
	//     "message": "Fabric Gateway API is running",
	//     "time":    time.Now().Format(time.RFC3339),
	// })
	// 客户端收到的就是一个 JSON 对象，包含 status、message 和 time 字段。
	// 除了 http.StatusOK（200）之外，常用的 HTTP 状态码还有：
	// - http.StatusCreated（201）：资源创建成功
	// - http.StatusBadRequest（400）：请求参数有误
	// - http.StatusUnauthorized（401）：未认证
	// - http.StatusForbidden（403）：无权限
	// - http.StatusNotFound（404）：资源未找到
	// - http.StatusConflict（409）：资源冲突
	// - http.StatusInternalServerError（500）：服务器内部错误
	// - http.StatusServiceUnavailable（503）：服务不可用
	// 这些常量都定义在 net/http 包中，可以直接使用。
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Fabric Gateway API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// getAllAssets returns all assets
func (s *Server) getAllAssets(c *gin.Context) {
	assets, err := s.assetService.GetAllAssets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
		"count":  len(assets),
	})
}

// getAsset returns a specific asset by ID
func (s *Server) getAsset(c *gin.Context) {
	//get直接通过Param获取
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	asset, err := s.assetService.ReadAsset(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"asset": asset})
}

// createAsset creates a new asset
// POST请求
func (s *Server) createAsset(c *gin.Context) {
	//struct后边的也是类型的一部分
	var req struct {
		ID             string `json:"id" binding:"required"`
		Color          string `json:"color" binding:"required"`
		Size           string `json:"size" binding:"required"`
		Owner          string `json:"owner" binding:"required"`
		AppraisedValue string `json:"appraisedValue" binding:"required"`
	}

	// 是的，这一步是将前端传来的JSON数据自动绑定（反序列化）到Go的结构体（struct）变量req中。
	// 这样后续就可以直接通过req.ID、req.Color等字段来访问请求体中的数据了。
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.assetService.CreateAsset(req.ID, req.Color, req.Size, req.Owner, req.AppraisedValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "asset created successfully",
		"id":      req.ID,
	})
}

// updateAsset updates an existing asset
func (s *Server) updateAsset(c *gin.Context) {
	//它要从路径中获取信息
	id := c.Param("id")
	if id == "" {
		//参数就是bad request
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	var req struct {
		Color          string `json:"color" binding:"required"`
		Size           string `json:"size" binding:"required"`
		Owner          string `json:"owner" binding:"required"`
		AppraisedValue string `json:"appraisedValue" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.assetService.UpdateAsset(id, req.Color, req.Size, req.Owner, req.AppraisedValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "asset updated successfully",
		"id":      id,
	})
}

// transferAsset transfers asset ownership
func (s *Server) transferAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	var req struct {
		NewOwner string `json:"newOwner" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.assetService.TransferAsset(id, req.NewOwner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "asset transferred successfully",
		"id":      id,
		"owner":   req.NewOwner,
	})
}

// deleteAsset deletes an asset
func (s *Server) deleteAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset ID is required"})
		return
	}

	if err := s.assetService.DeleteAsset(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "asset deleted successfully",
		"id":      id,
	})
}

// streamEvents streams events to client (WebSocket-like)
func (s *Server) streamEvents(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Start event listening
	events, err := s.network.ChaincodeEvents(ctx, "basic")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error: %v", err))
		return
	}
	//除了c.JSON,这里c.Stream
	// 这里类似于监听链码事件，持续推送给前端
	// 事件流会在下方的c.Stream中处理，这里无需额外代码
	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-events:
			asset := formatJSON(event.Payload)
			fmt.Fprintf(w, "data: %s - %s\n\n", event.EventName, asset)
			return true
		case <-ctx.Done():
			return false
		}
	})
}

// Start starts the HTTP server
func (s *Server) Start(address string) error {
	log.Printf("🚀 Starting server on %s", address)
	return s.router.Run(address)
}

// formatJSON formats JSON data with proper indentation
func formatJSON(data []byte) string {
	var result bytes.Buffer
	//不一样的是第一个是接收值
	if err := json.Indent(&result, data, "", "  "); err != nil {
		return string(data)
	}
	return result.String()
}