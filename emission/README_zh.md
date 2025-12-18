# Emission - Go 的泛型事件发射器

[English](./README.md) | 简体中文

`emission` 是一个简单而强大的 Go 泛型事件发射器，为事件驱动编程提供发布-订阅模式，支持异步和同步事件处理。

## 特性

- **泛型支持**: 使用 Go 泛型（Go 1.18+）实现类型安全的事件处理
- **可比较事件**: 事件类型必须是可比较的（字符串、整数、枚举等）
- **异步和同步**: 支持异步和同步事件触发
- **一次性监听器**: 内置一次性事件监听器支持
- **Panic 恢复**: 可选的监听器函数 panic 恢复机制
- **线程安全**: 支持并发安全使用
- **最大监听器数**: 可配置的限制以检测潜在的内存泄漏

## 安装

```bash
go get github.com/wsshow/op/emission
```

## 使用示例

### 基本事件处理

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/emission"
)

func main() {
    // 创建一个使用字符串事件名称的发射器
    emitter := emission.NewEmitter[string]()
    
    // 添加监听器
    emitter.On("message", func(args ...string) {
        fmt.Println("接收到:", args)
    })
    
    // 同步触发事件
    emitter.EmitSync("message", "Hello", "World")
    // 输出: 接收到: [Hello World]
}
```

### 异步事件触发

```go
emitter := emission.NewEmitter[string]()

emitter.On("data", func(args ...string) {
    fmt.Println("处理中:", args[0])
})

// 异步触发（非阻塞）
emitter.Emit("data", "item1")
emitter.Emit("data", "item2")
```

### 一次性监听器

```go
emitter := emission.NewEmitter[string]()

// 监听器将在首次触发后自动移除
emitter.Once("startup", func(args ...string) {
    fmt.Println("应用程序已启动！")
})

emitter.EmitSync("startup") // 输出: 应用程序已启动！
emitter.EmitSync("startup") // 无输出（监听器已被移除）
```

### 移除监听器

```go
emitter := emission.NewEmitter[string]()

handler := func(args ...string) {
    fmt.Println("事件已触发")
}

emitter.On("test", handler)
emitter.EmitSync("test") // 输出: 事件已触发

emitter.Off("test", handler)
emitter.EmitSync("test") // 无输出（监听器已移除）
```

### Panic 恢复

```go
emitter := emission.NewEmitter[string]()

// 设置恢复处理器
emitter.RecoverWith(func(event string, listener interface{}, err error) {
    fmt.Printf("事件 '%s' 中的 Panic: %v\n", event, err)
})

// 添加会 panic 的监听器
emitter.On("error", func(args ...string) {
    panic("出错了！")
})

emitter.EmitSync("error") // Panic 被捕获并记录
// 输出: 事件 'error' 中的 Panic: panic occurred in listener: 出错了！
```

### 使用整数事件类型

```go
const (
    EventStart = iota
    EventStop
    EventPause
)

emitter := emission.NewEmitter[int]()

emitter.On(EventStart, func(args ...int) {
    fmt.Println("已启动")
})

emitter.On(EventStop, func(args ...int) {
    fmt.Println("已停止")
})

emitter.EmitSync(EventStart) // 输出: 已启动
emitter.EmitSync(EventStop)  // 输出: 已停止
```

### 自定义事件类型

```go
type EventType string

const (
    UserLogin    EventType = "user:login"
    UserLogout   EventType = "user:logout"
    DataReceived EventType = "data:received"
)

emitter := emission.NewEmitter[EventType]()

emitter.On(UserLogin, func(args ...EventType) {
    fmt.Println("用户已登录")
})

emitter.On(UserLogout, func(args ...EventType) {
    fmt.Println("用户已登出")
})

emitter.EmitSync(UserLogin)  // 输出: 用户已登录
emitter.EmitSync(UserLogout) // 输出: 用户已登出
```

## API 概览

### 创建
- `NewEmitter[T comparable]() *Emitter[T]`: 创建新的事件发射器

### 添加监听器
- `On(event T, listener Listener[T]) *Emitter[T]`: 添加监听器（别名：AddListener）
- `Once(event T, listener Listener[T]) *Emitter[T]`: 添加一次性监听器

### 移除监听器
- `Off(event T, listener Listener[T]) *Emitter[T]`: 移除监听器（别名：RemoveListener）

### 触发事件
- `Emit(event T, args ...T) *Emitter[T]`: 异步触发事件
- `EmitSync(event T, args ...T) *Emitter[T]`: 同步触发事件

### 配置
- `SetMaxListeners(max int) *Emitter[T]`: 设置每个事件的最大监听器数（-1 表示无限制）
- `GetListenerCount(event T) int`: 获取事件的监听器数量
- `RecoverWith(listener RecoveryListener[T]) *Emitter[T]`: 设置 panic 恢复处理器

### 类型
- `Listener[T comparable] func(args ...T)`: 监听器函数签名
- `RecoveryListener[T comparable] func(event T, listener interface{}, err error)`: 恢复处理器签名

## 注意事项

- **事件类型**: 事件类型 `T` 必须是可比较的（string、int、自定义可比较类型）
- **异步 vs 同步**: `Emit()` 在 goroutine 中运行监听器，`EmitSync()` 顺序执行
- **内存安全**: Once 监听器在执行后自动移除
- **Panic 处理**: 设置恢复监听器以捕获事件处理器中的 panic
- **线程安全**: 所有操作都由互斥锁保护，支持并发使用

## 参考来源

本实现受 [chuckpreslar/emission](https://github.com/chuckpreslar/emission) 启发。

## 许可证

MIT License
