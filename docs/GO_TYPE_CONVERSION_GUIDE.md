# Go 语言类型转换详解：从 string 到各种类型

本文档详细解释 Go 语言中类型转换的各种方式，特别关注 `string` 与其他类型的转换，以及为什么 `string` 和 `[]byte` 可以直接转换而其他类型不能。

## 🔍 特殊转换：string ↔ []byte

### 1. 为什么可以直接转换？

```go
// 直接转换，语法简单
str := "Hello"
bytes := []byte(str)     // string -> []byte
str2 := string(bytes)    // []byte -> string

// 语法糖形式
str := "Hello"
bytes := []byte(str)     // 实际上是底层字节数组的拷贝
```

**根本原因：**
- `string` 在 Go 内部本质上是只读的字节数组
- `[]byte` 是可读写的字节数组
- Go 语言为这两种类型提供了内置的转换语法

### 2. 转换发生了什么？

```go
package main

import (
	"fmt"
)

func main() {
	str := "Hello"

	// 转换为字节数组
	bytes := []byte(str)

	// 内存地址不同 - 说明是拷贝，不是引用
	fmt.Printf("string地址: %p\n", &str)
	fmt.Printf("[]byte地址: %p\n", &bytes)

	// 修改字节数组不会影响原字符串
	bytes[0] = 'h'  // 改为小写h
	fmt.Println("原字符串:", str)    // "Hello" - 不变
	fmt.Println("字节数组:", string(bytes)) // "hello"
}
```

### 3. 项目中的实际例子

来自 `smartcontract.go:493` 的代码：

```go
// JSON 解析 - 常见用法
jsonData := "{\"novels\": [{\"title\": \"小说1\"}]}"

// 转换步骤：
jsonData    // string 类型
[]byte(jsonData)  // 转换为 []byte，因为 json.Unmarshal 需要字节数组

// json.Unmarshal 的函数签名
func Unmarshal(data []byte, v interface{}) error
```

## 🔧 String 转换为其他类型的处理方式

### 1. 转换为数值类型（int, float64等）

```go
import (
	"fmt"
	"strconv"
)

func numericConversions() {
	// String -> Int
	str := "123"

	// ❌ 错误：不能直接转换
	// num := int(str)  // 编译错误：cannot convert str (type string) to type int

	// ✅ 正确：使用 strconv 包
	num, err := strconv.Atoi(str)  // Atoi = "ASCII to Integer"
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	fmt.Printf("数字: %d, 类型: %T\n", num, num)  // 123, int

	// 更灵活的方式（可以指定进制）
	num64, err := strconv.ParseInt(str, 10, 64)  // 10进制，64位
	fmt.Printf("int64: %d\n", num64)

	// String -> Float
	floatStr := "3.14"
	f, err := strconv.ParseFloat(floatStr, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	fmt.Printf("浮点数: %.2f\n", f)  // 3.14

	// String -> Uint (无符号整数)
	uintStr := "42"
	uintVal, err := strconv.ParseUint(uintStr, 10, 32)  // 10进制，32位
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	fmt.Printf("uint32: %d\n", uintVal)
}
```

### 2. 转换为布尔值

```go
func booleanConversion() {
	// String -> Bool
	str := "true"

	// ParseBool 支持多种格式：
	// "1", "t", "T", "true", "True", "TRUE" -> true
	// "0", "f", "F", "false", "False", "FALSE" -> false
	b, err := strconv.ParseBool(str)
	if err != nil {
		fmt.Println("转换失败:", err)
	}
	fmt.Printf("布尔值: %t\n", b)  // true

	// 测试不同格式
	testCases := []string{"true", "false", "1", "0", "T", "F", "yes", "no"}
	for _, tc := range testCases {
		if b, err := strconv.ParseBool(tc); err == nil {
			fmt.Printf("%s -> %t\n", tc, b)
		} else {
			fmt.Printf("%s -> 错误: %v\n", tc, err)
		}
	}
}
```

