package op

import (
	"github.com/wsshow/op/deque"
	"github.com/wsshow/op/emission"
	"github.com/wsshow/op/linq"
	"github.com/wsshow/op/slice"
	"github.com/wsshow/op/str"
	"github.com/wsshow/op/workerpool"
)

// 创建一个字符串对象
func NewString(s string) *str.String {
	return str.NewString(s)
}

// 创建一个切片对象
func NewSlice[T any](values ...T) *slice.Slice[T] {
	return slice.New(values...)
}

// 创建一个双端队列对象
// size 0: 最大容量 1: 最小容量(容量缩放时保证容量不小于最小容量)
func NewDeque[T any](size ...int) *deque.Deque[T] {
	return deque.New[T](size...)
}

// 创建一个工作池对象
func NewWorkerPool(maxWorkers int) *workerpool.WorkerPool {
	return workerpool.New(maxWorkers)
}

// 创建一个事件发射器对象
func NewEmitter() *emission.Emitter {
	return emission.NewEmitter()
}

// 创建一个Linq对象
func LinqFrom[T comparable](arr []T) linq.Linq[T] {
	return linq.From(arr)
}
