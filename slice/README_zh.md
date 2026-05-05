# Slice - Go 的泛型切片包装器

[English](./README.md) | 简体中文

`slice` 是一个 Go 泛型切片包装器，提供了一套丰富的实用方法来进行常见的切片操作，其 API 设计受到 JavaScript 数组方法和函数式编程模式的启发。

## 特性

- **泛型支持**: 使用 Go 泛型（Go 1.18+）支持任意类型
- **链式 API**: 大多数方法返回 `*Slice[T]` 以支持方法链
- **丰富操作**: Push、pop、shift、unshift、insert、remove、filter、map、reduce、sort 等
- **熟悉的语法**: API 受 JavaScript 数组启发，易于上手
- **类型安全**: 泛型提供完整的编译时类型检查

## 设计原则

- **容器操作**（Push、Pop、Shift、Unshift、Insert、Remove、Set、Clear）：原地修改，返回 self 以支持链式调用。
- **原地变换**（Sort、Reverse）：原地修改，与 Go 标准库一致。
- **不可变变换**（Map、Filter、Concat、Sub、Clone）：返回新 `*Slice[T]`，不修改原切片。
- **查询**（Find、FindIndex、First、Last、Every、Some、Reduce 等）：返回计算结果。

## 安装

```bash
go get github.com/wsshow/op/slice
```

## 使用示例

### 创建和基本操作

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/slice"
)

func main() {
    s := slice.New(1, 2, 3)

    s.Push(4, 5)
    fmt.Println(s.Data()) // 输出: [1 2 3 4 5]

    last, ok := s.Pop()
    fmt.Println(last, ok)  // 输出: 5 true
    fmt.Println(s.Data())  // 输出: [1 2 3 4]
}
```

### 类数组操作

```go
s := slice.New(1, 2, 3)

s.Unshift(0)
fmt.Println(s.Data()) // 输出: [0 1 2 3]

first, ok := s.Shift()
fmt.Println(first, ok) // 输出: 0 true
fmt.Println(s.Data())  // 输出: [1 2 3]
```

### 插入和删除

```go
s := slice.New(1, 3).Insert(1, 2)
fmt.Println(s.Data()) // 输出: [1 2 3]

val, ok := s.Remove(1)
fmt.Println(val, ok)  // 输出: 2 true
fmt.Println(s.Data()) // 输出: [1 3]
```

### 过滤和映射

```go
numbers := slice.New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

evens := numbers.Filter(func(x int) bool {
    return x%2 == 0
})
fmt.Println(evens.Data()) // 输出: [2 4 6 8 10]
// 原始切片不变
fmt.Println(numbers.Data()) // 输出: [1 2 3 4 5 6 7 8 9 10]

doubled := numbers.Map(func(x int) int {
    return x * 2
})
fmt.Println(doubled.Data())  // 输出: [2 4 6 8 10 12 14 16 18 20]
fmt.Println(numbers.Data())  // 输出: [1 2 3 4 5 6 7 8 9 10]（未修改）
```

### 跨类型映射

```go
s := slice.New(1, 2, 3)
strs := slice.MapTo(s, func(v int) string {
    return fmt.Sprintf("n=%d", v)
})
fmt.Println(strs.Data()) // 输出: [n=1 n=2 n=3]
```

### 搜索

```go
users := slice.New(
    struct{ Name string; Age int }{"Alice", 25},
    struct{ Name string; Age int }{"Bob", 30},
    struct{ Name string; Age int }{"Charlie", 35},
)

user, found := users.Find(func(u struct{ Name string; Age int }) bool {
    return u.Age > 28
})
if found {
    fmt.Println(user.Name) // 输出: Bob
}

// 按索引查找
idx := users.FindIndex(func(u struct{ Name string; Age int }) bool {
    return u.Age > 28
})
fmt.Println(idx) // 输出: 1
```

### IndexOf 和 Contains（可比较类型）

```go
names := slice.New("Alice", "Bob", "Charlie")

index := slice.IndexOf(names, "Bob")
fmt.Println(index) // 输出: 1

index = slice.IndexOf(names, "David")
fmt.Println(index) // 输出: -1

fmt.Println(slice.Contains(names, "Bob"))   // 输出: true
fmt.Println(slice.Contains(names, "David")) // 输出: false
```

### 首尾元素

```go
s := slice.New(10, 20, 30)

first, ok := s.First()
fmt.Println(first, ok) // 输出: 10 true

last, ok := s.Last()
fmt.Println(last, ok) // 输出: 30 true
```

### 检查条件

```go
numbers := slice.New(2, 4, 6, 8, 10)

allEven := numbers.Every(func(x int) bool {
    return x%2 == 0
})
fmt.Println(allEven) // 输出: true

someGreater := numbers.Some(func(x int) bool {
    return x > 5
})
fmt.Println(someGreater) // 输出: true
```

### 归约

```go
numbers := slice.New(1, 2, 3, 4, 5)

sum := numbers.Reduce(func(acc, curr int) int {
    return acc + curr
}, 0)
fmt.Println(sum) // 输出: 15

// 使用不同累加器类型
total := slice.ReduceTo(numbers, func(acc string, v int) string {
    return acc + fmt.Sprint(v)
}, "")
fmt.Println(total) // 输出: 12345
```

### 排序和反转

```go
numbers := slice.New(5, 2, 8, 1, 9, 3)

numbers.Sort(func(a, b int) bool {
    return a < b
})
fmt.Println(numbers.Data()) // 输出: [1 2 3 5 8 9]

numbers.Reverse()
fmt.Println(numbers.Data()) // 输出: [9 8 5 3 2 1]
```

### 连接

```go
s1 := slice.New(1, 2)
s2 := slice.New(3, 4)
s3 := slice.New(5, 6)

