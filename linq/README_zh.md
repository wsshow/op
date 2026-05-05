# LINQ - Go 的 LINQ 风格查询 API

[English](./README.md) | 简体中文

`linq` 是一个针对 Go 切片的泛型 LINQ 风格查询库，提供流畅的链式 API 来执行过滤、映射、排序、分组等常见数据转换操作。

## 特性

- **泛型** — 支持任意类型（Go 1.22+）。
- **链式调用** — 方法链串联出清晰的查询管线。
- **操作丰富** — 过滤、投影、排序、分组、连接、聚合等。
- **类型安全投影** — `SelectT` 可在不同类型间映射。
- **多级排序** — `OrderBy` / `ThenBy` 搭配稳定排序。
- **自定义比较器** — 灵活的排序、去重、最值比较。
- **错误传播** — 错误沿链传播，不抛异常。
- **零依赖** — 纯 Go 标准库实现。

## 安装

```bash
go get github.com/wsshow/op/linq
```

## 快速入门

```go
data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// 取偶数、翻倍、只取前 3 个
result := linq.From(data).
    Where(func(x int) bool { return x%2 == 0 }).
    Select(func(x int) int { return x * 2 }).
    Take(3).
    Results()

fmt.Println(result) // [4 8 12]
```

## 使用示例

### 过滤

```go
data := []int{1, 2, 3, 4, 5, 6}

// 保留偶数
evens := linq.From(data).
    Where(func(x int) bool { return x%2 == 0 }).
    Results()
// [2 4 6]

// 从头开始取满足条件的元素
leading := linq.From(data).
    TakeWhile(func(x int) bool { return x < 4 }).
    Results()
// [1 2 3]

// 从头跳过满足条件的元素，返回剩余
rest := linq.From(data).
    SkipWhile(func(x int) bool { return x < 4 }).
    Results()
// [4 5 6]
```

### 类型投影（Select / SelectT）

```go
// 同类型变换——用 Select 方法
squares := linq.From([]int{1, 2, 3}).
    Select(func(x int) int { return x * x }).
    Results()
// [1 4 9]

// 跨类型变换——用 SelectT 函数
labels := linq.SelectT(
    linq.From([]int{1, 2, 3}),
    func(x int) string { return fmt.Sprintf("第%d项", x) },
).Results()
// ["第1项" "第2项" "第3项"]

// 扁平化嵌套切片——SelectMany
nested := linq.From([][]int{{1, 2}, {3, 4}})
flat := linq.SelectMany(nested, func(x []int) []int { return x }).Results()
// [1 2 3 4]
```

### 排序

```go
// 基本排序——传入 less 函数
sorted := linq.From([]int{3, 1, 4, 2}).
    Sort(func(a, b int) bool { return a < b }).
    Results()
// [1 2 3 4]

// 按键升序
type Person struct{ Name string; Age int }
people := []Person{{"Bob", 30}, {"Alice", 25}, {"Charlie", 35}}

sorted := linq.OrderBy(linq.From(people), func(p Person) string { return p.Name }).Results()
// [{Alice 25} {Bob 30} {Charlie 35}]

// 按键降序
sorted = linq.OrderByDescending(linq.From(people), func(p Person) int { return p.Age }).Results()
// [{Charlie 35} {Bob 30} {Alice 25}]

// 多级排序：OrderBy + ThenBy
sorted = linq.OrderBy(linq.From(people), func(p Person) string { return p.Name }).
    ThenBy(func(a, b Person) int { return cmp.Compare(a.Age, b.Age) }).
    Results()
// [{Alice 25} {Bob 30} {Charlie 35}] — 先按 Name，再按 Age
```

### 去重

