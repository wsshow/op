package process

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- cross-platform helper ----------
// TestHelperProcess 作为子进程 stub：通过 os.Args[0] 发起，
// 传入 "-test.run=TestHelperProcess" 和 "--" 分隔命令参数。
// helper 检测到 "--" 时执行对应行为，否则退化为普通空测试。

func TestHelperProcess(t *testing.T) {
	hasRun := false
	hasSep := false
	for _, a := range os.Args {
		if a == "-test.run=TestHelperProcess" {
			hasRun = true
		}
		if a == "--" {
			hasSep = true
		}
	}
	if !hasRun || !hasSep {
		return
	}

	// 找到 "--" 之后的参数
	var cmdArgs []string
	for i, a := range os.Args {
		if a == "--" {
			cmdArgs = os.Args[i+1:]
			break
		}
	}
	if len(cmdArgs) == 0 {
		os.Exit(0)
	}

	switch cmdArgs[0] {
	case "exit":
		code := 0
		if len(cmdArgs) > 1 {
			fmt.Sscanf(cmdArgs[1], "%d", &code)
		}
		os.Exit(code)
	case "stdout":
		for _, s := range cmdArgs[1:] {
			fmt.Println(s)
		}
	case "stderr":
		for _, s := range cmdArgs[1:] {
			fmt.Fprintln(os.Stderr, s)
		}
	case "mixed":
		for _, s := range cmdArgs[1:] {
			if after, ok := strings.CutPrefix(s, "e:"); ok {
				fmt.Fprintln(os.Stderr, after)
			} else {
				fmt.Println(s)
			}
		}
	case "sleep":
		secs := 10
		if len(cmdArgs) > 1 {
			fmt.Sscanf(cmdArgs[1], "%d", &secs)
		}
		time.Sleep(time.Duration(secs) * time.Second)
	}
	os.Exit(0)
}

// helperOpts 返回使用测试二进制作为子进程的 Options。
func helperOpts(args ...string) Options {
	all := make([]string, 0, 2+len(args))
	all = append(all, "-test.run=TestHelperProcess", "--")
	all = append(all, args...)
	return Options{ExecPath: os.Args[0], Args: all}
}

func mustStart(t *testing.T, p *Process) {
	t.Helper()
	if err := p.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
}

func mustWait(t *testing.T, p *Process) {
	t.Helper()
	if err := p.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

// ========================================================================
// Start / Wait / Run
// ========================================================================

func TestStartWaitSuccess(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	mustStart(t, p)
	mustWait(t, p)
	if ec := p.ExitCode(); ec != 0 {
		t.Fatalf("exit code = %d, want 0", ec)
	}
	if p.IsRunning() {
		t.Fatal("should not be running after exit")
	}
	if s := p.State(); s == nil {
		t.Fatal("State should not be nil after exit")
	}
}

func TestStartWaitNonZero(t *testing.T) {
	p := New(helperOpts("exit", "3"))
	mustStart(t, p)
	err := p.Wait()
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if ec := p.ExitCode(); ec != 3 {
		t.Fatalf("exit code = %d, want 3", ec)
	}
}

func TestRunSuccess(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	if err := p.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if ec := p.ExitCode(); ec != 0 {
		t.Fatalf("exit code = %d, want 0", ec)
	}
}

func TestRunNonZero(t *testing.T) {
	p := New(helperOpts("exit", "2"))
	if err := p.Run(); err == nil {
		t.Fatal("Run should return error for exit 2")
	}
}

func TestDoubleStart(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	if err := p.Start(); err == nil {
		t.Fatal("expected error on double start")
	}
}

func TestWaitNeverStarted(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	if err := p.Wait(); err == nil {
		t.Fatal("expected error when waiting for never-started process")
	}
}

// ---------- Pid / ExitCode / State ----------

func TestPidWhileRunning(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	if pid := p.Pid(); pid != -1 {
		t.Fatalf("Pid before start = %d, want -1", pid)
	}
	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if pid := p.Pid(); pid <= 0 {
		t.Fatalf("Pid while running = %d, want > 0", pid)
	}
}

func TestExitCodeWhileRunning(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if ec := p.ExitCode(); ec != -1 {
		t.Fatalf("ExitCode while running = %d, want -1", ec)
	}
}

func TestStateQueriesConcurrentWithWait(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	mustStart(t, p)
	done := p.Done()

	for {
		select {
		case <-done:
			if state := p.State(); state == nil {
				t.Fatal("State should be published when Done is closed")
			}
			if code := p.ExitCode(); code != 0 {
				t.Fatalf("ExitCode = %d, want 0", code)
			}
			return
		default:
			_ = p.State()
			_ = p.ExitCode()
			_ = p.Pid()
		}
	}
}

func TestStateBeforeStart(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	if s := p.State(); s != nil {
		t.Fatal("State before start should be nil")
	}
	if p.IsRunning() {
		t.Fatal("should not be running before start")
	}
}

func TestIsRunningLifecycle(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	if p.IsRunning() {
		t.Fatal("should not be running before Start")
	}
	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	if !p.IsRunning() {
		t.Fatal("should be running after Start")
	}
	p.Stop()
	if p.IsRunning() {
		t.Fatal("should not be running after Stop")
	}
}

// ========================================================================
// Callbacks
// ========================================================================

func TestOnStdout(t *testing.T) {
	var lines []string
	var mu sync.Mutex
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "stdout", "hello", "world"},
		OnStdout: func(s string) {
			mu.Lock()
			lines = append(lines, s)
			mu.Unlock()
		},
	})
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %v", len(lines), lines)
	}
}

