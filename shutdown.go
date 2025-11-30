package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// 检查是否以后台模式运行
	if len(os.Args) > 1 && os.Args[1] == "--daemon" {
		fmt.Println("🔄 启动后台监控模式...")
		runDaemonMode()
		return
	}

	// 设置信号处理，支持Ctrl+C优雅关闭
	setupSignalHandlers()

	fmt.Println("🛑 开始关闭完整项目...")
	fmt.Println("📋 关闭流程:")
	fmt.Println("   1️⃣ 停止Novel资源管理系统 (novel-resource-management)")
	fmt.Println("   2️⃣ 关闭Hyperledger Fabric网络 (test-network)")
	fmt.Println("   3️⃣ 清理Docker容器和资源")
	fmt.Println()

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前目录失败: %v", err)
	}
	fmt.Printf("📁 当前工作目录: %s\n", wd)
	fmt.Println()

	// 第一步：停止Novel资源管理系统
	if err := shutdownNovelManagement(); err != nil {
		log.Printf("⚠️ 停止Novel资源管理系统时出现警告: %v", err)
	} else {
		fmt.Println("✅ Novel资源管理系统已停止")
	}
	fmt.Println()

	// 第二步：关闭Fabric网络
	if err := shutdownFabricNetwork(); err != nil {
		log.Printf("⚠️ 关闭Fabric网络时出现警告: %v", err)
	} else {
		fmt.Println("✅ Fabric网络已关闭")
	}
	fmt.Println()

	// 第三步：清理Docker资源
	if err := cleanupDocker(); err != nil {
		log.Printf("⚠️ 清理Docker资源时出现警告: %v", err)
	} else {
		fmt.Println("✅ Docker资源已清理")
	}
	fmt.Println()

	// 验证关闭状态
	if err := verifyShutdown(); err != nil {
		log.Printf("⚠️ 验证关闭状态时出现警告: %v", err)
	} else {
		fmt.Println("✅ 所有服务已成功关闭")
	}

	fmt.Println("🎉 项目关闭完成!")
	fmt.Println("📋 状态总结:")
	fmt.Println("   🛑 Novel API: 已停止")
	fmt.Println("   🛑 Fabric网络: 已关闭")
	fmt.Println("   🛑 Docker容器: 已清理")
	fmt.Println("   🛑 MongoDB: 保持运行（如需关闭请手动执行）")
	fmt.Println()
}

// shutdownNovelManagement 停止Novel资源管理系统
func shutdownNovelManagement() error {
	fmt.Println("🔧 第一步：停止Novel资源管理系统")
	fmt.Println(repeat("=", 50))

	// 检查novel-resource-management目录是否存在
	novelDir := "novel-resource-management"
	if _, err := os.Stat(novelDir); os.IsNotExist(err) {
		return fmt.Errorf("novel-resource-management目录不存在: %s", novelDir)
	}

	// 切换到novel-resource-management目录
	originalDir, _ := os.Getwd()
	if err := os.Chdir(novelDir); err != nil {
		return fmt.Errorf("切换到novel-resource-management目录失败: %v", err)
	}
	defer func() {
		os.Chdir(originalDir) // 返回原目录
	}()

	fmt.Println("📁 当前目录: " + getCurrentDir())

	// 检查是否有docker-compose.yml文件
	if _, err := os.Stat("docker-compose.yml"); os.IsNotExist(err) {
		fmt.Println("⚠️ 未找到docker-compose.yml文件，跳过Docker容器停止")
		return nil
	}

	fmt.Println("🛑 停止Docker Compose服务...")

	// 尝试优雅停止
	stopCmd := exec.Command("docker-compose", "down")
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr

	if err := stopCmd.Run(); err != nil {
		fmt.Printf("⚠️ 优雅停止失败，尝试强制停止: %v\n", err)

		// 强制停止
		forceCmd := exec.Command("docker-compose", "kill")
		forceCmd.Stdout = os.Stdout
		forceCmd.Stderr = os.Stderr
		if err := forceCmd.Run(); err != nil {
			return fmt.Errorf("强制停止Docker Compose服务失败: %v", err)
		}

		// 再次尝试清理
		cleanCmd := exec.Command("docker-compose", "down", "--remove-orphans")
		cleanCmd.Stdout = os.Stdout
		cleanCmd.Stderr = os.Stderr
		if err := cleanCmd.Run(); err != nil {
			fmt.Printf("⚠️ 清理容器失败，继续执行: %v\n", err)
		}
	}

	fmt.Println("✅ Docker Compose服务已停止")

	return nil
}

