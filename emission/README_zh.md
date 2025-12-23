# Emission - Go 的泛型事件发射器

[English](./README.md) | 简体中文

`emission` 是一个强大的 Go 泛型事件发射器，为事件驱动编程提供发布-订阅模式。它采用双泛型设计，分离事件标识类型和监听器参数类型，提供更灵活的类型安全。

## 特性

- **双泛型设计**: 事件标识类型和监听器参数类型完全分离，提供最大灵活性
- **类型安全**: 使用 Go 泛型（Go 1.18+）确保编译时类型检查
- **异步和同步**: 支持异步和同步事件触发
- **一次性监听器**: 内置一次性事件监听器支持
- **Panic 恢复**: 可选的监听器函数 panic 恢复机制
- **线程安全**: 支持并发安全使用
- **最大监听器数**: 可配置的限制以检测潜在的内存泄漏

## 安装

```bash
go get github.com/wsshow/op/emission
```

## 核心概念

`Emitter[E comparable, T any]` 使用两个泛型参数：

- **E**: 事件标识类型（必须是 comparable，如 string、int 等）
- **T**: 监听器参数类型（可以是任意类型）

这种设计允许你用简单的类型（如字符串）标识事件，同时传递复杂的数据结构给监听器。

## 使用示例

### 基本事件处理

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    // 创建一个发射器：事件名为 string，参数为 string
    emitter := emission.NewEmitter[string, string]()

    // 添加监听器
    emitter.On("message", func(args ...string) {
        fmt.Println("接收到:", args)
    })

    // 同步触发事件
    emitter.EmitSync("message", "Hello", "World")
    // 输出: 接收到: [Hello World]
}
```

### 使用自定义数据结构

```go
type User struct {
    Name string
    Age  int
}

// 事件名为 string，参数为 User 结构体
emitter := emission.NewEmitter[string, User]()

emitter.On("user_login", func(users ...User) {
    for _, user := range users {
        fmt.Printf("用户 %s (年龄: %d) 已登录\n", user.Name, user.Age)
    }
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

emitter.On("message", func(msgs ...Message) {
    for _, msg := range msgs {
        fmt.Printf("处理消息 #%d: %s\n", msg.ID, msg.Content)
    }
})

// 异步触发（非阻塞）
emitter.Emit("message", Message{ID: 1, Content: "Hello"})
emitter.Emit("message", Message{ID: 2, Content: "World"})
```

### 一次性监听器

```go
emitter := emission.NewEmitter[string, string]()

// 监听器将在首次触发后自动移除
emitter.Once("startup", func(args ...string) {
    fmt.Println("应用程序已启动！")
})

emitter.EmitSync("startup") // 输出: 应用程序已启动！
emitter.EmitSync("startup") // 无输出（监听器已被移除）
```

### 移除所有监听器

```go
emitter := emission.NewEmitter[string, int]()

emitter.On("event", func(args ...int) { fmt.Println("Listener 1") })
emitter.On("event", func(args ...int) { fmt.Println("Listener 2") })

fmt.Println("监听器数量:", emitter.GetListenerCount("event")) // 输出: 2

emitter.RemoveAllListeners("event")
fmt.Println("监听器数量:", emitter.GetListenerCount("event")) // 输出: 0
```

### Panic 恢复

```go
type ErrorData struct {
    Code    int
    Message string
}

emitter := emission.NewEmitter[string, ErrorData]()

// 设置恢复处理器
emitter.RecoverWith(func(event string, listener interface{}, err error) {
    fmt.Printf("事件 '%s' 中的 Panic: %v\n", event, err)
})

// 添加会 panic 的监听器
emitter.On("error", func(args ...ErrorData) {
    panic("出错了！")
})

emitter.EmitSync("error", ErrorData{Code: 500, Message: "Internal Error"})
// Panic 被捕获并记录
```

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

emitter.On(EventStart, func(states ...AppState) {
    fmt.Printf("应用启动，状态: %s\n", states[0].Status)
})

emitter.On(EventStop, func(states ...AppState) {
    fmt.Printf("应用停止，状态: %s\n", states[0].Status)
})

emitter.EmitSync(EventStart, AppState{Timestamp: 1234567890, Status: "running"})
emitter.EmitSync(EventStop, AppState{Timestamp: 1234567900, Status: "stopped"})
```

### 自定义事件类型

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
    fmt.Printf("用户 %s (ID: %s) 从 %s 登录\n", evt.Username, evt.UserID, evt.IP)
})

