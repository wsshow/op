# Deque - 一个高性能的泛型双端队列

[English](./README.md) | 简体中文

`deque` 是一个高性能的泛型双端队列（Double-Ended Queue）实现，支持在队列两端高效地添加和移除元素。它基于环形缓冲区（circular buffer）设计，容量按 2 的幂次动态调整，支持多种操作，如插入、移除、旋转和搜索等。

## 特性

- **泛型支持**：适用于任何类型（Go 1.18+）。
- **高效操作**：在队列两端添加和移除元素的时间复杂度为 O(1)。
- **动态调整**：容量按需扩展或缩减，始终保持 2 的幂。
- **丰富功能**：支持旋转、搜索、插入、移除等操作。
- **安全设计**：对空队列或无效索引的操作会触发 panic。

## 安装

将包添加到你的 Go 项目中：

```bash
go get github.com/wsshow/op/deque
```

## 使用示例

以下是一些基本用法示例：

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/deque"
)

func main() {
    // 创建一个新的双端队列
    d := deque.New[int]()

    // 在尾部添加元素
    d.PushBack(1)
    d.PushBack(2)
    d.PushBack(3)
    fmt.Println("Size:", d.Size()) // 输出: Size: 3

    // 在头部添加元素
    d.PushFront(0)
    fmt.Println("Front:", d.Front()) // 输出: Front: 0
    fmt.Println("Back:", d.Back())   // 输出: Back: 3

    // 访问指定索引
    fmt.Println("At 1:", d.At(1)) // 输出: At 1: 1

    // 移除元素
    front := d.PopFront()
    back := d.PopBack()
    fmt.Println("Popped Front:", front) // 输出: Popped Front: 0
    fmt.Println("Popped Back:", back)   // 输出: Popped Back: 3

    // 旋转队列
    d.PushBack(4)
    d.Rotate(1) // 正向旋转1步
    fmt.Println("After Rotate:", d.At(0)) // 输出: After Rotate: 2

    // 搜索元素
    idx := d.Index(func(x int) bool { return x > 1 })
    fmt.Println("Index of >1:", idx) // 输出: Index of >1: 1
}
```

## API 概览

### 创建和初始化

- `New[T]() *Deque[T]`: 创建一个新的双端队列实例。

### 基本操作

- `PushBack(elem T)`: 在尾部添加元素。
- `PushFront(elem T)`: 在头部添加元素。
- `PopFront() T`: 从头部移除并返回元素。
- `PopBack() T`: 从尾部移除并返回元素。
- `Front() T`: 返回头部元素。
- `Back() T`: 返回尾部元素。

### 容量管理

- `Capacity() int`: 返回当前容量。
- `Size() int`: 返回当前元素数量。
- `Grow(n int)`: 确保有空间容纳额外 n 个元素。
- `SetBaseCap(baseCap int)`: 设置基础容量。

### 其他操作

- `At(index int) T`: 获取指定索引处的元素。
- `Set(index int, item T)`: 设置指定索引处的值。
- `Insert(at int, item T)`: 在指定位置插入元素。
- `Remove(at int) T`: 移除并返回指定索引处的元素。
- `Rotate(steps int)`: 旋转队列。
- `Index(match func(T) bool) int`: 从头部搜索满足条件的元素索引。
- `RIndex(match func(T) bool) int`: 从尾部搜索满足条件的元素索引。
- `Swap(idxA, idxB int)`: 交换两个索引处的值。
- `Clear()`: 清空队列但保留容量。

## 注意事项

- 队列操作（如 `PopFront`、`Front` 等）在空队列上调用会触发 panic。
- 中间插入（`Insert`）和移除（`Remove`）的时间复杂度为 O(n)，不适合频繁使用。
- 容量调整时，队列大小始终为 2 的幂次。

## 参考来源

本实现参考了 [gammazero/deque](https://github.com/gammazero/deque)
