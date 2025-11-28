# Go WaitGroup 完全教程

## WaitGroup 的本质

`WaitGroup` 是 Go 语言中用于等待一组 goroutine 完成的同步工具。它的本质是一个**计数器**，用来跟踪还有多少个 goroutine 没有完成。

## 通俗比喻：班级春游

想象一下**班级春游**的场景：

```
老师（主程序）拿着一个点名册（WaitGroup）：

1. 老师问："今天有几个同学要去春游？"
   老师在点名册上记下："总共 30 人"  ← wg.Add(30)

2. 同学们各自出发去春游（启动 goroutine）
   每个同学完成春游后向老师报到  ← wg.Done()

3. 老师在原地等，看着点名册
   等所有 30 人都报到，老师才回家  ← wg.Wait()
```

## WaitGroup 的内部原理

### 基本结构
```go
// Go 标准库中的 WaitGroup 结构（简化版）
type WaitGroup struct {
    noCopy noCopy          // 防止复制
    state1 [3]uint32      // 状态存储（32位系统）
    state2 [3]uint32      // 状态存储（64位系统）
}
```

### 简化版的内部实现
```go
// 为了理解，我们来看一个简化版的实现
type WaitGroup struct {
    counter int64    // 当前还在运行的任务数
    waiters int64    // 有多少个 goroutine 在等待
    sema    uint32   // 信号量，用于唤醒等待者
}

func (wg *WaitGroup) Add(delta int) {
    // 使用原子操作增加计数器
    state := atomic.AddInt64(&wg.counter, int64(delta))

    // 如果计数器变成 0，说明所有任务都完成了
    if state == 0 {
        // 唤醒所有等待的 goroutine
        for i := 0; i < int(atomic.LoadInt64(&wg.waiters)); i++ {
            runtime_Semrelease(&wg.sema, false, 0)
        }
    }
}

func (wg *WaitGroup) Done() {
    // Done() 实际上就是 Add(-1)
    wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
    // 如果还有任务没完成，就等待
    if atomic.LoadInt64(&wg.counter) > 0 {
        atomic.AddInt64(&wg.waiters, 1)
        runtime_Semacquire(&wg.sema)  // 阻塞等待
        atomic.AddInt64(&wg.waiters, -1)
    }
}
```

## 核心方法详解

### 1. Add(delta int) - 增加计数

```go
// 增加要等待的 goroutine 数量
wg.Add(3)  // 现在要等待 3 个 goroutine
wg.Add(1)  // 再增加 1 个，总共要等待 4 个
wg.Add(-1) // 减少 1 个，现在要等待 3 个
```

**为什么可以 Add？**
- WaitGroup 就像一个计数器，可以动态调整
- 你可以在启动 goroutine 之前增加计数
- 也可以在运行过程中动态调整

### 2. Done() - 完成

```go
// 相当于 wg.Add(-1)
wg.Done()  // 计数器减 1
```

### 3. Wait() - 等待

```go
// 阻塞等待，直到计数器为 0
wg.Wait()  // 如果还有 goroutine 没完成，这里会阻塞
```

