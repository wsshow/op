package emission

import (
	"sync"
	"testing"
	"time"
)

// TestNewEmitter 测试创建新的 Emitter 实例
func TestNewEmitter(t *testing.T) {
	em := NewEmitter[string]()
	if em.maxListeners != DefaultMaxListeners {
		t.Errorf("NewEmitter should set maxListeners to %d, got %d", DefaultMaxListeners, em.maxListeners)
	}
	if len(em.events) != 0 {
		t.Errorf("NewEmitter should initialize empty events map, got %d events", len(em.events))
	}
	if len(em.listenerIDs) != 0 {
		t.Errorf("NewEmitter should initialize empty listenerIDs map, got %d entries", len(em.listenerIDs))
	}
	if em.nextID != 1 {
		t.Errorf("NewEmitter should initialize nextID to 1, got %d", em.nextID)
	}
}

// TestAddListener 测试添加监听器
func TestAddListener(t *testing.T) {
	em := NewEmitter[string]()
	listener := func(args ...string) {}
	em.AddListener("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("AddListener should add one listener, got %d", count)
	}

	// 测试超过最大监听器数量的警告
	em.SetMaxListeners(1)
	em.AddListener("test", func(args ...string) {}) // 应打印警告
	if count := em.GetListenerCount("test"); count != 2 {
		t.Errorf("AddListener should add second listener despite warning, got %d", count)
	}
}

// TestOn 测试 On 方法（AddListener 的别名）
func TestOn(t *testing.T) {
	em := NewEmitter[string]()
	listener := func(args ...string) {}
	em.On("test", listener)

	if count := em.GetListenerCount("test"); count != 1 {
		t.Errorf("On should add one listener, got %d", count)
	}
}

// TestRemoveListener 测试移除监听器
func TestRemoveListener(t *testing.T) {
	em := NewEmitter[string]()
	listener := func(args ...string) {}
	em.On("test", listener)
	em.RemoveListener("test", listener)

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("RemoveListener should remove listener, got %d remaining", count)
	}

	// 测试移除不存在的事件
	em.RemoveListener("unknown", listener) // 应无错误
}

// TestOff 测试 Off 方法（RemoveListener 的别名）
func TestOff(t *testing.T) {
	em := NewEmitter[string]()
	listener := func(args ...string) {}
	em.On("test", listener)
	em.Off("test", listener)

	if count := em.GetListenerCount("test"); count != 0 {
		t.Errorf("Off should remove listener, got %d remaining", count)
	}
}

// TestOnce 测试一次性监听器
func TestOnce(t *testing.T) {
	em := NewEmitter[string]()
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
	em := NewEmitter[string]()
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

	em.On("test", listener1).On("test", listener2)
	em.Emit("test", "a", "b")
	wg.Wait()

	// 测试不存在的事件
	em.Emit("unknown") // 应无错误
}

// TestEmitSync 测试同步事件触发
func TestEmitSync(t *testing.T) {
	em := NewEmitter[string]()
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
	em := NewEmitter[string]()
	recovered := false
	em.RecoverWith(func(event string, listener interface{}, err error) {
		recovered = true
		if err == nil {
			t.Error("RecoverWith should receive an error")
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
	em := NewEmitter[string]()
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
	em := NewEmitter[string]()
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
	em := NewEmitter[string]()
	var wg sync.WaitGroup
	listener := func(args ...string) {}

	// 并发添加和移除监听器
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			em.On("test", listener)
		}()
		go func() {
			defer wg.Done()
			em.Off("test", listener)
		}()
	}

	wg.Wait()
	// 检查是否发生数据竞争（依赖 go test -race 检测）
}
