# Go脚本编写教程

## 🎯 概述

Go语言是编写命令行脚本和自动化工具的优秀选择。本文档以项目中的 `scripts/deploy.go` 为例，详细介绍如何使用Go编写脚本。

## 🏗️ Go脚本基本结构 


### 1. Package和Import

每个Go可执行脚本都必须以 `package main` 开始，并导入需要的包：

```go
package main  // 必须的！表示这是可执行程序

import (
    // Go内置包
    "context"    // 处理超时和取消
    "fmt"        // 格式化输出（打印）
    "log"        // 日志记录
    "net"        // 网络操作
    "os"         // 操作系统功能
    "os/exec"    // 执行外部命令
    "strings"    // 字符串处理
    "time"       // 时间操作

    // 第三方包
    "github.com/joho/godotenv"  // 加载.env文件
)
```

### 2. 常量和变量定义

```go
// 常量定义 - 不会改变的值
const (
    MongoPort     = "27017"
    MongoDatabase = "admin"
)

// 函数定义 - 可复用的逻辑块
func getMongoConfig() (string, string) {
    user := getEnv("MONGO_USER", "admin")    // 获取用户名，默认admin
    pass := getEnv("MONGO_PASS", "password") // 获取密码，默认password
    return user, pass                        // 返回两个值
}
```

## 🎮 主函数 - 程序入口点

```go
func main() {
    // 1. 加载配置文件
    if err := godotenv.Load("../.env"); err != nil {
        log.Printf("警告: 无法加载.env文件: %v", err)  // 只是警告，继续执行
    }

    fmt.Println("🚀 开始自动化部署novel-resource-management...")  // 打印消息

    // 2. 获取宿主机IP
    hostIP, err := getHostIP()  // 调用函数
    if err != nil {            // 错误处理
        log.Fatalf("❌ 获取宿主机IP失败: %v", err)  // 致命错误，程序退出
    }
    fmt.Printf("✅ 宿主机IP: %s\n", hostIP)  // 成功消息

    // 3. 配置MongoDB
    if err := configureMongoDBReplicaSet(hostIP); err != nil {
        log.Fatalf("❌ MongoDB副本集配置失败: %v", err)
    }
}
```

## 🌐 常用功能实现

### 1. 环境变量处理

```go
// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// 敏感信息通过环境变量获取
func getMongoConfig() (string, string) {
    user := getEnv("MONGO_USER", "admin")
    pass := getEnv("MONGO_PASS", "password")
    return user, pass
}
```

### 2. 网络操作

```go
func getHostIP() (string, error) {
    // 1. 获取所有网络接口
    interfaces, err := net.Interfaces()
    if err != nil {
        return "", fmt.Errorf("获取网络接口失败: %v", err)  // 返回错误
    }

    // 2. 遍历网络接口
    for _, inter := range interfaces {
        // 跳过回环接口和down状态的接口
        if inter.Flags&net.FlagLoopback != 0 || inter.Flags&net.FlagUp == 0 {
            continue  // 继续下一个
        }

        // 3. 获取接口地址
        addrs, err := inter.Addrs()
        if err != nil {
            continue  // 忽略错误，继续
        }

        // 4. 检查每个地址
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

            // 优先选择172.16网段
            if strings.HasPrefix(ip.String(), "172.16.") {
                return ip.String(), nil
            }
        }
    }

    // 返回默认IP
    return "172.16.181.101", nil
}
```

### 3. 执行外部命令

```go
func runDockerDeploy() error {
    // 1. 简单命令执行
    dockerCmd := exec.Command("docker", "--version")
    if err := dockerCmd.Run(); err != nil {
        return fmt.Errorf("Docker未运行或未安装: %v", err)
    }

    // 2. 停止现有容器（忽略错误）
    exec.Command("docker-compose", "down").Run()

    // 3. 带参数的命令，显示输出
    cmd := exec.Command("docker-compose", "up", "-d")
    cmd.Stdout = os.Stdout  // 输出到控制台
    cmd.Stderr = os.Stderr  // 错误输出到控制台

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Docker Compose启动失败: %v", err)
    }

    // 4. 获取命令输出
    statusCmd := exec.Command("docker-compose", "ps")
    output, err := statusCmd.CombinedOutput()  // 获取标准输出和错误输出
    if err != nil {
        return fmt.Errorf("检查服务状态失败: %v", err)
    }

    fmt.Printf("📊 服务状态:\n%s", string(output))  // 输出结果
    return nil
}
```

### 4. 超时控制

```go
func configureMongoDBReplicaSet(hostIP string) error {
    // 设置10秒超时
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()  // 确保取消函数被调用

    // 使用带超时的命令
    checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
    if output, err := checkCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("MongoDB连接失败: %v, 输出: %s", err, string(output))
    }

    return nil
}
```

