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
	"novel-resource-management/service"
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
		// INSERT_YOUR_CODE
		/*
			这些 timeout 主要控制与 Fabric 网络交互的不同阶段的超时时间，单位是 time.Duration（如 15*time.Second）：

			- client.WithEvaluateTimeout(15*time.Second)
			   「查询 Transaction 的超时时间」
			   用户用 gateway.Evaluate 获取链码数据（只查不写），比如 GET/查询小说等。
			   这个超时，控制的是 read 类型（evaluate）请求，网络里如果 15 秒都没响应会报超时。

			- client.WithEndorseTimeout(30*time.Second)
			   「背书过程的超时时间」
			   用户提交“写入”请求时，Fabric 要让各背书节点模拟执行交易并签名背书。这个过程太慢可能就会超时。
			   这里的 30 秒主要保证集群背书时网络分布较慢时也能等一会。

			- client.WithSubmitTimeout(15*time.Second)
			   「提交到排序服务（orderer）的超时时间」
			   交易背书好后，需发给 orderer 排序。这个主流程较快（一般不需要很长），15 秒足够。
			
			- client.WithCommitStatusTimeout(2*time.Minute)
			   「等待区块最终提交的超时时间」
			   你的交易被排序后，会入账并要等 peer 节点确认提交。
			   这个阶段可能等得最久，因为涉及区块打包、排序网络广播等，所以时间拉长到 2 分钟。
		*/
		client.WithEvaluateTimeout(15*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(15*time.Second),
		client.WithCommitStatusTimeout(2*time.Minute),
	)
	
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gateWay.Close()

	// 创建事件服务并启动事件监听
	eventService := service.NewEventService(gateWay)
	ctx := context.Background()

	// 启动事件监听（在后台goroutine中运行），如果不启动服务是没有意义的
	go func() {
		log.Println("🎧 启动事件监听器...")
		if err := eventService.StartEventListening(ctx); err != nil {
			log.Printf("❌ 事件监听器启动失败: %v", err)
		}
	}()

	// 启动时自动初始化链码（从MongoDB同步数据）
	go func() {
		log.Println("🔄 开始从MongoDB初始化链码...")
		chaincodeService, err := service.NewChaincodeMigrationService(gateWay)
		if err != nil {
			log.Printf("❌ 创建链码迁移服务失败: %v", err)
			return
		}

		// 等待3秒，确保MongoDB连接稳定
		time.Sleep(3 * time.Second)

		result, err := chaincodeService.InitChaincodeFromMongoDB(ctx)
		if err != nil {
			log.Printf("❌ 链码初始化失败: %v", err)
		} else {
			log.Printf("✅ 链码初始化成功: %s", result)
		}

		// 验证数据一致性
		log.Println("🔍 验证链上链下数据一致性...")
		consistencyReport, err := chaincodeService.ValidateDataConsistency(ctx)
		if err != nil {
			log.Printf("❌ 数据一致性验证失败: %v", err)
		} else {
			if consistencyReport["consistent"].(bool) {
				log.Println("✅ 链上链下数据一致性验证通过")
			} else {
				log.Printf("⚠️ 数据一致性验证发现问题: %+v", consistencyReport["discrepancies"])
			}
		}
	}()

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