// shutdownFabricNetwork 关闭Fabric网络
func shutdownFabricNetwork() error {
	fmt.Println("🔧 第二步：关闭Hyperledger Fabric网络")
	fmt.Println(repeat("=", 50))

	// 检查test-network目录是否存在
	testNetworkDir := "test-network"
	if _, err := os.Stat(testNetworkDir); os.IsNotExist(err) {
		return fmt.Errorf("test-network目录不存在: %s", testNetworkDir)
	}

	// 执行Fabric网络关闭命令
	fmt.Println("🛑 执行Fabric网络关闭命令...")

	// 先切换到test-network目录执行
	script := `
cd test-network

echo "=== Stopping Fabric network ==="
./network.sh down

echo "=== Network stopped successfully ==="
`

	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Fabric网络关闭失败: %v", err)
	}

	fmt.Println("✅ Fabric网络已关闭")

	return nil
}

// cleanupDocker 清理Docker资源
func cleanupDocker() error {
	fmt.Println("🔧 第三步：清理Docker资源")
	fmt.Println(repeat("=", 50))

	fmt.Println("🧹 清理停止的容器...")

	// 清理停止的容器
	cleanCmd := exec.Command("docker", "container", "prune", "-f")
	if output, err := cleanCmd.CombinedOutput(); err != nil {
		fmt.Printf("⚠️ 清理容器失败: %v (输出: %s)\n", err, string(output))
	} else {
		fmt.Println("✅ 停止的容器已清理")
	}

	fmt.Println("🧹 清理未使用的镜像...")

	// 清理未使用的镜像（可选，避免频繁下载）
	imageCmd := exec.Command("docker", "image", "prune", "-f")
	if output, err := imageCmd.CombinedOutput(); err != nil {
		fmt.Printf("⚠️ 清理镜像失败: %v (输出: %s)\n", err, string(output))
	} else {
		fmt.Println("✅ 未使用的镜像已清理")
	}

	fmt.Println("🧹 清理未使用的网络...")

	// 清理未使用的网络
	networkCmd := exec.Command("docker", "network", "prune", "-f")
	if output, err := networkCmd.CombinedOutput(); err != nil {
		fmt.Printf("⚠️ 清理网络失败: %v (输出: %s)\n", err, string(output))
	} else {
		fmt.Println("✅ 未使用的网络已清理")
	}

	return nil
}

// verifyShutdown 验证关闭状态
func verifyShutdown() error {
	fmt.Println("🔍 验证关闭状态...")
	fmt.Println(repeat("=", 30))

	// 检查Fabric容器
	fabricContainers := []string{
		"peer0.org1.example.com",
		"peer0.org2.example.com",
		"orderer.example.com",
	}

	for _, container := range fabricContainers {
		checkCmd := exec.Command("docker", "ps", "-a", "--filter", "name="+container, "--format", "{{.Names}}\t{{.Status}}")
		if output, err := checkCmd.CombinedOutput(); err == nil {
			if len(output) > 0 {
				fmt.Printf("⚠️ Fabric容器仍在运行: %s\n", string(output))
			}
		}
	}

	// 检查novel-api容器
	apiCheckCmd := exec.Command("docker", "ps", "-a", "--filter", "name=novel-api", "--format", "{{.Names}}\t{{.Status}}")
	if output, err := apiCheckCmd.CombinedOutput(); err == nil {
		if len(output) > 0 && string(output) != "" {
			fmt.Printf("⚠️ Novel API容器仍在运行: %s\n", string(output))
		}
	}

	// 检查端口占用
	checkPort := func(port string) {
		checkCmd := exec.Command("lsof", "-i", ":"+port)
		if err := checkCmd.Run(); err == nil {
			fmt.Printf("⚠️ 端口 %s 仍在使用中\n", port)
		}
	}

	// 检查常见端口
	portsToCheck := []string{"8080", "7051", "7050", "9051", "9050"}
	for _, port := range portsToCheck {
		checkPort(port)
	}

	// 等待一下确保所有进程完全退出
	time.Sleep(2 * time.Second)

	fmt.Println("✅ 关闭状态验证完成")
	return nil
}

// killProcessesByPort 强制杀死占用指定端口的进程
func killProcessesByPort(ports []string) error {
	for _, port := range ports {
		fmt.Printf("🔪 检查并处理端口 %s...\n", port)

		// 查找占用端口的进程
		findCmd := exec.Command("lsof", "-t", "-i", ":"+port)
		output, err := findCmd.Output()
		if err != nil || len(output) == 0 {
			fmt.Printf("✅ 端口 %s 未被占用\n", port)
			continue
		}

		pids := string(output)
		fmt.Printf("⚠️ 发现端口 %s 被进程占用: %s\n", port, pids)

		// 优雅关闭
		killCmd := exec.Command("kill", "-TERM", pids)
		if err := killCmd.Run(); err != nil {
			fmt.Printf("⚠️ 优雅关闭端口 %s 进程失败: %v\n", port, err)
		}

		// 等待一下
		time.Sleep(2 * time.Second)

		// 检查是否还在运行
		checkCmd := exec.Command("lsof", "-t", "-i", ":"+port)
		if checkOutput, err := checkCmd.Output(); err == nil && len(checkOutput) > 0 {
			fmt.Printf("🔪 强制杀死端口 %s 的进程...\n", port)
			forceCmd := exec.Command("kill", "-9", string(checkOutput))
			if err := forceCmd.Run(); err != nil {
				fmt.Printf("⚠️ 强制杀死端口 %s 进程失败: %v\n", port, err)
			}
		}

		fmt.Printf("✅ 端口 %s 已释放\n", port)
	}

	return nil
}

