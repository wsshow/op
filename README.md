# OP - A Collection of Go Utility Packages

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

English | [简体中文](README_zh.md)

`op` is a carefully crafted Go utility toolkit providing reusable, generic-first packages for common programming tasks. Each package is designed for performance and usability, with clean APIs that integrate naturally into Go projects. Import the top-level package to access all functionality without managing individual sub-package dependencies.

## Features

- **High Performance**: Optimized implementations — amortized O(1) deque end operations, ring-buffer backing, and allocation-conscious hot paths.
- **Generic Support**: Full support for Go generics with type-safe APIs across all collection and utility packages.
- **Modular Design**: Each sub-package is self-contained and can be used independently or through the unified entry point.
- **Clean API**: Consistent patterns — method chaining, safe variants (`Try*`/`Peek*`), and idiomatic error handling.
- **Well Tested**: Comprehensive unit test coverage across all packages.

## Packages

### deque - Double-Ended Queue

A high-performance generic deque backed by a ring buffer. Head and tail operations are amortized O(1), including occasional resizing. The buffer auto-expands on push and shrinks when sparsely populated. Supports random access, rotation, insertion, and predicate-based search.

```go
d := op.NewDeque[int](64)       // pre-allocate ring buffer capacity
d.PushBack(1)                    // [1]
d.PushBack(2)                    // [1, 2]
d.PushFront(0)                   // [0, 1, 2]
d.PushBack(3)                    // [0, 1, 2, 3]

d.PopFront()                    // returns 0, panics if empty
d.Rotate(1)                      // rotate forward: [3, 1, 2]
d.Insert(1, 99)                  // [3, 99, 1, 2]

// Safe access without panics
if v, ok := d.PeekFront(); ok {
    // use v
}

// Search for matching element
idx := d.Index(func(x int) bool { return x > 50 })  // idx=1 → 99
```

**Docs**: [deque/README.md](deque/README.md) | [中文文档](deque/README_zh.md)

### emission - Event Emitter

A generic publish-subscribe system with typed events and payloads. Supports asynchronous fire-and-forget, concurrent emit-with-wait, and fully synchronous dispatch. Provides listener lifecycle management via subscriptions, one-time listeners, panic recovery, and configurable concurrency limits.

```go
// E = event type (comparable), T = payload type
em := op.NewEmitter[string, int]()

// Subscribe with lifecycle management
sub := em.On("order.created", func(amount int) {
    fmt.Printf("New order: $%d\n", amount)
})
defer sub.Unsubscribe()

// One-shot listener
em.Once("startup", func(v int) { fmt.Println("Init complete") })

// Fire and forget (async)
em.Emit("order.created", 150)

// Fire concurrently and wait for all listeners
em.EmitWait("order.created", 200)

// Fire synchronously, in registration order
em.EmitSync("order.created", 300)

// Recover from listener panics
em.RecoverWith(func(event string, listener any, panicVal any) {
    log.Printf("Panic in listener for %s: %v", event, panicVal)
})

// Limit concurrent listener goroutines
em.SetConcurrency(4)
```

**Docs**: [emission/README.md](emission/README.md)

### linq - LINQ-Style Queries

A chainable query API for Go slices inspired by .NET LINQ. Provides filtering, projection, sorting, grouping, aggregation, set operations, and joins. Operations evaluate eagerly. `Linq` is a value type; most chain methods return new values.

