package main

import (
	"fmt"
	"log"

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

	// Create asset service
	assetService := service.NewAssetService(gateway)

	// Demonstrate chaincode operations
	fmt.Println("\n📋 Asset Management Operations:")

	// Query all assets
	allAssets, err := assetService.GetAllAssets()
	if err != nil {
		log.Printf("Failed to get all assets: %v", err)
	} else {
		fmt.Printf("All assets: %s\n", allAssets)
	}

	// Read a specific asset
	asset, err := assetService.ReadAsset("asset1")
	if err != nil {
		log.Printf("Failed to read asset: %v", err)
	} else {
		fmt.Printf("Asset details: %s\n", asset)
	}

	// Create a new asset
	err = assetService.CreateAsset("asset7", "purple", "8", "Alice", "900")
	if err != nil {
		log.Printf("Failed to create asset: %v", err)
	}

	// Query again to see new asset
	allAssets, err = assetService.GetAllAssets()
	if err != nil {
		log.Printf("Failed to get all assets: %v", err)
	} else {
		fmt.Printf("Updated assets: %s\n", allAssets)
	}

	fmt.Println("\n✅ Operations completed!")
}