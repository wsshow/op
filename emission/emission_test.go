package emission

import (
	"sync"
	"testing"
	"time"
)

// User 示例数据结构
type User struct {
	Name string
	Age  int
}

// TestNewEmitter 测试创建新的 Emitter 实例
func TestNewEmitter(t *testing.T) {
	em := NewEmitter[string, string]()
	if em.maxListeners != DefaultMaxListeners {
		t.Errorf("NewEmitter should set maxListeners to %d, got %d", DefaultMaxListeners, em.maxListeners)
	}
	if len(em.events) != 0 {
		t.Errorf("NewEmitter should initialize empty events map, got %d events", len(em.events))
	}
	if em.nextID != 1 {
		t.Errorf("NewEmitter should initialize nextID to 1, got %d", em.nextID)
	}
}

// TestAddListener 测试添加监听器
func TestAddListener(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(args ...string) {}
	unsubscribe := em.AddListener("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("AddListener should add one listener, got %d", count)
	}

	// 测试取消监听器
	unsubscribe()
	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("Unsubscribe should remove listener, got %d", count)
	}

	// 测试超过最大监听器数量的警告
	em.SetMaxListeners(1)
	em.AddListener("test", func(args ...string) {}) // 如果设置了logger会记录警告
	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("AddListener should add listener despite warning, got %d", count)
	}
}

// TestOn 测试 On 方法（AddListener 的别名）
func TestOn(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(args ...string) {}
	unsubscribe := em.On("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("On should add one listener, got %d", count)
	}

	// 测试取消
	unsubscribe()
	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("Unsubscribe should remove listener, got %d", count)
	}
}

// TestRemoveAllListeners 测试移除所有监听器
func TestRemoveAllListeners(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(args ...string) {}
	em.On("test", listener)
	em.On("test", func(args ...string) {})
	em.RemoveAllListeners("test")

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("RemoveAllListeners should remove all listeners, got %d remaining", count)
	}

	// 测试移除不存在的事件
	em.RemoveAllListeners("unknown") // 应无错误
}

// TestOnce 测试一次性监听器
func TestOnce(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0
	listener := func(args ...string) { called++ }
	em.Once("test", listener)

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("Once listener should be called once, got %d calls", called)
	}

	em.EmitSync("test", "data")
	if called != 1 {
		t.Errorf("Once listener should not be called again, got %d calls", called)
	}

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("Once should remove listener after call, got %d remaining", count)
	}
}

// TestEmit 测试异步事件触发
func TestEmit(t *testing.T) {
	em := NewEmitter[string, string]()
	var wg sync.WaitGroup
	wg.Add(2)

	listener1 := func(args ...string) {
		defer wg.Done()
		if len(args) != 2 || args[0] != "a" || args[1] != "b" {
			t.Errorf("Listener1 expected [a b], got %v", args)
		}
	}
	listener2 := func(args ...string) {
		defer wg.Done()
		if len(args) != 2 || args[0] != "a" || args[1] != "b" {
			t.Errorf("Listener2 expected [a b], got %v", args)
		}
	}

	em.On("test", listener1)
	em.On("test", listener2)
	em.Emit("test", "a", "b")
	wg.Wait()

	// 测试不存在的事件
	em.Emit("unknown")                // 应无错误
	time.Sleep(50 * time.Millisecond) // 等待异步操作完成
}

// TestEmitSync 测试同步事件触发
func TestEmitSync(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0
	listener := func(args ...string) {
		called++
		if len(args) != 1 || args[0] != "sync" {
			t.Errorf("Listener expected [sync], got %v", args)
		}
	}

	em.On("test", listener)
	em.EmitSync("test", "sync")
	if called != 1 {
		t.Errorf("EmitSync should call listener once, got %d calls", called)
	}

	// 测试不存在的事件
	em.EmitSync("unknown") // 应无错误
}

// TestRecoverWith 测试恢复监听器处理 panic
func TestRecoverWith(t *testing.T) {
	em := NewEmitter[string, string]()
	recovered := false
	em.RecoverWith(func(event string, listener interface{}, panicValue interface{}) {
		recovered = true
		if panicValue == nil {
			t.Error("RecoverWith should receive a panic value")
		}
		// 验证 panic 值是预期的字符串
		if panicStr, ok := panicValue.(string); !ok || panicStr != "test panic" {
			t.Errorf("Expected panic value 'test panic', got %v", panicValue)
		}
	})

	listener := func(args ...string) {
		panic("test panic")
	}
	em.On("test", listener)
	em.EmitSync("test", "data")

	time.Sleep(100 * time.Millisecond) // 等待恢复处理
	if !recovered {
		t.Error("RecoverWith should have been called")
	}
}

