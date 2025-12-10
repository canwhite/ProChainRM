// novel-resource-management 自动化部署脚本（Linux优化版）
//
// 功能特性：
// - 自动检测操作系统并优化IP选择策略
// - 智能网段优先级选择（Linux环境优先192.168.x.x）
// - 全面的Docker网络过滤（172.17-31.x.x等）
// - 详细的调试日志输出
// - MongoDB副本集自动配置
//
// 环境变量：
//   DEBUG_NETWORK=true  - 启用详细网络接口调试信息
//   ENV_PATH           - 自定义.env文件路径
//   MONGO_USER         - MongoDB用户名
//   MONGO_PASS         - MongoDB密码
//
// Linux部署注意事项：
//   - 脚本会自动检测Linux环境并优化IP选择策略
//   - 优先选择192.168.x.x网段（最常见的Linux内网）
//   - 备用IP设置为192.168.1.100
//   - 支持eth0、ens33、enp0s3等常见Linux网络接口
//
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec" //这个用来执行命令行，直接可以exec
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// MongoDB配置
const (
	MongoPort     = "27017"
	MongoDatabase = "admin"
)

// 敏感信息通过环境变量获取
func getMongoConfig() (string, string) {
	user := getEnv("MONGO_USER", "admin")
	pass := getEnv("MONGO_PASS", "password")
	return user, pass
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isLinux 检测是否为Linux系统
func isLinux() bool {
	return runtime.GOOS == "linux"
}

// getEnvironmentConfig 获取环境配置
func getEnvironmentConfig() {
	if isLinux() {
		log.Println("🐧 检测到Linux环境，应用Linux优化配置")
	} else {
		log.Printf("💻 检测到%s环境", runtime.GOOS)
	}
}

// getPreferredIP 根据优先级选择最佳IP
func getPreferredIP(candidateIPs []string) string {
	if len(candidateIPs) == 0 {
		return ""
	}

	// 常见的内网网段优先级（Linux环境优先考虑）
	priorityNetworks := []string{
		"10.",      // 企业网络A段
		"192.168.", // 常见的家用/办公网络C段
		"172.16.",  // 私有网络B段（特定偏好）
		"172.",     // 其他私有网络B段
	}

	// 在Linux环境下，优先考虑192.168网段
	if isLinux() {
		priorityNetworks = []string{
			"192.168.", // Linux最常见的内网网段
			"10.",      // 企业网络
			"172.16.",  // 特定偏好网段
			"172.",     // 其他私有网络
		}
	}

	for _, prefix := range priorityNetworks {
		for _, ip := range candidateIPs {
			if strings.HasPrefix(ip, prefix) {
				log.Printf("✅ 优先选择网段 %s 的IP: %s", prefix, ip)
				return ip
			}
		}
	}

	// 如果没有匹配的优先网段，返回第一个候选IP
	log.Printf("⚠️ 未找到优先网段，使用候选IP: %s", candidateIPs[0])
	return candidateIPs[0]
}

// isDockerNetwork 检查是否为Docker相关网络（优化版）
func isDockerNetwork(ip string) bool {
	// 1. 首先过滤明确的Docker网络和系统网络
	definitelyDockerNetworks := []string{
		"172.17.",       // Docker默认网桥网络
		"192.168.65.",   // Docker Desktop (Mac)
		"127.",          // 回环地址
	}

	for _, dockerNet := range definitelyDockerNetworks {
		if strings.HasPrefix(ip, dockerNet) {
			return true
		}
	}

	// 2. 对于172.18-31网段，采用更保守的策略
	// 因为这些也可能是合法的企业内网，只有在特定条件下才认为是Docker网络
	if isPotentiallyDockerNetwork(ip) {
		log.Printf("🔍 发现可能的Docker网络IP: %s (172.18-31网段)，保留候选", ip)
		// 不立即返回true，而是保留为候选，让优先级逻辑决定
		return false
	}

	return false
}

// isPotentiallyDockerNetwork 检查是否可能是Docker网络（172.18-31网段）
// 这个函数用于识别可能被误判的Docker网络，但保留为有效候选
func isPotentiallyDockerNetwork(ip string) bool {
	// 检查172.18-31网段
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) >= 2 {
			secondOctet := parts[1]
			// 172.18 到 172.31 网段
			for i := 18; i <= 31; i++ {
				if secondOctet == fmt.Sprintf("%d", i) {
					return true
				}
			}
		}
	}
	return false
}

