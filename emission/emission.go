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

// Emitter 是一个泛型事件发射器，用于管理事件的监听和触发
// T 必须是 comparable 类型，因为它用作 map 的键
type Emitter[T comparable] struct {
	mu           sync.Mutex                    // 互斥锁，确保线程安全
	events       map[T][]Listener[T]           // 事件到监听器列表的映射
	recoverer    RecoveryListener[T]           // 可选的恢复监听器，用于处理 panic
	maxListeners int                           // 每个事件的最大监听器数量，用于调试内存泄漏
	onces        map[*Listener[T]]*Listener[T] // 用于记录 Once 包装的监听器
}

// NewEmitter 创建一个新的泛型事件发射器
// 返回初始化好的 Emitter 实例，默认最大监听器数为 DefaultMaxListeners
func NewEmitter[T comparable]() *Emitter[T] {
	return &Emitter[T]{
		events:       make(map[T][]Listener[T]),
		maxListeners: DefaultMaxListeners,
		onces:        make(map[*Listener[T]]*Listener[T]),
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

	e.events[event] = append(e.events[event], listener)
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
		if once, exists := e.onces[lPtr]; exists {
			lPtr = once // 如果是 Once 包装的监听器，使用原始指针
		}

		newListeners := make([]Listener[T], 0, len(listeners))
		for _, l := range listeners {
			if &l != lPtr {
				newListeners = append(newListeners, l)
			}
		}
		e.events[event] = newListeners
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
	var once Listener[T]
	once = func(args ...T) {
		defer e.RemoveListener(event, once)
		listener(args...)
	}

	lPtr := &listener
	e.mu.Lock()
	e.onces[lPtr] = &once
	e.mu.Unlock()

	return e.AddListener(event, once)
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
	listenersCopy := make([]Listener[T], len(listeners))
	copy(listenersCopy, listeners)
	e.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(listenersCopy))

	for _, listener := range listenersCopy {
		go func(l Listener[T]) {
			defer wg.Done()
			e.callListener(event, l, args...)
		}(listener)
	}

	wg.Wait()
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
	listenersCopy := make([]Listener[T], len(listeners))
	copy(listenersCopy, listeners)
	e.mu.Unlock()

	for _, listener := range listenersCopy {
		e.callListener(event, listener, args...)
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
