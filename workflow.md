### 基本流程

🎯 完整流程总结（小白版）

🏗️ 就像开一家新商场

| 步骤    | 现实类比         | 区块链操作        | 实际结果                          |
| ------- | ---------------- | ----------------- | --------------------------------- |
| 第 1 步 | 申请营业执照     | 创建 3 个配置文件 | 定义了 R3、R4、Orderer 三个"公司" |
| 第 2 步 | 刻公章、办身份证 | 生成证书          | 每个"公司"都有了官方证件          |
| 第 3 步 | 制作名片         | 生成连接配置      | 每个公司有了"联系方式"            |
| 第 4 步 | 租店面、装修     | 创建 Docker 容器  | 3 个"公司"开始实体运营            |
| 第 6 步 | 建立 VIP 会员群  | 创建通道          | R3 和 R4 有了专属"私聊群"         |

📋 最终得到什么？

- 3 个运行的容器：R3 餐厅、R4 餐厅、Orderer 邮局
- 1 个私聊通道：R3↔R4 专属交易通道
- 3 套证书：每个组织的"身份证"
- 1 个创世区块：通道的"出生证明"

🚀 你现在可以...

- R3 和 R4 可以在这个通道里安全交易
- 所有交易都会被 Orderer 排序记录
- 其他组织无法看到 R3 和 R4 的交易内容

简单说：你建立了一个只有 R3 和 R4 知道的"秘密交易市场"！

### 如何创建组织和节点

#### 前置准备

1. 进入测试目录网络

```
  cd /Users/zack/Desktop/fabric-samples/test-network
```

2. 确保工具可用

```
  export PATH=${PWD}/../bin:$PATH
  export FABRIC_CFG_PATH=${PWD}/configtx
```

#### 具体步骤

1. 创建组织配置文件

```
 1.1 创建R3组织配置文件

  创建文件：organizations/cryptogen/crypto-config-r3.yaml
  PeerOrgs:
    - Name: R3
      Domain: r3.example.com
      EnableNodeOUs: true
      Template:
        Count: 1
      Users:
        Count: 1

  1.2 创建R4组织配置文件

  创建文件：organizations/cryptogen/crypto-config-r4.yaml
  PeerOrgs:
    - Name: R4
      Domain: r4.example.com
      EnableNodeOUs: true
      Template:
        Count: 1
      Users:
        Count: 1

  1.3 创建排序组织配置文件
  // 排序组织配置文件就是用来定义“区块链网络里负责打包和排序交易的组织”的身份和节点信息的，没它区块链没法出块！

  创建文件：organizations/cryptogen/crypto-config-orderer-r3r4.yaml
  OrdererOrgs:
    - Name: Orderer
      Domain: r3r4.example.com
      EnableNodeOUs: true
      Specs:
        - Hostname: orderer

```

解释一下

```
 是的！创建了3个组织：

  1. R3组织（对等组织）- 有自己的对等节点
  2. R4组织（对等组织）- 有自己的对等节点
  3. Orderer组织（排序组织）- 负责排序服务

  这3个组织相互独立：
  - R3和R4是业务组织，处理交易
  - Orderer是基础设施组织，负责打包区块

  就像：
  - R3 = 银行A
  - R4 = 银行B
  - Orderer = 中央银行清算系统
```

2. 生成证书
   PS：排序

```
  # 清理旧数据（如果有）
  rm -rf organizations/peerOrganizations/r3.example.com
  rm -rf organizations/peerOrganizations/r4.example.com
  rm -rf organizations/ordererOrganizations/r3r4.example.com

  # 生成R3组织证书



  cryptogen generate --config=./organizations/cryptogen/crypto-config-r3.yaml --output="organizations"

  # 生成R4组织证书
  cryptogen generate --config=./organizations/cryptogen/crypto-config-r4.yaml --output="organizations"

  # 生成排序组织证书
  cryptogen generate --config=./organizations/cryptogen/crypto-config-orderer-r3r4.yaml
  --output="organizations"


```

3. 创建连接配置文件

```
  创建文件：organizations/ccp-r3r4-generate.sh（复制现有的ccp-generate.sh并修改）
  运行：
  # 生成R3和R4的连接配置文件
  ./organizations/ccp-r3r4-generate.sh
```

具体修改的内容吗

