package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"sdk-go/network"
	"sdk-go/service"
)

func main() {
	fmt.Println("🚀 Starting Asset Management Client...")

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
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection))
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gateway.Close()

	// Create services
	assetService := service.NewAssetService(gateway)
	eventService := service.NewEventService(gateway)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// 该行创建了一个带缓冲区的信号通道，用于接收操作系统发来的中断（如Ctrl+C）或终止信号，实现优雅关闭程序
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start event listening
	fmt.Println("\n🎧 Starting Event Listener...")
	err = eventService.StartEventListening(ctx)
	if err != nil {
		log.Printf("Failed to start event listening: %v", err)
	}

	// Give event listener time to start
	time.Sleep(2 * time.Second)

	// Demonstrate chaincode operations that will trigger events
	fmt.Println("\n📋 Asset Management Operations:")

	// Create a new asset (will trigger CreateAsset event)
	err = assetService.CreateAsset("asset1", "purple", "8", "Alice", "900")
	if err != nil {
		log.Printf("Failed to create asset: %v", err)
	}

	// Update asset (will trigger UpdateAsset event)
	err = assetService.UpdateAsset("asset1", "blue", "10", "Alice", "1200")
	if err != nil {
		log.Printf("Failed to update asset: %v", err)
	}

	// Transfer asset (will trigger TransferAsset event)
	err = assetService.TransferAsset("asset1", "Bob")
	if err != nil {
		log.Printf("Failed to transfer asset: %v", err)
	}

	// Query assets
	allAssets, err := assetService.GetAllAssets()
	if err != nil {
		log.Printf("Failed to get all assets: %v", err)
	} else {
		fmt.Printf("All assets: %s\n", allAssets)
	}

	// Keep running to receive events
	fmt.Println("\n🔍 Listening for events... Press Ctrl+C to stop")
	
	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\n🛑 Shutting down gracefully...")
	fmt.Println("\n✅ Client stopped!")
}