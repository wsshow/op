# Emission - Go 的泛型事件发射器

[English](./README.md) | 简体中文

`emission` 是一个强大的 Go 泛型事件发射器，为事件驱动编程提供发布-订阅模式。它采用双泛型设计，分离事件标识类型和监听器参数类型，提供更灵活的类型安全。

## 特性

- **双泛型设计**: 事件标识类型和监听器参数类型完全分离，提供最大灵活性
- **类型安全**: 使用 Go 泛型（模块要求 Go 1.24+）确保编译时类型检查
- **三种触发模式**: `Emit`（异步即弃）、`EmitWait`（并发等待）、`EmitSync`（同步顺序）
- **一次性监听器**: 内置一次性事件监听器，触发后自动移除
- **Panic 恢复**: 始终隔离监听器 panic，并可选交给恢复回调或日志处理
- **全局并发控制**: 通过共享信号量限制同时执行的异步监听器数量
- **线程安全**: 互斥锁保护所有操作，支持并发使用
- **最大监听器告警**: 可配置的软限制，用于检测潜在内存泄漏（Node.js 风格）

## 安装

```bash
go get github.com/wsshow/op/emission
```

## 核心概念

`Emitter[E comparable, T any]` 使用两个泛型参数：

- **E**: 事件标识类型（必须是 comparable，如 string、int 等）
- **T**: 监听器参数类型（可以是任意类型）

这种设计允许你用简单的类型标识事件，同时传递复杂的数据结构给监听器。

## 使用示例

### 基本事件处理

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    emitter := emission.NewEmitter[string, string]()

    // 添加监听器
    emitter.On("message", func(msg string) {
        fmt.Println("接收到:", msg)
    })

    // 同步触发事件
    emitter.EmitSync("message", "Hello, World")
    // 输出: 接收到: Hello, World
}
```

### 使用自定义数据结构

```go
type User struct {
    Name string
    Age  int
}

emitter := emission.NewEmitter[string, User]()

emitter.On("user_login", func(u User) {
    fmt.Printf("用户 %s (年龄: %d) 已登录\n", u.Name, u.Age)
})

emitter.EmitSync("user_login", User{Name: "Alice", Age: 30})
// 输出: 用户 Alice (年龄: 30) 已登录
```

### 异步事件触发

```go
type Message struct {
    ID      int
    Content string
}

emitter := emission.NewEmitter[string, Message]()

emitter.On("message", func(m Message) {
    fmt.Printf("处理消息 #%d: %s\n", m.ID, m.Content)
})

// 异步触发（非阻塞）
emitter.Emit("message", Message{ID: 1, Content: "Hello"})
emitter.Emit("message", Message{ID: 2, Content: "World"})
```

### 并发触发并等待

```go
emitter := emission.NewEmitter[string, string]()
emitter.SetConcurrency(4) // 全局最多同时执行 4 个监听器

emitter.On("process", func(data string) {
    // 耗时操作
})

// 阻塞直到所有监听器执行完成
emitter.EmitWait("process", "data")
```

### 一次性监听器

```go
emitter := emission.NewEmitter[string, string]()

// 监听器将在首次触发后自动移除
emitter.Once("startup", func(msg string) {
    fmt.Println("应用程序已启动！")
})

emitter.EmitSync("startup", "ready") // 输出: 应用程序已启动！
emitter.EmitSync("startup", "ready") // 无输出（监听器已被移除）
```

### 通过 Subscription 取消监听

```go
emitter := emission.NewEmitter[string, int]()

sub := emitter.On("event", func(n int) { fmt.Println(n) })
emitter.EmitSync("event", 1) // 输出: 1

sub.Unsubscribe()
emitter.EmitSync("event", 2) // 无输出
```

### 移除所有监听器

```go
emitter := emission.NewEmitter[string, int]()

emitter.On("event", func(n int) { fmt.Println("L1") })
emitter.On("event", func(n int) { fmt.Println("L2") })

fmt.Println(emitter.GetListenerCount("event")) // 2

emitter.RemoveAllListeners("event")
fmt.Println(emitter.GetListenerCount("event")) // 0
```

### Panic 恢复

```go
emitter := emission.NewEmitter[string, string]()

// 设置自定义恢复处理器
emitter.RecoverWith(func(event string, listener interface{}, panicValue interface{}) {
    fmt.Printf("事件 '%s' 中的 Panic: %v\n", event, panicValue)
})

emitter.On("error", func(msg string) {
    panic("出错了！")
})

emitter.EmitSync("error", "data")
// Panic 被捕获并处理
```

未设置 RecoverWith 时，panic 仍然会被安全恢复，若配置了 Logger 则会记录日志。

### 使用整数事件标识

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
    fmt.Printf("应用启动，状态: %s\n", s.Status)
})

emitter.On(EventStop, func(s AppState) {
    fmt.Printf("应用停止，状态: %s\n", s.Status)
})

emitter.EmitSync(EventStart, AppState{Timestamp: 1234567890, Status: "running"})
```

