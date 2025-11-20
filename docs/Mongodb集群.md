MongoDB 真正进入“生产集群”阶段时，和单机随便玩完全是两回事。  
下面按 2025 年真实生产场景，把你必须做的操作一次性讲透（单机 → 副本集 → 分片集群逐级进阶）。

| 阶段               | 必须做的安全&集群操作                             | 一句话目的                              | 推荐最小规模（生产） |
|--------------------|--------------------------------------------------|-----------------------------------------|----------------------|
| 单机（你现在）     | 1. 创建管理员用户<br>2. 开启 --auth                | 不被黑客一秒删库                        | 1 台（仅开发）       |
| 副本集 Replica Set（生产最小单位） | 1. 每台都创建相同管理员<br>2. 用 keyfile 内部认证<br>3. 初始化 rs.initiate() | 高可用 + 自动故障转移                   | 最少 3 台（2 数据 + 1 投票） |
| 分片集群 Sharded Cluster | 1. configsvr（配置服务器副本集）<br>2. mongos（路由）<br>3. 每个 shard 都是副本集<br>4. 所有节点 keyfile 统一<br>5. enableSharding + shardCollection | 海量数据 + 水平扩展                     | 最少 11 台（简化也可 7 台） |

### 2025 年生产最常见部署方式（99% 公司都这么干）

| 组件             | 台数（最小生产） | 端口    | 说明                                     |
|------------------|------------------|---------|------------------------------------------|
| mongos（路由）   | 2~3 台           | 27017   | 应用连接这里                             |
| configsvr（配置服务器） | 3 台（副本集）   | 27019   | 存元数据，必须是副本集                   |
| shard1（分片1）  | 3 台（副本集）   | 27018   | 真实存数据                               |
| shard2（分片2）  | 3 台（副本集）   | 27018   | 真实存数据                               |
| 总计             | 最少 11 台       |         | 真正能扛住生产流量 + 自动故障转移        |

### 关键步骤：从单机 → 副本集（你下一步就要做的事）

1. 准备 3 台机器（Mac + 云服务器 / Docker / 本地虚拟机都行）
2. 每台都安装 MongoDB + 创建相同管理员（单机那套）
3. 生成一个 keyfile（内部成员互相认证用，超级重要！）
   ```bash
   openssl rand -base64 756 > mongodb-keyfile
   chmod 400 mongodb-keyfile
   # 复制到每台服务器的 /data/mongodb-keyfile
   ```
4. 每台 mongod 启动都加这三行（配置文件或命令行）
   ```yaml
   security:
     authorization: enabled          # 外部认证
     keyFile: /data/mongodb-keyfile  # 内部成员认证（必须！）
   replication:
     replSetName: rs0                # 副本集名字统一
   ```
5. 在任意一台初始化副本集
   ```bash
   mongosh --host localhost --port 27017 -u admin -p 你的密码
   rs.initiate({
     _id: "rs0",
     members: [
       { _id: 0, host: "ip1:27017" },
       { _id: 1, host: "ip2:27017" },
       { _id: 2, host: "ip3:27017" }
     ]
   })
   ```
6. 看状态（看到 PRIMARY 和 SECONDARY 就成功了）
   ```bash
   rs.status()
   ```

### 再进阶到分片集群（大数据量时再做）

1. 先启动 3 台 configsvr（也是副本集，名字叫 configRepl）
2. 启动每组 shard（上面建好的副本集）
3. 启动 mongos，连接 configsvr
   ```bash
   mongos --configdb configRepl/c1:27019,c2:27019,c3:27019 --port 27017
   ```
4. 通过 mongos 添加 shard
   ```bash
   sh.addShard("rs0/ip1:27017,ip2:27017,ip3:27017")
   ```
5. 开启数据库分片 + 选片键
   ```bash
   sh.enableSharding("mydb")
   sh.shardCollection("mydb.users", { "user_id": 1 })
   ```

### 一句话总结你现在的行动路线

| 当前阶段     | 下一件必须做的事                             | 命令关键词            |
|--------------|----------------------------------------------|-----------------------|
| 单机玩着     | 立刻创建管理员 + 开启 auth                   | createUser + authorization: enabled |
| 想高可用     | 建 3 台副本集 + keyfile 内部认证             | keyfile + rs.initiate |
| 数据要上亿   | 再加 configsvr + mongos + 多个 shard         | mongos + sh.addShard  |

生产一句话铁律：  
**没有 keyfile + 没有副本集的 MongoDB = 随时可能死无葬身之地**

现在就去把 keyfile 建了，然后搭个 3 节点副本集吧！  
需要我给你一键 Docker Compose 完整搭建脚本（3节点副本集 + keyfile + 管理员）吗？一句话就发你，10 秒启动生产级集群。