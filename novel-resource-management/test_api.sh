#!/bin/bash

# API测试脚本
# 测试novel-resource-management的所有API接口

BASE_URL="http://localhost:8080/api/v1"
NOVEL_URL="$BASE_URL/novels"
USER_URL="$BASE_URL/users"
HEALTH_URL="http://localhost:8080/health"

echo "🚀 开始API测试..."
echo "=================="

# 1. 健康检查
echo "1️⃣  健康检查..."
health_response=$(curl -s "$HEALTH_URL")
health_code=$(curl -s -w "%{http_code}" "$HEALTH_URL" | tail -c 3)
echo "健康检查响应: $health_response"
echo "HTTP状态码: $health_code"

if [ "$health_code" -eq 200 ]; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败"
fi
echo ""

# 2. 获取所有小说
echo "2️⃣  获取所有小说..."
novels_response=$(curl -s -w "%{http_code}" "$NOVEL_URL")
echo "HTTP状态码: ${novels_response: -3}"
echo "响应内容: ${novels_response%???}"
echo ""

# 3. 创建新小说
echo "3️⃣  创建新小说..."
create_novel_data='{
    "ID": "test_novel_001",
    "author": "测试作者",
    "storyOutline": "这是一个测试小说的故事大纲",
    "subsections": "第一章,第二章,第三章",
    "characters": "主角,配角,反派",
    "items": "魔法剑,神秘护符",
    "totalScenes": "3"
}'

create_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$create_novel_data" \
    -w "%{http_code}" \
    "$NOVEL_URL")

echo "HTTP状态码: ${create_response: -3}"
echo "响应内容: ${create_response%???}"
echo ""

# 4. 获取单个小说
echo "4️⃣  获取单个小说..."
get_novel_response=$(curl -s -w "%{http_code}" "$NOVEL_URL/test_novel_001")
echo "HTTP状态码: ${get_novel_response: -3}"
echo "响应内容: ${get_novel_response%???}"
echo ""

# 5. 更新小说
echo "5️⃣  更新小说..."
update_novel_data='{
    "ID": "test_novel_001",
    "author": "更新的测试作者",
    "storyOutline": "这是更新后的故事大纲",
    "subsections": "第一章,第二章,第三章,第四章",
    "characters": "主角,配角,反派,新角色",
    "items": "魔法剑,神秘护符,新道具",
    "totalScenes": "4"
}'

update_response=$(curl -s -X PUT \
    -H "Content-Type: application/json" \
    -d "$update_novel_data" \
    -w "%{http_code}" \
    "$NOVEL_URL/test_novel_001")

echo "HTTP状态码: ${update_response: -3}"
echo "响应内容: ${update_response%???}"
echo ""

# 6. 获取所有用户积分
echo "6️⃣  获取所有用户积分..."
users_response=$(curl -s -w "%{http_code}" "$USER_URL")
echo "HTTP状态码: ${users_response: -3}"
echo "响应内容: ${users_response%???}"
echo ""

# 7. 创建用户积分
echo "7️⃣  创建用户积分..."
create_user_data='{
    "userId": "test_user_001",
    "credit": 100,
    "totalUsed": 0,
    "totalRecharge": 100
}'

create_user_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$create_user_data" \
    -w "%{http_code}" \
    "$USER_URL")

echo "HTTP状态码: ${create_user_response: -3}"
echo "响应内容: ${create_user_response%???}"
echo ""

# 8. 获取单个用户积分
echo "8️⃣  获取单个用户积分..."
get_user_response=$(curl -s -w "%{http_code}" "$USER_URL/test_user_001")
echo "HTTP状态码: ${get_user_response: -3}"
echo "响应内容: ${get_user_response%???}"
echo ""

# 9. 更新用户积分
echo "9️⃣  更新用户积分..."
update_user_data='{
    "userId": "test_user_001",
    "credit": 150,
    "totalUsed": 25,
    "totalRecharge": 175
}'

update_user_response=$(curl -s -X PUT \
    -H "Content-Type: application/json" \
    -d "$update_user_data" \
    -w "%{http_code}" \
    "$USER_URL/test_user_001")

echo "HTTP状态码: ${update_user_response: -3}"
echo "响应内容: ${update_user_response%???}"
echo ""

# 10. 再次获取所有小说和用户，验证更新
echo "🔟  再次获取所有小说验证更新..."
final_novels_response=$(curl -s -w "%{http_code}" "$NOVEL_URL")
echo "HTTP状态码: ${final_novels_response: -3}"
echo "响应内容: ${final_novels_response%???}"
echo ""

echo "🔟  再次获取所有用户积分验证更新..."
final_users_response=$(curl -s -w "%{http_code}" "$USER_URL")
echo "HTTP状态码: ${final_users_response: -3}"
echo "响应内容: ${final_users_response%???}"
echo ""

echo "=================="
echo "🏁 API测试完成"

# 清理测试数据（可选）
echo ""
echo "🧹 清理测试数据..."

# 删除测试小说
delete_novel_status=$(curl -s -X DELETE "$NOVEL_URL/test_novel_001" -w "%{http_code}" | tail -c 3)
if [ "$delete_novel_status" -eq 200 ]; then
    echo "✅ 已删除测试小说: test_novel_001"
else
    echo "❌ 删除测试小说失败: test_novel_001 (状态码: $delete_novel_status)"
fi

# 删除测试用户
delete_user_status=$(curl -s -X DELETE "$USER_URL/test_user_001" -w "%{http_code}" | tail -c 3)
if [ "$delete_user_status" -eq 200 ]; then
    echo "✅ 已删除测试用户: test_user_001"
else
    echo "❌ 删除测试用户失败: test_user_001 (状态码: $delete_user_status)"
fi

echo "🏁 测试数据清理完成"