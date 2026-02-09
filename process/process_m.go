package process

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessManager 管理多个进程的实例，提供进程的增删改查功能。
type ProcessManager struct {
	processMap map[string]*Process // 存储进程的映射表，键为进程名称
	mu         sync.RWMutex        // 读写锁，确保线程安全
}

// NewProcessManager 创建一个新的 ProcessManager 实例。
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processMap: make(map[string]*Process),
	}
}

// GetProcess 获取指定名称的进程。
// 返回进程实例和是否存在的标志。
func (pm *ProcessManager) GetProcess(name string) (*Process, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, exists := pm.processMap[name]
	return p, exists
}

// GetProcesses 获取所有进程的列表。
func (pm *ProcessManager) GetProcesses() []*Process {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	processes := make([]*Process, 0, len(pm.processMap))
	for _, p := range pm.processMap {
		processes = append(processes, p)
	}
	return processes
}

// AddProcess 添加并启动一个新进程。
// 如果进程名称已存在，返回错误。
// 注意：进程异步启动，启动错误需通过 Process.Error() 或 Process.Wait() 检查。
func (pm *ProcessManager) AddProcess(co CmdOptions) error {
	if co.Name == "" {
		return errors.New("process name cannot be empty")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processMap[co.Name]; exists {
		return fmt.Errorf("process %q already exists", co.Name)
	}

	process := NewProcess(co).AsyncRun()
	pm.processMap[co.Name] = process
	return nil
}

// UpdateProcess 更新现有进程。
// 如果进程不存在，返回错误。停止旧进程并用新进程替换。
func (pm *ProcessManager) UpdateProcess(process *Process) error {
	if process == nil || process.CmdOptions().Name == "" {
		return errors.New("invalid process or empty name")
	}

	name := process.CmdOptions().Name

	pm.mu.Lock()
	oldProcess, exists := pm.processMap[name]
	if !exists {
		pm.mu.Unlock()
		return fmt.Errorf("process %q not found", name)
	}
	pm.processMap[name] = process
	pm.mu.Unlock()

	// 释放锁后停止旧进程，避免长时间持锁阻塞
	if oldProcess.IsRunning() {
		oldProcess.Stop()
	}
	return nil
}

// RemoveProcess 移除指定名称的进程。
// 如果进程存在且正在运行，先停止再删除。
func (pm *ProcessManager) RemoveProcess(name string) error {
	pm.mu.Lock()
	process, exists := pm.processMap[name]
	if exists {
		delete(pm.processMap, name)
	}
	pm.mu.Unlock()

	if !exists {
		return nil
	}

	// 释放锁后停止进程，避免长时间持锁阻塞
	if process.IsRunning() {
		process.Stop()
		if err := process.Error(); err != nil {
			return fmt.Errorf("failed to stop process %q: %w", name, err)
		}
	}
	return nil
}

// StartAll 启动所有已添加但未运行的进程。
// 返回启动过程中遇到的所有错误（合并）。
func (pm *ProcessManager) StartAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errs []error
	for name, process := range pm.processMap {
		if !process.IsRunning() {
			newProcess := NewProcess(process.CmdOptions()).AsyncRun()
			// 检查立即发生的初始化错误（如 ExecPath 为空）
			if err := newProcess.Error(); err != nil {
				errs = append(errs, fmt.Errorf("failed to start process %q: %w", name, err))
				continue
			}
			pm.processMap[name] = newProcess
		}
	}
	return errors.Join(errs...)
}

// StopAll 停止所有正在运行的进程。
// 返回停止过程中遇到的所有错误（合并）。
func (pm *ProcessManager) StopAll() error {
	pm.mu.RLock()
	toStop := make([]*Process, 0, len(pm.processMap))
	names := make([]string, 0, len(pm.processMap))
	for name, process := range pm.processMap {
		if process.IsRunning() {
			toStop = append(toStop, process)
			names = append(names, name)
		}
	}
	pm.mu.RUnlock()

	// 释放锁后停止进程，避免长时间持锁阻塞
	var errs []error
	for i, process := range toStop {
		process.Stop()
		if err := process.Error(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop process %q: %w", names[i], err))
		}
	}
	return errors.Join(errs...)
}

// Count 返回当前管理的进程数量。
func (pm *ProcessManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.processMap)
}

// Clear 移除所有进程并停止运行中的进程。
// 返回停止过程中遇到的所有错误（合并）。
func (pm *ProcessManager) Clear() error {
	pm.mu.Lock()
	toStop := make([]*Process, 0, len(pm.processMap))
	names := make([]string, 0, len(pm.processMap))
	for name, process := range pm.processMap {
		if process.IsRunning() {
			toStop = append(toStop, process)
			names = append(names, name)
		}
		delete(pm.processMap, name)
	}
	pm.mu.Unlock()

	// 释放锁后停止进程，避免长时间持锁阻塞
	var errs []error
	for i, process := range toStop {
		process.Stop()
		if err := process.Error(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop process %q: %w", names[i], err))
		}
	}
	return errors.Join(errs...)
}
