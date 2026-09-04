# Emission - A Generic Event Emitter for Go

English | [简体中文](README_zh.md)

`emission` is a powerful generic event emitter for Go, providing a publish-subscribe pattern for event-driven programming. It features a dual-generic design that separates event identifier types from listener parameter types for maximum flexibility and type safety.

## Features

- **Dual-Generic Design**: Complete separation of event identifier and listener parameter types
- **Type Safety**: Compile-time type checking with Go generics (Go 1.24+ module requirement)
- **Async & Sync**: `Emit` (fire-and-forget), `EmitWait` (concurrent with wait), and `EmitSync` (sequential)
- **Once Listeners**: One-time event listeners that auto-remove after first trigger
- **Panic Recovery**: Always-on listener panic isolation with optional recovery callbacks or logging
- **Global Concurrency Control**: A shared semaphore caps concurrently executing async listeners
- **Thread-Safe**: All operations protected by mutex for concurrent use
- **Max Listeners Warning**: Configurable soft limit to detect potential memory leaks (Node.js-style)

## Installation

```bash
go get github.com/wsshow/op/emission
```

## Core Concept

`Emitter[E comparable, T any]` uses two generic parameters:

- **E**: Event identifier type (must be comparable, like string, int, etc.)
- **T**: Listener parameter type (can be any type)

This design lets you identify events with simple types while passing complex data structures to listeners.

## Usage Examples

### Basic Event Handling

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    emitter := emission.NewEmitter[string, string]()

    // Add a listener
    emitter.On("message", func(msg string) {
        fmt.Println("Received:", msg)
    })

    // Emit synchronously
    emitter.EmitSync("message", "Hello, World")
    // Output: Received: Hello, World
}
```

### Using Custom Data Structures

```go
type User struct {
    Name string
    Age  int
}

emitter := emission.NewEmitter[string, User]()

emitter.On("user_login", func(u User) {
    fmt.Printf("User %s (age: %d) logged in\n", u.Name, u.Age)
})

emitter.EmitSync("user_login", User{Name: "Alice", Age: 30})
// Output: User Alice (age: 30) logged in
```

### Async Event Emission

```go
type Message struct {
    ID      int
    Content string
}

emitter := emission.NewEmitter[string, Message]()

emitter.On("message", func(m Message) {
    fmt.Printf("Processing message #%d: %s\n", m.ID, m.Content)
})

// Fire-and-forget (non-blocking)
emitter.Emit("message", Message{ID: 1, Content: "Hello"})
emitter.Emit("message", Message{ID: 2, Content: "World"})
```

### Concurrent Emit with Wait

```go
emitter := emission.NewEmitter[string, string]()
emitter.SetConcurrency(4) // max 4 concurrent listeners globally

emitter.On("process", func(data string) {
    // heavy work
})

// Blocks until all listeners complete
emitter.EmitWait("process", "data")
```

### Once Listeners

```go
emitter := emission.NewEmitter[string, string]()

// Listener auto-removes after first trigger
emitter.Once("startup", func(msg string) {
    fmt.Println("Application started!")
})

emitter.EmitSync("startup", "ready") // Output: Application started!
emitter.EmitSync("startup", "ready") // No output (listener removed)
```

### Unsubscribe via Subscription

```go
emitter := emission.NewEmitter[string, int]()

sub := emitter.On("event", func(n int) { fmt.Println(n) })
emitter.EmitSync("event", 1) // Output: 1

sub.Unsubscribe()
emitter.EmitSync("event", 2) // No output
```

### Removing All Listeners

```go
emitter := emission.NewEmitter[string, int]()

emitter.On("event", func(n int) { fmt.Println("L1") })
emitter.On("event", func(n int) { fmt.Println("L2") })

fmt.Println(emitter.GetListenerCount("event")) // 2

emitter.RemoveAllListeners("event")
fmt.Println(emitter.GetListenerCount("event")) // 0
```

### Panic Recovery

```go
emitter := emission.NewEmitter[string, string]()

// Set custom recovery handler
emitter.RecoverWith(func(event string, listener interface{}, panicValue interface{}) {
    fmt.Printf("Panic in event '%s': %v\n", event, panicValue)
})

emitter.On("error", func(msg string) {
    panic("something went wrong!")
})

emitter.EmitSync("error", "data")
// Panic is caught and handled by the recovery handler
```

Without a custom recoverer, panics are still safely recovered and logged if a logger is configured.

### Using Integer Event Identifiers

```go
const (
    EventStart = iota
    EventStop
    EventPause
)

type AppState struct {
    Timestamp int64
    Status    string
}

emitter := emission.NewEmitter[int, AppState]()

emitter.On(EventStart, func(s AppState) {
    fmt.Printf("App started, status: %s\n", s.Status)
})

emitter.On(EventStop, func(s AppState) {
    fmt.Printf("App stopped, status: %s\n", s.Status)
})

