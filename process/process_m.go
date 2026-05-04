package process

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Manager 管理一组具名进程，所有方法均并发安全。
type Manager struct {
	mu    sync.RWMutex
	procs map[string]*Process
}

// NewManager 创建 Manager 实例。
func NewManager() *Manager {
	return &Manager{procs: make(map[string]*Process)}
}

// Add 注册并异步启动一个新进程。name 为空或已存在时返回错误。
func (m *Manager) Add(name string, opts Options) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.procs[name]; exists {
		return fmt.Errorf("process %q already exists", name)
	}
	p := New(opts)
	if err := p.Start(); err != nil {
		return fmt.Errorf("start %q: %w", name, err)
	}
	m.procs[name] = p
	return nil
}

// Get 返回指定名称的进程，不存在时第二个返回值为 false。
func (m *Manager) Get(name string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.procs[name]
	return p, ok
}

// Has 报告指定名称的进程是否已注册。
func (m *Manager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.procs[name]
	return ok
}

// List 返回所有已注册进程的快照，顺序不定。
func (m *Manager) List() []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Process, 0, len(m.procs))
	for _, p := range m.procs {
		list = append(list, p)
	}
	return list
}

// Range 对已注册进程的快照逐一调用 fn，fn 返回 false 时停止。
func (m *Manager) Range(fn func(name string, p *Process) bool) {
	m.mu.RLock()
	snapshot := make([]namedProc, 0, len(m.procs))
	for name, p := range m.procs {
		snapshot = append(snapshot, namedProc{name, p})
	}
	m.mu.RUnlock()

	for _, e := range snapshot {
		if !fn(e.name, e.p) {
			return
		}
	}
}

// SetOptions 替换已注册进程的配置，不重启。进程运行中时返回错误。
func (m *Manager) SetOptions(name string, opts Options) error {
	m.mu.RLock()
	p, ok := m.procs[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process %q not found", name)
	}
	return p.SetOptions(opts)
}

// Restart 重启指定进程，沿用当前配置。
func (m *Manager) Restart(name string) error {
	m.mu.RLock()
	p, ok := m.procs[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process %q not found", name)
	}
	return p.Restart()
}

// Remove 注销并停止指定进程。名称不存在时静默返回。
func (m *Manager) Remove(name string) {
	m.mu.Lock()
	p, exists := m.procs[name]
	if exists {
		delete(m.procs, name)
	}
	m.mu.Unlock()

	if exists {
		p.Stop()
	}
}

// StartAll 启动所有当前未运行的进程（异步启动，不阻塞）。
// 启动阶段的错误（如进程已在运行）会立即返回；
// 运行时错误（如 ExecPath 无效）通过各进程的 OnAfter 或 Wait 获取。
func (m *Manager) StartAll() error {
	m.mu.RLock()
	var idle []namedProc
	for name, p := range m.procs {
		if !p.IsRunning() {
			idle = append(idle, namedProc{name, p})
		}
	}
	m.mu.RUnlock()

	var errs []error
	for _, e := range idle {
		if err := e.p.Start(); err != nil {
			errs = append(errs, fmt.Errorf("start %q: %w", e.name, err))
		}
	}
	return errors.Join(errs...)
}

// StopAll 并发停止所有运行中的进程。
func (m *Manager) StopAll() {
	m.StopAllWithTimeout(defaultStopTimeout)
}

// StopAllWithTimeout 并发停止所有运行中的进程，使用自定义宽限时间。
func (m *Manager) StopAllWithTimeout(timeout time.Duration) {
	targets := m.runningSnapshot()
	stopConcurrently(targets, timeout)
}

// RestartAll 并发停止所有进程后重新启动。返回启动阶段的错误。
func (m *Manager) RestartAll() error {
	targets := m.runningSnapshot()
	stopConcurrently(targets, defaultStopTimeout)
	return m.StartAll()
}

// Clear 并发停止并注销所有进程。
func (m *Manager) Clear() {
	m.mu.Lock()
	snapshot := make([]namedProc, 0, len(m.procs))
	for name, p := range m.procs {
		snapshot = append(snapshot, namedProc{name, p})
	}
	m.procs = make(map[string]*Process)
	m.mu.Unlock()

	stopConcurrently(snapshot, defaultStopTimeout)
}

// Count 返回已注册的进程总数。
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.procs)
}

// RunningCount 返回当前正在运行的进程数。
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, p := range m.procs {
		if p.IsRunning() {
			n++
		}
	}
	return n
}

type namedProc struct {
	name string
	p    *Process
}

// runningSnapshot 返回当前运行中进程的快照。
func (m *Manager) runningSnapshot() []namedProc {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []namedProc
	for name, p := range m.procs {
		if p.IsRunning() {
			list = append(list, namedProc{name, p})
		}
	}
	return list
}

// stopConcurrently 并发停止一组进程，等待全部退出。
func stopConcurrently(targets []namedProc, timeout time.Duration) {
	if len(targets) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, t := range targets {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.p.StopWithTimeout(timeout)
		}()
	}
	wg.Wait()
}
