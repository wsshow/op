package op

import (
	"testing"
)

func TestNewString(t *testing.T) {
	s := NewString("hello")
	if s.String() != "hello" {
		t.Errorf("NewString: expected 'hello', got '%s'", s.String())
	}
}

func TestNewSlice(t *testing.T) {
	s := NewSlice(1, 2, 3)
	if s.Length() != 3 {
		t.Errorf("NewSlice: expected length 3, got %d", s.Length())
	}
}

func TestNewDeque(t *testing.T) {
	d := NewDeque[int]()
	if d.Size() != 0 {
		t.Errorf("NewDeque: expected size 0, got %d", d.Size())
	}
}

func TestNewDequeWithCapacity(t *testing.T) {
	d := NewDeque[int](128)
	d.PushBack(1)
	if d.Capacity() < 128 {
		t.Errorf("NewDeque with capacity: expected capacity >= 128 after first push, got %d", d.Capacity())
	}
}

func TestNewEmitter(t *testing.T) {
	e := NewEmitter[string, int]()
	if e.TotalListenerCount() != 0 {
		t.Errorf("NewEmitter: expected 0 listeners, got %d", e.TotalListenerCount())
	}
}

func TestLinqFrom(t *testing.T) {
	l := LinqFrom([]int{1, 2, 3})
	if l.Count() != 3 {
		t.Errorf("LinqFrom: expected count 3, got %d", l.Count())
	}
}

func TestLinqFromNonComparable(t *testing.T) {
	type nonComp struct{ vals []int }
	s := []nonComp{{vals: []int{1}}, {vals: []int{2}}}
	l := LinqFrom(s)
	if l.Count() != 2 {
		t.Errorf("LinqFrom with non-comparable type: expected count 2, got %d", l.Count())
	}
}

func TestLinqEmpty(t *testing.T) {
	l := LinqEmpty[int]()
	if l.Count() != 0 {
		t.Errorf("LinqEmpty: expected count 0, got %d", l.Count())
	}
}

func TestLinqRange(t *testing.T) {
	l := LinqRange(5, 3)
	r := l.Results()
	if len(r) != 3 || r[0] != 5 || r[1] != 6 || r[2] != 7 {
		t.Errorf("LinqRange(5,3): expected [5 6 7], got %v", r)
	}
	// count <= 0 returns empty
	e := LinqRange(0, -1)
	if e.Count() != 0 {
		t.Errorf("LinqRange(0,-1): expected empty, got count %d", e.Count())
	}
}

func TestLinqRepeat(t *testing.T) {
	l := LinqRepeat("x", 3)
	r := l.Results()
	if len(r) != 3 || r[0] != "x" || r[1] != "x" || r[2] != "x" {
		t.Errorf("LinqRepeat('x',3): expected [x x x], got %v", r)
	}
}

func TestNewProcess(t *testing.T) {
	p := NewProcess(Options{})
	if p.IsRunning() {
		t.Error("NewProcess: process should not be running")
	}
}

func TestNewProcessManager(t *testing.T) {
	m := NewProcessManager()
	if m.Count() != 0 {
		t.Errorf("NewProcessManager: expected count 0, got %d", m.Count())
	}
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(func(yield Yield[int]) {
		for i := 0; i < 3; i++ {
			yield.Send(i)
		}
	})
	for i := 0; i < 3; i++ {
		v, done := g.Next()
		if done {
			t.Errorf("NewGenerator: unexpected done at iteration %d", i)
		}
		if v != i {
			t.Errorf("NewGenerator: expected %d, got %d", i, v)
		}
	}
	_, done := g.Next()
	if !done {
		t.Error("NewGenerator: expected done after all values")
	}
}

func TestNewWorkerPool(t *testing.T) {
	p := NewWorkerPool(4)
	if p.Size() != 4 {
		t.Errorf("NewWorkerPool: expected size 4, got %d", p.Size())
	}
	p.Submit(func() {})
	p.Stop()
}

// Verify type aliases are usable in variable declarations.
func TestTypeAliases(t *testing.T) {
	var _ String
	var _ Slice[int]
	var _ Deque[int]
	var _ Emitter[string, int]
	var _ Linq[int]
	var _ *Process
	var _ *Manager
	var _ *Generator[int]
	var _ *WorkerPool
}
