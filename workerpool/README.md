# WorkerPool - A Generic Worker Pool in Go

English | [简体中文](README_zh.md)

`workerpool` is a high-performance worker pool implementation in Go, designed to limit the number of goroutines concurrently executing tasks. It leverages channels and dynamic worker management, supporting task submission, pausing, stopping, and waiting queue management. When no tasks arrive, workers are gradually stopped to conserve resources.

## Features

- **Concurrency Control**: Limits the maximum number of concurrent workers, ensuring manageable resource usage.
- **Dynamic Adjustment**: Creates or terminates workers dynamically based on task load.
- **Task Queue**: Supports a waiting queue for tasks when all workers are busy.
- **Pause and Stop**: Allows pausing all workers or stopping the pool, with an option to wait for queued tasks to complete.
- **Efficient Design**: Submission hands tasks to the scheduler without waiting for execution; idle workers automatically shut down after a timeout.

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
    "context"
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

- `New(maxWorkers int, opts ...Option) *WorkerPool`: Creates a new worker pool with the specified maximum number of concurrent workers. Values below 1 are normalized to 1.

### Basic Operations

- `Submit(task func())`: Submits an asynchronous task; it may briefly block until the scheduler receives it.
- `SubmitWait(task func())`: Submits a task and waits for completion. If a concurrent `Stop` discards it before execution, the call returns when the pool stops.
- `Size() int`: Returns the maximum number of concurrent workers.
- `WaitingQueueSize() int`: Returns the number of tasks in the waiting queue.

### Lifecycle Management

- `Stop()`: Stops the worker pool, completing only currently running tasks and abandoning pending ones.
- `StopWait()`: Stops the worker pool and waits for all queued tasks to complete.
- `Stopped() bool`: Returns whether the worker pool has been stopped.
- `Pause(ctx context.Context)`: Pauses all workers until the context is canceled or times out. Concurrent pauses are serialized, and a waiting call still honors its own context.

### Configuration

- `WithIdleTimeout(d time.Duration)`: Sets the idle-worker retirement interval. Non-positive values keep the 2-second default.
- `WithPanicHandler(handler func(any))`: Observes task panics. A panic raised by the handler itself is also isolated.

## Notes

- Submitting tasks after the pool has stopped panics.
- During a `Pause`, tasks continue to queue but are not executed until the pause is lifted.
- With the default configuration, the scheduler attempts to retire one idle worker per 2-second interval without new tasks. Running tasks are never interrupted by idle retirement.
- Task functions must capture external values via closures, and return values should be sent over channels.
- `Stop` and `StopWait` are synchronous; do not call them from a task running in the same pool.

## Reference

This implementation is inspired by [gammazero/workerpool](https://github.com/gammazero/workerpool).
