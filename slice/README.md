# Slice - A Generic Slice Wrapper in Go

English | [简体中文](README_zh.md)

`slice` is a generic slice wrapper in Go that provides a rich set of utility methods for common slice operations, inspired by JavaScript array methods and functional programming patterns.

## Features

- **Generic Support**: Works with any type using Go generics (Go 1.24+ module requirement).
- **Chainable API**: Most methods return `*Slice[T]` for method chaining.
- **Rich Operations**: Push, pop, shift, unshift, insert, remove, filter, map, reduce, sort, and more.
- **Familiar Syntax**: API inspired by JavaScript arrays for easy adoption.
- **Type-Safe**: Full compile-time type checking with generics.

## Design Principles

- **Container mutations** (Push, Pop, Shift, Unshift, Insert, Remove, Set, Clear): operate in-place, return self for chaining.
- **In-place transforms** (Sort, Reverse): modify in-place, consistent with Go's standard library.
- **Immutable transforms** (Map, Filter, Concat, Sub, Clone): return a new `*Slice[T]`, leaving the original unchanged.
- **Queries** (Find, FindIndex, First, Last, Every, Some, Reduce, etc.): return computed values.

## Installation

```bash
go get github.com/wsshow/op/slice
```

## Usage Examples

### Creating and Basic Operations

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/slice"
)

func main() {
    s := slice.New(1, 2, 3)

    s.Push(4, 5)
    fmt.Println(s.Data()) // Output: [1 2 3 4 5]

    last, ok := s.Pop()
    fmt.Println(last, ok)  // Output: 5 true
    fmt.Println(s.Data())  // Output: [1 2 3 4]
}
```

### Array-like Operations

```go
s := slice.New(1, 2, 3)

s.Unshift(0)
fmt.Println(s.Data()) // Output: [0 1 2 3]

first, ok := s.Shift()
fmt.Println(first, ok) // Output: 0 true
fmt.Println(s.Data())  // Output: [1 2 3]
```

### Insert and Remove

```go
s := slice.New(1, 3).Insert(1, 2)
fmt.Println(s.Data()) // Output: [1 2 3]

val, ok := s.Remove(1)
fmt.Println(val, ok)  // Output: 2 true
fmt.Println(s.Data()) // Output: [1 3]
```

### Filtering and Mapping

```go
numbers := slice.New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

evens := numbers.Filter(func(x int) bool {
    return x%2 == 0
})
fmt.Println(evens.Data()) // Output: [2 4 6 8 10]
// Original unchanged
fmt.Println(numbers.Data()) // Output: [1 2 3 4 5 6 7 8 9 10]

doubled := numbers.Map(func(x int) int {
    return x * 2
})
fmt.Println(doubled.Data())  // Output: [2 4 6 8 10 12 14 16 18 20]
fmt.Println(numbers.Data())  // Output: [1 2 3 4 5 6 7 8 9 10] (unchanged)
```

### Type-Changing Map

```go
s := slice.New(1, 2, 3)
strs := slice.MapTo(s, func(v int) string {
    return fmt.Sprintf("n=%d", v)
})
fmt.Println(strs.Data()) // Output: [n=1 n=2 n=3]
```

### Searching

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
    fmt.Println(user.Name) // Output: Bob
}

// Find by index
idx := users.FindIndex(func(u struct{ Name string; Age int }) bool {
    return u.Age > 28
})
fmt.Println(idx) // Output: 1
```

### IndexOf and Contains (Comparable Types)

```go
names := slice.New("Alice", "Bob", "Charlie")

index := slice.IndexOf(names, "Bob")
fmt.Println(index) // Output: 1

index = slice.IndexOf(names, "David")
fmt.Println(index) // Output: -1

fmt.Println(slice.Contains(names, "Bob"))   // Output: true
fmt.Println(slice.Contains(names, "David")) // Output: false
```

### First and Last

```go
s := slice.New(10, 20, 30)

first, ok := s.First()
fmt.Println(first, ok) // Output: 10 true

last, ok := s.Last()
fmt.Println(last, ok) // Output: 30 true
```

### Checking Conditions

```go
numbers := slice.New(2, 4, 6, 8, 10)

allEven := numbers.Every(func(x int) bool {
    return x%2 == 0
})
fmt.Println(allEven) // Output: true

someGreater := numbers.Some(func(x int) bool {
    return x > 5
})
fmt.Println(someGreater) // Output: true
```

### Reducing

```go
numbers := slice.New(1, 2, 3, 4, 5)

sum := numbers.Reduce(func(acc, curr int) int {
    return acc + curr
}, 0)
fmt.Println(sum) // Output: 15

// Reduce with different accumulator type
total := slice.ReduceTo(numbers, func(acc string, v int) string {
    return acc + fmt.Sprint(v)
}, "")
fmt.Println(total) // Output: 12345
```

### Sorting and Reversing

```go
numbers := slice.New(5, 2, 8, 1, 9, 3)

numbers.Sort(func(a, b int) bool {
    return a < b
})
fmt.Println(numbers.Data()) // Output: [1 2 3 5 8 9]

numbers.Reverse()
fmt.Println(numbers.Data()) // Output: [9 8 5 3 2 1]
```

