package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
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

func main() {
	// 加载.env文件
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
		log.Println("将使用默认环境变量")
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

// getHostIP 获取宿主机在局域网中的真实IP
func getHostIP() (string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口失败: %v", err)
	}

	var candidateIPs []string

	for _, inter := range interfaces {
		// 跳过回环接口和down状态的接口
		if inter.Flags&net.FlagLoopback != 0 || inter.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := inter.Addrs()
		if err != nil {
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
				continue
			}

			// 优先选择172.16网段（你的局域网）
			if strings.HasPrefix(ip.String(), "172.16.") {
				fmt.Printf("🔍 找到172.16网段IP: %s\n", ip.String())
				return ip.String(), nil
			}

			// 收集其他候选IP（跳过Docker网络）
			if !strings.HasPrefix(ip.String(), "192.168.65.") &&
			   !strings.HasPrefix(ip.String(), "172.17.") &&
			   !strings.HasPrefix(ip.String(), "127.") {
				candidateIPs = append(candidateIPs, ip.String())
			}
		}
	}

	// 如果没有找到172.16网段，使用其他候选IP
	if len(candidateIPs) > 0 {
		fmt.Printf("🔍 使用候选IP: %s\n", candidateIPs[0])
		return candidateIPs[0], nil
	}

	// 最后的备用方案
	fmt.Println("⚠️ 使用备用IP: 172.16.181.101")
	return "172.16.181.101", nil
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
	if strings.Contains(mongoPass, "%40") {
		// 如果密码包含%40，先解码得到原始密码
		actualPassword = strings.ReplaceAll(mongoPass, "%40", "@")
	} else {
		// 否则直接使用
		actualPassword = mongoPass
	}

	// 然后进行正确的URL编码
	encodedPassword := strings.ReplaceAll(actualPassword, "@", "%40")
	realMongoURI := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%s/%s?authSource=admin",
		mongoUser, encodedPassword, MongoPort, MongoDatabase)
	checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
	if output, err := checkCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("MongoDB连接失败: %v, 输出: %s", err, string(output))
	}
	fmt.Println("✅ MongoDB连接成功")

	// 检查副本集状态
	checkRSCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
		"try { rs.status().ok } catch(e) { 0 }")
	output, err := checkRSCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("检查副本集状态失败: %v", err)
	}

	status := strings.TrimSpace(string(output))
	if status == "1" {
		fmt.Println("✅ 副本集已配置，检查IP配置...")

		// 获取当前配置
		getConfigCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "rs.conf().members[0].host")
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
		fmt.Println("🔧 初始化副本集...")
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

	// 验证副本集状态
	fmt.Println("🔍 验证副本集状态...")
	verifyCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval",
		`rs.status().members.forEach(function(member) { print("- " + member.name + ": " + member.healthStr + " (" + member.stateStr + ")") })`)
	var verifyOutput []byte
	verifyOutput, err = verifyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("验证副本集状态失败: %v", err)
	}
	fmt.Printf("📊 副本集状态:\n%s", string(verifyOutput))

	return nil
}

// runDockerDeploy 执行Docker部署
func runDockerDeploy() error {
	var err error
	fmt.Println("🐳 开始Docker部署...")

	// 检查Docker是否运行
	dockerCmd := exec.Command("docker", "--version")
	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("Docker未运行或未安装: %v", err)
	}
	fmt.Println("✅ Docker服务正常")

	// 停止现有容器（如果存在）
	fmt.Println("🔄 停止现有容器...")
	exec.Command("docker-compose", "down").Run()

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
			if strings.Contains(string(output), "ok") {
				fmt.Println("✅ 服务健康检查通过")
				return nil
			}
		}

		fmt.Printf("⏳ 等待服务就绪... (%d/30)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("服务健康检查超时")
}