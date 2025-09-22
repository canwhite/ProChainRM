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
		novelService:  service.NewNovelService(gateway),
		creditService: service.NewUserCreditService(gateway),
		eventService:  service.NewEventService(gateway),
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
		//TODO，RESTFUL API
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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//参数顺序
	//id, author, storyOutline, subsections, characters, items, totalScenes string
	if err = s.novelService.CreateNovel(req.ID, req.Author, req.StoryOutline, req.Subsections, req.Characters, req.Items, req.TotalScenes); err != nil {
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

	if err = s.novelService.UpdateNovel(id, req.Author, req.StoryOutline, req.Subsections, req.Characters, req.Items, req.TotalScenes); err != nil {
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

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Fabric Gateway API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}
