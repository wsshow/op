# WorkerPool - A Generic Worker Pool in Go

English | [简体中文](README_zh.md)

`workerpool` is a high-performance worker pool implementation in Go, designed to limit the number of goroutines concurrently executing tasks. It leverages channels and dynamic worker management, supporting task submission, pausing, stopping, and waiting queue management. When no tasks arrive, workers are gradually stopped to conserve resources.

## Features

- **Concurrency Control**: Limits the maximum number of concurrent workers, ensuring manageable resource usage.
- **Dynamic Adjustment**: Creates or terminates workers dynamically based on task load.
- **Task Queue**: Supports a waiting queue for tasks when all workers are busy.
- **Pause and Stop**: Allows pausing all workers or stopping the pool, with an option to wait for queued tasks to complete.
- **Efficient Design**: Non-blocking task submission, with idle workers automatically shut down after a timeout.

## Installation

Add the package to your Go project:

```bash
go get github.com/wsshow/op/workerpool
```

## Usage Example

Here are some basic usage examples:

```go
package main

import (
    "fmt"
    "time"
    "github.com/wsshow/op/workerpool"
)

func main() {
    // Create a worker pool with a maximum of 2 concurrent workers
    pool := workerpool.New(2)

    // Submit asynchronous tasks
    for i := 0; i < 5; i++ {
        i := i
        pool.Submit(func() {
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("Task %d completed\n", i)
        })
    }

    // Submit a synchronous task and wait for completion
    pool.SubmitWait(func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Println("Synchronous task completed")
    })

    // Pause the worker pool for 1 second
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    pool.Pause(ctx)
    fmt.Println("Pool paused for 1 second")

    // Stop the worker pool and wait for all tasks to complete
    pool.StopWait()
    fmt.Println("Pool stopped, all tasks completed")
}
```

## API Overview

### Creation and Initialization

- `New(maxWorkers int) *WorkerPool`: Creates a new worker pool with the specified maximum number of concurrent workers.

### Basic Operations

- `Submit(task func())`: Submits an asynchronous task to the worker pool.
- `SubmitWait(task func())`: Submits a task and waits for its execution to complete.
- `Size() int`: Returns the maximum number of concurrent workers.
- `WaitingQueueSize() int`: Returns the number of tasks in the waiting queue.

### Lifecycle Management

- `Stop()`: Stops the worker pool, completing only currently running tasks and abandoning pending ones.
- `StopWait()`: Stops the worker pool and waits for all queued tasks to complete.
- `Stopped() bool`: Returns whether the worker pool has been stopped.
- `Pause(ctx context.Context)`: Pauses all workers until the Context is canceled or times out.

## Notes

- Submitting tasks after calling `Stop` or `StopWait` may cause a panic.
- During a `Pause`, tasks continue to queue but are not executed until the pause is lifted.
- Idle workers are automatically shut down after 2 seconds (`idleTimeout`) of inactivity.
- Task functions must capture external values via closures, and return values should be sent over channels.

## Reference

This implementation is inspired by [gammazero/workerpool](https://github.com/gammazero/workerpool).
