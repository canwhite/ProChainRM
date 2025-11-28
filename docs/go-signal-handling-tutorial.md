# Go 信号处理完全教程

## 信号处理是什么？

在 Go 语言中，信号（Signal）是操作系统用来通知进程发生了某种事件的一种机制。理解信号处理对于编写健壮的服务器程序非常重要。

## 基本概念：信号就像"门铃"

想象一下你的程序是一个办公楼：

```
┌─────────────────┐
│   你的程序       │
│                 │
│  ┌──────────┐   │
│  │  门铃声   │   │  <-- 操作系统发送信号
│  │  通知系统  │   │
│  └──────────┘   │
│        ↓         │
│  ┌──────────┐   │
│  │  看门人   │   │  <-- <-sigChan 等待信号
│  │  等待中   │   │
│  └──────────┘   │
└─────────────────┘
```

- **信号**：门铃声（Ctrl+C、系统关闭等）
- **signal.Notify**：告诉看门人要听什么门铃声
- **<-sigChan**：看门人坐在门口等待门铃响

## signal.Notify 详细讲解

### 基本语法
```go
signal.Notify(c chan<- os.Signal, sig ...os.Signal)
```

### 参数解释

#### 第一个参数：信号通道
```go
sigChan := make(chan os.Signal, 1)
```
- `chan<- os.Signal`：只能发送信号的通道
- `1`：缓冲大小，表示最多存1个信号

#### 第二个参数开始的：要监听的信号
```go
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
```
- `syscall.SIGINT`：中断信号（Ctrl+C）
- `syscall.SIGTERM`：终止信号（系统关闭）
- 可以监听多个信号

### 常见的系统信号

| 信号 | 含义 | 触发方式 | 可否捕获 |
|------|------|----------|----------|
| `SIGINT` | 中断信号 | 用户按 `Ctrl+C` | ✅ 可以 |
| `SIGTERM` | 终止信号 | 系统要求退出 | ✅ 可以 |
| `SIGHUP` | 挂断信号 | 终端关闭 | ✅ 可以 |
| `SIGQUIT` | 退出信号 | 用户按 `Ctrl+\` | ✅ 可以 |
| `SIGKILL` | 强制终止 | `kill -9` | ❌ 不可以 |
| `SIGSTOP` | 停止进程 | `kill -19` | ❌ 不可以 |

## 基本用法示例

### 1. 最简单的信号处理
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)

    // 注册要监听的信号
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("程序启动，等待信号...")

    // 阻塞等待信号
    sig := <-sigChan
    fmt.Printf("收到信号: %v\n", sig)

    fmt.Println("程序即将退出")
}
```

### 2. 优雅关闭服务器
```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "context"
)

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 创建 HTTP 服务器
    server := &http.Server{
        Addr: ":8080",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Hello, World!")
        }),
    }

    // 在后台启动服务器
    go func() {
        log.Println("服务器启动，监听 :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("服务器启动失败: %v", err)
        }
    }()

    // 等待信号
    sig := <-sigChan
    log.Printf("收到信号: %v，开始优雅关闭...\n", sig)

    // 设置关闭超时
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 优雅关闭服务器
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("服务器关闭失败: %v", err)
    }

    log.Println("服务器已关闭")
}
```

## 主流程等待的实用案例

### 案例1：简单的计数器服务
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 计数器
    counter := 0

    // 在后台运行计数器
    go func() {
        for {
            counter++
            fmt.Printf("计数: %d\n", counter)
            time.Sleep(1 * time.Second)
        }
    }()

    fmt.Println("计数器服务启动，按 Ctrl+C 停止...")

    // 等待信号
    <-sigChan
    fmt.Printf("计数器服务停止，最终计数: %d\n", counter)
}
```

### 案例2：日志收集服务
```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
}

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 日志通道
    logChan := make(chan LogEntry, 100)

    // 模拟日志生成器
    go func() {
        for i := 0; ; i++ {
            logChan <- LogEntry{
                Timestamp: time.Now(),
                Level:     "INFO",
                Message:   fmt.Sprintf("处理任务 %d", i),
            }
            time.Sleep(500 * time.Millisecond)
        }
    }()

    // 日志处理器
    go func() {
        for entry := range logChan {
            log.Printf("[%s] %s: %s\n",
                entry.Timestamp.Format("15:04:05"),
                entry.Level,
                entry.Message)
        }
    }()

    fmt.Println("日志收集服务启动，按 Ctrl+C 停止...")

    // 等待信号
    sig := <-sigChan
    fmt.Printf("\n收到信号: %v，停止日志收集...\n", sig)

    // 关闭日志通道
    close(logChan)

    // 等待最后一个日志处理完成
    time.Sleep(1 * time.Second)

    fmt.Println("日志收集服务已停止")
}
```

### 案例3：文件监控服务
```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 监控的文件
    fileName := "monitor.log"

    // 创建文件（如果不存在）
    if _, err := os.Stat(fileName); os.IsNotExist(err) {
        ioutil.WriteFile(fileName, []byte(""), 0644)
    }

    fmt.Printf("开始监控文件: %s\n", fileName)

    // 文件监控器
    go func() {
        var lastSize int64 = 0

        for {
            info, err := os.Stat(fileName)
            if err != nil {
                log.Printf("无法获取文件信息: %v", err)
                time.Sleep(1 * time.Second)
                continue
            }

            currentSize := info.Size()
            if currentSize > lastSize {
                fmt.Printf("文件大小发生变化: %d -> %d\n", lastSize, currentSize)
                lastSize = currentSize
            }

            time.Sleep(1 * time.Second)
        }
    }()

    fmt.Println("文件监控服务启动，按 Ctrl+C 停止...")
    fmt.Println("你可以修改 monitor.log 文件来测试监控")

    // 等待信号
    <-sigChan
    fmt.Println("\n收到停止信号，文件监控服务已停止")
}
```

### 案例4：健康检查服务
```go
package main

