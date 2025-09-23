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
	"novel-resource-management/service"
)

type Server struct {
	router        *gin.Engine
	novelService  *service.NovelService
	creditService *service.UserCreditService
	eventService  *service.EventService
	network       *client.Network
}

// create new service interface
func NewServer(gateway *client.Gateway) *Server {
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
		router:        gin.GetDefault(),
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
	//先接路由，再接方法
	s.router.GET("/health", s.healthCheck)

	novels := s.router.Group("/api/v1/novels")
	{
		//RESTFUL API
		novels.GET("", s.getAllNovels)
		novels.GET("/:id", s.getNovel)
		//create
		novels.POST("", s.createNovel)
		//update,PUT是整体更新，PATCH是部分更新
		novels.PUT("/:id", s.updateNovel)
		//delete
		novels.DELETE("/:id", s.deleteNovel)

	}

	users := s.router.Group("/api/v1/users")
	{
		//TODO，RESTFUL API

	}

	events := s.router.Group("/api/v1/events")
	{
		events.GET("/listen",s.streamEvents)
	}

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
		CreatedAt    string `json:"createAt" binding:"omitempty"`
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
		ID           string `json:"id" binding:"required"`
		Author       string `json:"author" binding:"required"`
		StoryOutline string `json:"storyOutline" binding:"required"`
		Subsections  string `json:"subsections" binding:"required"`
		Characters   string `json:"characters" binding:"required"`
		Items        string `json:"items" binding:"required"`
		TotalScenes  string `json:"totalScenes" binding:"required"`
		CreatedAt    string `json:"createAt" binding:"omitempty"`
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


func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Fabric Gateway API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) start(address string) error{
	log.Printf("🚀 Starting server on %s", address)
	return s.router.Run(address)
}

func (s *Server) formatJSON(data []byte)string {
	var result bytes.Buffer
	//第三个参数字符串的前缀，第四个参数是缩进
	if err :=json.Indent(&result,data,"","    "); err != nil{
		return string(data)
	}
	return result.String()
}