package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[0;34m"
)

// Config 配置结构
type Config struct {
	MongoUser string
	MongoPass string
	MongoPort string
}

// ReplicaSetConfig 副本集配置
type ReplicaSetConfig struct {
	ID      string             `bson:"_id"`
	Members []ReplicaSetMember `bson:"members"`
	Version int                `bson:"version"`
}

// ReplicaSetMember 副本集成员
type ReplicaSetMember struct {
	ID   int    `bson:"_id"`
	Host string `bson:"host"`
}

func main() {
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBlue, colorReset)
	fmt.Printf("%s  MongoDB 副本集地址刷新工具 (Go版本)%s\n", colorBlue, colorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBlue, colorReset)

	// 1. 加载配置
	config := loadConfig()

	// 2. 获取当前局域网 IP
	hostIP := getLocalIP()
	if hostIP == "" {
		log.Fatalf("%s❌ 无法获取局域网 IP 地址%s\n", colorRed, colorReset)
	}
	fmt.Printf("%s✅ 当前局域网 IP: %s%s\n\n", colorGreen, hostIP, colorReset)

	// 3. 连接 MongoDB
	client := connectToMongo(config)
	defer client.Disconnect(context.Background())

	// 4. 检查副本集状态
	fmt.Printf("%s🔍 检查副本集状态...%s\n", colorYellow, colorReset)
	replConfig := checkReplicaSet(client)

	if replConfig == nil {
		// 副本集未初始化
		fmt.Printf("%s❌ 副本集未配置%s\n", colorRed, colorReset)
		fmt.Printf("%s是否需要初始化副本集? (y/n): %s", colorYellow, colorReset)
		var answer string
		fmt.Scanln(&answer)

		if strings.ToLower(answer) == "y" {
			initializeReplicaSet(client, hostIP, config.MongoPort)
		} else {
			fmt.Println("取消操作")
			return
		}
	} else {
		// 副本集已配置，检查是否需要更新
		if len(replConfig.Members) == 0 {
			fmt.Printf("%s⚠️  副本集配置异常，没有成员%s\n", colorYellow, colorReset)
			return
		}

		currentMember := replConfig.Members[0].Host
		fmt.Printf("%s📊 当前副本集配置: %s%s\n", colorYellow, colorReset, currentMember)

		expectedHost := fmt.Sprintf("%s:%s", hostIP, config.MongoPort)

		if currentMember == expectedHost {
			fmt.Printf("%s✅ 副本集配置已是最新，无需更新%s\n\n", colorGreen, colorReset)
			showReplicaSetStatus(client)
			printNoRestartTip()
			return
		}

		// 需要更新
		fmt.Printf("%s🔧 检测到网络环境变化%s\n", colorYellow, colorReset)
		fmt.Printf("   旧地址: %s%s%s\n", colorRed, currentMember, colorReset)
		fmt.Printf("   新地址: %s%s%s\n", colorGreen, expectedHost, colorReset)
		fmt.Printf("\n%s是否确认更新? (y/n): %s", colorYellow, colorReset)

		var confirm string
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) == "y" {
			updateReplicaSet(client, hostIP, config.MongoPort)
		} else {
			fmt.Println("取消操作")
			return
		}
	}

	// 5. 验证并显示状态
	fmt.Printf("\n%s🔍 验证更新结果...%s\n", colorYellow, colorReset)
	time.Sleep(2 * time.Second)

	replConfig = checkReplicaSet(client)
	if replConfig != nil && len(replConfig.Members) > 0 {
		newMember := replConfig.Members[0].Host
		expectedHost := fmt.Sprintf("%s:%s", hostIP, config.MongoPort)

		if newMember == expectedHost {
			fmt.Printf("%s✅ 副本集配置更新成功%s\n\n", colorGreen, colorReset)
			showReplicaSetStatus(client)
			printNoRestartTip()
		} else {
			log.Fatalf("%s❌ 验证失败%s\n", colorRed, colorReset)
		}
	}
}

// loadConfig 加载配置
func loadConfig() Config {
	return Config{
		MongoUser: getEnv("MONGO_USER", "admin"),
		MongoPass: getEnv("MONGO_PASS", "password"),
		MongoPort: getEnv("MONGO_PORT", "27017"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getLocalIP 获取本地局域网 IP
func getLocalIP() string {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	var candidates []string

	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
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

			// 只支持 IPv4
			ip = ip.To4()
			if ip == nil {
				continue
			}

			ipStr := ip.String()

			// 过滤掉 Docker 和虚拟网卡
			if strings.HasPrefix(ipStr, "172.17.") || strings.HasPrefix(ipStr, "192.168.65.") {
				continue
			}

			// 优先选择 172.16 网段（你的局域网）
			if strings.HasPrefix(ipStr, "172.16.") {
				candidates = append([]string{ipStr}, candidates...)
			} else if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
				candidates = append(candidates, ipStr)
			} else {
				candidates = append(candidates, ipStr)
			}
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}

	return ""
}

// connectToMongo 连接到 MongoDB
func connectToMongo(config Config) *mongo.Client {
	fmt.Printf("%s🔗 连接 MongoDB...%s\n", colorYellow, colorReset)

	uri := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%s/admin?authSource=admin",
		config.MongoUser, config.MongoPass, config.MongoPort)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("%s❌ MongoDB 连接失败: %v%s\n", colorRed, err, colorReset)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("%s❌ MongoDB Ping 失败: %v%s\n", colorRed, err, colorReset)
	}

	fmt.Printf("%s✅ MongoDB 连接成功%s\n\n", colorGreen, colorReset)
	return client
}

