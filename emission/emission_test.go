package emission

import (
	"sync"
	"testing"
	"time"
)

type User struct {
	Name string
	Age  int
}

func TestNewEmitter(t *testing.T) {
	em := NewEmitter[string, string]()
	if em.maxListeners != DefaultMaxListeners {
		t.Errorf("expected maxListeners=%d, got %d", DefaultMaxListeners, em.maxListeners)
	}
	if len(em.events) != 0 {
		t.Errorf("expected empty events map, got %d events", len(em.events))
	}
	if em.nextID != 1 {
		t.Errorf("expected nextID=1, got %d", em.nextID)
	}
}

func TestOn(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(s string) {}
	sub := em.On("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("expected 1 listener, got %d", count)
	}

	sub.Unsubscribe()
	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("expected 0 listeners after unsubscribe, got %d", count)
	}
}

func TestAddListener(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(s string) {}
	sub := em.AddListener("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("expected 1 listener, got %d", count)
	}

	sub.Unsubscribe()
	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("expected 0 listeners after unsubscribe, got %d", count)
	}
}

func TestRemoveAllListeners(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("test", func(s string) {})
	em.On("test", func(s string) {})
	em.RemoveAllListeners("test")

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("expected 0 listeners, got %d", count)
	}

	// 不存在事件无副作用
	em.RemoveAllListeners("unknown")
}

func TestOnce(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0
	em.Once("test", func(s string) { called++ })

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("expected still 1 call, got %d", called)
	}

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("expected 0 listeners after once, got %d", count)
	}
}

func TestOnceUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0

	sub := em.Once("test", func(s string) { called++ })
	sub.Unsubscribe()

	if em.GetListenerCount("test") != 0 {
		t.Errorf("expected 0 listeners after unsubscribe, got %d", em.GetListenerCount("test"))
	}

	em.EmitSync("test", "data")
	if called != 0 {
		t.Errorf("expected 0 calls after unsubscribe, got %d", called)
	}
}

func TestEmit(t *testing.T) {
	em := NewEmitter[string, string]()
	var wg sync.WaitGroup
	wg.Add(2)

	em.On("test", func(s string) {
		defer wg.Done()
		if s != "hello" {
			t.Errorf("expected 'hello', got %q", s)
		}
	})
	em.On("test", func(s string) {
		defer wg.Done()
		if s != "hello" {
			t.Errorf("expected 'hello', got %q", s)
		}
	})

	em.Emit("test", "hello")
	wg.Wait()

	// 不存在的事件无副作用
	em.Emit("unknown", "data")
}

func TestEmitSync(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0
	em.On("test", func(s string) {
		called++
		if s != "sync" {
			t.Errorf("expected 'sync', got %q", s)
		}
	})

	em.EmitSync("test", "sync")
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}

	// 不存在的事件无副作用
	em.EmitSync("unknown", "data")
}

func TestEmitWait(t *testing.T) {
	em := NewEmitter[string, string]()
	called := false

	em.On("test", func(s string) {
		time.Sleep(50 * time.Millisecond)
		called = true
	})

	start := time.Now()
	em.EmitWait("test", "data")
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("EmitWait should block, took only %v", elapsed)
	}
	if !called {
		t.Error("listener should have been called")
	}
}

func TestEmitAsyncBehavior(t *testing.T) {
	em := NewEmitter[string, string]()
	done := make(chan bool)

	em.On("test", func(s string) {
		time.Sleep(100 * time.Millisecond)
		done <- true
	})

	start := time.Now()
	em.Emit("test", "data")
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Emit should return immediately, took %v", elapsed)
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("listener did not complete")
	}
}

func TestRecoverWith(t *testing.T) {
	em := NewEmitter[string, string]()
	recovered := false
	em.RecoverWith(func(event string, listener any, panicValue any) {
		recovered = true
		if panicValue == nil {
			t.Error("should receive a panic value")
		}
		if panicStr, ok := panicValue.(string); !ok || panicStr != "test panic" {
			t.Errorf("expected 'test panic', got %v", panicValue)
		}
	})

	em.On("test", func(s string) { panic("test panic") })
	em.EmitSync("test", "data")

	if !recovered {
		t.Error("RecoverWith should have been called")
	}
}

