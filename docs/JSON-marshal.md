# JSON到Go interface{}的类型映射规则

当使用`json.Unmarshal`解析JSON到`interface{}`时，Go标准库有以下默认的类型映射规则：

## 基本类型映射

| JSON类型 | Go interface{}中的类型    | 说明 |
|---------|--------------------------|------|
| JSON对象 | `map[string]interface{}` | 所有键值对对象 |
| JSON数组 | `[]interface{}`.         | 所有数组类型 |
| JSON字符串 | `string`               | 所有字符串 |
| JSON数字 | `float64` | **无论是否为整数**，都解析为float64 |
| JSON布尔值 | `bool` | true/false |
| JSON null | `nil` | 空值 |

## 重要注意事项

### 1. 数字类型处理
```go
// JSON: {"age": 25, "height": 175.5}
// 解析后:
age := data["age"].(float64)        // 25.0 (虽然是整数，但类型是float64)
height := data["height"].(float64)  // 175.5

// 转换为整数:
ageInt := int(age)                  // 25
```

### 2. 嵌套结构
```go
// JSON: {"user": {"name": "张三", "age": 25}}
// 解析后:
user := data["user"].(map[string]interface{})  // 嵌套对象也是map[string]interface{}
name := user["name"].(string)                  // "张三"
age := int(user["age"].(float64))              // 25
```

### 3. 数组处理
```go
// JSON: {"scores": [90, 85, 95]}
// 解析后:
scores := data["scores"].([]interface{})       // 数组类型
for _, score := range scores {
    scoreInt := int(score.(float64))           // 每个元素仍然是float64
    fmt.Println(scoreInt)                      // 90, 85, 95
}
```

## 在你的项目中的应用

在`UserCreditService`中：
```go
func (us *UserCreditService) ReadUserCredit(userId string) (map[string]interface{}, error) {
    // 从区块链获取JSON数据
    result, err := us.contract.EvaluateTransaction("ReadUserCredit", userId)

    var data map[string]interface{}
    err = json.Unmarshal(result, &data)  // 数字变成float64

    return data, nil
}

// 使用时需要类型转换
credit := int(userCredit["credit"].(float64))         // 必须这样转换
totalUsed := int(userCredit["totalUsed"].(float64))   // 即使原始数据是int
totalRecharge := int(userCredit["totalRecharge"].(float64))
```

## 更好的解决方案

### 1. 定义结构体（推荐）
```go
type UserCredit struct {
    Credit        int `json:"credit"`
    TotalUsed     int `json:"totalUsed"`
    TotalRecharge int `json:"totalRecharge"`
}

func (us *UserCreditService) ReadUserCredit(userId string) (*UserCredit, error) {
    result, err := us.contract.EvaluateTransaction("ReadUserCredit", userId)
    if err != nil {
        return nil, err
    }

    var data UserCredit
    err = json.Unmarshal(result, &data)  // 直接解析到结构体
    if err != nil {
        return nil, err
    }

    return &data, nil  // 无需类型转换
}
```

### 2. 使用更安全的类型检查
```go
func safeInt(value interface{}) (int, error) {
    switch v := value.(type) {
    case float64:
        return int(v), nil
    case int:
        return v, nil
    default:
        return 0, fmt.Errorf("无法将 %v 转换为int", value)
    }
}

credit, err := safeInt(userCredit["credit"])
if err != nil {
    return err
}
```

## 总结

Go的JSON解析器为了确保精度和一致性，将所有数字统一解析为`float64`。这是Go语言的设计选择，虽然有时看起来有些繁琐，但可以避免精度丢失和类型歧义。

在项目中，建议：
1. 优先使用结构体定义明确的数据结构
2. 如果必须使用`map[string]interface{}`，记得正确处理数字类型转换
3. 考虑使用类型断言的安全形式避免panic