package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNew 测试创建工作协程池
func TestNew(t *testing.T) {
	pool := New(0) // 测试最小值边界
	if pool.Size() != 1 {
		t.Errorf("New(0) should create pool with size 1, got %d", pool.Size())
	}

	pool = New(5)
	if pool.Size() != 5 {
		t.Errorf("New(5) should create pool with size 5, got %d", pool.Size())
	}
	if pool.Stopped() {
		t.Error("Newly created pool should not be stopped")
	}
}

// TestSubmit 测试任务提交
func TestSubmit(t *testing.T) {
	pool := New(2)
	var counter int32
	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond) // 模拟任务耗时
			atomic.AddInt32(&counter, 1)
			wg.Done()
		})
	}

	wg.Wait()
	if counter != 3 {
		t.Errorf("Expected 3 tasks to complete, got %d", counter)
	}

	pool.Submit(nil)                  // 测试空任务
	time.Sleep(10 * time.Millisecond) // 等待处理
	if pool.WaitingQueueSize() != 0 {
		t.Errorf("Submitting nil task should not increase waiting queue, got %d", pool.WaitingQueueSize())
	}
}

// TestSubmitWait 测试同步任务提交
func TestSubmitWait(t *testing.T) {
	pool := New(1)
	var result int
	pool.SubmitWait(func() {
		time.Sleep(50 * time.Millisecond)
		result = 42
	})
	if result != 42 {
		t.Errorf("SubmitWait should complete task, expected result 42, got %d", result)
	}

	pool.SubmitWait(nil) // 测试空任务
	if result != 42 {
		t.Errorf("SubmitWait with nil should not modify state, result changed to %d", result)
	}
}

// TestStop 测试停止协程池（不等待队列任务）
func TestStop(t *testing.T) {
	pool := New(1)
	var counter int32
	pool.Submit(func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	})
	pool.Submit(func() { // 此任务将被放弃
		atomic.AddInt32(&counter, 1)
	})

	time.Sleep(10 * time.Millisecond) // 确保第一个任务开始执行
	pool.Stop()
	if !pool.Stopped() {
		t.Error("Pool should be stopped after Stop()")
	}
	if counter != 1 {
		t.Errorf("Only running task should complete, expected counter 1, got %d", counter)
	}

	assertPanics(t, "Submit after Stop should panic", func() {
		pool.Submit(func() {})
	})
}

// TestStopWait 测试停止并等待所有任务完成
func TestStopWait(t *testing.T) {
	pool := New(1)
	var counter int32
	var wg sync.WaitGroup
	wg.Add(2)

	pool.Submit(func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
		wg.Done()
	})
	pool.Submit(func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
		wg.Done()
	})

	pool.StopWait()
	wg.Wait()
	if !pool.Stopped() {
		t.Error("Pool should be stopped after StopWait()")
	}
	if counter != 2 {
		t.Errorf("All queued tasks should complete, expected counter 2, got %d", counter)
	}
}

// TestWaitingQueueSize 测试等待队列大小
func TestWaitingQueueSize(t *testing.T) {
	pool := New(1)
	var counter int32
	pool.Submit(func() {
		time.Sleep(100 * time.Millisecond) // 占用唯一工作协程
		atomic.AddInt32(&counter, 1)
	})

	for i := 0; i < 3; i++ {
		pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	}

	time.Sleep(10 * time.Millisecond) // 等待任务进入队列
	if size := pool.WaitingQueueSize(); size != 3 {
		t.Errorf("Expected waiting queue size 3, got %d", size)
	}

	pool.StopWait()
	if pool.WaitingQueueSize() != 0 {
		t.Errorf("Waiting queue should be empty after StopWait, got %d", pool.WaitingQueueSize())
	}
	if counter != 4 {
		t.Errorf("All tasks should complete, expected counter 4, got %d", counter)
	}
}

// TestPause 测试暂停功能
func TestPause(t *testing.T) {
	pool := New(2)
	var counter int32
	var wg sync.WaitGroup
	wg.Add(2)

	// 提交任务以占用工作协程
	for i := 0; i < 2; i++ {
		pool.Submit(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
			wg.Done()
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	pool.Pause(ctx) // 暂停协程池

	// 在暂停期间提交任务
	pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	time.Sleep(10 * time.Millisecond) // 确保任务进入队列
	if pool.WaitingQueueSize() != 1 {
		t.Errorf("Task should be queued during pause, expected queue size 1, got %d", pool.WaitingQueueSize())
	}

	wg.Wait()
	<-ctx.Done()                      // 等待暂停结束
	time.Sleep(60 * time.Millisecond) // 等待第三个任务完成
	if counter != 3 {
		t.Errorf("All tasks should complete after pause, expected counter 3, got %d", counter)
	}
}

// TestIdleWorkerShutdown 测试空闲工作协程的关闭
func TestIdleWorkerShutdown(t *testing.T) {
	pool := New(3)
	var counter int32
	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
			wg.Done()
		})
	}

	wg.Wait()
	time.Sleep(idleTimeout + 100*time.Millisecond) // 等待空闲超时
	pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	time.Sleep(10 * time.Millisecond) // 确保新任务被处理

	if pool.WaitingQueueSize() != 0 {
		t.Errorf("No tasks should be waiting after idle timeout, got %d", pool.WaitingQueueSize())
	}
	pool.StopWait()
	if counter != 4 {
		t.Errorf("All tasks should complete, expected counter 4, got %d", counter)
	}
}

// TestMultiplePause 测试多次暂停
func TestMultiplePause(t *testing.T) {
	pool := New(1)
	var counter int32

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	go func() {
		pool.Pause(ctx1)
	}()

	time.Sleep(10 * time.Millisecond) // 确保第一个 Pause 开始
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	pool.Pause(ctx2) // 第二个 Pause 应等待第一个完成

	pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	time.Sleep(10 * time.Millisecond)
	if pool.WaitingQueueSize() != 1 {
		t.Errorf("Task should be queued during pause, expected queue size 1, got %d", pool.WaitingQueueSize())
	}

	cancel1()                         // 取消第一个 Pause
	time.Sleep(60 * time.Millisecond) // 等待任务执行
	if counter != 1 {
		t.Errorf("Task should complete after first pause ends, expected counter 1, got %d", counter)
	}
}

// assertPanics 检查函数是否引发 panic
func assertPanics(t *testing.T, msg string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: didn't panic as expected", msg)
		}
	}()
	f()
}
