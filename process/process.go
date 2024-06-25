package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type CmdOptions struct {
	Name        string
	ExecPath    string
	Args        []string
	OnRunBefore func(*Process)
	OnRunAfter  func(*Process)
	OnStdout    func(string)
	OnStderr    func(string)
}

type Process struct {
	cmdOptions CmdOptions
	pExec      *exec.Cmd
	cancelFunc context.CancelFunc
	isRunning  bool
	err        error
	stdout     func(*bufio.Reader)
	stderr     func(*bufio.Reader)
}

func NewProcess(co CmdOptions) *Process {
	p := &Process{
		cmdOptions: co,
		isRunning:  false,
		cancelFunc: nil,
		err:        nil,
		stdout:     nil,
		stderr:     nil,
	}

	if p.cmdOptions.OnStdout == nil {
		p.stdout = func(reader *bufio.Reader) {}
	} else {
		p.stdout = func(reader *bufio.Reader) {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
				line = strings.TrimSuffix(line, "\n")
				p.cmdOptions.OnStdout(line)
			}
		}
	}

	if p.cmdOptions.OnStderr == nil {
		p.stderr = func(reader *bufio.Reader) {}
	} else {
		p.stderr = func(reader *bufio.Reader) {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
				line = strings.TrimSuffix(line, "\n")
				p.cmdOptions.OnStderr(line)
			}
		}
	}

	return p
}

func (p *Process) AsyncRun() *Process {
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.cancelFunc = cancelFunc

	exit := make(chan struct{})
	go func() {

		var (
			err    error
			stdout io.ReadCloser
			stderr io.ReadCloser
		)

		defer func() {
			p.isRunning = false
			if p.cmdOptions.OnRunAfter != nil {
				p.cmdOptions.OnRunAfter(p)
			}
			if err != nil {
				exit <- struct{}{}
				p.err = err
			}
		}()

		p.pExec = exec.CommandContext(ctx, p.cmdOptions.ExecPath, p.cmdOptions.Args...)

		if stdout, err = p.pExec.StdoutPipe(); err != nil {
			return
		} else {
			go p.stdout(bufio.NewReader(stdout))
		}

		if stderr, err = p.pExec.StderrPipe(); err != nil {
			return
		} else {
			go p.stderr(bufio.NewReader(stderr))
		}

		if err = p.pExec.Start(); err != nil {
			return
		}

		p.isRunning = true

		exit <- struct{}{}

		if err = p.pExec.Wait(); err != nil {
			return
		}

	}()

	<-exit

	if p.cmdOptions.OnRunBefore != nil {
		p.cmdOptions.OnRunBefore(p)
	}

	return p
}

func (p *Process) Run() *Process {
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.cancelFunc = cancelFunc

	var (
		err    error
		stdout io.ReadCloser
		stderr io.ReadCloser
	)

	defer func() {
		p.isRunning = false
		if p.cmdOptions.OnRunAfter != nil {
			p.cmdOptions.OnRunAfter(p)
		}
		if err != nil {
			p.err = err
		}
	}()

	p.pExec = exec.CommandContext(ctx, p.cmdOptions.ExecPath, p.cmdOptions.Args...)

	if stdout, err = p.pExec.StdoutPipe(); err != nil {
		return p
	} else {
		go p.stdout(bufio.NewReader(stdout))
	}

	if stderr, err = p.pExec.StderrPipe(); err != nil {
		return p
	} else {
		go p.stderr(bufio.NewReader(stderr))
	}

	if p.cmdOptions.OnRunBefore != nil {
		p.cmdOptions.OnRunBefore(p)
	}

	if err = p.pExec.Start(); err != nil {
		return p
	}

	p.isRunning = true

	if err = p.pExec.Wait(); err != nil {
		return p
	}

	return p
}

func (p *Process) Start() {
	if p.isRunning {
		p.err = fmt.Errorf("process was running")
		return
	}
	if len(p.cmdOptions.ExecPath) == 0 {
		p.err = fmt.Errorf("not found execpath, process couldn't to start")
		return
	}
	p.AsyncRun()
}

func (p *Process) Stop() {
	if !p.isRunning {
		p.err = fmt.Errorf("process has not started")
		return
	}
	defer func() { p.isRunning = false }()
	p.cancelFunc()
	for p.pExec.ProcessState == nil {
		select {
		case <-time.After(3 * time.Second):
			p.err = p.pExec.Process.Kill()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (p *Process) ReStart() {
	if len(p.cmdOptions.ExecPath) == 0 {
		p.err = fmt.Errorf("not found execpath, process couldn't to restart")
		return
	}
	p.Stop()
	p.Start()
}

func (p *Process) Pid() int {
	return p.pExec.Process.Pid
}

func (p *Process) CmdOptions() CmdOptions {
	return p.cmdOptions
}

func (p *Process) IsRunning() bool {
	return p.isRunning
}

func (p *Process) Error() error {
	return p.err
}
