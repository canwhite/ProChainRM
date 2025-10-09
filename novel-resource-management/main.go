package main

import(
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"novel-resource-management/network"
	"novel-resource-management/api"
)


func main(){
	clientConnection,err := network.NewGrpcConnection()
	if err != nil{
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer clientConnection.Close()

	id := network.NewIdentity()
	sign := network.NewSign()

	// INSERT_YOUR_CODE
	// 这里的gateWay（其实应该叫gateway，变量名建议统一）是 *client.Gateway 类型的指针，代表 Fabric Gateway 客户端的连接对象，不是地址字符串。
	// 它不是返回网络地址，而是一个已经建立好连接、可以用于后续链码交互的网关客户端对象。
	// 你可以用它来获取 network/channel、提交交易、查询等。
	gateWay,err := client.Connect(
		id,
		client.WithSign(sign),
		//hash and connect,确实应该先有hash，再有connect
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		//几个timeout
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gateWay.Close()

	server := api.NewServer(gateWay)

	//handle gracefully shutdown 
	sigChan := make(chan os.Signal,1)
	// INSERT_YOUR_CODE
	/*
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) 这行代码的第二个和第三个参数分别是 syscall.SIGINT 和 syscall.SIGTERM。

		- syscall.SIGINT：表示“中断信号”，通常是用户在终端按下 Ctrl+C 时，操作系统发送给进程的信号。收到这个信号后，程序可以选择优雅地退出。
		- syscall.SIGTERM：表示“终止信号”，是操作系统或其他进程请求当前进程终止时发送的信号。它是让程序“正常退出”的标准信号，程序可以捕获并做清理工作。

		这两个参数的作用是告诉 signal.Notify：当进程收到 SIGINT 或 SIGTERM 信号时，把信号发送到 sigChan 这个 channel 里。这样主程序就能感知到“要退出了”，从而做一些优雅关闭的操作（如资源清理、日志记录等）。
	*/
	signal.Notify(sigChan,syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		log.Println("🚀 Starting Fabric Gateway API Server...")
		log.Println("📋 Available endpoints:")
		log.Println("  GET    /api/v1/novels")
		log.Println("  GET    /api/v1/novels/:id")
		log.Println("  POST   /api/v1/novels")
		log.Println("  PUT    /api/v1/novels/:id")
		log.Println("  DELETE /api/v1/novels/:id")
		log.Println("  GET    /api/v1/users")
		log.Println("  GET    /api/v1/users/:id")
		log.Println("  POST   /api/v1/users")
		log.Println("  PUT    /api/v1/users/:id")
		log.Println("  DELETE /api/v1/users/:id")
		log.Println("  GET    /api/v1/events/listen")
		log.Println("  GET    /health")
		
		if err := server.Start(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()


	// Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutting down gracefully...")

	//防止泄漏
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用Shutdown方法清理gin资源
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Warning: graceful shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped")
}