```
⏺ 📋 第三步：连接配置文件的作用和修改

  🎯 作用是什么？

  连接配置文件（Connection Profile）是告诉你的应用程序：
  - "R3组织的对等节点在哪里？"
  - "R4组织的对等节点在哪里？"
  - "如何连接它们？"

  就像给应用程序一个"地址簿"，让它知道如何找到各个节点。

  🔧 需要修改什么？

  你可以偷懒不修改，直接复制现有的然后运行：

  # 简单方法：直接复制现有的
  cp organizations/ccp-generate.sh organizations/ccp-r3r4-generate.sh

  # 然后运行
  ./organizations/ccp-r3r4-generate.sh

  或者手动修改（推荐理解）：

  打开 organizations/ccp-r3r4-generate.sh，找到这几行：

  # 原来的是ORG1和ORG2，改成R3和R4
  ORG_LOWER="r3"
  ORG_UPPER="R3"
  ORG_CAPITAL="R3"
  PEER_PORT="7051"

  ORG_LOWER="r4"
  ORG_UPPER="R4"
  ORG_CAPITAL="R4"
  PEER_PORT="9051"

  📁 生成的文件

  运行后会生成两个文件：
  - organizations/peerOrganizations/r3.example.com/connection-r3.json
  - organizations/peerOrganizations/r4.example.com/connection-r4.json

  🗺️ 文件内容示例

  {
    "name": "r3",
    "version": "1.0.0",
    "client": {
      "organization": "R3"
    },
    "organizations": {
      "R3": {
        "mspid": "R3MSP",
        "peers": ["peer0.r3.example.com"]
      }
    },
    "peers": {
      "peer0.r3.example.com": {
        "url": "grpcs://localhost:7051",
        "tlsCACerts": {
          "pem": "-----BEGIN CERTIFICATE-----..."
        }
      }
    }
  }

  ✅ 一句话总结

  第三步就是"给每个组织制作一张名片，告诉其他程序怎么找到它" -
  你可以复制现有的脚本直接运行，不需要复杂修改。
```

前三步总结:

```

⏺ 📋 前三步总结

  第1步：创建3个配置文件

  目的：告诉系统要创建哪些组织

  | 文件                              | 作用     | 创建的组织       |
  |---------------------------------|--------|-------------|
  | crypto-config-r3.yaml           | 定义R3组织 | R3对等组织      |
  | crypto-config-r4.yaml           | 定义R4组织 | R4对等组织      |
  | crypto-config-orderer-r3r4.yaml | 定义排序组织 | Orderer排序组织 |

  第2步：生成证书

  目的：给每个组织发"身份证"

  # 生成3套证书，共3个组织
  cryptogen generate ... (执行3次命令)

  生成目录：
  organizations/
  ├── peerOrganizations/r3.example.com/     # R3的证书
  ├── peerOrganizations/r4.example.com/     # R4的证书
  └── ordererOrganizations/r3r4.example.com/ # Orderer的证书

  第3步：连接配置文件

  目的：给每个组织制作"名片"，告诉其他程序怎么连接它

  实际操作：运行一个脚本生成两个JSON文件：
  - connection-r3.json（R3的名片）
  - connection-r4.json（R4的名片）

  🎯 前三步完成状态

  ✅ 3个组织已定义
  ✅ 3套证书已生成
  ✅ 2个连接配置已准备

  下一步：创建Docker容器配置文件，把这3个组织变成可以运行的容器。

```

4. 创建 docker 配置文件
   PS:TODO,解释一下这里到底做了什么事儿？我为什么没有看到 node 的创建

```
 创建文件：compose/compose-r3r4-test-net.yaml

  version: '3.7'

  volumes:
    orderer.r3r4.example.com:
    peer0.r3.example.com:
    peer0.r4.example.com:

  networks:
    r3r4:
      name: fabric_r3r4

  services:

    orderer.r3r4.example.com:
      container_name: orderer.r3r4.example.com
      image: hyperledger/fabric-orderer:latest
      environment:
        - FABRIC_LOGGING_SPEC=INFO
        - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
        - ORDERER_GENERAL_LISTENPORT=7050
        - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
        - ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp
        - ORDERER_GENERAL_TLS_ENABLED=true
        - ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key
        - ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt
        - ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
        - ORDERER_GENERAL_BOOTSTRAPMETHOD=none
        - ORDERER_CHANNELPARTICIPATION_ENABLED=true
      volumes:
        - ../organizations/ordererOrganizations/r3r4.example.com/orderers/orderer.r3r4.example.com/msp:/var/
  hyperledger/orderer/msp
        - ../organizations/ordererOrganizations/r3r4.example.com/orderers/orderer.r3r4.example.com/tls/:/var
  /hyperledger/orderer/tls
        - orderer.r3r4.example.com:/var/hyperledger/production/orderer
      ports:
        - 7050:7050
      networks:
        - r3r4

    peer0.r3.example.com:
      container_name: peer0.r3.example.com
      image: hyperledger/fabric-peer:latest
      environment:
        - FABRIC_CFG_PATH=/etc/hyperledger/peercfg
        - FABRIC_LOGGING_SPEC=INFO
        - CORE_PEER_TLS_ENABLED=true
        - CORE_PEER_ID=peer0.r3.example.com
        - CORE_PEER_ADDRESS=peer0.r3.example.com:7051
        - CORE_PEER_LISTENADDRESS=0.0.0.0:7051
        - CORE_PEER_LOCALMSPID=R3MSP
        - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/fabric/msp
      volumes:
        -
  ../organizations/peerOrganizations/r3.example.com/peers/peer0.r3.example.com:/etc/hyperledger/fabric
        - peer0.r3.example.com:/var/hyperledger/production
      ports:
        - 7051:7051
      networks:
        - r3r4

    peer0.r4.example.com:
      container_name: peer0.r4.example.com
      image: hyperledger/fabric-peer:latest
      environment:
        - FABRIC_CFG_PATH=/etc/hyperledger/peercfg
        - FABRIC_LOGGING_SPEC=INFO
        - CORE_PEER_TLS_ENABLED=true
        - CORE_PEER_ID=peer0.r4.example.com
        - CORE_PEER_ADDRESS=peer0.r4.example.com:9051
        - CORE_PEER_LISTENADDRESS=0.0.0.0:9051
        - CORE_PEER_LOCALMSPID=R4MSP
        - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/fabric/msp
      volumes:
        -
  ../organizations/peerOrganizations/r4.example.com/peers/peer0.r4.example.com:/etc/hyperledger/fabric
        - peer0.r4.example.com:/var/hyperledger/production
      ports:
        - 9051:9051
      networks:
        - r3r4

```

