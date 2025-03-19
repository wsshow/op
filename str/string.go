package str

import (
	"fmt"
	"strconv"
	"strings"
)

// String 是一个字符串包装器，提供丰富的字符串操作方法
type String struct {
	str string // 底层字符串
}

// NewString 创建一个新的 String 实例
func NewString(s string) *String {
	return &String{str: s}
}

// Contain 检查字符串是否包含指定子串
func (s *String) Contain(substr string) bool {
	return strings.Contains(s.str, substr)
}

// Index 返回子串第一次出现的位置，若未找到返回 -1
func (s *String) Index(substr string) int {
	return strings.Index(s.str, substr)
}

// LastIndex 返回子串最后一次出现的位置，若未找到返回 -1
func (s *String) LastIndex(substr string) int {
	return strings.LastIndex(s.str, substr)
}

// Split 将字符串按分隔符分割成切片
func (s *String) Split(sep string) []string {
	return strings.Split(s.str, sep)
}

// Length 返回字符串的字符数（字节长度）
func (s *String) Length() int {
	return len(s.str)
}

// RuneLength 返回字符串的 Unicode 字符数（rune 长度）
func (s *String) RuneLength() int {
	return len([]rune(s.str))
}

// ReplaceAll 替换所有匹配的子串，返回自身以支持链式调用
func (s *String) ReplaceAll(old, new string) *String {
	s.str = strings.ReplaceAll(s.str, old, new)
	return s
}

// Trim 去除字符串两端的指定字符集，返回自身
func (s *String) Trim(cutset string) *String {
	s.str = strings.Trim(s.str, cutset)
	return s
}

// TrimSpace 去除字符串两端的空白字符，返回自身
func (s *String) TrimSpace() *String {
	s.str = strings.TrimSpace(s.str)
	return s
}

// ToLower 将字符串转换为小写，返回自身
func (s *String) ToLower() *String {
	s.str = strings.ToLower(s.str)
	return s
}

// ToUpper 将字符串转换为大写，返回自身
func (s *String) ToUpper() *String {
	s.str = strings.ToUpper(s.str)
	return s
}

// Concat 连接多个字符串，返回自身
func (s *String) Concat(ss ...string) *String {
	var builder strings.Builder
	builder.WriteString(s.str)
	for _, st := range ss {
		builder.WriteString(st)
	}
	s.str = builder.String()
	return s
}

// StartsWith 检查字符串是否以指定前缀开头
func (s *String) StartsWith(prefix string) bool {
	return strings.HasPrefix(s.str, prefix)
}

// EndsWith 检查字符串是否以指定后缀结尾
func (s *String) EndsWith(suffix string) bool {
	return strings.HasSuffix(s.str, suffix)
}

// ToInt 将字符串转换为整数，若失败返回错误
func (s *String) ToInt() (int, error) {
	return strconv.Atoi(strings.TrimSpace(s.str))
}

// ToFloat 将字符串转换为浮点数，若失败返回错误
func (s *String) ToFloat() (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s.str), 64)
}

// Format 使用格式化参数生成新字符串，返回自身
func (s *String) Format(args ...interface{}) *String {
	s.str = fmt.Sprintf(s.str, args...)
	return s
}

// Substring 返回指定范围的子串，若越界则调整到合法范围
func (s *String) Substring(start, end int) *String {
	runes := []rune(s.str)
	length := len(runes)
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		s.str = ""
		return s
	}
	s.str = string(runes[start:end])
	return s
}

// IsEmpty 检查字符串是否为空
func (s *String) IsEmpty() bool {
	return len(s.str) == 0
}

// Clone 创建字符串的副本，返回新实例
func (s *String) Clone() *String {
	return NewString(s.str)
}

// String 返回底层字符串值，满足 fmt.Stringer 接口
func (s *String) String() string {
	return s.str
}