```go
// 可比较类型（int、string 等）
unique := linq.DistinctComparable(linq.From([]int{1, 2, 2, 3, 1})).Results()
// [1 2 3]

// 按键去重
type User struct{ ID int; Name string }
users := []User{{1, "Alice"}, {2, "Bob"}, {1, "Alice(重复)"}}
uniqueUsers := linq.DistinctBy(linq.From(users), func(u User) int { return u.ID }).Results()
// [{1 Alice} {2 Bob}] — 每个键保留首次出现的元素

// 自定义比较（忽略大小写、不可比较的结构体等）
uniqueCase := linq.From([]string{"a", "A", "b"}).
    WithComparer(func(a, b string) int {
        return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
    }).
    Distinct().
    Results()
// ["a" "b"]（或 ["A" "b"]）
```

### 聚合

```go
data := linq.From([]int{1, 2, 3, 4, 5})

count := data.Count()                                      // 5
evenCount := data.CountBy(func(x int) bool { return x%2 == 0 }) // 2
sum := linq.Sum(data)                                      // 15
avg := linq.Average(linq.From([]float64{1, 2, 3})) // 2.0

// 最小值 / 最大值——需 WithComparer，返回 (值, bool)
nums := linq.From([]int{5, 2, 8, 1, 3}).WithComparer(func(a, b int) int { return a - b })
min, ok := nums.Min() // 1, true
max, ok := nums.Max() // 8, true

// MinBy / MaxBy——无需比较器，按 key 取最值，返回 (元素, 是否找到)
type Product struct{ Name string; Price float64 }
products := []Product{{"A", 9.9}, {"B", 5.0}, {"C", 12.0}}
cheapest, ok := linq.MinBy(linq.From(products), func(p Product) float64 { return p.Price })
// {B 5.0}, true

// 累积（Fold / Reduce）
result := linq.Aggregate(
    linq.From([]int{1, 2, 3, 4}),
    0,
    func(acc, x int) int { return acc + x },
) // 10
```

### 元素访问

```go
data := linq.From([]int{10, 20, 30, 40})

first, ok := data.First()           // 10, true
first, ok = data.FirstBy(func(x int) bool { return x > 15 }) // 20, true
last, ok := data.Last()             // 40, true
at, ok := data.ElementAt(2)         // 30, true

// Single——断言恰有一个元素
only, ok := linq.From([]int{42}).Single()   // 42, true
_, ok = linq.From([]int{1, 2}).Single()      // false（多个元素）
_, ok = linq.From([]int{1, 2, 3}).SingleBy(func(x int) bool { return x > 5 }) // false（无匹配）

// 条件检查
hasLarge := data.Any(func(x int) bool { return x > 25 })  // true
allEven := data.All(func(x int) bool { return x%2 == 0 }) // false
has := linq.Contains(data, 20) // true
```

### 分页

```go
data := linq.From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

// 取前 3 个
data.Take(3).Results() // [1 2 3]

// 跳过 5 个再取 3 个（相当于第 2 页，每页 5 条）
data.Skip(5).Take(3).Results() // [6 7 8]

// 分块
chunks := data.Chunk(3) // [[1 2 3] [4 5 6] [7 8 9] [10]]
```

### 集合运算

```go
a := linq.From([]int{1, 2, 3})
b := linq.From([]int{2, 3, 4})

linq.Union(a, b).Results()      // [1 2 3 4]
linq.Intersect(a, b).Results()  // [2 3]
linq.Except(a, b).Results()     // [1]

// 逐元素相等判断
linq.SequenceEqual(
    linq.From([]int{1, 2, 3}),
    linq.From([]int{1, 2, 3}),
) // true
```

### 分组与连接

