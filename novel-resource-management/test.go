package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"

	"novel-resource-management/network"
	"novel-resource-management/service"
)

//1.建立网路连接 => 2.创建身份和签名 => 3.创建网关 => 4.创建服务实例 
func main() {
	// 1. 建立网络连接
	fmt.Println("正在建立网络连接...")
	grpcConnection, err := network.NewGrpcConnection()
	if err != nil {
		log.Fatalf("无法建立gRPC连接: %v", err)
	}
	defer grpcConnection.Close()

	// 2. 创建身份和签名
	id := network.NewIdentity() 
	sign := network.NewSign()

	// 3. 创建网关
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConnection),
	)
	if err != nil {
		log.Fatalf("无法连接到网关: %v", err)
	}
	defer gateway.Close()

	// 4. 创建服务实例
	novelService, err := service.NewNovelService(gateway)
	if err != nil {
		log.Fatalf("无法创建小说服务: %v", err)
	}

	userCreditService, err := service.NewUserCreditService(gateway)
	if err != nil {
		log.Fatalf("无法创建用户积分服务: %v", err)
	}

	// 5. 测试小说服务
	fmt.Println("\n=== 测试小说服务 ===")
	testNovelService(novelService)

	// 6. 测试用户积分服务
	fmt.Println("\n=== 测试用户积分服务 ===")
	testUserCreditService(userCreditService)

	fmt.Println("\n测试完成！")
}

func testNovelService(novelService *service.NovelService) {
	// 创建小说
	fmt.Println("创建小说...")
	err := novelService.CreateNovel(
		"novel1",
		"张三",
		"这是一个科幻小说的故事大纲",
		"第一章: 起始\n第二章: 发展\n第三章: 高潮",
		"主角: 李明\n配角: 王小红",
		"道具1: 时光机\n道具2: 激光枪",
		"100",
	)
	if err != nil {
		fmt.Printf("创建小说失败: %v\n", err)
		return
	}
	fmt.Println("小说创建成功")

	// 读取小说
	fmt.Println("读取小说信息...")
	novelData, err := novelService.ReadNovel("novel1")
	if err != nil {
		fmt.Printf("读取小说失败: %v\n", err)
		return
	}
	fmt.Printf("小说信息: %+v\n", novelData)

	// 获取所有小说
	fmt.Println("获取所有小说...")
	allNovels, err := novelService.GetAllNovels()
	if err != nil {
		fmt.Printf("获取所有小说失败: %v\n", err)
		return
	}
	fmt.Printf("共有 %d 本小说\n", len(allNovels))
	for i, novel := range allNovels {
		fmt.Printf("小说 %d: %+v\n", i+1, novel)
	}

	// 更新小说
	fmt.Println("更新小说...")
	err = novelService.UpdateNovel(
		"novel1",
		"张三（更新）",
		"更新后的故事大纲",
		"第一章: 起始（更新）\n第二章: 发展（更新）",
		"主角: 李明（更新）",
		"道具1: 时光机（更新）\n道具2: 激光枪（更新）",
		"120",
	)
	if err != nil {
		fmt.Printf("更新小说失败: %v\n", err)
		return
	}
	fmt.Println("小说更新成功")

	// 删除小说
	fmt.Println("删除小说...")
	err = novelService.DeleteNovel("novel1")
	if err != nil {
		fmt.Printf("删除小说失败: %v\n", err)
		return
	}
	fmt.Println("小说删除成功")
}

func testUserCreditService(userCreditService *service.UserCreditService) {
	// 创建用户积分
	fmt.Println("创建用户积分...")
	err := service.CreateUserCredit(userCreditService, "user1", 1000, 500, 1500)
	if err != nil {
		fmt.Printf("创建用户积分失败: %v\n", err)
		return
	}
	fmt.Println("用户积分创建成功")

	// 读取用户积分
	fmt.Println("读取用户积分...")
	userCredit, err := service.ReadUserCredit(userCreditService, "user1")
	if err != nil {
		fmt.Printf("读取用户积分失败: %v\n", err)
		return
	}
	fmt.Printf("用户积分信息: %+v\n", userCredit)

	// 获取所有用户积分
	fmt.Println("获取所有用户积分...")
	allUserCredits, err := service.GetAllUserCredits(userCreditService)
	if err != nil {
		fmt.Printf("获取所有用户积分失败: %v\n", err)
		return
	}
	fmt.Printf("共有 %d 个用户积分记录\n", len(allUserCredits))
	for i, credit := range allUserCredits {
		fmt.Printf("用户积分 %d: %+v\n", i+1, credit)
	}

	// 更新用户积分
	fmt.Println("更新用户积分...")
	err = service.UpdateUserCredit(userCreditService, "user1", 1200, 700, 1900)
	if err != nil {
		fmt.Printf("更新用户积分失败: %v\n", err)
		return
	}
	fmt.Println("用户积分更新成功")

	// 删除用户积分
	fmt.Println("删除用户积分...")
	err = service.DeleteUserCredit(userCreditService, "user1")
	if err != nil {
		fmt.Printf("删除用户积分失败: %v\n", err)
		return
	}
	fmt.Println("用户积分删除成功")
}