详细解析：

```
  第四步就是**"写一份说明书，告诉Docker如何创建和启动我们需要的容器"**

  📋 为什么要用Docker？

  想象你要开3家餐厅：
  - R3餐厅（需要厨师、服务员、收银台）
  - R4餐厅（需要厨师、服务员、收银台）
  - Orderer邮局（需要分拣员、邮递员）

  Docker就是**"标准化厨房"**，我们把每样东西都装进标准化的"集装箱"，这样不管在哪台机器上都能一模一样地运行。

  🔍 compose-r3r4-test-net.yaml文件详解

  1. 文件结构（就像餐厅的布局图）

  version: '3.7'          # Docker语法版本，不用管

  volumes:                # 数据存储位置
    orderer.r3r4.example.com:    # 邮局的数据存储
    peer0.r3.example.com:        # R3餐厅的数据存储
    peer0.r4.example.com:        # R4餐厅的数据存储

  networks:               # 网络配置，让3个餐厅可以互相打电话
    r3r4:
      name: fabric_r3r4

  2. 排序节点配置（邮局）

  services:
    orderer.r3r4.example.com:    # 邮局的名字
      container_name: orderer.r3r4.example.com  # Docker容器名字
      image: hyperledger/fabric-orderer:latest  # 使用官方镜像（就像用标准厨房设备）

      environment:              # 环境变量（就像给员工的制服和工牌）
        - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0    # 监听所有IP
        - ORDERER_GENERAL_LISTENPORT=7050          # 邮局门口：7050号
        - ORDERER_GENERAL_LOCALMSPID=OrdererMSP    # "我是Orderer组织的"

      volumes:                  # 文件挂载（把证书放进容器）
        - ../organizations/ordererOrganizations/r3r4.example.com/...:/var/hyperledger/orderer/msp
        # 左边：宿主机的证书文件
        # 右边：容器内看到的路径

      ports:                    # 端口映射（就像给每个窗口编号）
        - 7050:7050             # 宿主机的7050 → 容器的7050

      networks:
        - r3r4                 # 加入"r3r4"网络，可以和R3、R4通话

  3. 对等节点配置（R3餐厅）

    peer0.r3.example.com:       # R3餐厅的名字
      container_name: peer0.r3.example.com
      image: hyperledger/fabric-peer:latest

      environment:              # 餐厅员工的配置
        - CORE_PEER_ID=peer0.r3.example.com        # 我的ID
        - CORE_PEER_ADDRESS=peer0.r3.example.com:7051    # 我的地址
        - CORE_PEER_LOCALMSPID=R3MSP              # 我是R3组织的
        - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/fabric/msp  # 我的证书位置

      volumes:                  # 文件挂载
        - ../organizations/peerOrganizations/r3.example.com/...:/etc/hyperledger/fabric
        # 把R3的证书、密钥等文件映射到容器内

      ports:
        - 7051:7051             # 餐厅前门：7051号

      networks:
        - r3r4                  # 可以和邮局、R4餐厅通话

  🚀 第五步如何执行

  1. 启动命令（实际发生了什么）

  docker-compose -f compose/compose-r3r4-test-net.yaml up -d

  背后的过程：
  1. Docker读取yaml文件
  2. 下载镜像（hyperledger/fabric-orderer, hyperledger/fabric-peer）
  3. 创建3个容器（邮局+R3餐厅+R4餐厅）
  4. 把证书文件挂载到容器内
  5. 启动所有服务
  6. 设置网络连接，让它们可以互相通信

  2. 验证启动成功

  docker ps
  # 看到3个容器在运行：
  # - orderer.r3r4.example.com
  # - peer0.r3.example.com
  # - peer0.r4.example.com

  🎯 小白理解版

  第四步就是：
  1. 写一份清单：告诉Docker要开哪几家店（3个容器）
  2. 每家店的配置：地址、员工制服、设备摆放
  3. 共享文件：把证书文件放到每家店的指定位置
  4. 网络连接：让3家店可以互相打电话沟通


```

