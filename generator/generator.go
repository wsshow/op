package generator

type Yield[T any] struct {
	value  chan T
	result chan any
}

func (y *Yield[T]) Yield(value T) (result any) {
	y.value <- value
	select {
	case result = <-y.result:
	default:
	}
	return
}

type Generator[T any] struct {
	done chan bool
	y    Yield[T]
}

func NewGenerator[T any](genFunc func(yield Yield[T])) *Generator[T] {
	g := &Generator[T]{
		y:    Yield[T]{value: make(chan T), result: make(chan any)},
		done: make(chan bool),
	}
	go g.run(genFunc)
	return g
}

func (g *Generator[T]) run(f func(yield Yield[T])) {
	f(g.y)
	close(g.y.value)
	close(g.done)
}

func (g *Generator[T]) Next(values ...any) (value T, done bool) {
	if len(values) > 0 {
		g.y.result <- values[0]
	}
	select {
	case value, ok := <-g.y.value:
		return value, !ok
	case <-g.done:
		return value, true
	}
}
