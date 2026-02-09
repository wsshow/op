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
		opt(pool)
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
// 则新任务将加入等待队列。Submit 不会阻塞调用方。
// task 为 nil 时将被忽略。协程池停止后调用将触发 panic。
func (p *WorkerPool) Submit(task func()) {
	if task != nil {
		p.taskChan <- task
	}
}

// SubmitWait 将任务提交到协程池并阻塞等待其执行完成。
// task 为 nil 时立即返回。
func (p *WorkerPool) SubmitWait(task func()) {
	if task == nil {
		return
	}
	doneChan := make(chan struct{})
	p.taskChan <- func() {
		defer close(doneChan)
		task()
	}
	<-doneChan
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

	// 提交占位任务以阻塞所有 worker
	readyWG := new(sync.WaitGroup)
	doneWG := new(sync.WaitGroup)
	readyWG.Add(p.maxWorkers)
	doneWG.Add(p.maxWorkers)

	for i := 0; i < p.maxWorkers; i++ {
		p.Submit(func() {
			readyWG.Done()
			defer doneWG.Done()
			select {
			case <-ctx.Done():
			case <-p.stopSignal:
			}
		})
	}

	readyWG.Wait() // 等待所有暂停任务开始执行
	<-ctx.Done()   // 等待 context 取消
	doneWG.Wait()  // 等待所有暂停任务完成
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
		case task, ok := <-p.taskChan:
			if !ok {
				break dispatchLoop
			}
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
		}
	}

	if p.waitAll {
		p.runQueuedTasks()
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
			go worker(task, p.workerChan, wg)
			*workerCount++
		} else {
			p.waitingQueue.PushBack(task)
			p.waitingCount.Store(int32(p.waitingQueue.Size()))
		}
	}
}

// worker 是工作协程的执行函数。
// 持续从 workerChan 接收并执行任务，收到 nil 时退出。
func worker(task func(), workerChan chan func(), wg *sync.WaitGroup) {
	defer wg.Done()
	for task != nil {
		task()
		task = <-workerChan
	}
}

// stop 执行协程池的停止操作。wait 为 true 时等待所有排队任务完成。
func (p *WorkerPool) stop(wait bool) {
	p.stopOnce.Do(func() {
		close(p.stopSignal)
		p.stopMutex.Lock()
		p.isStopped = true
		p.waitAll = wait
		p.stopMutex.Unlock()
		close(p.taskChan)
	})
	<-p.stoppedChan
}

// processWaitingQueue 处理等待队列：接收新任务或将队首任务分派给工作协程。
// 返回 false 表示任务通道已关闭，协程池应停止。
func (p *WorkerPool) processWaitingQueue() bool {
	select {
	case task, ok := <-p.taskChan:
		if !ok {
			return false
		}
		p.waitingQueue.PushBack(task)
	case p.workerChan <- p.waitingQueue.Front():
		p.waitingQueue.PopFront()
	}
	p.waitingCount.Store(int32(p.waitingQueue.Size()))
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
		p.waitingCount.Store(int32(p.waitingQueue.Size()))
	}
}
