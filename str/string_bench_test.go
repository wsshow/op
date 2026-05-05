package str

import "testing"

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		New("hello world")
	}
}

func BenchmarkContains(b *testing.B) {
	s := New("hello world foo bar baz")
	b.ResetTimer()
	for b.Loop() {
		s.Contains("foo")
	}
}

func BenchmarkReplaceAll(b *testing.B) {
	s := New("hello world world")
	b.ResetTimer()
	for b.Loop() {
		s.ReplaceAll("world", "earth")
	}
}

func BenchmarkToUpper(b *testing.B) {
	s := New("hello world")
	b.ResetTimer()
	for b.Loop() {
		s.ToUpper()
	}
}

func BenchmarkToLower(b *testing.B) {
	s := New("HELLO WORLD")
	b.ResetTimer()
	for b.Loop() {
		s.ToLower()
	}
}

func BenchmarkReverse(b *testing.B) {
	s := New("Hello, 世界")
	b.ResetTimer()
	for b.Loop() {
		s.Reverse()
	}
}

func BenchmarkConcat(b *testing.B) {
	s := New("hello")
	b.ResetTimer()
	for b.Loop() {
		s.Concat(" ", "world")
	}
}

func BenchmarkTrimSpace(b *testing.B) {
	s := New("  hello world  ")
	b.ResetTimer()
	for b.Loop() {
		s.TrimSpace()
	}
}

func BenchmarkSplit(b *testing.B) {
	s := New("a,b,c,d,e,f,g,h,i,j")
	b.ResetTimer()
	for b.Loop() {
		s.Split(",")
	}
}

func BenchmarkToInt(b *testing.B) {
	s := New("12345")
	b.ResetTimer()
	for b.Loop() {
		s.ToInt()
	}
}

func BenchmarkSubstring(b *testing.B) {
	s := New("Hello, 世界!")
	b.ResetTimer()
	for b.Loop() {
		s.Substring(0, 5)
	}
}