// checkReplicaSet 检查副本集配置
func checkReplicaSet(client *mongo.Client) *ReplicaSetConfig {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试获取副本集配置
	result := client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "replSetGetConfig", Value: 1},
	})

	if result.Err() != nil {
		return nil
	}

	// 解码结果
	var rawResult bson.M
	if err := result.Decode(&rawResult); err != nil {
		return nil
	}

	// 配置在 "config" 字段中
	configData, ok := rawResult["config"].(bson.M)
	if !ok {
		return nil
	}

	// 将 config 数据转换为 BSON 再解码到结构体
	configBytes, err := bson.Marshal(configData)
	if err != nil {
		return nil
	}

	var config ReplicaSetConfig
	if err := bson.Unmarshal(configBytes, &config); err != nil {
		return nil
	}

	return &config
}

// initializeReplicaSet 初始化副本集
func initializeReplicaSet(client *mongo.Client, hostIP, port string) {
	fmt.Printf("%s🔧 正在初始化副本集...%s\n", colorYellow, colorReset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config := ReplicaSetConfig{
		ID:      "rs0",
		Version: 1,
		Members: []ReplicaSetMember{
			{ID: 0, Host: fmt.Sprintf("%s:%s", hostIP, port)},
		},
	}

	result := client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "replSetInitiate", Value: config},
	})

	if result.Err() != nil {
		log.Fatalf("%s❌ 初始化失败: %v%s\n", colorRed, result.Err(), colorReset)
	}

	fmt.Printf("%s✅ 副本集初始化完成%s\n", colorGreen, colorReset)
	fmt.Printf("%s⏳ 等待副本集选举完成...%s\n", colorYellow, colorReset)
	time.Sleep(5 * time.Second)
}

// updateReplicaSet 更新副本集配置
func updateReplicaSet(client *mongo.Client, hostIP, port string) {
	fmt.Printf("%s⏳ 正在更新副本集配置...%s\n", colorYellow, colorReset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先获取当前配置以获取 version
	var currentConfig ReplicaSetConfig
	configResult := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetConfig", Value: 1}})
	if err := configResult.Decode(&currentConfig); err != nil {
		log.Fatalf("%s❌ 获取当前配置失败: %v%s\n", colorRed, err, colorReset)
	}

	// 更新配置
	newConfig := ReplicaSetConfig{
		ID:      "rs0",
		Version: currentConfig.Version + 1, // 版本号必须增加
		Members: []ReplicaSetMember{
			{ID: 0, Host: fmt.Sprintf("%s:%s", hostIP, port)},
		},
	}

	result := client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "replSetReconfig", Value: newConfig},
	})

	if result.Err() != nil {
		// 尝试使用 force 选项
		result = client.Database("admin").RunCommand(ctx, bson.D{
			{Key: "replSetReconfig", Value: newConfig},
			{Key: "force", Value: true},
		})

		if result.Err() != nil {
			log.Fatalf("%s❌ 更新失败: %v%s\n", colorRed, result.Err(), colorReset)
		}
	}

	fmt.Printf("%s✅ 副本集配置更新成功%s\n", colorGreen, colorReset)
}

// showReplicaSetStatus 显示副本集状态
func showReplicaSetStatus(client *mongo.Client) {
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBlue, colorReset)
	fmt.Printf("%s  当前副本集状态%s\n", colorBlue, colorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBlue, colorReset)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}})

	var status bson.M
	if err := result.Decode(&status); err != nil {
		fmt.Printf("%s❌ 获取状态失败: %v%s\n", colorRed, err, colorReset)
		return
	}

	// 解析并显示状态
	if setName, ok := status["set"].(string); ok {
		fmt.Printf("📊 副本集信息:\n")
		fmt.Printf("   名称: %s\n", setName)
		fmt.Printf("   状态: %s✅ 正常%s\n\n", colorGreen, colorReset)
	}

	if members, ok := status["members"].(bson.A); ok {
		fmt.Printf("🖥️  节点列表:\n")
		for _, member := range members {
			if m, ok := member.(bson.M); ok {
				name := fmt.Sprintf("%v", m["name"])
				stateStr := fmt.Sprintf("%v", m["stateStr"])
				healthStr := fmt.Sprintf("%v", m["healthStr"])

				icon := "🔹"
				if healthStr == "PRIMARY" {
					icon = "👑"
				}

				fmt.Printf("   %s %s\n", icon, name)
				fmt.Printf("      状态: %s (%s)\n", healthStr, stateStr)
			}
		}
	}

	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBlue, colorReset)
	fmt.Printf("%s  ✅ 刷新完成 - Docker 容器将自动连接%s\n", colorGreen, colorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBlue, colorReset)
}

// printNoRestartTip 打印无需重启的提示
func printNoRestartTip() {
	fmt.Printf("%s💡 提示:%s\n", colorYellow, colorReset)
	fmt.Println("   副本集配置已更新为当前网络 IP")
	fmt.Println("   Docker 容器通过 host.docker.internal 自动连接")
	fmt.Println("   ✅ 无需重启 Docker 容器")
	fmt.Println("\n   下次切换网络环境时，再次运行此脚本即可:")
	fmt.Printf("   %sMONGO_PASS=你的密码 go run refresh-mongodb.go%s\n\n", colorGreen, colorReset)
}

func init() {
	// Windows 不支持 ANSI 颜色
	if runtime.GOOS == "windows" {
		colorReset = ""
		colorRed = ""
		colorGreen = ""
		colorYellow = ""
		colorBlue = ""
	}
}
