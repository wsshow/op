package emission

import (
	"fmt"
	"testing"
)

func BenchmarkAddRemoveListener(b *testing.B) {
	em := NewEmitter[string, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub := em.AddListener("test", func(s string) {})
		sub.Unsubscribe()
	}
}

func BenchmarkRemoveListenerByID(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			em := NewEmitter[string, string]()

			subs := make([]*Subscription[string, string], size)
			for i := 0; i < size; i++ {
				subs[i] = em.AddListener("test", func(s string) {})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % size
				subs[idx].Unsubscribe()
				subs[idx] = em.AddListener("test", func(s string) {})
			}
		})
	}
}

func BenchmarkEmitWithOnce(b *testing.B) {
	em := NewEmitter[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			em.Once("test", func(s string) {})
		}
		em.EmitSync("test", "data")
	}
}

func BenchmarkOnceListenerCleanup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em := NewEmitter[string, string]()
		for j := 0; j < 100; j++ {
			em.Once("test", func(s string) {})
		}
		em.EmitSync("test", "data")
	}
}

func BenchmarkConcurrentEmit(b *testing.B) {
	em := NewEmitter[string, string]()
	for i := 0; i < 10; i++ {
		em.AddListener("test", func(s string) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			em.Emit("test", "data")
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	em := NewEmitter[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub := em.AddListener("test", func(s string) {})
		em.Once("test", func(s string) {})
		em.EmitSync("test", "data")
		sub.Unsubscribe()
	}
}

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
				em.AddListener("test", func(s string) {})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				em.EmitWait("test", "data")
			}
		})
	}
}

func BenchmarkHighFrequencyEmitWithConcurrency(b *testing.B) {
	em := NewEmitter[string, string]()
	em.SetConcurrency(4)
	for i := 0; i < 10; i++ {
		em.AddListener("test", func(s string) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			em.Emit("test", "data")
		}
	})
}
