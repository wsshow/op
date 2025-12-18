# LINQ - Go 的 LINQ 风格查询 API

[English](./README.md) | 简体中文

`linq` 是一个针对 Go 切片的泛型 LINQ 风格查询库，提供流畅的链式 API 来执行过滤、映射、排序、分组等常见数据转换操作。

## 特性

- **泛型支持**: 使用 Go 泛型（Go 1.18+）支持任意类型
- **链式 API**: 支持方法链式调用，实现优雅的查询表达式
- **丰富操作**: 过滤、映射、排序、分组、去重、分页、连接等
- **自定义比较器**: 灵活的比较函数支持，用于排序和去重
- **零依赖**: 纯 Go 实现，无外部依赖

## 安装

将包添加到你的 Go 项目中：

```bash
go get github.com/wsshow/op/linq
```

## 使用示例

### 基本操作

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/linq"
)

func main() {
    data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // 过滤偶数并翻倍
    result := linq.From(data).
        Where(func(x int) bool { return x%2 == 0 }).
        Select(func(x int) int { return x * 2 }).
        Results()
    
    fmt.Println(result) // 输出: [4 8 12 16 20]
}
```

### 排序和获取元素

```go
data := []int{5, 2, 8, 1, 9, 3}

// 降序排序并取前3个
top3 := linq.From(data).
    Sort(func(a, b int) bool { return a > b }).
    Take(3).
    Results()

fmt.Println(top3) // 输出: [9 8 5]
```

### 使用 Comparable 类型去重

```go
data := []int{1, 2, 2, 3, 3, 3, 4}

// 为可比较类型去重
unique := linq.DistinctComparable(linq.From(data)).
    Results()

fmt.Println(unique) // 输出: [1 2 3 4]
```

### 使用自定义比较器去重

```go
type Person struct {
    Name string
    Age  int
}

people := []Person{
    {"Alice", 25},
    {"Bob", 30},
    {"Alice", 25}, // 重复
    {"Charlie", 35},
}

// 使用自定义比较器去重
unique := linq.From(people).
    WithComparer(func(a, b Person) int {
        if a.Name != b.Name {
            if a.Name < b.Name {
                return -1
            }
            return 1
        }
        return a.Age - b.Age
    }).
    Distinct().
    Results()

fmt.Println(len(unique)) // 输出: 3
```

### 分组

```go
type Product struct {
    Name     string
    Category string
    Price    float64
}

products := []Product{
    {"Laptop", "Electronics", 999.99},
    {"Mouse", "Electronics", 29.99},
    {"Desk", "Furniture", 299.99},
    {"Chair", "Furniture", 199.99},
}

// 按类别分组
groups := linq.GroupBy(linq.From(products), func(p Product) string {
    return p.Category
})

for _, group := range groups {
    fmt.Printf("%s: %d 项\n", group.Key, len(group.Items))
}
// 输出:
// Electronics: 2 项
// Furniture: 2 项
```

### 连接

```go
type Order struct {
    ID         int
    CustomerID int
    Amount     float64
}

type Customer struct {
    ID   int
    Name string
}

orders := []Order{
    {1, 101, 50.0},
    {2, 102, 75.0},
    {3, 101, 100.0},
}

customers := []Customer{
    {101, "Alice"},
    {102, "Bob"},
}

// 连接订单和客户
type OrderDetail struct {
    OrderID      int
    CustomerName string
    Amount       float64
}

result := linq.Join(
    linq.From(orders),
    linq.From(customers),
    func(o Order) int { return o.CustomerID },
    func(c Customer) int { return c.ID },
    func(o Order, c Customer) OrderDetail {
        return OrderDetail{o.ID, c.Name, o.Amount}
    },
).Results()

for _, detail := range result {
    fmt.Printf("订单 #%d - %s: $%.2f\n", 
        detail.OrderID, detail.CustomerName, detail.Amount)
}
// 输出:
// 订单 #1 - Alice: $50.00
// 订单 #2 - Bob: $75.00
// 订单 #3 - Alice: $100.00
```

### 使用比较器求最小/最大值

```go
data := []int{5, 2, 8, 1, 9, 3}

linq := linq.From(data).WithComparer(func(a, b int) int {
    return a - b
})

min, _ := linq.Min()
max, _ := linq.Max()

fmt.Printf("最小: %d, 最大: %d\n", min, max) // 输出: 最小: 1, 最大: 9
```

### 分页

```go
data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// 跳过3个，取4个（分页）
page := linq.From(data).
    Skip(3).
    Take(4).
    Results()

fmt.Println(page) // 输出: [4 5 6 7]
```

### 检查条件

```go
data := []int{2, 4, 6, 8, 10}

// 检查是否有元素大于5
hasLarge := linq.From(data).Any(func(x int) bool { return x > 5 })
fmt.Println(hasLarge) // 输出: true
```

### 复杂查询链

```go
type Student struct {
    Name  string
    Grade int
    Score float64
}

students := []Student{
    {"Alice", 10, 95.5},
    {"Bob", 10, 78.0},
    {"Charlie", 11, 88.5},
    {"David", 11, 92.0},
    {"Eve", 10, 85.0},
}

// 查找10年级分数最高的2名学生
topStudents := linq.From(students).
    Where(func(s Student) bool { return s.Grade == 10 }).
    Sort(func(a, b Student) bool { return a.Score > b.Score }).
    Take(2).
    Results()

for _, s := range topStudents {
    fmt.Printf("%s: %.1f\n", s.Name, s.Score)
}
// 输出:
// Alice: 95.5
// Eve: 85.0
```

## API 概览

### 创建
- `From[T any](data []T) Linq[T]`: 从切片创建 Linq 实例

### 过滤
- `Where(predicate func(T) bool) Linq[T]`: 按条件过滤元素
- `Any(predicate func(T) bool) bool`: 检查是否有元素满足条件
- `Distinct() Linq[T]`: 去重（需要比较器）
- `DistinctComparable[T comparable](l Linq[T]) Linq[T]`: 为可比较类型去重

### 转换
- `Select(selector func(T) T) Linq[T]`: 转换每个元素
- `Concat(other Linq[T]) Linq[T]`: 合并两个数据集
- `Reverse() Linq[T]`: 反转顺序

### 排序
- `Sort(compareFn func(a, b T) bool) Linq[T]`: 使用自定义比较排序
- `WithComparer(compare func(a, b T) int) Linq[T]`: 设置用于 Min/Max/Distinct 的比较器

### 聚合
- `Min() (T, bool)`: 获取最小元素（需要比较器）
- `Max() (T, bool)`: 获取最大元素（需要比较器）

### 分页
- `Take(n int) Linq[T]`: 取前 n 个元素
- `Skip(n int) Linq[T]`: 跳过前 n 个元素

### 分组与连接
- `GroupBy[K comparable, T any](l Linq[T], keySelector func(T) K) []Group[K, T]`: 按键分组
- `Join[T, U, K comparable, R any](outer, inner, outerKey, innerKey, resultSelector) Linq[R]`: 连接两个数据集

### 结果提取
- `Results() []T`: 获取最终的切片结果

## 注意事项

- **需要比较器**: `Distinct()`、`Min()` 和 `Max()` 需要通过 `WithComparer()` 设置比较器
- **可比较类型**: 对于内置的可比较类型（int、string等），使用 `DistinctComparable()`
- **不可变性**: 操作返回新的 Linq 实例；原始数据不被修改（底层切片引用除外）
- **性能**: 对于大型数据集，注意长链中的多次内存分配

## 许可证

MIT License
