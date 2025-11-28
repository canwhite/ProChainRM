# pre

PS：先到根目录，然后到 test-network

./network.sh up

./network.sh createChannel

PS:如果 createChannel 卡住，注意查看下 9443 端口是不是被占了，kill 下

# 1. Set up environment variables

1. 环境变量:
   source set-env.sh

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

./network.sh deployCC \ 
-ccn novel-basic \ 
-ccp ../novel-resource-events \
-ccl go \
-cci InitLedger

PS: endorse 的时候有些问题, -ccv 2.0,所以这是一个很棒的问题

./network.sh deployCC \
 -ccn novel-basic \
 -ccp ../novel-resource-events \
 -ccl go \
 -cci InitLedger \
 -ccep "OR('Org1MSP.member','Org2MSP.member')"

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

### 关于ccep
  🎯 核心作用

  -ccep "OR('Org1MSP.member','Org2MSP.member')" 参数为你的 novel-basic 智能合约设置了灵活的背书策略。

  📝 具体含义

  1. OR 策略：使用逻辑 OR 操作符
  2. 两个组织：Org1MSP 和 Org2MSP
  3. 成员范围：每个组织的所有成员 (.member)

  最终效果：你的智能合约交易只需要以下任一条件满足即可：
  - Org1 的任何 peer 背书（比如 peer0.org1.example.com）
  - Org2 的任何 peer 背书（比如 peer0.org2.example.com）

  🔄 与其他策略的对比

  | 策略类型 | 示例                 | 要求           | 优点   | 缺点    |
  |------|--------------------|--------------|------|-------|
  | AND  | AND('Org1','Org2') | 两个组织都必须背书    | 安全性高 | 灵活性低  |
  | OR   | OR('Org1','Org2')  | 任一组织背书即可     | 灵活性高 | 安全性较低 |
  | 单一   | 'Org1MSP.member'   | 只有 Org1 成员背书 | 简单明确 | 限制性强  |

  🏢 实际应用场景

  对于你的 novel-basic 智能合约（小说资源管理），这个策略意味着：

  1. 小说创作：
    - Org1 的用户可以独立发布小说（只需 Org1 背书）
    - Org2 的用户可以独立发布小说（只需 Org2 背书）
    - 跨组织的协作也可以（任一组织支持即可）
  2. 资源管理：
    - 不会因为另一个组织暂时不可用而阻塞操作
    - 提高了系统的可用性和容错能力

  ⚖️ 安全性考虑

  虽然 OR 策略降低了背书要求，但在 Fabric 中仍然安全：

  1. 数字身份验证：每个背书都需要有效的数字证书
  2. MSP 验证：只有组织内的合法成员才能背书
  3. 通道权限控制：只有通道成员才能提交交易

  🎛 最佳实践建议

  根据你的业务需求，可以考虑不同的策略：

  高安全场景（重要数据）

  # 推荐用于敏感操作
  -ccep "AND('Org1MSP.admin','Org2MSP.admin')"

  高可用场景（日常操作）

  # 推荐用于一般业务操作（当前使用的）
  -ccep "OR('Org1MSP.member','Org2MSP.member')"

  分层权限场景

  # 读操作：任一组织即可
  -ccep "OR('Org1MSP.member','Org2MSP.member')"

  # 写操作：需要管理员权限
  -ccep "AND('Org1MSP.admin','Org2MSP.admin')"

  🔍 调试和验证

  部署后，你可以通过以下方式验证背书策略：

  # 查看链码的背书策略
  peer chaincode query --channel mychannel --name novel-basic --ctor
  '{"function":"getEndorsementPolicy"}'

  # 或者查看通道配置
  peer channel getinfo --channel mychannel

  总结

  这个 -ccep 参数是 Hyperledger Fabric 中配置智能合约背书策略的关键参数：

  - 作用：定义哪些组织成员需要背书交易
  - 策略：OR('Org1MSP.member','Org2MSP.member') - 任一组织背书即可
  - 效果：提高系统的灵活性和可用性
  - 应用：适合小说资源管理系统这类需要高可用性的场景

  选择合适的背书策略是 Fabric 网络设计中的重要决策，需要根据业务需求、安全要求和性能考虑来平衡！