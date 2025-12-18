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

// CmdOptions 定义进程的配置选项
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

// Process 封装了一个外部进程的执行和管理
type Process struct {
	cmdOptions CmdOptions          // 进程配置
	pExec      *exec.Cmd           // 底层命令实例
	cancelFunc context.CancelFunc  // 用于取消进程的上下文函数
	isRunning  bool                // 进程是否正在运行
	err        error               // 最近的错误
	stdout     func(*bufio.Reader) // 处理标准输出的函数
	stderr     func(*bufio.Reader) // 处理标准错误的函数
	mu         sync.Mutex          // 保护进程状态的锁
	wg         sync.WaitGroup      // 等待输出处理协程完成
}

// NewProcess 创建一个新的 Process 实例
func NewProcess(co CmdOptions) *Process {
	p := &Process{
		cmdOptions: co,
		isRunning:  false,
		err:        nil,
	}

	// 初始化标准输出处理
	if co.OnStdout == nil {
		p.stdout = func(*bufio.Reader) {}
	} else {
		p.stdout = func(reader *bufio.Reader) {
			p.wg.Add(1)
			defer p.wg.Done()
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						p.setError(fmt.Errorf("stdout read error: %v", err))
					}
					return
				}
				co.OnStdout(strings.TrimSuffix(line, "\n"))
			}
		}
	}

	// 初始化标准错误处理
	if co.OnStderr == nil {
		p.stderr = func(*bufio.Reader) {}
	} else {
		p.stderr = func(reader *bufio.Reader) {
			p.wg.Add(1)
			defer p.wg.Done()
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						p.setError(fmt.Errorf("stderr read error: %v", err))
					}
					return
				}
				co.OnStderr(strings.TrimSuffix(line, "\n"))
			}
		}
	}

	return p
}

// Run 同步运行进程，阻塞直到进程结束
func (p *Process) Run() *Process {
	p.mu.Lock()
	if p.isRunning {
		p.setError(fmt.Errorf("process is already running"))
		p.mu.Unlock()
		return p
	}
	p.isRunning = true
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	p.execCommand(ctx)
	return p
}

// AsyncRun 异步运行进程，立即返回
func (p *Process) AsyncRun() *Process {
	p.mu.Lock()
	if p.isRunning {
		p.setError(fmt.Errorf("process is already running"))
		p.mu.Unlock()
		return p
	}
	p.isRunning = true
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	go p.execCommand(ctx)
	return p
}

// execCommand 执行命令的核心逻辑
func (p *Process) execCommand(ctx context.Context) {
	defer func() {
		p.mu.Lock()
		p.isRunning = false
		p.mu.Unlock()
		p.wg.Wait() // 等待输出处理协程完成
		if p.cmdOptions.OnRunAfter != nil {
			p.cmdOptions.OnRunAfter(p)
		}
	}()

	if p.cmdOptions.ExecPath == "" {
		p.setError(fmt.Errorf("exec path is empty"))
		return
	}

	p.pExec = exec.CommandContext(ctx, p.cmdOptions.ExecPath, p.cmdOptions.Args...)
	p.pExec.SysProcAttr = p.cmdOptions.SysProcAttr

	stdout, err := p.pExec.StdoutPipe()
	if err != nil {
		p.setError(fmt.Errorf("failed to get stdout pipe: %v", err))
		return
	}
	defer stdout.Close()

	stderr, err := p.pExec.StderrPipe()
	if err != nil {
		p.setError(fmt.Errorf("failed to get stderr pipe: %v", err))
		return
	}
	defer stderr.Close()

	go p.stdout(bufio.NewReader(stdout))
	go p.stderr(bufio.NewReader(stderr))

	if p.cmdOptions.OnRunBefore != nil {
		p.cmdOptions.OnRunBefore(p)
	}

	if err := p.pExec.Start(); err != nil {
		p.setError(fmt.Errorf("failed to start process: %v", err))
		return
	}

	if err := p.pExec.Wait(); err != nil {
		p.setError(fmt.Errorf("process wait error: %v", err))
	}
}

// Start 启动进程（异步方式）
func (p *Process) Start() *Process {
	if p.cmdOptions.ExecPath == "" {
		p.setError(fmt.Errorf("exec path is empty, cannot start process"))
		return p
	}
	return p.AsyncRun()
}

// Stop 停止正在运行的进程
func (p *Process) Stop() *Process {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		p.setError(fmt.Errorf("process is not running"))
		return p
	}

	cancelFunc := p.cancelFunc
	p.mu.Unlock()

	if cancelFunc != nil {
		cancelFunc()
	}

	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			p.mu.Lock()
			if p.pExec != nil && p.pExec.Process != nil {
				err := p.pExec.Process.Kill()
				p.mu.Unlock()
				if err != nil {
					p.setError(err)
				}
			} else {
				p.mu.Unlock()
			}
			return p
		case <-ticker.C:
			p.mu.Lock()
			if p.pExec != nil && p.pExec.ProcessState != nil {
				p.mu.Unlock()
				return p
			}
			p.mu.Unlock()
		}
	}
}

// Restart 重启进程
func (p *Process) Restart() *Process {
	p.Stop()
	return p.Start()
}

// Wait 等待进程完成，返回错误
func (p *Process) Wait() error {
	p.wg.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// State 返回进程状态
func (p *Process) State() *os.ProcessState {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pExec != nil && p.pExec.ProcessState != nil {
		return p.pExec.ProcessState
	}
	return nil
}

// Pid 返回进程 ID，若进程未启动则返回 -1
func (p *Process) Pid() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pExec != nil && p.pExec.Process != nil {
		return p.pExec.Process.Pid
	}
	return -1
}

// CmdOptions 返回进程的配置选项
func (p *Process) CmdOptions() CmdOptions {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmdOptions
}

// IsRunning 检查进程是否正在运行
func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isRunning
}

// Error 返回最近的错误
func (p *Process) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// setError 设置错误并加锁保护
func (p *Process) setError(err error) {
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
}