func TestOnStderr(t *testing.T) {
	var lines []string
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "stderr", "err1", "err2"},
		OnStderr: func(s string) { lines = append(lines, s) },
	})
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d stderr lines, want 2: %v", len(lines), lines)
	}
}

func TestOnBefore(t *testing.T) {
	var called bool
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		OnBefore: func(proc *Process) {
			called = true
			if !proc.IsRunning() {
				t.Error("IsRunning should be true in OnBefore")
			}
		},
	})
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("OnBefore was not called")
	}
}

func TestOnAfterSuccess(t *testing.T) {
	var called bool
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		OnAfter: func(proc *Process) {
			called = true
			if proc.IsRunning() {
				t.Error("IsRunning should be false in OnAfter")
			}
			if proc.State() == nil {
				t.Error("State should be non-nil in OnAfter")
			}
		},
	})
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("OnAfter was not called")
	}
}

func TestOnAfterFailure(t *testing.T) {
	var called bool
	p := New(Options{
		ExecPath: "nonexistent_binary_xyz",
		OnAfter: func(proc *Process) {
			called = true
			if proc.Error() == nil {
				t.Error("Error should be non-nil after exec failure")
			}
		},
	})
	p.Start()
	p.Wait()
	if !called {
		t.Fatal("OnAfter was not called on failure")
	}
}

func TestEmptyExecPath(t *testing.T) {
	var afterCalled bool
	p := New(Options{OnAfter: func(_ *Process) { afterCalled = true }})
	p.Start()
	p.Wait()
	if !afterCalled {
		t.Fatal("OnAfter should be called even when ExecPath is empty")
	}
	if err := p.Error(); err == nil {
		t.Fatal("expected error for empty ExecPath")
	}
}

// ========================================================================
// Stop / StopWithTimeout
// ========================================================================

func TestStopRunning(t *testing.T) {
	p := New(helperOpts("sleep", "30"))
	mustStart(t, p)
	p.Stop()
	if p.IsRunning() {
		t.Fatal("process should not be running after Stop")
	}
	// kill 路径下 cmd.ProcessState 可能未被 Wait 填充（平台差异），
	// 仅验证进程已退出即可。
}

func TestStopIdle(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	p.Stop()
	p.StopWithTimeout(10 * time.Millisecond)
}

func TestStopAlreadyExited(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	p.Run()
	p.Stop()
	p.StopWithTimeout(10 * time.Millisecond)
}

func TestStopWithTimeoutKill(t *testing.T) {
	p := New(helperOpts("sleep", "60"))
	mustStart(t, p)
	p.StopWithTimeout(50 * time.Millisecond)
	if p.IsRunning() {
		t.Fatal("process should not be running after forced kill")
	}
}

