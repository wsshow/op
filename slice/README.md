# Slice - A Generic Slice Wrapper in Go

English | [简体中文](README_zh.md)

`slice` is a generic slice wrapper in Go that provides a rich set of utility methods for common slice operations, inspired by JavaScript array methods and functional programming patterns.

## Features

- **Generic Support**: Works with any type using Go generics (Go 1.18+).
- **Chainable API**: Most methods return `*Slice[T]` for method chaining.
- **Rich Operations**: Push, pop, shift, unshift, filter, map, reduce, sort, and more.
- **Familiar Syntax**: API inspired by JavaScript arrays for easy adoption.
- **Type-Safe**: Full compile-time type checking with generics.

## Installation

Add the package to your Go project:

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
    // Create a new slice
    s := slice.New(1, 2, 3)
    
    // Add elements
    s.Push(4, 5)
    fmt.Println(s.Data()) // Output: [1 2 3 4 5]
    
    // Remove last element
    last := s.Pop()
    fmt.Println(last)     // Output: 5
    fmt.Println(s.Data()) // Output: [1 2 3 4]
}
```

### Array-like Operations

```go
s := slice.New(1, 2, 3)

// Add to beginning
s.Unshift(0)
fmt.Println(s.Data()) // Output: [0 1 2 3]

// Remove from beginning
first := s.Shift()
fmt.Println(first)     // Output: 0
fmt.Println(s.Data())  // Output: [1 2 3]
```

### Filtering and Mapping

```go
numbers := slice.New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

// Filter even numbers
evens := numbers.Filter(func(x int) bool {
    return x%2 == 0
})
fmt.Println(evens.Data()) // Output: [2 4 6 8 10]

// Double each number (modifies in-place)
numbers.Map(func(x int) int {
    return x * 2
})
fmt.Println(numbers.Data()) // Output: [2 4 6 8 10 12 14 16 18 20]
```

### Searching

```go
users := slice.New(
    struct{ Name string; Age int }{"Alice", 25},
    struct{ Name string; Age int }{"Bob", 30},
    struct{ Name string; Age int }{"Charlie", 35},
)

// Find first user over 28
user, found := users.Find(func(u struct{ Name string; Age int }) bool {
    return u.Age > 28
})
if found {
    fmt.Println(user.Name) // Output: Bob
}
```

### Finding Index (Comparable Types)

```go
names := slice.New("Alice", "Bob", "Charlie")

// Find index of "Bob"
index := slice.IndexOf(names, "Bob")
fmt.Println(index) // Output: 1

// Not found returns -1
index = slice.IndexOf(names, "David")
fmt.Println(index) // Output: -1
```

### Checking Conditions

```go
numbers := slice.New(2, 4, 6, 8, 10)

// Check if all are even
allEven := numbers.Every(func(x int) bool {
    return x%2 == 0
})
fmt.Println(allEven) // Output: true

// Check if some are greater than 5
someGreater := numbers.Some(func(x int) bool {
    return x > 5
})
fmt.Println(someGreater) // Output: true
```

### Reducing

```go
numbers := slice.New(1, 2, 3, 4, 5)

// Sum all numbers
sum := numbers.Reduce(func(acc, curr int) int {
    return acc + curr
}, 0)
fmt.Println(sum) // Output: 15

// Find max
max := numbers.Reduce(func(acc, curr int) int {
    if curr > acc {
        return curr
    }
    return acc
}, 0)
fmt.Println(max) // Output: 5
```

### Sorting

```go
numbers := slice.New(5, 2, 8, 1, 9, 3)

// Sort ascending
numbers.Sort(func(a, b int) bool {
    return a < b
})
fmt.Println(numbers.Data()) // Output: [1 2 3 5 8 9]

// Sort descending
numbers.Sort(func(a, b int) bool {
    return a > b
})
fmt.Println(numbers.Data()) // Output: [9 8 5 3 2 1]
```

### Reversing

```go
s := slice.New(1, 2, 3, 4, 5)

s.Reverse()
fmt.Println(s.Data()) // Output: [5 4 3 2 1]
```

### Concatenation

```go
s1 := slice.New(1, 2, 3)
s2 := slice.New(4, 5, 6)

// Create new slice with combined elements
combined := s1.Concat(s2)
fmt.Println(combined.Data()) // Output: [1 2 3 4 5 6]

