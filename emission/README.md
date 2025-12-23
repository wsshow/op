# Emission - A Generic Event Emitter for Go

English | [简体中文](README_zh.md)

`emission` is a powerful generic event emitter for Go, providing a publish-subscribe pattern for event-driven programming. It features a dual-generic design that separates event identifier types from listener parameter types for maximum flexibility and type safety.

## Features

- **Dual-Generic Design**: Complete separation of event identifier and listener parameter types for maximum flexibility
- **Type Safety**: Compile-time type checking with Go generics (Go 1.18+)
- **Async & Sync**: Support for both asynchronous and synchronous event emission
- **Once Listeners**: Built-in support for one-time event listeners
- **Panic Recovery**: Optional panic recovery for listener functions
- **Thread-Safe**: Safe for concurrent use
- **Max Listeners**: Configurable limit to detect potential memory leaks

## Installation

```bash
go get github.com/wsshow/op/emission
```

## Core Concept

`Emitter[E comparable, T any]` uses two generic parameters:

- **E**: Event identifier type (must be comparable, like string, int, etc.)
- **T**: Listener parameter type (can be any type)

This design allows you to identify events with simple types (like strings) while passing complex data structures to listeners.

## Usage Examples

### Basic Event Handling

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    // Create an emitter: event names are strings, parameters are strings
    emitter := emission.NewEmitter[string, string]()

    // Add a listener
    emitter.On("message", func(args ...string) {
        fmt.Println("Received:", args)
    })

    // Emit event synchronously
    emitter.EmitSync("message", "Hello", "World")
    // Output: Received: [Hello World]
}
```

### Using Custom Data Structures

```go
type User struct {
    Name string
    Age  int
}

// Event names are strings, parameters are User structs
emitter := emission.NewEmitter[string, User]()

