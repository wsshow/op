package emission

import (
	"fmt"
	"testing"
)

// BenchmarkAddRemoveListener 测试添加和删除监听器的性能
func BenchmarkAddRemoveListener(b *testing.B) {
	em := NewEmitter[string, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsubscribe := em.AddListener("test", func(args ...string) {})
		unsubscribe()
	}
}

// BenchmarkRemoveListenerByID 测试在不同大小的监听器列表中删除监听器的性能
func BenchmarkRemoveListenerByID(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			em := NewEmitter[string, string]()

			// 预先添加固定数量的监听器
			cancels := make([]func(), size)
			for i := 0; i < size; i++ {
				cancels[i] = em.AddListener("test", func(args ...string) {})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % size
				// 删除一个监听器
				cancels[idx]()
				// 重新添加以保持列表大小不变
				cancels[idx] = em.AddListener("test", func(args ...string) {})
			}
		})
	}
}

// BenchmarkEmitWithOnce 测试带有Once监听器的Emit性能
func BenchmarkEmitWithOnce(b *testing.B) {
	em := NewEmitter[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次迭代添加10个once监听器
		for j := 0; j < 10; j++ {
			em.Once("test", func(args ...string) {})
		}
		em.EmitSync("test", "data")
	}
}

// BenchmarkOnceListenerCleanup 测试Once监听器清理性能
func BenchmarkOnceListenerCleanup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em := NewEmitter[string, string]()
		// 添加100个once监听器
		for j := 0; j < 100; j++ {
			em.Once("test", func(args ...string) {})
		}
		// 触发清理
		em.EmitSync("test", "data")
	}
}

// BenchmarkConcurrentEmit 测试并发Emit性能
func BenchmarkConcurrentEmit(b *testing.B) {
	em := NewEmitter[string, string]()
	for i := 0; i < 10; i++ {
		em.AddListener("test", func(args ...string) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			em.Emit("test", "data")
		}
	})
}

// BenchmarkMixedOperations 测试混合操作性能
func BenchmarkMixedOperations(b *testing.B) {
	em := NewEmitter[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 添加常规监听器
		cancel1 := em.AddListener("test", func(args ...string) {})
		// 添加once监听器
		em.Once("test", func(args ...string) {})
		// 触发事件
		em.EmitSync("test", "data")
		// 删除常规监听器
		cancel1()
	}
}

// BenchmarkEmitWaitWithConcurrency 测试设置并发度的 EmitWait 性能
func BenchmarkEmitWaitWithConcurrency(b *testing.B) {
	for _, concurrency := range []int{0, 2, 4, 8} {
		name := "Unlimited"
		if concurrency > 0 {
			name = fmt.Sprintf("Max-%d", concurrency)
		}
		b.Run(name, func(b *testing.B) {
			em := NewEmitter[string, string]()
			if concurrency > 0 {
				em.SetConcurrency(concurrency)
			}
			for i := 0; i < 10; i++ {
				em.AddListener("test", func(args ...string) {})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				em.EmitWait("test", "data")
			}
		})
	}
}

// BenchmarkHighFrequencyEmitWithConcurrency 测试高频 Emit + 并发限制
func BenchmarkHighFrequencyEmitWithConcurrency(b *testing.B) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(4)
	for i := 0; i < 10; i++ {
		em.AddListener("test", func(args ...string) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			em.Emit("test", "data")
		}
	})
}