// Original slices unchanged
fmt.Println(s1.Data()) // Output: [1 2 3]
fmt.Println(s2.Data()) // Output: [4 5 6]
```

### Slicing

```go
s := slice.New(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

// Get elements from index 2 to 5 (exclusive)
sub := s.Slice(2, 5)
fmt.Println(sub.Data()) // Output: [2 3 4]

// Original unchanged
fmt.Println(s.Length()) // Output: 10
```

### Getting and Setting Elements

```go
s := slice.New(10, 20, 30, 40, 50)

// Get element at index 2
value, ok := s.Get(2)
if ok {
    fmt.Println(value) // Output: 30
}

// Set element at index 3
success := s.Set(3, 99)
fmt.Println(success)   // Output: true
fmt.Println(s.Data())  // Output: [10 20 30 99 50]
```

### Iteration

```go
s := slice.New("apple", "banana", "cherry")

s.Foreach(func(fruit string) {
    fmt.Println(fruit)
})
// Output:
// apple
// banana
// cherry
```

### Clearing and Cloning

```go
s := slice.New(1, 2, 3, 4, 5)

// Clone the slice
clone := s.Clone()
fmt.Println(clone.Data()) // Output: [1 2 3 4 5]

// Clear the original
s.Clear()
fmt.Println(s.Length())   // Output: 0
fmt.Println(s.IsEmpty())  // Output: true

// Clone remains unchanged
fmt.Println(clone.Data()) // Output: [1 2 3 4 5]
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

// Find expensive in-stock products and increase price by 10%
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
    fmt.Printf("%s: $%.2f (Stock: %d)\n", p.Name, p.Price, p.Stock)
})
// Output:
// Laptop: $1099.99 (Stock: 5)
// Monitor: $329.99 (Stock: 10)
// Keyboard: $87.99 (Stock: 0)
```

## API Overview

### Creation
- `New[T any](values ...T) *Slice[T]`: Create a new slice with initial values

### Adding Elements
- `Push(values ...T) *Slice[T]`: Add elements to the end
- `Unshift(values ...T) *Slice[T]`: Add elements to the beginning

### Removing Elements
- `Pop() T`: Remove and return last element
- `Shift() T`: Remove and return first element
- `Clear() *Slice[T]`: Remove all elements

### Querying
- `Length() int`: Get the number of elements
- `IsEmpty() bool`: Check if slice is empty
- `Get(index int) (T, bool)`: Get element at index
- `Set(index int, value T) bool`: Set element at index

### Searching
- `Find(predicate func(T) bool) (T, bool)`: Find first matching element
- `IndexOf[T comparable](s *Slice[T], value T) int`: Find index of value

### Filtering and Transforming
- `Filter(predicate func(T) bool) *Slice[T]`: Filter elements (returns new slice)
- `Map(callbackfn func(T) T) *Slice[T]`: Transform elements (modifies in-place)
- `Foreach(callbackfn func(T)) *Slice[T]`: Execute function for each element

### Checking
- `Every(predicate func(T) bool) bool`: Check if all match
- `Some(predicate func(T) bool) bool`: Check if any match

### Aggregation
- `Reduce(callbackfn func(prev, curr T) T, initialValue T) T`: Reduce to single value

### Sorting and Ordering
- `Sort(compareFn func(a, b T) bool) *Slice[T]`: Sort elements
- `Reverse() *Slice[T]`: Reverse order

### Combining and Slicing
- `Concat(other *Slice[T]) *Slice[T]`: Concatenate slices (returns new slice)
- `Slice(start, end int) *Slice[T]`: Get sub-slice (returns new slice)

### Utility
- `Data() []T`: Get copy of underlying slice
- `Clone() *Slice[T]`: Create a deep copy

## Notes

- **Mutability**: Some methods modify the slice in-place (`Push`, `Pop`, `Map`, `Sort`, etc.), while others return new slices (`Filter`, `Concat`, `Slice`, `Clone`).
- **Empty Operations**: `Pop()` and `Shift()` return zero value if the slice is empty.
- **Bounds Checking**: `Get()` and `Set()` perform bounds checking and return/accept a boolean flag.
- **Type Safety**: All operations are type-safe at compile time thanks to generics.

## License

MIT License
