package str

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	s := New("hello")
	if s.String() != "hello" {
		t.Errorf("New: expected 'hello', got '%s'", s.String())
	}
}

func TestContains(t *testing.T) {
	s := New("hello world")
	if !s.Contains("world") {
		t.Error("Contains: expected true for 'world'")
	}
	if s.Contains("xyz") {
		t.Error("Contains: expected false for 'xyz'")
	}
}

func TestIndex(t *testing.T) {
	s := New("hello world")
	if idx := s.Index("world"); idx != 6 {
		t.Errorf("Index: expected 6, got %d", idx)
	}
	if idx := s.Index("xyz"); idx != -1 {
		t.Errorf("Index: expected -1, got %d", idx)
	}
}

func TestLastIndex(t *testing.T) {
	s := New("hello hello")
	if idx := s.LastIndex("hello"); idx != 6 {
		t.Errorf("LastIndex: expected 6, got %d", idx)
	}
	if idx := s.LastIndex("xyz"); idx != -1 {
		t.Errorf("LastIndex: expected -1, got %d", idx)
	}
}

func TestCount(t *testing.T) {
	s := New("hello hello hello")
	if n := s.Count("hello"); n != 3 {
		t.Errorf("Count: expected 3, got %d", n)
	}
	if n := s.Count("xyz"); n != 0 {
		t.Errorf("Count: expected 0, got %d", n)
	}
}

func TestSplit(t *testing.T) {
	s := New("a,b,c")
	parts := s.Split(",")
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Errorf("Split: expected [a b c], got %v", parts)
	}
}

func TestFields(t *testing.T) {
	s := New("hello  world\tfoo")
	parts := s.Fields()
	if len(parts) != 3 || parts[0] != "hello" || parts[1] != "world" || parts[2] != "foo" {
		t.Errorf("Fields: expected [hello world foo], got %v", parts)
	}
}

func TestLength(t *testing.T) {
	s := New("hello")
	if s.Length() != 5 {
		t.Errorf("Length: expected 5, got %d", s.Length())
	}
	s2 := New("世界")
	if s2.Length() != 6 {
		t.Errorf("Length: expected 6 (UTF-8 bytes), got %d", s2.Length())
	}
}

func TestRuneLength(t *testing.T) {
	s := New("hello 世界")
	if s.RuneLength() != 8 {
		t.Errorf("RuneLength: expected 8, got %d", s.RuneLength())
	}
}

func TestBytes(t *testing.T) {
	s := New("hello")
	b := s.Bytes()
	if string(b) != "hello" {
		t.Errorf("Bytes: expected 'hello', got '%s'", string(b))
	}
	// Mutation safety
	b[0] = 'x'
	if s.String() != "hello" {
		t.Error("Bytes should return a copy, not the underlying data")
	}
}

func TestReplaceAll(t *testing.T) {
	s := New("hello world world")
	s.ReplaceAll("world", "go")
	if s.String() != "hello go go" {
		t.Errorf("ReplaceAll: expected 'hello go go', got '%s'", s.String())
	}
}

func TestReplace(t *testing.T) {
	s := New("hello world world")
	s.Replace("world", "go", 1)
	if s.String() != "hello go world" {
		t.Errorf("Replace: expected 'hello go world', got '%s'", s.String())
	}
	s.Replace("go", "world", -1)
	if s.String() != "hello world world" {
		t.Errorf("Replace all: expected 'hello world world', got '%s'", s.String())
	}
}

func TestTrim(t *testing.T) {
	s := New("...hello...")
	s.Trim(".")
	if s.String() != "hello" {
		t.Errorf("Trim: expected 'hello', got '%s'", s.String())
	}
}

func TestTrimSpace(t *testing.T) {
	s := New("  hello  ")
	s.TrimSpace()
	if s.String() != "hello" {
		t.Errorf("TrimSpace: expected 'hello', got '%s'", s.String())
	}
}

func TestTrimPrefix(t *testing.T) {
	s := New("https://example.com")
	s.TrimPrefix("https://")
	if s.String() != "example.com" {
		t.Errorf("TrimPrefix: expected 'example.com', got '%s'", s.String())
	}
	// No match — unchanged
	s2 := New("http://example.com")
	s2.TrimPrefix("https://")
	if s2.String() != "http://example.com" {
		t.Errorf("TrimPrefix: expected 'http://example.com', got '%s'", s2.String())
	}
}

func TestTrimSuffix(t *testing.T) {
	s := New("hello.txt")
	s.TrimSuffix(".txt")
	if s.String() != "hello" {
		t.Errorf("TrimSuffix: expected 'hello', got '%s'", s.String())
	}
	// No match — unchanged
	s2 := New("hello.pdf")
	s2.TrimSuffix(".txt")
	if s2.String() != "hello.pdf" {
		t.Errorf("TrimSuffix: expected 'hello.pdf', got '%s'", s2.String())
	}
}

func TestToLower(t *testing.T) {
	s := New("HELLO")
	s.ToLower()
	if s.String() != "hello" {
		t.Errorf("ToLower: expected 'hello', got '%s'", s.String())
	}
}

