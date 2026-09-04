package generator

import (
	"sync"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := range 3 {
			yield.Send(i)
		}
	})
	// 检查通道已初始化
	if gen.yield.valueCh == nil || gen.yield.resultCh == nil || gen.stopCh == nil || gen.closeCh == nil {
		t.Error("generator channels should be initialized")
	}
}

func TestNewGeneratorNilPanicsSynchronously(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewGenerator(nil) did not panic")
		}
	}()
	NewGenerator[int](nil)
}

func TestSend(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		result := yield.Send(42)
		if result != "ack" {
			t.Errorf("expected result 'ack', got %v", result)
		}
	})

	value, done := gen.Next("ack")
	if done {
		t.Error("first Next should not be done")
	}
	if value != 42 {
		t.Errorf("expected value 42, got %d", value)
	}
}

func TestNext(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := 0; i < 3; i++ {
			yield.Send(i)
		}
	})

	expected := []int{0, 1, 2}
	for i := range expected {
		value, done := gen.Next()
		if done {
			t.Errorf("Next should not be done at iteration %d", i)
		}
		if value != expected[i] {
			t.Errorf("expected value %d, got %d", expected[i], value)
		}
	}

	value, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}
	if value != 0 {
		t.Errorf("expected zero value after done, got %d", value)
	}

	_, done = gen.Next()
	if !done {
		t.Error("Next should remain done after completion")
	}
}

func TestNextWithResult(t *testing.T) {
	gen := NewGenerator(func(yield Yield[string]) {
		for i := 0; i < 2; i++ {
			result := yield.Send("value-" + string(rune('A'+i)))
			if result != "ack-"+string(rune('A'+i)) {
				t.Errorf("expected result 'ack-%c', got %v", 'A'+i, result)
			}
		}
	})

	for i := 0; i < 2; i++ {
		value, done := gen.Next("ack-" + string(rune('A'+i)))
		if done {
			t.Errorf("Next should not be done at iteration %d", i)
		}
		if value != "value-"+string(rune('A'+i)) {
			t.Errorf("expected value 'value-%c', got %s", 'A'+i, value)
		}
	}

	_, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}
}

func TestGeneratorDone(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		yield.Send(1)
	})

	value, done := gen.Next()
	if done {
		t.Error("first Next should not be done")
	}
	if value != 1 {
		t.Errorf("expected value 1, got %d", value)
	}

	_, done = gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}

	// 验证通道关闭
	select {
	case _, ok := <-gen.yield.valueCh:
		if ok {
			t.Error("valueCh should be closed after generator completes")
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("valueCh should be closed immediately")
	}
}

func TestEmptyGenerator(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		// 不产生任何值
	})

	value, done := gen.Next()
	if !done {
		t.Error("Next should return done=true for empty generator")
	}
	if value != 0 {
		t.Errorf("expected zero value for empty generator, got %d", value)
	}
}

func TestStop(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := 0; ; i++ {
			if yield.Stopped() {
				return
			}
			yield.Send(i)
		}
	})

	// 消费几个值
	for i := 0; i < 3; i++ {
		value, done := gen.Next()
		if done {
			t.Fatalf("Next unexpectedly done at iteration %d", i)
		}
		if value != i {
			t.Errorf("expected %d, got %d", i, value)
		}
	}

	gen.Stop()

	// 生成器应很快结束
	_, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after Stop")
	}
}

func TestStopDuringSend(t *testing.T) {
	ready := make(chan struct{})
	gen := NewGenerator(func(yield Yield[int]) {
		close(ready)
		result := yield.Send(42)
		if result != nil {
			t.Errorf("Send should return nil when stopped, got %v", result)
		}
	})

	<-ready
	gen.Stop()

	_, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after Stop")
	}
}

func TestStopMultiple(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		<-time.After(50 * time.Millisecond)
	})
	gen.Stop()
	gen.Stop() // 不应 panic
	gen.Stop()
}

func TestStopConcurrent(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		<-yield.stopCh
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gen.Stop()
		}()
	}
	wg.Wait()

	select {
	case <-gen.closeCh:
	case <-time.After(time.Second):
		t.Fatal("generator did not exit after concurrent Stop calls")
	}
}

func TestConcurrentStopUnblocksConsumerAndGenerator(t *testing.T) {
	started := make(chan struct{})
	gen := NewGenerator(func(yield Yield[int]) {
		close(started)
		for !yield.Stopped() {
			yield.Send(1)
		}
	})
	<-started

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		for {
			if _, done := gen.Next(); done {
				return
			}
		}
	}()

	var stops sync.WaitGroup
	for range 100 {
		stops.Add(1)
		go func() {
			defer stops.Done()
			gen.Stop()
		}()
	}
	stops.Wait()

	select {
	case <-consumerDone:
	case <-time.After(time.Second):
		t.Fatal("consumer remained blocked after concurrent Stop calls")
	}
	select {
	case <-gen.closeCh:
	case <-time.After(time.Second):
		t.Fatal("generator remained blocked after concurrent Stop calls")
	}
}

func TestStoppedBeforeSend(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		if yield.Stopped() {
			return // 已停止，不发送值
		}
		yield.Send(1)
	})

	gen.Stop()
	_, done := gen.Next()
	if !done {
		t.Error("Next should return done=true")
	}
}