```go
import (
    "github.com/wsshow/op"
    "github.com/wsshow/op/linq"
)

// --- Filtering and projection ---
results := op.LinqFrom([]int{1, 2, 3, 4, 5, 6}).
    Where(func(x int) bool { return x%2 == 0 }).
    Select(func(x int) int { return x * 10 }).
    Results()
// results = [20, 40, 60]

// --- Sorting with multi-level keys ---
users := []struct{ Name string; Age int }{{"Alice", 30}, {"Bob", 25}, {"Carol", 35}}
ordered := linq.OrderBy(op.LinqFrom(users),
    func(u struct{ Name string; Age int }) int { return u.Age },
).ThenByDescending(func(a, b struct{ Name string; Age int }) int {
    return strings.Compare(a.Name, b.Name)
})
for _, u := range ordered.Results() {
    fmt.Println(u.Name, u.Age)
}

// --- Aggregation ---
nums := op.LinqFrom([]int{10, 20, 30, 40})
sum := linq.Sum(nums)           // 100
avg := linq.Average(nums)       // 25.0
min, _ := linq.MinVal(nums)     // 10
cnt := nums.CountBy(func(x int) bool { return x > 20 })  // 2

// --- Set operations (comparable types) ---
a := op.LinqFrom([]int{1, 2, 3, 4})
b := op.LinqFrom([]int{3, 4, 5, 6})
union := linq.Union(a, b)        // [1, 2, 3, 4, 5, 6]
inter := linq.Intersect(a, b)    // [3, 4]
diff  := linq.Except(a, b)       // [1, 2]

// --- Grouping ---
words := op.LinqFrom([]string{"apple", "banana", "apricot", "blueberry", "avocado"})
groups := linq.GroupBy(words, func(w string) string { return string(w[0]) })
for _, g := range groups {
    fmt.Printf("Key %s: %v\n", g.Key, g.Items)
}

// --- Joins ---
orders := op.LinqFrom([]struct{ ID, UserID int }{{1, 100}, {2, 200}})
customers := op.LinqFrom([]struct{ ID int; Name string }{{100, "Alice"}, {200, "Bob"}})
joined := linq.Join(
    orders, customers,
    func(o struct{ ID, UserID int }) int { return o.UserID },
    func(c struct{ ID int; Name string }) int { return c.ID },
    func(o struct{ ID, UserID int }, c struct{ ID int; Name string }) string {
        return fmt.Sprintf("Order #%d by %s", o.ID, c.Name)
    },
)
```

**Docs**: [linq/README.md](linq/README.md)

### process - Process Management

Tools for spawning, monitoring, and managing external processes with full lifecycle control. Supports stdout/stderr line callbacks, automatic restart with interval gating, graceful shutdown with configurable timeouts, and multi-process orchestration via `Manager`.

```go
// --- Single process ---
proc := op.NewProcess(op.Options{
    ExecPath: "my-server",
    Args:     []string{"--port", "8080", "--verbose"},
    Env:      []string{"LOG_LEVEL=debug"},
    OnStdout: func(line string) { log.Println("OUT:", line) },
    OnStderr: func(line string) { log.Println("ERR:", line) },
    OnBefore: func(p *op.Process) { log.Println("Starting...") },
    OnAfter:  func(p *op.Process) { log.Printf("Exited with code %d", p.ExitCode()) },
})

if err := proc.Start(); err != nil {
    log.Fatal(err)
}

// Wait for completion
<-proc.Done()
log.Printf("Exit code: %d", proc.ExitCode())

// Graceful restart with backpressure
proc.Restart()

// Signal handling
proc.Signal(os.Interrupt)

// Custom stop timeout
proc.StopWithTimeout(10 * time.Second)

// --- Multi-process manager ---
mgr := op.NewProcessManager()

mgr.Add("api", op.Options{
    ExecPath: "./api-server",
    Args:     []string{"--port", "8080"},
})
mgr.Add("worker", op.Options{
    ExecPath: "./worker",
    Args:     []string{"--queue", "default"},
})

// Query and control
if p, ok := mgr.Get("api"); ok {
    log.Printf("API PID: %d", p.Pid())
}

// Iterate all processes
mgr.Range(func(name string, p *op.Process) bool {
    log.Printf("%s: running=%v", name, p.IsRunning())
    return true
})

// Bulk operations
mgr.RestartAll()
defer mgr.StopAllWithTimeout(15 * time.Second)
```

**Docs**: [process/README.md](process/README.md) | [中文文档](process/README_zh.md)

### slice - Slice Utilities

A generic slice wrapper with functional operations inspired by JavaScript's array methods. Supports map, filter, reduce, element insertion/removal at arbitrary positions, sorting, reversal, concatenation, and safe access. Most mutation methods return `*Slice` for chaining.