### Concatenation

```go
s1 := slice.New(1, 2)
s2 := slice.New(3, 4)
s3 := slice.New(5, 6)

combined := s1.Concat(s2, s3)
fmt.Println(combined.Data()) // Output: [1 2 3 4 5 6]
```

### Sub-slicing

```go
s := slice.New(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

sub := s.Sub(2, 5)
fmt.Println(sub.Data()) // Output: [2 3 4]

// Original unchanged
fmt.Println(s.Length()) // Output: 10
```

### Get, Set, and Iteration

```go
s := slice.New(10, 20, 30, 40, 50)

value, ok := s.Get(2)
if ok {
    fmt.Println(value) // Output: 30
}

success := s.Set(3, 99)
fmt.Println(success)   // Output: true
fmt.Println(s.Data())  // Output: [10 20 30 99 50]

// Iteration
s.ForEach(func(v int) {
    fmt.Println(v)
})

// Iteration with index
s.ForEachIndex(func(i, v int) {
    fmt.Printf("[%d] = %d\n", i, v)
})
```

### Clearing and Cloning

```go
s := slice.New(1, 2, 3, 4, 5)

clone := s.Clone()
fmt.Println(clone.Data()) // Output: [1 2 3 4 5]

s.Clear()
fmt.Println(s.Length())   // Output: 0
fmt.Println(s.IsEmpty())  // Output: true

fmt.Println(clone.Data()) // Output: [1 2 3 4 5]
```

### Data and Raw Access

```go
s := slice.New(1, 2, 3)

// Data returns a safe copy
data := s.Data()
data[0] = 10
fmt.Println(s.Data()) // Output: [1 2 3] (unchanged)

// Raw returns the underlying slice directly (no copy)
raw := s.Raw()
fmt.Println(raw) // Output: [1 2 3]
```

### Complex Example: Data Processing Pipeline

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
    fmt.Printf("%s: $%.2f (Stock: %d)\n", p.Name, p.Price, p.Stock)
})
// Output:
// Laptop: $1099.99 (Stock: 5)
// Monitor: $329.99 (Stock: 10)
```

## API Overview

### Creation
- `New[T any](values ...T) *Slice[T]`: Create a new slice with initial values

### Container Mutations (in-place, return self)

| Method | Description |
|--------|-------------|
| `Push(values ...T) *Slice[T]` | Add elements to the end |
| `Pop() (T, bool)` | Remove and return last element |
| `Unshift(values ...T) *Slice[T]` | Add elements to the beginning |
| `Shift() (T, bool)` | Remove and return first element |
| `Insert(index int, values ...T) *Slice[T]` | Insert elements at index |
| `Remove(index int) (T, bool)` | Remove element at index |
| `Set(index int, value T) bool` | Set element at index |
| `Clear() *Slice[T]` | Remove all elements |

### In-place Transforms (mutate, return self)

| Method | Description |
|--------|-------------|
| `Sort(less func(a, b T) bool) *Slice[T]` | Sort elements |
| `Reverse() *Slice[T]` | Reverse order |

### Immutable Transforms (return new Slice)

| Method | Description |
|--------|-------------|
| `Map(fn func(T) T) *Slice[T]` | Transform each element |
| `Filter(predicate func(T) bool) *Slice[T]` | Filter elements |
| `Concat(others ...*Slice[T]) *Slice[T]` | Concatenate slices |
| `Sub(start, end int) *Slice[T]` | Get sub-slice |
| `Clone() *Slice[T]` | Copy the slice storage (referenced values remain shared) |

### Package-Level Functions

| Function | Description |
|----------|-------------|
| `IndexOf[T comparable](s *Slice[T], value T) int` | Find index of value |
| `Contains[T comparable](s *Slice[T], value T) bool` | Check if value exists |
| `MapTo[T, U any](s *Slice[T], fn func(T) U) *Slice[U]` | Map with type change |
| `ReduceTo[T, U any](s *Slice[T], fn func(U, T) U, initial U) U` | Reduce with different accumulator type |

### Queries

| Method | Description |
|--------|-------------|
| `Length() int` | Number of elements |
| `IsEmpty() bool` | Check if empty |
| `Get(index int) (T, bool)` | Get element at index |
| `Find(predicate func(T) bool) (T, bool)` | Find first matching element |
| `FindIndex(predicate func(T) bool) int` | Find index of first match |
| `First() (T, bool)` | Get first element |
| `Last() (T, bool)` | Get last element |
| `Every(predicate func(T) bool) bool` | Check if all match |
| `Some(predicate func(T) bool) bool` | Check if any match |
| `Reduce(fn func(prev, curr T) T, initialValue T) T` | Reduce to single value |
| `ForEach(fn func(T)) *Slice[T]` | Execute function for each element |
| `ForEachIndex(fn func(int, T)) *Slice[T]` | Execute function with index |
| `Data() []T` | Get safe copy of underlying slice |
| `Raw() []T` | Get underlying slice directly (read-only intent) |

## License

MIT License
