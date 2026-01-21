//go:build test

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"novel-resource-management/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("=== MongoDB 多个查询示例 ===")

	// 获取数据库实例
	mongoInstance := database.GetMongoInstance()
	if !mongoInstance.IsConnected() {
		log.Fatal("数据库未连接")
	}

	// 准备一些测试数据
	setupTestData()

	// 运行各种查询示例
	runQueryExamples()

	fmt.Println("\n=== 查询示例完成 ===")
}

// 设置测试数据
func setupTestData() {
	fmt.Println("\n📦 准备测试数据...")

	collections := []string{
		"user_credits",
		"novels",
	}

	// 清理旧数据
	for _, collectionName := range collections {
		collection := database.GetMongoInstance().GetCollection(collectionName)
		//这个Drop就是删表了
		collection.Drop(context.Background())
	}

	// 插入测试数据
	insertTestUserCredits()
	insertTestNovels()

	fmt.Println("✅ 测试数据准备完成")
}

// 插入测试用户积分数据
func insertTestUserCredits() {
	collection := database.GetMongoInstance().GetCollection("user_credits")
	//时间格式--- ：：：
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	userCredits := []interface{}{
		database.UserCredit{
			UserID:        "user_001",
			Credit:        100,
			TotalUsed:     20,
			TotalRecharge: 100,
			CreatedAt:     currentTime,
			UpdatedAt:     currentTime,
		},
		database.UserCredit{
			UserID:        "user_002",
			Credit:        250,
			TotalUsed:     50,
			TotalRecharge: 250,
			CreatedAt:     currentTime,
			UpdatedAt:     currentTime,
		},
		database.UserCredit{
			UserID:        "user_003",
			Credit:        75,
			TotalUsed:     25,
			TotalRecharge: 100,
			CreatedAt:     currentTime,
			UpdatedAt:     currentTime,
		},
		database.UserCredit{
			UserID:        "user_004",
			Credit:        500,
			TotalUsed:     100,
			TotalRecharge: 500,
			CreatedAt:     currentTime,
			UpdatedAt:     currentTime,
		},
	}

	_, err := collection.InsertMany(context.Background(), userCredits)
	if err != nil {
		log.Printf("插入用户积分数据失败: %v", err)
	}
}

// 插入测试小说数据
func insertTestNovels() {
	collection := database.GetMongoInstance().GetCollection("novels")
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	novels := []interface{}{
		database.Novel{
			Author:       "张三",
			StoryOutline: "这是一个玄幻小说的故事大纲",
			Subsections:  "第一章,第二章,第三章,第四章",
			Characters:   "主角A,配角B,反派C",
			Items:        "魔法剑,神秘药水",
			TotalScenes:  "20",
			CreatedAt:    "2024-01-15 10:00:00",
			UpdatedAt:    currentTime,
		},
		database.Novel{
			Author:       "李四",
			StoryOutline: "都市爱情小说",
			Subsections:  "序章,第一章,第二章",
			Characters:   "男主角,女主角",
			Items:        "玫瑰花,戒指",
			TotalScenes:  "15",
			CreatedAt:    "2024-02-20 14:30:00",
			UpdatedAt:    currentTime,
		},
		database.Novel{
			Author:       "张三",
			StoryOutline: "科幻冒险故事",
			Subsections:  "开端,发展,高潮,结局",
			Characters:   "太空人,外星人",
			Items:        "宇宙飞船,激光枪",
			TotalScenes:  "25",
			CreatedAt:    "2024-03-10 09:15:00",
			UpdatedAt:    currentTime,
		},
		database.Novel{
			Author:       "王五",
			StoryOutline: "悬疑推理小说",
			Subsections:  "案件发生,调查过程,真相揭露",
			Characters:   "侦探,嫌疑人,证人",
			Items:        "放大镜,证据袋",
			TotalScenes:  "18",
			CreatedAt:    "2024-01-25 16:45:00",
			UpdatedAt:    currentTime,
		},
	}

	_, err := collection.InsertMany(context.Background(), novels)
	if err != nil {
		log.Printf("插入小说数据失败: %v", err)
	}
}


// 运行查询示例
func runQueryExamples() {
	// 1. 查询所有用户积分
	queryAllUserCredits()

	// 2. 条件查询 - 积分大于100的用户
	queryUsersWithHighCredit()

	// 3. 范围查询 - 积分在50-200之间的用户
	queryUsersWithCreditRange()

	// 4. 复杂条件查询
	queryWithComplexConditions()

	// 5. 查询指定作者的小说
	queryNovelsByAuthor()

	// 6. 正则表达式查询
	queryWithRegex()

	// 7. 分页查询
	queryWithPagination()

	// 8. 排序查询
	queryWithSort()

	// 9. 只查询特定字段
	queryWithProjection()
}

