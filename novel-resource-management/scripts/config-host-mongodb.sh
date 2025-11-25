#!/bin/sh

# 宿主机MongoDB副本集自动配置脚本
# 在Alpine容器中执行，配置宿主机的MongoDB副本集

echo "🔧 开始配置宿主机MongoDB副本集..."

# MongoDB连接配置
MONGO_ADMIN_USER=${MONGO_ADMIN_USER:-"admin"}
MONGO_ADMIN_PASS=${MONGO_ADMIN_PASS:-"715705@Qc123"}
MONGO_HOST=${MONGO_HOST:-"127.0.0.1"}
MONGO_PORT=${MONGO_PORT:-"27017"}

echo "📋 配置信息:"
echo "   用户: $MONGO_ADMIN_USER"
echo "   主机: $MONGO_HOST"
echo "   端口: $MONGO_PORT"

# 安装必要的工具
apk add --no-cache mongodb-tools curl iproute2

# 获取宿主机在局域网中的真实IP地址
echo "🔍 获取宿主机局域网IP地址..."
HOST_IP=""

# 方法1: 通过ifconfig获取宿主机真实局域网IP（跳过Docker网络）
if command -v ifconfig >/dev/null 2>&1; then
    for interface in en0 eth0; do
        INTERFACE_IP=$(ifconfig "$interface" 2>/dev/null | grep 'inet ' | grep -v '127.0.0.1' | grep -v '192.168.65' | grep -v '172.17' | awk '{print $2}' | head -1)
        # 清理addr:前缀（如果存在）
        INTERFACE_IP=$(echo "$INTERFACE_IP" | sed 's/addr://' | sed 's/inet://')
        if [ -n "$INTERFACE_IP" ] && [ "$INTERFACE_IP" != "127.0.0.1" ]; then
            # 优先选择172.16网段（你的局域网）
            if [[ "$INTERFACE_IP" == 172.16.* ]]; then
                HOST_IP="$INTERFACE_IP"
                echo "✅ 通过网络接口 $interface 获取到局域网IP: $HOST_IP"
                break
            fi
        fi
    done
fi

# 如果没找到172.16网段，尝试其他局域网段
if [ -z "$HOST_IP" ] && command -v ifconfig >/dev/null 2>&1; then
    for interface in en0 eth0; do
        INTERFACE_IP=$(ifconfig "$interface" 2>/dev/null | grep 'inet ' | grep -v '127.0.0.1' | grep -v '192.168.65' | grep -v '172.17' | awk '{print $2}' | head -1)
        INTERFACE_IP=$(echo "$INTERFACE_IP" | sed 's/addr://' | sed 's/inet://')
        if [ -n "$INTERFACE_IP" ] && [ "$INTERFACE_IP" != "127.0.0.1" ]; then
            HOST_IP="$INTERFACE_IP"
            echo "✅ 通过网络接口 $interface 获取到其他局域网IP: $HOST_IP"
            break
        fi
    done
fi

# 方法2: 通过ip命令获取（Linux兼容）
if [ -z "$HOST_IP" ] && command -v ip >/dev/null 2>&1; then
    for interface in en0 eth0; do
        INTERFACE_IP=$(ip addr show "$interface" 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1 | head -1)
        if [ -n "$INTERFACE_IP" ] && [ "$INTERFACE_IP" != "127.0.0.1" ]; then
            HOST_IP="$INTERFACE_IP"
            echo "✅ 通过网络接口 $interface 获取到IP: $HOST_IP"
            break
        fi
    done
fi

# 方法3: 扫描172.16网段查找宿主机IP
if [ -z "$HOST_IP" ]; then
    echo "🔍 扫描172.16网段查找宿主机IP..."
    for ip in 172.16.181.{100..110}; do
        if ping -c 1 -W 1 "$ip" >/dev/null 2>&1; then
            # 尝试连接该IP的MongoDB
            if mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$ip:$MONGO_PORT/admin?authSource=admin" --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
                HOST_IP="$ip"
                echo "✅ 找到可用的MongoDB宿主机IP: $HOST_IP"
                break
            fi
        fi
    done
