// Package emission 提供了一个类型安全的泛型事件发射器实现，
// 支持同步/异步事件触发、一次性监听器、并发度控制和 panic 恢复。
package emission

import (
	"sync"
)

// DefaultMaxListeners 默认的最大监听器数量
const DefaultMaxListeners = 10

// Logger 定义日志接口，用于记录警告和错误信息
type Logger interface {
	Warnf(format string, args ...any)
}

// RecoveryListener 定义恢复监听器的签名，用于处理 panic
// E: 事件类型，T: 监听器参数类型
// panicValue: 原始的 panic 值，可以是任意类型
type RecoveryListener[E comparable, T any] func(event E, listener any, panicValue any)

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
	logger       Logger                      // 可选的日志记录器
	semaphore    chan struct{}               // 并发度限制信号量，nil 表示无限制
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

// addListener 内部方法，添加监听器到指定事件
// 参数 once: 是否为一次性监听器
// 返回一个取消函数，调用该函数可移除此监听器
func (e *Emitter[E, T]) addListener(event E, listener Listener[T], once bool) func() {
	e.mu.Lock()

	if e.maxListeners != -1 && len(e.events[event])+1 > e.maxListeners {
		if e.logger != nil {
			e.logger.Warnf("event `%v` exceeds max listeners limit of %d", event, e.maxListeners)
		}
	}

	id := e.nextID
	e.nextID++
	wrapper := &listenerWrapper[T]{
		id:       id,
		listener: listener,
		isOnce:   once,
	}
	e.events[event] = append(e.events[event], wrapper)
	e.mu.Unlock()

	// 返回取消函数
	return func() {
		e.removeListenerByID(event, id)
	}
}

// AddListener 添加监听器到指定事件
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 返回一个取消函数，调用该函数可移除此监听器
// 如果监听器数量超过 maxListeners，会通过 logger 记录警告
func (e *Emitter[E, T]) AddListener(event E, listener Listener[T]) func() {
	return e.addListener(event, listener, false)
}

// On 是 AddListener 的别名
// 返回一个取消函数，调用该函数可移除此监听器
func (e *Emitter[E, T]) On(event E, listener Listener[T]) func() {
	return e.addListener(event, listener, false)
}

// Once 添加一个只触发一次的监听器
// 参数 event: 事件标识
// 参数 listener: 监听器函数
// 返回一个取消函数，调用该函数可在触发前移除此监听器
// 触发后自动移除
func (e *Emitter[E, T]) Once(event E, listener Listener[T]) func() {
	return e.addListener(event, listener, true)
}

// removeListenerByID 通过 ID 移除监听器（内部方法）
// 使用 swap-remove 技术优化性能，时间复杂度 O(1)
// 注意：swap-remove 不保留监听器的注册顺序
func (e *Emitter[E, T]) removeListenerByID(event E, id uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, ok := e.events[event]
	if !ok {
		return
	}

	// 查找并使用 swap-remove 技术删除（交换最后一个元素，然后截断）
	for i, wrapper := range listeners {
		if wrapper.id == id {
			lastIdx := len(listeners) - 1
			listeners[i] = listeners[lastIdx]
			listeners[lastIdx] = nil // 避免内存泄漏
			e.events[event] = listeners[:lastIdx]

			if lastIdx == 0 {
				delete(e.events, event)
			}
			return
		}
	}
}

// RemoveAllListeners 移除指定事件的所有监听器
// 参数 event: 事件标识
func (e *Emitter[E, T]) RemoveAllListeners(event E) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.events, event)
	return e
}

// prepareEmit 原子地复制监听器列表并移除 once 监听器
// 在持有锁的情况下完成快照操作，避免 once 监听器在并发 Emit 中被重复触发
// 返回要执行的监听器副本、信号量快照和恢复监听器快照，若无监听器则返回 nil
func (e *Emitter[E, T]) prepareEmit(event E) ([]*listenerWrapper[T], chan struct{}, RecoveryListener[E, T]) {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, ok := e.events[event]
	if !ok || len(listeners) == 0 {
		return nil, nil, nil
	}

	// 快照信号量和恢复监听器引用，保证后续使用无数据竞争
	sem := e.semaphore
	recoverer := e.recoverer

	// 复制监听器列表供调用方使用
	result := make([]*listenerWrapper[T], len(listeners))
	copy(result, listeners)

	// 检查是否存在 once 监听器
	hasOnce := false
	for _, w := range listeners {
		if w.isOnce {
			hasOnce = true
			break
		}
	}
	if !hasOnce {
		return result, sem, recoverer
	}

	// 就地过滤，从源列表中移除 once 监听器
	n := 0
	for _, w := range listeners {
		if !w.isOnce {
			listeners[n] = w
			n++
		}
	}
	// 清除剩余引用，避免内存泄漏
	for i := n; i < len(listeners); i++ {
		listeners[i] = nil
	}
	if n == 0 {
		delete(e.events, event)
	} else {
		e.events[event] = listeners[:n]
	}

	return result, sem, recoverer
}

