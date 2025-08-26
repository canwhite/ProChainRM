### **Hyperledger Fabric v2.5.x 快速入门教程**

本教程将指导您在 macOS 或 Linux 系统上使用 Hyperledger Fabric v2.5.13 搭建一个简单的测试网络，部署链码，并进行基本操作。以下步骤假设您使用的是 macOS（Apple M1/M2 或 Intel）或 Ubuntu 系统。

#### **1. 准备环境**

确保您的系统满足以下要求：

- **操作系统**: macOS 或 Linux（Windows 用户需使用 WSL2 或虚拟机）。
- **工具**:
  - Docker（版本 17.06.2-ce 或更高）及 Docker Compose（1.14.0 或更高）。
  - Go（1.17.x 或更高，用于链码开发）。
  - Node.js（12.x 或 14.x，推荐 14.x）。
  - Git（最新版本）。
  - curl（最新版本，避免 HTTP/2 错误）。
- **硬件**: 至少 4GB 内存和稳定的网络连接。

**安装步骤**（以 macOS 为例，使用 Homebrew）：

```bash
# 安装 Homebrew（若未安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装依赖
brew install docker docker-compose git go node
```

**验证安装**：

```bash
docker --version
go version
node --version
curl --version
```

#### **2. 下载 Fabric Samples 和二进制文件**

==========以下是 grok 给的方法==========
Hyperledger Fabric 提供了一个官方脚本 `install-fabric.sh` 来下载 `fabric-samples` 存储库和二进制文件。

```bash
# 下载并赋予执行权限
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
chmod +x install-fabric.sh

# 下载 Fabric v2.5.13 和 Fabric CA v1.5.15 的二进制文件和样本
./install-fabric.sh -f 2.5.13 -c 1.5.15
```

这会在当前目录创建 `fabric-samples` 文件夹，包含 `test-network` 示例和必要的二进制文件（`peer`、`orderer` 等）。

#### **3. 进入 Test Network 目录**

`test-network` 是 v2.5.x 中替代 `first-network` 的示例网络，包含两个组织（Org1、Org2）和一个排序节点。

```bash
cd fabric-samples/test-network
```

#### **4. 启动测试网络**

使用 `network.sh` 脚本启动网络。此脚本会创建一个包含两个对等节点（peer nodes）和一个排序节点（orderer）的网络。

```bash
./network.sh up
```

**说明**：

- 该命令会启动 Docker 容器，包括两个组织的对等节点（Org1 和 Org2）、一个排序服务和一个 Fabric CA。
- 证书和通道配置会自动生成。
- 检查 Docker 容器是否运行：
  ```bash
  docker ps
  ```
  您应该看到类似 `orderer.example.com`、`peer0.org1.example.com` 和 `peer0.org2.example.com` 的容器。

#### **5. 创建通道**

PS: 先说明下通道这个概念

```
在 Hyperledger Fabric 中，通道是一个私有子网络，用于隔离不同组织之间的交易和账本数据。

 * 数据隔离：每个通道维护一个独立的区块链账本，只有加入该通道的组织（及其对等节点）可以访问该账本的数据和交易。这确保了数据的隐私性和权限控制。

 * 交易协作：通道允许特定组织在私有环境中协作，执行交易（如资产转移），而其他未加入通道的组织无法看到这些交易。

* 链码执行：链码（智能合约）部署在通道上，只有通道成员可以调用链码执行交易逻辑。

* 一致性保证：通道内的对等节点通过排序服务（ordering service）同步账本，确保所有成员的账本副本一致。

```

`test-network` 默认使用通道 `mychannel`。创建通道的命令如下：

```bash
./network.sh createChannel
```

可以看有哪些 channel

```
peer channel list
```

要执行 peer 命令，需要先设置环境
eg:step 7
建议先写一个 sh 文件，然后

```
source set-env.sh
```

在 Fabric v2.5.x 的 test-network 中，运行 ./network.sh createChannel 创建 mychannel 时，Org1 和 Org2 的对等节点（peer0.org1.example.com 和 peer0.org2.example.com）默认自动加入通道，由脚本通过 **peer channel join** 实现，无需手动添加。**通道隔离交易和账本，仅成员节点参与**。生产环境或添加新节点时，需手动生成证书、启动节点并运行 peer channel join。可用 peer channel list 验证节点加入状态。test-network 简化了测试流程，生产需手动配置

这将创建一个名为 `mychannel` 的通道，并将 Org1 和 Org2 的对等节点加入通道。

```
+ . scripts/orderer.sh mychannel
+ '[' 0 -eq 1 ']'
+ res=0
Status: 201
{
	"name": "mychannel",
	"url": "/participation/v1/channels/mychannel",
	"consensusRelation": "consenter",
	"status": "active",
	"height": 1
}

Channel 'mychannel' created
Joining org1 peer to the channel.

```

#### **6. 部署链码**

链码（chaincode）是 Fabric 的智能合约。我们将使用 `fabric-samples` 提供的 `asset-transfer-basic` 链码示例（用 Go 编写）。

如果已经执行过了 set-env.sh,可以先查是否已经部署过了

```
 peer chaincode list --installed
```

