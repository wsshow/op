// Package workerpool 提供了一个高性能的工作协程池实现，
// 支持限制并发任务数、任务排队、暂停/恢复以及优雅停止等功能。
package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wsshow/op/deque"
)

// DefaultIdleTimeout 是工作协程的默认空闲超时时间。
// 若工作协程空闲超过此时间且无新任务到来，该协程将被自动回收。
const DefaultIdleTimeout = 2 * time.Second

// Option 定义 WorkerPool 的可选配置函数。
type Option func(*WorkerPool)

// WithIdleTimeout 设置工作协程的空闲超时时间。
// 若 d <= 0，将使用 DefaultIdleTimeout。
func WithIdleTimeout(d time.Duration) Option {
	return func(p *WorkerPool) {
		if d > 0 {
			p.idleTimeout = d
		}
	}
}

// WithPanicHandler 设置 panic 处理器，当任务发生 panic 时调用。
// 若未设置，任务中的 panic 将被静默恢复，不会导致程序崩溃。
func WithPanicHandler(handler func(any)) Option {
	return func(p *WorkerPool) {
		p.panicHandler = handler
	}
}

// WorkerPool 是一个工作协程池，限制并发执行任务的协程数量不超过指定最大值。
// 当所有工作协程繁忙时，新任务将被放入等待队列。
// 空闲的工作协程在超过空闲超时时间后会被自动回收。
type WorkerPool struct {
	maxWorkers  int
	idleTimeout time.Duration

	taskChan     chan func()
	workerChan   chan func()
	stopSignal   chan struct{}
	stoppedChan  chan struct{}
	waitingQueue deque.Deque[func()]

	stopMutex  sync.Mutex
	pauseMutex sync.Mutex
	stopOnce   sync.Once

	isStopped    bool
	waitAll      bool
	waitingCount atomic.Int32
	panicHandler func(any)
}

// New 创建并启动一个工作协程池。
//
// maxWorkers 指定最大并发工作协程数，最小值为 1。
// 若无任务到来，工作协程会在空闲超时后逐渐被回收。
// 可通过 opts 自定义配置，例如 WithIdleTimeout。
func New(maxWorkers int, opts ...Option) *WorkerPool {
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	pool := &WorkerPool{
		maxWorkers:  maxWorkers,
		idleTimeout: DefaultIdleTimeout,
		taskChan:    make(chan func()),
		workerChan:  make(chan func()),
		stopSignal:  make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(pool)
		}
	}

	go pool.dispatch()

	return pool
}

// Size 返回最大并发工作协程数。
func (p *WorkerPool) Size() int {
	return p.maxWorkers
}

// Stop 停止工作协程池，仅等待当前运行的任务完成。
// 等待队列中未运行的任务将被丢弃。调用后不得再次提交任务。
func (p *WorkerPool) Stop() {
	p.stop(false)
}

// StopWait 停止工作协程池，并等待所有已排队的任务执行完成。
// 调用后不得再次提交任务。
func (p *WorkerPool) StopWait() {
	p.stop(true)
}

// Stopped 返回工作协程池是否已停止。
func (p *WorkerPool) Stopped() bool {
	p.stopMutex.Lock()
	defer p.stopMutex.Unlock()
	return p.isStopped
}

// Submit 将任务提交到协程池中执行。
//
// 任务将被立即分配给可用的工作协程，若所有协程都在执行任务，
// 则新任务将加入等待队列。Submit 会阻塞直到调度器接收任务。
// task 为 nil 时将被忽略。协程池停止后调用将触发 panic。
func (p *WorkerPool) Submit(task func()) {
	if task == nil {
		return
	}
	p.submit(task, true)
}

// SubmitWait 将任务提交到协程池并阻塞等待其执行完成。
// task 为 nil 时立即返回。
func (p *WorkerPool) SubmitWait(task func()) {
	if task == nil {
		return
	}
	doneChan := make(chan struct{})
	p.submit(func() {
		defer close(doneChan)
		task()
	}, true)
	select {
	case <-doneChan:
	case <-p.stoppedChan:
		// Stop 会丢弃等待队列。任务若已开始，Stop 会等待它结束，因而
		// doneChan 总会先关闭；走到这里表示任务未执行。
	}
}

// submit 在任务通道和停止信号之间仲裁。任务通道永不关闭，从而避免 Submit
// 与 Stop 并发时发生 send/close panic；公开提交在停止后仍按既有契约 panic。
func (p *WorkerPool) submit(task func(), panicIfStopped bool) bool {
	select {
	case p.taskChan <- task:
		return true
	case <-p.stopSignal:
		if panicIfStopped {
			panic("workerpool: submit on stopped pool")
		}
		return false
	}
}

func (p *WorkerPool) syncWaitingCount() {
	p.waitingCount.Store(int32(p.waitingQueue.Size()))
}

// WaitingQueueSize 返回等待队列中的任务数量。
func (p *WorkerPool) WaitingQueueSize() int {
	return int(p.waitingCount.Load())
}