// Emit 异步触发事件的所有监听器（Fire-and-forget）
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
// 注意：此方法立即返回，不等待监听器执行完成
// 若通过 SetConcurrency 设置了并发度，则使用 worker pool 限制 goroutine 创建数量
func (e *Emitter[E, T]) Emit(event E, args ...T) {
	listeners, sem, recoverer := e.prepareEmit(event)
	if len(listeners) == 0 {
		return
	}

	go func() {
		if sem == nil {
			// 无并发限制：直接为每个监听器创建 goroutine
			var wg sync.WaitGroup
			wg.Add(len(listeners))
			for _, wrapper := range listeners {
				go func() {
					defer wg.Done()
					e.callListener(event, wrapper.listener, recoverer, args...)
				}()
			}
			wg.Wait()
		} else {
			// 有并发限制：使用 worker pool 模式
			e.runWithWorkerPool(event, listeners, sem, recoverer, args...)
		}
	}()
}

// EmitWait 并发触发事件的所有监听器，并等待所有监听器执行完成
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
// 注意：此方法会阻塞直到所有监听器执行完成
// 若通过 SetConcurrency 设置了并发度，则使用 worker pool 限制 goroutine 创建数量
func (e *Emitter[E, T]) EmitWait(event E, args ...T) {
	listeners, sem, recoverer := e.prepareEmit(event)
	if len(listeners) == 0 {
		return
	}

	if sem == nil {
		// 无并发限制：直接为每个监听器创建 goroutine
		var wg sync.WaitGroup
		wg.Add(len(listeners))
		for _, wrapper := range listeners {
			go func() {
				defer wg.Done()
				e.callListener(event, wrapper.listener, recoverer, args...)
			}()
		}
		wg.Wait()
	} else {
		// 有并发限制：使用 worker pool 模式
		e.runWithWorkerPool(event, listeners, sem, recoverer, args...)
	}
}

// EmitSync 同步触发事件的所有监听器
// 参数 event: 事件标识
// 参数 args: 传递给监听器的参数
// 注意：此方法按顺序同步执行所有监听器，不受 SetConcurrency 影响
func (e *Emitter[E, T]) EmitSync(event E, args ...T) {
	listeners, _, recoverer := e.prepareEmit(event)
	if len(listeners) == 0 {
		return
	}

	for _, wrapper := range listeners {
		e.callListener(event, wrapper.listener, recoverer, args...)
	}
}

// runWithWorkerPool 使用 worker pool 模式执行监听器
// 只创建 min(cap(sem), len(listeners)) 个 worker goroutine，避免创建多余的空闲 goroutine
func (e *Emitter[E, T]) runWithWorkerPool(event E, listeners []*listenerWrapper[T], sem chan struct{}, recoverer RecoveryListener[E, T], args ...T) {
	workerCount := min(cap(sem), len(listeners))
	tasks := make(chan *listenerWrapper[T], len(listeners))

	// 填充任务队列
	for _, wrapper := range listeners {
		tasks <- wrapper
	}
	close(tasks)

	// 启动固定数量的 worker goroutine
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()
			for wrapper := range tasks {
				e.callListener(event, wrapper.listener, recoverer, args...)
			}
		}()
	}
	wg.Wait()
}

// callListener 调用监听器并处理可能的 panic
// recoverer 必须是在持有锁期间快照的值，避免数据竞争
func (e *Emitter[E, T]) callListener(event E, listener Listener[T], recoverer RecoveryListener[E, T], args ...T) {
	if recoverer != nil {
		defer func() {
			if r := recover(); r != nil {
				recoverer(event, listener, r)
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

// SetLogger 设置日志记录器
// 参数 logger: 实现 Logger 接口的日志记录器
func (e *Emitter[E, T]) SetLogger(logger Logger) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
	return e
}

// SetConcurrency 设置并发执行监听器的最大数量
// 参数 n: 最大并发数，n <= 0 表示无限制（默认）
// 影响 Emit 和 EmitWait，不影响 EmitSync
// 信号量在所有事件间共享，用于全局 goroutine 背压控制
func (e *Emitter[E, T]) SetConcurrency(n int) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	if n <= 0 {
		e.semaphore = nil
	} else {
		e.semaphore = make(chan struct{}, n)
	}
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
