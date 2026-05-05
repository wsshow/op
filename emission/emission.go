// Package emission 提供了一个类型安全的泛型事件发射器，
// 支持同步/异步事件触发、一次性监听器、并发度控制和 panic 恢复。
package emission

import (
	"sync"
)

// DefaultMaxListeners 默认的最大监听器数量警告阈值
const DefaultMaxListeners = 10

// Logger 定义日志接口
type Logger interface {
	Warnf(format string, args ...any)
}

// Listener 定义监听器函数签名，接收单个泛型参数
type Listener[T any] func(T)

// RecoveryListener 定义 panic 恢复监听器签名
// E: 事件类型，T: 监听器参数类型
type RecoveryListener[E comparable, T any] func(event E, listener any, panicValue any)

// listenerWrapper 包装监听器并添加元数据
type listenerWrapper[T any] struct {
	id       uint64
	listener Listener[T]
	isOnce   bool
}

// Subscription 表示一个已注册的监听器订阅，可用于取消监听
type Subscription[E comparable, T any] struct {
	emitter *Emitter[E, T]
	event   E
	id      uint64
}

// Unsubscribe 取消此订阅，从事件中移除对应的监听器。
// 多次调用安全，后续调用为 no-op。
func (s *Subscription[E, T]) Unsubscribe() {
	if s.emitter != nil {
		s.emitter.removeListenerByID(s.event, s.id)
		s.emitter = nil
	}
}

// Emitter 是一个泛型事件发射器，用于管理事件的监听和触发
// E: 事件标识类型（必须是 comparable），T: 监听器参数类型
type Emitter[E comparable, T any] struct {
	mu           sync.Mutex
	events       map[E][]*listenerWrapper[T]
	recoverer    RecoveryListener[E, T]
	maxListeners int
	nextID       uint64
	logger       Logger
	semaphore    chan struct{} // 并发度信号量，nil 表示无限制
}

// NewEmitter 创建一个新的事件发射器
func NewEmitter[E comparable, T any]() *Emitter[E, T] {
	return &Emitter[E, T]{
		events:       make(map[E][]*listenerWrapper[T]),
		maxListeners: DefaultMaxListeners,
		nextID:       1,
	}
}

// addListener 内部方法，添加监听器到指定事件，返回 Subscription
func (e *Emitter[E, T]) addListener(event E, listener Listener[T], once bool) *Subscription[E, T] {
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

	return &Subscription[E, T]{
		emitter: e,
		event:   event,
		id:      id,
	}
}

// On 添加监听器到指定事件，返回 Subscription 用于取消
func (e *Emitter[E, T]) On(event E, listener Listener[T]) *Subscription[E, T] {
	return e.addListener(event, listener, false)
}

// AddListener 是 On 的别名
func (e *Emitter[E, T]) AddListener(event E, listener Listener[T]) *Subscription[E, T] {
	return e.addListener(event, listener, false)
}

// Once 添加一个只触发一次的监听器，返回 Subscription 可用于触发前取消
func (e *Emitter[E, T]) Once(event E, listener Listener[T]) *Subscription[E, T] {
	return e.addListener(event, listener, true)
}

// removeListenerByID 通过 ID 移除监听器，使用 swap-remove 优化为 O(1)
func (e *Emitter[E, T]) removeListenerByID(event E, id uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, ok := e.events[event]
	if !ok {
		return
	}

	for i, wrapper := range listeners {
		if wrapper.id == id {
			lastIdx := len(listeners) - 1
			listeners[i] = listeners[lastIdx]
			listeners[lastIdx] = nil
			e.events[event] = listeners[:lastIdx]

			if lastIdx == 0 {
				delete(e.events, event)
			}
			return
		}
	}
}

// RemoveAllListeners 移除指定事件的所有监听器
func (e *Emitter[E, T]) RemoveAllListeners(event E) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.events, event)
	return e
}

// preparedEmit 持有 prepareEmit 的快照结果
type preparedEmit[E comparable, T any] struct {
	listeners []*listenerWrapper[T]
	sem       chan struct{}
	recoverer RecoveryListener[E, T]
	logger    Logger
}

