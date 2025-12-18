# LINQ - LINQ-Style Query API for Go

English | [简体中文](README_zh.md)

`linq` is a generic LINQ-style query library for Go slices, providing a fluent chainable API for filtering, mapping, sorting, grouping, and other common data transformation operations.

## Features

- **Generic Support**: Works with any type using Go generics (Go 1.18+).
- **Chainable API**: Supports method chaining for elegant query expressions.
- **Rich Operations**: Filter, map, sort, group, distinct, take, skip, join, and more.
- **Custom Comparers**: Flexible comparison function support for sorting and deduplication.
- **Zero Dependencies**: Pure Go implementation with no external dependencies.

## Installation

Add the package to your Go project:

```bash
go get github.com/wsshow/op/linq
```

## Usage Examples

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/linq"
)

func main() {
    data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // Filter even numbers and double them
    result := linq.From(data).
        Where(func(x int) bool { return x%2 == 0 }).
        Select(func(x int) int { return x * 2 }).
        Results()
    
    fmt.Println(result) // Output: [4 8 12 16 20]
}
```

### Sorting and Taking Elements

```go
data := []int{5, 2, 8, 1, 9, 3}

// Sort descending and take top 3
top3 := linq.From(data).
    Sort(func(a, b int) bool { return a > b }).
    Take(3).
    Results()

fmt.Println(top3) // Output: [9 8 5]
```

### Deduplication with Comparable Types

```go
data := []int{1, 2, 2, 3, 3, 3, 4}

// Remove duplicates for comparable types
unique := linq.DistinctComparable(linq.From(data)).
    Results()

fmt.Println(unique) // Output: [1 2 3 4]
```

### Deduplication with Custom Comparer

```go
type Person struct {
    Name string
    Age  int
}

people := []Person{
    {"Alice", 25},
    {"Bob", 30},
    {"Alice", 25}, // duplicate
    {"Charlie", 35},
}

// Remove duplicates using custom comparer
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

fmt.Println(len(unique)) // Output: 3
```

### Grouping

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

// Group by category
groups := linq.GroupBy(linq.From(products), func(p Product) string {
    return p.Category
})

for _, group := range groups {
    fmt.Printf("%s: %d items\n", group.Key, len(group.Items))
}
// Output:
// Electronics: 2 items
// Furniture: 2 items
```

### Joining

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

// Join orders with customers
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
    fmt.Printf("Order #%d - %s: $%.2f\n", 
        detail.OrderID, detail.CustomerName, detail.Amount)
}
// Output:
// Order #1 - Alice: $50.00
// Order #2 - Bob: $75.00
// Order #3 - Alice: $100.00
```

### Min/Max with Comparer

```go
data := []int{5, 2, 8, 1, 9, 3}

linq := linq.From(data).WithComparer(func(a, b int) int {
    return a - b
})

min, _ := linq.Min()
max, _ := linq.Max()

fmt.Printf("Min: %d, Max: %d\n", min, max) // Output: Min: 1, Max: 9
```

### Pagination

```go
data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// Skip 3, take 4 (pagination)
page := linq.From(data).
    Skip(3).
    Take(4).
    Results()

fmt.Println(page) // Output: [4 5 6 7]
```

### Checking Conditions

```go
data := []int{2, 4, 6, 8, 10}

// Check if any element is greater than 5
hasLarge := linq.From(data).Any(func(x int) bool { return x > 5 })
fmt.Println(hasLarge) // Output: true
```

### Complex Query Chain

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

// Find top 2 students in grade 10, sorted by score
topStudents := linq.From(students).
    Where(func(s Student) bool { return s.Grade == 10 }).
    Sort(func(a, b Student) bool { return a.Score > b.Score }).
    Take(2).
    Results()

for _, s := range topStudents {
    fmt.Printf("%s: %.1f\n", s.Name, s.Score)
}
// Output:
// Alice: 95.5
// Eve: 85.0
```

## API Overview

### Creation
- `From[T any](data []T) Linq[T]`: Create a Linq instance from a slice

### Filtering
- `Where(predicate func(T) bool) Linq[T]`: Filter elements by predicate
- `Any(predicate func(T) bool) bool`: Check if any element matches
- `Distinct() Linq[T]`: Remove duplicates (requires comparer)
- `DistinctComparable[T comparable](l Linq[T]) Linq[T]`: Remove duplicates for comparable types

### Transformation
- `Select(selector func(T) T) Linq[T]`: Transform each element
- `Concat(other Linq[T]) Linq[T]`: Merge two datasets
- `Reverse() Linq[T]`: Reverse order

### Sorting
- `Sort(compareFn func(a, b T) bool) Linq[T]`: Sort with custom comparison
- `WithComparer(compare func(a, b T) int) Linq[T]`: Set comparer for Min/Max/Distinct

### Aggregation
- `Min() (T, bool)`: Get minimum element (requires comparer)
- `Max() (T, bool)`: Get maximum element (requires comparer)

### Pagination
- `Take(n int) Linq[T]`: Take first n elements
- `Skip(n int) Linq[T]`: Skip first n elements

### Grouping & Joining
- `GroupBy[K comparable, T any](l Linq[T], keySelector func(T) K) []Group[K, T]`: Group by key
- `Join[T, U, K comparable, R any](outer, inner, outerKey, innerKey, resultSelector) Linq[R]`: Join two datasets

### Result Extraction
- `Results() []T`: Get the final slice result

## Notes

- **Comparer Required**: `Distinct()`, `Min()`, and `Max()` require a comparer set via `WithComparer()`.
- **Comparable Types**: Use `DistinctComparable()` for built-in comparable types (int, string, etc.).
- **Immutability**: Operations return new Linq instances; original data is not modified (except the underlying slice reference).
- **Performance**: For large datasets, be mindful of multiple allocations in long chains.

## License

MIT License
