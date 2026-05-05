package generator

import "testing"

func BenchmarkNewGenerator(b *testing.B) {
	for b.Loop() {
		g := NewGenerator(func(y Yield[int]) {
			y.Send(1)
		})
		g.Next()
		g.Stop()
	}
}

func BenchmarkGeneratorNext(b *testing.B) {
	g := NewGenerator(func(y Yield[int]) {
		for i := 0; i < b.N; i++ {
			if y.Stopped() {
				return
			}
			y.Send(i)
		}
	})
	defer g.Stop()
	b.ResetTimer()
	for b.Loop() {
		g.Next()
	}
}

func BenchmarkGeneratorSend(b *testing.B) {
	g := NewGenerator(func(y Yield[int]) {
		for i := 0; i < b.N; i++ {
			if y.Stopped() {
				return
			}
			result := y.Send(i)
			_ = result
		}
	})
	defer g.Stop()
	b.ResetTimer()
	for b.Loop() {
		g.Next()
	}
}