## 完整示例：春游比喻

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup

    // 1. 老师统计要去春游的学生数量
    studentCount := 5
    fmt.Printf("老师：今天有 %d 个同学要去春游\n", studentCount)

    // 在点名册上记录要等待的人数
    wg.Add(studentCount)  // 等待 5 个同学

    // 2. 同学们各自出发（启动 goroutine）
    for i := 1; i <= studentCount; i++ {
        go func(studentID int) {
            fmt.Printf("同学 %d：我出发去春游了！\n", studentID)

            // 模拟春游时间
            time.Sleep(time.Duration(studentID) * time.Second)

            fmt.Printf("同学 %d：我玩完了，回家了！\n", studentID)

            // 同学回家，在点名册上划掉自己的名字
            wg.Done()  // 这个同学完成了
        }(i)
    }

    // 3. 老师在原地等待所有同学都回家
    fmt.Println("老师：我在这里等所有同学都回家...")
    wg.Wait()  // 阻塞，直到所有同学都调用了 Done()

    // 所有同学都回家后
    fmt.Println("老师：所有同学都回家了，我也回家了！春游结束！")
}
```

**运行结果**：
```
老师：今天有 5 个同学要去春游
老师：我在这里等所有同学都回家...
同学 1：我出发去春游了！
同学 2：我出发去春游了！
同学 3：我出发去春游了！
同学 4：我出发去春游了！
同学 5：我出发去春游了！
同学 1：我玩完了，回家了！
同学 2：我玩完了，回家了！
同学 3：我玩完了，回家了！
同学 4：我玩完了，回家了！
同学 5：我玩完了，回家了！
老师：所有同学都回家了，我也回家了！春游结束！
```

## Add 方法的灵活性

### 1. 批量添加
```go
var wg sync.WaitGroup

// 方式1：一次性添加多个
wg.Add(10)
for i := 0; i < 10; i++ {
    go func() {
        defer wg.Done()
        doWork(i)
    }()
}
wg.Wait()

// 方式2：逐个添加
for i := 0; i < 10; i++ {
    wg.Add(1)  // 每次启动一个 goroutine 前添加 1
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}
wg.Wait()
```

### 2. 动态调整
```go
var wg sync.WaitGroup

func dynamicWork() {
    // 启动基础任务
    wg.Add(2)
    go baseTask1()
    go baseTask2()

    // 根据条件启动额外任务
    if someCondition() {
        wg.Add(1)  // 动态增加任务
        go extraTask()
    }

    wg.Wait()  // 等待所有任务完成
}
```

### 3. 分阶段执行
```go
var wg sync.WaitGroup

func stagedWork() {
    // 第一阶段
    wg.Add(3)
    go stage1Task1()
    go stage1Task2()
    go stage1Task3()
    wg.Wait()  // 等待第一阶段完成

    fmt.Println("第一阶段完成，开始第二阶段...")

    // 第二阶段
    wg.Add(2)
    go stage2Task1()
    go stage2Task2()
    wg.Wait()  // 等待第二阶段完成

    fmt.Println("所有阶段完成！")
}
```

## 常见使用模式

### 1. 基本模式（推荐）
```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)  // 在启动 goroutine 前添加
    go func(id int) {
        defer wg.Done()  // 使用 defer 确保一定会调用
        doWork(id)
    }(i)
}
wg.Wait()
```

### 2. 工作池模式
```go
func workerPool() {
    var wg sync.WaitGroup

    // 启动多个 worker
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go worker(i, &wg)
    }

    wg.Wait()
}

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()

    for task := range taskChan {
        processTask(task)
    }
}
```

### 3. 批量处理模式
```go
func batchProcessing(items []string) {
    var wg sync.WaitGroup
    batchSize := 10

    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }

        batch := items[i:end]
        wg.Add(1)
        go func(b []string) {
            defer wg.Done()
            processBatch(b)
        }(batch)
    }

    wg.Wait()
}
```

## 错误使用示例

### 1. 忘记 Add
```go
// ❌ 错误：忘记 Add
var wg sync.WaitGroup

go func() {
    wg.Done()  // panic: WaitGroup is reused before previous Wait has returned
}()

wg.Wait()
```

### 2. Add 和 Done 数量不匹配
```go
// ❌ 错误：Add 和 Done 数量不匹配
var wg sync.WaitGroup

wg.Add(2)  // 等待 2 个 goroutine

go task1()  // 调用 wg.Done()
go task2()  // 没有调用 wg.Done() -> 永远等待

wg.Wait()  // 永远阻塞
```

### 3. 在 goroutine 内部 Add
```go
// ❌ 错误：在 goroutine 内部 Add
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    go func() {
        wg.Add(1)  // 在 goroutine 内部 Add，可能导致竞争
        defer wg.Done()
        doWork(i)
    }()
}