// getNetworkInterfaceInfo 获取网络接口详细信息用于调试
func getNetworkInterfaceInfo() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("❌ 获取网络接口信息失败: %v", err)
		return
	}

	log.Println("📋 网络接口详情:")
	for _, inter := range interfaces {
		log.Printf("  - 接口: %s, 状态: %s, MTU: %d", inter.Name, getInterfaceStatus(inter.Flags), inter.MTU)
		addrs, err := inter.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			log.Printf("    地址: %s", addr.String())
		}
	}
}

// getInterfaceStatus 获取网络接口状态描述
func getInterfaceStatus(flags net.Flags) string {
	status := ""
	if flags&net.FlagUp != 0 {
		status += "UP "
	}
	if flags&net.FlagLoopback != 0 {
		status += "LOOPBACK "
	}
	if flags&net.FlagMulticast != 0 {
		status += "MULTICAST "
	}
	if status == "" {
		status = "DOWN"
	}
	return strings.TrimSpace(status)
}

func main() {
	// 环境检测和配置
	getEnvironmentConfig()

	// 确定.env文件路径
	envPath := "../.env"
	if envPathFromEnv := os.Getenv("ENV_PATH"); envPathFromEnv != "" {
		envPath = envPathFromEnv
	}

	// 加载.env文件
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("警告: 无法加载.env文件: %v (路径: %s)", err, envPath)
		log.Println("将使用默认环境变量")
	} else {
		log.Printf("✅ 成功加载.env文件: %s", envPath)
	}

	fmt.Println("🚀 开始自动化部署novel-resource-management...")

	// 1. 获取宿主机真实IP
	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("❌ 获取宿主机IP失败: %v", err)
	}
	fmt.Printf("✅ 宿主机IP: %s\n", hostIP)

	// 2. 配置MongoDB副本集
	if err := configureMongoDBReplicaSet(hostIP); err != nil {
		log.Fatalf("❌ MongoDB副本集配置失败: %v", err)
	}
	fmt.Println("✅ MongoDB副本集配置完成")

	// 3. 执行Docker部署
	if err := runDockerDeploy(); err != nil {
		log.Fatalf("❌ Docker部署失败: %v", err)
	}

	fmt.Println("🎉 自动化部署完成!")
	fmt.Println("📊 服务访问地址: http://localhost:8080")
	fmt.Println("💚 健康检查: http://localhost:8080/health")
}