// forceShutdown 强制关闭所有相关进程
func forceShutdown() error {
	fmt.Println("🔪 执行强制关闭模式...")

	// 强制杀死占用关键端口的进程
	ports := []string{"8080", "7051", "7050", "9051", "9050", "27017"}
	if err := killProcessesByPort(ports); err != nil {
		return fmt.Errorf("强制关闭进程失败: %v", err)
	}

	fmt.Println("✅ 强制关闭完成")
	return nil
}

// getCurrentDir 获取当前工作目录
func getCurrentDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "unknown"
}

// repeat 重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// runDaemonMode 后台监控模式
func runDaemonMode() {
	fmt.Println("👻 进入后台守护模式，监控项目服务状态...")
	fmt.Println("📋 功能:")
	fmt.Println("   🔄 自动检测服务状态")
	fmt.Println("   🚨 检测异常时自动关闭")
	fmt.Println("   📊 定期报告服务状态")
	fmt.Println("   ⏹️  Ctrl+C 停止守护进程")
	fmt.Println()

	// 创建守护进程的PID文件
	pidFile := "shutdown-daemon.pid"
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		log.Printf("⚠️ 创建PID文件失败: %v", err)
	} else {
		fmt.Printf("✅ 守护进程PID: %d (已写入 %s)\n", os.Getpid(), pidFile)
	}

	defer os.Remove(pidFile)

	// 监控循环
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := monitorServices(); err != nil {
				log.Printf("⚠️ 监控服务时出错: %v", err)
			}
		case sig := <-getSignalChan():
			fmt.Printf("\n收到信号 %v，停止守护进程...\n", sig)
			fmt.Println("🛑 后台守护进程已停止")
			return
		}
	}
}

// monitorServices 监控服务状态
func monitorServices() error {
	fmt.Printf("🔍 [%s] 检查服务状态...\n", time.Now().Format("15:04:05"))

	// 检查关键端口
	portsToCheck := map[string]string{
		"8080": "Novel API",
		"7051": "Fabric Peer1",
		"7050": "Fabric Peer2",
		"9051": "Fabric Orderer1",
		"9050": "Fabric Orderer2",
	}

	var activeServices []string
	var failedServices []string

	for port, serviceName := range portsToCheck {
		if isPortActive(port) {
			activeServices = append(activeServices, serviceName)
		} else {
			failedServices = append(failedServices, serviceName)
		}
	}

	// 报告状态
	if len(activeServices) > 0 {
		fmt.Printf("✅ 运行中的服务: %v\n", activeServices)
	}

	if len(failedServices) > 0 {
		fmt.Printf("❌ 停止的服务: %v\n", failedServices)
	}

	// 如果所有服务都停止了，自动退出
	if len(activeServices) == 0 {
		fmt.Println("🎉 所有服务已停止，守护进程退出")
		os.Exit(0)
	}

	// 检查是否有异常状态
	if err := checkAbnormalStatus(); err != nil {
		fmt.Printf("🚨 检测到异常状态: %v\n", err)
		fmt.Println("⚠️ 建议手动执行: go run shutdown.go")
	}

	return nil
}

// isPortActive 检查端口是否活跃
func isPortActive(port string) bool {
	cmd := exec.Command("lsof", "-t", "-i", ":"+port)
	output, err := cmd.Output()
	return err == nil && len(output) > 0
}

// checkAbnormalStatus 检查异常状态
func checkAbnormalStatus() error {
	// 检查Docker容器状态异常
	cmd := exec.Command("docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.Names}}\t{{.Status}}")
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" &&
			   (strings.Contains(line, "novel") || strings.Contains(line, "peer") || strings.Contains(line, "orderer")) {
				return fmt.Errorf("容器异常退出: %s", line)
			}
		}
	}

	return nil
}

// getSignalChan 获取信号通道
func getSignalChan() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}

// 信号处理函数，支持优雅关闭
func setupSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n收到信号 %v，开始优雅关闭...\n", sig)

		// 执行快速关闭流程
		fmt.Println("🚨 执行快速关闭...")

		// 强制关闭所有进程
		if err := forceShutdown(); err != nil {
			fmt.Printf("⚠️ 强制关闭失败: %v\n", err)
		}

		os.Exit(0)
	}()
}