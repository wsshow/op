# Generator - Go 中的泛型生成器

[English](./README.md) | 简体中文

`generator` 是一个轻量级的泛型生成器，使用协程模式在 Go 中迭代生成值。支持双向通信——消费者可以在每次迭代时将结果回传给生成器。

### 特性

- **泛型支持**：支持任意类型（模块要求 Go 1.24+）。
- **双向通信**：`Send` 产生值的同时可接收调用者回传的结果。
- **提前终止**：`Stop` 通知生成器停止，防止 goroutine 泄漏。
- **简单迭代**：使用 `Next` 方法获取值，直到生成器完成。
- **资源安全**：生成完成后自动关闭通道。

### 设计

每个 `Generator` 仅支持**单个消费者 goroutine**。生成器函数在独立的 goroutine 中运行，通过无缓冲通道与消费者通信，确保生产者与消费者之间严格握手。

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/generator"
)

func main() {
    gen := generator.NewGenerator(func(yield generator.Yield[int]) {
        for i := 0; i < 5; i++ {
            result := yield.Send(i)
            fmt.Printf("发送 %d，接收到: %v\n", i, result)
        }
    })

    for i := 0; ; i++ {
        value, done := gen.Next(fmt.Sprintf("ack-%d", i))
        if done {
            break
        }
        fmt.Printf("接收到的值: %d\n", value)
    }
}
```

**输出：**

```
发送 0，接收到: ack-0
接收到的值: 0
发送 1，接收到: ack-1
接收到的值: 1
发送 2，接收到: ack-2
接收到的值: 2
发送 3，接收到: ack-3
接收到的值: 3
发送 4，接收到: ack-4
接收到的值: 4
```

### 提前终止

调用 `Stop()` 通知生成器停止。生成器函数应检查 `yield.Stopped()` 并及时返回。

```go
gen := generator.NewGenerator(func(yield generator.Yield[int]) {
    for i := 0; ; i++ {
        result := yield.Send(i)
        if yield.Stopped() {
            return
        }
        fmt.Println("接收到:", result)
    }
})

gen.Next()
gen.Next()
gen.Stop() // 生成器 goroutine 将退出
```

### API 概览

#### 创建

- `NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T]` — 创建并启动一个新的生成器。

#### Yield

- `Send(value T) any` — 向消费者发送值并阻塞等待结果。若生成器已停止，返回 `nil`。
- `Stopped() bool` — 报告消费者是否已请求停止生成器。

#### Generator

- `Next(values ...any) (value T, done bool)` — 获取下一个值；`done` 为 `true` 表示生成结束或已停止。
- `Stop()` — 通知生成器停止。可安全地多次调用。