func TestPanicRecoveryWithoutRecoverer(t *testing.T) {
	em := NewEmitter[string, string]()
	done := make(chan bool)

	em.On("test", func(s string) {
		defer func() { done <- true }()
		panic("silent panic")
	})

	// Emit async — 若无 recover 整个 goroutine 会挂，后面的 done <- true 不会执行
	em.Emit("test", "data")

	select {
	case <-done:
		// panic 被内置 recover 捕获，defer 正常执行
	case <-time.After(200 * time.Millisecond):
		t.Error("panic was not recovered, goroutine died silently")
	}
}

func TestSetMaxListeners(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetMaxListeners(5)
	if em.maxListeners != 5 {
		t.Errorf("expected max=5, got %d", em.maxListeners)
	}

	em.SetMaxListeners(-1)
	if em.maxListeners != -1 {
		t.Errorf("expected max=-1 (unlimited), got %d", em.maxListeners)
	}
}

func TestGetListenerCount(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("test", func(s string) {})

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
	if count := em.GetListenerCount("unknown"); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestEvents(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("a", func(s string) {})
	em.On("b", func(s string) {})
	em.On("c", func(s string) {})

	events := em.Events()
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	eventSet := make(map[string]bool)
	for _, e := range events {
		eventSet[e] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !eventSet[expected] {
			t.Errorf("expected event %q in list", expected)
		}
	}
}

func TestTotalListenerCount(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("a", func(s string) {})
	em.On("a", func(s string) {})
	em.On("b", func(s string) {})

	if total := em.TotalListenerCount(); total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
}

func TestDifferentTypes(t *testing.T) {
	em := NewEmitter[string, User]()
	received := false

	sub := em.On("user_login", func(u User) {
		received = true
		if u.Name != "Alice" || u.Age != 30 {
			t.Errorf("expected User{Alice, 30}, got %v", u)
		}
	})
	defer sub.Unsubscribe()

	em.EmitSync("user_login", User{Name: "Alice", Age: 30})
	if !received {
		t.Error("listener should have been called")
	}
}

func TestIntEventKey(t *testing.T) {
	em := NewEmitter[int, string]()
	called := false

	em.On(100, func(s string) {
		called = true
		if s != "test" {
			t.Errorf("expected 'test', got %q", s)
		}
	})

	em.EmitSync(100, "test")
	if !called {
		t.Error("listener should have been called")
	}
}

func TestComplexDataTypes(t *testing.T) {
	type Message struct {
		ID      int
		Content string
		Tags    []string
	}

	em := NewEmitter[string, Message]()
	var received Message

	em.On("message", func(m Message) { received = m })

	expected := Message{ID: 1, Content: "Hello", Tags: []string{"urgent"}}
	em.EmitSync("message", expected)

	if received.ID != expected.ID || received.Content != expected.Content || len(received.Tags) != 1 {
		t.Errorf("expected %v, got %v", expected, received)
	}
}

func TestMultipleOnceListeners(t *testing.T) {
	em := NewEmitter[string, int]()
	count1, count2 := 0, 0

	em.Once("event", func(n int) { count1++ })
	em.Once("event", func(n int) { count2++ })

	em.EmitSync("event", 1)
	em.EmitSync("event", 2)

	if count1 != 1 || count2 != 1 {
		t.Errorf("expected count1=1, count2=1; got %d, %d", count1, count2)
	}
	if em.GetListenerCount("event") != 0 {
		t.Errorf("expected 0 listeners, got %d", em.GetListenerCount("event"))
	}
}

func TestUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0

	sub := em.On("test", func(s string) { called++ })

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}

	sub.Unsubscribe()
	if em.GetListenerCount("test") != 0 {
		t.Errorf("expected 0 listeners, got %d", em.GetListenerCount("test"))
	}

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("expected still 1 call, got %d", called)
	}
}

func TestDoubleUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	sub := em.On("test", func(s string) {})

	sub.Unsubscribe()
	// 第二次 Unsubscribe 应为安全 no-op，不 panic
	sub.Unsubscribe()

	if em.GetListenerCount("test") != 0 {
		t.Errorf("expected 0 listeners, got %d", em.GetListenerCount("test"))
	}
}

func TestSetConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(2)

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0
	totalCalls := 0

	for i := 0; i < 10; i++ {
		em.On("test", func(s string) {
			mu.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			current--
			totalCalls++
			mu.Unlock()
		})
	}

	em.EmitWait("test", "data")

	if totalCalls != 10 {
		t.Errorf("expected 10 calls, got %d", totalCalls)
	}
	if maxConcurrent > 2 {
		t.Errorf("maxConcurrent should be <= 2, got %d", maxConcurrent)
	}
}

func TestSetConcurrencyEmit(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(3)

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0
	done := make(chan struct{})
	remaining := 6

	for i := 0; i < 6; i++ {
		em.On("test", func(s string) {
			mu.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			current--
			remaining--
			allDone := remaining == 0
			mu.Unlock()

			if allDone {
				close(done)
			}
		})
	}

	em.Emit("test", "data")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for listeners")
	}

	if maxConcurrent > 3 {
		t.Errorf("maxConcurrent should be <= 3, got %d", maxConcurrent)
	}
}

func TestSetConcurrencyZeroMeansUnlimited(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(2)
	em.SetConcurrency(0)

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0

	for i := 0; i < 5; i++ {
		em.On("test", func(s string) {
			mu.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			current--
			mu.Unlock()
		})
	}

	em.EmitWait("test", "data")

	if maxConcurrent < 3 {
		t.Errorf("without concurrency limit expected higher concurrency, got %d", maxConcurrent)
	}
}

func TestSetConcurrencyDoesNotAffectEmitSync(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(1)

	callOrder := make([]int, 0, 3)
	var mu sync.Mutex

	for i := range 3 {
		idx := i
		em.On("test", func(s string) {
			mu.Lock()
			callOrder = append(callOrder, idx)
			mu.Unlock()
		})
	}

	em.EmitSync("test", "data")

	if len(callOrder) != 3 {
		t.Errorf("expected 3 calls, got %d", len(callOrder))
	}
}

func TestChaining(t *testing.T) {
	em := NewEmitter[string, string]()

	em.SetMaxListeners(20).
		SetConcurrency(5).
		RecoverWith(func(event string, listener any, panicValue any) {}).
		EmitSync("init", "ready")

	// 验证链式调用后 emitter 正常可用
	em.On("test", func(s string) {})
	if em.GetListenerCount("test") != 1 {
		t.Errorf("expected 1 listener after chaining, got %d", em.GetListenerCount("test"))
	}
}

func TestConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			em.On("test", func(s string) {})
		}()
		go func() {
			defer wg.Done()
			em.Emit("test", "data")
		}()
	}

	wg.Wait()
	// 依赖 go test -race 检测数据竞争
}

// testLogger 实现 Logger 接口用于测试
type testLogger struct {
	warnings []string
	mu       sync.Mutex
}

func (l *testLogger) Warnf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnings = append(l.warnings, format)
}

func TestMaxListenersWarning(t *testing.T) {
	em := NewEmitter[string, string]()
	logger := &testLogger{}
	em.SetLogger(logger)
	em.SetMaxListeners(1)

	em.On("test", func(s string) {})
	if len(logger.warnings) != 0 {
		t.Error("no warning expected for first listener")
	}

	em.On("test", func(s string) {})
	if len(logger.warnings) != 1 {
		t.Error("expected warning for exceeding max listeners")
	}

	// 即便超出限制，监听器仍然被添加
	if em.GetListenerCount("test") != 2 {
		t.Errorf("expected 2 listeners despite warning, got %d", em.GetListenerCount("test"))
	}
}