wg.Wait()  // 可能 Add 还没执行就 Wait 了
```

### 4. 复制 WaitGroup
```go
// ❌ 错误：复制 WaitGroup
var wg sync.WaitGroup

wg.Add(1)
go func(wgCopy sync.WaitGroup) {  // 复制了 WaitGroup
    defer wgCopy.Done()
    doWork()
}(wg)

wg.Wait()  // 可能导致问题
```

## 高级用法

### 1. 带错误处理的 WaitGroup
```go
type WorkerResult struct {
    ID    int
    Error error
    Data  interface{}
}

func workersWithErrors() error {
    var wg sync.WaitGroup
    resultChan := make(chan WorkerResult, 10)

    // 启动多个 worker
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            result := WorkerResult{ID: id}
            result.Data, result.Error = doWork(id)

            resultChan <- result
        }(i)
    }

    // 等待所有 worker 完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 收集结果和错误
    var errors []error
    var successes []interface{}

    for result := range resultChan {
        if result.Error != nil {
            errors = append(errors, result.Error)
        } else {
            successes = append(successes, result.Data)
        }
    }

    fmt.Printf("成功: %d, 失败: %d\n", len(successes), len(errors))

    if len(errors) > 0 {
        return fmt.Errorf("有 %d 个任务失败", len(errors))
    }

    return nil
}
```

### 2. 限制并发数量
```go
func limitedConcurrency(tasks []Task) {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 3)  // 最多 3 个并发

    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {
            defer wg.Done()

            // 获取信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            processTask(t)
        }(task)
    }

    wg.Wait()
}
```

### 3. 超时控制
```go
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
    done := make(chan struct{})

    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("等待超时")
    }
}
```

### 4. 分层 WaitGroup
```go
func layeredWaitGroup() {
    // 外层 WaitGroup
    var outerWg sync.WaitGroup

    // 启动多个工作组
    for group := 0; group < 3; group++ {
        outerWg.Add(1)
        go func(groupID int) {
            defer outerWg.Done()

            // 内层 WaitGroup
            var innerWg sync.WaitGroup

            // 每个组启动多个任务
            for task := 0; task < 5; task++ {
                innerWg.Add(1)
                go func(gID, tID int) {
                    defer innerWg.Done()
                    fmt.Printf("组 %d 任务 %d\n", gID, tID)
                    time.Sleep(time.Second)
                }(groupID, task)
            }

            innerWg.Wait()  // 等待当前组完成
            fmt.Printf("组 %d 完成所有任务\n", groupID)
        }(group)
    }

    outerWg.Wait()  // 等待所有组完成
    fmt.Println("所有组和任务都完成了")
}
```

## 实际应用案例

### 1. 网络爬虫
```go
func webCrawler(urls []string) {
    var wg sync.WaitGroup
    resultChan := make(chan string, len(urls))

    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()

            content, err := fetchURL(u)
            if err != nil {
                fmt.Printf("抓取 %s 失败: %v\n", u, err)
                return
            }

            resultChan <- content
        }(url)
    }

    // 等待所有抓取完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 处理结果
    for content := range resultChan {
        processContent(content)
    }
}
```

### 2. 文件处理
```go
func processFiles(filePaths []string) error {
    var wg sync.WaitGroup
    errorChan := make(chan error, len(filePaths))

    for _, filePath := range filePaths {
        wg.Add(1)
        go func(path string) {
            defer wg.Done()

            if err := processFile(path); err != nil {
                errorChan <- fmt.Errorf("处理文件 %s 失败: %v", path, err)
            }
        }(filePath)
    }

    // 等待所有文件处理完成
    go func() {
        wg.Wait()
        close(errorChan)
    }()

    // 检查错误
    var errors []error
    for err := range errorChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("处理过程中发生 %d 个错误", len(errors))
    }

    return nil
}
```

### 3. 数据库批量操作
```go
func batchInsert(records []Record) error {
    var wg sync.WaitGroup
    batchSize := 100
    errorChan := make(chan error, (len(records)+batchSize-1)/batchSize)

    for i := 0; i < len(records); i += batchSize {
        end := i + batchSize
        if end > len(records) {
            end = len(records)
        }

        batch := records[i:end]
        wg.Add(1)
        go func(b []Record, batchID int) {
            defer wg.Done()

            if err := insertBatch(b); err != nil {
                errorChan <- fmt.Errorf("批次 %d 插入失败: %v", batchID, err)
            }
        }(batch, i/batchSize)
    }

    go func() {
        wg.Wait()
        close(errorChan)
    }()

    // 检查错误
    for err := range errorChan {
        return err  // 返回第一个错误
    }

    return nil
}
```

## WaitGroup 与其他同步工具的对比

### WaitGroup vs Channel
```go
// 使用 WaitGroup
func withWaitGroup() {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            doWork(id)
        }(i)
    }

    wg.Wait()
}