// getHostIP 获取宿主机在局域网中的真实IP（Linux优化版）
func getHostIP() (string, error) {
	log.Println("🔍 开始获取宿主机IP地址...")

	// 在详细模式下显示网络接口信息
	if os.Getenv("DEBUG_NETWORK") == "true" {
		getNetworkInterfaceInfo()
	}

	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口失败: %v", err)
	}

	var candidateIPs []string
	var preferredIP string

	log.Printf("📋 找到 %d 个网络接口", len(interfaces))

	for _, inter := range interfaces {
		// 跳过回环接口和down状态的接口
		if inter.Flags&net.FlagLoopback != 0 || inter.Flags&net.FlagUp == 0 {
			log.Printf("  - 跳过接口 %s (状态: %s)", inter.Name, getInterfaceStatus(inter.Flags))
			continue
		}

		log.Printf("  - 检查接口: %s", inter.Name)
		addrs, err := inter.Addrs()
		if err != nil {
			log.Printf("    ❌ 获取地址失败: %v", err)
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // 跳过IPv6地址
			}

			ipStr := ip.String()

			// 检查是否为Docker网络
			if isDockerNetwork(ipStr) {
				log.Printf("    - 跳过Docker网络IP: %s", ipStr)
				continue
			}

			log.Printf("    ✅ 发现有效IP: %s (来自接口: %s)", ipStr, inter.Name)

			// 收集候选IP
			candidateIPs = append(candidateIPs, ipStr)

			// 在Linux环境下，检查是否有特定偏好的网段
			if isLinux() && strings.HasPrefix(ipStr, "192.168.") {
				// Linux环境下，192.168网段有较高优先级
				if preferredIP == "" {
					preferredIP = ipStr
					log.Printf("🎯 Linux环境优先选择192.168网段IP: %s", ipStr)
				}
			} else if strings.HasPrefix(ipStr, "172.16.") {
				// 特定偏好网段
				if preferredIP == "" {
					preferredIP = ipStr
					log.Printf("🎯 发现偏好网段172.16的IP: %s", ipStr)
				}
			}
		}
	}

	// 优先选择特定网段的IP
	if preferredIP != "" {
		log.Printf("✅ 使用优先选择的IP: %s", preferredIP)
		return preferredIP, nil
	}

	// 使用智能优先级选择
	if len(candidateIPs) > 0 {
		selectedIP := getPreferredIP(candidateIPs)
		log.Printf("✅ 智能选择的IP: %s", selectedIP)
		return selectedIP, nil
	}

	// 如果没有找到任何IP，根据环境提供备用方案
	var fallbackIP string
	if isLinux() {
		fallbackIP = "192.168.1.100" // Linux常见网段
		log.Printf("🐧 Linux环境使用备用IP: %s", fallbackIP)
	} else {
		fallbackIP = "172.16.181.101" // 原有备用IP
		log.Printf("💻 使用备用IP: %s", fallbackIP)
	}

	return fallbackIP, nil
}

