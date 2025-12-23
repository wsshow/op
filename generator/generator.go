package generator

import "sync"

// Yield 用于在生成器中产生值并接收返回值
type Yield[T any] struct {
	valueChan  chan T   // 用于发送生成的值
	resultChan chan any // 用于接收调用者传递的返回值
}

// Yield 将值发送给调用者，并等待接收返回值
// 如果没有返回值，则返回 nil
func (y *Yield[T]) Yield(value T) any {
	y.valueChan <- value
	// 阻塞等待返回值或通道关闭
	result, ok := <-y.resultChan
	if !ok {
		return nil // 通道已关闭
	}
	return result
}

// Generator 是一个泛型生成器，支持迭代生成值
type Generator[T any] struct {
	yield     Yield[T]  // 用于值传递的 Yield 实例
	doneChan  chan bool // 标记生成器是否完成
	isDone    bool      // 内部状态，标记是否已完成
	closeOnce sync.Once // 确保通道只关闭一次
}

// NewGenerator 创建并启动一个新的生成器
// genFunc 是生成逻辑，接收 Yield[T] 用于产生值
func NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T] {
	g := &Generator[T]{
		yield:    Yield[T]{valueChan: make(chan T), resultChan: make(chan any)},
		doneChan: make(chan bool),
	}
	go g.run(genFunc) // 在 goroutine 中运行生成逻辑
	return g
}

// run 执行生成器的核心逻辑
// 在生成完成后关闭通道
func (g *Generator[T]) run(genFunc func(yield Yield[T])) {
	defer g.close() // 确保在函数退出时关闭通道
	genFunc(g.yield)
}

// close 安全地关闭生成器的通道
func (g *Generator[T]) close() {
	g.closeOnce.Do(func() {
		close(g.yield.valueChan)
		close(g.yield.resultChan)
		close(g.doneChan)
		g.isDone = true
	})
}

// Next 获取生成器的下一个值
// values 可选参数，用于向生成器传递返回值
// 返回值：生成的 value 和 done 状态（true 表示生成结束）
func (g *Generator[T]) Next(values ...any) (value T, done bool) {
	if g.isDone {
		return value, true // 如果已完成，直接返回
	}

	// 先等待生成的下一个值或完成信号
	select {
	case val, ok := <-g.yield.valueChan:
		if !ok {
			g.isDone = true
			return value, true // 通道关闭，表示生成结束
		}
		// 发送返回值（如果提供）或 nil
		var result any
		if len(values) > 0 {
			result = values[0]
		}
		select {
		case g.yield.resultChan <- result:
		case <-g.doneChan:
			g.isDone = true
			return value, true
		}
		return val, false
	case <-g.doneChan:
		g.isDone = true
		return value, true // 生成器完成
	}
}