// 使用 Channel
func withChannel() {
    done := make(chan struct{}, 10)

    for i := 0; i < 10; i++ {
        go func(id int) {
            doWork(id)
            done <- struct{}{}
        }(i)
    }

    // 等待 10 个完成信号
    for i := 0; i < 10; i++ {
        <-done
    }
}
```

**对比**：
- **WaitGroup**：更直观，专门用于等待
- **Channel**：更灵活，可以传递数据

### WaitGroup vs Mutex
```go
// WaitGroup：等待一组任务完成
func withWaitGroup() {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        doWork()
    }()
    wg.Wait()
}

// Mutex：保护共享资源
func withMutex() {
    var mu sync.Mutex
    var counter int

    for i := 0; i < 10; i++ {
        go func() {
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
}
```

## 最佳实践

### 1. 使用 defer
```go
// ✅ 推荐：使用 defer 确保一定会调用
go func() {
    defer wg.Done()

    // 即使这里发生 panic，Done 也会被调用
    doWork()
}()
```

### 2. 在正确的时机 Add
```go
// ✅ 推荐：在启动 goroutine 前添加
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()

// ❌ 不推荐：在 goroutine 内部添加
go func() {
    wg.Add(1)  // 可能竞争
    defer wg.Done()
    doWork()
}()
```

### 3. 不要复制 WaitGroup
```go
// ✅ 推荐：传递指针
func worker(wg *sync.WaitGroup) {
    defer wg.Done()
    doWork()
}

wg.Add(1)
go worker(&wg)
```

### 4. 错误处理
```go
// ✅ 推荐：使用 channel 收集错误
func withErrorHandling() error {
    var wg sync.WaitGroup
    errorChan := make(chan error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            if err := doWork(id); err != nil {
                errorChan <- err
            }
        }(i)
    }

    go func() {
        wg.Wait()
        close(errorChan)
    }()

    var errors []error
    for err := range errorChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("发生 %d 个错误", len(errors))
    }
    return nil
}
```

## 总结

**WaitGroup 的本质**：
1. **一个计数器**：跟踪还有多少个 goroutine 没完成
2. **原子操作**：保证并发安全
3. **信号量机制**：实现等待和唤醒

**为什么可以 Add**：
1. **动态计数**：它可以随时增加或减少要等待的数量
2. **灵活性**：支持复杂的并发场景
3. **原子安全**：使用原子操作保证并发安全

**核心用法**：
- `wg.Add(n)`：告诉 WaitGroup 要等待 n 个 goroutine
- `wg.Done()`：告诉 WaitGroup 其中一个完成了
- `wg.Wait()`：等待所有 goroutine 都完成

WaitGroup 是 Go 语言中实现**"等待一组任务完成"**的标准工具，简单而强大！它特别适合用于：
- 并行处理一批任务
- 工作池模式
- 批量操作
- 分阶段执行

掌握了 WaitGroup，你就掌握了 Go 并发编程的一个重要基础！