```go
// 按类别分组
products := []struct{ Name, Category string }{
    {"笔记本", "电子产品"}, {"鼠标", "电子产品"},
    {"桌子", "家具"}, {"椅子", "家具"},
}
groups := linq.GroupBy(linq.From(products), func(p struct{ Name, Category string }) string {
    return p.Category
})
for _, g := range groups {
    fmt.Printf("%s: %d 件\n", g.Key, len(g.Items))
}
// 电子产品: 2 件
// 家具: 2 件

// 内连接
type Order struct{ ID, CustomerID int }
type Customer struct{ ID int; Name string }

orders := linq.From([]Order{{1, 101}, {2, 102}, {3, 101}})
customers := linq.From([]Customer{{101, "Alice"}, {102, "Bob"}})

result := linq.Join(orders, customers,
    func(o Order) int { return o.CustomerID },
    func(c Customer) int { return c.ID },
    func(o Order, c Customer) string {
        return fmt.Sprintf("订单#%d 客户:%s", o.ID, c.Name)
    },
).Results()
// ["订单#1 客户:Alice" "订单#2 客户:Bob" "订单#3 客户:Alice"]

// 左外连接——保留未匹配的 outer 元素
result = linq.LeftJoin(orders, customers,
    func(o Order) int { return o.CustomerID },
    func(c Customer) int { return c.ID },
    func(o Order, c Customer) string { return fmt.Sprintf("#%d: %s", o.ID, c.Name) },
    func(o Order) string { return fmt.Sprintf("#%d: (无客户)", o.ID) },
).Results()

// Zip——按位置配对
zipped := linq.Zip(
    linq.From([]int{1, 2, 3}),
    linq.From([]string{"a", "b", "c"}),
    func(x int, s string) string { return fmt.Sprintf("%d%s", x, s) },
).Results()
// ["1a" "2b" "3c"]
```

### 生成序列

```go
linq.Range(5, 4).Results()      // [5 6 7 8]
linq.Repeat("hi", 3).Results()  // ["hi" "hi" "hi"]
linq.Empty[int]().Count()       // 0
```

### 错误处理

```go
// 缺少比较器时，Distinct 将错误记录在 Linq 实例中
l := linq.From([]int{1, 2, 3}).Distinct()
if err := l.Error(); err != nil {
    fmt.Println("错误:", err) // "linq.Distinct: requires a comparer, use WithComparer"
}

// Min/Max 返回 (T, bool)，false 表示空序列或缺少比较器
_, ok := linq.From([]int{1, 2}).Min()
if !ok {
    fmt.Println("缺少比较器或序列为空")
}

// 错误沿链传播——在链末检查即可
result := linq.From(data).
    WithComparer(func(a, b MyType) int { ... }).
    Distinct().
    Where(...).
    Results()
if err := linq.From(data).WithComparer(...).Distinct().Error(); err != nil {
    // 处理错误
}
```

## API 概览

### 创建
| 函数 | 说明 |
|------|------|
| `From[T](data []T) Linq[T]` | 从切片创建 |
| `Empty[T]() Linq[T]` | 创建空实例 |
| `Range(start, count int) Linq[int]` | 生成连续整数 |
| `Repeat[T](value T, count int) Linq[T]` | 生成重复值 |

### 过滤
| 函数 | 说明 |
|------|------|
| `Where(predicate) Linq[T]` | 保留满足条件的元素 |
| `Distinct() Linq[T]` | 去重（需 WithComparer） |
| `DistinctComparable[T comparable](l) Linq[T]` | 对可比较类型去重 |
| `DistinctBy[T, K comparable](l, keyFn) Linq[T]` | 按键去重 |
| `TakeWhile(predicate) Linq[T]` | 从头取满足条件的元素 |
| `SkipWhile(predicate) Linq[T]` | 从头跳过满足条件的元素 |

### 投影
| 函数 | 说明 |
|------|------|
| `Select(func(T) T) Linq[T]` | 同类型变换 |
| `SelectT[T, R any](l, func(T) R) Linq[R]` | 跨类型变换 |
| `SelectMany[T, R any](l, func(T) []R) Linq[R]` | 扁平化映射 |

### 排序
| 函数 | 说明 |
|------|------|
| `Sort(less func(T,T) bool) Linq[T]` | 用 less 函数排序 |
| `WithComparer(cmp func(T,T) int) Linq[T]` | 设置三值比较器（供 Distinct/Min/Max 使用） |
| `OrderBy[T, K ordered](l, keyFn) OrderedLinq[T]` | 按键升序 |
| `OrderByDescending[T, K ordered](l, keyFn) OrderedLinq[T]` | 按键降序 |
| `ThenBy(cmp func(T,T) int) OrderedLinq[T]` | 在 OrderBy 后追加次级排序 |

