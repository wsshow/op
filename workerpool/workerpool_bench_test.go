package workerpool

import (
	"sync/atomic"
	"testing"
)

func BenchmarkSubmit(b *testing.B) {
	pool := New(4)
	defer pool.Stop()
	var counter atomic.Int32
	b.ResetTimer()
	for b.Loop() {
		pool.Submit(func() {
			counter.Add(1)
		})
	}
}

func BenchmarkSubmitWait(b *testing.B) {
	pool := New(4)
	defer pool.Stop()
	b.ResetTimer()
	for b.Loop() {
		pool.SubmitWait(func() {})
	}
}

func BenchmarkSubmitParallel(b *testing.B) {
	pool := New(8)
	defer pool.Stop()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Submit(func() {})
		}
	})
}

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		p := New(4)
		p.Stop()
	}
}
