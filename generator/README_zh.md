# Generator - Go 中的泛型生成器

[English](./README.md) | 简体中文

`generator` 是一个轻量级的泛型生成器实现，旨在使用协程模式在 Go 中迭代生成值。它通过通道实现值的产生和接收调用者的返回值，提供了一种灵活高效的迭代器实现方式。

### 特性

- **泛型支持**：支持任意类型（需 Go 1.18+）。
- **生成机制**：通过 `Yield` 结构体产生值并接收返回值。
- **简单迭代**：使用 `Next` 方法获取值，直到生成器完成。
- **资源安全**：生成完成后自动关闭通道，防止资源泄漏。
- **并发安全**：通过适当的同步机制支持并发访问。

### 安装

将包添加到你的 Go 项目中：

```bash
go get github.com/wsshow/op/generator
```

### 使用示例

以下是一个基本的使用示例：

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/generator"
)

func main() {
    // 创建一个生成器，生成 0 到 4 的数字
    gen := generator.NewGenerator[int](func(yield generator.Yield[int]) {
        for i := 0; i < 5; i++ {
            result := yield.Yield(i)
            fmt.Printf("生成了 %d，接收到: %v\n", i, result)
        }
    })

    // 迭代获取值
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
生成了 0，接收到: ack-0
接收到的值: 0
生成了 1，接收到: ack-1
接收到的值: 1
生成了 2，接收到: ack-2
接收到的值: 2
生成了 3，接收到: ack-3
接收到的值: 3
生成了 4，接收到: ack-4
接收到的值: 4
```

### API 概览

#### 创建和初始化

- `NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T]`：创建并启动一个新的生成器，传入生成逻辑函数。

#### 核心结构

- `Yield[T any]`：用于产生值和接收返回值的结构体。

  - `Yield(value T) any`：产生一个值，并可选地返回调用者的结果。

- `Generator[T any]`：生成器实例。
  - `Next(values ...any) (value T, done bool)`：获取下一个值，`done` 为 `true` 时表示生成结束。

### 注意事项

- 生成器在单独的 goroutine 中运行，生成函数退出时会关闭所有通道。
- `Next` 可选地接受一个值传递给生成器，若不提供，则 `Yield` 返回 `nil`。
- 生成完成后，`Next` 的后续调用将返回 `T` 的零值和 `done=true`。
- 确保生成函数不会无限阻塞，以避免死锁。
