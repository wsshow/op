# Generator - A Generic Generator in Go

English | [简体中文](README_zh.md)

`generator` is a lightweight, generic generator implementation in Go, designed to produce values iteratively using a coroutine-like pattern. It supports bidirectional communication — the consumer can send results back to the generator on each iteration.

### Features

- **Generic**: Works with any type using Go generics (Go 1.24+ module requirement).
- **Bidirectional**: `Send` yields a value and optionally receives a result from the caller.
- **Early Termination**: `Stop` signals the generator to stop, preventing goroutine leaks.
- **Simple Iteration**: `Next` fetches values until the generator is done.
- **Resource Safe**: Completion is signaled automatically; call `Stop` when abandoning iteration early.

### Design

Each `Generator` supports a **single consumer goroutine**. The generator function runs in its own goroutine and communicates with the consumer through unbuffered channels, ensuring a strict handshake between producer and consumer.

### Usage Example

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
            fmt.Printf("sent %d, received: %v\n", i, result)
        }
    })

    for i := 0; ; i++ {
        value, done := gen.Next(fmt.Sprintf("ack-%d", i))
        if done {
            break
        }
        fmt.Printf("received value: %d\n", value)
    }
}
```

**Output:**

```
sent 0, received: ack-0
received value: 0
sent 1, received: ack-1
received value: 1
sent 2, received: ack-2
received value: 2
sent 3, received: ack-3
received value: 3
sent 4, received: ack-4
received value: 4
```

### Early Termination

Call `Stop()` to tell the generator to stop. The generator function should check `yield.Stopped()` and return promptly.

```go
gen := generator.NewGenerator(func(yield generator.Yield[int]) {
    for i := 0; ; i++ {
        result := yield.Send(i)
        if yield.Stopped() {
            return
        }
        fmt.Println("received:", result)
    }
})

gen.Next()
gen.Next()
gen.Stop() // generator goroutine will exit
```

### API Overview

#### Creation

- `NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T]` — Creates and starts a new generator.

#### Yield

- `Send(value T) any` — Sends a value to the consumer and blocks until a result is received. Returns `nil` if the generator was stopped.
- `Stopped() bool` — Reports whether the consumer has requested the generator to stop.

#### Generator

- `Next(values ...any) (value T, done bool)` — Retrieves the next value; `done` is `true` when generation is complete or the generator was stopped.
- `Stop()` — Signals the generator to stop. Safe to call multiple times.