combined := s1.Concat(s2, s3)
fmt.Println(combined.Data()) // 输出: [1 2 3 4 5 6]
```

### 子切片

```go
s := slice.New(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

sub := s.Sub(2, 5)
fmt.Println(sub.Data()) // 输出: [2 3 4]

// 原始切片不变
fmt.Println(s.Length()) // 输出: 10
```

### 获取、设置和迭代

```go
s := slice.New(10, 20, 30, 40, 50)

value, ok := s.Get(2)
if ok {
    fmt.Println(value) // 输出: 30
}

success := s.Set(3, 99)
fmt.Println(success)   // 输出: true
fmt.Println(s.Data())  // 输出: [10 20 30 99 50]

// 迭代
s.ForEach(func(v int) {
    fmt.Println(v)
})

// 带索引的迭代
s.ForEachIndex(func(i, v int) {
    fmt.Printf("[%d] = %d\n", i, v)
})
```

### 清空和克隆

```go
s := slice.New(1, 2, 3, 4, 5)

clone := s.Clone()
fmt.Println(clone.Data()) // 输出: [1 2 3 4 5]

s.Clear()
fmt.Println(s.Length())   // 输出: 0
fmt.Println(s.IsEmpty())  // 输出: true

fmt.Println(clone.Data()) // 输出: [1 2 3 4 5]
```

### Data 和 Raw 访问

```go
s := slice.New(1, 2, 3)

// Data 返回安全副本
data := s.Data()
data[0] = 10
fmt.Println(s.Data()) // 输出: [1 2 3]（未被修改）

// Raw 直接返回底层切片（无拷贝）
raw := s.Raw()
fmt.Println(raw) // 输出: [1 2 3]
```

### 复杂示例：数据处理管道

```go
type Product struct {
    Name  string
    Price float64
    Stock int
}

products := slice.New(
    Product{"Laptop", 999.99, 5},
    Product{"Mouse", 29.99, 50},
    Product{"Keyboard", 79.99, 0},
    Product{"Monitor", 299.99, 10},
    Product{"USB Cable", 9.99, 100},
)

expensive := products.
    Filter(func(p Product) bool {
        return p.Stock > 0 && p.Price > 50
    }).
    Map(func(p Product) Product {
        p.Price *= 1.1
        return p
    }).
    Sort(func(a, b Product) bool {
        return a.Price > b.Price
    })

expensive.ForEach(func(p Product) {
    fmt.Printf("%s: $%.2f (库存: %d)\n", p.Name, p.Price, p.Stock)
})
// 输出:
// Laptop: $1099.99 (库存: 5)
// Monitor: $329.99 (库存: 10)
```

## API 概览

### 创建
- `New[T any](values ...T) *Slice[T]`: 创建包含初始值的新切片

### 容器操作（原地修改，返回 self）

| 方法 | 描述 |
|------|------|
| `Push(values ...T) *Slice[T]` | 在末尾添加元素 |
| `Pop() (T, bool)` | 移除并返回最后一个元素 |
| `Unshift(values ...T) *Slice[T]` | 在开头添加元素 |
| `Shift() (T, bool)` | 移除并返回第一个元素 |
| `Insert(index int, values ...T) *Slice[T]` | 在指定索引处插入元素 |
| `Remove(index int) (T, bool)` | 移除指定索引处的元素 |
| `Set(index int, value T) bool` | 设置指定索引处的值 |
| `Clear() *Slice[T]` | 清空所有元素 |

### 原地变换（修改自身，返回 self）

| 方法 | 描述 |
|------|------|
| `Sort(less func(a, b T) bool) *Slice[T]` | 排序 |
| `Reverse() *Slice[T]` | 反转顺序 |

### 不可变变换（返回新 Slice）

| 方法 | 描述 |
|------|------|
| `Map(fn func(T) T) *Slice[T]` | 映射每个元素 |
| `Filter(predicate func(T) bool) *Slice[T]` | 过滤元素 |
| `Concat(others ...*Slice[T]) *Slice[T]` | 连接多个切片 |
| `Sub(start, end int) *Slice[T]` | 获取子切片 |
| `Clone() *Slice[T]` | 创建深拷贝 |

### 包级函数

| 函数 | 描述 |
|------|------|
| `IndexOf[T comparable](s *Slice[T], value T) int` | 查找值的索引 |
| `Contains[T comparable](s *Slice[T], value T) bool` | 检查是否包含 |
| `MapTo[T, U any](s *Slice[T], fn func(T) U) *Slice[U]` | 跨类型映射 |
| `ReduceTo[T, U any](s *Slice[T], fn func(U, T) U, initial U) U` | 跨类型归约 |

### 查询

| 方法 | 描述 |
|------|------|
| `Length() int` | 元素数量 |
| `IsEmpty() bool` | 检查是否为空 |
| `Get(index int) (T, bool)` | 获取指定索引的元素 |
| `Find(predicate func(T) bool) (T, bool)` | 查找第一个匹配的元素 |
| `FindIndex(predicate func(T) bool) int` | 查找第一个匹配的索引 |
| `First() (T, bool)` | 获取第一个元素 |
| `Last() (T, bool)` | 获取最后一个元素 |
| `Every(predicate func(T) bool) bool` | 检查是否全部匹配 |
| `Some(predicate func(T) bool) bool` | 检查是否有匹配 |
| `Reduce(fn func(prev, curr T) T, initialValue T) T` | 归约为单个值 |
| `ForEach(fn func(T)) *Slice[T]` | 对每个元素执行函数 |
| `ForEachIndex(fn func(int, T)) *Slice[T]` | 带索引遍历 |
| `Data() []T` | 获取底层切片的安全副本 |
| `Raw() []T` | 直接获取底层切片（调用者不应修改） |

## 许可证

MIT License
