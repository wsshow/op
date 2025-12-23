package op

import (
	"github.com/wsshow/op/deque"
	"github.com/wsshow/op/emission"
	"github.com/wsshow/op/linq"
	"github.com/wsshow/op/process"
	"github.com/wsshow/op/slice"
	"github.com/wsshow/op/str"
	"github.com/wsshow/op/workerpool"
)

// NewString 创建一个新的字符串对象
// 参数 s: 初始字符串值
func NewString(s string) *str.String {
	return str.NewString(s)
}

// NewSlice 创建一个新的泛型切片对象
// 参数 values: 可变参数，表示初始值
func NewSlice[T any](values ...T) *slice.Slice[T] {
	return slice.New(values...)
}

// NewDeque 创建一个新的泛型双端队列对象
func NewDeque[T any]() *deque.Deque[T] {
	return deque.New[T]()
}

// NewEmitter 创建一个新的事件发射器对象
// E: 事件标识类型（必须是 comparable），T: 监听器参数类型（任意类型）
func NewEmitter[E comparable, T any]() *emission.Emitter[E, T] {
	return emission.NewEmitter[E, T]()
}

// LinqFrom 从切片创建一个 Linq 对象，用于链式查询
// 参数 arr: 初始切片，类型需满足 comparable 约束
func LinqFrom[T comparable](arr []T) linq.Linq[T] {
	return linq.From(arr)
}

// LinqFromAny 从任意类型切片创建一个 Linq 对象
// 参数 arr: 初始切片，无需 comparable 约束
func LinqFromAny[T any](arr []T) linq.Linq[T] {
	return linq.From(arr)
}

// NewProcess 创建一个新的进程对象
// 参数 co: 进程配置选项
func NewProcess(co process.CmdOptions) *process.Process {
	return process.NewProcess(co)
}

// NewProcessManager 创建一个新的进程管理器对象
func NewProcessManager() *process.ProcessManager {
	return process.NewProcessManager()
}

// NewWorkerPool 创建一个新的工作池对象
// 参数 maxWorkers: 最大工作线程数
func NewWorkerPool(maxWorkers int) *workerpool.WorkerPool {
	return workerpool.New(maxWorkers)
}
