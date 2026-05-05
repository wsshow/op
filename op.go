// Package op 是工具库的统一入口，重导出各子包的核心类型与构造函数。
// 导入本包后无需单独导入子包即可使用全部功能。
package op

import (
	"github.com/wsshow/op/deque"
	"github.com/wsshow/op/emission"
	"github.com/wsshow/op/generator"
	"github.com/wsshow/op/linq"
	"github.com/wsshow/op/process"
	"github.com/wsshow/op/slice"
	"github.com/wsshow/op/str"
	"github.com/wsshow/op/workerpool"
)

// ---------------------------------------------------------------------------
// 类型别名 — 无需导入子包即可使用这些类型
// ---------------------------------------------------------------------------

// String 字符串包装器类型。
type String = str.String

// Slice 泛型切片包装器类型。
type Slice[T any] = slice.Slice[T]

// Deque 泛型双端队列类型。
type Deque[T any] = deque.Deque[T]

// Emitter 泛型事件发射器类型。
type Emitter[E comparable, T any] = emission.Emitter[E, T]

// Linq 泛型 LINQ 查询类型。
type Linq[T any] = linq.Linq[T]

// Process 外部进程管理器类型。
type Process = process.Process

// Manager 进程管理器（管理一组具名进程）。
type Manager = process.Manager

// Generator 泛型生成器类型。
type Generator[T any] = generator.Generator[T]

// Yield 生成器的产出句柄，用于向消费者发送值并接收回传结果。
type Yield[T any] = generator.Yield[T]

// WorkerPool 工作协程池类型。
type WorkerPool = workerpool.WorkerPool

// Option 工作协程池的可选配置函数。
type Option = workerpool.Option

// Options 进程启动配置。
type Options = process.Options

// ---------------------------------------------------------------------------
// 数据结构
// ---------------------------------------------------------------------------

// NewString 创建一个新的字符串对象。
func NewString(s string) *String { return str.New(s) }

// JoinStrings 使用分隔符连接字符串切片，返回新 String 对象。
func JoinStrings(elems []string, sep string) *String { return str.Join(elems, sep) }

// NewSlice 创建一个新的泛型切片对象，可传入初始值。
func NewSlice[T any](values ...T) *Slice[T] { return slice.New(values...) }

// NewDeque 创建一个新的泛型双端队列对象。
// 可选 capacity 参数设置初始基础容量。
func NewDeque[T any](capacity ...int) *Deque[T] { return deque.New[T](capacity...) }

// ---------------------------------------------------------------------------
// 事件
// ---------------------------------------------------------------------------

// NewEmitter 创建一个新的事件发射器对象。
// E 为事件标识类型，T 为监听器参数类型。
func NewEmitter[E comparable, T any]() *Emitter[E, T] {
	return emission.NewEmitter[E, T]()
}

// ---------------------------------------------------------------------------
// 查询
// ---------------------------------------------------------------------------

// LinqFrom 从切片创建一个 Linq 对象，用于链式查询。
func LinqFrom[T any](arr []T) Linq[T] { return linq.From(arr) }

// LinqEmpty 创建一个空的 Linq 对象。
func LinqEmpty[T any]() Linq[T] { return linq.Empty[T]() }

// LinqRange 生成从 start 开始的 count 个连续整数。
func LinqRange(start, count int) Linq[int] { return linq.Range(start, count) }

// LinqRepeat 生成 count 个重复的 value。
func LinqRepeat[T any](value T, count int) Linq[T] { return linq.Repeat(value, count) }

// ---------------------------------------------------------------------------
// 并发
// ---------------------------------------------------------------------------

// NewGenerator 创建并启动一个新的生成器，在后台 goroutine 中运行 genFunc。
func NewGenerator[T any](genFunc func(Yield[T])) *Generator[T] {
	return generator.NewGenerator(genFunc)
}

// NewWorkerPool 创建一个新的工作协程池。
// maxWorkers 指定最大并发工作协程数，opts 可选配置空闲超时等参数。
func NewWorkerPool(maxWorkers int, opts ...Option) *WorkerPool {
	return workerpool.New(maxWorkers, opts...)
}

// WithPanicHandler 设置工作协程池的 panic 处理器，当任务发生 panic 时调用。
func WithPanicHandler(handler func(any)) Option {
	return workerpool.WithPanicHandler(handler)
}

// ---------------------------------------------------------------------------
// 系统
// ---------------------------------------------------------------------------

// NewProcess 创建一个新的进程对象。
func NewProcess(co Options) *Process { return process.New(co) }

// NewProcessManager 创建一个新的进程管理器对象。
func NewProcessManager() *Manager { return process.NewManager() }
