package generator

import (
	"sync"
	"testing"
	"time"
)

// TestNewGenerator 测试创建生成器
func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := range 3 {
			yield.Yield(i)
		}
	})

	// 检查初始状态
	if gen.isDone {
		t.Error("Newly created generator should not be done")
	}
	if gen.doneChan == nil || gen.yield.valueChan == nil || gen.yield.resultChan == nil {
		t.Error("Generator channels should be initialized")
	}
}

// TestYield 测试 Yield 方法的值生成和返回值接收
func TestYield(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		result := yield.Yield(42)
		if result != "ack" {
			t.Errorf("Expected result 'ack', got %v", result)
		}
	})

	value, done := gen.Next("ack")
	if done {
		t.Error("First call to Next should not be done")
	}
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

// TestNext 测试 Next 方法的迭代行为
func TestNext(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := 0; i < 3; i++ {
			yield.Yield(i)
		}
	})

	expected := []int{0, 1, 2}
	for i := range expected {
		value, done := gen.Next()
		if done {
			t.Errorf("Next should not be done at iteration %d", i)
		}
		if value != expected[i] {
			t.Errorf("Expected value %d, got %d", expected[i], value)
		}
	}

	// 检查生成器完成
	value, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}
	if value != 0 { // 检查默认值
		t.Errorf("Expected zero value after done, got %d", value)
	}

	// 再次调用 Next，应保持 done 状态
	_, done = gen.Next()
	if !done {
		t.Error("Next should remain done after completion")
	}
}

// TestNextWithResult 测试带返回值的 Next 调用
func TestNextWithResult(t *testing.T) {
	gen := NewGenerator(func(yield Yield[string]) {
		for i := 0; i < 2; i++ {
			result := yield.Yield("value-" + string(rune('A'+i)))
			if result != "ack-"+string(rune('A'+i)) {
				t.Errorf("Expected result 'ack-%c', got %v", 'A'+i, result)
			}
		}
	})

	for i := 0; i < 2; i++ {
		value, done := gen.Next("ack-" + string(rune('A'+i)))
		if done {
			t.Errorf("Next should not be done at iteration %d", i)
		}
		if value != "value-"+string(rune('A'+i)) {
			t.Errorf("Expected value 'value-%c', got %s", 'A'+i, value)
		}
	}

	_, done := gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}
}

// TestGeneratorDone 测试生成器完成后的行为
func TestGeneratorDone(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		yield.Yield(1)
	})

	// 获取第一个值
	value, done := gen.Next()
	if done {
		t.Error("First call to Next should not be done")
	}
	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}

	// 检查完成状态
	value, done = gen.Next()
	if !done {
		t.Error("Next should return done=true after generator completes")
	}
	if !gen.isDone {
		t.Error("Generator should be marked as done")
	}

	// 验证通道关闭
	select {
	case _, ok := <-gen.yield.valueChan:
		if ok {
			t.Error("valueChan should be closed after generator completes")
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("valueChan should be closed immediately")
	}
}

// TestEmptyGenerator 测试空生成器
func TestEmptyGenerator(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		// 空生成器，不产生任何值
	})

	value, done := gen.Next()
	if !done {
		t.Error("Next should return done=true for empty generator")
	}
	if value != 0 {
		t.Errorf("Expected zero value for empty generator, got %d", value)
	}
	if !gen.isDone {
		t.Error("Empty generator should be marked as done")
	}
}

// TestConcurrentSafety 测试并发安全关闭
func TestConcurrentSafety(t *testing.T) {
	gen := NewGenerator(func(yield Yield[int]) {
		for i := range 5 {
			yield.Yield(i)
		}
	})

	var wg sync.WaitGroup
	wg.Add(3)

	// 并发调用 Next
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			for {
				_, done := gen.Next()
				if done {
					break
				}
				time.Sleep(1 * time.Millisecond) // 模拟并发延迟
			}
		}()
	}

	wg.Wait()
	if !gen.isDone {
		t.Error("Generator should be marked as done after concurrent access")
	}
}
