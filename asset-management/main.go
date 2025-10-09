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
		// log.Fatalf 是 Go 语言标准库 log 包中的一个函数。它的作用是先按照指定的格式输出一条日志（类似 fmt.Printf），
		// 然后调用 os.Exit(1) 终止程序运行。也就是说，log.Fatalf 会输出错误信息并让程序异常退出，常用于遇到致命错误时的处理。
		// 例如：log.Fatalf("Failed to create gRPC connection: %v", err) 会输出错误信息并退出程序。
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

	/*
		所以，缓冲区的作用就是：让你可以先放，等会儿再取，不用卡着等。
	*/

	// INSERT_YOUR_CODE
	/*
		这句代码：
			sigChan := make(chan os.Signal, 1)

		意思是：创建一个“带缓冲区”的通道（channel），类型是 os.Signal，缓冲区大小为 1。

		详细解释：
		- chan os.Signal：表示这个通道里只能传递 os.Signal 类型的值（比如操作系统的中断信号）。
		- make(chan os.Signal, 1)：用 make 创建一个带 1 个缓冲槽的 channel。这样，最多可以有 1 个信号被发送到通道里而不会阻塞发送方。

		为什么要这样用？
		- 在 Go 里，通道（channel）是用来在 goroutine 之间传递数据的。
		- signal.Notify(sigChan, ...) 会把收到的操作系统信号（如 Ctrl+C）发送到 sigChan 里。
		- 缓冲区大小为 1，意味着即使主 goroutine还没来得及处理信号，最多也只会存一个信号，不会丢失。

		常见场景：
		- 用于优雅关闭服务（graceful shutdown），比如 Web 服务器收到 SIGINT/SIGTERM 信号后，先做清理再退出。

		小结：
		- make(chan os.Signal, 1) 就是造了一个“信号邮箱”，能暂存 1 个信号，方便主程序检测和处理。
	*/
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
	// <-sigChan //don‘t need this line
	log.Println("🛑 Shutting down gracefully...")
	
	// Allow time for cleanup
	time.Sleep(1 * time.Second)
	log.Println("✅ Server stopped")
}