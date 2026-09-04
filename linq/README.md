# LINQ - LINQ-Style Query API for Go

English | [简体中文](README_zh.md)

`linq` is a generic LINQ-style query library for Go slices, providing a fluent chainable API for filtering, mapping, sorting, grouping, and other common data transformation operations.

## Features

- **Generic** — Works with any type (Go 1.24+ module requirement).
- **Chainable** — Method chaining for expressive query pipelines.
- **Rich operations** — Filter, project, sort, group, join, aggregate, and more.
- **Type-safe projection** — `SelectT` maps between different types.
- **Multi-key sorting** — `OrderBy` / `ThenBy` with stable sort.
- **Custom comparers** — Flexible comparison for sorting, deduplication, min/max.
- **Error propagation** — Errors flow through the chain without panicking.
- **Zero dependencies** — Pure Go, only standard library.

## Installation

```bash
go get github.com/wsshow/op/linq
```

## Quick Start

```go
data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// Filter even numbers, double them, keep the first 3
result := linq.From(data).
    Where(func(x int) bool { return x%2 == 0 }).
    Select(func(x int) int { return x * 2 }).
    Take(3).
    Results()

fmt.Println(result) // [4 8 12]
```

## Usage Examples

### Filtering

```go
data := []int{1, 2, 3, 4, 5, 6}

// Keep elements matching a condition
evens := linq.From(data).
    Where(func(x int) bool { return x%2 == 0 }).
    Results()
// [2 4 6]

// Take elements from the start while a condition holds
leading := linq.From(data).
    TakeWhile(func(x int) bool { return x < 4 }).
    Results()
// [1 2 3]

// Skip elements from the start while a condition holds
rest := linq.From(data).
    SkipWhile(func(x int) bool { return x < 4 }).
    Results()
// [4 5 6]
```

### Type Projection (Select / SelectT)

```go
// Same-type transformation
squares := linq.From([]int{1, 2, 3}).
    Select(func(x int) int { return x * x }).
    Results()
// [1 4 9]

// Cross-type transformation — use SelectT
labels := linq.SelectT(
    linq.From([]int{1, 2, 3}),
    func(x int) string { return fmt.Sprintf("Item #%d", x) },
).Results()
// ["Item #1" "Item #2" "Item #3"]

// Flatten nested slices with SelectMany
nested := linq.From([][]int{{1, 2}, {3, 4}})
flat := linq.SelectMany(nested, func(x []int) []int { return x }).Results()
// [1 2 3 4]
```

### Sorting

```go
// Basic sort with a less function
sorted := linq.From([]int{3, 1, 4, 2}).
    Sort(func(a, b int) bool { return a < b }).
    Results()
// [1 2 3 4]

// Sort by a key (ascending)
type Person struct{ Name string; Age int }
people := []Person{{"Bob", 30}, {"Alice", 25}, {"Charlie", 35}}

sorted := linq.OrderBy(linq.From(people), func(p Person) string { return p.Name }).Results()
// [{Alice 25} {Bob 30} {Charlie 35}]

// Sort descending
sorted = linq.OrderByDescending(linq.From(people), func(p Person) int { return p.Age }).Results()
// [{Charlie 35} {Bob 30} {Alice 25}]

// Multi-key: OrderBy + ThenBy
sorted = linq.OrderBy(linq.From(people), func(p Person) string { return p.Name }).
    ThenBy(func(a, b Person) int { return cmp.Compare(a.Age, b.Age) }).
    Results()
// [{Alice 25} {Bob 30} {Charlie 35}] — by Name, then Age
```

### Deduplication

