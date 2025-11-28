# Go os/exec 完全教程

## os/exec 是什么？

`os/exec` 是 Go 语言的标准包，用于**执行外部命令**。它就像是在 Go 程序中打开了一个"终端窗口"，可以运行各种系统命令。

## 基本概念

### 通俗比喻
想象你的 Go 程序是一个**指挥官**，`os/exec` 就是**传令兵**：

```
Go 程序（指挥官） → os/exec（传令兵） → 系统命令（士兵）
     │                   │                    │
   发出指令            传达指令            执行任务
     │                   │                    │
   检查结果            返回结果            完成任务
```

## 基本用法

### 1. 导入包
```go
import (
    "os/exec"
    "fmt"
    "log"
)
```

### 2. 创建命令
```go
// 基本语法：exec.Command("命令名", "参数1", "参数2", ...)
cmd := exec.Command("ls", "-l", "/home")
```

### 3. 执行命令的不同方式

#### 方式1：Run() - 只执行，不获取输出
```go
func runCommand() {
    // 执行命令，不关心输出
    cmd := exec.Command("docker", "--version")
    err := cmd.Run()  // 如果命令执行失败，err 不为 nil
    if err != nil {
        log.Printf("命令执行失败: %v", err)
    }
    fmt.Println("命令执行完成")
}
```

#### 方式2：CombinedOutput() - 获取标准输出和错误
```go
func runWithOutput() {
    // 执行命令并获取输出
    cmd := exec.Command("echo", "Hello, World!")
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("输出: %s\n", string(output))
    // 结果: 输出: Hello, World!
}
```

#### 方式3：Output() - 只获取标准输出
```go
func runWithStdoutOnly() {
    // 只获取标准输出，错误信息通过 err 返回
    cmd := exec.Command("date")
    output, err := cmd.Output()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("当前时间: %s\n", string(output))
}
```

## 在你的项目中的实际应用

### 1. MongoDB 连接检查
```go
// 代码来自 deploy.go 第162行
checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
if output, err := checkCmd.CombinedOutput(); err != nil {
    return fmt.Errorf("MongoDB连接失败: %v, 输出: %s", err, string(output))
}
```

**解释**：
- `mongosh`：命令名（MongoDB shell）
- `realMongoURI`：第一个参数（连接字符串）
- `"--eval"`：第二个参数
- `"db.adminCommand('ping')"`：第三个参数

**相当于执行**：
```bash
mongosh "mongodb://user:pass@127.0.0.1:27017/admin" --eval "db.adminCommand('ping')"
```

### 2. Docker 版本检查
```go
// 代码来自 deploy.go 第238行
dockerCmd := exec.Command("docker", "--version")
if err := dockerCmd.Run(); err != nil {
    return fmt.Errorf("Docker未运行或未安装: %v", err)
}
fmt.Println("✅ Docker服务正常")
```

**解释**：
- 使用 `Run()` 因为只关心命令是否成功
- 不需要获取输出版本信息
- 如果 Docker 安装正常，命令返回 nil 错误

### 3. Docker Compose 操作
```go
// 代码来自 deploy.go 第246行
exec.Command("docker-compose", "down").Run()

// 代码来自 deploy.go 第252行
cmd := exec.Command("docker-compose", "up", "-d")
cmd.Stdout = os.Stdout  // 标准输出重定向
cmd.Stderr = os.Stderr  // 标准错误重定向
if err := cmd.Run(); err != nil {
    return fmt.Errorf("Docker Compose启动失败: %v", err)
}
```

**解释**：
- `docker-compose down`：停止并删除容器
- `docker-compose up -d`：后台启动服务
- 重定向输出到终端，用户可以看到 Docker 的输出

### 4. 服务状态检查
```go
// 代码来自 deploy.go 第265行
statusCmd := exec.Command("docker-compose", "ps")
var statusOutput []byte
statusOutput, err = statusCmd.CombinedOutput()
if err != nil {
    return fmt.Errorf("检查服务状态失败: %v", err)
}
fmt.Printf("📊 服务状态:\n%s", string(statusOutput))
```

**解释**：
- 使用 `CombinedOutput()` 获取完整输出
- 输出内容显示所有容器的状态

### 5. 健康检查
```go
// 代码来自 deploy.go 第276行
healthCmd := exec.Command("curl", "-s", "http://localhost:8080/health")
if output, err := healthCmd.CombinedOutput(); err == nil {
    if strings.Contains(string(output), "ok") {
        fmt.Println("✅ 服务健康检查通过")
        return nil
    }
}
```

