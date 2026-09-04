package process

import (
	"sync"
	"testing"
	"time"
)

// Manager 测试复用 process_test.go 中的 helperOpts / TestHelperProcess。

func TestManagerAddGetHasCount(t *testing.T) {
	m := NewManager()
	defer m.Clear()

	if err := m.Add("echo", helperOpts("exit", "0")); err != nil {
		t.Fatal(err)
	}
	if m.Count() != 1 {
		t.Fatalf("Count = %d, want 1", m.Count())
	}
	if !m.Has("echo") {
		t.Fatal("Has(echo) = false")
	}
	p, ok := m.Get("echo")
	if !ok {
		t.Fatal("Get(echo) not found")
	}
	p.Wait()
}

func TestManagerZeroValue(t *testing.T) {
	var m Manager
	defer m.Clear()

	if err := m.Add("zero", helperOpts("exit", "0")); err != nil {
		t.Fatalf("zero-value Manager.Add: %v", err)
	}
	p, ok := m.Get("zero")
	if !ok {
		t.Fatal("zero-value Manager did not retain added process")
	}
	if err := p.Wait(); err != nil {
		t.Fatalf("zero-value Manager process: %v", err)
	}
}

func TestManagerAddEmptyName(t *testing.T) {
	m := NewManager()
	if err := m.Add("", helperOpts("exit", "0")); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestManagerAddDuplicate(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	if err := m.Add("p1", helperOpts("exit", "0")); err != nil {
		t.Fatal(err)
	}
	if err := m.Add("p1", helperOpts("exit", "0")); err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if p, ok := m.Get("p1"); ok {
		p.Wait()
	}
}

func TestManagerGetNotFound(t *testing.T) {
	m := NewManager()
	if _, ok := m.Get("nonexistent"); ok {
		t.Fatal("Get should return false for non-existent")
	}
	if m.Has("nonexistent") {
		t.Fatal("Has should return false for non-existent")
	}
}

func TestManagerRemove(t *testing.T) {
	m := NewManager()
	m.Add("p1", helperOpts("exit", "0"))
	if p, ok := m.Get("p1"); ok {
		p.Wait()
	}
	m.Remove("p1")
	if m.Has("p1") {
		t.Fatal("Has(p1) should be false after Remove")
	}
	// Remove non-existent is a no-op
	m.Remove("nonexistent")
}

func TestManagerList(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "0"))
	m.Add("b", helperOpts("exit", "0"))

	list := m.List()
	if len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
	for _, p := range list {
		p.Wait()
	}
}

func TestManagerRange(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "0"))
	m.Add("b", helperOpts("exit", "0"))

	var names []string
	m.Range(func(name string, _ *Process) bool {
		names = append(names, name)
		return true
	})
	if len(names) != 2 {
		t.Fatalf("Range visited %d names, want 2", len(names))
	}

	// Early stop
	names = nil
	m.Range(func(name string, _ *Process) bool {
		names = append(names, name)
		return false
	})
	if len(names) != 1 {
		t.Fatalf("Range with early stop visited %d names, want 1", len(names))
	}

	for _, p := range m.List() {
		p.Wait()
	}
}

func TestManagerStartAll(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "0"))
	m.Add("b", helperOpts("exit", "0"))

	// Wait for initial auto-start to finish
	for _, p := range m.List() {
		p.Wait()
	}
	if m.RunningCount() != 0 {
		t.Fatalf("RunningCount after exit = %d, want 0", m.RunningCount())
	}

	if err := m.StartAll(); err != nil {
		t.Fatal(err)
	}
	if m.RunningCount() != 2 {
		t.Fatalf("RunningCount after StartAll = %d, want 2", m.RunningCount())
	}

	// Verify start-all with already-running skips correctly
	if err := m.StartAll(); err != nil {
		t.Fatal(err)
	}

	m.StopAll()
}