### 3. 转换为结构体（最常见）

```go
import (
	"encoding/json"
	"fmt"
)

type Novel struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
}

func structConversion() {
	// JSON String -> Struct
	jsonStr := `{
		"title": "Go语言编程",
		"author": "张三",
		"pages": 300
	}`

	var novel Novel

	// 需要 string -> []byte -> struct 的转换步骤
	err := json.Unmarshal([]byte(jsonStr), &novel)
	if err != nil {
		fmt.Println("JSON解析失败:", err)
		return
	}

	fmt.Printf("小说: %+v\n", novel)
	// 输出: 小说: {Title:Go语言编程 Author:张三 Pages:300}
}

// 反向转换：Struct -> JSON String
func structToString() {
	novel := Novel{
		Title:  "区块链技术",
		Author: "李四",
		Pages:  250,
	}

	// Struct -> []byte -> String
	jsonBytes, err := json.Marshal(novel)
	if err != nil {
		fmt.Println("JSON序列化失败:", err)
		return
	}

	jsonStr := string(jsonBytes)  // []byte -> string
	fmt.Println("JSON字符串:", jsonStr)
}
```

## 🏢 项目中的实际使用案例

### 1. 用户服务中的 strconv 使用

来自 `user_service.go`：

```go
// Int -> String（因为区块链接口需要字符串参数）
credit := 100
totalUsed := 50
totalRecharge := 150

// 使用 strconv.Itoa 将整数转为字符串
_, err := us.contract.SubmitTransaction("CreateUserCredit",
    userId,
    strconv.Itoa(credit),         // 100 -> "100"
    strconv.Itoa(totalUsed),      // 50 -> "50"
    strconv.Itoa(totalRecharge))  // 150 -> "150"

// String -> Int（从配置文件读取）
maxPool := "10"  // 从.env文件读取
if size, err := strconv.ParseUint(maxPool, 10, 64); err == nil {
    config.MaxPoolSize = size  // "10" -> 10 (uint64)
}
```

### 2. 数据库配置中的使用

来自 `mongodb.go`：

```go
// 从环境变量读取连接池大小配置
maxPool := os.Getenv("MONGO_MAX_POOL_SIZE")
if maxPool != "" {
    // String -> Uint64
    if size, err := strconv.ParseUint(maxPool, 10, 64); err == nil {
        config.MaxPoolSize = size
    } else {
        log.Printf("无效的 MONGO_MAX_POOL_SIZE: %v", err)
    }
}

minPool := os.Getenv("MONGO_MIN_POOL_SIZE")
if minPool != "" {
    if size, err := strconv.ParseUint(minPool, 10, 64); err == nil {
        config.MinPoolSize = size
    }
}
```

### 3. 订单生成中的使用

来自 `sync-map-examples.md`：

```go
// 生成订单号
orderNum := 12345
orderID := "ORD" + strconv.Itoa(orderNum)  // "ORD12345"

// 生成房间号
roomID := "ROOM_" + strconv.Itoa(rand.Intn(10000))  // "ROOM_1234"
```

## 📋 转换方式总结表

| 转换类型 | 直接转换 | 需要的包 | 示例 | 错误处理 |
|----------|----------|----------|------|----------|
| `string` ↔ `[]byte` | ✅ 支持 | 无 | `[]byte(str)` / `string(bytes)` | 无错误（但可能有数据丢失） |
| `string` → `int` | ❌ 不支持 | `strconv` | `strconv.Atoi("123")` | 返回 `(int, error)` |
| `int` → `string` | ❌ 不支持 | `strconv` | `strconv.Itoa(123)` | 无错误 |
| `string` → `int64` | ❌ 不支持 | `strconv` | `strconv.ParseInt("123", 10, 64)` | 返回 `(int64, error)` |
| `string` → `uint64` | ❌ 不支持 | `strconv` | `strconv.ParseUint("123", 10, 64)` | 返回 `(uint64, error)` |
| `string` → `float64` | ❌ 不支持 | `strconv` | `strconv.ParseFloat("3.14", 64)` | 返回 `(float64, error)` |
| `string` → `bool` | ❌ 不支持 | `strconv` | `strconv.ParseBool("true")` | 返回 `(bool, error)` |
| `string` ↔ `struct` | ❌ 不支持 | `encoding/json` | `json.Unmarshal([]byte(jsonStr), &obj)` | 返回 `error` |

