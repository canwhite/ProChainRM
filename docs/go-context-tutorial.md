# Go 语言 Context 完全教程

## Context 是什么？

Go 的 `context` 包提供了一种在 API 边界之间传递请求范围的值、取消信号和超时的机制。它是 Go 语言处理并发、超时和取消的核心工具。

## Context.WithTimeout 的工作原理

### ❌ 常见误解
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```
**不是** 10秒后结束整个函数，而是 10秒后让**所有使用这个 context 的操作**超时取消。

### ✅ 正确理解
```go
func connectMongoDBWithTimeout() error {
    // 设置一个10秒超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()  // 确保函数退出时取消资源

    // 这个命令会在10秒内执行，超过10秒会被取消
    cmd := exec.CommandContext(ctx, "mongosh", "mongodb://...", "--eval", "db.adminCommand('ping')")

    output, err := cmd.CombinedOutput()

    // 检查错误类型
    if err != nil {
        // 如果是超时错误
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("MongoDB连接超时（10秒）")
        }
        // 其他错误
        return fmt.Errorf("MongoDB连接失败: %v", err)
    }

    fmt.Println("MongoDB连接成功")
    return nil
}
```

## 为什么需要 Context？

### 1. 防止无限等待

```go
// ❌ 没有context的版本 - 可能永远卡住
func badMongoConnect() error {
    cmd := exec.Command("mongosh", "mongodb://...", "--eval", "db.adminCommand('ping')")
    // 如果MongoDB挂了，这里可能永远不返回！
    output, err := cmd.CombinedOutput()
    return err
}

