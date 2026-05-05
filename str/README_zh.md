# Str - 字符串包装器与实用方法

[English](./README.md) | 简体中文

`str` 是一个 Go 字符串包装器，提供了一套丰富的实用方法来进行常见的字符串操作，具有受面向对象编程模式启发的链式 API。

## 特性

- **链式 API**: 修改类方法返回 `*String` 以支持方法链
- **丰富操作**: 包含检查、分割、替换、修剪、大小写转换等
- **类型转换**: 内置错误处理和 panic 变体的 int、float 转换
- **Unicode 感知**: Reverse、Substring 和 RuneLength 基于 rune 实现
- **不可变派生**: Substring、Concat 和 Format 返回新实例，不修改原值
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
    s := str.New("Hello, World!")

    // 检查是否包含子串
    fmt.Println(s.Contains("World")) // 输出: true

    // 查找索引
    fmt.Println(s.Index("World"))     // 输出: 7
    fmt.Println(s.LastIndex("o"))     // 输出: 8

    // 统计出现次数
    fmt.Println(s.Count("o"))         // 输出: 2
}
```

### 链式字符串操作

```go
s := str.New("  Hello, World!  ")

// 链式调用（修改类方法原地修改）
result := s.TrimSpace().
    ReplaceAll("World", "Go").
    ToUpper()

fmt.Println(result.String()) // 输出: HELLO, GO!
```

### 大小写转换

```go
s := str.New("Hello World")

fmt.Println(s.ToLower().String()) // 输出: hello world
fmt.Println(s.ToUpper().String()) // 输出: HELLO WORLD

// 忽略大小写比较
s2 := str.New("Hello")
fmt.Println(s2.EqualFold("hello")) // 输出: true
```

### 分割

```go
s := str.New("apple,banana,cherry")

parts := s.Split(",")
fmt.Println(parts) // 输出: [apple banana cherry]

// 按空白字符分割
s2 := str.New("hello  world\tfoo")
fields := s2.Fields()
fmt.Println(fields) // 输出: [hello world foo]
```

### 前缀和后缀检查

```go
s := str.New("https://example.com")

fmt.Println(s.StartsWith("https://")) // 输出: true
fmt.Println(s.EndsWith(".com"))       // 输出: true

// 修剪前缀/后缀
s.TrimPrefix("https://")
fmt.Println(s.String()) // 输出: example.com
```

### 字符串长度

```go
s := str.New("Hello 世界")

// 字节长度
fmt.Println(s.Length())      // 输出: 12

// Unicode 字符数（rune 长度）
fmt.Println(s.RuneLength())  // 输出: 8
```

### 修剪

```go
s := str.New("...Hello...")

// 修剪指定字符
s.Trim(".")
fmt.Println(s.String()) // 输出: Hello

// 修剪空白字符
s2 := str.New("  spaces  ")
s2.TrimSpace()
fmt.Println(s2.String()) // 输出: spaces

// 修剪前缀/后缀
s3 := str.New("hello.txt")
s3.TrimSuffix(".txt")
fmt.Println(s3.String()) // 输出: hello
```

### 连接

```go
s := str.New("Hello")

// Concat 返回新实例，原值不变
result := s.Concat(", ", "World", "!")
fmt.Println(result.String()) // 输出: Hello, World!
fmt.Println(s.String())      // 输出: Hello（未修改）
```

### 反转

```go
s := str.New("hello")
s.Reverse()
fmt.Println(s.String()) // 输出: olleh

// Unicode 感知
s2 := str.New("你好世界")
s2.Reverse()
fmt.Println(s2.String()) // 输出: 界世好你
```

### 子串

```go
s := str.New("Hello 世界")

// 提取子串（Unicode 感知，返回新实例）
sub := s.Substring(0, 7)
fmt.Println(sub.String()) // 输出: Hello 世

// 原值不变
fmt.Println(s.String()) // 输出: Hello 世界

// 负数索引（从末尾算起）
sub2 := str.New("hello.txt").Substring(0, -4)
fmt.Println(sub2.String()) // 输出: hello
```

### 类型转换

```go
// 字符串转 int
s1 := str.New("42")
num, err := s1.ToInt()
if err == nil {
    fmt.Println(num) // 输出: 42
}

// MustInt 失败时 panic
s2 := str.New("42")
fmt.Println(s2.MustInt()) // 输出: 42

// 字符串转 float
s3 := str.New("3.14")
f, err := s3.ToFloat()
if err == nil {
    fmt.Println(f) // 输出: 3.14
}

