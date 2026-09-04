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

// processRun 保存一次启动的全部可变状态。每次启动使用独立实例，避免退出回调
// 重启进程时，旧运行代次关闭或覆盖新代次的状态。
type processRun struct {
	done          chan struct{}
	cancel        context.CancelFunc
	cmd           *exec.Cmd
	processState  *os.ProcessState
	stopRequested bool
	err           error
}

// Process 管理单个外部进程的完整生命周期。
// 所有方法均并发安全，请通过 New 创建实例。
type Process struct {
	opts Options
	run  *processRun
	mu   sync.Mutex

	state      atomic.Int32 // stateIdle / stateRunning / stateStopped
	restarting atomic.Bool
	lastStart  atomic.Int64 // 上次启动的 UnixNano 时间戳
}

// New 创建 Process 实例。
func New(opts Options) *Process {
	p := &Process{opts: cloneOptions(opts)}
	p.state.Store(stateIdle)
	return p
}

// Run 同步运行进程，阻塞至退出，返回运行期间的累计错误。
func (p *Process) Run() error {
	r, err := p.launch(false)
	if err != nil {
		return err
	}
	return p.waitRun(r)
}

// Start 异步启动进程，立即返回。运行错误通过 Wait 或 OnAfter 获取。
// 进程已在运行时返回错误。
func (p *Process) Start() error {
	_, err := p.launch(true)
	return err
}

// SetOptions 在进程未运行时替换配置。运行中调用返回错误。
func (p *Process) SetOptions(opts Options) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state.Load() == stateRunning {
		return errors.New("cannot update options while process is running")
	}
	p.opts = cloneOptions(opts)
	return nil
}

func (p *Process) launch(async bool) (*processRun, error) {
	p.mu.Lock()
	if p.state.Load() == stateRunning {
		p.mu.Unlock()
		return nil, errors.New("process is already running")
	}

	opts := cloneOptions(p.opts)
	parentCtx := opts.Context
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	r := &processRun{done: make(chan struct{}), cancel: cancel}
	p.run = r
	p.state.Store(stateRunning)
	p.lastStart.Store(time.Now().UnixNano())
	p.mu.Unlock()

	if async {
		go p.exec(ctx, opts, r)
	} else {
		p.exec(ctx, opts, r)
	}
	return r, nil
}

func (p *Process) exec(ctx context.Context, opts Options, r *processRun) {
	defer p.finishRun(opts, r)
	if opts.ExecPath == "" {
		p.addRunError(r, errors.New("exec path is empty"))
		return
	}

	cmd := exec.CommandContext(ctx, opts.ExecPath, opts.Args...)
	cmd.Env = opts.Env
	cmd.Dir = opts.Dir
	cmd.Stdin = opts.Stdin
	cmd.SysProcAttr = opts.SysProcAttr
	p.mu.Lock()
	r.cmd = cmd
	p.mu.Unlock()

	var stdout, stderr io.ReadCloser
	if opts.OnStdout != nil {
		out, err := cmd.StdoutPipe()
		if err != nil {
			p.addRunError(r, fmt.Errorf("stdout pipe: %w", err))
			return
		}
		stdout = out
	}
	if opts.OnStderr != nil {
		out, err := cmd.StderrPipe()
		if err != nil {
			p.addRunError(r, fmt.Errorf("stderr pipe: %w", err))
			return
		}
		stderr = out
	}

	if opts.OnBefore != nil {
		opts.OnBefore(p)
	}

	// Start 会写入 cmd.Process；与 Pid、Signal 和 Stop 对该字段的读取使用同一把锁。
	p.mu.Lock()
	err := cmd.Start()
	p.mu.Unlock()
	if err != nil {
		p.addRunError(r, fmt.Errorf("start: %w", err))
		return
	}

	var readers sync.WaitGroup
	if stdout != nil {
		readers.Add(1)
		go func() {
			defer readers.Done()
			p.readLines(r, stdout, opts.OnStdout, "stdout")
		}()
	}
	if stderr != nil {
		readers.Add(1)
		go func() {
			defer readers.Done()
			p.readLines(r, stderr, opts.OnStderr, "stderr")
		}()
	}

	readers.Wait()
	waitErr := cmd.Wait()
	// Wait 会写 ProcessState；只在它返回后发布不可变结果。
	p.mu.Lock()
	r.processState = cmd.ProcessState
	stopRequested := r.stopRequested
	p.mu.Unlock()
	if waitErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			// Stop 是正常的生命周期操作，不作为运行错误报告；父 context
			// 的取消则必须保留，方便调用方使用 errors.Is 判断原因。
			if !stopRequested {
				p.addRunError(r, fmt.Errorf("context: %w", ctxErr))
			}
		} else {
			p.addRunError(r, fmt.Errorf("wait: %w", waitErr))
		}
	}
}

