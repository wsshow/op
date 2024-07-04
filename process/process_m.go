package process

import "errors"

type ProcessManager struct {
	processMap map[string]*Process
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processMap: make(map[string]*Process),
	}
}

func (pm *ProcessManager) GetProcess(name string) (p *Process, bFound bool) {
	p, bFound = pm.processMap[name]
	return
}

func (pm *ProcessManager) GetProcesses() (processes []*Process) {
	for _, p := range pm.processMap {
		processes = append(processes, p)
	}
	return
}

func (pm *ProcessManager) AddProcess(co CmdOptions) (err error) {
	name := co.Name
	if _, bFound := pm.GetProcess(name); bFound {
		return errors.New("process already exists")
	}
	process := NewProcess(co).AsyncRun()
	if err := process.Error(); err != nil {
		return err
	}
	pm.processMap[name] = process
	return
}

func (pm *ProcessManager) RemoveProcess(name string) (err error) {
	if p, bFound := pm.GetProcess(name); bFound {
		p.Stop()
		err = p.Error()
	}
	delete(pm.processMap, name)
	return
}
