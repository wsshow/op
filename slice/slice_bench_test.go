package slice

import "testing"

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		New(1, 2, 3, 4, 5)
	}
}

func BenchmarkPush(b *testing.B) {
	s := New[int]()
	b.ResetTimer()
	for b.Loop() {
		s.Push(1)
		if s.Length() > 100 {
			s.Clear()
		}
	}
}

func BenchmarkPop(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		if s.IsEmpty() {
			for i := 0; i < 1000; i++ {
				s.Push(i)
			}
		}
		s.Pop()
	}
}

func BenchmarkShift(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		if s.IsEmpty() {
			for i := 0; i < 1000; i++ {
				s.Push(i)
			}
		}
		s.Shift()
	}
}

func BenchmarkMap(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Map(func(x int) int { return x * 2 })
	}
}

func BenchmarkFilter(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Filter(func(x int) bool { return x%2 == 0 })
	}
}

func BenchmarkFilterNone(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Filter(func(int) bool { return false })
	}
}

func BenchmarkForEach(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	sum := 0
	b.ResetTimer()
	for b.Loop() {
		s.ForEach(func(x int) { sum += x })
	}
}

func BenchmarkSort(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(1000 - i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Sort(func(a, b int) bool { return a < b })
	}
}

func BenchmarkReverse(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Reverse()
	}
}

func BenchmarkReduce(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		s.Reduce(func(a, b int) int { return a + b }, 0)
	}
}

func BenchmarkContains(b *testing.B) {
	s := New[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for b.Loop() {
		Contains(s, 999)
	}
}