## 🎯 为什么设计成这样？

### 1. 历史和性能原因
- `string` 和 `[]byte` 在底层都是字节数组，转换开销小
- Go 语言设计者认为这两种类型转换足够常见，值得语法支持
- 内置转换语法更简洁，性能更好

### 2. 类型安全考虑
- 其他类型转换可能失败（如 `"abc"` 无法转为数字）
- 需要明确的错误处理机制，所以使用函数返回 `(result, error)`
- 强制开发者处理可能的转换错误

### 3. 灵活性需求
- 数值转换支持不同进制（二进制、八进制、十进制、十六进制）
- 支持不同位大小（8位、16位、32位、64位）
- JSON 解析需要处理复杂的嵌套结构

## 💡 实用技巧和最佳实践

### 1. 安全的转换函数

```go
// 快速判断字符串是否为数字
func isNumeric(s string) bool {
    _, err := strconv.Atoi(s)
    return err == nil
}

// 安全的字符串转数字（带默认值）
func safeParseInt(s string, defaultValue int) int {
    if num, err := strconv.Atoi(s); err == nil {
        return num
    }
    return defaultValue
}

// 安全的字符串转数字（带范围检查）
func safeParseIntRange(s string, min, max, defaultValue int) int {
    num, err := strconv.Atoi(s)
    if err != nil {
        return defaultValue
    }
    if num < min || num > max {
        return defaultValue
    }
    return num
}
```

### 2. 批量转换

```go
// 字符串切片转整数切片
func stringSliceToIntSlice(strs []string) []int {
    nums := make([]int, 0, len(strs))
    for _, str := range strs {
        if num, err := strconv.Atoi(str); err == nil {
            nums = append(nums, num)
        }
    }
    return nums
}

// 处理转换失败的详细日志
func stringSliceToIntSliceWithLogging(strs []string, logger *log.Logger) []int {
    nums := make([]int, 0, len(strs))
    for i, str := range strs {
        if num, err := strconv.Atoi(str); err == nil {
            nums = append(nums, num)
        } else {
            logger.Printf("索引 %d: 无法转换 '%s' 为整数: %v", i, str, err)
        }
    }
    return nums
}
```

### 3. 配置文件读取的最佳实践

```go
type Config struct {
    MaxPoolSize    uint64
    MinPoolSize    uint64
    ConnectionTimeout time.Duration
}

func loadConfig() (*Config, error) {
    config := &Config{}

    // 读取数值配置
    if maxPool := os.Getenv("MONGO_MAX_POOL_SIZE"); maxPool != "" {
        if size, err := strconv.ParseUint(maxPool, 10, 64); err == nil {
            config.MaxPoolSize = size
        } else {
            return nil, fmt.Errorf("无效的 MONGO_MAX_POOL_SIZE: %v", err)
        }
    }

    // 读取超时配置
    if timeout := os.Getenv("MONGO_TIMEOUT"); timeout != "" {
        if duration, err := time.ParseDuration(timeout); err == nil {
            config.ConnectionTimeout = duration
        } else {
            return nil, fmt.Errorf("无效的 MONGO_TIMEOUT: %v", err)
        }
    }

    // 设置默认值
    if config.MaxPoolSize == 0 {
        config.MaxPoolSize = 10  // 默认值
    }

    return config, nil
}
```