import (
    "fmt"
    "math/rand"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type ServiceStatus struct {
    Name   string
    Status string
    LastCheck time.Time
}

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 要监控的服务
    services := []string{
        "Web服务器",
        "数据库",
        "缓存服务",
        "消息队列",
    }

    // 状态通道
    statusChan := make(chan ServiceStatus, 10)

    // 健康检查器
    go func() {
        for {
            for _, service := range services {
                // 模拟健康检查
                time.Sleep(500 * time.Millisecond)

                status := "正常"
                if rand.Intn(10) < 2 { // 20% 概率异常
                    status = "异常"
                }

                statusChan <- ServiceStatus{
                    Name:      service,
                    Status:    status,
                    LastCheck: time.Now(),
                }
            }
            time.Sleep(2 * time.Second)
        }
    }()

    // 状态显示器
    go func() {
        for status := range statusChan {
            emoji := "✅"
            if status.Status == "异常" {
                emoji = "❌"
            }

            fmt.Printf("[%s] %s %s (检查时间: %s)\n",
                emoji,
                status.Name,
                status.Status,
                status.LastCheck.Format("15:04:05"))
        }
    }()

    fmt.Println("健康检查服务启动，按 Ctrl+C 停止...")
    fmt.Println("正在监控以下服务:")
    for _, service := range services {
        fmt.Printf("  - %s\n", service)
    }

    // 等待信号
    <-sigChan
    fmt.Println("\n收到停止信号，健康检查服务已停止")
}
```

### 案例5：简单的聊天服务器
```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/signal"
    "strings"
    "syscall"
)

type Client struct {
    Conn net.Conn
    Name string
}

func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 客户端通道
    clientChan := make(chan Client, 10)
    messageChan := make(chan string, 100)

    // 启动聊天服务器
    go startChatServer(":8080", clientChan, messageChan)

    // 消息广播器
    go func() {
        clients := make(map[net.Conn]Client)

        for {
            select {
            case client := <-clientChan:
                clients[client.Conn] = client
                messageChan <- fmt.Sprintf("欢迎 %s 加入聊天室", client.Name)

            case msg := <-messageChan:
                // 广播消息给所有客户端
                for conn := range clients {
                    conn.Write([]byte(msg + "\n"))
                }
            }
        }
    }()

    fmt.Println("聊天服务器启动，监听 :8080")
    fmt.Println("按 Ctrl+C 停止服务器")

    // 等待信号
    <-sigChan
    fmt.Println("\n收到停止信号，聊天服务器已停止")
}

func startChatServer(addr string, clientChan chan<- Client, messageChan chan<- string) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        panic(err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        go handleClient(conn, clientChan, messageChan)
    }
}

func handleClient(conn net.Conn, clientChan chan<- Client, messageChan chan<- string) {
    defer conn.Close()

    // 读取客户端名称
    reader := bufio.NewReader(conn)
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)

    // 通知新客户端加入
    client := Client{Conn: conn, Name: name}
    clientChan <- client

    // 处理客户端消息
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            break
        }

        messageChan <- fmt.Sprintf("%s: %s", name, strings.TrimSpace(msg))
    }

    messageChan <- fmt.Sprintf("%s 离开了聊天室", name)
}
```

## 高级用法

### 1. 处理多个不同信号
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    sigChan := make(chan os.Signal, 1)

    // 监听多个信号
    signal.Notify(sigChan,
        syscall.SIGINT,   // Ctrl+C
        syscall.SIGTERM,  // 终止信号
        syscall.SIGHUP,   // 挂断信号
        syscall.SIGQUIT,  // Ctrl+\
    )

    for sig := range sigChan {
        switch sig {
        case syscall.SIGINT:
            fmt.Println("收到 SIGINT (Ctrl+C)")
            // 可以选择不退出，只是处理信号
        case syscall.SIGTERM:
            fmt.Println("收到 SIGTERM，准备退出")
            goto Done
        case syscall.SIGHUP:
            fmt.Println("收到 SIGHUP，重新加载配置")
            // 重新加载配置但不退出
        case syscall.SIGQUIT:
            fmt.Println("收到 SIGQUIT，强制退出")
            goto Done
        }
    }

Done:
    fmt.Println("程序即将退出")
}
```

