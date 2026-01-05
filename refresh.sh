#!/bin/bash

###############################################################################
# MongoDB 副本集刷新 - 便捷脚本
###############################################################################

# 从 .env 文件读取密码
if [ -f /Users/zack/Desktop/ProChainRM/novel-resource-management/.env ]; then
    export $(grep '^MONGO_PASS=' /Users/zack/Desktop/ProChainRM/novel-resource-management/.env | xargs)
fi

# 运行 Go 脚本
cd /Users/zack/Desktop/ProChainRM
go run refresh-mongodb.go
