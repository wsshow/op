package str

import (
	"strconv"
	"strings"
)

type String struct {
	str string
}

func NewString(s string) *String {
	return &String{str: s}
}

func (s *String) Contain(substr string) bool {
	return strings.Contains(s.str, substr)
}

func (s *String) Index(substr string) int {
	return strings.Index(s.str, substr)
}

func (s *String) LastIndex(substr string) int {
	return strings.LastIndex(s.str, substr)
}

func (s *String) Split(sep string) []string {
	return strings.Split(s.str, sep)
}

func (s *String) Length() int {
	return len(s.str)
}

func (s *String) ReplaceAll(old, new string) *String {
	s.str = strings.ReplaceAll(s.str, old, new)
	return s
}

func (s *String) String() string {
	return s.str
}

func (s *String) ToInt() (int, error) {
	return strconv.Atoi(s.str)
}

func (s *String) Concat(ss ...string) *String {
	for _, st := range ss {
		s.str += st
	}
	return s
}