// ========================================================================
// Restart
// ========================================================================

func TestRestart(t *testing.T) {
	var starts int32
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		OnBefore: func(_ *Process) { atomic.AddInt32(&starts, 1) },
	})
	mustStart(t, p)
	time.Sleep(50 * time.Millisecond)
	if err := p.Restart(); err != nil {
		t.Fatal(err)
	}
	mustWait(t, p)
	if n := atomic.LoadInt32(&starts); n != 2 {
		t.Fatalf("start count = %d, want 2", n)
	}
}

func TestRestartMinInterval(t *testing.T) {
	p := New(Options{
		ExecPath:           os.Args[0],
		Args:               []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		MinRestartInterval: 200 * time.Millisecond,
	})
	mustStart(t, p)
	// 不额外 sleep：lastStart 由 launch 刚刚设置，Restart 应按间隔等待。
	start := time.Now()
	if err := p.Restart(); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 180*time.Millisecond {
		t.Fatalf("Restart took %v, want >= ~200ms (backoff)", elapsed)
	}
	mustWait(t, p)
}

func TestRestartConcurrent(t *testing.T) {
	var starts int32
	p := New(Options{
		ExecPath:           os.Args[0],
		Args:               []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		MinRestartInterval: 500 * time.Millisecond,
		OnBefore:           func(_ *Process) { atomic.AddInt32(&starts, 1) },
	})
	mustStart(t, p)

	var wg sync.WaitGroup
	barrier := make(chan struct{})
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier
			errs <- p.Restart()
		}()
	}
	close(barrier)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Restart() error = %v", err)
		}
	}
	mustWait(t, p)

	if n := atomic.LoadInt32(&starts); n != 2 {
		t.Fatalf("start count = %d, want 2 (only one restart)", n)
	}
}

func TestOnAfterCanRestart(t *testing.T) {
	var exits atomic.Int32
	completed := make(chan struct{})
	startErr := make(chan error, 1)

	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		OnAfter: func(proc *Process) {
			if exits.Add(1) == 1 {
				startErr <- proc.Start()
				return
			}
			close(completed)
		},
	})

	mustStart(t, p)
	select {
	case err := <-startErr:
		if err != nil {
			t.Fatalf("Start from OnAfter: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OnAfter did not restart process")
	}

	select {
	case <-completed:
	case <-time.After(2 * time.Second):
		t.Fatal("restarted process did not complete")
	}
	if got := exits.Load(); got != 2 {
		t.Fatalf("exit callbacks = %d, want 2", got)
	}
}

// ========================================================================
// Signal
// ========================================================================

func TestSignalKill(t *testing.T) {
	p := New(helperOpts("sleep", "30"))
	mustStart(t, p)
	time.Sleep(100 * time.Millisecond)

	if err := p.Signal(os.Kill); err != nil {
		t.Fatalf("Signal: %v", err)
	}
	p.Wait()
	if p.IsRunning() {
		t.Fatal("process should be dead after Kill signal")
	}
}

func TestSignalNotRunning(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	if err := p.Signal(os.Kill); err == nil {
		t.Fatal("expected error when signalling non-running process")
	}
}

// ========================================================================
// Done
// ========================================================================

func TestDone(t *testing.T) {
	p := New(helperOpts("exit", "0"))
	if d := p.Done(); d != nil {
		t.Fatal("Done should be nil before start")
	}
	mustStart(t, p)
	done := p.Done()
	if done == nil {
		t.Fatal("Done should not be nil after start")
	}
	mustWait(t, p)
	select {
	case <-done:
	default:
		t.Fatal("Done channel should be closed after exit")
	}
}

// ========================================================================
// Error
// ========================================================================

func TestErrorAccumulation(t *testing.T) {
	p := New(Options{ExecPath: "nonexistent_binary_xyz"})
	p.Start()
	p.Wait()
	if p.Error() == nil {
		t.Fatal("expected error after failed exec")
	}
}

func TestErrorResetBetweenRuns(t *testing.T) {
	p := New(helperOpts("exit", "2"))
	p.Run()
	if p.Error() == nil {
		t.Fatal("expected error after non-zero exit")
	}
	p.SetOptions(helperOpts("exit", "0"))
	p.Run()
	if err := p.Error(); err != nil {
		t.Fatalf("expected no error after clean run, got: %v", err)
	}
}