emitter.On("user_login", func(users ...User) {
    for _, user := range users {
        fmt.Printf("User %s (age: %d) logged in\n", user.Name, user.Age)
    }
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

emitter.On("message", func(msgs ...Message) {
    for _, msg := range msgs {
        fmt.Printf("Processing message #%d: %s\n", msg.ID, msg.Content)
    }
})

// Emit asynchronously (non-blocking)
emitter.Emit("message", Message{ID: 1, Content: "Hello"})
emitter.Emit("message", Message{ID: 2, Content: "World"})
```

### Once Listeners

```go
emitter := emission.NewEmitter[string, string]()

// Listener will be automatically removed after first trigger
emitter.Once("startup", func(args ...string) {
    fmt.Println("Application started!")
})

emitter.EmitSync("startup") // Output: Application started!
emitter.EmitSync("startup") // No output (listener removed)
```

### Removing All Listeners

```go
emitter := emission.NewEmitter[string, int]()

emitter.On("event", func(args ...int) { fmt.Println("Listener 1") })
emitter.On("event", func(args ...int) { fmt.Println("Listener 2") })

fmt.Println("Listener count:", emitter.GetListenerCount("event")) // Output: 2

emitter.RemoveAllListeners("event")
fmt.Println("Listener count:", emitter.GetListenerCount("event")) // Output: 0
```

### Panic Recovery

```go
type ErrorData struct {
    Code    int
    Message string
}

emitter := emission.NewEmitter[string, ErrorData]()

// Set recovery handler
emitter.RecoverWith(func(event string, listener interface{}, err error) {
    fmt.Printf("Panic in event '%s': %v\n", event, err)
})

// Add a listener that panics
emitter.On("error", func(args ...ErrorData) {
    panic("Something went wrong!")
})

emitter.EmitSync("error", ErrorData{Code: 500, Message: "Internal Error"})
// Panic is caught and logged
```

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

emitter.On(EventStart, func(states ...AppState) {
    fmt.Printf("App started, status: %s\n", states[0].Status)
})

emitter.On(EventStop, func(states ...AppState) {
    fmt.Printf("App stopped, status: %s\n", states[0].Status)
})

emitter.EmitSync(EventStart, AppState{Timestamp: 1234567890, Status: "running"})
emitter.EmitSync(EventStop, AppState{Timestamp: 1234567900, Status: "stopped"})
```

### Custom Event Types

```go
type EventType string

const (
    UserLogin    EventType = "user:login"
    UserLogout   EventType = "user:logout"
    DataReceived EventType = "data:received"
)

type UserEvent struct {
    UserID   string
    Username string
    IP       string
}

emitter := emission.NewEmitter[EventType, UserEvent]()

emitter.On(UserLogin, func(events ...UserEvent) {
    evt := events[0]
    fmt.Printf("User %s (ID: %s) logged in from %s\n", evt.Username, evt.UserID, evt.IP)
})

emitter.On(UserLogout, func(events ...UserEvent) {
    evt := events[0]
    fmt.Printf("User %s (ID: %s) logged out\n", evt.Username, evt.UserID)
})

emitter.EmitSync(UserLogin, UserEvent{
    UserID:   "u123",
    Username: "alice",
    IP:       "192.168.1.1",
})
```

### Method Chaining

```go
emitter := emission.NewEmitter[string, int]()

emitter.
    SetMaxListeners(20).
    On("event1", func(args ...int) { fmt.Println("Event 1") }).
    On("event2", func(args ...int) { fmt.Println("Event 2") }).
    Once("event3", func(args ...int) { fmt.Println("Event 3") }).
    EmitSync("event1", 1, 2, 3)
```

## API Overview

### Creation

- `NewEmitter[E comparable, T any]() *Emitter[E, T]`: Create a new event emitter
  - `E`: Event identifier type (must be comparable)
  - `T`: Listener parameter type (can be any type)

### Adding Listeners

- `On(event E, listener Listener[T]) *Emitter[E, T]`: Add a listener (alias: AddListener)
- `Once(event E, listener Listener[T]) *Emitter[E, T]`: Add a one-time listener

### Removing Listeners

- `RemoveAllListeners(event E) *Emitter[E, T]`: Remove all listeners for a specific event
- `Off(event E, listener Listener[T]) *Emitter[E, T]`: Kept for API compatibility (doesn't work due to Go function comparison limitations)

### Emitting Events

- `Emit(event E, args ...T) *Emitter[E, T]`: Emit event asynchronously
- `EmitSync(event E, args ...T) *Emitter[E, T]`: Emit event synchronously

### Configuration

- `SetMaxListeners(max int) *Emitter[E, T]`: Set maximum number of listeners per event (-1 for unlimited)
- `GetListenerCount(event E) int`: Get the number of listeners for an event
- `RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T]`: Set panic recovery handler

### Types

- `Listener[T any] func(args ...T)`: Listener function signature
- `RecoveryListener[E comparable, T any] func(event E, listener interface{}, err error)`: Recovery handler signature

## Design Rationale

### Why Dual Generics?

Traditional single-generic designs force event identifiers and parameters to use the same type:

```go
// Single-generic design limitation
emitter := NewEmitter[string]()
emitter.On("login", func(args ...string) {
    // Can only receive string parameters, can't handle complex User objects
})
```

Dual-generic design solves this problem:

```go
// Dual-generic design flexibility
emitter := NewEmitter[string, User]()
emitter.On("login", func(users ...User) {
    // Can use simple strings to identify events while passing complex data structures
})
```

### Function Removal Limitations

Due to Go language limitations, function values cannot be directly compared. Therefore, `RemoveListener` and `Off` methods cannot accurately identify which listener to remove. It's recommended to use `RemoveAllListeners` to clear all listeners for a specific event.

## Important Notes

- **Event Type**: Event identifier type `E` must be comparable (string, int, custom comparable types)
- **Parameter Type**: Listener parameter type `T` can be any type
- **Async vs Sync**: `Emit()` runs listeners in goroutines, `EmitSync()` executes sequentially
- **Memory Safety**: Once listeners are automatically removed after execution
- **Panic Handling**: Set a recovery listener to catch panics in event handlers
- **Thread Safety**: All operations are protected by mutexes for concurrent use
- **Function Comparison**: Functions cannot be directly compared in Go, use `RemoveAllListeners` instead of `RemoveListener`

## Best Practices

1. **Use Meaningful Event Names**: Use clear strings or constants as event identifiers
2. **Define Event Types**: Define custom event type enums for complex applications
3. **Use Structs**: Use structs for complex data instead of multiple parameters
4. **Set Recovery Handlers**: Always set panic recovery handlers in production
5. **Monitor Listener Count**: Use `GetListenerCount` to detect memory leaks
6. **Clean Up Listeners**: Remove listeners promptly when no longer needed

## Inspiration

This implementation is inspired by [chuckpreslar/emission](https://github.com/chuckpreslar/emission) and extends it with a dual-generic design for better type safety and flexibility.

## License

MIT License
