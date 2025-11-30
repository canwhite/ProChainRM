package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("🚀 开始完整项目部署...")
	fmt.Println("📋 部署流程:")
	fmt.Println("   1️⃣ 部署Hyperledger Fabric网络 (test-network)")
	fmt.Println("   2️⃣ 部署Novel资源管理系统 (novel-resource-management)")
	fmt.Println()

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前目录失败: %v", err)
	}
	fmt.Printf("📁 当前工作目录: %s\n", wd)
	fmt.Println()

	// 第一步：部署Fabric网络
	if err := deployFabricNetwork(); err != nil {
		log.Fatalf("❌ Fabric网络部署失败: %v", err)
	}
	fmt.Println("✅ Fabric网络部署完成!")
	fmt.Println()

	// 等待网络稳定
	fmt.Println("⏳ 等待Fabric网络稳定...")
	time.Sleep(5 * time.Second)

	// 第二步：部署Novel资源管理系统
	if err := deployNovelManagement(); err != nil {
		log.Fatalf("❌ Novel资源管理系统部署失败: %v", err)
	}
	fmt.Println("✅ Novel资源管理系统部署完成!")
	fmt.Println()

	// 部署成功信息
	fmt.Println("🎉 完整项目部署成功!")
	fmt.Println("📋 服务信息:")
	fmt.Println("   🔗 Fabric网络: test-network目录")
	fmt.Println("   🌐 Novel API: http://localhost:8080")
	fmt.Println("   💚 健康检查: http://localhost:8080/health")
	fmt.Println("   📊 API文档: http://localhost:8080/swagger")
	fmt.Println()
}

// deployFabricNetwork 部署Hyperledger Fabric网络
func deployFabricNetwork() error {
	fmt.Println("🔧 第一步：部署Hyperledger Fabric网络")
	fmt.Println(repeat("=", 50))

	// 检查test-network目录是否存在
	testNetworkDir := "test-network"
	if _, err := os.Stat(testNetworkDir); os.IsNotExist(err) {
		return fmt.Errorf("test-network目录不存在: %s", testNetworkDir)
	}

	// Fabric部署脚本内容（基于之前的分析，包含修复时序问题的版本）
	script := `
# 先切换到test-network目录并保持在其中
cd test-network

echo "=== Step 1: Stopping previous network ==="
./network.sh down

echo ""
echo "=== Step 2: Starting network ==="
./network.sh up

echo ""
echo "=== Step 3: Creating channel ==="
./network.sh createChannel

echo ""
echo "=== Step 4: Setting environment and deploying chaincode ==="
source set-env.sh
./network.sh deployCC -ccn novel-basic -ccp ../novel-resource-events -ccl go -ccv 1.0 -cci InitLedger -ccep 'OR("Org1MSP.member","Org2MSP.member")'

echo ""
echo "=== Step 5: Waiting for chaincode to be ready ==="
sleep 10

echo ""
echo "=== Step 6: Querying chaincode ==="
peer chaincode query -C mychannel -n novel-basic -c '{"function":"GetAllNovels","Args":[]}'

echo ""
echo "=== Fabric network deployment completed ==="
`

	// 执行Fabric部署脚本
	fmt.Println("🔨 执行Fabric网络部署脚本...")
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Fabric部署脚本执行失败: %v", err)
	}

	return nil
}

// deployNovelManagement 部署Novel资源管理系统
func deployNovelManagement() error {
	fmt.Println("🔧 第二步：部署Novel资源管理系统")
	fmt.Println(repeat("=", 50))

	// 检查novel-resource-management目录是否存在
	novelDir := "novel-resource-management"
	if _, err := os.Stat(novelDir); os.IsNotExist(err) {
		return fmt.Errorf("novel-resource-management目录不存在: %s", novelDir)
	}

	// 切换到novel-resource-management目录
	if err := os.Chdir(novelDir); err != nil {
		return fmt.Errorf("切换到novel-resource-management目录失败: %v", err)
	}
	defer func() {
		// 返回原目录
		if err := os.Chdir(".."); err != nil {
			log.Printf("⚠️ 返回原目录失败: %v", err)
		}
	}()

	fmt.Println("📁 当前目录: " + getCurrentDir())

	// 检查必要的文件
	requiredFiles := []string{"scripts/deploy.go", "docker-compose.yml"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("必要文件不存在: %s", file)
		}
		fmt.Printf("✅ 找到文件: %s\n", file)
	}

	// 执行novel-resource-management部署脚本
	fmt.Println("🔨 编译并执行Novel资源管理系统部署脚本...")

	// 进入scripts目录执行部署（确保相对路径正确）
	if err := os.Chdir("scripts"); err != nil {
		return fmt.Errorf("切换到scripts目录失败: %v", err)
	}
	defer func() {
		// 返回novel-resource-management目录
		if err := os.Chdir(".."); err != nil {
			log.Printf("⚠️ 返回novel-resource-management目录失败: %v", err)
		}
	}()

	// 编译并执行部署脚本
	compileCmd := exec.Command("go", "run", "deploy.go")
	compileCmd.Stdout = os.Stdout
	compileCmd.Stderr = os.Stderr
	// 设置环境变量，确保可以找到.env文件
	compileCmd.Env = append(os.Environ(), "ENV_PATH=../.env")

	if err := compileCmd.Run(); err != nil {
		return fmt.Errorf("Novel资源管理系统部署失败: %v", err)
	}

	// 验证服务启动
	fmt.Println("🔍 验证Novel API服务...")
	if err := verifyNovelAPIService(); err != nil {
		return fmt.Errorf("Novel API服务验证失败: %v", err)
	}

	return nil
}

// verifyNovelAPIService 验证Novel API服务是否正常运行
func verifyNovelAPIService() error {
	fmt.Println("🏥 执行健康检查...")

	// 等待服务启动
	for i := 0; i < 30; i++ {
		healthCmd := exec.Command("curl", "-s", "http://localhost:8080/health")
		if output, err := healthCmd.CombinedOutput(); err == nil {
			outputStr := strings.TrimSpace(string(output))
			if strings.Contains(outputStr, "ok") || strings.Contains(outputStr, "OK") ||
			   strings.Contains(outputStr, "healthy") || strings.Contains(outputStr, "success") {
				fmt.Println("✅ Novel API服务健康检查通过")
				return nil
			}
		}

		fmt.Printf("⏳ 等待服务就绪... (%d/30)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("Novel API服务健康检查超时")
}

// getCurrentDir 获取当前工作目录
func getCurrentDir() string {
	if dir, err := os.Getwd(); err == nil {
		return filepath.Base(dir)
	}
	return "unknown"
}

// 辅助函数：重复字符串
func repeat(s string, count int) string {
	return strings.Repeat(s, count)
}