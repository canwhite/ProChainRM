package api 

import(
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

type Server struct{
	router *gin.Engine
	novelService *service.NovelService
	creditService *service.UserCreditService
	eventService *service.EventService
	network *client.Network
}

//create new service interface
func NewServer(gateway *client.Gateway) *Server{
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
		router: gin.GetDefault(),
		novelService: service.NewNovelService(gateway),
		creditService: service.NewUserCreditService(gateway),
		eventService: service.NewEventService(gateway),
		network: network,
	}

	server.setupRoutes()
	return server
} 

//方法指示器
func (s *Server)setupRoutes(){
	//先接路由，再接方法
	s.router.GET("/health", s.healthCheck)


}

func (s *Server)healthCheck(c *gin.Context){
	c.JSON(http.StatusOK,gin.H{
		"status":  "ok",
		"message": "Fabric Gateway API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}