### 自定义事件类型

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
    fmt.Printf("用户 %s 从 %s 登录\n", e.Username, e.IP)
})

emitter.EmitSync(UserLogin, UserEvent{
    UserID: "u123", Username: "alice", IP: "192.168.1.1",
})
```

### 链式调用

```go
emitter := emission.NewEmitter[string, int]()

emitter.
    SetMaxListeners(20).
    SetConcurrency(5).
    RecoverWith(func(event string, listener interface{}, panicValue interface{}) {}).
    EmitSync("init", 0)

// On/Once 返回 *Subscription 用于单独取消
sub1 := emitter.On("event1", func(n int) { fmt.Println("Event 1:", n) })
sub2 := emitter.On("event2", func(n int) { fmt.Println("Event 2:", n) })

emitter.EmitSync("event1", 1).EmitSync("event2", 2)

sub1.Unsubscribe()
sub2.Unsubscribe()
```

### 内省

```go
emitter := emission.NewEmitter[string, int]()
emitter.On("a", func(n int) {})
emitter.On("a", func(n int) {})
emitter.On("b", func(n int) {})

fmt.Println(emitter.Events())              // 包含 a 和 b；顺序不确定
fmt.Println(emitter.GetListenerCount("a")) // 2
fmt.Println(emitter.TotalListenerCount())  // 3
```

## API 概览

### 创建

- `NewEmitter[E comparable, T any]() *Emitter[E, T]`: 创建新的事件发射器

### 添加监听器

- `On(event E, listener Listener[T]) *Subscription[E, T]`: 添加监听器
- `AddListener(event E, listener Listener[T]) *Subscription[E, T]`: On 的别名
- `Once(event E, listener Listener[T]) *Subscription[E, T]`: 添加一次性监听器
- `Subscription.Unsubscribe()`: 取消订阅，移除监听器

### 移除监听器

- `RemoveAllListeners(event E) *Emitter[E, T]`: 移除指定事件的所有监听器

### 触发事件

- `Emit(event E, value T) *Emitter[E, T]`: 异步即弃（立即返回）
- `EmitWait(event E, value T) *Emitter[E, T]`: 异步等待（阻塞直到所有监听器完成）
- `EmitSync(event E, value T) *Emitter[E, T]`: 同步顺序执行

### 内省

- `GetListenerCount(event E) int`: 获取事件的监听器数量
- `Events() []E`: 以不确定顺序列出所有已注册的事件标识
- `TotalListenerCount() int`: 获取所有事件的监听器总数

### 配置

- `SetMaxListeners(max int) *Emitter[E, T]`: 设置每事件告警阈值（-1 无限制）
- `SetConcurrency(n int) *Emitter[E, T]`: 设置全局最大并发执行数（0 无限制）
- `SetLogger(logger Logger) *Emitter[E, T]`: 设置日志记录器
- `RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T]`: 设置自定义 panic 恢复处理器

### 类型

- `Listener[T any] func(T)`: 监听器函数签名
- `RecoveryListener[E comparable, T any] func(event E, listener interface{}, panicValue interface{})`: 恢复处理器签名
- `Logger`: 包含 `Warnf(format string, args ...any)` 方法的接口

## 设计说明

### 为什么使用双泛型？

传统单泛型设计会强制事件标识和参数使用相同的类型。双泛型设计允许用简单的标识类型（string、int、枚举）来区分事件，同时传递复杂数据结构。

### 单参数监听器

`Listener[T]` 接收单个 `T` 值而非可变参数 `...T`。这使得监听器签名更简洁 —— 需要多个字段时使用 struct 即可。

### 全局并发控制

`SetConcurrency(n)` 创建信号量，由设置变更后开始的 `Emit` 和 `EmitWait` 调用共享。信号量满时内部调度会等待空位；`Emit` 本身仍是即发即忘并立即返回。已经开始的触发继续使用启动时捕获的信号量。

如果所有许可都可能已被占用，`EmitWait` 的监听器不能在同一个 emitter 上同步重入调用 `EmitWait`，否则会等待自己持有的许可。该路径应使用 `Emit`、`EmitSync` 或独立的 emitter。

### Panic 恢复

Panic 恢复始终开启。设置自定义 `RecoveryListener` 时，panic 会路由到该处理器；否则在配置了 `Logger` 时记录日志。两者都未配置时，panic 会被安全恢复并丢弃。恢复回调或 logger 自身的 panic 也会被隔离。

## 参考来源

受 [chuckpreslar/emission](https://github.com/chuckpreslar/emission) 启发，扩展了双泛型设计、真正的信号量并发控制和始终开启的 panic 恢复。

## 许可证

MIT License