fi

# 方法4: 备用方案 - 尝试常见局域网IP
if [ -z "$HOST_IP" ]; then
    echo "⚠️  无法自动获取，尝试常见IP地址..."
    for test_ip in 172.16.181.101 192.168.1.100 192.168.0.100 10.0.2.15; do
        if ping -c 1 -W 1 "$test_ip" >/dev/null 2>&1; then
            if mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$test_ip:$MONGO_PORT/admin?authSource=admin" --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
                HOST_IP="$test_ip"
                echo "✅ 找到可用的MongoDB宿主机IP: $HOST_IP"
                break
            fi
        fi
    done
fi

# 最终检查
if [ -z "$HOST_IP" ]; then
    echo "❌ 无法获取宿主机IP地址"
    exit 1
fi

echo "📍 最终使用宿主机IP: $HOST_IP"

# 在host网络模式下，MongoDB连接使用127.0.0.1
MONGO_HOST="127.0.0.1"
echo "🔗 MongoDB连接地址: $MONGO_HOST:$MONGO_PORT"

# 等待MongoDB服务可用
echo "⏳ 等待MongoDB服务启动..."
max_attempts=20
attempt=1

while [ $attempt -le $max_attempts ]; do
    if mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        echo "✅ MongoDB服务已就绪"
        break
    fi

    echo "⏳ 第 $attempt 次尝试连接MongoDB..."
    sleep 2
    attempt=$((attempt + 1))
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ MongoDB服务连接失败"
    exit 1
fi

# 检查副本集状态
echo "🔍 检查MongoDB副本集状态..."
REPLICA_STATUS=$(mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "try { rs.status().ok } catch(e) { 0 }" --quiet)

if [ "$REPLICA_STATUS" = "1" ]; then
    echo "✅ 副本集已配置，检查是否需要更新IP..."

    # 检查当前副本集配置
    CURRENT_MEMBER=$(mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "rs.conf().members[0].host" --quiet)
    echo "📊 当前副本集配置: $CURRENT_MEMBER"

    # 如果当前配置不是宿主机IP，则更新
    if [ "$CURRENT_MEMBER" != "$HOST_IP:$MONGO_PORT" ]; then
        echo "🔧 更新副本集配置到宿主机IP..."
        mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "
            rs.reconfig({
                _id: 'rs0',
                members: [
                    { _id: 0, host: '$HOST_IP:$MONGO_PORT' }
                ]
            }, { force: true });
            print('✅ 副本集配置已更新到: $HOST_IP:$MONGO_PORT');
        "
    else
        echo "✅ 副本集配置已正确"
    fi
else
    echo "🔧 初始化副本集..."
    mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "
        try {
            rs.initiate({
                _id: 'rs0',
                members: [
                    { _id: 0, host: '$HOST_IP:$MONGO_PORT' }
                ]
            });
            print('✅ 副本集初始化成功');
        } catch(e) {
            print('⚠️  初始化警告: ' + e.message);
        }
    "

    # 等待副本集选举完成
    echo "⏳ 等待副本集选举完成..."
    sleep 10
fi

# 验证副本集状态
echo "🔍 验证副本集状态..."
sleep 5

mongosh "mongodb://$MONGO_ADMIN_USER:$MONGO_ADMIN_PASS@$MONGO_HOST:$MONGO_PORT/admin?authSource=admin" --eval "
    try {
        var status = rs.status();
        print('🎉 副本集状态:');
        print('   副本集名称: ' + status.set);
        print('   节点数量: ' + status.members.length);
        status.members.forEach(function(member) {
            print('   - ' + member.name + ': ' + member.healthStr + ' (' + member.stateStr + ')');
        });
        print('✅ MongoDB副本集配置验证成功');
    } catch(e) {
        print('❌ 验证失败: ' + e.message);
        exit(1);
    }
"

echo "🎉 宿主机MongoDB副本集配置完成！"