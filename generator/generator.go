// Package generator 提供一个泛型生成器，支持通过 goroutine 与 channel
// 迭代生成值，并允许调用者在每次迭代时向生成器回传结果。
//
// 每个 Generator 仅支持单个消费者 goroutine。
package generator

import "sync"

// Yield 供生成器函数用于向消费者发送值并接收回传结果。
type Yield[T any] struct {
	valueCh  chan T          // 向消费者发送值
	resultCh chan any        // 从消费者接收结果
	stopCh   <-chan struct{} // 消费者已请求停止
}

// Send 向消费者发送值并阻塞等待其结果。
// 若消费者已停止生成器，则立即返回 nil。
func (y *Yield[T]) Send(value T) any {
	select {
	case y.valueCh <- value:
	case <-y.stopCh:
		return nil
	}
	select {
	case result := <-y.resultCh:
		return result
	case <-y.stopCh:
		return nil
	}
}

// Stopped 报告消费者是否已请求停止生成器。
func (y *Yield[T]) Stopped() bool {
	select {
	case <-y.stopCh:
		return true
	default:
		return false
	}
}

// Generator 从生成器函数中产生值。
type Generator[T any] struct {
	yield     Yield[T]
	stopCh    chan struct{}
	closeCh   chan struct{}
	stopOnce  sync.Once
	closeOnce sync.Once
}

// NewGenerator 创建并启动一个新的生成器，在后台 goroutine 中运行 genFunc。
func NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T] {
	g := &Generator[T]{
		yield: Yield[T]{
			valueCh:  make(chan T),
			resultCh: make(chan any),
		},
		stopCh:  make(chan struct{}),
		closeCh: make(chan struct{}),
	}
	g.yield.stopCh = g.stopCh
	go g.run(genFunc)
	return g
}

func (g *Generator[T]) run(genFunc func(yield Yield[T])) {
	defer g.close()
	genFunc(g.yield)
}

func (g *Generator[T]) close() {
	g.closeOnce.Do(func() {
		close(g.closeCh)
		close(g.yield.valueCh)
		// resultCh 只有生成器接收、消费者发送。它无需关闭；保持打开可避免
		// Stop 与 Next 并发时，Next 的 select 包含向已关闭 channel 发送。
	})
}

// Next 返回生成的下一个值。done 为 true 时表示生成器已结束或被停止，
// 此时返回值是类型 T 的零值。
func (g *Generator[T]) Next(values ...any) (value T, done bool) {
	select {
	case val, ok := <-g.yield.valueCh:
		if !ok {
			return value, true
		}
		var result any
		if len(values) > 0 {
			result = values[0]
		}
		select {
		case g.yield.resultCh <- result:
		case <-g.closeCh:
		case <-g.stopCh:
		}
		return val, false
	case <-g.closeCh:
		return value, true
	case <-g.stopCh:
		return value, true
	}
}

// Stop 通知生成器停止生成值。可安全地多次调用。
// 调用 Stop 后，不应再调用 Next。
func (g *Generator[T]) Stop() {
	g.stopOnce.Do(func() { close(g.stopCh) })
}
