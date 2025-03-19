# Deque - A Generic Double-Ended Queue in Go

English | [简体中文](README_zh.md)

`deque` is a high-performance generic double-ended queue (Deque) implementation in Go, designed to efficiently add and remove elements from both ends. It is built on a circular buffer, with capacity dynamically adjusted in powers of 2, and supports a variety of operations such as insertion, removal, rotation, and searching.

## Features

- **Generic Support**: Works with any type (Go 1.18+).
- **Efficient Operations**: Adding and removing elements at both ends has O(1) time complexity.
- **Dynamic Resizing**: Capacity expands or shrinks as needed, always maintaining a power of 2.
- **Rich Functionality**: Includes rotation, searching, insertion, and removal operations.
- **Safety Design**: Operations on an empty queue or invalid indices trigger a panic.

## Installation

Add the package to your Go project:

```bash
go get github.com/wsshow/op/deque
```

## Usage Example

Here are some basic usage examples:

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/deque"
)

func main() {
    // Create a new double-ended queue
    d := deque.New[int]()

    // Add elements to the back
    d.PushBack(1)
    d.PushBack(2)
    d.PushBack(3)
    fmt.Println("Size:", d.Size()) // Output: Size: 3

    // Add an element to the front
    d.PushFront(0)
    fmt.Println("Front:", d.Front()) // Output: Front: 0
    fmt.Println("Back:", d.Back())   // Output: Back: 3

    // Access element at a specific index
    fmt.Println("At 1:", d.At(1)) // Output: At 1: 1

    // Remove elements
    front := d.PopFront()
    back := d.PopBack()
    fmt.Println("Popped Front:", front) // Output: Popped Front: 0
    fmt.Println("Popped Back:", back)   // Output: Popped Back: 3

    // Rotate the queue
    d.PushBack(4)
    d.Rotate(1) // Rotate forward by 1 step
    fmt.Println("After Rotate:", d.At(0)) // Output: After Rotate: 2

    // Search for an element
    idx := d.Index(func(x int) bool { return x > 1 })
    fmt.Println("Index of >1:", idx) // Output: Index of >1: 1
}
```

## API Overview

### Creation and Initialization

- `New[T]() *Deque[T]`: Creates a new double-ended queue instance.

### Basic Operations

- `PushBack(elem T)`: Adds an element to the back of the queue.
- `PushFront(elem T)`: Adds an element to the front of the queue.
- `PopFront() T`: Removes and returns the element from the front.
- `PopBack() T`: Removes and returns the element from the back.
- `Front() T`: Returns the element at the front.
- `Back() T`: Returns the element at the back.

### Capacity Management

- `Capacity() int`: Returns the current capacity of the queue.
- `Size() int`: Returns the current number of elements.
- `Grow(n int)`: Ensures space for at least n additional elements.
- `SetBaseCap(baseCap int)`: Sets the base capacity.

### Additional Operations

- `At(index int) T`: Retrieves the element at the specified index.
- `Set(index int, item T)`: Sets the value at the specified index.
- `Insert(at int, item T)`: Inserts an element at the specified position.
- `Remove(at int) T`: Removes and returns the element at the specified index.
- `Rotate(steps int)`: Rotates the queue by the specified number of steps.
- `Index(match func(T) bool) int`: Searches for the first element satisfying the condition from the front.
- `RIndex(match func(T) bool) int`: Searches for the first element satisfying the condition from the back.
- `Swap(idxA, idxB int)`: Swaps the elements at the specified indices.
- `Clear()`: Clears the queue while retaining its capacity.

## Notes

- Operations like `PopFront`, `Front`, etc., will panic if called on an empty queue.
- Middle insertions (`Insert`) and removals (`Remove`) have O(n) time complexity and are not suitable for frequent use.
- During capacity adjustments, the queue size is always maintained as a power of 2.

## Reference

This implementation is inspired by [gammazero/deque](https://github.com/gammazero/deque).
