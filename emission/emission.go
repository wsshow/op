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
type RecoveryListener[T comparable] func(event T, listener interface{}, err error)

// Listener 定义监听器函数的签名，接受泛型参数
type Listener[T comparable] func(args ...T)

// listenerWrapper 包装监听器并添加唯一标识
type listenerWrapper[T comparable] struct {
	id       uint64      // 唯一标识符
	listener Listener[T] // 实际的监听器函数
	isOnce   bool        // 是否为 Once 监听器
}

// Emitter 是一个泛型事件发射器，用于管理事件的监听和触发
// T 必须是 comparable 类型，因为它用作 map 的键
type Emitter[T comparable] struct {
	mu           sync.Mutex                  // 互斥锁，确保线程安全
	events       map[T][]*listenerWrapper[T] // 事件到监听器列表的映射
	recoverer    RecoveryListener[T]         // 可选的恢复监听器，用于处理 panic
	maxListeners int                         // 每个事件的最大监听器数量，用于调试内存泄漏
	nextID       uint64                      // 下一个监听器的ID
	listenerIDs  map[*Listener[T]]uint64     // 监听器函数到ID的映射
}

// NewEmitter 创建一个新的泛型事件发射器
// 返回初始化好的 Emitter 实例，默认最大监听器数为 DefaultMaxListeners
func NewEmitter[T comparable]() *Emitter[T] {
	return &Emitter[T]{
		events:       make(map[T][]*listenerWrapper[T]),
		maxListeners: DefaultMaxListeners,
		nextID:       1,
		listenerIDs:  make(map[*Listener[T]]uint64),
	}
}

// AddListener 添加监听器到指定事件
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 如果监听器数量超过 maxListeners，会打印警告
func (e *Emitter[T]) AddListener(event T, listener Listener[T]) *Emitter[T] {
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
	e.listenerIDs[&listener] = id
	return e
}

// On 是 AddListener 的别名，便于链式调用
func (e *Emitter[T]) On(event T, listener Listener[T]) *Emitter[T] {
	return e.AddListener(event, listener)
}

// RemoveListener 从指定事件中移除监听器
// 参数 event: 事件标识
// 参数 listener: 要移除的监听器函数
func (e *Emitter[T]) RemoveListener(event T, listener Listener[T]) *Emitter[T] {
	e.mu.Lock()
	defer e.mu.Unlock()

	if listeners, ok := e.events[event]; ok {
		lPtr := &listener
		id, exists := e.listenerIDs[lPtr]
		if !exists {
			return e // 监听器未注册，直接返回
		}

		newListeners := make([]*listenerWrapper[T], 0, len(listeners))
		for _, wrapper := range listeners {
			if wrapper.id != id {
				newListeners = append(newListeners, wrapper)
			}
		}
		e.events[event] = newListeners
		delete(e.listenerIDs, lPtr)
	}
	return e
}

// Off 是 RemoveListener 的别名，便于链式调用
func (e *Emitter[T]) Off(event T, listener Listener[T]) *Emitter[T] {
	return e.RemoveListener(event, listener)
}

// Once 添加一个只触发一次的监听器
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 触发后自动移除
func (e *Emitter[T]) Once(event T, listener Listener[T]) *Emitter[T] {
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
	e.listenerIDs[&listener] = id
	return e
}

// Emit 异步触发事件的所有监听器
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
func (e *Emitter[T]) Emit(event T, args ...T) *Emitter[T] {
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
				} else {
					// 从 listenerIDs 中移除
					for lPtr, id := range e.listenerIDs {
						if id == wrapper.id {
							delete(e.listenerIDs, lPtr)
							break
						}
					}
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
func (e *Emitter[T]) EmitSync(event T, args ...T) *Emitter[T] {
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
				} else {
					// 从 listenerIDs 中移除
					for lPtr, id := range e.listenerIDs {
						if id == wrapper.id {
							delete(e.listenerIDs, lPtr)
							break
						}
					}
				}
			}
			e.events[event] = newListeners
		}
		e.mu.Unlock()
	}
	return e
}

// callListener 调用监听器并处理可能的 panic
func (e *Emitter[T]) callListener(event T, listener Listener[T], args ...T) {
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
func (e *Emitter[T]) RecoverWith(listener RecoveryListener[T]) *Emitter[T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.recoverer = listener
	return e
}

// SetMaxListeners 设置每个事件的最大监听器数量
// 参数 max: 最大数量，若为 -1 则无限制
func (e *Emitter[T]) SetMaxListeners(max int) *Emitter[T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = max
	return e
}

// GetListenerCount 获取指定事件的监听器数量
// 参数 event: 事件标识
func (e *Emitter[T]) GetListenerCount(event T) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.events[event])
}