// ✅ 有context的版本 - 10秒超时
func goodMongoConnect() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "mongosh", "mongodb://...", "--eval", "db.adminCommand('ping')")
    // 最多等10秒，超时自动取消
    output, err := cmd.CombinedOutput()
    return err
}
```

### 2. 优雅取消操作

```go
func processWithTimeout() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 可以同时启动多个操作
    go func() {
        select {
        case <-time.After(10 * time.Second):
            fmt.Println("这个操作会被取消")
        case <-ctx.Done():
            fmt.Println("上下文被取消，操作退出")
        }
    }()

    // 模拟工作
    time.Sleep(2 * time.Second)
    fmt.Println("主操作完成")
}
```

## Context 的四种创建方式

### 1. WithTimeout（最常用）

设置超时时间，超过时间自动取消。

```go
func operationWithTimeout() error {
    // 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    // 模拟网络请求
    result := make(chan string, 1)
    go func() {
        // 模拟耗时操作
        time.Sleep(5 * time.Second)
        result <- "完成"
    }()

    // 等待结果或超时
    select {
    case res := <-result:
        fmt.Printf("操作成功: %s\n", res)
        return nil
    case <-ctx.Done():
        return fmt.Errorf("操作超时: %v", ctx.Err())
    }
}
```

**实际应用场景**：
```go
func fetchDataFromAPI(url string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 处理响应...
    return nil
}
```

### 2. WithCancel（手动取消）

手动触发取消操作，用于控制多个 goroutine。

```go
func operationWithCancel() error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 启动监控goroutine
    done := make(chan bool)
    go func() {
        // 模拟监控某个条件
        for i := 0; i < 10; i++ {
            time.Sleep(1 * time.Second)
            select {
            case <-ctx.Done():
                fmt.Println("监控被取消")
                return
            default:
                fmt.Printf("监控中... %d/10\n", i+1)
            }
        }
        done <- true
    }()

    // 模拟某个条件触发取消
    time.Sleep(3 * time.Second)
    fmt.Println("触发取消条件")
    cancel()  // 手动取消

    <-done
    return nil
}
```

**实际应用场景**：
```go
func handleRequestWithGracefulShutdown() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 监听系统信号
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        fmt.Println("收到停止信号，开始优雅关闭...")
        cancel()
    }()

    // 处理请求
    for {
        select {
        case <-ctx.Done():
            fmt.Println("服务关闭")
            return
        default:
            // 处理请求...
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

### 3. WithDeadline（指定时间点）

设置具体的截止时间点。

```go
func operationWithDeadline() error {
    // 设置到今晚10点截止
    deadline := time.Date(time.Now().Year(), time.Now().Month(),
                         time.Now().Day(), 22, 0, 0, 0, time.Local)

    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()

    for {
        select {
        case <-time.After(1 * time.Second):
            fmt.Printf("工作中... %s\n", time.Now().Format("15:04:05"))
        case <-ctx.Done():
            return fmt.Errorf("操作截止时间到: %v", ctx.Err())
        }
    }
}
```

**实际应用场景**：
```go
func scheduleTask() error {
    // 设置每天午夜截止
    midnight := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
    ctx, cancel := context.WithDeadline(context.Background(), midnight)
    defer cancel()

    // 执行批量任务
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            fmt.Printf("任务在 %d 时截止\n", i)
            return ctx.Err()
        default:
            // 处理单个任务
            processItem(i)
        }
    }
    return nil
}
```

### 4. WithValue（传递参数）

在 context 中传递请求范围的数据。

```go
type contextKey string

const (
    requestIDKey contextKey = "requestID"
    userIDKey    contextKey = "userID"
    traceKey     contextKey = "trace"
)

func operationWithValue() {
    // 在context中传递参数
    ctx := context.WithValue(context.Background(), requestIDKey, "req-12345")
    ctx = context.WithValue(ctx, userIDKey, 42)
    ctx = context.WithValue(ctx, traceKey, "trace-abc-def")

    // 在下游操作中使用
    processRequest(ctx)
}

func processRequest(ctx context.Context) {
    if requestID, ok := ctx.Value(requestIDKey).(string); ok {
        fmt.Printf("处理请求: %s\n", requestID)
    }

    if userID, ok := ctx.Value(userIDKey).(int); ok {
        fmt.Printf("用户ID: %d\n", userID)
    }

    if traceID, ok := ctx.Value(traceKey).(string); ok {
        fmt.Printf("追踪ID: %s\n", traceID)
    }
}
```

**实际应用场景**：
```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 生成请求ID
        requestID := uuid.New().String()

        // 将请求ID存入context
        ctx := context.WithValue(r.Context(), "requestID", requestID)

        // 记录请求日志
        log.Printf("Request ID: %s, Method: %s, Path: %s", requestID, r.Method, r.URL.Path)

        // 继续处理请求
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func businessHandler(w http.ResponseWriter, r *http.Request) {
    // 从context获取请求ID
    if requestID, ok := r.Context().Value("requestID").(string); ok {
        // 在业务逻辑中使用请求ID
        processData(requestID)
    }
}
```

## Context 的传播和继承

### Context 树结构

```go
func demonstrateContextHierarchy() {
    // 根context
    rootCtx := context.Background()

    // 第一层：总超时
    ctx1, cancel1 := context.WithTimeout(rootCtx, 60*time.Second)
    defer cancel1()

    // 第二层：数据库操作超时
    ctx2, cancel2 := context.WithTimeout(ctx1, 10*time.Second)
    defer cancel2()

    // 第三层：添加追踪信息
    ctx3 := context.WithValue(ctx2, "traceID", "trace-123")

    // 使用最内层的context
    processDataWithTrace(ctx3)
}
```

### Context 取消的传播

```go
func contextCancellationPropagation() {
    ctx1, cancel1 := context.WithCancel(context.Background())
    defer cancel1()

    // 创建子context
    ctx2, cancel2 := context.WithTimeout(ctx1, 30*time.Second)
    defer cancel2()

    // 创建孙子context
    ctx3 := context.WithValue(ctx2, "data", "some-data")

    go func() {
        select {
        case <-ctx1.Done():
            fmt.Println("ctx1被取消:", ctx1.Err())
        case <-ctx2.Done():
            fmt.Println("ctx2被取消:", ctx2.Err())
        case <-ctx3.Done():
            fmt.Println("ctx3被取消:", ctx3.Err())
        }
    }()

    // 2秒后取消根context
    time.Sleep(2 * time.Second)
    cancel1()

    // 所有子context都会收到取消信号
    time.Sleep(1 * time.Second)
}
```

## Context 在实际项目中的应用

### 1. HTTP 服务器中的 Context

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
    // 从请求中获取context
    ctx := r.Context()

    // 设置操作超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 添加追踪信息
    traceID := r.Header.Get("X-Trace-ID")
    if traceID != "" {
        ctx = context.WithValue(ctx, "traceID", traceID)
    }

    // 执行业务逻辑
    if err := businessOperation(ctx); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write([]byte("操作完成"))
}

func businessOperation(ctx context.Context) error {
    // 模拟数据库查询
    if err := databaseQuery(ctx); err != nil {
        return err
    }

    // 模拟外部API调用
    if err := externalAPICall(ctx); err != nil {
        return err
    }

    return nil
}

func databaseQuery(ctx context.Context) error {
    // 使用context的数据库操作
    db, err := sql.Open("mysql", "user:password@/dbname")
    if err != nil {
        return err
    }

    // 设置查询超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var result string
    err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", 1).Scan(&result)
    if err != nil {
        return err
    }

    fmt.Printf("查询结果: %s\n", result)
    return nil
}

func externalAPICall(ctx context.Context) error {
    // 创建带context的HTTP请求
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        return err
    }

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 处理响应...
    return nil
}
```

### 2. 在你的 deploy.go 中的实际应用

```go
// configureMongoDBReplicaSet 配置MongoDB副本集
func configureMongoDBReplicaSet(hostIP string) error {
    fmt.Println("🔧 开始配置MongoDB副本集...")

    // 获取MongoDB认证信息
    mongoUser, mongoPass := getMongoConfig()

    // 设置10秒超时 - 防止MongoDB连接卡死
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()  // 确保函数退出时清理资源

    // 处理密码编码
    var actualPassword string
    if strings.Contains(mongoPass, "%40") {
        actualPassword = strings.ReplaceAll(mongoPass, "%40", "@")
    } else {
        actualPassword = mongoPass
    }
    encodedPassword := strings.ReplaceAll(actualPassword, "@", "%40")
    realMongoURI := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%s/%s?authSource=admin",
        mongoUser, encodedPassword, MongoPort, MongoDatabase)

    // 使用context的命令执行 - 检查连接
    checkCmd := exec.CommandContext(ctx, "mongosh", realMongoURI, "--eval", "db.adminCommand('ping')")
    if output, err := checkCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("MongoDB连接失败: %v, 输出: %s", err, string(output))
    }
    fmt.Println("✅ MongoDB连接成功")

    // 使用相同的context检查副本集状态
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
```

### 3. 微服务架构中的 Context

```go
// 服务间调用的Context传播
func callUserService(ctx context.Context, userID int) (*User, error) {
    // 创建带超时的context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // 从context获取追踪信息
    traceID, _ := ctx.Value("traceID").(string)

    // 创建请求
    req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://user-service/users/%d", userID), nil)
    if err != nil {
        return nil, err
    }

    // 传播追踪信息
    if traceID != "" {
        req.Header.Set("X-Trace-ID", traceID)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 从请求中获取用户ID
    userID := extractUserID(r)

    // 设置操作超时
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // 调用用户服务
    user, err := callUserService(ctx, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 返回用户信息
    json.NewEncoder(w).Encode(user)
}
```

## Context 最佳实践

### 1. Context 传递规则

```go
// ✅ 正确：将context作为第一个参数
func databaseOperation(ctx context.Context) error {
    // 使用传入的context
    cmd := exec.CommandContext(ctx, "mongosh", "--eval", "db.stats()")
    return cmd.Run()
}

// ❌ 错误：创建新的context
func databaseOperation() error {
    // 忽略了上游的取消信号
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, "mongosh", "--eval", "db.stats()")
    return cmd.Run()
}
```

### 2. Context 的生命周期管理

```go
func contextLifecycleExample() {
    // ✅ 正确：及时调用cancel
    func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()  // 确保资源清理

        doSomething(ctx)
    }()

    // ❌ 错误：忘记调用cancel
    func() {
        ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
        doSomething(ctx)  // 可能导致资源泄漏
    }()
}
```

### 3. 超时设置策略

```go
func timeoutStrategy() {
    // ✅ 合理的分层超时
    func() {
        // 总操作超时：30秒
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // 数据库操作超时：5秒
        dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
        defer dbCancel()

        if err := databaseOperation(dbCtx); err != nil {
            log.Printf("数据库操作失败: %v", err)
            return
        }

        // API调用超时：10秒
        apiCtx, apiCancel := context.WithTimeout(ctx, 10*time.Second)
        defer apiCancel()

        if err := apiOperation(apiCtx); err != nil {
            log.Printf("API操作失败: %v", err)
            return
        }
    }()
}
```

### 4. 错误处理模式

```go
func errorHandlingPatterns(ctx context.Context) error {
    // 方案1：区分超时和其他错误
    result := make(chan error, 1)

    go func() {
        result <- lengthyOperation()
    }()

    select {
    case err := <-result:
        return err
    case <-ctx.Done():
        // 区分不同类型的取消
        switch ctx.Err() {
        case context.DeadlineExceeded:
            return fmt.Errorf("操作超时")
        case context.Canceled:
            return fmt.Errorf("操作被取消")
        default:
            return fmt.Errorf("上下文错误: %v", ctx.Err())
        }
    }
}