// prepareEmit 原子地复制监听器列表并移除 once 监听器
func (e *Emitter[E, T]) prepareEmit(event E) preparedEmit[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, ok := e.events[event]
	if !ok || len(listeners) == 0 {
		return preparedEmit[E, T]{}
	}

	snap := preparedEmit[E, T]{
		sem:       e.semaphore,
		recoverer: e.recoverer,
		logger:    e.logger,
	}

	snap.listeners = make([]*listenerWrapper[T], len(listeners))
	copy(snap.listeners, listeners)

	hasOnce := false
	for _, w := range listeners {
		if w.isOnce {
			hasOnce = true
			break
		}
	}
	if !hasOnce {
		return snap
	}

	n := 0
	for _, w := range listeners {
		if !w.isOnce {
			listeners[n] = w
			n++
		}
	}
	for i := n; i < len(listeners); i++ {
		listeners[i] = nil
	}
	if n == 0 {
		delete(e.events, event)
	} else {
		e.events[event] = listeners[:n]
	}

	return snap
}

// Emit 异步触发事件（fire-and-forget），立即返回
// 若设置了并发度，信号量在多个 Emit 调用间共享以提供全局背压控制
func (e *Emitter[E, T]) Emit(event E, value T) *Emitter[E, T] {
	snap := e.prepareEmit(event)
	if len(snap.listeners) == 0 {
		return e
	}

	go func() {
		for _, w := range snap.listeners {
			if snap.sem != nil {
				snap.sem <- struct{}{} // 获取信号量令牌（全局背压点）
			}
			w := w
			go func() {
				if snap.sem != nil {
					defer func() { <-snap.sem }() // 释放信号量令牌
				}
				e.callListener(event, w.listener, snap.recoverer, snap.logger, value)
			}()
		}
	}()

	return e
}

// EmitWait 并发触发事件的所有监听器，等待全部完成后返回
// 若设置了并发度，信号量在多个调用间共享以提供全局背压控制
func (e *Emitter[E, T]) EmitWait(event E, value T) *Emitter[E, T] {
	snap := e.prepareEmit(event)
	if len(snap.listeners) == 0 {
		return e
	}

	var wg sync.WaitGroup
	for _, w := range snap.listeners {
		if snap.sem != nil {
			snap.sem <- struct{}{} // 获取信号量令牌
		}
		wg.Add(1)
		w := w
		go func() {
			defer wg.Done()
			if snap.sem != nil {
				defer func() { <-snap.sem }() // 释放信号量令牌
			}
			e.callListener(event, w.listener, snap.recoverer, snap.logger, value)
		}()
	}
	wg.Wait()

	return e
}

// EmitSync 同步顺序触发事件的所有监听器，不受并发度设置影响
func (e *Emitter[E, T]) EmitSync(event E, value T) *Emitter[E, T] {
	snap := e.prepareEmit(event)
	for _, w := range snap.listeners {
		e.callListener(event, w.listener, snap.recoverer, snap.logger, value)
	}
	return e
}

// callListener 调用监听器，始终 recover panic
func (e *Emitter[E, T]) callListener(event E, listener Listener[T], recoverer RecoveryListener[E, T], logger Logger, value T) {
	defer func() {
		if r := recover(); r != nil {
			if recoverer != nil {
				recoverer(event, listener, r)
			} else if logger != nil {
				logger.Warnf("panic in listener for event `%v`: %v", event, r)
			}
		}
	}()
	listener(value)
}

// RecoverWith 设置自定义 panic 恢复监听器
func (e *Emitter[E, T]) RecoverWith(listener RecoveryListener[E, T]) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.recoverer = listener
	return e
}

// SetLogger 设置日志记录器
func (e *Emitter[E, T]) SetLogger(logger Logger) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
	return e
}

// SetConcurrency 设置全局并发执行监听器的最大数量
// n <= 0 表示无限制。影响 Emit 和 EmitWait，不影响 EmitSync
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

// SetMaxListeners 设置每个事件的最大监听器数量警告阈值，-1 表示无限制
func (e *Emitter[E, T]) SetMaxListeners(max int) *Emitter[E, T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = max
	return e
}

// GetListenerCount 获取指定事件的监听器数量
func (e *Emitter[E, T]) GetListenerCount(event E) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.events[event])
}

// Events 返回所有已注册的事件列表
func (e *Emitter[E, T]) Events() []E {
	e.mu.Lock()
	defer e.mu.Unlock()
	events := make([]E, 0, len(e.events))
	for event := range e.events {
		events = append(events, event)
	}
	return events
}

// TotalListenerCount 返回所有事件的监听器总数
func (e *Emitter[E, T]) TotalListenerCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	total := 0
	for _, listeners := range e.events {
		total += len(listeners)
	}
	return total
}