// 1. 查询所有用户积分
func queryAllUserCredits() {
	fmt.Println("\n🔍 1. 查询所有用户积分")

	// 📖 小白解释：获取MongoDB数据库连接，然后拿到"user_credits"这个表（集合）
	// 就像拿到一个Excel文件，然后打开名为"user_credits"的工作表
	collection := database.GetMongoInstance().GetCollection("user_credits")

	// 📖 小白解释：在数据库中查找所有数据
	// context.Background() 表示这是一个独立的操作，没有超时限制
	// bson.M{} 是一个空的查询条件，相当于SQL中的"SELECT * FROM"，即查找所有记录
	// cursor 就像一个指向查询结果的指针，需要遍历它才能看到具体数据
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}

	// 📖 小白解释：defer就像一个"事后清理"的承诺
	// 无论函数是正常结束还是因为错误提前退出，这行代码都会在最后执行
	// 关闭cursor可以释放数据库资源，防止内存泄漏
	defer cursor.Close(context.Background())

	// 📖 小白解释：创建一个空的UserCredit数组，用来存放从数据库查出来的所有用户数据
	// 就像准备一个空篮子，等下要把超市里查到的所有商品都放进去
	var userCredits []database.UserCredit

	// 📖 小白解释：把cursor（查询结果）中的所有数据一次性全部读取到userCredits数组中
	// &userCredits 表示把这个数组的内存地址传给All方法，让它知道数据要存到哪里
	// 就像告诉收银员："请把所有商品都装到这个篮子里"
	err = cursor.All(context.Background(), &userCredits)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 📖 小白解释：打印查找到的用户总数
	// len(userCredits) 就是数组userCredits中元素的个数
	fmt.Printf("✅ 找到 %d 个用户:\n", len(userCredits))

	// 📖 小白解释：遍历所有用户数据并打印每个用户的信息
	// for _, user := range userCredentials 的意思是：
	//   range userCredits：逐个取出userCredits数组中的用户数据
	//   user：当前取出的这个用户数据
	//   _：表示我们不关心索引（第几个用户），只关心用户数据本身
	for _, user := range userCredits {
		fmt.Printf("   👤 %s: %d积分 (已用:%d, 充值:%d)\n",
			user.UserID, user.Credit, user.TotalUsed, user.TotalRecharge)
	}
}

// 2. 条件查询 - 积分大于100的用户
func queryUsersWithHighCredit() {
	fmt.Println("\n🔍 2. 查询积分大于100的用户")

	// 📖 小白解释：获取数据库连接，拿到用户积分表
	collection := database.GetMongoInstance().GetCollection("user_credits")

	// 📖 小白解释：设置查询条件，只查找积分大于100的用户
	// bson.M{"credit": bson.M{"$gt": 100}} 的含义：
	//   - 外层的 bson.M{"credit": ...} 表示要查询credit字段
	//   - 内层的 bson.M{"$gt": 100} 表示大于100
	//   - "$gt" 是MongoDB中的"大于"操作符（Greater Than）
	// 相当于SQL中的：WHERE credit > 100
	filter := bson.M{"credit": bson.M{"$gt": 100}}

	// 📖 小白解释：使用设置好的条件查询数据库
	// 只会返回积分大于100的用户记录
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	// 📖 小白解释：创建数组来存放查询结果
	var userCredits []database.UserCredit

	// 📖 小白解释：将查询结果读取到数组中
	err = cursor.All(context.Background(), &userCredits)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 📖 小白解释：打印符合条件用户的总数
	fmt.Printf("✅ 找到 %d 个积分大于100的用户:\n", len(userCredits))

	// 📖 小白解释：遍历所有符合条件的用户，显示他们的积分
	for _, user := range userCredits {
		fmt.Printf("   💰 %s: %d积分\n", user.UserID, user.Credit)
	}
}

