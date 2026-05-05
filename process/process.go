// Package process 提供外部进程的生命周期管理，支持启动、停止、重启、
// 标准输出/错误的逐行回调以及多进程管理。
//
// Process 和 Manager 均并发安全。
package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	defaultStopTimeout = 5 * time.Second
	killWaitTimeout    = 2 * time.Second
)

const (
	stateIdle int32 = iota
	stateRunning
	stateStopped
)

// Options 进程启动配置。
type Options struct {
	ExecPath    string               // 可执行文件路径
	Args        []string             // 命令行参数
	Env         []string             // 环境变量，nil 时继承父进程环境
	Dir         string               // 工作目录，空字符串使用当前目录
	Stdin       io.Reader            // 标准输入，nil 时进程从 /dev/null 读取
	Context     context.Context      // 父上下文，nil 时使用 context.Background
	OnBefore    func(*Process)       // 启动前回调
	OnAfter     func(*Process)       // 退出后回调（无论成功与否）
	OnStdout    func(string)         // 标准输出逐行回调
	OnStderr    func(string)         // 标准错误逐行回调
	SysProcAttr *syscall.SysProcAttr // 系统级进程属性

	// MinRestartInterval 是两次启动之间的最小间隔，0 表示无限制。
	// 用于防止 OnAfter 中调用 Restart 导致的紧密重启循环。
	MinRestartInterval time.Duration
}

// Process 管理单个外部进程的完整生命周期。
// 所有方法均并发安全，请通过 New 创建实例。
type Process struct {
	opts       Options
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	done       chan struct{} // 每次启动时重建，退出时关闭
	err        error
	mu         sync.Mutex
	wg         sync.WaitGroup
	state      atomic.Int32 // stateIdle / stateRunning / stateStopped
	restarting atomic.Bool
	lastStart  atomic.Int64 // 上次启动的 UnixNano 时间戳
}

// New 创建 Process 实例。
func New(opts Options) *Process { //nolint:gocritic // Options by value is intentional for API simplicity
	p := &Process{opts: opts}
	p.state.Store(stateIdle)
	return p
}

// Run 同步运行进程，阻塞至退出，返回运行期间的累计错误。
func (p *Process) Run() error {
	if err := p.launch(false); err != nil {
		return err
	}
	return p.Wait()
}

// Start 异步启动进程，立即返回。运行错误通过 Wait 或 OnAfter 获取。
// 进程已在运行时返回错误。
func (p *Process) Start() error {
	return p.launch(true)
}

// SetOptions 在进程未运行时替换配置。运行中调用返回错误。
func (p *Process) SetOptions(opts Options) error { //nolint:gocritic
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state.Load() == stateRunning {
		return errors.New("cannot update options while process is running")
	}
	p.opts = opts
	return nil
}

func (p *Process) launch(async bool) error {
	p.mu.Lock()
	if p.state.Load() == stateRunning {
		p.mu.Unlock()
		return errors.New("process is already running")
	}
	p.err = nil
	p.done = make(chan struct{})
	parentCtx := p.opts.Context
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	p.cancel = cancel
	p.state.Store(stateRunning)
	p.lastStart.Store(time.Now().UnixNano())
	p.mu.Unlock()

	if async {
		go p.exec(ctx)
	} else {
		p.exec(ctx)
	}
	return nil
}

func (p *Process) exec(ctx context.Context) {
	defer func() {
		p.state.Store(stateStopped)
		if p.opts.OnAfter != nil {
			p.opts.OnAfter(p)
		}
		close(p.done)
	}()

	if p.opts.ExecPath == "" {
		p.addError(errors.New("exec path is empty"))
		return
	}

	p.mu.Lock()
	//nolint:gosec // ExecPath and Args come from library consumer configuration
	p.cmd = exec.CommandContext(ctx, p.opts.ExecPath, p.opts.Args...)
	p.cmd.Env = p.opts.Env
	p.cmd.Dir = p.opts.Dir
	p.cmd.Stdin = p.opts.Stdin
	p.cmd.SysProcAttr = p.opts.SysProcAttr
	p.mu.Unlock()

	var stdout, stderr io.ReadCloser

	if p.opts.OnStdout != nil {
		out, err := p.cmd.StdoutPipe()
		if err != nil {
			p.addError(fmt.Errorf("stdout pipe: %w", err))
			return
		}
		stdout = out
	}
	if p.opts.OnStderr != nil {
		out, err := p.cmd.StderrPipe()
		if err != nil {
			p.addError(fmt.Errorf("stderr pipe: %w", err))
			return
		}
		stderr = out
	}

	if p.opts.OnBefore != nil {
		p.opts.OnBefore(p)
	}

	if err := p.cmd.Start(); err != nil {
		p.addError(fmt.Errorf("start: %w", err))
		return
	}

	if stdout != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.readLines(stdout, p.opts.OnStdout, "stdout")
		}()
	}
	if stderr != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.readLines(stderr, p.opts.OnStderr, "stderr")
		}()
	}

	// 必须先等读取协程结束，再调用 cmd.Wait（exec.Cmd 文档要求）。
	p.wg.Wait()

	if err := p.cmd.Wait(); err != nil && ctx.Err() == nil {
		p.addError(fmt.Errorf("wait: %w", err))
	}
}

