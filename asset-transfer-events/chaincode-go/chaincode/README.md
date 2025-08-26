# Hyperledger Fabric 事件监听指南 (Go 版本)

## 监听 `ctx.GetStub().SetEvent("UpdateAsset", assetJSON)` 事件

### 1. 基本监听代码

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/hyperledger/fabric-gateway/pkg/client"
    "github.com/hyperledger/fabric-gateway/pkg/hash"
)

const (
    channelName   = "mychannel"
    chaincodeName = "events"
)

func listenEvents() {
    // 建立连接
    clientConnection := newGrpcConnection()
    defer clientConnection.Close()

    id := newIdentity()
    sign := newSign()

    gateway, err := client.Connect(
        id,
        client.WithSign(sign),
        client.WithHash(hash.SHA256),
        client.WithClientConnection(clientConnection),
    )
    if err != nil {
        panic(err)
    }
    defer gateway.Close()

    network := gateway.GetNetwork(channelName)
    // context最重要的作用是用于控制协程的生命周期（如取消、超时）以及在协程间传递请求范围的数据。
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() //可以在结束的时候跳出

    events, err := network.ChaincodeEvents(ctx, chaincodeName)
    if err != nil {
        panic(fmt.Errorf("failed to start chaincode event listening: %w", err))
    }

    // 异步处理事件，goroutine
    go func() {
        for event := range events {
            switch event.EventName {
            case "UpdateAsset":
                handleUpdateAsset(event.Payload)
            case "CreateAsset":
                handleCreateAsset(event.Payload)
            case "DeleteAsset":
                handleDeleteAsset(event.Payload)
            case "TransferAsset":
                handleTransferAsset(event.Payload)
            }
        }
    }()
}

func handleUpdateAsset(payload []byte) {
    var asset Asset //定义一个变量
    //用来在这里接值
    if err := json.Unmarshal(payload, &asset); err != nil {
        fmt.Printf("Error parsing UpdateAsset payload: %v\n", err)
        return
    }

    fmt.Printf("UpdateAsset事件 - ID: %s, Owner: %s, Value: %d\n",
        asset.ID, asset.Owner, asset.AppraisedValue)
}
```

### 2. 事件重放

```go
func replayEvents(startBlock uint64) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    events, err := network.ChaincodeEvents(ctx, chaincodeName,
        client.WithStartBlock(startBlock))
    if err != nil {
        panic(fmt.Errorf("failed to replay events: %w", err))
    }

    // 这里叫“重放”是因为我们可以从指定的区块高度（startBlock）开始，重新读取并处理历史上已经发生过的链码事件。
    // 这对于恢复服务、数据同步、审计等场景非常有用。通过重放事件，可以确保应用不会错过任何重要的链码事件。
    for event := range events {
        if event.EventName == "UpdateAsset" {
            fmt.Printf("重放UpdateAsset: %s\n", string(event.Payload))
        }
    }
}
```

### 3. 事件过滤

```go
func listenSpecificEvents(eventName string) {
    ctx := context.Background()
    events, _ := network.ChaincodeEvents(ctx, chaincodeName)

    for event := range events {
        if event.EventName == eventName {
            // 只处理指定类型的事件
            fmt.Printf("收到%s事件: %s\n", eventName, string(event.Payload))
        }
    }
}
```

### 4. 完整示例

查看完整示例代码: `../application-gateway-go/app.go`

该文件包含：

- 完整的连接建立流程
- 实时事件监听
- 事件重放功能
- 交易提交示例

### 5. 事件数据结构

当链码中的 `UpdateAsset` 函数执行时，会发出以下事件：

```json
{
  "eventName": "UpdateAsset",
  "payload": {
    "ID": "asset123",
    "Color": "blue",
    "Size": 10,
    "Owner": "Sam",
    "AppraisedValue": 200
  },
  "transactionId": "tx123...",
  "blockNumber": 12345
}
```

### 6. 错误处理

```go
func safeListenEvents() {
    events, err := network.ChaincodeEvents(ctx, chaincodeName)
    if err != nil {
        // 处理连接错误
        fmt.Printf("连接失败: %v\n", err)
        return
    }

    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("监听异常: %v\n", r)
        }
    }()

    for event := range events {
        // 处理事件
    }
}
```
