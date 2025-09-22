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

		// struct tag（结构体标签）是Go语言本身的特性，不是gin特有的。
		// Go的struct tag允许你为结构体字段添加元数据，常用于序列化（如json、xml）、数据库映射、表单校验等。
		// gin框架只是利用了Go的struct tag机制，
		// 定义了自己的tag（如`binding`、`form`等）来实现参数绑定和校验。
		// 总结：struct tag是Go语言的，gin只是用它来做参数绑定和校验。
		// 这里的 `json:"id" binding:"required"` 是结构体标签（struct tag），用于指定该字段在序列化/反序列化（
		// 如 JSON <-> Go struct）时的映射关系，以及在绑定请求参数时的校验规则。
		// 具体来说：
		// - `json:"id"` 表示该字段在 JSON 数据中的键名是 "id"。
		// - `binding:"required"` 表示在使用 gin 框架绑定请求体（如 ShouldBindJSON）时，这个字段是必填的，否则会校验失败。
		// 常见的 tag 选项有：


		// 1. json 标签：
		//    - `json:"name"`：指定 JSON 字段名
		//    - `json:"name,omitempty"`：如果该字段为零值则序列化时忽略
		//    - `json:"-"`：序列化/反序列化时忽略该字段


		// 2. binding 标签（gin 框架）：
		//    - `binding:"required"`：必填
		//    - `binding:"omitempty"`：可选
		//    - `binding:"min=1,max=10"`：长度或数值范围校验
		//    - `binding:"email"`：邮箱格式校验
		//    - `binding:"gte=0,lte=100"`：数值区间校验
		//    - 还可以组合多个校验条件，如：`binding:"required,min=3,max=20"`
		
		// 3. form 标签（用于表单绑定）：
		//    - `form:"username"`：指定表单字段名

	
		// 4. uri 标签（用于路径参数绑定）：
		//    - `uri:"id"`：指定路径参数名
		// 这些标签可以组合使用，具体取决于你的业务需求和数据来源。
		ID             string `json:"id" binding:"required"`
		Color          string `json:"color" binding:"required"`
		Size           string `json:"size" binding:"required"`
		Owner          string `json:"owner" binding:"required"`
		AppraisedValue string `json:"appraisedValue" binding:"required"`
	}

	// 是的，这一步是将前端传来的JSON数据自动绑定（反序列化）到Go的结构体（struct）变量req中。
	// 这样后续就可以直接通过req.ID、req.Color等字段来访问请求体中的数据了。
	// 除了 ShouldBindJSON，gin 还提供了多种数据绑定方法，常见的有：
	// 1. ShouldBind：自动根据 Content-Type 选择绑定方式（JSON、表单、XML等）
	//    err := c.ShouldBind(&req)
	// 2. ShouldBindQuery：绑定 URL 查询参数（?id=xxx&color=red）
	//    err := c.ShouldBindQuery(&req)
	// 3. ShouldBindForm：绑定表单数据（Content-Type: application/x-www-form-urlencoded 或 multipart/form-data）
	//    err := c.ShouldBindForm(&req)
	// 4. ShouldBindUri：绑定路径参数（如 /assets/:id）
	//    err := c.ShouldBindUri(&req)
	// 5. ShouldBindHeader：绑定请求头参数
	//    err := c.ShouldBindHeader(&req)
	// 6. ShouldBindXML：绑定 XML 数据
	//    err := c.ShouldBindXML(&req
	// 7. ShouldBindYAML：绑定 YAML 数据
	//    err := c.ShouldBindYAML(&req)
	// 8. ShouldBindTOML：绑定 TOML 数据
	//    err := c.ShouldBindTOML(&req)
	// 这些方法都可以用于将请求中的数据自动映射到结构体中，具体选择哪种方法取决于你的数据来源和 Content-Type。
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