# Generator - A Generic Generator in Go

English | [简体中文](README_zh.md)

`generator` is a lightweight, generic generator implementation in Go, designed to produce values iteratively using a coroutine-like pattern. It leverages channels to yield values and optionally receive results from the caller, providing a flexible and efficient way to implement iterators.

### Features

- **Generic Support**: Works with any type using Go generics (Go 1.18+).
- **Yield Mechanism**: Allows producing values and receiving results via a `Yield` struct.
- **Simple Iteration**: Uses a `Next` method to fetch values until the generator is done.
- **Resource Safety**: Automatically closes channels upon completion to prevent leaks.
- **Concurrency Ready**: Safe for concurrent access with proper synchronization.

### Installation

Add the package to your Go project:

```bash
go get github.com/wsshow/op/generator
```

### Usage Example

Here’s a basic example demonstrating how to use the generator:

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/generator"
)

func main() {
    // Create a generator that yields numbers 0 to 4
    gen := generator.NewGenerator[int](func(yield generator.Yield[int]) {
        for i := 0; i < 5; i++ {
            result := yield.Yield(i)
            fmt.Printf("Yielded %d, received: %v\n", i, result)
        }
    })

    // Iterate over the generator
    for i := 0; ; i++ {
        value, done := gen.Next(fmt.Sprintf("ack-%d", i))
        if done {
            break
        }
        fmt.Printf("Received value: %d\n", value)
    }
}
```

**Output:**

```
Yielded 0, received: ack-0
Received value: 0
Yielded 1, received: ack-1
Received value: 1
Yielded 2, received: ack-2
Received value: 2
Yielded 3, received: ack-3
Received value: 3
Yielded 4, received: ack-4
Received value: 4
```

### API Overview

#### Creation and Initialization

- `NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T]`: Creates and starts a new generator with the provided generation function.

#### Core Structures

- `Yield[T any]`: Struct for yielding values and receiving results.

  - `Yield(value T) any`: Yields a value and optionally returns a result from the caller.

- `Generator[T any]`: The generator instance.
  - `Next(values ...any) (value T, done bool)`: Retrieves the next value; `done` is `true` when generation is complete.
