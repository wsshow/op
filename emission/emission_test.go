package emission

import (
	"fmt"
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

func TestConcurrentUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	sub := em.On("test", func(string) {})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub.Unsubscribe()
		}()
	}
	wg.Wait()
	if count := em.GetListenerCount("test"); count != 0 {
		t.Fatalf("listener count = %d, want 0", count)
	}

	var nilSub *Subscription[string, string]
	nilSub.Unsubscribe()
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

func TestEventsEmptyEmitter(t *testing.T) {
	em := NewEmitter[string, string]()
	events := em.Events()
	if len(events) != 0 {
		t.Errorf("expected 0 events for new emitter, got %d", len(events))
	}
}

func TestEventsAfterRemoveAll(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("a", func(s string) {})
	em.On("b", func(s string) {})

	em.RemoveAllListeners("a")
	events := em.Events()
	if len(events) != 1 {
		t.Errorf("expected 1 event after removing 'a', got %d", len(events))
	}

	em.RemoveAllListeners("b")
	events = em.Events()
	if len(events) != 0 {
		t.Errorf("expected 0 events after removing all, got %d", len(events))
	}
}

func TestTotalListenerCountEmpty(t *testing.T) {
	em := NewEmitter[string, string]()
	if total := em.TotalListenerCount(); total != 0 {
		t.Errorf("expected 0 total for new emitter, got %d", total)
	}
}

func TestTotalListenerCountAfterChanges(t *testing.T) {
	em := NewEmitter[string, string]()
	em.On("a", func(s string) {})
	em.On("a", func(s string) {})
	sub := em.On("b", func(s string) {})

	if total := em.TotalListenerCount(); total != 3 {
		t.Errorf("expected 3, got %d", total)
	}

	sub.Unsubscribe()
	if total := em.TotalListenerCount(); total != 2 {
		t.Errorf("expected 2 after unsubscribe, got %d", total)
	}

	em.RemoveAllListeners("a")
	if total := em.TotalListenerCount(); total != 0 {
		t.Errorf("expected 0 after RemoveAll, got %d", total)
	}
}

func TestPointerTypeParameter(t *testing.T) {
	type Data struct{ Value int }
	em := NewEmitter[string, *Data]()
	var received *Data

	em.On("ptr_event", func(d *Data) { received = d })

	original := &Data{Value: 42}
	em.EmitSync("ptr_event", original)

	if received != original {
		t.Error("pointer identity should be preserved")
	}
	if received.Value != 42 {
		t.Errorf("expected 42, got %d", received.Value)
	}
}

func TestInterfaceTypeParameter(t *testing.T) {
	em := NewEmitter[string, any]()
	var received any

	em.On("any_event", func(v any) { received = v })
	em.EmitSync("any_event", "a string value")

	if s, ok := received.(string); !ok || s != "a string value" {
		t.Errorf("expected 'a string value', got %v", received)
	}
}

func TestEmitWaitWithConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(2)

	var mu sync.Mutex
	executed := make([]int, 0, 4)

	for i := range 4 {
		idx := i
		em.On("test", func(s string) {
			time.Sleep(20 * time.Millisecond)
			mu.Lock()
			executed = append(executed, idx)
			mu.Unlock()
		})
	}

	em.EmitWait("test", "data")

	if len(executed) != 4 {
		t.Errorf("expected 4 executions, got %d", len(executed))
	}
}

func TestRecoverWithAsyncEmit(t *testing.T) {
	em := NewEmitter[string, string]()
	recovered := make(chan any, 1)

	em.RecoverWith(func(event string, listener any, panicValue any) {
		recovered <- panicValue
	})

	em.On("test", func(s string) { panic("async panic") })
	em.Emit("test", "data")

	select {
	case pv := <-recovered:
		if pv != "async panic" {
			t.Errorf("expected 'async panic', got %v", pv)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for async recover")
	}
}

func TestConcurrentSetConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			em.SetConcurrency(2)
		}()
		go func() {
			defer wg.Done()
			em.EmitWait("test", "data")
		}()
	}

	wg.Wait()
	// 依赖 go test -race 检测数据竞争
}

func TestEmitReturnValueChaining(t *testing.T) {
	em := NewEmitter[string, int]()

	result := em.
		EmitSync("a", 1).
		EmitSync("b", 2).
		Emit("c", 3).
		EmitWait("d", 4)

	if result != em {
		t.Error("chained emits should return the same emitter")
	}
}

// testLogger 实现 Logger 接口用于测试
type testLogger struct {
	warnings []string
	mu       sync.Mutex
}

type loggerFunc func(format string, args ...any)

func (f loggerFunc) Warnf(format string, args ...any) { f(format, args...) }

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

func TestMaxListenersLoggerCanReenterEmitter(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetMaxListeners(0)
	logged := make(chan struct{})
	em.SetLogger(loggerFunc(func(string, ...any) {
		_ = em.GetListenerCount("test")
		close(logged)
	}))

	registered := make(chan struct{})
	go func() {
		em.On("test", func(string) {})
		close(registered)
	}()
	select {
	case <-registered:
	case <-time.After(time.Second):
		t.Fatal("listener registration deadlocked in reentrant logger")
	}
	select {
	case <-logged:
	default:
		t.Fatal("logger was not called")
	}
}

// Example 演示基本的事件发射和监听流程。
func Example_basic() {
	em := NewEmitter[string, string]()
	em.On("greet", func(msg string) {
		fmt.Println("received:", msg)
	})
	em.EmitSync("greet", "hello")
	// Output: received: hello
}

// Example 演示 Once 一次性监听器。
func Example_once() {
	em := NewEmitter[string, int]()
	counter := 0
	em.Once("increment", func(n int) {
		counter += n
	})
	em.EmitSync("increment", 1)
	em.EmitSync("increment", 10) // 不会执行
	fmt.Println("counter:", counter)
	// Output: counter: 1
}

// Example 演示 Subscription 取消监听。
func ExampleSubscription_Unsubscribe() {
	em := NewEmitter[string, string]()
	sub := em.On("event", func(msg string) {
		fmt.Println(msg)
	})
	sub.Unsubscribe()
	em.EmitSync("event", "should not print")
	fmt.Println("done")
	// Output: done
}

// Example 演示链式配置和并发控制。
func Example_chaining() {
	em := NewEmitter[string, string]()
	em.SetMaxListeners(20).
		SetConcurrency(4).
		RecoverWith(func(event string, listener any, panicValue any) {
			fmt.Println("recovered:", panicValue)
		})

	em.On("task", func(data string) {
		fmt.Println("processing:", data)
	})
	em.EmitSync("task", "job-1")
	// Output: processing: job-1
}