**解释**：
- `curl -s`：静默模式，不显示进度
- 检查 `/health` 端点是否返回 "ok"
- 如果成功，服务健康

## 带 Context 的用法

### 基本语法
```go
exec.CommandContext(ctx, "命令名", "参数1", "参数2", ...)
```

### 在你的项目中的应用
```go
// 设置10秒超时的上下文
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 使用带上下文的命令
checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
output, err := checkCmd.CombinedOutput()
```

### Context 的作用
1. **超时控制**：如果命令执行时间超过10秒，自动终止
2. **取消机制**：可以手动取消正在执行的命令
3. **资源管理**：确保命令不会永久阻塞

### 超时示例
```go
func timeoutExample() {
    // 设置2秒超时
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // 模拟一个耗时5秒的命令
    cmd := exec.CommandContext(ctx, "sleep", "5")

    output, err := cmd.CombinedOutput()
    if err != nil {
        // 2秒后命令会被自动终止
        fmt.Printf("命令超时或失败: %v\n", err)
        return
    }

    fmt.Println("命令执行成功")
}
```

## 高级用法

### 1. 设置工作目录
```go
func setWorkingDirectory() {
    cmd := exec.Command("ls", "-l")
    cmd.Dir = "/tmp"  // 设置工作目录

    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("临时目录内容:\n%s", string(output))
}
```

### 2. 设置环境变量
```go
func setEnvironmentVariables() {
    cmd := exec.Command("env")

    // 添加环境变量
    cmd.Env = append(os.Environ(), "MY_VAR=hello", "OTHER_VAR=world")

    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("环境变量:\n%s", string(output))
}
```

### 3. 分别处理标准输出和标准错误
```go
func separateOutputs() {
    cmd := exec.Command("ls", "/nonexistent", "/tmp")

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    fmt.Printf("标准输出: %s\n", stdout.String())
    fmt.Printf("标准错误: %s\n", stderr.String())
    fmt.Printf("错误信息: %v\n", err)
}
```

### 4. 实时输出处理
```go
func realTimeOutput() {
    cmd := exec.Command("ping", "-c", "3", "google.com")

    // 创建管道
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }

    // 启动命令
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    // 实时读取输出
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }

    // 等待命令完成
    if err := cmd.Wait(); err != nil {
        log.Printf("命令执行失败: %v", err)
    }
}
```

## 错误处理模式

### 1. 基本错误处理
```go
func basicErrorHandling() {
    cmd := exec.Command("nonexistent-command")

    output, err := cmd.CombinedOutput()
    if err != nil {
        // 处理不同类型的错误
        if execErr, ok := err.(*exec.Error); ok {
            if execErr.Err == exec.ErrNotFound {
                fmt.Println("命令不存在")
                return
            }
        }

        if exitErr, ok := err.(*exec.ExitError); ok {
            fmt.Printf("命令退出码: %d\n", exitErr.ExitCode())
            fmt.Printf("输出: %s\n", string(output))
            return
        }

        fmt.Printf("未知错误: %v\n", err)
        return
    }

    fmt.Printf("命令执行成功: %s\n", string(output))
}
```

### 2. 重试机制
```go
func retryCommand() error {
    var output []byte
    var err error

    // 重试3次
    for i := 0; i < 3; i++ {
        cmd := exec.Command("curl", "http://example.com")
        output, err = cmd.CombinedOutput()

        if err == nil {
            fmt.Println("命令执行成功")
            return nil
        }

        fmt.Printf("第%d次尝试失败: %v\n", i+1, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }

    return fmt.Errorf("重试3次后仍然失败: %v", err)
}
```

## 常见问题和解决方案

### 1. 路径包含空格
```go
func handleSpacesInPath() {
    // ❌ 错误：包含空格的路径
    // cmd := exec.Command("C:/Program Files/app.exe")

    // ✅ 正确：使用引号
    cmd := exec.Command("C:/Program Files/app.exe")

    // ✅ 正确：或者使用绝对路径
    cmd := exec.Command(`"C:\Program Files\app.exe"`)

    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("输出: %s\n", string(output))
}
```

