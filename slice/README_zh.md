# Slice - Go 的泛型切片包装器

[English](./README.md) | 简体中文

`slice` 是一个 Go 泛型切片包装器，提供了一套丰富的实用方法来进行常见的切片操作，其 API 设计受到 JavaScript 数组方法和函数式编程模式的启发。

## 特性

- **泛型支持**: 使用 Go 泛型（Go 1.18+）支持任意类型
- **链式 API**: 大多数方法返回 `*Slice[T]` 以支持方法链
- **丰富操作**: Push、pop、shift、unshift、filter、map、reduce、sort 等
- **熟悉的语法**: API 受 JavaScript 数组启发，易于上手
- **类型安全**: 泛型提供完整的编译时类型检查

## 安装

将包添加到你的 Go 项目中：

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
    // 创建一个新切片
    s := slice.New(1, 2, 3)
    
    // 添加元素
    s.Push(4, 5)
    fmt.Println(s.Data()) // 输出: [1 2 3 4 5]
    
    // 移除最后一个元素
    last := s.Pop()
    fmt.Println(last)     // 输出: 5
    fmt.Println(s.Data()) // 输出: [1 2 3 4]
}
```

### 类数组操作

```go
s := slice.New(1, 2, 3)

// 在开头添加
s.Unshift(0)
fmt.Println(s.Data()) // 输出: [0 1 2 3]

// 从开头移除
first := s.Shift()
fmt.Println(first)     // 输出: 0
fmt.Println(s.Data())  // 输出: [1 2 3]
```

### 过滤和映射

```go
numbers := slice.New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

// 过滤偶数
evens := numbers.Filter(func(x int) bool {
    return x%2 == 0
})
fmt.Println(evens.Data()) // 输出: [2 4 6 8 10]

// 将每个数字翻倍（就地修改）
numbers.Map(func(x int) int {
    return x * 2
})
fmt.Println(numbers.Data()) // 输出: [2 4 6 8 10 12 14 16 18 20]
```

### 搜索

```go
users := slice.New(
    struct{ Name string; Age int }{"Alice", 25},
    struct{ Name string; Age int }{"Bob", 30},
    struct{ Name string; Age int }{"Charlie", 35},
)

// 查找第一个年龄大于28的用户
user, found := users.Find(func(u struct{ Name string; Age int }) bool {
    return u.Age > 28
})
if found {
    fmt.Println(user.Name) // 输出: Bob
}
```

### 查找索引（可比较类型）

```go
names := slice.New("Alice", "Bob", "Charlie")

// 查找 "Bob" 的索引
index := slice.IndexOf(names, "Bob")
fmt.Println(index) // 输出: 1

// 未找到返回 -1
index = slice.IndexOf(names, "David")
fmt.Println(index) // 输出: -1
```

### 检查条件

```go
numbers := slice.New(2, 4, 6, 8, 10)

// 检查是否全部为偶数
allEven := numbers.Every(func(x int) bool {
    return x%2 == 0
})
fmt.Println(allEven) // 输出: true

// 检查是否有大于5的
someGreater := numbers.Some(func(x int) bool {
    return x > 5
})
fmt.Println(someGreater) // 输出: true
```

### 归约

```go
numbers := slice.New(1, 2, 3, 4, 5)

// 求和
sum := numbers.Reduce(func(acc, curr int) int {
    return acc + curr
}, 0)
fmt.Println(sum) // 输出: 15

// 找最大值
max := numbers.Reduce(func(acc, curr int) int {
    if curr > acc {
        return curr
    }
    return acc
}, 0)
fmt.Println(max) // 输出: 5
```

### 排序

```go
numbers := slice.New(5, 2, 8, 1, 9, 3)

// 升序排序
numbers.Sort(func(a, b int) bool {
    return a < b
})
fmt.Println(numbers.Data()) // 输出: [1 2 3 5 8 9]

// 降序排序
numbers.Sort(func(a, b int) bool {
    return a > b
})
fmt.Println(numbers.Data()) // 输出: [9 8 5 3 2 1]
```

### 反转

```go
s := slice.New(1, 2, 3, 4, 5)

s.Reverse()
fmt.Println(s.Data()) // 输出: [5 4 3 2 1]
```

### 连接

```go
s1 := slice.New(1, 2, 3)
s2 := slice.New(4, 5, 6)

// 创建包含组合元素的新切片
combined := s1.Concat(s2)
fmt.Println(combined.Data()) // 输出: [1 2 3 4 5 6]