```go
// For comparable types (int, string, etc.)
unique := linq.DistinctComparable(linq.From([]int{1, 2, 2, 3, 1})).Results()
// [1 2 3]

// Deduplicate by a key
type User struct{ ID int; Name string }
users := []User{{1, "Alice"}, {2, "Bob"}, {1, "Alice (dup)"}}
uniqueUsers := linq.DistinctBy(linq.From(users), func(u User) int { return u.ID }).Results()
// [{1 Alice} {2 Bob}] — keeps first occurrence per key

// Custom comparison (case-insensitive, non-comparable structs, etc.)
uniqueCase := linq.From([]string{"b", "A", "a", "c"}).
    WithComparer(func(a, b string) int {
        return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
    }).
    Distinct().
    Results()
// ["b" "A" "c"] — preserves the first occurrence from the input
```

### Aggregation

```go
data := linq.From([]int{1, 2, 3, 4, 5})

count := data.Count()                         // 5
evenCount := data.CountBy(func(x int) bool { return x%2 == 0 }) // 2
sum := linq.Sum(data)                         // 15
avg := linq.Average(linq.From([]float64{1, 2, 3})) // 2.0

// MinVal / MaxVal -- for cmp.Ordered types, returns (value, bool)
minVal, ok := linq.MinVal(data) // 1, true
maxVal, ok := linq.MaxVal(data) // 5, true

// Min / Max — requires WithComparer, returns (value, bool)
nums := linq.From([]int{5, 2, 8, 1, 3}).WithComparer(cmp.Compare[int])
min, ok := nums.Min() // 1, true
max, ok := nums.Max() // 8, true

// MinBy / MaxBy — no comparer needed, returns (value, bool)
type Product struct{ Name string; Price float64 }
products := []Product{{"A", 9.9}, {"B", 5.0}, {"C", 12.0}}
cheapest, ok := linq.MinBy(linq.From(products), func(p Product) float64 { return p.Price })
// {B 5.0}, true

// Fold / Reduce
result := linq.Aggregate(
    linq.From([]int{1, 2, 3, 4}),
    0,
    func(acc, x int) int { return acc + x },
) // 10
```

### Element Access

```go
data := linq.From([]int{10, 20, 30, 40})

first, ok := data.First()           // 10, true
first, ok = data.FirstBy(func(x int) bool { return x > 15 }) // 20, true
last, ok := data.Last()             // 40, true
at, ok := data.ElementAt(2)         // 30, true

// Single — asserts exactly one element
only, ok := linq.From([]int{42}).Single()       // 42, true
_, ok = linq.From([]int{1, 2}).Single()          // false (multiple)
_, ok = linq.From([]int{1, 2, 3}).SingleBy(func(x int) bool { return x > 5 }) // false (none)

// Check conditions
hasLarge := data.Any(func(x int) bool { return x > 25 }) // true
allEven := data.All(func(x int) bool { return x%2 == 0 }) // false
has := linq.Contains(data, 20) // true
```

### Pagination

```go
data := linq.From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

// Take first 3
data.Take(3).Results() // [1 2 3]

// Skip 5, take 3 (page 2 of page-size 5)
data.Skip(5).Take(3).Results() // [6 7 8]

// Split into chunks
chunks := data.Chunk(3) // [[1 2 3] [4 5 6] [7 8 9] [10]]
```

### Set Operations

```go
a := linq.From([]int{1, 2, 3})
b := linq.From([]int{2, 3, 4})

linq.Union(a, b).Results()      // [1 2 3 4]
linq.Intersect(a, b).Results()  // [2 3]
linq.Except(a, b).Results()     // [1]

// Element-wise equality
linq.SequenceEqual(
    linq.From([]int{1, 2, 3}),
    linq.From([]int{1, 2, 3}),
) // true
```

### Grouping & Joining