### 4. 错误处理的最佳实践

```go
// 错误处理的模式匹配
func parseUserInput(input string) (int, error) {
    // 尝试转换
    num, err := strconv.Atoi(input)
    if err != nil {
        // 根据错误类型提供不同的错误信息
        if numError, ok := err.(*strconv.NumError); ok {
            switch numError.Err {
            case strconv.ErrRange:
                return 0, fmt.Errorf("数字 '%s' 超出范围", input)
            case strconv.ErrSyntax:
                return 0, fmt.Errorf("'%s' 不是有效的数字", input)
            default:
                return 0, fmt.Errorf("解析 '%s' 时发生未知错误: %v", input, err)
            }
        }
        return 0, fmt.Errorf("解析失败: %v", err)
    }

    // 额外的业务逻辑验证
    if num < 0 {
        return 0, fmt.Errorf("数字不能为负数")
    }
    if num > 1000 {
        return 0, fmt.Errorf("数字不能超过1000")
    }

    return num, nil
}
```

## ⚠️ 常见错误和陷阱

### 1. 忘记处理错误

```go
// ❌ 错误：忽略错误
num, _ := strconv.Atoi("abc")  // num = 0，但转换失败了

// ✅ 正确：处理错误
num, err := strconv.Atoi("abc")
if err != nil {
    log.Printf("转换失败: %v", err)
    return
}
fmt.Println("转换成功:", num)
```

### 2. 字符编码问题

```go
// UTF-8 字符处理
str := "你好"
bytes := []byte(str)  // 会被编码为 UTF-8 字节
fmt.Println(len(bytes))  // 6 (每个中文字符3个字节)
fmt.Println(len(str))     // 2 (2个字符)

// 反向转换
str2 := string(bytes)
fmt.Println(str2)  // "你好"
```

### 3. JSON 解析陷阱

```go
// ❌ 错误：直接解析 string
var data map[string]interface{}
err := json.Unmarshal("json string", &data)  // 编译错误！

// ✅ 正确：先转换为 []byte
jsonStr := `{"name": "张三", "age": 25}`
err = json.Unmarshal([]byte(jsonStr), &data)  // 正确
```

## 📚 相关函数速查

### strconv 包常用函数

```go
// 字符串 -> 整数
strconv.Atoi(s string) (int, error)                    // 10进制字符串转int
strconv.ParseInt(s string, base, bitSize int) (int64, error)  // 指定进制
strconv.ParseUint(s string, base, bitSize int) (uint64, error)

// 整数 -> 字符串
strconv.Itoa(i int) string                             // int转10进制字符串
strconv.FormatInt(i int64, base int) string            // 指定进制
strconv.FormatUint(i uint64, base int) string

// 浮点数
strconv.ParseFloat(s string, bitSize int) (float64, error)
strconv.FormatFloat(f float64, fmt byte, prec, bitSize int) string

// 布尔值
strconv.ParseBool(s string) (bool, error)
strconv.FormatBool(b bool) string

// 其他格式化
strconv.Quote(s string) string           // 添加引号
strconv.Unquote(s string) (string, error) // 移除引号
```

### encoding/json 包常用函数

```go
// 解析JSON
json.Unmarshal(data []byte, v interface{}) error

// 生成JSON
json.Marshal(v interface{}) ([]byte, error)
json.MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
```

## 🎓 总结

1. **只有 `string` 和 `[]byte` 可以直接转换**，因为它们底层都是字节数组
2. **其他类型转换需要使用专门的包**（`strconv`、`encoding/json` 等）
3. **必须处理转换错误**，Go 强制开发者关注可能的失败情况
4. **选择合适的转换函数**，根据具体需求（进制、精度、错误处理等）
5. **使用最佳实践**，包括安全转换、默认值、详细错误处理等

掌握这些转换技巧是 Go 编程的基础技能，特别是在处理配置文件、用户输入、API 数据和数据库交互时。