func TestToUpper(t *testing.T) {
	s := New("hello")
	s.ToUpper()
	if s.String() != "HELLO" {
		t.Errorf("ToUpper: expected 'HELLO', got '%s'", s.String())
	}
}

func TestReverse(t *testing.T) {
	s := New("hello")
	s.Reverse()
	if s.String() != "olleh" {
		t.Errorf("Reverse: expected 'olleh', got '%s'", s.String())
	}
	// Unicode
	s2 := New("你好世界")
	s2.Reverse()
	if s2.String() != "界世好你" {
		t.Errorf("Reverse: expected '界世好你', got '%s'", s2.String())
	}
}

func TestConcat(t *testing.T) {
	s := New("hello")
	result := s.Concat(", ", "world", "!")
	if result.String() != "hello, world!" {
		t.Errorf("Concat: expected 'hello, world!', got '%s'", result.String())
	}
	// Original unchanged
	if s.String() != "hello" {
		t.Errorf("Concat: original should be 'hello', got '%s'", s.String())
	}
}

func TestStartsWith(t *testing.T) {
	s := New("hello.txt")
	if !s.StartsWith("hello") {
		t.Error("StartsWith: expected true for 'hello'")
	}
	if s.StartsWith("world") {
		t.Error("StartsWith: expected false for 'world'")
	}
}

func TestEndsWith(t *testing.T) {
	s := New("hello.txt")
	if !s.EndsWith(".txt") {
		t.Error("EndsWith: expected true for '.txt'")
	}
	if s.EndsWith(".pdf") {
		t.Error("EndsWith: expected false for '.pdf'")
	}
}

func TestEqualFold(t *testing.T) {
	s := New("Hello")
	if !s.EqualFold("hello") {
		t.Error("EqualFold: expected true for 'hello'")
	}
	if !s.EqualFold("HELLO") {
		t.Error("EqualFold: expected true for 'HELLO'")
	}
	if s.EqualFold("world") {
		t.Error("EqualFold: expected false for 'world'")
	}
}

func TestToInt(t *testing.T) {
	s := New("42")
	n, err := s.ToInt()
	if err != nil || n != 42 {
		t.Errorf("ToInt: expected 42, got %d (err: %v)", n, err)
	}
	// With whitespace
	s2 := New("  123  ")
	n, err = s2.ToInt()
	if err != nil || n != 123 {
		t.Errorf("ToInt: expected 123, got %d (err: %v)", n, err)
	}
	// Invalid
	s3 := New("abc")
	_, err = s3.ToInt()
	if err == nil {
		t.Error("ToInt: expected error for 'abc'")
	}
}

func TestMustInt(t *testing.T) {
	s := New("42")
	if n := s.MustInt(); n != 42 {
		t.Errorf("MustInt: expected 42, got %d", n)
	}
	// Invalid should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustInt: expected panic for invalid input")
		}
	}()
	New("abc").MustInt()
}

func TestToFloat(t *testing.T) {
	s := New("3.14")
	f, err := s.ToFloat()
	if err != nil || f != 3.14 {
		t.Errorf("ToFloat: expected 3.14, got %f (err: %v)", f, err)
	}
	// Invalid
	s2 := New("abc")
	_, err = s2.ToFloat()
	if err == nil {
		t.Error("ToFloat: expected error for 'abc'")
	}
}

func TestMustFloat(t *testing.T) {
	s := New("3.14")
	if f := s.MustFloat(); f != 3.14 {
		t.Errorf("MustFloat: expected 3.14, got %f", f)
	}
}

func TestFormat(t *testing.T) {
	template := New("Hello, %s! You have %d messages.")
	result := template.Format("Alice", 5)
	if result.String() != "Hello, Alice! You have 5 messages." {
		t.Errorf("Format: unexpected result '%s'", result.String())
	}
	// Original unchanged
	if template.String() != "Hello, %s! You have %d messages." {
		t.Errorf("Format: template was modified: '%s'", template.String())
	}
}

func TestSubstring(t *testing.T) {
	s := New("hello world")
	// Normal range
	sub := s.Substring(0, 5)
	if sub.String() != "hello" {
		t.Errorf("Substring: expected 'hello', got '%s'", sub.String())
	}
	// Original unchanged
	if s.String() != "hello world" {
		t.Errorf("Substring: original was modified: '%s'", s.String())
	}
	// Unicode
	s2 := New("你好世界")
	sub2 := s2.Substring(0, 2)
	if sub2.String() != "你好" {
		t.Errorf("Substring: expected '你好', got '%s'", sub2.String())
	}
	// Negative end (from end)
	sub3 := New("hello.txt").Substring(0, -4)
	if sub3.String() != "hello" {
		t.Errorf("Substring(0,-4): expected 'hello', got '%s'", sub3.String())
	}
	// Negative start (from end)
	sub4 := New("hello").Substring(-3, -1)
	if sub4.String() != "ll" {
		t.Errorf("Substring(-3,-1): expected 'll', got '%s'", sub4.String())
	}
	// Negative start exceeding length (clamped to 0)
	sub5 := New("ab").Substring(-10, 2)
	if sub5.String() != "ab" {
		t.Errorf("Substring(-10,2): expected 'ab', got '%s'", sub5.String())
	}
	// start >= end returns empty
	sub6 := New("hello").Substring(3, 1)
	if sub6.String() != "" {
		t.Errorf("Substring(3,1): expected '', got '%s'", sub6.String())
	}
}

