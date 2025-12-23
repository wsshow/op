package emission

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// DefaultMaxListeners 默认的最大监听器数量
const DefaultMaxListeners = 10

// ErrNoneFunction 当监听器不是函数时返回的错误
var ErrNoneFunction = errors.New("listener must be a function")

// RecoveryListener 定义恢复监听器的签名，用于处理 panic
// E: 事件类型，T: 监听器参数类型
type RecoveryListener[E comparable, T any] func(event E, listener interface{}, err error)

// Listener 定义监听器函数的签名，接受泛型参数
type Listener[T any] func(args ...T)

// listenerWrapper 包装监听器并添加唯一标识
type listenerWrapper[T any] struct {
	id       uint64      // 唯一标识符
	listener Listener[T] // 实际的监听器函数
	isOnce   bool        // 是否为 Once 监听器
}

// Emitter 是一个泛型事件发射器，用于管理事件的监听和触发
// E: 事件标识类型（必须是 comparable），T: 监听器参数类型（可以是任意类型）
type Emitter[E comparable, T any] struct {
	mu           sync.Mutex                  // 互斥锁，确保线程安全
	events       map[E][]*listenerWrapper[T] // 事件到监听器列表的映射
	recoverer    RecoveryListener[E, T]      // 可选的恢复监听器，用于处理 panic
	maxListeners int                         // 每个事件的最大监听器数量，用于调试内存泄漏
	nextID       uint64                      // 下一个监听器的ID
}

// NewEmitter 创建一个新的泛型事件发射器
// E: 事件标识类型，T: 监听器参数类型
// 返回初始化好的 Emitter 实例，默认最大监听器数为 DefaultMaxListeners
func NewEmitter[E comparable, T any]() *Emitter[E, T] {
	return &Emitter[E, T]{
		events:       make(map[E][]*listenerWrapper[T]),
		maxListeners: DefaultMaxListeners,
		nextID:       1,
	}
}

// AddListener 添加监听器到指定事件
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 如果监听器数量超过 maxListeners，会打印警告
func (e *Emitter[E, T]) AddListener(event E, listener Listener[T]) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.maxListeners != -1 && len(e.events[event])+1 > e.maxListeners {
		fmt.Fprintf(os.Stdout, "Warning: event `%v` exceeds max listeners limit of %d\n", event, e.maxListeners)
	}

	id := e.nextID
	e.nextID++
	wrapper := &listenerWrapper[T]{
		id:       id,
		listener: listener,
		isOnce:   false,
	}
	e.events[event] = append(e.events[event], wrapper)
	return e
}

// On 是 AddListener 的别名，便于链式调用
func (e *Emitter[E, T]) On(event E, listener Listener[T]) *Emitter[E, T] {
	return e.AddListener(event, listener)
}

// RemoveListener 从指定事件中移除监听器
// 参数 event: 事件标识
// 参数 listener: 要移除的监听器函数
// 注意：由于 Go 函数比较的限制，RemoveListener 无法准确识别要移除的监听器
// 建议使用 RemoveAllListeners 清除所有监听器，或保存监听器引用后移除
func (e *Emitter[E, T]) RemoveListener(event E, listener Listener[T]) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 由于 Go 中函数无法直接比较，此方法仅用于保持 API 兼容性
	// 实际使用中建议通过其他方式管理监听器
	return e
}

// Off 是 RemoveListener 的别名，便于链式调用
func (e *Emitter[E, T]) Off(event E, listener Listener[T]) *Emitter[E, T] {
	return e.RemoveListener(event, listener)
}

// RemoveAllListeners 移除指定事件的所有监听器
// 参数 event: 事件标识
func (e *Emitter[E, T]) RemoveAllListeners(event E) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.events, event)
	return e
}

// Once 添加一个只触发一次的监听器
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 触发后自动移除
func (e *Emitter[E, T]) Once(event E, listener Listener[T]) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.maxListeners != -1 && len(e.events[event])+1 > e.maxListeners {
		fmt.Fprintf(os.Stdout, "Warning: event `%v` exceeds max listeners limit of %d\n", event, e.maxListeners)
	}

	id := e.nextID
	e.nextID++

	wrapper := &listenerWrapper[T]{
		id:       id,
		listener: listener,
		isOnce:   true,
	}
	e.events[event] = append(e.events[event], wrapper)
	return e
}