// TestSetMaxListeners 测试设置最大监听器数量
func TestSetMaxListeners(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetMaxListeners(5)
	if em.maxListeners != 5 {
		t.Errorf("SetMaxListeners should set max to 5, got %d", em.maxListeners)
	}

	em.SetMaxListeners(-1)
	if em.maxListeners != -1 {
		t.Errorf("SetMaxListeners should set max to -1 (unlimited), got %d", em.maxListeners)
	}
}

// TestGetListenerCount 测试获取监听器数量
func TestGetListenerCount(t *testing.T) {
	em := NewEmitter[string, string]()
	listener := func(args ...string) {}
	em.On("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("GetListenerCount should return 1, got %d", count)
	}

	if count := em.GetListenerCount("unknown"); count != 0 {
		t.Errorf("GetListenerCount for unknown event should return 0, got %d", count)
	}
}

// TestConcurrency 测试并发安全性
func TestConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	var wg sync.WaitGroup
	listener := func(args ...string) {}

	// 并发添加和触发事件
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			em.On("test", listener)
		}()
		go func() {
			defer wg.Done()
			em.Emit("test", "data")
		}()
	}

	wg.Wait()
	// 检查是否发生数据竞争（依赖 go test -race 检测）
}

// TestDifferentTypes 测试事件类型和参数类型分离
func TestDifferentTypes(t *testing.T) {
	// 测试事件名为 string，参数为自定义结构体
	em := NewEmitter[string, User]()
	received := false

	unsubscribe := em.On("user_login", func(users ...User) {
		received = true
		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
		if users[0].Name != "Alice" || users[0].Age != 30 {
			t.Errorf("Expected User{Alice, 30}, got %v", users[0])
		}
	})
	defer unsubscribe()

	em.EmitSync("user_login", User{Name: "Alice", Age: 30})

	if !received {
		t.Error("Listener should have been called")
	}
}

// TestIntEventKey 测试使用整数作为事件标识
func TestIntEventKey(t *testing.T) {
	em := NewEmitter[int, string]()
	called := false

	em.On(100, func(args ...string) {
		called = true
		if len(args) != 1 || args[0] != "test" {
			t.Errorf("Expected [test], got %v", args)
		}
	})

	em.EmitSync(100, "test")

	if !called {
		t.Error("Listener should have been called")
	}
}

// TestComplexDataTypes 测试复杂数据类型
func TestComplexDataTypes(t *testing.T) {
	type Message struct {
		ID      int
		Content string
		Tags    []string
	}

	em := NewEmitter[string, Message]()
	var receivedMsg Message

	em.On("message", func(msgs ...Message) {
		if len(msgs) > 0 {
			receivedMsg = msgs[0]
		}
	})

	expectedMsg := Message{
		ID:      1,
		Content: "Hello World",
		Tags:    []string{"urgent", "important"},
	}

	em.EmitSync("message", expectedMsg)

	if receivedMsg.ID != expectedMsg.ID ||
		receivedMsg.Content != expectedMsg.Content ||
		len(receivedMsg.Tags) != len(expectedMsg.Tags) {
		t.Errorf("Expected %v, got %v", expectedMsg, receivedMsg)
	}
}

// TestMultipleOnceListeners 测试多个 Once 监听器
func TestMultipleOnceListeners(t *testing.T) {
	em := NewEmitter[string, int]()
	count1 := 0
	count2 := 0

	em.Once("event", func(args ...int) { count1++ })
	em.Once("event", func(args ...int) { count2++ })

	em.EmitSync("event", 1)
	em.EmitSync("event", 2)

	if count1 != 1 || count2 != 1 {
		t.Errorf("Each Once listener should be called exactly once, got count1=%d, count2=%d", count1, count2)
	}

	if em.GetListenerCount("event") != 0 {
		t.Errorf("All Once listeners should be removed, got %d remaining", em.GetListenerCount("event"))
	}
}

// TestUnsubscribe 测试取消函数
func TestUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0

	unsubscribe := em.On("test", func(args ...string) {
		called++
	})

	em.EmitSync("test")
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}

	// 取消监听器
	unsubscribe()

	if em.GetListenerCount("test") != 0 {
		t.Errorf("Expected 0 listeners after unsubscribe, got %d", em.GetListenerCount("test"))
	}

	em.EmitSync("test")
	if called != 1 {
		t.Errorf("Expected still 1 call after unsubscribe, got %d", called)
	}
}

// TestOnceUnsubscribe 测试Once监听器的取消
func TestOnceUnsubscribe(t *testing.T) {
	em := NewEmitter[string, string]()
	called := 0

	unsubscribe := em.Once("test", func(args ...string) {
		called++
	})

	// 在触发前取消
	unsubscribe()

	if em.GetListenerCount("test") != 0 {
		t.Errorf("Expected 0 listeners after unsubscribe, got %d", em.GetListenerCount("test"))
	}

	em.EmitSync("test")
	if called != 0 {
		t.Errorf("Expected 0 calls after unsubscribe, got %d", called)
	}
}

