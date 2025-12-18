# Str - 字符串包装器与实用方法

[English](./README.md) | 简体中文

`str` 是一个 Go 字符串包装器，提供了一套丰富的实用方法来进行常见的字符串操作，具有受面向对象编程模式启发的链式 API。

## 特性

- **链式 API**: 大多数方法返回 `*String` 以支持方法链
- **丰富操作**: 包含检查、分割、替换、修剪、大小写转换等
- **类型转换**: 内置错误处理的 int、float 转换
- **Unicode 感知**: 分别提供字节和 rune 长度方法
- **熟悉的方法**: API 受其他语言常见字符串操作启发

## 安装

```bash
go get github.com/wsshow/op/str
```

## 使用示例

### 基本字符串操作

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/str"
)

func main() {
    s := str.NewString("Hello, World!")
    
    // 检查是否包含子串
    fmt.Println(s.Contain("World")) // 输出: true
    
    // 查找索引
    fmt.Println(s.Index("World"))     // 输出: 7
    fmt.Println(s.LastIndex("o"))     // 输出: 8
}
```

### 链式字符串操作

```go
s := str.NewString("  Hello, World!  ")

// 链式调用多个操作
result := s.TrimSpace().
    ReplaceAll("World", "Go").
    ToUpper()

fmt.Println(result.String()) // 输出: HELLO, GO!
```

### 大小写转换

```go
s := str.NewString("Hello World")

fmt.Println(s.ToLower().String()) // 输出: hello world
fmt.Println(s.ToUpper().String()) // 输出: HELLO WORLD
```

### 分割

```go
s := str.NewString("apple,banana,cherry")

parts := s.Split(",")
fmt.Println(parts) // 输出: [apple banana cherry]
```

### 前缀和后缀检查

```go
s := str.NewString("hello.txt")

fmt.Println(s.StartsWith("hello")) // 输出: true
fmt.Println(s.EndsWith(".txt"))     // 输出: true
fmt.Println(s.EndsWith(".pdf"))     // 输出: false
```

### 字符串长度

```go
s := str.NewString("Hello 世界")

// 字节长度
fmt.Println(s.Length())      // 输出: 12

// Unicode 字符数（rune 长度）
fmt.Println(s.RuneLength())  // 输出: 8
```

### 修剪

```go
s := str.NewString("...Hello...")

// 修剪指定字符
s.Trim(".")
fmt.Println(s.String()) // 输出: Hello

// 修剪空白字符
s2 := str.NewString("  spaces  ")
s2.TrimSpace()
fmt.Println(s2.String()) // 输出: spaces
```

### 连接

```go
s := str.NewString("Hello")

s.Concat(", ", "World", "!")
fmt.Println(s.String()) // 输出: Hello, World!
```

### 子串

```go
s := str.NewString("Hello 世界")

// 提取子串（Unicode 感知）
sub := s.Substring(0, 7)
fmt.Println(sub.String()) // 输出: Hello 世
```

### 类型转换

```go
// 字符串转 int
s1 := str.NewString("42")
num, err := s1.ToInt()
if err == nil {
    fmt.Println(num) // 输出: 42
}

// 字符串转 float
s2 := str.NewString("3.14")
f, err := s2.ToFloat()
if err == nil {
    fmt.Println(f) // 输出: 3.14
}

// 带空白字符自动修剪
s3 := str.NewString("  123  ")
num, err = s3.ToInt() // 自动修剪空白字符
fmt.Println(num)       // 输出: 123
```

### 格式化

```go
template := str.NewString("Hello, %s! You have %d messages.")
formatted := template.Format("Alice", 5)

fmt.Println(formatted.String()) 
// 输出: Hello, Alice! You have 5 messages.
```

### 检查空字符串

```go
s1 := str.NewString("")
s2 := str.NewString("text")

fmt.Println(s1.IsEmpty()) // 输出: true
fmt.Println(s2.IsEmpty()) // 输出: false
```

### 克隆

```go
original := str.NewString("original")
clone := original.Clone()

clone.ToUpper()

fmt.Println(original.String()) // 输出: original
fmt.Println(clone.String())    // 输出: ORIGINAL
```

### 复杂示例：文本处理

```go
// 处理用户输入
input := str.NewString("  HELLO@EXAMPLE.COM  ")

email := input.
    TrimSpace().
    ToLower().
    Clone()

if email.EndsWith("@example.com") && !email.IsEmpty() {
    username := email.
        ReplaceAll("@example.com", "").
        String()
    
    fmt.Printf("用户名: %s\n", username) // 输出: 用户名: hello
}
```

### URL 处理

```go
url := str.NewString("https://example.com/api/v1/users")

if url.StartsWith("https://") {
    path := str.NewString(url.String())
    path.ReplaceAll("https://example.com", "")
    fmt.Println(path.String()) // 输出: /api/v1/users
}
```

### 数据解析

```go
data := str.NewString("Name:Alice,Age:25,City:NYC")

parts := data.Split(",")
for _, part := range parts {
    kv := str.NewString(part).Split(":")
    if len(kv) == 2 {
        fmt.Printf("%s = %s\n", kv[0], kv[1])
    }
}
// 输出:
// Name = Alice
// Age = 25
// City = NYC
```

## API 概览

### 创建
- `NewString(s string) *String`: 创建新的 String 实例

### 搜索
- `Contain(substr string) bool`: 检查是否包含子串
- `Index(substr string) int`: 查找首次出现的索引（未找到返回 -1）
- `LastIndex(substr string) int`: 查找最后出现的索引（未找到返回 -1）

### 分割
- `Split(sep string) []string`: 分割为切片

### 长度
- `Length() int`: 获取字节长度
- `RuneLength() int`: 获取 Unicode 字符数（rune 长度）

### 修改（可链式）
- `ReplaceAll(old, new string) *String`: 替换所有匹配项
- `Trim(cutset string) *String`: 从两端修剪字符
- `TrimSpace() *String`: 从两端修剪空白字符
- `ToLower() *String`: 转换为小写
- `ToUpper() *String`: 转换为大写
- `Concat(ss ...string) *String`: 连接字符串
- `Substring(start, end int) *String`: 提取子串（Unicode 感知）

### 检查
- `StartsWith(prefix string) bool`: 检查是否以前缀开头
- `EndsWith(suffix string) bool`: 检查是否以后缀结尾
- `IsEmpty() bool`: 检查字符串是否为空

### 转换
- `ToInt() (int, error)`: 转换为整数
- `ToFloat() (float64, error)`: 转换为 float64
- `Format(args ...interface{}) *String`: 使用参数格式化字符串

### 工具方法
- `Clone() *String`: 创建副本
- `String() string`: 获取底层字符串值

## 注意事项

- **链式调用**: 修改字符串的方法返回 `*String` 以支持链式调用
- **可变性**: 与 Go 内置字符串不同，String 方法会修改内部状态
- **Unicode**: `Substring()` 和 `RuneLength()` 是 Unicode 感知的（基于 rune）
- **转换**: `ToInt()` 和 `ToFloat()` 在解析前会自动修剪空白字符

## 许可证

MIT License