emitter.On(UserLogout, func(events ...UserEvent) {
    evt := events[0]
    fmt.Printf("用户 %s (ID: %s) 已登出\n", evt.Username, evt.UserID)
})

emitter.EmitSync(UserLogin, UserEvent{
    UserID:   "u123",
    Username: "alice",
    IP:       "192.168.1.1",
})
```

### 链式调用

```go
emitter := emission.NewEmitter[string, int]()

emitter.
    SetMaxListeners(20).
    On("event1", func(args ...int) { fmt.Println("Event 1") }).
    On("event2", func(args ...int) { fmt.Println("Event 2") }).
    Once("event3", func(args ...int) { fmt.Println("Event 3") }).
    EmitSync("event1", 1, 2, 3)
```

## API 概览

### 创建

- `NewEmitter[E comparable, T any]() *Emitter[E, T]`: 创建新的事件发射器
  - `E`: 事件标识类型（必须可比较）
  - `T`: 监听器参数类型（任意类型）

### 添加监听器

- `On(event E, listener Listener[T]) *Emitter[E, T]`: 添加监听器（别名：AddListener）
- `Once(event E, listener Listener[T]) *Emitter[E, T]`: 添加一次性监听器

### 移除监听器

- `RemoveAllListeners(event E) *Emitter[E, T]`: 移除指定事件的所有监听器
- `Off(event E, listener Listener[T]) *Emitter[E, T]`: 保留用于 API 兼容性（由于 Go 函数比较限制，实际不工作）

### 触发事件

- `Emit(event E, args ...T) *Emitter[E, T]`: 异步触发事件
- `EmitSync(event E, args ...T) *Emitter[E, T]`: 同步触发事件

### 配置

- `SetMaxListeners(max int) *Emitter[E, T]`: 设置每个事件的最大监听器数（-1 表示无限制）
- `GetListenerCount(event E) int`: 获取事件的监听器数量
- `RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T]`: 设置 panic 恢复处理器

### 类型

- `Listener[T any] func(args ...T)`: 监听器函数签名
- `RecoveryListener[E comparable, T any] func(event E, listener interface{}, err error)`: 恢复处理器签名

## 设计说明

### 为什么使用双泛型？

传统的单泛型设计会强制事件标识和参数使用相同的类型：

```go
// 单泛型设计的限制
emitter := NewEmitter[string]()
emitter.On("login", func(args ...string) {
    // 只能接收 string 参数，无法处理复杂的 User 对象
})
```

双泛型设计解决了这个问题：

```go
// 双泛型设计的灵活性
emitter := NewEmitter[string, User]()
emitter.On("login", func(users ...User) {
    // 可以使用简单的字符串标识事件，同时传递复杂的数据结构
})
```

### 函数移除的限制

由于 Go 语言的限制，函数值不能直接比较，因此 `RemoveListener` 和 `Off` 方法无法准确识别要移除的监听器。建议使用 `RemoveAllListeners` 清除指定事件的所有监听器。

## 注意事项

- **事件类型**: 事件标识类型 `E` 必须是可比较的（string、int、自定义可比较类型）
- **参数类型**: 监听器参数类型 `T` 可以是任意类型
- **异步 vs 同步**: `Emit()` 在 goroutine 中运行监听器，`EmitSync()` 顺序执行
- **内存安全**: Once 监听器在执行后自动移除
- **Panic 处理**: 设置恢复监听器以捕获事件处理器中的 panic
- **线程安全**: 所有操作都由互斥锁保护，支持并发使用
- **函数比较**: Go 中函数无法直接比较，建议使用 `RemoveAllListeners` 而非 `RemoveListener`

## 最佳实践

1. **使用有意义的事件名称**: 使用清晰的字符串或常量作为事件标识
2. **定义事件类型**: 为复杂应用定义自定义事件类型枚举
3. **使用结构体**: 为复杂数据使用结构体而非多个参数
4. **设置恢复处理器**: 在生产环境中始终设置 panic 恢复处理器
5. **监控监听器数量**: 使用 `GetListenerCount` 检测内存泄漏
6. **清理监听器**: 不再需要时及时移除监听器

## 参考来源

本实现受 [chuckpreslar/emission](https://github.com/chuckpreslar/emission) 启发，并扩展了双泛型设计以提供更好的类型安全性和灵活性。

## 许可证

MIT License
