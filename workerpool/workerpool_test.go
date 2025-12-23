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

	// 等待任务开始执行
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// 在暂停期间提交任务（启动 Pause 的 goroutine，不阻塞当前测试）
	pauseDone := make(chan struct{})
	go func() {
		pool.Pause(ctx) // 暂停协程池（会阻塞直到 pause 任务开始执行）
		close(pauseDone)
	}()

	time.Sleep(150 * time.Millisecond) // 等待原任务完成 + pause 任务开始
	pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	time.Sleep(20 * time.Millisecond) // 确保任务进入队列

	queueSize := pool.WaitingQueueSize()
	// 新提交的任务应该在队列中，因为所有 worker 都被 pause 任务阻塞
	if queueSize < 1 {
		t.Errorf("Task should be queued during pause, expected queue size >= 1, got %d", queueSize)
	}

	wg.Wait()
	<-pauseDone                       // 等待 Pause 返回
	time.Sleep(50 * time.Millisecond) // 等待第三个任务完成
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

	time.Sleep(10 * time.Millisecond) // 确保第一个 Pause 开始并占据 worker

	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	pause2Done := make(chan struct{})
	go func() {
		pool.Pause(ctx2) // 第二个 Pause 应等待第一个完成
		close(pause2Done)
	}()

	time.Sleep(10 * time.Millisecond) // 给第二个 Pause 时间尝试获取锁
	pool.Submit(func() { atomic.AddInt32(&counter, 1) })
	time.Sleep(10 * time.Millisecond)
	queueSize := pool.WaitingQueueSize()
	// 任务应该在队列中，因为 worker 被第一个 Pause 占据
	if queueSize != 1 {
		t.Errorf("Task should be queued during pause, expected queue size 1, got %d", queueSize)
	}

	cancel1()                         // 取消第一个 Pause
	<-pause2Done                      // 等待第二个 Pause 完成
	time.Sleep(20 * time.Millisecond) // 等待任务执行
	if counter != 1 {
		t.Errorf("Task should complete after pauses end, expected counter 1, got %d", counter)
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