func TestManagerStopAll(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("sleep", "30"))
	m.Add("b", helperOpts("sleep", "30"))
	time.Sleep(100 * time.Millisecond)

	if m.RunningCount() != 2 {
		t.Fatalf("RunningCount = %d, want 2", m.RunningCount())
	}
	m.StopAll()
	if m.RunningCount() != 0 {
		t.Fatalf("RunningCount after StopAll = %d, want 0", m.RunningCount())
	}
}

func TestManagerStopAllWithTimeout(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("sleep", "60"))
	m.Add("b", helperOpts("sleep", "60"))
	time.Sleep(100 * time.Millisecond)

	m.StopAllWithTimeout(50 * time.Millisecond)
	if m.RunningCount() != 0 {
		t.Fatalf("RunningCount after StopAllWithTimeout = %d, want 0", m.RunningCount())
	}
}

func TestManagerRestartAll(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "0"))

	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	if err := m.RestartAll(); err != nil {
		t.Fatal(err)
	}
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	// After RestartAll, previously idle processes are also started.
}

func TestManagerClear(t *testing.T) {
	m := NewManager()
	m.Add("a", helperOpts("sleep", "30"))
	m.Add("b", helperOpts("sleep", "30"))
	m.Clear()

	if m.Count() != 0 {
		t.Fatalf("Count after Clear = %d, want 0", m.Count())
	}
	if m.RunningCount() != 0 {
		t.Fatalf("RunningCount after Clear = %d, want 0", m.RunningCount())
	}
}

func TestManagerRunningCount(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	if n := m.RunningCount(); n != 0 {
		t.Fatalf("RunningCount empty = %d, want 0", n)
	}
	m.Add("a", helperOpts("exit", "0"))
	if n := m.RunningCount(); n != 1 {
		t.Fatalf("RunningCount after Add = %d, want 1", n)
	}
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	if n := m.RunningCount(); n != 0 {
		t.Fatalf("RunningCount after exit = %d, want 0", n)
	}
}

func TestManagerSetOptions(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "1"))
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	if err := m.SetOptions("a", helperOpts("exit", "0")); err != nil {
		t.Fatal(err)
	}
	if err := m.Restart("a"); err != nil {
		t.Fatal(err)
	}
	if p, ok := m.Get("a"); ok {
		p.Wait()
		if p.ExitCode() != 0 {
			t.Fatalf("ExitCode = %d, want 0 after SetOptions", p.ExitCode())
		}
	}
}

func TestManagerSetOptionsNotFound(t *testing.T) {
	m := NewManager()
	if err := m.SetOptions("noexist", helperOpts("exit", "0")); err == nil {
		t.Fatal("expected error for non-existent process")
	}
}

func TestManagerRestartNotFound(t *testing.T) {
	m := NewManager()
	if err := m.Restart("noexist"); err == nil {
		t.Fatal("expected error for non-existent process")
	}
}

func TestManagerRestart(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	m.Add("a", helperOpts("exit", "0"))
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	// Restart via manager
	if err := m.Restart("a"); err != nil {
		t.Fatal(err)
	}
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
}

func TestManagerCount(t *testing.T) {
	m := NewManager()
	defer m.Clear()
	if m.Count() != 0 {
		t.Fatalf("Count = %d, want 0", m.Count())
	}
	m.Add("a", helperOpts("exit", "0"))
	m.Add("b", helperOpts("exit", "0"))
	if m.Count() != 2 {
		t.Fatalf("Count = %d, want 2", m.Count())
	}
	if p, ok := m.Get("a"); ok {
		p.Wait()
	}
	if p, ok := m.Get("b"); ok {
		p.Wait()
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	m := NewManager()
	defer m.Clear()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := string(rune('a' + idx))
			m.Add(name, helperOpts("exit", "0"))
		}(i)
	}
	wg.Wait()

	if m.Count() != 10 {
		t.Fatalf("Count = %d, want 10", m.Count())
	}

	// Cleanup
	for _, p := range m.List() {
		p.Wait()
	}
}