// 方案2：包装context错误
func wrappedContextErrors(ctx context.Context) error {
    if err := someOperation(ctx); err != nil {
        // 检查是否是context相关的错误
        if ctx.Err() != nil {
            return fmt.Errorf("操作失败: %v (原因: %w)", err, ctx.Err())
        }
        return fmt.Errorf("操作失败: %w", err)
    }
    return nil
}
```

### 5. Context 的合理使用

```go
func contextUsage() {
    // ✅ 合理使用：用于外部调用和超时控制
    func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // HTTP请求
        resp, err := http.DefaultClient.Get("https://api.example.com")
        if err != nil {
            log.Fatal(err)
        }
        defer resp.Body.Close()

        // 数据库查询
        db.QueryRowContext(ctx, "SELECT * FROM users")
    }()

    // ❌ 不合理使用：简单的计算不需要context
    func() {
        // 这是过度使用context
        ctx := context.Background()
        result := calculateSomething(ctx)  // 简单计算不需要context
    }()

    // ✅ 合理使用：并发控制和取消
    func() {
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        var wg sync.WaitGroup
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                worker(ctx, id)
            }(i)
        }

        // 某个条件触发取消
        time.Sleep(5 * time.Second)
        cancel()

        wg.Wait()
    }()
}
```

## Context 的高级用法

### 1. Context 和并发模式

```go
// 使用context控制多个goroutine
func workerPool(ctx context.Context) {
    const numWorkers = 5
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // 启动工作goroutine
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for {
                select {
                case job := <-jobs:
                    results <- processJob(job)
                case <-ctx.Done():
                    fmt.Printf("Worker %d 停止\n", id)
                    return
                }
            }
        }(i)
    }

    // 分发任务
    go func() {
        for i := 0; i < 100; i++ {
            select {
            case jobs <- i:
            case <-ctx.Done():
                return
            }
        }
        close(jobs)
    }()

    // 等待完成
    go func() {
        wg.Wait()
        close(results)
    }()

    // 处理结果
    for result := range results {
        fmt.Printf("结果: %d\n", result)
    }
}
```

### 2. Context 和链式操作

```go
func pipelineWithTimeout() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 阶段1：数据获取
    data, err := stage1(ctx)
    if err != nil {
        return fmt.Errorf("阶段1失败: %w", err)
    }

    // 阶段2：数据处理
    processed, err := stage2(ctx, data)
    if err != nil {
        return fmt.Errorf("阶段2失败: %w", err)
    }

    // 阶段3：数据存储
    err = stage3(ctx, processed)
    if err != nil {
        return fmt.Errorf("阶段3失败: %w", err)
    }

    return nil
}