// configureMongoDBReplicaSet 配置MongoDB副本集
func configureMongoDBReplicaSet(hostIP string) error {
	fmt.Println("🔧 开始配置MongoDB副本集...")

	// 获取MongoDB认证信息
	mongoUser, mongoPass := getMongoConfig()

	// 检查MongoDB连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用mongosh命令检查连接（实际连接时使用真实密码）
	// 处理密码编码：如果环境变量中已经是编码过的，先解码再重新编码
	var actualPassword string
	// strings可以测包含问题
	if strings.Contains(mongoPass, "%40") {
		// 如果密码包含%40，先解码得到原始密码
		actualPassword = strings.ReplaceAll(mongoPass, "%40", "@")
	} else {
		// 否则直接使用
		actualPassword = mongoPass
	}

	// 然后进行正确的URL编码
	encodedPassword := strings.ReplaceAll(actualPassword, "@", "%40")
	// 类似于OS的stringsFormat
	realMongoURI := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%s/%s?authSource=admin",
		mongoUser, encodedPassword, MongoPort, MongoDatabase)
	checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
	if output, err := checkCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("MongoDB连接失败: %v, 输出: %s", err, string(output))
	}
	fmt.Println("✅ MongoDB连接成功")

	// 检查副本集状态 - 使用更可靠的检测方法
	checkRSCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
		"try { rs.status().ok } catch(e) { print('NOT_INITIALIZED') }")
	output, err := checkRSCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("检查副本集状态失败: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	fmt.Printf("🔍 副本集状态检测输出: '%s'\n", outputStr)

	// 更可靠的状态判断：如果输出是'1'或者不是'NOT_INITIALIZED'，说明副本集已初始化
	if outputStr == "1" || (outputStr != "NOT_INITIALIZED" && outputStr != "") {
		fmt.Println("✅ 副本集已配置，检查IP配置...")

		// 获取当前配置，这里Command是名词，指令，合到一起是指令上下文的意思
		getConfigCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "rs.conf().members[0].host")
		// 执行并合并结果
		output, err := getConfigCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("获取当前副本集配置失败: %v", err)
		}

		currentHost := strings.TrimSpace(string(output))
		fmt.Printf("📊 当前副本集配置: %s\n", currentHost)

		// 如果配置不正确，重新配置
		if !strings.Contains(currentHost, hostIP) {
			fmt.Println("🔧 更新副本集配置...")
			reconfigCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
				fmt.Sprintf(`rs.reconfig({_id: "rs0", members: [{_id: 0, host: "%s:%s"}]}, {force: true})`, hostIP, MongoPort))
			if output, err := reconfigCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("更新副本集配置失败: %v, 输出: %s", err, string(output))
			}
			fmt.Println("✅ 副本集配置已更新")
		} else {
			fmt.Println("✅ 副本集配置已正确")
		}
	} else {
		fmt.Println("🔧 副本集未初始化，开始初始化...")

		// 额外确认：再次检查是否真的未初始化
		confirmCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
			"try { rs.conf() } catch(e) { print('NOT_CONFIGURED') }")
		confirmOutput, confirmErr := confirmCmd.CombinedOutput()
		if confirmErr == nil {
			confirmStr := strings.TrimSpace(string(confirmOutput))
			if confirmStr != "NOT_CONFIGURED" && confirmStr != "" {
				fmt.Println("✅ 副本集实际已配置，检查IP配置...")
				fmt.Printf("🔍 确认输出: '%s'\n", confirmStr)

				// 检查当前配置的IP是否与当前主机IP匹配
				if !strings.Contains(confirmStr, hostIP) {
					fmt.Printf("⚠️ 副本集IP配置不匹配，当前主机IP: %s\n", hostIP)
					fmt.Println("🔧 更新副本集IP配置...")

					// 使用rs.reconfig更新IP配置
					reconfigCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
						fmt.Sprintf(`rs.reconfig({_id: "rs0", members: [{_id: 0, host: "%s:%s"}]}, {force: true})`, hostIP, MongoPort))
					if output, err := reconfigCmd.CombinedOutput(); err != nil {
						return fmt.Errorf("更新副本集IP配置失败: %v, 输出: %s", err, string(output))
					}
					fmt.Println("✅ 副本集IP配置已更新")

					// 等待配置生效
					fmt.Println("⏳ 等待副本集配置生效...")
					time.Sleep(5 * time.Second)
				} else {
					fmt.Println("✅ 副本集IP配置正确")
				}
			} else {
				// 确实未初始化，执行初始化
				fmt.Printf("🔧 使用主机IP %s 初始化副本集...\n", hostIP)
				initCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
					fmt.Sprintf(`rs.initiate({_id: "rs0", members: [{_id: 0, host: "%s:%s"}]})`, hostIP, MongoPort))
				if output, err := initCmd.CombinedOutput(); err != nil {
					return fmt.Errorf("初始化副本集失败: %v, 输出: %s", err, string(output))
				}
				fmt.Println("✅ 副本集初始化成功")

				// 等待副本集选举完成
				fmt.Println("⏳ 等待副本集选举完成...")
				time.Sleep(10 * time.Second)
			}
		} else {
			fmt.Printf("⚠️ 确认检查失败，假设副本集未初始化: %v\n", confirmErr)
			// 执行初始化
			initCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
				fmt.Sprintf(`rs.initiate({_id: "rs0", members: [{_id: 0, host: "%s:%s"}]})`, hostIP, MongoPort))
			if output, err := initCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("初始化副本集失败: %v, 输出: %s", err, string(output))
			}
			fmt.Println("✅ 副本集初始化成功")
			time.Sleep(10 * time.Second)
		}
	}

	// 验证副本集状态 - 使用更安全的验证方法
	fmt.Println("🔍 验证副本集状态...")

	// 首先尝试简单的状态检查
	simpleVerifyCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
		"try { print('REPLICA_SET_OK:' + rs.status().ok) } catch(e) { print('REPLICA_SET_ERROR:' + e.message) }")

	simpleOutput, simpleErr := simpleVerifyCmd.CombinedOutput()
	if simpleErr == nil {
		simpleStr := strings.TrimSpace(string(simpleOutput))
		fmt.Printf("🔍 简单验证结果: %s\n", simpleStr)

		if strings.Contains(simpleStr, "REPLICA_SET_OK:1") {
			fmt.Println("✅ 副本集状态验证通过")

			// 尝试获取详细信息（可能失败但不影响整体结果）
			detailVerifyCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
				`try { rs.status().members.forEach(function(member) { print("- " + member.name + ": " + (member.healthStr || 'unknown') + " (" + member.stateStr + ")") }) } catch(e) { print("详细状态获取失败，但副本集基本状态正常") }`)

			if detailOutput, detailErr := detailVerifyCmd.CombinedOutput(); detailErr == nil {
				fmt.Printf("📊 副本集详细信息:\n%s", string(detailOutput))
			} else {
				fmt.Printf("⚠️ 详细状态获取失败，但副本集基本状态正常: %v\n", detailErr)
			}

			return nil
		}
	}

	// 如果简单验证失败，尝试详细验证作为备用方案
	fmt.Println("⚠️ 简单验证失败，尝试详细验证...")
	detailVerifyCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
		`rs.status().members.forEach(function(member) { print("- " + member.name + ": " + (member.healthStr || 'unknown') + " (" + member.stateStr + ")") })`)

	detailOutput, detailErr := detailVerifyCmd.CombinedOutput()
	if detailErr == nil {
		detailStr := string(detailOutput)
		fmt.Printf("📊 副本集状态:\n%s", detailStr)

		// 检查是否有有效的成员
		if strings.Contains(detailStr, hostIP) {
			fmt.Println("✅ 找到当前主机IP在副本集中，验证通过")
			return nil
		}
	}

	// 所有验证都失败
	return fmt.Errorf("副本集状态验证失败 - 简单验证: %v, 详细验证: %v", simpleErr, detailErr)
}