```go
// Group by category
products := []struct{ Name, Category string }{
    {"Laptop", "Electronics"}, {"Mouse", "Electronics"},
    {"Desk", "Furniture"}, {"Chair", "Furniture"},
}
groups := linq.GroupBy(linq.From(products), func(p struct{ Name, Category string }) string {
    return p.Category
})
for _, g := range groups {
    fmt.Printf("%s: %d items\n", g.Key, len(g.Items))
}
// Electronics: 2 items
// Furniture: 2 items

// Inner join
type Order struct{ ID, CustomerID int }
type Customer struct{ ID int; Name string }

orders := linq.From([]Order{{1, 101}, {2, 102}, {3, 101}})
customers := linq.From([]Customer{{101, "Alice"}, {102, "Bob"}})

result := linq.Join(orders, customers,
    func(o Order) int { return o.CustomerID },
    func(c Customer) int { return c.ID },
    func(o Order, c Customer) string {
        return fmt.Sprintf("Order #%d by %s", o.ID, c.Name)
    },
).Results()
// ["Order #1 by Alice" "Order #2 by Bob" "Order #3 by Alice"]

// Left join — includes unmatched outer elements
result = linq.LeftJoin(orders, customers,
    func(o Order) int { return o.CustomerID },
    func(c Customer) int { return c.ID },
    func(o Order, c Customer) string { return fmt.Sprintf("#%d: %s", o.ID, c.Name) },
    func(o Order) string { return fmt.Sprintf("#%d: (no customer)", o.ID) },
).Results()

// Zip — pair elements positionally
zipped := linq.Zip(
    linq.From([]int{1, 2, 3}),
    linq.From([]string{"a", "b", "c"}),
    func(x int, s string) string { return fmt.Sprintf("%d%s", x, s) },
).Results()
// ["1a" "2b" "3c"]
```

### Building Sequences

```go
linq.Range(5, 4).Results()         // [5 6 7 8]
linq.Repeat("hi", 3).Results()     // ["hi" "hi" "hi"]
linq.Empty[int]().Count()          // 0
```

### Error Handling

```go
// Operations that need a comparer set an error if it's missing
l := linq.From([]int{1, 2, 3}).Distinct()
if err := l.Error(); err != nil {
    fmt.Println("Error:", err) // "linq.Distinct: requires a comparer, use WithComparer"
}

// Min/Max return (T, bool), false means empty or missing comparer
_, ok := linq.From([]int{1, 2}).Min()
if !ok {
    fmt.Println("Missing comparer or empty")
}

// Errors propagate through the chain — check at the end
result := linq.From(data).
    WithComparer(func(a, b MyType) int { ... }).
    Distinct().
    Where(...).
    Results()
if err := linq.From(data).WithComparer(...).Distinct().Error(); err != nil {
    // handle
}
```

## API Overview

### Creation
| Function | Description |
|----------|-------------|
| `From[T](data []T) Linq[T]` | Create from a slice |
| `Empty[T]() Linq[T]` | Create an empty instance |
| `Range(start, count int) Linq[int]` | Generate consecutive integers |
| `Repeat[T](value T, count int) Linq[T]` | Generate repeated values |

### Filtering
| Function | Description |
|----------|-------------|
| `Where(predicate) Linq[T]` | Keep elements matching predicate |
| `Distinct() Linq[T]` | Remove duplicates (needs WithComparer) |
| `DistinctComparable[T comparable](l) Linq[T]` | Remove duplicates for `==` types |
| `DistinctBy[T, K comparable](l, keyFn) Linq[T]` | Remove duplicates by key |
| `TakeWhile(predicate) Linq[T]` | Take from start while true |
| `SkipWhile(predicate) Linq[T]` | Skip from start while true |

### Projection
| Function | Description |
|----------|-------------|
| `Select(func(T) T) Linq[T]` | Transform same-type |
| `SelectT[T, R any](l, func(T) R) Linq[R]` | Transform to different type |
| `SelectMany[T, R any](l, func(T) []R) Linq[R]` | Flatten nested slices |

### Sorting
| Function | Description |
|----------|-------------|
| `Sort(less func(T,T) bool) Linq[T]` | Sort with less function |
| `WithComparer(cmp func(T,T) int) Linq[T]` | Set comparer for Distinct/Min/Max |
| `OrderBy[T, K ordered](l, keyFn) OrderedLinq[T]` | Sort ascending by key |
| `OrderByDescending[T, K ordered](l, keyFn) OrderedLinq[T]` | Sort descending by key |
| `ThenBy(cmp func(T,T) int) OrderedLinq[T]` | Add tie-breaker after OrderBy |