### 2. 命令注入防护
```go
func preventCommandInjection(userInput string) {
    // ❌ 危险：直接使用用户输入
    // cmd := exec.Command("echo", userInput)

    // ✅ 安全：验证输入
    if strings.Contains(userInput, ";") || strings.Contains(userInput, "&") {
        log.Printf("非法输入: %s", userInput)
        return
    }

    cmd := exec.Command("echo", userInput)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    fmt.Printf("输出: %s\n", string(output))
}
```

### 3. 输出编码处理
```go
func handleOutputEncoding() {
    cmd := exec.Command("ls", "-l")

    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("命令执行失败: %v", err)
        return
    }

    // 检查输出编码
    if !utf8.Valid(output) {
        // 尝试转换编码
        output, err = iconv.ConvertString(string(output), "utf-8", "gbk")
        if err != nil {
            log.Printf("编码转换失败: %v", err)
            return
        }
    }

    fmt.Printf("输出: %s\n", string(output))
}
```

## 性能优化

### 1. 并发执行多个命令
```go
func concurrentCommands() {
    var wg sync.WaitGroup

    // 执行多个命令
    commands := [][]string{
        {"ping", "-c", "1", "google.com"},
        {"ping", "-c", "1", "baidu.com"},
        {"ping", "-c", "1", "github.com"},
    }

    for _, cmdArgs := range commands {
        wg.Add(1)
        go func(args []string) {
            defer wg.Done()

            cmd := exec.Command(args[0], args[1:]...)
            output, err := cmd.CombinedOutput()
            if err != nil {
                fmt.Printf("%s 失败: %v\n", strings.Join(args, " "), err)
                return
            }

            fmt.Printf("%s 成功\n", strings.Join(args, " "))
        }(cmdArgs)
    }

    wg.Wait()
}
```

### 2. 缓存命令结果
```go
func cacheCommandResult() {
    cache := make(map[string][]byte)
    cacheMutex := sync.RWMutex{}

    func getCachedOutput(command []string) ([]byte, error) {
        // 生成缓存键
        key := strings.Join(command, "_")

        // 尝试从缓存获取
        cacheMutex.RLock()
        if output, exists := cache[key]; exists {
            cacheMutex.RUnlock()
            return output, nil
        }
        cacheMutex.RUnlock()

        // 缓存中没有，执行命令
        cmd := exec.Command(command[0], command[1:]...)
        output, err := cmd.CombinedOutput()
        if err != nil {
            return nil, err
        }

        // 存入缓存
        cacheMutex.Lock()
        cache[key] = output
        cacheMutex.Unlock()

        return output, nil
    }

    // 使用缓存的命令
    if output, err := getCachedOutput([]string{"hostname"}); err == nil {
        fmt.Printf("主机名: %s\n", string(output))
    }
}
```

## 在你的项目中总结

### 命令使用统计

| 命令 | 用途 | 执行方式 | 输出处理 |
|------|------|----------|----------|
| `mongosh` | MongoDB操作 | `CommandContext` + `CombinedOutput` | 检查连接状态 |
| `docker --version` | Docker检查 | `Run` | 无输出处理 |
| `docker-compose down` | 停止容器 | `Run` | 无输出处理 |
| `docker-compose up -d` | 启动服务 | `Run` + 重定向 | 实时显示输出 |
| `docker-compose ps` | 查看状态 | `CombinedOutput` | 显示服务状态 |
| `curl` | 健康检查 | `CombinedOutput` | 检查服务健康 |

### 使用模式

1. **状态检查类**：使用 `CombinedOutput()` 获取结果进行判断
2. **控制命令类**：使用 `Run()` 执行，不关心输出
3. **实时输出类**：重定向 `Stdout` 和 `Stderr`
4. **超时控制**：所有数据库相关命令使用 `CommandContext`

## 最佳实践

1. **总是检查错误**：所有外部命令都可能失败
2. **使用 Context**：设置合理的超时时间
3. **清理输出**：使用 `strings.TrimSpace()` 处理结果
4. **日志记录**：记录命令执行结果，便于调试
5. **安全考虑**：验证用户输入，防止命令注入

## 总结

`os/exec` 包让 Go 程序能够：
- 执行系统命令
- 获取命令输出
- 控制命令执行环境
- 处理命令执行错误
- 实现超时控制

在你的项目中，`os/exec` 主要用于：
1. **数据库管理**：MongoDB 连接和副本集配置
2. **容器管理**：Docker 和 Docker Compose 操作
3. **服务监控**：健康检查和状态查询

掌握了 `os/exec`，你就可以在 Go 程序中自动化地执行各种系统命令了！