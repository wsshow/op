package op

import (
	"github.com/wsshow/op/slice"
	"github.com/wsshow/op/str"
)

func NewString(s string) *str.String {
	return str.NewString(s)
}

func NewSlice[T any](values ...T) *slice.Slice[T] {
	return slice.New(values...)
}
