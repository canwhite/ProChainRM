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
	"novel-resource-management/middleware"
	"novel-resource-management/service"
	"novel-resource-management/utils"
)

type Server struct {
	router        *gin.Engine
	httpServer    *http.Server
	novelService  *service.NovelService
	creditService *service.UserCreditService
	eventService  *service.EventService
	network       *client.Network
}

// create new service interface
func NewServer(gateway *client.Gateway) *Server {
	// 初始化RSA加密解密器
	if err := utils.InitRSACrypto(); err != nil {
		log.Printf("警告: RSA加密解密器初始化失败: %v", err)
	}

	//network
	network := gateway.GetNetwork("mychannel")

	novelService, err := service.NewNovelService(gateway)
	if err != nil {
		panic(fmt.Sprintf("初始化 NovelService 失败: %v", err))
	}
	creditService, err := service.NewUserCreditService(gateway)
	if err != nil {
		panic(fmt.Sprintf("初始化 UserCreditService 失败: %v", err))
	}
	eventService := service.NewEventService(gateway)

	server := &Server{
		router:        gin.Default(),
		novelService:  novelService,
		creditService: creditService,
		eventService:  eventService,
		network:       network,
	}

	server.setupRoutes()
	
	return server
}

// 方法指示器
func (s *Server) setupRoutes() {
	// 先接路由，再接方法

	// 添加全局防抖中间件
	debounce := middleware.NewDebounceMiddleware()
	s.router.Use(debounce.Debounce())

	s.router.GET("/health", s.healthCheck)

	novels := s.router.Group("/api/v1/novels")
	{
		//RESTFUL API
		novels.GET("", s.getAllNovels)
		novels.GET("/:id", s.getNovel)
		//delete
		novels.DELETE("/:id", s.deleteNovel)

		//先不用
		novels.POST("", s.createNovel)
		//update,PUT是整体更新，PATCH是部分更新
		novels.PUT("/:id", s.updateNovel)

		/*
		// 需要RSA加密的路由（POST和PUT）
		encryptedNovels := novels.Group("")
		encryptedNovels.Use(middleware.RSARequestMiddleware())
		{
			//create
			encryptedNovels.POST("", s.createNovel)
			//update,PUT是整体更新，PATCH是部分更新
			encryptedNovels.PUT("/:id", s.updateNovel)
		}
		*/
	}

	users := s.router.Group("/api/v1/users")
	{
		//get
		users.GET("",s.getAllUserCredits)
		users.GET("/:id",s.getUserCredit)

		//delete
		users.DELETE("/:id",s.deleteUserCredit)

		// 需要RSA加密的路由
		
		encryptedUsers := users.Group("")
		// 虽然通常建议包名和文件夹名一致，但 Go 并不强制要求。
		// 如果 package 和文件夹名不一致——比如文件在 middleware 目录，但声明 package mware——
		// 你在 import 时依然写 import "novel-resource-management/middleware"，但代码中用 mware.xxx 来访问。
		// 总之，“import 路径”（即文件夹路径）用于定位代码源文件，而“包名”决定了代码里实际的调用前缀。
		encryptedUsers.Use(middleware.RSARequestMiddleware())
		{
			//create
			encryptedUsers.POST("",s.createUserCredit)

			//update
			encryptedUsers.PUT("/:id",s.updateUserCredit)
		}
	}

	events := s.router.Group("/api/v1/events")
	{
		events.GET("/listen",s.streamEvents)
	}

	
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// GIN do not need to return some data
func (s *Server) getAllNovels(c *gin.Context) {

	novels, err := s.novelService.GetAllNovels()
	if err != nil {
		//注意c.JSON和gin.H
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if novels == nil {
		novels = []map[string]interface{}{}
	}

	//c.JSON不用return
	c.JSON(
		http.StatusOK,
		gin.H{
			"novels": novels,
			"count":  len(novels),
		},
	)
}

func (s *Server) getNovel(c *gin.Context) {
	id := c.Param("id")
	//3.id == nil是指针判断
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Don't get the novel id",
		})
		return
	}
	//1. 短变量声明
	novel, err := s.novelService.ReadNovel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			//2.结构体逗号
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"novel": novel,
	})
}

