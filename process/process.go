package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// stopTimeout 定义停止进程时的最大等待时间。
const stopTimeout = 5 * time.Second

// CmdOptions 定义进程的配置选项。
type CmdOptions struct {
	Name        string               // 进程名称，用于标识
	ExecPath    string               // 可执行文件的路径
	Args        []string             // 命令行参数
	OnRunBefore func(*Process)       // 进程启动前的回调
	OnRunAfter  func(*Process)       // 进程结束后的回调
	OnStdout    func(string)         // 标准输出行回调
	OnStderr    func(string)         // 标准错误行回调
	SysProcAttr *syscall.SysProcAttr // 系统进程属性，用于控制进程行为
}

// Process 封装了一个外部进程的执行和生命周期管理。
type Process struct {
	cmdOptions CmdOptions         // 进程配置
	pExec      *exec.Cmd          // 底层命令实例
	cancelFunc context.CancelFunc // 用于取消进程的上下文函数
	isRunning  bool               // 进程是否正在运行
	err        error              // 最近的错误
	done       chan struct{}      // 进程执行完毕时关闭
	mu         sync.Mutex         // 保护进程状态的锁
	wg         sync.WaitGroup     // 等待 stdout/stderr 读取协程完成
}

// NewProcess 创建一个新的 Process 实例。
func NewProcess(co CmdOptions) *Process {
	return &Process{
		cmdOptions: co,
	}
}

// Run 同步运行进程，阻塞直到进程结束。
func (p *Process) Run() *Process {
	p.mu.Lock()
	if p.isRunning {
		p.err = fmt.Errorf("process is already running")
		p.mu.Unlock()
		return p
	}
	p.isRunning = true
	p.err = nil
	p.done = make(chan struct{})
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	p.execCommand(ctx)
	return p
}

// AsyncRun 异步运行进程，立即返回。
func (p *Process) AsyncRun() *Process {
	p.mu.Lock()
	if p.isRunning {
		p.err = fmt.Errorf("process is already running")
		p.mu.Unlock()
		return p
	}
	p.isRunning = true
	p.err = nil
	p.done = make(chan struct{})
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	go p.execCommand(ctx)
	return p
}

// execCommand 执行命令的核心逻辑。
func (p *Process) execCommand(ctx context.Context) {
	defer func() {
		p.mu.Lock()
		p.isRunning = false
		p.mu.Unlock()
		if p.cmdOptions.OnRunAfter != nil {
			p.cmdOptions.OnRunAfter(p)
		}
		close(p.done)
	}()

	if p.cmdOptions.ExecPath == "" {
		p.setError(fmt.Errorf("exec path is empty"))
		return
	}

	p.mu.Lock()
	p.pExec = exec.CommandContext(ctx, p.cmdOptions.ExecPath, p.cmdOptions.Args...)
	p.pExec.SysProcAttr = p.cmdOptions.SysProcAttr
	p.mu.Unlock()

	stdout, err := p.pExec.StdoutPipe()
	if err != nil {
		p.setError(fmt.Errorf("failed to get stdout pipe: %w", err))
		return
	}

	stderr, err := p.pExec.StderrPipe()
	if err != nil {
		p.setError(fmt.Errorf("failed to get stderr pipe: %w", err))
		return
	}

	if p.cmdOptions.OnRunBefore != nil {
		p.cmdOptions.OnRunBefore(p)
	}

	if err := p.pExec.Start(); err != nil {
		p.setError(fmt.Errorf("failed to start process: %w", err))
		return
	}

	// 在 Start 成功后启动输出读取协程。
	// wg.Add 必须在协程外调用以避免竞态条件。
	if p.cmdOptions.OnStdout != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.readOutput(bufio.NewReader(stdout), p.cmdOptions.OnStdout, "stdout")
		}()
	}
	if p.cmdOptions.OnStderr != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.readOutput(bufio.NewReader(stderr), p.cmdOptions.OnStderr, "stderr")
		}()
	}

	// 先等待读取完成，再调用 Wait 收集退出状态（符合 exec.Cmd 文档要求）。
	p.wg.Wait()

	if err := p.pExec.Wait(); err != nil {
		// 上下文取消（Stop 调用）导致的退出错误属于预期行为，不记录
		if ctx.Err() == nil {
			p.setError(fmt.Errorf("process exited with error: %w", err))
		}
	}
}

// readOutput 逐行读取输出并调用 handler 处理。
// ReadString 在遇到 EOF 时可能同时返回数据和错误，需先处理数据再检查错误。
func (p *Process) readOutput(reader *bufio.Reader, handler func(string), source string) {
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			handler(strings.TrimSuffix(line, "\n"))
		}
		if err != nil {
			if err != io.EOF {
				p.setError(fmt.Errorf("%s read error: %w", source, err))
			}
			return
		}
	}
}

// Start 启动进程（异步方式），等同于 AsyncRun。
func (p *Process) Start() *Process {
	return p.AsyncRun()
}

// Stop 停止正在运行的进程。
// 先通过上下文取消发送终止信号，超时后强制终止。
func (p *Process) Stop() *Process {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return p
	}
	cancelFunc := p.cancelFunc
	done := p.done
	p.mu.Unlock()

	// 取消上下文，CommandContext 会向进程发送终止信号
	if cancelFunc != nil {
		cancelFunc()
	}

	// 等待进程退出，超时则强制终止
	select {
	case <-done:
	case <-time.After(stopTimeout):
		p.mu.Lock()
		proc := p.pExec
		p.mu.Unlock()
		if proc != nil && proc.Process != nil {
			_ = proc.Process.Kill()
		}
		<-done
	}
	return p
}

// Restart 重启进程。
func (p *Process) Restart() *Process {
	p.Stop()
	return p.Start()
}

// Wait 等待进程执行完毕并返回错误。
func (p *Process) Wait() error {
	p.mu.Lock()
	done := p.done
	p.mu.Unlock()

	if done != nil {
		<-done
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// State 返回进程退出状态，若进程未结束则返回 nil。
func (p *Process) State() *os.ProcessState {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pExec != nil {
		return p.pExec.ProcessState
	}
	return nil
}

// Pid 返回进程 ID，若进程未启动则返回 -1。
func (p *Process) Pid() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pExec != nil && p.pExec.Process != nil {
		return p.pExec.Process.Pid
	}
	return -1
}

// CmdOptions 返回进程的配置选项。
func (p *Process) CmdOptions() CmdOptions {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmdOptions
}

// IsRunning 检查进程是否正在运行。
func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isRunning
}

// Error 返回最近的错误。
func (p *Process) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// setError 线程安全地设置错误。
func (p *Process) setError(err error) {
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
}
