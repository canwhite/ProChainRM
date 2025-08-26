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

	"github.com/gin-gonic/gin"
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
	s.router.GET("/health", s.healthCheck)

	// Assets API
	assets := s.router.Group("/api/v1/assets")
	{
		assets.GET("", s.getAllAssets)
		assets.GET("/:id", s.getAsset)
		assets.POST("", s.createAsset)
		assets.PUT("/:id", s.updateAsset)
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
func (s *Server) createAsset(c *gin.Context) {
	var req struct {
		ID             string `json:"id" binding:"required"`
		Color          string `json:"color" binding:"required"`
		Size           string `json:"size" binding:"required"`
		Owner          string `json:"owner" binding:"required"`
		AppraisedValue string `json:"appraisedValue" binding:"required"`
	}

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
	id := c.Param("id")
	if id == "" {
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
	if err := json.Indent(&result, data, "", "  "); err != nil {
		return string(data)
	}
	return result.String()
}