// Pause 暂停协程池中所有工作协程的任务执行。
//
// 调用后将阻塞直到所有工作协程进入暂停状态。暂停期间提交的新任务
// 将被放入等待队列，待 ctx 取消或超时后恢复执行。
// 若协程池已处于暂停状态，本次调用将等待前一次暂停结束后再执行。
func (p *WorkerPool) Pause(ctx context.Context) {
	p.pauseMutex.Lock()
	defer p.pauseMutex.Unlock()

	p.stopMutex.Lock()
	if p.isStopped {
		p.stopMutex.Unlock()
		return
	}
	p.stopMutex.Unlock()

	// 提交占位任务以阻塞所有 worker。使用带缓冲的通知通道，使停止和
	// context 取消都能中断等待，即使部分占位任务仍在等待队列中。
	ready := make(chan struct{}, p.maxWorkers)
	done := make(chan struct{}, p.maxWorkers)
	submitted := 0
	for ; submitted < p.maxWorkers; submitted++ {
		if !p.submit(func() {
			ready <- struct{}{}
			defer func() { done <- struct{}{} }()
			select {
			case <-ctx.Done():
			case <-p.stopSignal:
			}
		}, false) {
			return
		}
	}

	for range submitted {
		select {
		case <-ready:
		case <-ctx.Done():
			return
		case <-p.stopSignal:
			return
		}
	}
	select {
	case <-ctx.Done():
	case <-p.stopSignal:
		return
	}
	for range submitted {
		<-done
	}
}

// dispatch 是任务分发器的主循环，运行在独立的 goroutine 中。
// 负责将任务分配给可用的工作协程，并管理工作协程的生命周期。
func (p *WorkerPool) dispatch() {
	defer close(p.stoppedChan)
	timeout := time.NewTimer(p.idleTimeout)
	defer timeout.Stop()

	var (
		workerCount int
		idle        bool
		wg          sync.WaitGroup
	)

dispatchLoop:
	for {
		if p.waitingQueue.Size() > 0 {
			if !p.processWaitingQueue() {
				break dispatchLoop
			}
			continue
		}

		select {
		case task := <-p.taskChan:
			p.handleTask(task, &workerCount, &wg)
			idle = false
			// 收到新任务后重置空闲计时器，确保超时时间一致
			if !timeout.Stop() {
				select {
				case <-timeout.C:
				default:
				}
			}
			timeout.Reset(p.idleTimeout)
		case <-timeout.C:
			if idle && workerCount > 0 {
				if p.killIdleWorker() {
					workerCount--
				}
			}
			idle = true
			timeout.Reset(p.idleTimeout)
		case <-p.stopSignal:
			break dispatchLoop
		}
	}

	if p.waitAll {
		p.runQueuedTasks()
	} else {
		// Stop 明确丢弃尚未运行的任务，同时释放闭包捕获的对象，并让
		// WaitingQueueSize 在 Stop 返回前反映最终状态。
		p.waitingQueue.Clear()
		p.syncWaitingCount()
	}

	// 停止所有剩余工作协程
	for workerCount > 0 {
		p.workerChan <- nil
		workerCount--
	}
	wg.Wait()
}

// handleTask 将任务分配给可用的工作协程，或创建新协程，或加入等待队列。
func (p *WorkerPool) handleTask(task func(), workerCount *int, wg *sync.WaitGroup) {
	select {
	case p.workerChan <- task:
	default:
		if *workerCount < p.maxWorkers {
			wg.Add(1)
			go worker(task, p.workerChan, wg, p.panicHandler)
			*workerCount++
		} else {
			p.waitingQueue.PushBack(task)
			p.syncWaitingCount()
		}
	}
}

// worker 是工作协程的执行函数。
// 持续从 workerChan 接收并执行任务，收到 nil 时退出。
// 任务中的 panic 会被恢复，不会导致整个程序崩溃。
func worker(task func(), workerChan chan func(), wg *sync.WaitGroup, panicHandler func(any)) {
	defer wg.Done()
	for task != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if panicHandler != nil {
						panicHandler(r)
					}
				}
			}()
			task()
		}()
		task = <-workerChan
	}
}

// stop 执行协程池的停止操作。wait 为 true 时等待所有排队任务完成。
func (p *WorkerPool) stop(wait bool) {
	p.stopOnce.Do(func() {
		p.stopMutex.Lock()
		p.isStopped = true
		p.waitAll = wait
		close(p.stopSignal)
		p.stopMutex.Unlock()
	})
	<-p.stoppedChan
}

// processWaitingQueue 处理等待队列：接收新任务或将队首任务分派给工作协程。
// 返回 false 表示协程池已收到停止信号。
func (p *WorkerPool) processWaitingQueue() bool {
	select {
	case task := <-p.taskChan:
		p.waitingQueue.PushBack(task)
	case p.workerChan <- p.waitingQueue.Front():
		p.waitingQueue.PopFront()
	case <-p.stopSignal:
		return false
	}
	p.syncWaitingCount()
	return true
}

// killIdleWorker 向工作协程通道发送 nil 以回收一个空闲协程。
func (p *WorkerPool) killIdleWorker() bool {
	select {
	case p.workerChan <- nil:
		return true
	default:
		return false
	}
}

// runQueuedTasks 将等待队列中的所有任务依次分派给工作协程执行。
func (p *WorkerPool) runQueuedTasks() {
	for p.waitingQueue.Size() > 0 {
		p.workerChan <- p.waitingQueue.PopFront()
		p.syncWaitingCount()
	}
}