// 原始切片不变
fmt.Println(s1.Data()) // 输出: [1 2 3]
fmt.Println(s2.Data()) // 输出: [4 5 6]
```

### 切分

```go
s := slice.New(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

// 获取索引2到5的元素（不包括5）
sub := s.Slice(2, 5)
fmt.Println(sub.Data()) // 输出: [2 3 4]

// 原始切片不变
fmt.Println(s.Length()) // 输出: 10
```

### 获取和设置元素

```go
s := slice.New(10, 20, 30, 40, 50)

// 获取索引2的元素
value, ok := s.Get(2)
if ok {
    fmt.Println(value) // 输出: 30
}

// 设置索引3的元素
success := s.Set(3, 99)
fmt.Println(success)   // 输出: true
fmt.Println(s.Data())  // 输出: [10 20 30 99 50]
```

### 迭代

```go
s := slice.New("apple", "banana", "cherry")

s.Foreach(func(fruit string) {
    fmt.Println(fruit)
})
// 输出:
// apple
// banana
// cherry
```

### 清空和克隆

```go
s := slice.New(1, 2, 3, 4, 5)

// 克隆切片
clone := s.Clone()
fmt.Println(clone.Data()) // 输出: [1 2 3 4 5]

// 清空原始切片
s.Clear()
fmt.Println(s.Length())   // 输出: 0
fmt.Println(s.IsEmpty())  // 输出: true

// 克隆切片保持不变
fmt.Println(clone.Data()) // 输出: [1 2 3 4 5]
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

// 查找有库存且价格高的产品，并将价格提高10%
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

expensive.Foreach(func(p Product) {
    fmt.Printf("%s: $%.2f (库存: %d)\n", p.Name, p.Price, p.Stock)
})
// 输出:
// Laptop: $1099.99 (库存: 5)
// Monitor: $329.99 (库存: 10)
// Keyboard: $87.99 (库存: 0)
```

## API 概览

### 创建
- `New[T any](values ...T) *Slice[T]`: 创建包含初始值的新切片

### 添加元素
- `Push(values ...T) *Slice[T]`: 在末尾添加元素
- `Unshift(values ...T) *Slice[T]`: 在开头添加元素

### 移除元素
- `Pop() T`: 移除并返回最后一个元素
- `Shift() T`: 移除并返回第一个元素
- `Clear() *Slice[T]`: 移除所有元素

### 查询
- `Length() int`: 获取元素数量
- `IsEmpty() bool`: 检查切片是否为空
- `Get(index int) (T, bool)`: 获取指定索引的元素
- `Set(index int, value T) bool`: 设置指定索引的元素

### 搜索
- `Find(predicate func(T) bool) (T, bool)`: 查找第一个匹配的元素
- `IndexOf[T comparable](s *Slice[T], value T) int`: 查找值的索引

### 过滤和转换
- `Filter(predicate func(T) bool) *Slice[T]`: 过滤元素（返回新切片）
- `Map(callbackfn func(T) T) *Slice[T]`: 转换元素（就地修改）
- `Foreach(callbackfn func(T)) *Slice[T]`: 对每个元素执行函数

### 检查
- `Every(predicate func(T) bool) bool`: 检查是否全部匹配
- `Some(predicate func(T) bool) bool`: 检查是否有匹配

### 聚合
- `Reduce(callbackfn func(prev, curr T) T, initialValue T) T`: 归约为单个值

### 排序和排列
- `Sort(compareFn func(a, b T) bool) *Slice[T]`: 排序元素
- `Reverse() *Slice[T]`: 反转顺序

### 组合和切分
- `Concat(other *Slice[T]) *Slice[T]`: 连接切片（返回新切片）
- `Slice(start, end int) *Slice[T]`: 获取子切片（返回新切片）

### 工具方法
- `Data() []T`: 获取底层切片的副本
- `Clone() *Slice[T]`: 创建深拷贝

## 注意事项

- **可变性**: 某些方法就地修改切片（`Push`、`Pop`、`Map`、`Sort` 等），而其他方法返回新切片（`Filter`、`Concat`、`Slice`、`Clone`）
- **空操作**: 如果切片为空，`Pop()` 和 `Shift()` 返回零值
- **边界检查**: `Get()` 和 `Set()` 执行边界检查并返回/接受布尔标志
- **类型安全**: 所有操作在编译时都是类型安全的，得益于泛型

## 许可证

MIT License