// 自动修剪空白字符
s4 := str.New("  123  ")
num, err = s4.ToInt()
fmt.Println(num) // 输出: 123
```

### 格式化

```go
template := str.New("Hello, %s! You have %d messages.")

// Format 返回新实例，模板不变
formatted := template.Format("Alice", 5)

fmt.Println(formatted.String())
// 输出: Hello, Alice! You have 5 messages.

fmt.Println(template.String())
// 输出: Hello, %s! You have %d messages.（不变）
```

### 空值检查

```go
s1 := str.New("")
s2 := str.New("text")
s3 := str.New("   ")

fmt.Println(s1.IsEmpty()) // 输出: true
fmt.Println(s2.IsEmpty()) // 输出: false
fmt.Println(s1.IsBlank()) // 输出: true
fmt.Println(s3.IsBlank()) // 输出: true
fmt.Println(s2.IsBlank()) // 输出: false
```

### 克隆

```go
original := str.New("original")
clone := original.Clone()

clone.ToUpper()

fmt.Println(original.String()) // 输出: original
fmt.Println(clone.String())    // 输出: ORIGINAL
```

### 复杂示例：文本处理

```go
input := str.New("  HELLO@EXAMPLE.COM  ")

email := input.
    TrimSpace().
    ToLower()

if email.EndsWith("@example.com") && !email.IsEmpty() {
    username := email.Clone().
        ReplaceAll("@example.com", "").
        String()

    fmt.Printf("用户名: %s\n", username) // 输出: 用户名: hello
}
```

### 数据解析

```go
data := str.New("Name:Alice,Age:25,City:NYC")

parts := data.Split(",")
for _, part := range parts {
    kv := str.New(part).Split(":")
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
- `New(s string) *String`: 创建新的 String 实例

### 搜索
- `Contains(substr string) bool`: 检查是否包含子串
- `Index(substr string) int`: 查找首次出现的索引（未找到返回 -1）
- `LastIndex(substr string) int`: 查找最后出现的索引（未找到返回 -1）
- `Count(substr string) int`: 统计不重叠的子串出现次数

### 分割
- `Split(sep string) []string`: 按分隔符分割
- `Fields() []string`: 按空白字符分割

### 长度
- `Length() int`: 获取字节长度
- `RuneLength() int`: 获取 Unicode 字符数（rune 长度）

### 修改（可链式，原地修改）
- `ReplaceAll(old, new string) *String`: 替换所有匹配项
- `Replace(old, new string, n int) *String`: 替换前 n 个匹配项
- `Trim(cutset string) *String`: 从两端修剪字符
- `TrimSpace() *String`: 从两端修剪空白字符
- `TrimPrefix(prefix string) *String`: 去除前缀
- `TrimSuffix(suffix string) *String`: 去除后缀
- `ToLower() *String`: 转换为小写
- `ToUpper() *String`: 转换为大写
- `Reverse() *String`: 反转字符串（Unicode 感知）

### 派生（返回新实例，原值不变）
- `Concat(ss ...string) *String`: 连接字符串
- `Substring(start, end int) *String`: 提取子串（Unicode 感知，负数 end 表示从末尾算起）
- `Format(args ...interface{}) *String`: 使用参数格式化字符串

### 检查
- `StartsWith(prefix string) bool`: 检查是否以前缀开头
- `EndsWith(suffix string) bool`: 检查是否以后缀结尾
- `EqualFold(t string) bool`: 忽略大小写比较相等
- `IsEmpty() bool`: 检查字符串是否为空
- `IsBlank() bool`: 检查是否为空或仅含空白字符

### 转换
- `ToInt() (int, error)`: 转换为整数
- `MustInt() int`: 转换为整数，失败时 panic
- `ToFloat() (float64, error)`: 转换为 float64
- `MustFloat() float64`: 转换为 float64，失败时 panic

### 工具方法
- `Clone() *String`: 创建副本
- `String() string`: 获取底层字符串值（满足 fmt.Stringer）

## 设计说明

- **修改 vs 派生**: 修改类方法（Replace、Trim、ToLower 等）原地修改并返回 `*String` 以支持链式调用。派生类方法（Concat、Substring、Format）返回**新**实例且不影响原值。
- **Unicode**: `Reverse()`、`Substring()` 和 `RuneLength()` 是 Unicode 感知的（基于 rune）。
- **转换**: `ToInt()` 和 `ToFloat()` 在解析前会自动修剪空白字符。使用 `MustInt()`/`MustFloat()` 获取失败时 panic 的变体。
- **可变性**: 与 Go 内置不可变字符串不同，修改类方法会修改内部状态以支持链式调用。如需在修改前保留原值，请使用 `Clone()`。

## 许可证

MIT License
