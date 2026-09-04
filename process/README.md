# Process - Concurrent Process Lifecycle Management

English | [简体中文](README_zh.md)

`process` manages external commands and named groups of commands. `Process` and
`Manager` are safe for concurrent use and keep each process run isolated from
later restarts. The zero value of `Manager` is ready to use.

## Installation

```bash
go get github.com/wsshow/op/process
```

The module requires Go 1.24 or later.

## Basic usage

```go
p := process.New(process.Options{
	ExecPath: "my-server",
	Args:     []string{"--port", "8080"},
	OnStdout: func(line string) { log.Println("OUT:", line) },
	OnStderr: func(line string) { log.Println("ERR:", line) },
})

if err := p.Start(); err != nil {
	log.Fatal(err)
}
if err := p.Wait(); err != nil {
	log.Printf("server exited: %v", err)
}
```

Use `Run` when the caller should block from launch through completion:

```go
if err := process.New(opts).Run(); err != nil {
	log.Fatal(err)
}
```

## Lifecycle contract

- `Start` accepts a run asynchronously. An empty or invalid executable and
  other command-start failures are reported by `Wait` and `Error`, not by the
  successful `Start` call.
- `Run` starts synchronously and returns the error from that exact run.
- `Wait` and `Done` capture the run active when they are called. A concurrent
  restart cannot redirect an existing waiter to the new run.
- `SetOptions` succeeds only while the process is not running. `Args` and `Env`
  are copied, so later changes to the caller's slices do not affect a run.
- `Restart` coalesces overlapping calls. `MinRestartInterval` limits the rate
  of accepted restarts but does not make the wait cancelable.
- `ExitCode` returns `-1` until a process has exited and for abnormal exits.

`Options.Env` follows `os/exec` semantics: nil inherits the parent environment,
while a non-nil empty slice starts the process with an empty environment.

## Cancellation and stopping

Canceling `Options.Context` terminates the command. `Wait` and `Error` preserve
the context error, so callers can use `errors.Is(err, context.Canceled)` or
`errors.Is(err, context.DeadlineExceeded)`.

`Stop` and `StopWithTimeout` are explicit lifecycle operations. They cancel the
command context, wait for library cleanup, and retry with `Process.Kill` if the
run has not completed before the requested timeout. They do not send
`os.Interrupt`; use `Signal` first when the child implements graceful shutdown.

```go
if err := p.Signal(os.Interrupt); err != nil {
	log.Printf("interrupt: %v", err)
}
p.StopWithTimeout(5 * time.Second)
```

Signals and descendant-process handling are operating-system specific. Configure
`SysProcAttr` when process-group behavior is required.

## Callback rules

- `OnBefore` runs before `exec.Cmd.Start`.
- `OnStdout` and `OnStderr` run in independent reader goroutines and may execute
  concurrently. Their callbacks must synchronize shared state themselves.
- `OnAfter` runs after state and errors are published, and before that run's
  `Done` channel closes. It may call `Restart`, but must not wait on the same
  run with `Wait` or `<-Done()`.
- Callback panics are not recovered by this package.

Slow output callbacks delay `Wait`, because all captured output is delivered
before the run is considered complete.

## Managing named processes

```go
m := process.NewManager()
if err := m.Add("api", process.Options{ExecPath: "./api"}); err != nil {
	log.Fatal(err)
}
defer m.StopAll()

m.Range(func(name string, p *process.Process) bool {
	log.Printf("%s: running=%v pid=%d", name, p.IsRunning(), p.Pid())
	return true
})
```

`Range` and `List` operate on snapshots, so callbacks may safely add or remove
manager entries. Bulk stop operations run concurrently and wait for every
process in their initial snapshot.
