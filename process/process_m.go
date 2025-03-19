package process

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessManager 管理多个进程的实例，提供进程的增删改查功能
type ProcessManager struct {
	processMap map[string]*Process // 存储进程的映射表，键为进程名称
	mu         sync.RWMutex        // 读写锁，确保线程安全
}

// NewProcessManager 创建一个新的 ProcessManager 实例
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processMap: make(map[string]*Process),
	}
}

// GetProcess 获取指定名称的进程
// 返回进程实例和是否存在标志
func (pm *ProcessManager) GetProcess(name string) (*Process, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, exists := pm.processMap[name]
	return p, exists
}

// GetProcesses 获取所有进程的列表
func (pm *ProcessManager) GetProcesses() []*Process {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	processes := make([]*Process, 0, len(pm.processMap))
	for _, p := range pm.processMap {
		processes = append(processes, p)
	}
	return processes
}

// AddProcess 添加一个新进程
// 如果进程名称已存在或启动失败，返回错误
func (pm *ProcessManager) AddProcess(co CmdOptions) error {
	if co.Name == "" {
		return errors.New("process name cannot be empty")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processMap[co.Name]; exists {
		return errors.New("process already exists")
	}

	process := NewProcess(co).AsyncRun()
	if err := process.Error(); err != nil {
		return err
	}

	pm.processMap[co.Name] = process
	return nil
}

// UpdateProcess 更新现有进程
// 如果进程不存在，返回错误
func (pm *ProcessManager) UpdateProcess(process *Process) error {
	if process == nil || process.CmdOptions().Name == "" {
		return errors.New("invalid process or empty name")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := process.CmdOptions().Name
	if _, exists := pm.processMap[name]; !exists {
		return errors.New("process not found")
	}

	// 停止旧进程并替换
	if oldProcess := pm.processMap[name]; oldProcess.IsRunning() {
		oldProcess.Stop()
	}
	pm.processMap[name] = process
	return nil
}

// RemoveProcess 移除指定名称的进程
// 如果进程存在则停止并删除，返回停止时的错误（如果有）
func (pm *ProcessManager) RemoveProcess(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if process, exists := pm.processMap[name]; exists {
		if process.IsRunning() {
			process.Stop()
			if err := process.Error(); err != nil {
				return err
			}
		}
		delete(pm.processMap, name)
	}
	return nil
}

// StartAll 启动所有已添加但未运行的进程
func (pm *ProcessManager) StartAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var lastErr error
	for name, process := range pm.processMap {
		if !process.IsRunning() {
			newProcess := NewProcess(process.CmdOptions()).AsyncRun()
			if err := newProcess.Error(); err != nil {
				lastErr = fmt.Errorf("failed to start process %s: %v", name, err)
				continue
			}
			pm.processMap[name] = newProcess
		}
	}
	return lastErr
}

// StopAll 停止所有正在运行的进程
func (pm *ProcessManager) StopAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var lastErr error
	for name, process := range pm.processMap {
		if process.IsRunning() {
			process.Stop()
			if err := process.Error(); err != nil {
				lastErr = fmt.Errorf("failed to stop process %s: %v", name, err)
			}
		}
	}
	return lastErr
}

// Count 返回当前管理的进程数量
func (pm *ProcessManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.processMap)
}

// Clear 移除所有进程并停止运行中的进程
func (pm *ProcessManager) Clear() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var lastErr error
	for name, process := range pm.processMap {
		if process.IsRunning() {
			process.Stop()
			if err := process.Error(); err != nil {
				lastErr = fmt.Errorf("failed to stop process %s: %v", name, err)
			}
		}
		delete(pm.processMap, name)
	}
	return lastErr
}