// ========================================================================
// Options / String
// ========================================================================

func TestOptions(t *testing.T) {
	opts := helperOpts("exit", "0")
	p := New(opts)
	got := p.Options()
	if got.ExecPath != os.Args[0] {
		t.Fatalf("ExecPath = %q, want %q", got.ExecPath, os.Args[0])
	}
}

func TestOptionsAreCopied(t *testing.T) {
	opts := helperOpts("exit", "0")
	opts.Env = []string{"PROCESS_TEST=value"}
	p := New(opts)

	opts.Args[0] = "mutated"
	opts.Env[0] = "mutated"
	got := p.Options()
	if got.Args[0] == "mutated" || got.Env[0] == "mutated" {
		t.Fatal("New retained caller-owned option slices")
	}

	got.Args[0] = "changed again"
	got.Env[0] = "changed again"
	again := p.Options()
	if again.Args[0] == "changed again" || again.Env[0] == "changed again" {
		t.Fatal("Options exposed internal option slices")
	}
}

func TestOptionsPreserveNilAndEmptySlices(t *testing.T) {
	withNil := New(Options{}).Options()
	if withNil.Args != nil || withNil.Env != nil {
		t.Fatalf("nil slices changed: Args=%v Env=%v", withNil.Args, withNil.Env)
	}

	withEmpty := New(Options{Args: []string{}, Env: []string{}}).Options()
	if withEmpty.Args == nil || withEmpty.Env == nil {
		t.Fatalf("non-nil empty slices changed: Args=%v Env=%v", withEmpty.Args, withEmpty.Env)
	}
}

func TestString(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	s := p.String()
	if !strings.Contains(s, "idle") {
		t.Fatalf("String() = %q, should contain \"idle\"", s)
	}

	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	s = p.String()
	if !strings.Contains(s, "running") {
		t.Fatalf("String() = %q, should contain \"running\"", s)
	}

	p.Stop()
	s = p.String()
	if !strings.Contains(s, "stopped") {
		t.Fatalf("String() = %q, should contain \"stopped\"", s)
	}
}

// ========================================================================
// SetOptions
// ========================================================================

func TestSetOptionsIdle(t *testing.T) {
	p := New(helperOpts("exit", "1"))
	if err := p.SetOptions(helperOpts("exit", "0")); err != nil {
		t.Fatal(err)
	}
	if err := p.Run(); err != nil {
		t.Fatalf("after SetOptions should succeed: %v", err)
	}
}

func TestSetOptionsRunning(t *testing.T) {
	p := New(helperOpts("sleep", "5"))
	mustStart(t, p)
	defer p.StopWithTimeout(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	if err := p.SetOptions(helperOpts("exit", "0")); err == nil {
		t.Fatal("expected error when setting options while running")
	}
}

// ========================================================================
// Context
// ========================================================================

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "sleep", "60"},
		Context:  ctx,
	})
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	p.Wait()

	if p.IsRunning() {
		t.Fatal("process should not be running after context cancellation")
	}
	if s := p.State(); s == nil {
		t.Fatal("State should be non-nil after context cancellation")
	}
}

func TestContextCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "exit", "0"},
		Context:  ctx,
	})
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	p.Wait()
	if p.ExitCode() != -1 {
		t.Log("process started despite cancelled context (timing-dependent)")
	}
}

// ========================================================================
// 集成：同时使用 stdout 和 stderr
// ========================================================================

func TestMixedStdoutStderr(t *testing.T) {
	var outLines, errLines []string
	p := New(Options{
		ExecPath: os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--", "mixed", "hello", "e:oops"},
		OnStdout: func(s string) { outLines = append(outLines, s) },
		OnStderr: func(s string) { errLines = append(errLines, s) },
	})
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if len(outLines) != 1 || outLines[0] != "hello" {
		t.Fatalf("stdout = %v, want [hello]", outLines)
	}
	if len(errLines) != 1 || errLines[0] != "oops" {
		t.Fatalf("stderr = %v, want [oops]", errLines)
	}
}
