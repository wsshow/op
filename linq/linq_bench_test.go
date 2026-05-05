package linq

import (
	"testing"
)

func BenchmarkFrom(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		From(data)
	}
}

func BenchmarkWhere(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		l.Where(func(x int) bool { return x%2 == 0 })
	}
}

func BenchmarkSelect(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		l.Select(func(x int) int { return x * 2 })
	}
}

func BenchmarkSelectT(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		SelectT(l, func(x int) int64 { return int64(x) })
	}
}

func BenchmarkWhereSelect(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		From(data).
			Where(func(x int) bool { return x%2 == 0 }).
			Select(func(x int) int { return x * 2 })
	}
}

func BenchmarkSort(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = 1000 - i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		l.Sort(func(a, b int) bool { return a < b })
	}
}

func BenchmarkOrderBy(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = 1000 - i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		OrderBy(l, func(x int) int { return x })
	}
}

func BenchmarkAggregate(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		Aggregate(l, 0, func(acc, x int) int { return acc + x })
	}
}

func BenchmarkSum(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		Sum(l)
	}
}

func BenchmarkGroupBy(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		GroupBy(l, func(x int) int { return x % 10 })
	}
}

func BenchmarkDistinctComparable(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i % 50
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		DistinctComparable(l)
	}
}

func BenchmarkToMap(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	l := From(data)
	b.ResetTimer()
	for b.Loop() {
		ToMap(l, func(x int) int { return x }, func(x int) string { return string(rune(x)) })
	}
}
