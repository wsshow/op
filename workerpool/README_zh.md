# WorkerPool - Go 中的泛型工作协程池

[English](./README.md) | 简体中文

`workerpool` 是一个高性能的工作协程池实现，用于限制并发执行任务的协程数量。它基于通道和动态工作协程管理设计，支持任务提交、暂停、停止和等待队列管理等功能。当没有任务到来时，工作协程会逐渐停止，以节省资源。

## 特性

- **并发控制**：限制最大并发工作协程数，确保资源使用可控。
- **动态调整**：根据任务负载动态创建或关闭工作协程。
- **任务队列**：支持等待队列，当工作协程繁忙时任务会排队等待。
- **暂停与停止**：支持暂停所有工作协程或停止协程池，可选择是否等待队列任务完成。
- **高效设计**：任务提交不阻塞，空闲协程会在超时后自动关闭。

## 安装

将包添加到你的 Go 项目中：

```bash
go get github.com/wsshow/op/workerpool
```

## 使用示例

以下是一些基本用法示例：

```go
package main

import (
    "fmt"
    "time"
    "github.com/wsshow/op/workerpool"
)

func main() {
    // 创建一个最大并发数为 2 的工作协程池
    pool := workerpool.New(2)

    // 提交异步任务
    for i := 0; i < 5; i++ {
        i := i
        pool.Submit(func() {
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("Task %d completed\n", i)
        })
    }

    // 提交同步任务并等待完成
    pool.SubmitWait(func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Println("Synchronous task completed")
    })

    // 暂停协程池 1 秒
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    pool.Pause(ctx)
    fmt.Println("Pool paused for 1 second")

    // 停止协程池并等待所有任务完成
    pool.StopWait()
    fmt.Println("Pool stopped, all tasks completed")
}
```

## API 概览

### 创建和初始化

- `New(maxWorkers int) *WorkerPool`：创建一个新的工作协程池，指定最大并发工作协程数。

### 基本操作

- `Submit(task func())`：提交一个异步任务到协程池。
- `SubmitWait(task func())`：提交一个任务并等待其执行完成。
- `Size() int`：返回最大并发工作协程数。
- `WaitingQueueSize() int`：返回等待队列中的任务数。

### 生命周期管理

- `Stop()`：停止协程池，仅完成当前运行任务，未运行任务被放弃。
- `StopWait()`：停止协程池并等待所有排队任务完成。
- `Stopped() bool`：返回协程池是否已停止。
- `Pause(ctx context.Context)`：暂停所有工作协程，直到 Context 取消或超时。

## 注意事项

- 调用 `Stop` 或 `StopWait` 后不得再次提交任务，否则可能引发 panic。
- `Pause` 期间任务会继续排队，但不执行，直到暂停解除。
- 空闲工作协程在 2 秒（`idleTimeout`）无任务后自动关闭。
- 任务函数需通过闭包捕获外部值，返回值应通过通道传递。

## 参考来源

本实现参考了 [gammazero/workerpool](https://github.com/gammazero/workerpool).