emitter.EmitSync(EventStart, AppState{Timestamp: 1234567890, Status: "running"})
```

### Custom Event Types

```go
type EventType string

const (
    UserLogin  EventType = "user:login"
    UserLogout EventType = "user:logout"
)

type UserEvent struct {
    UserID   string
    Username string
    IP       string
}

emitter := emission.NewEmitter[EventType, UserEvent]()

emitter.On(UserLogin, func(e UserEvent) {
    fmt.Printf("User %s logged in from %s\n", e.Username, e.IP)
})

emitter.EmitSync(UserLogin, UserEvent{
    UserID: "u123", Username: "alice", IP: "192.168.1.1",
})
```

### Method Chaining

```go
emitter := emission.NewEmitter[string, int]()

emitter.
    SetMaxListeners(20).
    SetConcurrency(5).
    RecoverWith(func(event string, listener interface{}, panicValue interface{}) {}).
    EmitSync("init", 0)

// On/Once return *Subscription for individual unsubscribe
sub1 := emitter.On("event1", func(n int) { fmt.Println("Event 1:", n) })
sub2 := emitter.On("event2", func(n int) { fmt.Println("Event 2:", n) })

emitter.EmitSync("event1", 1).EmitSync("event2", 2)

sub1.Unsubscribe()
sub2.Unsubscribe()
```

### Introspection

```go
emitter := emission.NewEmitter[string, int]()
emitter.On("a", func(n int) {})
emitter.On("a", func(n int) {})
emitter.On("b", func(n int) {})

fmt.Println(emitter.Events())              // contains a and b; order is unspecified
fmt.Println(emitter.GetListenerCount("a")) // 2
fmt.Println(emitter.TotalListenerCount())  // 3
```

## API Overview

### Creation

- `NewEmitter[E comparable, T any]() *Emitter[E, T]`: Create a new event emitter

### Adding Listeners

- `On(event E, listener Listener[T]) *Subscription[E, T]`: Add a listener
- `AddListener(event E, listener Listener[T]) *Subscription[E, T]`: Alias for On
- `Once(event E, listener Listener[T]) *Subscription[E, T]`: Add a one-time listener
- `Subscription.Unsubscribe()`: Remove the listener from the event

### Removing Listeners

- `RemoveAllListeners(event E) *Emitter[E, T]`: Remove all listeners for an event

### Emitting Events

- `Emit(event E, value T) *Emitter[E, T]`: Fire-and-forget (async, returns immediately)
- `EmitWait(event E, value T) *Emitter[E, T]`: Async with wait (blocks until all listeners complete)
- `EmitSync(event E, value T) *Emitter[E, T]`: Synchronous sequential execution

### Introspection

- `GetListenerCount(event E) int`: Listener count for an event
- `Events() []E`: List all registered event identifiers in unspecified order
- `TotalListenerCount() int`: Total listeners across all events

### Configuration

- `SetMaxListeners(max int) *Emitter[E, T]`: Set warning threshold per event (-1 for unlimited)
- `SetConcurrency(n int) *Emitter[E, T]`: Set global max concurrent listener execution (0 for unlimited)
- `SetLogger(logger Logger) *Emitter[E, T]`: Set logger for warnings and panic recovery
- `RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T]`: Set custom panic recovery handler

### Types

- `Listener[T any] func(T)`: Listener function signature
- `RecoveryListener[E comparable, T any] func(event E, listener interface{}, panicValue interface{})`: Recovery handler signature
- `Logger`: Interface with `Warnf(format string, args ...any)`

## Design Notes

### Why Dual Generics?

Traditional single-generic designs force event identifiers and parameters to share a type. Dual generics let you use simple identifiers (strings, ints, enums) while passing complex data structures.

### Single-Value Listener

`Listener[T]` takes a single `T` value rather than variadic `...T`. This keeps listeners clean — use a struct when you need multiple fields.

### Global Concurrency Control

`SetConcurrency(n)` creates a semaphore shared by `Emit` and `EmitWait` calls that start after the setting changes. When it is full, internal dispatch waits for a slot; `Emit` itself remains fire-and-forget and returns immediately. In-flight emissions keep using the semaphore captured when they started.

An `EmitWait` listener must not synchronously call `EmitWait` on the same emitter when all permits may already be occupied; doing so would wait for its own permit. Use `Emit`, `EmitSync`, or a separate emitter for that reentrant path.

### Panic Recovery

Panic recovery is always active. With a custom `RecoveryListener`, panics are routed there; otherwise they are logged when a `Logger` is configured. With neither configured, the panic is safely recovered and discarded. Panics raised by the recovery callback or logger are isolated as well.

## Inspiration

Inspired by [chuckpreslar/emission](https://github.com/chuckpreslar/emission), extended with dual-generic design, real semaphore-based concurrency control, and always-on panic recovery.

## License

MIT License
