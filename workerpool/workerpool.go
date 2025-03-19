package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wsshow/op/deque"
)

const (
	// 如果工作协程空闲超过此时间，则停止一个工作协程
	idleTimeout = 2 * time.Second
)

// WorkerPool 是一个工作协程池，限制并发执行任务的协程数量不超过指定最大值
type WorkerPool struct {
	maxWorkers   int                 // 最大工作协程数
	taskChan     chan func()         // 任务通道
	workerChan   chan func()         // 工作协程通道
	stopSignal   chan struct{}       // 停止信号通道
	stoppedChan  chan struct{}       // 停止完成通道
	waitingQueue deque.Deque[func()] // 等待任务队列
	stopMutex    sync.Mutex          // 停止操作互斥锁
	stopOnce     sync.Once           // 确保停止只执行一次
	isStopped    bool                // 是否已停止
	waitingCount int32               // 等待队列中的任务数
	waitAll      bool                // 是否等待所有任务完成
}

// New 创建并启动一个工作协程池
//
// maxWorkers 指定最大并发工作协程数。若无任务到来，工作协程会逐渐停止直到没有剩余工作协程
func New(maxWorkers int) *WorkerPool {
	if maxWorkers < 1 {
		maxWorkers = 1 // 确保至少有一个工作协程
	}

	pool := &WorkerPool{
		maxWorkers:  maxWorkers,
		taskChan:    make(chan func()),
		workerChan:  make(chan func()),
		stopSignal:  make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}

	// 启动任务分发器
	go pool.dispatch()

	return pool
}

// Size 返回最大并发工作协程数
func (p *WorkerPool) Size() int {
	return p.maxWorkers
}

// Stop 停止工作协程池，仅等待当前运行任务完成，未运行的待处理任务将被放弃
// 调用后不得再次提交任务
func (p *WorkerPool) Stop() {
	p.stop(false)
}

// StopWait 停止工作协程池，并等待所有排队任务完成
// 调用后不得再次提交任务，所有待处理任务将在函数返回前执行完毕
func (p *WorkerPool) StopWait() {
	p.stop(true)
}

// Stopped 返回工作协程池是否已停止
func (p *WorkerPool) Stopped() bool {
	p.stopMutex.Lock()
	defer p.stopMutex.Unlock()
	return p.isStopped
}

// Submit 将任务加入队列，由工作协程执行
//
// 任务函数需通过闭包捕获外部值，返回值应通过闭包中的通道返回。
// Submit 不会阻塞，无论提交多少任务，新任务会立即分配给可用工作协程或加入等待队列。
func (p *WorkerPool) Submit(task func()) {
	if task != nil {
		p.taskChan <- task
	}
}

// SubmitWait 将任务加入队列并等待其执行完成
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

// WaitingQueueSize 返回等待队列中的任务数
func (p *WorkerPool) WaitingQueueSize() int {
	return int(atomic.LoadInt32(&p.waitingCount))
}

// Pause 使所有工作协程根据给定的 Context 暂停，暂停期间不执行任务
// 返回时所有工作协程已暂停。任务可继续排队，但需等待 Context 取消或超时后执行。
// 若协程池已暂停，则等待前次暂停取消后重新控制暂停状态。
func (p *WorkerPool) Pause(ctx context.Context) {
	p.stopMutex.Lock()
	defer p.stopMutex.Unlock()
	if p.isStopped {
		return
	}

	readyWG := new(sync.WaitGroup)
	readyWG.Add(p.maxWorkers)
	for i := 0; i < p.maxWorkers; i++ {
		p.Submit(func() {
			readyWG.Done()
			select {
			case <-ctx.Done():
			case <-p.stopSignal:
			}
		})
	}
	readyWG.Wait() // 等待所有工作协程暂停
}

// dispatch 分发任务给可用工作协程
func (p *WorkerPool) dispatch() {
	defer close(p.stoppedChan)
	timeout := time.NewTimer(idleTimeout)
	workerCount := 0
	idle := false
	var wg sync.WaitGroup

	for {
		// 处理等待队列中的任务
		if p.waitingQueue.Size() > 0 {
			if !p.processWaitingQueue() {
				break
			}
			continue
		}

		select {
		case task, ok := <-p.taskChan:
			if !ok {
				break
			}
			p.handleTask(task, &workerCount, &wg)
			idle = false
		case <-timeout.C:
			if idle && workerCount > 0 {
				if p.killIdleWorker() {
					workerCount--
				}
			}
			idle = true
			timeout.Reset(idleTimeout)
		}
	}

	// 如果需要等待，则运行所有排队任务
	if p.waitAll {
		p.runQueuedTasks()
	}

	// 停止所有剩余工作协程
	p.shutdownWorkers(workerCount, &wg)
	timeout.Stop()
}

// handleTask 处理单个任务，分配给工作协程或加入等待队列
func (p *WorkerPool) handleTask(task func(), workerCount *int, wg *sync.WaitGroup) {
	select {
	case p.workerChan <- task:
		// 任务直接分配给可用工作协程
	default:
		if *workerCount < p.maxWorkers {
			// 创建新工作协程
			wg.Add(1)
			go worker(task, p.workerChan, wg)
			*workerCount++
		} else {
			// 加入等待队列
			p.waitingQueue.PushBack(task)
			atomic.StoreInt32(&p.waitingCount, int32(p.waitingQueue.Size()))
		}
	}
}

// worker 执行任务，直到收到 nil 任务时停止
func worker(task func(), workerChan chan func(), wg *sync.WaitGroup) {
	for task != nil {
		task()
		task = <-workerChan
	}
	wg.Done()
}

// stop 停止协程池，wait 参数决定是否完成排队任务
func (p *WorkerPool) stop(wait bool) {
	p.stopOnce.Do(func() {
		close(p.stopSignal) // 发送停止信号以解除暂停
		p.stopMutex.Lock()
		p.isStopped = true
		p.waitAll = wait
		p.stopMutex.Unlock()
		close(p.taskChan) // 关闭任务通道
	})
	<-p.stoppedChan // 等待停止完成
}

// processWaitingQueue 处理等待队列中的任务，返回 false 表示协程池已停止
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
	atomic.StoreInt32(&p.waitingCount, int32(p.waitingQueue.Size()))
	return true
}

// killIdleWorker 杀死一个空闲工作协程，返回是否成功
func (p *WorkerPool) killIdleWorker() bool {
	select {
	case p.workerChan <- nil:
		return true
	default:
		return false
	}
}

// runQueuedTasks 执行所有等待队列中的任务
func (p *WorkerPool) runQueuedTasks() {
	for p.waitingQueue.Size() > 0 {
		p.workerChan <- p.waitingQueue.PopFront()
		atomic.StoreInt32(&p.waitingCount, int32(p.waitingQueue.Size()))
	}
}

// shutdownWorkers 停止所有剩余工作协程
func (p *WorkerPool) shutdownWorkers(workerCount int, wg *sync.WaitGroup) {
	for workerCount > 0 {
		p.workerChan <- nil
		workerCount--
	}
	wg.Wait()
}