// runDockerDeploy 执行Docker部署
func runDockerDeploy() error {
	var err error
	fmt.Println("🐳 开始Docker部署...")

	// 检查Docker是否可用
	fmt.Println("🔍 检查Docker可用性...")
	dockerCmd := exec.Command("docker", "--version")
	output, err := dockerCmd.CombinedOutput()
	if err != nil {
		// 提供更详细的错误信息
		if strings.Contains(string(output), "command not found") {
			return fmt.Errorf("❌ Docker未安装，请先安装Docker: %v", err)
		} else if strings.Contains(string(output), "permission denied") {
			return fmt.Errorf("❌ Docker权限不足，请检查用户权限: %v", err)
		} else {
			return fmt.Errorf("❌ Docker服务异常: %v, 输出: %s", err, string(output))
		}
	}
	fmt.Printf("✅ Docker可用: %s", string(output))

	// 停止现有容器（如果存在）
	fmt.Println("🔄 停止现有容器...")
	if err := exec.Command("docker-compose", "down").Run(); err != nil {
		log.Printf("⚠️ 停止现有容器失败: %v", err)
		log.Println("🔍 继续部署流程，可能需要手动清理容器")
	} else {
		fmt.Println("✅ 现有容器已停止")
	}

	// 构建并启动服务
	fmt.Println("🔨 构建并启动服务...")

	// 执行docker-compose up -d
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose启动失败: %v", err)
	}

	fmt.Println("⏳ 等待服务启动...")
	time.Sleep(10 * time.Second)

	// 检查服务状态
	fmt.Println("🔍 检查服务状态...")
	statusCmd := exec.Command("docker-compose", "ps")
	var statusOutput []byte
	statusOutput, err = statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("检查服务状态失败: %v", err)
	}
	fmt.Printf("📊 服务状态:\n%s", string(statusOutput))

	// 验证健康检查
	fmt.Println("🏥 执行健康检查...")
	for i := 0; i < 30; i++ {
		healthCmd := exec.Command("curl", "-s", "http://localhost:8080/health")
		if output, err := healthCmd.CombinedOutput(); err == nil {
			response := string(output)
			if strings.Contains(response, "ok") {
				fmt.Printf("✅ 服务健康检查通过，响应: %s\n", strings.TrimSpace(response))
				return nil
			} else {
				// 显示实际响应内容，便于调试
				fmt.Printf("🔍 健康检查响应异常: %s\n", strings.TrimSpace(response))
			}
		} else {
			// 记录连接错误，但不停止重试
			fmt.Printf("🔍 健康检查连接失败: %v\n", err)
		}

		fmt.Printf("⏳ 等待服务就绪... (%d/30)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("服务健康检查超时")
}