### 聚合
| 函数 | 说明 |
|------|------|
| `Count() int` | 元素数量 |
| `CountBy(predicate) int` | 满足条件的数量 |
| `Sum[T Numeric](l) T` | 数值求和 |
| `Average[T Numeric](l) float64` | 数值平均值 |
| `Min() (T, bool)` | 最小值（需 WithComparer） |
| `Max() (T, bool)` | 最大值（需 WithComparer） |
| `MinBy[T, K ordered](l, keyFn) (T, bool)` | 按 key 取最小元素 |
| `MaxBy[T, K ordered](l, keyFn) (T, bool)` | 按 key 取最大元素 |
| `MinVal[T ordered](l) (T, bool)` | ordered 类型最小值 |
| `MaxVal[T ordered](l) (T, bool)` | ordered 类型最大值 |
| `Aggregate[T, R any](l, seed, fn) R` | 累积运算（折叠） |

### 元素访问
| 函数 | 说明 |
|------|------|
| `First() (T, bool)` | 第一个元素 |
| `FirstBy(predicate) (T, bool)` | 第一个匹配的元素 |
| `Last() (T, bool)` | 最后一个元素 |
| `LastBy(predicate) (T, bool)` | 最后一个匹配的元素 |
| `ElementAt(index) (T, bool)` | 指定索引的元素 |
| `Single() (T, bool)` | 唯一的元素 |
| `SingleBy(predicate) (T, bool)` | 唯一匹配的元素 |
| `Contains[T comparable](l, v) bool` | 是否包含 |
| `Any(predicate) bool` | 是否存在 |
| `All(predicate) bool` | 是否全部满足 |

### 分页
| 函数 | 说明 |
|------|------|
| `Take(n int) Linq[T]` | 取前 n 个 |
| `Skip(n int) Linq[T]` | 跳过前 n 个 |
| `Chunk(size int) [][]T` | 分块 |

### 集合运算
| 函数 | 说明 |
|------|------|
| `Union[T comparable](l1, l2) Linq[T]` | 并集（去重） |
| `Intersect[T comparable](l1, l2) Linq[T]` | 交集 |
| `Except[T comparable](l1, l2) Linq[T]` | 差集 |
| `SequenceEqual[T comparable](l1, l2) bool` | 逐元素相等判断 |

### 分组与连接
| 函数 | 说明 |
|------|------|
| `GroupBy[K, T](l, keyFn) []Group[K, T]` | 按键分组 |
| `Join[T, U, K, R]` | 内连接 |
| `LeftJoin[T, U, K, R]` | 左外连接 |
| `Zip[T, U, R](l1, l2, fn) Linq[R]` | 按位置配对 |

### 转换与执行
| 函数 | 说明 |
|------|------|
| `Results() []T` | 获取底层切片（非副本） |
| `ToSlice() []T` | 获取安全副本 |
| `ToMap[T, K, V](l, kFn, vFn) map[K]V` | 转换为 map |
| `ForEach(action func(T))` | 对每个元素执行操作 |
| `DefaultIfEmpty(v T) Linq[T]` | 空序列时使用默认值 |
| `Error() error` | 获取链中第一个错误 |
| `Concat(other) Linq[T]` | 拼接序列 |
| `Append(elems ...T) Linq[T]` | 末尾追加 |
| `Prepend(elems ...T) Linq[T]` | 开头插入 |
| `Reverse() Linq[T]` | 反转顺序 |

## 注意事项

- **错误处理**：方法返回携带错误的 `Linq` 实例；终端函数返回 `(T, bool)`。建议在链末检查错误。
- **需要比较器**：`Distinct()`、`Min()`、`Max()` 需通过 `WithComparer()` 设置比较器。内置类型推荐使用 `DistinctComparable()` 或 `MinBy`/`MaxBy`。
- **不可变性**：操作返回新实例，不修改原始数据。
- **v1 起 Min() 和 Max() 返回 (T, bool)，与 MinBy/MaxBy/MinVal/MaxVal 保持一致。

## 许可证

MIT
