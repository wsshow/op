package op

import (
	"github.com/wsshow/op/deque"
	"github.com/wsshow/op/emission"
	"github.com/wsshow/op/slice"
	"github.com/wsshow/op/str"
	"github.com/wsshow/op/workerpool"
)

/**
 * @Description: 创建一个字符串对象
 * @param s
 * @return *String
 */
func NewString(s string) *str.String {
	return str.NewString(s)
}

/**
 * @Description: 创建一个切片对象
 * @param values
 * @return *Slice
 */
func NewSlice[T any](values ...T) *slice.Slice[T] {
	return slice.New(values...)
}

/**
 * @Description: 创建一个双端队列对象
 * @param size 0: 最大容量 1: 最小容量(容量缩放时保证容量不小于最小容量)
 * @return *Deque
 */
func NewDeque[T any](size ...int) *deque.Deque[T] {
	return deque.New[T](size...)
}

/**
 * @Description: 创建一个工作池对象
 * @param maxWorkers
 * @return *WorkerPool
 */
func NewWorkerPool(maxWorkers int) *workerpool.WorkerPool {
	return workerpool.New(maxWorkers)
}

/**
 * @Description: 创建一个事件发射器对象
 * @return *Emitter
 */
func NewEmitter() *emission.Emitter {
	return emission.NewEmitter()
}