### Aggregation
| Function | Description |
|----------|-------------|
| `Count() int` | Element count |
| `CountBy(predicate) int` | Count matches |
| `Sum[T Numeric](l) T` | Sum of numeric values |
| `Average[T Numeric](l) float64` | Average of numeric values |
| `Min() (T, bool)` | Minimum (needs WithComparer) |
| `Max() (T, bool)` | Maximum (needs WithComparer) |
| `MinBy[T, K ordered](l, keyFn) (T, bool)` | Element with min key |
| `MaxBy[T, K ordered](l, keyFn) (T, bool)` | Element with max key |
| `MinVal[T ordered](l) (T, bool)` | Min for ordered types |
| `MaxVal[T ordered](l) (T, bool)` | Max for ordered types |
| `Aggregate[T, R any](l, seed, fn) R` | Fold / reduce |

### Element Access
| Function | Description |
|----------|-------------|
| `First() (T, bool)` | First element |
| `FirstBy(predicate) (T, bool)` | First matching |
| `Last() (T, bool)` | Last element |
| `LastBy(predicate) (T, bool)` | Last matching |
| `ElementAt(index) (T, bool)` | Element at index |
| `Single() (T, bool)` | Exactly one element |
| `SingleBy(predicate) (T, bool)` | Exactly one matching |
| `Contains[T comparable](l, v) bool` | Check membership |
| `Any(predicate) bool` | Any match |
| `All(predicate) bool` | All match |

### Partitioning
| Function | Description |
|----------|-------------|
| `Take(n int) Linq[T]` | First n elements |
| `Skip(n int) Linq[T]` | Skip first n |
| `Chunk(size int) [][]T` | Split into chunks |

### Set Operations
| Function | Description |
|----------|-------------|
| `Union[T comparable](l1, l2) Linq[T]` | Union (distinct) |
| `Intersect[T comparable](l1, l2) Linq[T]` | Intersection |
| `Except[T comparable](l1, l2) Linq[T]` | Difference |
| `SequenceEqual[T comparable](l1, l2) bool` | Element-wise equality |

### Grouping & Joining
| Function | Description |
|----------|-------------|
| `GroupBy[K, T](l, keyFn) []Group[K, T]` | Group by key |
| `Join[T, U, K, R]` | Inner join |
| `LeftJoin[T, U, K, R]` | Left outer join |
| `Zip[T, U, R](l1, l2, fn) Linq[R]` | Pair-wise combine |

### Conversion & Execution
| Function | Description |
|----------|-------------|
| `Results() []T` | Get underlying slice (not a copy) |
| `ToSlice() []T` | Get a safe copy |
| `ToMap[T, K, V](l, kFn, vFn) map[K]V` | Convert to map |
| `ForEach(action func(T))` | Execute action per element |
| `DefaultIfEmpty(v T) Linq[T]` | Default value if empty |
| `Error() error` | Get first error in chain |
| `Concat(other) Linq[T]` | Concatenate sequences |
| `Append(elems ...T) Linq[T]` | Append elements |
| `Prepend(elems ...T) Linq[T]` | Prepend elements |
| `Reverse() Linq[T]` | Reverse order |

## Notes

- **Error handling**: Methods return new `Linq` instances with errors set; terminal functions return `(T, bool)`. Always check errors at the end of a chain.
- **Comparer required**: `Distinct()`, `Min()`, and `Max()` need a comparer via `WithComparer()`. For built-in types, use `DistinctComparable()` or `MinBy`/`MaxBy` instead.
- **Immutability**: Operations return new instances and do not modify original data.
- **Min() and Max() return (T, bool), consistent with MinBy/MaxBy/MinVal/MaxVal.

## License

MIT