### 2. 超时等待信号
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("等待信号，最多等待 10 秒...")

    select {
    case sig := <-sigChan:
        fmt.Printf("在 10 秒内收到信号: %v\n", sig)
    case <-time.After(10 * time.Second):
        fmt.Println("10 秒内没有收到信号，程序自动退出")
    }
}
```

### 3. 信号处理中的清理工作
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

type ResourceManager struct {
    databaseFile string
    logFile      *os.File
    connections  []string
}

func (rm *ResourceManager) Initialize() {
    rm.databaseFile = "app.db"
    rm.connections = []string{"conn1", "conn2", "conn3"}

    file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    rm.logFile = file

    fmt.Println("资源管理器初始化完成")
}

func (rm *ResourceManager) Cleanup() {
    fmt.Println("开始清理资源...")

    // 关闭数据库连接
    fmt.Printf("关闭数据库文件: %s\n", rm.databaseFile)

    // 关闭日志文件
    if rm.logFile != nil {
        rm.logFile.Close()
        fmt.Println("关闭日志文件")
    }

    // 关闭网络连接
    for i, conn := range rm.connections {
        fmt.Printf("关闭连接 %d: %s\n", i+1, conn)
    }

    fmt.Println("资源清理完成")
}

func main() {
    // 创建资源管理器
    rm := &ResourceManager{}
    rm.Initialize()
    defer rm.Cleanup()

    // 创建信号通道
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("应用启动，按 Ctrl+C 优雅退出...")

    // 等待信号
    sig := <-sigChan
    fmt.Printf("收到信号: %v，开始优雅关闭...\n", sig)

    // defer 会自动调用 rm.Cleanup()
}
```

## 最佳实践

### 1. 总是使用缓冲通道
```go
// ✅ 推荐：使用缓冲通道
sigChan := make(chan os.Signal, 1)

// ❌ 不推荐：使用无缓冲通道
sigChan := make(chan os.Signal)
```

### 2. 及时处理信号
```go
// ✅ 推荐：及时处理信号
go func() {
    sig := <-sigChan
    handleShutdown(sig)
}()

// ❌ 不推荐：信号处理太复杂
func handleComplexSignal() {
    // 复杂的逻辑可能会延迟信号处理
}
```

### 3. 区分不同信号
```go
// ✅ 推荐：区分处理不同信号
switch sig {
case syscall.SIGINT:
    handleGracefulShutdown()
case syscall.SIGHUP:
    handleConfigReload()
}
```

### 4. 使用 context 超时控制
```go
// ✅ 推荐：使用 context 控制关闭超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := server.Shutdown(ctx)
if err != nil {
    log.Printf("服务器关闭超时: %v", err)
}
```

## 在你的项目中的应用

回到你的 `novel-resource-management` 项目：

```go
// 在 main.go 中的应用
func main() {
    // 创建信号通道
    sigChan := make(chan os.Signal, 1)

    // 注册信号监听
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 启动 Fabric Gateway
    gateWay := connectToFabric()
    defer gateWay.Close()

    // 启动事件监听器
    go startEventListener(gateWay)

    // 启动 HTTP 服务器
    server := api.NewServer(gateWay)
    go server.Start(":8080")

    // 主程序等待信号
    fmt.Println("服务已启动，等待停止信号...")
    sig := <-sigChan
    fmt.Printf("收到信号: %v\n", sig)

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("服务器关闭失败: %v", err)
    }

    fmt.Println("所有服务已停止")
}
```

## 总结

信号处理是 Go 语言中实现优雅关闭的标准模式：

1. **创建信号通道**：`make(chan os.Signal, 1)`
2. **注册信号监听**：`signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)`
3. **阻塞等待信号**：`<-sigChan`
4. **优雅关闭**：清理资源、关闭连接、保存状态

这种模式确保了程序在收到退出信号时能够：
- 处理完现有的请求
- 关闭数据库连接
- 停止后台的 goroutine
- 释放网络连接
- 完成其他清理工作

这是编写生产环境服务器的必备技能！