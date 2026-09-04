# Process - 并发安全的进程生命周期管理

[English](README.md) | 简体中文

`process` 用于管理外部命令以及具名命令集合。`Process` 和 `Manager` 均可
并发使用，并且每次进程运行的状态都与后续重启代次相互隔离；`Manager` 零值
可以直接使用。

## 安装

```bash
go get github.com/wsshow/op/process
```

模块要求 Go 1.24 或更高版本。

## 基本用法

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
	log.Printf("服务退出：%v", err)
}
```

调用方需要从启动一直阻塞到退出时，可以使用 `Run`：

```go
if err := process.New(opts).Run(); err != nil {
	log.Fatal(err)
}
```

## 生命周期契约

- `Start` 异步接受一次运行。可执行文件为空、路径错误等启动失败通过 `Wait`
  和 `Error` 报告，而不是由已经成功返回的 `Start` 报告。
- `Run` 同步启动，并返回该次运行自身的错误。
- `Wait` 和 `Done` 捕获调用时对应的运行代次；并发重启不会把已有等待者转移到
  新代次。
- `SetOptions` 只能在进程未运行时成功。`Args` 和 `Env` 会被复制，调用方随后
  修改原切片不会影响运行。
- `Restart` 会合并相互重叠的调用。`MinRestartInterval` 限制成功重启的频率，
  但间隔等待本身不可取消。
- 进程尚未退出或异常终止时，`ExitCode` 返回 `-1`。

`Options.Env` 遵循 `os/exec` 语义：nil 表示继承父进程环境；非 nil 的空切片
表示以空环境启动进程。

## 取消与停止

取消 `Options.Context` 会终止命令。`Wait` 和 `Error` 会保留 context 错误，
调用方可以使用 `errors.Is(err, context.Canceled)` 或
`errors.Is(err, context.DeadlineExceeded)` 判断原因。

`Stop` 和 `StopWithTimeout` 是显式生命周期操作：取消命令 context，等待库内
清理，并在指定时间内仍未完成时再次调用 `Process.Kill`。它们不会发送
`os.Interrupt`；如果子进程支持优雅退出，应先调用 `Signal`。

```go
if err := p.Signal(os.Interrupt); err != nil {
	log.Printf("interrupt: %v", err)
}
p.StopWithTimeout(5 * time.Second)
```

信号和子孙进程处理具有操作系统差异；需要进程组行为时应配置 `SysProcAttr`。

## 回调规则

- `OnBefore` 在 `exec.Cmd.Start` 之前运行。
- `OnStdout` 和 `OnStderr` 在独立读取 goroutine 中运行，可能并发执行；回调
  自行负责同步共享状态。
- `OnAfter` 在状态和错误发布后、该次运行的 `Done` 关闭前执行。它可以调用
  `Restart`，但不能通过 `Wait` 或 `<-Done()` 等待同一个运行代次。
- 本包不会恢复回调中的 panic。

输出回调过慢会延迟 `Wait`，因为一次运行只有在捕获的输出全部交付后才算完成。

## 管理具名进程

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

`Range` 和 `List` 基于快照工作，因此回调可以安全地新增或删除 Manager 条目。
批量停止操作会并发执行，并等待初始快照中的所有进程结束。