### 5. 重试机制和健康检查

```go
func performHealthCheck() error {
    // 最多尝试30次，每次间隔2秒
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
```

## 🎓 Go脚本核心概念

### 1. 基本语法

```go
package main  // 可执行程序必须是main包

import (
    "fmt"     // 导入需要的包
    "os"      // 多个包用括号括起来
)

// 函数定义
func main() {  // main函数是程序入口
    fmt.Println("Hello, World!")  // 打印输出
}
```

### 2. 变量和错误处理

```go
func myFunction() error {
    // 变量声明
    var name string = "张三"
    age := 25  // 简短声明，自动推断类型

    // 多返回值
    result, err := someFunction()
    if err != nil {  // 错误处理是Go的重点！
        return fmt.Errorf("操作失败: %v", err)
    }

    fmt.Printf("结果: %s\n", result)
    return nil  // nil表示没有错误
}
```

### 3. 条件和循环

```go
// if条件
if age > 18 {
    fmt.Println("成年人")
} else if age > 12 {
    fmt.Println("青少年")
} else {
    fmt.Println("儿童")
}

// for循环
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// 遍历切片/数组
numbers := []int{1, 2, 3, 4, 5}
for index, value := range numbers {
    fmt.Printf("索引%d: 值%d\n", index, value)
}
```

### 4. 字符串处理

```go
text := "hello world"

// 检查包含
if strings.Contains(text, "hello") {
    fmt.Println("包含hello")
}

// 替换
newText := strings.ReplaceAll(text, "world", "Go")
fmt.Println(newText)  // hello Go

// 分割
parts := strings.Split("a,b,c", ",")
fmt.Println(parts)  // [a b c]

// 去除空白
trimmed := strings.TrimSpace("  hello  ")
fmt.Println(trimmed)  // hello
```

## 🛠️ 常用命令行操作

### 1. 文件操作

```go
// 读取文件
content, err := os.ReadFile("file.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("文件内容: %s\n", content)

// 写入文件
err = os.WriteFile("output.txt", []byte("你好世界"), 0644)
if err != nil {
    log.Fatal(err)
}

// 检查文件是否存在
if _, err := os.Stat("file.txt"); os.IsNotExist(err) {
    fmt.Println("文件不存在")
}
```

### 2. 命令行参数

```go
func main() {
    // os.Args[0] 是程序名
    // os.Args[1:] 是真正的参数
    if len(os.Args) < 2 {
        fmt.Println("用法: program <参数1> [参数2]")
        return
    }

    arg1 := os.Args[1]
    fmt.Printf("第一个参数: %s\n", arg1)

    if len(os.Args) >= 3 {
        arg2 := os.Args[2]
        fmt.Printf("第二个参数: %s\n", arg2)
    }
}
```

### 3. 退出码

```go
func main() {
    // 正常退出
    // os.Exit(0)

    // 错误退出
    // os.Exit(1)

    // 使用log.Fatal会自动调用os.Exit(1)
    if someError {
        log.Fatal("发生致命错误")
    }
}
```

## 📚 项目依赖管理

### 1. 初始化模块

```bash
# 在脚本目录下初始化Go模块
go mod init myscript

# 添加依赖
go get github.com/joho/godotenv
```

### 2. go.mod 文件示例

```go
module myscript

go 1.23.0

require (
    github.com/joho/godotenv v1.5.1
)
```

## 🚀 运行Go脚本

### 1. 直接运行

```bash
# 运行Go文件
go run script.go

# 带参数运行
go run script.go arg1 arg2
```

### 2. 编译后运行

```bash
# 编译
go build -o myscript script.go

# 运行
./myscript arg1 arg2

# 交叉编译（为Linux编译）
GOOS=linux GOARCH=amd64 go build -o myscript-linux script.go
```

## 💡 最佳实践

### 1. 错误处理

- 总是检查函数返回的错误
- 使用 `fmt.Errorf` 添加上下文信息
- 在适当的地方使用 `log.Fatal` 处理致命错误

### 2. 代码组织

- 将复杂逻辑拆分为小函数
- 使用有意义的函数名和变量名
- 添加必要的注释

### 3. 安全性

- 不要硬编码敏感信息
- 使用环境变量存储密码等敏感数据
- 将 `.env` 文件添加到 `.gitignore`

### 4. 日志记录

```go
// 普通信息
fmt.Println("正在处理...")

// 警告信息
log.Printf("警告: %v", err)

// 错误信息但继续执行
log.Printf("错误: %v，继续执行", err)

// 致命错误，程序退出
log.Fatalf("致命错误: %v", err)
```

## 🔗 参考资源

- [Go官方文档](https://golang.org/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [本项目的deploy.go](../scripts/deploy.go) - 实际案例参考