func (p *Process) readLines(r io.Reader, handler func(string), source string) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		// line 包含分隔符 '\n'，仅在干净 EOF（无数据）时为空字符串。
		if line != "" {
			handler(strings.TrimRight(line, "\r\n"))
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.addError(fmt.Errorf("%s: %w", source, err))
			}
			return
		}
	}
}

// Stop 发送取消信号，等待进程退出或默认超时后强制终止。
func (p *Process) Stop() {
	p.StopWithTimeout(defaultStopTimeout)
}

// StopWithTimeout 与 Stop 相同，但支持自定义宽限时间。
func (p *Process) StopWithTimeout(timeout time.Duration) {
	p.mu.Lock()
	if p.state.Load() != stateRunning {
		p.mu.Unlock()
		return
	}
	cancel := p.cancel
	done := p.done
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	select {
	case <-done:
	case <-time.After(timeout):
		p.mu.Lock()
		cmd := p.cmd
		p.mu.Unlock()
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-done:
		case <-time.After(killWaitTimeout):
			p.addError(errors.New("process could not be killed within timeout"))
		}
	}
}

// Restart 停止后重新启动。并发调用只有一个生效，其余立即返回。
// MinRestartInterval 配置的间隔会在两次启动之间强制等待。
func (p *Process) Restart() error {
	if !p.restarting.CompareAndSwap(false, true) {
		return nil
	}
	defer p.restarting.Store(false)

	p.mu.Lock()
	interval := p.opts.MinRestartInterval
	p.mu.Unlock()

	if interval > 0 {
		lastStart := time.Unix(0, p.lastStart.Load())
		if elapsed := time.Since(lastStart); elapsed < interval {
			time.Sleep(interval - elapsed)
		}
	}

	p.Stop()
	return p.Start()
}

// Wait 阻塞至进程退出，返回运行期间累计的错误。
// 进程从未启动时立即返回错误。
func (p *Process) Wait() error {
	p.mu.Lock()
	done := p.done
	p.mu.Unlock()

	if done == nil {
		return errors.New("process has not been started")
	}
	<-done
	return p.Error()
}

// Done 返回进程退出时关闭的只读 channel，便于 select 使用。
// 进程从未启动时返回 nil。
func (p *Process) Done() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done
}

// IsRunning 报告进程当前是否正在运行。
func (p *Process) IsRunning() bool {
	return p.state.Load() == stateRunning
}

// Pid 返回进程 ID，未启动时返回 -1。
func (p *Process) Pid() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return -1
}

// ExitCode 返回进程退出码；未退出或异常终止时返回 -1。
func (p *Process) ExitCode() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil && p.cmd.ProcessState != nil {
		return p.cmd.ProcessState.ExitCode()
	}
	return -1
}

// State 返回进程退出状态，未退出时返回 nil。
func (p *Process) State() *os.ProcessState {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil {
		return p.cmd.ProcessState
	}
	return nil
}

// Error 返回本次运行累计的所有错误（via errors.Join）。
func (p *Process) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// Options 返回当前配置的副本。
func (p *Process) Options() Options {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.opts
}

// Signal 向运行中的进程发送信号。进程未运行时返回错误。
func (p *Process) Signal(sig os.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd == nil || p.cmd.Process == nil {
		return errors.New("process is not running")
	}
	return p.cmd.Process.Signal(sig)
}

// String 返回进程的调试信息。
func (p *Process) String() string {
	s := p.state.Load()
	var stateStr string
	switch s {
	case stateIdle:
		stateStr = "idle"
	case stateRunning:
		stateStr = "running"
	case stateStopped:
		stateStr = "stopped"
	default:
		stateStr = "unknown"
	}
	return fmt.Sprintf("Process{pid: %d, state: %s}", p.Pid(), stateStr)
}

func (p *Process) addError(err error) {
	p.mu.Lock()
	p.err = errors.Join(p.err, err)
	p.mu.Unlock()
}
