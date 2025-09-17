# pre

PS：先到根目录，然后到 test-network

./network.sh up

./network.sh createChannel

# 1. Set up environment variables

1. 环境变量:
   source set-env.sh

2. 链码并初始化:
   ./network.sh deployCC -ccn novel-basic -ccp ../novel-resource-events -ccl go -cci InitLedger

执行的原因是：

这个脚本设置了 Hyperledger Fabric 测试网络所需的环境变量，包括：

- 命令行工具路径 (../bin)
- Fabric 配置文件路径 (../config/)
- TLS 连接配置
- Org1 的 MSP ID 和证书路径
- Peer 节点地址和端口

<!-- export $(./setOrgEnv.sh Org1 | xargs) -->

then:
source set-env.sh

# 2. Deploy your novel-resource-events chaincode

./network.sh deployCC -ccn novel-basic -ccp ../novel-resource-events -ccl go -cci InitLedger

# 3. Then invoke the chaincode

// create
peer chaincode invoke -C mychannel -n novel-basic -c '{"function":"CreateNovel","Args":["novel1","The Great
Novel","Author1","2025-08-29","Fiction","A great story"]}'

// find
peer chaincode query -C mychannel -n novel-basic -c '{"function":"GetAllNovels","Args":[]}'

// delete
peer chaincode invoke -C mychannel -n novel-basic -c '{"function":"DeleteNovel","Args":["test-novel-001"]}'

# PS

peer chaincode invoke -C mychannel -n novel-basic -c '{"function":"CreateNovel","Args":["novel1","The Great
Novel","Author1","2025-08-29","Fiction","A great story"]}'
2025-08-29 16:18:23.782 CST 0001 INFO [chaincodeCmd] InitCmdFactory -> Retrieved channel (mychannel) orderer endpoint: orderer.example.com:7050
Error: error getting broadcast client: orderer client failed to connect to orderer.example.com:7050: failed to create new connection: connection error: desc = "transport: error while dialing: dial tcp: lookup orderer.example.com: no such host"

## 我执行查询的时候可以，写入的时候就会出现问题：

- 查询 (peer chaincode query)：只连接到 peer 节点 (localhost:7051)，这个可以直接访问
- 写入 (peer chaincode invoke)：需要连接到 orderer 节点 (orderer.example.com:7050) 进行事务广播

你的 peer 节点配置在 localhost:7051 所以查询正常，但 orderer 节点使用 orderer.example.com
这个主机名，系统无法解析这个域名到正确的 IP 地址。

解决方案：在 /etc/hosts 文件中添加：
127.0.0.1 orderer.example.com

## 那为什么 init 的时候能够运行成功呢？

⏺ 这是个很好的问题！init（创建）操作能够成功是因为：

创建操作 通常是通过 网络脚本 (network.sh deployCC) 在 容器内部 执行的，容器内部的 DNS 配置能够解析 orderer.example.com
到正确的容器 IP。

而你现在在 容器外部 直接执行 peer chaincode invoke，这时使用的是 宿主机的 DNS 解析，无法识别 orderer.example.com。

简单说：

- 容器内执行：DNS 正常，能解析 orderer.example.com → 创建成功
- 宿主机执行：DNS 无法解析 orderer.example.com → 删除失败

## 那为什么 gateway 可以呢？

Gateway SDK（你在脚本中使用的）与 CLI 命令 的网络配置方式不同：

1. Gateway SDK：

   - 使用 连接配置文件 (connection-org1.json)
   - 自动处理网络发现和负载均衡
   - 可以在配置中指定具体的 IP 地址或主机名映射

2. CLI 命令 (peer chaincode invoke)：

   - 使用 环境变量 配置
   - 直接连接指定的 orderer 地址
   - 依赖系统的 DNS 解析

让我检查你的连接配置文件：

⏺ Read(test-network/organizations/peerOrganizations/org1.example.com/connection-org1.json)
⎿ Read 50 lines (ctrl+r to expand)

⏺ 关键发现！Gateway SDK 的连接配置使用的是 localhost:7051 直接连接 peer，而 CLI 命令使用 orderer.example.com:7050。

Gateway SDK 通过 服务发现 自动找到 orderer，不依赖 DNS 解析。这就是为什么你的脚本能用 Gateway 成功读写，但 CLI 命令失败。

根本原因总结：

- Gateway SDK：使用服务发现，自动处理网络拓扑
- CLI 命令：需要显式 DNS 解析 orderer.example.com

解决方案：添加 hosts 映射或改用服务发现方式。

# TODO, setEvent 还没有开始