```go
import (
    "github.com/wsshow/op"
    "github.com/wsshow/op/slice"
)

s := op.NewSlice(1, 2, 3)

// --- Mutation (in-place, chainable) ---
s.Push(4, 5).Unshift(0)
// s = [0, 1, 2, 3, 4, 5]

val, ok := s.Pop()     // val=5, ok=true
val, ok = s.Shift()    // val=0, ok=true

s.Insert(2, 99)        // [1, 2, 99, 3, 4]

// --- Functional operations (return new Slice) ---
doubled := s.Map(func(x int) int { return x * 2 })
evens := s.Filter(func(x int) bool { return x%2 == 0 })

// --- Type conversion ---
strs := slice.MapTo(s, func(x int) string { return strconv.Itoa(x) })
// strs is *Slice[string]

// --- Reduction ---
sum := s.Reduce(func(acc, cur int) int { return acc + cur }, 0)

// --- Sorting ---
s.Sort(func(a, b int) bool { return a < b })
s.Reverse()

// --- Safe access ---
if v, ok := s.Find(func(x int) bool { return x > 50 }); ok {
    // use v
}
found := s.Some(func(x int) bool { return x > 3 })   // true if any match
allPos := s.Every(func(x int) bool { return x > 0 })  // true if all match

// --- Combine slices ---
other := op.NewSlice(10, 20, 30)
merged := s.Concat(other)       // new Slice, originals unchanged
```

**Docs**: [slice/README.md](slice/README.md) | [中文文档](slice/README_zh.md)

### str - String Utilities

A string wrapper with common text operations. Most methods mutate in-place and return `*String` for chaining. Includes parsing helpers, Unicode-aware reversal, formatting, and substring extraction with Python-style negative indexing.

```go
s := op.NewString("  Hello, World!  ")

// --- Transformation (in-place, chainable) ---
s.TrimSpace().ToLower().ReplaceAll("world", "Gopher")
// s.String() = "hello, Gopher!"

// --- Inspection ---
s.Contains("Gopher")       // true
s.StartsWith("hello")      // true
s.Count("o")               // 2
s.Length()                 // 15
s.RuneLength()             // 15 (Unicode-aware)

// --- Parsing ---
numStr := op.NewString("  42  ")
val, err := numStr.TrimSpace().ToInt()  // val=42

// --- Non-mutating (return new *String) ---
cloned := s.Clone()
formatted := op.NewString("Hello, %s!").Format("World")  // "Hello, World!"
sub := s.Substring(7, 12)                                // "Gopher"
joined := op.JoinStrings([]string{"a", "b", "c"}, ",")   // "a,b,c"

// --- Unicode-aware operations ---
op.NewString("こんにちは").Reverse()                     // "はちにんこ"
```

**Docs**: [str/README.md](str/README.md)

### workerpool - Worker Pool

A high-performance goroutine pool that limits concurrency and queues overflow tasks. Workers are dynamically created on demand and reclaimed after an idle timeout. Supports pause/resume with context-based timeouts, graceful shutdown modes, and panic recovery.

```go
wp := op.NewWorkerPool(4,  // max concurrent workers
    op.WithPanicHandler(func(v any) {
        log.Printf("Task panicked: %v", v)
    }),
)

// Submit fire-and-forget tasks
for i := 0; i < 100; i++ {
    i := i
    wp.Submit(func() {
        // process item i
        time.Sleep(50 * time.Millisecond)
    })
}

// Submit and wait for a specific task to complete
wp.SubmitWait(func() {
    // critical pre-flight check
    log.Println("Pre-flight complete")
})

// Inspect queue pressure
queued := wp.WaitingQueueSize()
log.Printf("Tasks waiting: %d", queued)

// Pause all workers temporarily
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
wp.Pause(ctx)   // blocks until workers pause or context expires

// Graceful shutdown — complete queued tasks, then stop
wp.StopWait()

// Immediate shutdown — finish running tasks, discard queued
// wp.Stop()
```

**Docs**: [workerpool/README.md](workerpool/README.md) | [中文文档](workerpool/README_zh.md)

### generator - Generator

A lightweight coroutine-style generator using goroutines and channels. The generator function yields values via `Yield.Send()`. The consumer retrieves them with `Next()` and can optionally send results back at each step, enabling bidirectional communication between producer and consumer.

```go
// --- Basic value generation ---
g := op.NewGenerator(func(yield op.Yield[int]) {
    for i := 0; i < 5; i++ {
        if yield.Stopped() {
            return  // consumer requested stop
        }
        // yield.Send blocks until consumer calls Next
        result := yield.Send(i)
        fmt.Printf("Generator received: %v\n", result)
    }
})

// Consume all values
for {
    val, done := g.Next("ack")  // "ack" sent back to generator
    if done {
        break
    }
    fmt.Printf("Consumer got: %d\n", val)
}

// --- Infinite sequence with early stop ---
fibGen := op.NewGenerator(func(yield op.Yield[int]) {
    a, b := 0, 1
    for {
        if yield.Stopped() {
            return
        }
        yield.Send(a)
        a, b = b, a+b
    }
})

// Take first 10 Fibonacci numbers
for i := 0; i < 10; i++ {
    v, done := fibGen.Next()
    if done {
        break
    }
    fmt.Println(v)  // 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
}
fibGen.Stop()
```