func TestIsEmpty(t *testing.T) {
	if !New("").IsEmpty() {
		t.Error("IsEmpty: expected true for empty string")
	}
	if New("text").IsEmpty() {
		t.Error("IsEmpty: expected false for non-empty string")
	}
}

func TestIsBlank(t *testing.T) {
	if !New("").IsBlank() {
		t.Error("IsBlank: expected true for empty string")
	}
	if !New("   ").IsBlank() {
		t.Error("IsBlank: expected true for whitespace")
	}
	if !New("\t\n").IsBlank() {
		t.Error("IsBlank: expected true for tabs/newlines")
	}
	if New(" a ").IsBlank() {
		t.Error("IsBlank: expected false for ' a '")
	}
}

func TestClone(t *testing.T) {
	original := New("hello")
	clone := original.Clone()
	clone.ToUpper()
	if original.String() != "hello" {
		t.Errorf("Clone: original modified: '%s'", original.String())
	}
	if clone.String() != "HELLO" {
		t.Errorf("Clone: expected 'HELLO', got '%s'", clone.String())
	}
}

func TestString(t *testing.T) {
	s := New("test")
	if s.String() != "test" {
		t.Errorf("String: expected 'test', got '%s'", s.String())
	}
}

func TestChaining(t *testing.T) {
	s := New("  Hello, World!  ")
	result := s.TrimSpace().
		ReplaceAll("World", "Go").
		ToUpper()
	if result.String() != "HELLO, GO!" {
		t.Errorf("Chaining: expected 'HELLO, GO!', got '%s'", result.String())
	}
}

func TestRepeat(t *testing.T) {
	s := New("ab")
	result := s.Repeat(3)
	if result.String() != "ababab" {
		t.Errorf("Repeat: expected 'ababab', got '%s'", result.String())
	}
	s2 := New("x").Repeat(0)
	if s2.String() != "" {
		t.Errorf("Repeat(0): expected '', got '%s'", s2.String())
	}
}

func TestJoin(t *testing.T) {
	s := Join([]string{"a", "b", "c"}, "-")
	if s.String() != "a-b-c" {
		t.Errorf("Join: expected 'a-b-c', got '%s'", s.String())
	}
	s2 := Join([]string{}, "-")
	if s2.String() != "" {
		t.Errorf("Join empty: expected '', got '%s'", s2.String())
	}
	s3 := Join([]string{"hello"}, ", ")
	if s3.String() != "hello" {
		t.Errorf("Join single: expected 'hello', got '%s'", s3.String())
	}
}

func TestChainingWithConcatAndSubstring(t *testing.T) {
	s := New("hello world")
	// Concat returns new, so chain on the result
	result := s.Concat(" foo", " bar").Substring(0, 5).ToUpper()
	if result.String() != "HELLO" {
		t.Errorf("Chaining+Concat: expected 'HELLO', got '%s'", result.String())
	}
	// s is unchanged by Concat
	if s.String() != "hello world" {
		t.Errorf("Chaining+Concat: original modified: '%s'", s.String())
	}
}

func TestSplitN(t *testing.T) {
	s := New("a,b,c,d")
	parts := s.SplitN(",", 3)
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c,d" {
		t.Errorf("SplitN: expected [a b c,d], got %v", parts)
	}
}

func TestContainsAny(t *testing.T) {
	s := New("hello")
	if !s.ContainsAny("aeiou") {
		t.Error("ContainsAny: 'hello' contains vowels")
	}
	if s.ContainsAny("xyz") {
		t.Error("ContainsAny: 'hello' does not contain xyz")
	}
}

func TestContainsRune(t *testing.T) {
	s := New("hello")
	if !s.ContainsRune('e') {
		t.Error("ContainsRune: 'hello' contains 'e'")
	}
	if s.ContainsRune('x') {
		t.Error("ContainsRune: 'hello' does not contain 'x'")
	}
}

func ExampleNew() {
	s := New("Hello, World!")
	fmt.Println(s.ToLower().String())
	// Output: hello, world!
}

func ExampleString_ReplaceAll() {
	s := New("hello world")
	s.ReplaceAll("world", "earth")
	fmt.Println(s.String())
	// Output: hello earth
}

func ExampleString_Substring() {
	s := New("Hello, 世界!")
	sub := s.Substring(0, 5)
	fmt.Println(sub.String())
	// Output: Hello
}

func ExampleString_Reverse() {
	s := New("Hello")
	s.Reverse()
	fmt.Println(s.String())
	// Output: olleH
}

func ExampleJoin() {
	s := Join([]string{"a", "b", "c"}, "-")
	fmt.Println(s.String())
	// Output: a-b-c
}