5. 启动网络

```

 docker-compose -f compose/compose-r3r4-test-net.yaml up -d

```

6. 创建通道

```
  如果需要创建通道：

  创建文件：configtx/configtx-r3r4.yaml
  Organizations:
    - &R3
      Name: R3MSP
      ID: R3MSP
      MSPDir: ../organizations/peerOrganizations/r3.example.com/msp

    - &R4
      Name: R4MSP
      ID: R4MSP
      MSPDir: ../organizations/peerOrganizations/r4.example.com/msp

  Application: &ApplicationDefaults
    Organizations:
      - *R3
      - *R4

  Profiles:
    TwoOrgsApplicationGenesis:
      <<: *ApplicationDefaults
      Orderer:
        <<: *OrdererDefaults
        Organizations:
          - *OrdererOrg
      Consortiums:
        SampleConsortium:
          Organizations:
            - *R3
            - *R4

  执行：
  # 生成创世区块
  configtxgen -profile TwoOrgsApplicationGenesis -outputBlock ./system-genesis-block/r3r4-genesis.block
  -channelID r3r4-channel
```

详细理解：

```
第六步详解：创建通道（小白版）

  🎯 一句话理解

  创建通道 = "给R3和R4建立一个私人群聊，让它们可以安全地转账交易"

  🤔 为什么需要通道？

  想象：
  - R3 = 微信用户A
  - R4 = 微信用户B
  - 排序节点 = 微信服务器

  通道 = "R3和R4的私聊窗口"，其他人看不到

  📋 具体在做什么？

  1. 创建configtx-r3r4.yaml文件

  这个文件就是**"群聊规则说明书"**：

  # 定义参与者
  Organizations:
    - &R3        # R3的身份证信息
      Name: R3MSP
      ID: R3MSP
      MSPDir: ../organizations/peerOrganizations/r3.example.com/msp  # R3的证书位置

    - &R4        # R4的身份证信息
      Name: R4MSP
      ID: R4MSP
      MSPDir: ../organizations/peerOrganizations/r4.example.com/msp  # R4的证书位置

  # 定义私聊窗口设置
  Profiles:
    TwoOrgsApplicationGenesis:    # 私聊配置名称
      Application:                # 应用层设置
        Organizations:
          - *R3                   # 允许R3加入
          - *R4                   # 允许R4加入
      Orderer:                    # 排序服务设置
        Organizations:
          - *OrdererOrg           # 指定排序节点

  2. 执行configtxgen命令

  configtxgen -profile TwoOrgsApplicationGenesis \
    -outputBlock ./system-genesis-block/r3r4-genesis.block \
    -channelID r3r4-channel

  这个命令做了什么？
  1. 读取规则说明书（configtx-r3r4.yaml）
  2. 验证身份：检查R3和R4的证书是否有效
  3. 创建创世区块：就像创建一个新的微信群，生成第一个"群公告"
  4. 保存文件：把"群公告"保存为r3r4-genesis.block文件

  🏗️ 生成的文件

  system-genesis-block/
  └── r3r4-genesis.block    # ← 这就是"群聊窗口的创建证书"

  📱 类比微信群

  | 微信群     | 区块链通道    |
  |---------|----------|
  | 创建群聊    | 生成创世区块   |
  | 群成员列表   | R3和R4的证书 |
  | 群公告     | 通道配置规则   |
  | 群主（管理群） | 排序节点     |

  📝 小白理解总结

  第六步就是：
  1. 写群规：规定只有R3和R4能加入
  2. 验证身份：检查R3和R4的身份证（证书）
  3. 建群：生成一个"群聊证书"（创世区块）
  4. 保存证书：把证书存起来，后续加入群聊需要它

  结果：R3和R4现在有专属私聊通道了，可以开始安全交易！


```

7. 验证

```
  ✅ 验证步骤

  # 检查容器状态
  docker ps

  # 检查证书是否生成
  ls organizations/peerOrganizations/r3.example.com/
  ls organizations/peerOrganizations/r4.example.com/
  ls organizations/ordererOrganizations/r3r4.example.com/
```
