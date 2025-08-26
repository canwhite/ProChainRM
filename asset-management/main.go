package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"sdk-go/api"
	"sdk-go/network"
)

func main() {
	// Create gRPC client connection
	clientConnection, err := network.NewGrpcConnection()
	if err != nil {
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer clientConnection.Close()

	// Create identity and signing
	id := network.NewIdentity()
	sign := network.NewSign()

	// Create gateway connection
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gateway.Close()

	// Create HTTP server
	server := api.NewServer(gateway)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		log.Println("🚀 Starting Fabric Gateway API Server...")
		log.Println("📋 Available endpoints:")
		log.Println("  GET    /api/v1/assets")
		log.Println("  GET    /api/v1/assets/:id")
		log.Println("  POST   /api/v1/assets")
		log.Println("  PUT    /api/v1/assets/:id")
		log.Println("  PATCH  /api/v1/assets/:id/transfer")
		log.Println("  DELETE /api/v1/assets/:id")
		log.Println("  GET    /api/v1/events/listen")
		log.Println("  GET    /health")
		
		if err := server.Start(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutting down gracefully...")
	
	// Allow time for cleanup
	time.Sleep(1 * time.Second)
	log.Println("✅ Server stopped")
}