func (s *Server) createNovel(c *gin.Context) {
	//先声明后挂值，区别于短变量声明
	//不用逗号，key:"value"
	var req struct {
		ID           string `json:"id" binding:"required"`
		Author       string `json:"author" binding:"required"`
		StoryOutline string `json:"storyOutline" binding:"required"`
		Subsections  string `json:"subsections" binding:"required"`
		Characters   string `json:"characters" binding:"required"`
		Items        string `json:"items" binding:"required"`
		TotalScenes  string `json:"totalScenes" binding:"required"`
		CreatedAt    string `json:"createdAt" binding:"omitempty"`
	}

	//这个err := 只在这个if作用域； 
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//参数顺序
	//id, author, storyOutline, subsections, characters, items, totalScenes string
	if err := s.novelService.CreateNovel(req.ID, req.Author, req.StoryOutline, req.Subsections, req.Characters, req.Items, req.TotalScenes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "create novel successful",
		"id":      req.ID,
	})
}

func (s *Server) updateNovel(c *gin.Context) {
	//todo
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "you do not get the id!",
		})
		return
	}

	var req struct {
		// Update请求不需要ID字段，使用URL路径中的ID
		Author       string `json:"author" binding:"required"`
		StoryOutline string `json:"storyOutline" binding:"required"`
		Subsections  string `json:"subsections" binding:"required"`
		Characters   string `json:"characters" binding:"required"`
		Items        string `json:"items" binding:"required"`
		TotalScenes  string `json:"totalScenes" binding:"required"`
		UpdatedAt    string `json:"updatedAt" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.novelService.UpdateNovel(id, req.Author, req.StoryOutline, req.Subsections, req.Characters, req.Items, req.TotalScenes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update successfully",
		"id":      id,
	})
}

func (s *Server) deleteNovel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "the param is ignore",
		})
		return
	}
	//novel
	if err := s.novelService.DeleteNovel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "delete successfully",
		"id":      id,
	})
}

func (s *Server) streamEvents(c *gin.Context){
	// 这三个属性分别是：
	// 1. Content-Type: 设置为 "text/event-stream"，表示响应内容是 Server-Sent Events（SSE）流，前端可以实时接收事件推送。
	// 2. Cache-Control: 设置为 "no-cache"，告知浏览器不要对该响应进行缓存，确保每次都能获取到最新的事件数据。
	// 3. Connection: 设置为 "keep-alive"，保持 HTTP 连接不断开，以便持续推送事件数据给客户端。
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	//go思维
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	
	//chain code 都返回两个参数
	events,err := s.network.ChaincodeEvents(ctx,"novel-basic")

	if err != nil{
		//spritf会返回字符串，println不会
		c.String(http.StatusInternalServerError,fmt.Sprintf("error: %v", err))
		return
	}

	//c.stream和闭包
	c.Stream(func(w io.Writer) bool{
		select{
		case event := <- events:
			//todo,最终的操作
			//hyper success
			novel := s.formatJSON(event.Payload)
			//Fprintf用于将指定的字符串写入io.Writer
			fmt.Fprintf(w, "data: %s - %s\n\n", event.EventName, novel)
			return true
		case <- ctx.Done():
			return false
		}
	})
}


func (s *Server) getAllUserCredits(c *gin.Context){
	credits, err := s.creditService.GetAllUserCredits()
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":err.Error(),
		})
		return
	}
	//如果是nil，我们返回空数组
	if credits == nil{
		credits = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK,gin.H{
		"credits":credits,
		"count":len(credits),
	})
}

func (s *Server)getUserCredit(c *gin.Context){
	id := c.Param("id")

	if id == ""{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"can not get the user credit id",
		})
		return
	}
	credit,err :=  s.creditService.ReadUserCredit(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if credit == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到该用户积分信息",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credit": credit,
	})
}

func (s *Server)createUserCredit(c *gin.Context){
	//之前TotalUsed加了binding:"required"，因为传参为0报错了
	var req struct{
		UserID        string `json:"userId"`
		Credit        int    `json:"credit"`
		TotalUsed     int    `json:"totalUsed"`
		TotalRecharge int    `json:"totalRecharge"`
		CreatedAt     string `json:"createdAt,omitempty"`
		UpdatedAt     string `json:"updatedAt,omitempty"`
	}

	// 添加调试日志：读取原始请求体
	bodyBytes, _ := c.GetRawData()
	log.Printf("DEBUG: Raw request body: %s", string(bodyBytes))
	
	// 重新设置请求体，因为GetRawData会消耗它
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 然后从JSON转为interface{}
	if err := c.ShouldBindJSON(&req); err!= nil{
		log.Printf("DEBUG: JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":err.Error(),
		})
		return
	}

	// 添加调试日志：验证后的数据
	log.Printf("DEBUG: Parsed request - UserID: %s, Credit: %d, TotalUsed: %d, TotalRecharge: %d", 
		req.UserID, req.Credit, req.TotalUsed, req.TotalRecharge)

	// 手动验证必填字段
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"userId不能为空",
		})
		return
	}
	if req.Credit < 0 {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"credit不能为负数",
		})
		return
	}
	if req.TotalUsed < 0 {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"totalUsed不能为负数",
		})
		return
	}
	if req.TotalRecharge < 0 {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"totalRecharge不能为负数",
		})
		return
	}

	// then we create the user credit
	// userId string, credit int, totalUsed int, totalRecharge int
    if err:= s.creditService.CreateUserCredit(req.UserID,req.Credit,req.TotalUsed,req.TotalRecharge); err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"message":"create successfully",
		"id":req.UserID,
	})
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Fabric Gateway API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}


func (s * Server)updateUserCredit(c *gin.Context){

	id := c.Param("id")
	if id == ""{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"can not get user credit id",
		})
		return
	}

	var req  struct {
		UserID        string `json:"userId" binding:"required"`
		Credit        int    `json:"credit" binding:"required"`
		TotalUsed     int    `json:"totalUsed" binding:"required"`
		TotalRecharge int    `json:"totalRecharge" binding:"required"`
		CreatedAt     string `json:"createdAt" binding:"omitempty"`
		UpdatedAt     string `json:"updatedAt" binding:"omitempty"`
	}
	//then we get interface{}
	if err := c.ShouldBindJSON(&req); err!= nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":err.Error(),
		})
		return
	}	
	
	//拿到对应的参数去处理
	//userId string, credit int, totalUsed int, totalRecharge int
	if err := s.creditService.UpdateUserCredit(id,req.Credit,req.TotalUsed,req.TotalRecharge) ; err != nil{
		//todo
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"message":"update user credit successfully",
		"id":id,
	})
}

func (s *Server)deleteUserCredit(c * gin.Context){
	//todo, delete
	id := c.Param("id")
	if id == ""{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"id is not found",
		})
		return
	}

	if err := s.creditService.DeleteUserCredit(id); err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"message":"delete successfully!",
		"id":id,
	})
}


func (s *Server) Start(address string) error{
	// 初始化 http.Server，使用传入的地址
	s.httpServer = &http.Server{
		Addr:    address,
		Handler: s.router,
	}
	
	log.Printf("🚀 Starting server on %s", address)
	return s.httpServer.ListenAndServe()
}

func (s *Server) formatJSON(data []byte)string {
	var result bytes.Buffer
	//第三个参数字符串的前缀，第四个参数是缩进
	if err :=json.Indent(&result,data,"","    "); err != nil{
		return string(data)
	}
	return result.String()
}

