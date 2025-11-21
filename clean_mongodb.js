// MongoDB 数据清理脚本
// 运行方法：在 MongoDB shell 中执行: load("clean_mongodb.js")

use novel;

print("🔍 开始检查数据库状态...");

// 1. 查看 novels 集合的所有索引
print("\n📋 novels 集合的当前索引:");
db.novels.getIndexes().forEach(function(index) {
    print("  - " + index.name + ": " + JSON.stringify(index.key));
});

// 2. 删除错误的索引 (如果存在)
try {
    var result = db.novels.dropIndex("novels_userId_novelId_key");
    print("✅ 成功删除错误的索引: novels_userId_novelId_key");
} catch (e) {
    if (e.code === 27) {
        print("ℹ️ 错误索引不存在，无需删除");
    } else {
        print("⚠️ 删除索引时出错: " + e.message);
    }
}

// 3. 查看 novels 集合的现有数据
print("\n📚 novels 集合中的现有记录:");
var novels = db.novels.find().toArray();
if (novels.length === 0) {
    print("  (空集合)");
} else {
    novels.forEach(function(novel) {
        print("  ID: " + novel._id + ", Author: " + novel.author + ", StoryOutline长度: " + (novel.storyOutline ? novel.storyOutline.length : 0));
    });
}

// 4. 创建正确的 storyOutline 索引
try {
    db.novels.createIndex({"storyOutline": 1}, {unique: true});
    print("✅ 成功创建 storyOutline 唯一索引");
} catch (e) {
    if (e.code === 85) {
        print("⚠️ storyOutline 索引已存在");
    } else {
        print("❌ 创建 storyOutline 索引失败: " + e.message);
    }
}

// 5. 检查 user_credits 集合
print("\n💰 user_credits 集合状态:");
var userCredits = db.user_credits.find().toArray();
if (userCredits.length === 0) {
    print("  (空集合)");
} else {
    userCredits.forEach(function(credit) {
        print("  UserID: " + credit.userId + ", Credit: " + credit.credit);
    });
}

print("\n🎉 数据库清理完成！");