func stage1(ctx context.Context) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // 实现阶段1逻辑...
    return []byte("data"), nil
}

func stage2(ctx context.Context, data []byte) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
    defer cancel()

    // 实现阶段2逻辑...
    return data, nil
}

func stage3(ctx context.Context, data []byte) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // 实现阶段3逻辑...
    return nil
}
```

## 总结

`context.WithTimeout` 不是"霸道"，而是一种**保护机制**：

1. **防止无限等待**：避免网络操作卡死整个程序
2. **资源管理**：超时自动清理相关资源
3. **用户体验**：给用户明确的超时反馈
4. **系统稳定性**：防止资源泄漏和累积

### 核心原则

1. **Context作为第一个参数**：遵循Go的约定
2. **及时调用cancel**：避免资源泄漏
3. **合理设置超时**：根据操作特性设置合适的超时时间
4. **错误处理**：正确处理context相关的错误
5. **不要存储context**：context应该传递，不应该存储

### 在你的项目中

10秒超时设置是合理的，因为：
- MongoDB 连接通常应该在几秒内完成
- 10秒足够处理网络延迟和短暂的服务问题
- 防止因MongoDB挂掉导致整个部署脚本卡死

这是Go语言处理并发和超时的标准做法，体现了**"快速失败优于永远等待"**的设计哲学。