# Emission - A Generic Event Emitter for Go

English | [简体中文](README_zh.md)

`emission` is a simple yet powerful generic event emitter for Go, providing a publish-subscribe pattern for event-driven programming with both async and sync event handling.

## Features

- **Generic Support**: Type-safe event handling with Go generics (Go 1.18+)
- **Comparable Events**: Event types must be comparable (strings, ints, enums, etc.)
- **Async & Sync**: Support for both asynchronous and synchronous event emission
- **Once Listeners**: Built-in support for one-time event listeners
- **Panic Recovery**: Optional panic recovery for listener functions
- **Thread-Safe**: Safe for concurrent use
- **Max Listeners**: Configurable limit to detect potential memory leaks

## Installation

```bash
go get github.com/wsshow/op/emission
```

## Usage Examples

### Basic Event Handling

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    // Create an emitter with string event names
    emitter := emission.NewEmitter[string]()
    
    // Add a listener
    emitter.On("message", func(args ...string) {
        fmt.Println("Received:", args)
    })
    
    // Emit event synchronously
    emitter.EmitSync("message", "Hello", "World")
    // Output: Received: [Hello World]
}
```

### Async Event Emission

```go
emitter := emission.NewEmitter[string]()

emitter.On("data", func(args ...string) {
    fmt.Println("Processing:", args[0])
})

// Emit asynchronously (non-blocking)
emitter.Emit("data", "item1")
emitter.Emit("data", "item2")
```

### One-Time Listeners

```go
emitter := emission.NewEmitter[string]()

// Listener will be automatically removed after first trigger
emitter.Once("startup", func(args ...string) {
    fmt.Println("Application started!")
})

emitter.EmitSync("startup") // Output: Application started!
emitter.EmitSync("startup") // No output (listener was removed)
```

### Removing Listeners

```go
emitter := emission.NewEmitter[string]()

handler := func(args ...string) {
    fmt.Println("Event triggered")
}

emitter.On("test", handler)
emitter.EmitSync("test") // Output: Event triggered

emitter.Off("test", handler)
emitter.EmitSync("test") // No output (listener removed)
```

### Panic Recovery

```go
emitter := emission.NewEmitter[string]()

// Set up recovery handler
emitter.RecoverWith(func(event string, listener interface{}, err error) {
    fmt.Printf("Panic in event '%s': %v\n", event, err)
})

// Add listener that panics
emitter.On("error", func(args ...string) {
    panic("Something went wrong!")
})

emitter.EmitSync("error") // Panic is caught and logged
// Output: Panic in event 'error': panic occurred in listener: Something went wrong!
```

### Using Integer Event Types

```go
const (
    EventStart = iota
    EventStop
    EventPause
)

emitter := emission.NewEmitter[int]()

emitter.On(EventStart, func(args ...int) {
    fmt.Println("Started")
})

emitter.On(EventStop, func(args ...int) {
    fmt.Println("Stopped")
})

emitter.EmitSync(EventStart) // Output: Started
emitter.EmitSync(EventStop)  // Output: Stopped
```

### Custom Event Types

```go
type EventType string

const (
    UserLogin    EventType = "user:login"
    UserLogout   EventType = "user:logout"
    DataReceived EventType = "data:received"
)

emitter := emission.NewEmitter[EventType]()

emitter.On(UserLogin, func(args ...EventType) {
    fmt.Println("User logged in")
})

emitter.On(UserLogout, func(args ...EventType) {
    fmt.Println("User logged out")
})

emitter.EmitSync(UserLogin)  // Output: User logged in
emitter.EmitSync(UserLogout) // Output: User logged out
```

### Listener Count and Max Listeners

```go
emitter := emission.NewEmitter[string]()

// Set maximum listeners (default is 10)
emitter.SetMaxListeners(5)

emitter.On("test", func(args ...string) {})
emitter.On("test", func(args ...string) {})

count := emitter.GetListenerCount("test")
fmt.Println("Listeners:", count) // Output: Listeners: 2
```

### Complex Example: Application Events

```go
type AppEvent string

const (
    AppInit     AppEvent = "app:init"
    AppReady    AppEvent = "app:ready"
    AppShutdown AppEvent = "app:shutdown"
)

type App struct {
    emitter *emission.Emitter[AppEvent]
}

func NewApp() *App {
    app := &App{
        emitter: emission.NewEmitter[AppEvent](),
    }
    
    // Set up panic recovery
    app.emitter.RecoverWith(func(event AppEvent, listener interface{}, err error) {
        fmt.Printf("[ERROR] Event %s failed: %v\n", event, err)
    })
    
    return app
}

func (a *App) On(event AppEvent, handler emission.Listener[AppEvent]) {
    a.emitter.On(event, handler)
}

func (a *App) Start() {
    a.emitter.EmitSync(AppInit)
    // ... initialization logic ...
    a.emitter.EmitSync(AppReady)
}

func (a *App) Stop() {
    a.emitter.EmitSync(AppShutdown)
}

func main() {
    app := NewApp()
    
    app.On(AppInit, func(args ...AppEvent) {
        fmt.Println("Initializing...")
    })
    
    app.On(AppReady, func(args ...AppEvent) {
        fmt.Println("Application ready!")
    })
    
    app.On(AppShutdown, func(args ...AppEvent) {
        fmt.Println("Shutting down...")
    })
    
    app.Start()
    // Output:
    // Initializing...
    // Application ready!
    
    app.Stop()
    // Output:
    // Shutting down...
}
```

## API Overview

### Creation
- `NewEmitter[T comparable]() *Emitter[T]`: Create a new event emitter

### Adding Listeners
- `On(event T, listener Listener[T]) *Emitter[T]`: Add a listener (alias: AddListener)
- `Once(event T, listener Listener[T]) *Emitter[T]`: Add a one-time listener

### Removing Listeners
- `Off(event T, listener Listener[T]) *Emitter[T]`: Remove a listener (alias: RemoveListener)

### Emitting Events
- `Emit(event T, args ...T) *Emitter[T]`: Emit event asynchronously
- `EmitSync(event T, args ...T) *Emitter[T]`: Emit event synchronously

### Configuration
- `SetMaxListeners(max int) *Emitter[T]`: Set max listeners per event (-1 for unlimited)
- `GetListenerCount(event T) int`: Get number of listeners for an event
- `RecoverWith(listener RecoveryListener[T]) *Emitter[T]`: Set panic recovery handler

### Types
- `Listener[T comparable] func(args ...T)`: Listener function signature
- `RecoveryListener[T comparable] func(event T, listener interface{}, err error)`: Recovery handler signature

## Notes

- **Event Types**: Event type `T` must be comparable (string, int, custom comparable types)
- **Async vs Sync**: `Emit()` runs listeners in goroutines, `EmitSync()` runs sequentially
- **Memory Safety**: Once listeners are automatically removed after execution
- **Panic Handling**: Set a recovery listener to catch panics in event handlers
- **Thread Safety**: All operations are protected by mutex for concurrent use

## Reference

This implementation is inspired by [chuckpreslar/emission](https://github.com/chuckpreslar/emission).

## License

MIT License