**Docs**: [generator/README.md](generator/README.md) | [中文文档](generator/README_zh.md)

## Installation

```
go get github.com/wsshow/op
```

Import the top-level package to access all types and constructors:

```go
import "github.com/wsshow/op"
```

Alternatively, import individual sub-packages for a lighter dependency footprint:

```go
import (
    "github.com/wsshow/op/deque"
    "github.com/wsshow/op/linq"
)
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/wsshow/op"
)

func main() {
    // String manipulation with chaining
    s := op.NewString("  Hello, World!  ")
    s.TrimSpace().ToUpper().ReplaceAll("WORLD", "GOPHER")
    fmt.Println(s)  // "HELLO, GOPHER!"

    // Slice with functional operations
    sl := op.NewSlice(1, 2, 3, 4, 5).
        Filter(func(x int) bool { return x%2 != 0 }).
        Map(func(x int) int { return x * x })
    fmt.Println(sl.Data())  // [1, 9, 25]

    // Type-safe event emitter
    em := op.NewEmitter[string, string]()
    em.On("message", func(payload string) {
        fmt.Println("Received:", payload)
    })
    em.Emit("message", "Hello from emitter")

    // LINQ filtering and chaining
    scores := op.LinqFrom([]int{85, 92, 78, 95, 88})
    passed := scores.
        Where(func(x int) bool { return x >= 80 }).
        Sort(func(a, b int) bool { return a < b })
    fmt.Println(passed.Results())                                 // [85, 88, 92, 95]
    fmt.Println("Passed:", passed.Count(), "out of", scores.Count())  // 4 out of 5

    // High-performance deque
    d := op.NewDeque[string](8)
    d.PushBack("alpha")
    d.PushBack("beta")
    d.PushFront("omega")
    for d.Size() > 0 {
        fmt.Println(d.PopFront())
    }

    // Coroutine-style generator
    g := op.NewGenerator(func(yield op.Yield[int]) {
        for i := 1; i <= 3; i++ {
            yield.Send(i * 10)
        }
    })
    for {
        v, done := g.Next()
        if done {
            break
        }
        fmt.Println(v)  // 10, 20, 30
    }

    // Worker pool with panic recovery
    wp := op.NewWorkerPool(4, op.WithPanicHandler(func(v any) {
        log.Printf("Recovered from panic: %v", v)
    }))
    for i := 0; i < 20; i++ {
        i := i
        wp.Submit(func() {
            fmt.Printf("Task %d running\n", i)
        })
    }
    wp.StopWait()

    // Process management
    proc := op.NewProcess(op.Options{
        ExecPath: "echo",
        Args:     []string{"hello"},
        OnStdout: func(line string) { fmt.Println("OUT:", line) },
    })
    if err := proc.Run(); err != nil {
        log.Fatal(err)
    }

    // Multi-process manager
    mgr := op.NewProcessManager()
    mgr.Add("healthcheck", op.Options{
        ExecPath: "curl",
        Args:     []string{"-s", "http://localhost:8080/health"},
    })
    defer mgr.StopAll()
}
```

## Directory Structure

```
op/
├── deque/              # Generic ring-buffer deque
├── emission/           # Typed event emitter for pub/sub
├── linq/               # LINQ-style chainable query library
├── process/            # External process lifecycle management
├── slice/              # Generic slice wrapper with functional ops
├── str/                # String wrapper with chaining
├── workerpool/         # Bounded goroutine pool
├── generator/          # Coroutine-style generator
└── op.go               # Unified entry point with type aliases
```

## Contributing

Contributions are welcome. Please ensure existing tests pass and new functionality includes test coverage. Open an issue to discuss significant changes before submitting a PR.

## License

MIT - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [deque](https://github.com/gammazero/deque) - Inspiration for the ring-buffer deque
- [workerpool](https://github.com/gammazero/workerpool) - Inspiration for the worker pool
- [emission](https://github.com/chuckpreslar/emission) - Inspiration for the event emitter