func (p *Process) finishRun(opts Options, r *processRun) {
	r.cancel()
	p.mu.Lock()
	if p.run == r {
		p.state.Store(stateStopped)
	}
	p.mu.Unlock()

	// done 属于本次运行。OnAfter 可以启动新代次，旧代次只关闭自己的 done。
	defer close(r.done)
	if opts.OnAfter != nil {
		opts.OnAfter(p)
	}
}

func (p *Process) readLines(r *processRun, reader io.Reader, handler func(string), source string) {
	br := bufio.NewReader(reader)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
				if line != "" && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
			}
			handler(line)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.addRunError(r, fmt.Errorf("%s: %w", source, err))
			}
			return
		}
	}
}

// Stop 发送取消信号，等待进程退出或默认超时后强制终止。
func (p *Process) Stop() { p.StopWithTimeout(defaultStopTimeout) }

// StopWithTimeout 与 Stop 相同，但支持自定义宽限时间。
func (p *Process) StopWithTimeout(timeout time.Duration) {
	p.mu.Lock()
	if p.state.Load() != stateRunning || p.run == nil {
		p.mu.Unlock()
		return
	}
	r := p.run
	cancel := r.cancel
	r.stopRequested = true
	p.mu.Unlock()

	cancel()
	if waitDone(r.done, timeout) {
		return
	}

	p.mu.Lock()
	cmd := r.cmd
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	p.mu.Unlock()
	if !waitDone(r.done, killWaitTimeout) {
		p.addRunError(r, errors.New("process could not be killed within timeout"))
	}
}

func waitDone(done <-chan struct{}, timeout time.Duration) bool {
	if timeout <= 0 {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-timer.C:
		return false
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

// Wait 阻塞至调用时对应的运行代次退出，并返回该次运行累计的错误。
// 进程从未启动时立即返回错误。
func (p *Process) Wait() error {
	p.mu.Lock()
	r := p.run
	p.mu.Unlock()
	if r == nil {
		return errors.New("process has not been started")
	}
	return p.waitRun(r)
}

func (p *Process) waitRun(r *processRun) error {
	<-r.done
	p.mu.Lock()
	defer p.mu.Unlock()
	return r.err
}

// Done 返回调用时对应的运行代次退出时关闭的只读 channel，便于 select 使用。
// 进程从未启动时返回 nil。
func (p *Process) Done() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.run == nil {
		return nil
	}
	return p.run.done
}

// IsRunning 报告进程当前是否正在运行。
func (p *Process) IsRunning() bool { return p.state.Load() == stateRunning }

// Pid 返回当前或最近一次进程 ID，尚未成功启动时返回 -1。
func (p *Process) Pid() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.run != nil && p.run.cmd != nil && p.run.cmd.Process != nil {
		return p.run.cmd.Process.Pid
	}
	return -1
}

// ExitCode 返回最近一次进程退出码；未退出或异常终止时返回 -1。
func (p *Process) ExitCode() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.run != nil && p.run.processState != nil {
		return p.run.processState.ExitCode()
	}
	return -1
}

// State 返回最近一次进程退出状态，未退出时返回 nil。
func (p *Process) State() *os.ProcessState {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.run == nil {
		return nil
	}
	return p.run.processState
}

// Error 返回当前或最近一次运行累计的所有错误（via errors.Join）。
func (p *Process) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.run == nil {
		return nil
	}
	return p.run.err
}

// Options 返回当前配置的副本。Args 和 Env 也会被复制。
func (p *Process) Options() Options {
	p.mu.Lock()
	defer p.mu.Unlock()
	return cloneOptions(p.opts)
}

// Signal 向运行中的进程发送信号。进程未运行时返回错误。
func (p *Process) Signal(sig os.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state.Load() != stateRunning || p.run == nil || p.run.cmd == nil || p.run.cmd.Process == nil {
		return errors.New("process is not running")
	}
	return p.run.cmd.Process.Signal(sig)
}

// String 返回进程的调试信息。
func (p *Process) String() string {
	var stateStr string
	switch p.state.Load() {
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

func (p *Process) addRunError(r *processRun, err error) {
	p.mu.Lock()
	r.err = errors.Join(r.err, err)
	p.mu.Unlock()
}

func cloneOptions(opts Options) Options {
	opts.Args = cloneStrings(opts.Args)
	opts.Env = cloneStrings(opts.Env)
	return opts
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append(make([]string, 0, len(values)), values...)
}