```bash
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```

**说明**：

- `-ccn basic`: 链码名称为 `basic`。
- `-ccp ../asset-transfer-basic/chaincode-go`: 链码代码路径。
- `-ccl go`: 链码语言为 Go。
- 该命令会打包、安装、批准并提交链码到 `mychannel`。

#### **7. 测试链码**

现在网络和链码已部署，可以通过 `peer` 命令测试链码。

**配置参数**：
这里我简化了过程,执行下述命令即可:

```
source set-env.sh
```

set-env.sh 具体做的事情：

```
# 设置执行文件目录
export PATH=${PWD}/../bin:$PATH
# 设置配置文件目录
export FABRIC_CFG_PATH=${PWD}/../config/
# 标识对等节点的组织身份，用于认证和授权。
export CORE_PEER_LOCALMSPID=Org1MSP
# 提供加密通信，test-network 中默认启用以模拟生产环境。
export CORE_PEER_TLS_ENABLED=true
# 指定 peer0.org1.example.com 的 TLS 根证书路径。证书由 Fabric CA（ca_org1）在 ./network.sh up 时生成。验证对等节点的 TLS 身份。
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
# 指定 Org1 管理员用户的 MSP 配置路径。
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
# 指定对等节点 peer0.org1.example.com 的地址。
export CORE_PEER_ADDRESS=localhost:7051

```

**init**

```
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'
```

**查询链码**：

```bash
peer chaincode query -C mychannel -n basic -c '{"function":"GetAllAssets","Args":[]}'
```

**输出示例**：
您将看到类似以下的资产列表：

```json
[
  {"ID":"asset1","color":"blue","size":5,"owner":"Tomoko","appraisedValue":300},
  {"ID":"asset2","color":"red","size":5,"owner":"Brad","appraisedValue":400},
  ...
]
```

**创建新资产**：

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"CreateAsset","Args":["asset7","green","10","Alice","500"]}'
```

**再次查询以验证**：

```bash
peer chaincode query -C mychannel -n basic -c '{"function":"GetAllAssets","Args":[]}'
```

#### **8. 关闭网络**

完成测试后，关闭网络并清理 Docker 容器：

```bash
./network.sh down
```

这将停止并删除所有 Docker 容器、网络和生成的证书。

#### **9. 下一步（可选）**

- **开发自定义链码**：参考 `asset-transfer-basic/chaincode-go` 目录，编写自己的 Go 或 JavaScript 链码。
- **客户端应用**：使用 Fabric SDK（Node.js 或 Java）开发客户端应用与网络交互。参考 `fabric-samples/asset-transfer-basic/application-javascript`。
- **生产环境**：生产环境需使用 Fabric CA 管理证书，并配置多节点网络。参考官方文档：https://hyperledger-fabric.readthedocs.io/en/release-2.5/

---

### **与原教程的差异**

- **网络示例**：原教程使用 `first-network` 和 `byfn.sh`，现更新为 `test-network` 和 `network.sh`，更简洁且支持 v2.5.x。
- **Composer 已废弃**：原教程依赖 Hyperledger Composer，现直接使用 Fabric 链码和 SDK，避免过时工具。
- **链码部署**：v2.5.x 使用链码生命周期管理（`peer lifecycle chaincode`），比 v1.x 的部署流程更规范。
- **环境配置**：更新了依赖版本（Docker、Go、Node.js），并推荐 `install-fabric.sh` 脚本，简化安装。

---

### **注意事项**

- **平台兼容性**：若使用 Apple M1/M2，确保 Docker Desktop 支持 arm64，下载的二进制文件需为 `darwin-arm64`（v2.5.13 已支持）。
- **网络问题**：若遇到下载失败（如 HTTP/2 错误），尝试更新 curl 或使用 HTTP/1.1：
  ```bash
  brew install curl
  /usr/local/opt/curl/bin/curl --http1.1 -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
  ```
- **文档参考**：官方文档（https://hyperledger-fabric.readthedocs.io/en/release-2.5/）提供了更详细的配置说明。
- **社区支持**：如遇问题，可在 GitHub（https://github.com/hyperledger/fabric/issues）或 Hyperledger 邮件列表（https://lists.lfdecentralizedtrust.org）寻求帮助。

---

### **总结（100 字）**

原 Medium 教程基于 Fabric v1.x 和已废弃的 Composer，已不适用于 v2.5.x。本教程更新为 Fabric v2.5.13，使用 `test-network` 替代 `first-network`，通过 `install-fabric.sh` 安装环境，`network.sh` 启动网络，部署 `asset-transfer-basic` 链码，并测试资产管理功能。避免了 Composer，直接使用 Fabric 链码和 SDK，流程更简洁，适合初学者。若需生产环境，建议深入学习 Fabric CA 和多节点配置。参考官方文档以获取更多细节。

如果您需要更详细的链码开发或客户端应用指导，请告诉我！[](https://mycoralhealth.medium.com/start-your-own-hyperledger-blockchain-the-easy-way-5758cb4ed2d1)[](https://mycoralhealth.medium.com/build-a-dapp-on-hyperledger-the-easy-way-178c39e503fa)