// Emit 异步触发事件的所有监听器
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
func (e *Emitter[E, T]) Emit(event E, args ...T) *Emitter[E, T] {
	e.mu.Lock()
	listeners, ok := e.events[event]
	if !ok {
		e.mu.Unlock()
		return e
	}
	// 复制监听器列表以避免在执行期间被修改
	listenersCopy := make([]*listenerWrapper[T], len(listeners))
	copy(listenersCopy, listeners)

	// 收集需要移除的 once 监听器的 ID
	var onceIDs []uint64
	for _, wrapper := range listenersCopy {
		if wrapper.isOnce {
			onceIDs = append(onceIDs, wrapper.id)
		}
	}
	e.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(listenersCopy))

	for _, wrapper := range listenersCopy {
		go func(w *listenerWrapper[T]) {
			defer wg.Done()
			e.callListener(event, w.listener, args...)
		}(wrapper)
	}

	wg.Wait()

	// 移除已触发的 once 监听器
	if len(onceIDs) > 0 {
		e.mu.Lock()
		if currentListeners, exists := e.events[event]; exists {
			onceIDSet := make(map[uint64]bool)
			for _, id := range onceIDs {
				onceIDSet[id] = true
			}
			newListeners := make([]*listenerWrapper[T], 0, len(currentListeners))
			for _, wrapper := range currentListeners {
				if !onceIDSet[wrapper.id] {
					newListeners = append(newListeners, wrapper)
				}
			}
			e.events[event] = newListeners
		}
		e.mu.Unlock()
	}
	return e
}

// EmitSync 同步触发事件的所有监听器
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
func (e *Emitter[E, T]) EmitSync(event E, args ...T) *Emitter[E, T] {
	e.mu.Lock()
	listeners, ok := e.events[event]
	if !ok {
		e.mu.Unlock()
		return e
	}
	// 复制监听器列表
	listenersCopy := make([]*listenerWrapper[T], len(listeners))
	copy(listenersCopy, listeners)

	// 收集需要移除的 once 监听器的 ID
	var onceIDs []uint64
	for _, wrapper := range listenersCopy {
		if wrapper.isOnce {
			onceIDs = append(onceIDs, wrapper.id)
		}
	}
	e.mu.Unlock()

	// 同步执行监听器
	for _, wrapper := range listenersCopy {
		e.callListener(event, wrapper.listener, args...)
	}

	// 移除已触发的 once 监听器
	if len(onceIDs) > 0 {
		e.mu.Lock()
		if currentListeners, exists := e.events[event]; exists {
			onceIDSet := make(map[uint64]bool)
			for _, id := range onceIDs {
				onceIDSet[id] = true
			}
			newListeners := make([]*listenerWrapper[T], 0, len(currentListeners))
			for _, wrapper := range currentListeners {
				if !onceIDSet[wrapper.id] {
					newListeners = append(newListeners, wrapper)
				}
			}
			e.events[event] = newListeners
		}
		e.mu.Unlock()
	}
	return e
}

// callListener 调用监听器并处理可能的 panic
func (e *Emitter[E, T]) callListener(event E, listener Listener[T], args ...T) {
	if e.recoverer != nil {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic occurred in listener: %v", r)
				e.recoverer(event, listener, err)
			}
		}()
	}
	listener(args...)
}

// RecoverWith 设置恢复监听器，用于处理 panic
// 参数 listener: 恢复监听器函数
func (e *Emitter[E, T]) RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.recoverer = listener
	return e
}

// SetMaxListeners 设置每个事件的最大监听器数量
// 参数 max: 最大数量，若为 -1 则无限制
func (e *Emitter[E, T]) SetMaxListeners(max int) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = max
	return e
}

// GetListenerCount 获取指定事件的监听器数量
// 参数 event: 事件标识
func (e *Emitter[E, T]) GetListenerCount(event E) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.events[event])
}