// 3. 范围查询 - 积分在50-200之间的用户
func queryUsersWithCreditRange() {
	fmt.Println("\n🔍 3. 查询积分在50-200之间的用户")

	// 📖 小白解释：获取数据库连接，拿到用户积分表
	collection := database.GetMongoInstance().GetCollection("user_credits")

	// 📖 小白解释：设置范围查询条件，查找积分在50到200之间的用户
	// bson.M 的结构解释：
	//   - "credit": bson.M{...} 表示要查询credit字段
	//   - "$gte": 50 表示大于等于50（Greater Than or Equal）
	//   - "$lte": 200 表示小于等于200（Less Than or Equal）
	// 相当于SQL中的：WHERE credit >= 50 AND credit <= 200
	// 或者更简洁的：WHERE credit BETWEEN 50 AND 200
	filter := bson.M{
		"credit": bson.M{
			"$gte": 50,
			"$lte": 200,
		},
	}

	// 📖 小白解释：使用范围条件查询数据库
	// 只会返回积分在50-200之间的用户记录
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	// 📖 小白解释：创建数组来存放查询结果
	var userCredits []database.UserCredit

	// 📖 小白解释：将查询结果读取到数组中
	err = cursor.All(context.Background(), &userCredits)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 📖 小白解释：打印符合条件的用户总数
	fmt.Printf("✅ 找到 %d 个积分在50-200之间的用户:\n", len(userCredits))

	// 📖 小白解释：遍历所有符合条件的用户，显示他们的积分
	// 使用📊表情符号表示这是一个统计/数据分析的结果
	for _, user := range userCredits {
		fmt.Printf("   📊 %s: %d积分\n", user.UserID, user.Credit)
	}
}

// 4. 复杂条件查询
func queryWithComplexConditions() {
	fmt.Println("\n🔍 4. 复杂条件查询 - 张三的小说且场景数大于15")

	collection := database.GetMongoInstance().GetCollection("novels")

	filter := bson.M{
		"author": "张三",
		"totalScenes": bson.M{"$gt": "15"},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	var novels []database.Novel
	err = cursor.All(context.Background(), &novels)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 找到 %d 个符合条件的小说:\n", len(novels))
	for _, novel := range novels {
		fmt.Printf("   📚 《%s》: %s场景\n", novel.StoryOutline, novel.TotalScenes)
	}
}

// 5. 查询指定作者的小说
func queryNovelsByAuthor() {
	fmt.Println("\n🔍 5. 查询张三的所有小说")

	collection := database.GetMongoInstance().GetCollection("novels")

	filter := bson.M{"author": "张三"}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	var novels []database.Novel
	err = cursor.All(context.Background(), &novels)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 张三写了 %d 本小说:\n", len(novels))
	for _, novel := range novels {
		fmt.Printf("   📖 %s (%s场景)\n", novel.StoryOutline, novel.TotalScenes)
	}
}

// 6. 正则表达式查询
func queryWithRegex() {
	fmt.Println("\n🔍 6. 正则表达式查询 - 作者名包含'张'或'李'")

	collection := database.GetMongoInstance().GetCollection("novels")

	filter := bson.M{
		"author": bson.M{"$regex": "张|李"},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	var novels []database.Novel
	err = cursor.All(context.Background(), &novels)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 找到 %d 个作者名包含'张'或'李'的小说:\n", len(novels))
	for _, novel := range novels {
		fmt.Printf("   ✍️  %s: %s\n", novel.Author, novel.StoryOutline)
	}
}

// 7. 分页查询
func queryWithPagination() {
	fmt.Println("\n🔍 7. 分页查询 - 小说列表(第1页，每页2条)")

	collection := database.GetMongoInstance().GetCollection("novels")

	page := 1
	limit := 2
	skip := (page - 1) * limit

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"createdAt", -1}})

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	var novels []database.Novel
	err = cursor.All(context.Background(), &novels)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 第%d页，每页%d条，共%d条记录:\n", page, limit, len(novels))
	for i, novel := range novels {
		fmt.Printf("   %d. %s - %s\n", i+1, novel.Author, novel.StoryOutline)
	}
}

// 8. 排序查询
func queryWithSort() {
	fmt.Println("\n🔍 8. 排序查询 - 按积分降序排列用户")

	collection := database.GetMongoInstance().GetCollection("user_credits")

	opts := options.Find().SetSort(bson.D{{"credit", -1}})

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	var userCredits []database.UserCredit
	err = cursor.All(context.Background(), &userCredits)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 用户积分排名:\n")
	for i, user := range userCredits {
		fmt.Printf("   %d. %s: %d积分\n", i+1, user.UserID, user.Credit)
	}
}

// 9. 只查询特定字段
func queryWithProjection() {
	fmt.Println("\n🔍 9. 只查询特定字段 - 只获取小说作者和大纲")

	collection := database.GetMongoInstance().GetCollection("novels")

	projection := bson.M{
		"author": 1,
		"storyOutline": 1,
		"_id": 0, // 不返回_id字段
	}

	opts := options.Find().SetProjection(projection)

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	type NovelSummary struct {
		Author       string `bson:"author"`
		StoryOutline string `bson:"storyOutline"`
	}

	var summaries []NovelSummary
	err = cursor.All(context.Background(), &summaries)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 小说概要(只显示作者和大纲):\n")
	for _, summary := range summaries {
		fmt.Printf("   📝 %s: %s\n", summary.Author, summary.StoryOutline)
	}
}

