package str

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// String 是一个字符串包装器，提供丰富的字符串操作方法。
type String struct {
	str string
}

// New 创建一个新的 String 实例。
func New(s string) *String {
	return &String{str: s}
}

// Contains 检查字符串是否包含指定子串。
func (s *String) Contains(substr string) bool {
	return strings.Contains(s.str, substr)
}

// Index 返回子串第一次出现的位置，若未找到返回 -1。
func (s *String) Index(substr string) int {
	return strings.Index(s.str, substr)
}

// LastIndex 返回子串最后一次出现的位置，若未找到返回 -1。
func (s *String) LastIndex(substr string) int {
	return strings.LastIndex(s.str, substr)
}

// Count 返回子串在字符串中出现的次数。
func (s *String) Count(substr string) int {
	return strings.Count(s.str, substr)
}

// Split 将字符串按分隔符分割成切片。
func (s *String) Split(sep string) []string {
	return strings.Split(s.str, sep)
}

// Fields 将字符串按空白字符分割成切片。
func (s *String) Fields() []string {
	return strings.Fields(s.str)
}

// Length 返回字符串的字节长度。
func (s *String) Length() int {
	return len(s.str)
}

// RuneLength 返回字符串的 Unicode 字符数（rune 长度）。
func (s *String) RuneLength() int {
	return utf8.RuneCountInString(s.str)
}

// ReplaceAll 替换所有匹配的子串，返回自身以支持链式调用。
func (s *String) ReplaceAll(old, new string) *String {
	s.str = strings.ReplaceAll(s.str, old, new)
	return s
}

// Replace 替换前 n 个匹配的子串（n < 0 表示替换所有），返回自身。
func (s *String) Replace(old, new string, n int) *String {
	s.str = strings.Replace(s.str, old, new, n)
	return s
}

// Trim 去除字符串两端的指定字符集，返回自身。
func (s *String) Trim(cutset string) *String {
	s.str = strings.Trim(s.str, cutset)
	return s
}

// TrimSpace 去除字符串两端的空白字符，返回自身。
func (s *String) TrimSpace() *String {
	s.str = strings.TrimSpace(s.str)
	return s
}

// TrimPrefix 去除字符串的前缀，返回自身。
func (s *String) TrimPrefix(prefix string) *String {
	s.str = strings.TrimPrefix(s.str, prefix)
	return s
}

// TrimSuffix 去除字符串的后缀，返回自身。
func (s *String) TrimSuffix(suffix string) *String {
	s.str = strings.TrimSuffix(s.str, suffix)
	return s
}

// ToLower 将字符串转换为小写，返回自身。
func (s *String) ToLower() *String {
	s.str = strings.ToLower(s.str)
	return s
}

// ToUpper 将字符串转换为大写，返回自身。
func (s *String) ToUpper() *String {
	s.str = strings.ToUpper(s.str)
	return s
}

// Reverse 反转字符串（Unicode 感知），返回自身。
func (s *String) Reverse() *String {
	runes := []rune(s.str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	s.str = string(runes)
	return s
}

// Concat 将当前字符串与传入字符串连接，返回新实例。
func (s *String) Concat(ss ...string) *String {
	var builder strings.Builder
	builder.WriteString(s.str)
	for _, st := range ss {
		builder.WriteString(st)
	}
	return New(builder.String())
}

// StartsWith 检查字符串是否以指定前缀开头。
func (s *String) StartsWith(prefix string) bool {
	return strings.HasPrefix(s.str, prefix)
}

// EndsWith 检查字符串是否以指定后缀结尾。
func (s *String) EndsWith(suffix string) bool {
	return strings.HasSuffix(s.str, suffix)
}

// EqualFold 在忽略大小写的情况下比较两个字符串是否相等。
func (s *String) EqualFold(t string) bool {
	return strings.EqualFold(s.str, t)
}

// ToInt 将字符串转换为整数，若失败返回错误。
func (s *String) ToInt() (int, error) {
	return strconv.Atoi(strings.TrimSpace(s.str))
}

// MustInt 将字符串转换为整数，若失败则 panic。
func (s *String) MustInt() int {
	n, err := s.ToInt()
	if err != nil {
		panic(fmt.Sprintf("str: cannot convert %q to int: %v", s.str, err))
	}
	return n
}

// ToFloat 将字符串转换为 float64，若失败返回错误。
func (s *String) ToFloat() (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s.str), 64)
}

// MustFloat 将字符串转换为 float64，若失败则 panic。
func (s *String) MustFloat() float64 {
	f, err := s.ToFloat()
	if err != nil {
		panic(fmt.Sprintf("str: cannot convert %q to float64: %v", s.str, err))
	}
	return f
}

// Format 使用格式化参数生成新字符串，返回新实例。
func (s *String) Format(args ...any) *String {
	return New(fmt.Sprintf(s.str, args...))
}

// Substring 返回指定 rune 范围的子串，返回新实例。
func (s *String) Substring(start, end int) *String {
	runes := []rune(s.str)
	length := len(runes)
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = length + end
	}
	if end > length {
		end = length
	}
	if start >= end {
		return New("")
	}
	return New(string(runes[start:end]))
}

// IsEmpty 检查字符串是否为空。
func (s *String) IsEmpty() bool {
	return len(s.str) == 0
}

// IsBlank 检查字符串是否为空或仅包含空白字符。
func (s *String) IsBlank() bool {
	return strings.TrimSpace(s.str) == ""
}

// Clone 创建字符串的副本，返回新实例。
func (s *String) Clone() *String {
	return New(s.str)
}

// String 返回底层字符串值，满足 fmt.Stringer 接口。
func (s *String) String() string {
	return s.str
}