// TestEmitAsyncBehavior 测试Emit的真正异步行为
func TestEmitAsyncBehavior(t *testing.T) {
	em := NewEmitter[string, string]()
	done := make(chan bool)

	em.On("test", func(args ...string) {
		time.Sleep(100 * time.Millisecond)
		done <- true
	})

	start := time.Now()
	em.Emit("test")
	elapsed := time.Since(start)

	// Emit应该立即返回
	if elapsed > 50*time.Millisecond {
		t.Errorf("Emit should return immediately, took %v", elapsed)
	}

	// 等待监听器完成
	select {
	case <-done:
		// 成功
	case <-time.After(200 * time.Millisecond):
		t.Error("Listener did not complete")
	}
}

// TestEmitWait 测试EmitWait会等待
func TestEmitWait(t *testing.T) {
	em := NewEmitter[string, string]()
	called := false

	em.On("test", func(args ...string) {
		time.Sleep(100 * time.Millisecond)
		called = true
	})

	start := time.Now()
	em.EmitWait("test")
	elapsed := time.Since(start)

	// EmitWait应该等待监听器完成
	if elapsed < 100*time.Millisecond {
		t.Errorf("EmitWait should wait for listeners, took only %v", elapsed)
	}

	if !called {
		t.Error("Listener should have been called")
	}
}

// TestSetConcurrency 测试并发度限制
func TestSetConcurrency(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(2) // 最多同时执行 2 个监听器

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0
	totalCalls := 0

	for i := 0; i < 10; i++ {
		em.On("test", func(args ...string) {
			mu.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond) // 模拟耗时操作

			mu.Lock()
			current--
			totalCalls++
			mu.Unlock()
		})
	}

	em.EmitWait("test")

	if totalCalls != 10 {
		t.Errorf("Expected 10 calls, got %d", totalCalls)
	}
	if maxConcurrent > 2 {
		t.Errorf("Max concurrent should be <= 2, got %d", maxConcurrent)
	}
}

// TestSetConcurrencyEmit 测试异步 Emit 的并发度限制
func TestSetConcurrencyEmit(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(3)

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0
	done := make(chan struct{})
	remaining := 6

	for i := 0; i < 6; i++ {
		em.On("test", func(args ...string) {
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

	em.Emit("test")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for listeners")
	}

	if maxConcurrent > 3 {
		t.Errorf("Max concurrent should be <= 3, got %d", maxConcurrent)
	}
}

// TestSetConcurrencyZeroMeansUnlimited 测试设置 0 表示无限制
func TestSetConcurrencyZeroMeansUnlimited(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(2)
	em.SetConcurrency(0) // 取消限制

	var mu sync.Mutex
	maxConcurrent := 0
	current := 0

	for i := 0; i < 5; i++ {
		em.On("test", func(args ...string) {
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

	em.EmitWait("test")

	// 无限制时，5 个监听器应当全部并发
	if maxConcurrent < 3 {
		t.Errorf("Without concurrency limit, expected higher concurrency, got %d", maxConcurrent)
	}
}

// TestSetConcurrencyDoesNotAffectEmitSync 测试并发度不影响 EmitSync
func TestSetConcurrencyDoesNotAffectEmitSync(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(1)

	callOrder := make([]int, 0, 3)
	var mu sync.Mutex

	for i := range 3 {
		idx := i
		em.On("test", func(args ...string) {
			mu.Lock()
			callOrder = append(callOrder, idx)
			mu.Unlock()
		})
	}

	em.EmitSync("test")

	if len(callOrder) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(callOrder))
	}
}

// TestWorkerPoolActuallyLimitsGoroutines 测试 worker pool 真正限制了 goroutine 创建数量
func TestWorkerPoolActuallyLimitsGoroutines(t *testing.T) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(3) // 最多 3 个 worker goroutine

	var mu sync.Mutex
	activeGoroutines := 0
	maxActive := 0
	totalExecutions := 0

	// 添加 20 个慢速监听器
	for i := 0; i < 20; i++ {
		em.On("test", func(args ...string) {
			mu.Lock()
			activeGoroutines++
			if activeGoroutines > maxActive {
				maxActive = activeGoroutines
			}
			totalExecutions++
			mu.Unlock()

			time.Sleep(20 * time.Millisecond) // 模拟耗时操作

			mu.Lock()
			activeGoroutines--
			mu.Unlock()
		})
	}

	em.EmitWait("test")

	if totalExecutions != 20 {
		t.Errorf("Expected 20 executions, got %d", totalExecutions)
	}

	// worker pool 模式下，活跃执行数应该 <= concurrency
	if maxActive > 3 {
		t.Errorf("With SetConcurrency(3) and worker pool, maxActive should be <= 3, got %d", maxActive)
	}
}
