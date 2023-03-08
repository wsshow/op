package array

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_Add(t *testing.T) {
	type args struct {
		elems []any
	}
	tests := []struct {
		name     string
		a        *Array
		args     args
		expected []any
	}{
		{name: "int", a: &Array{data: []any{1, 2}}, args: args{elems: []any{3}}, expected: []any{1, 2, 3}},
		{name: "string", a: &Array{data: []any{"1"}}, args: args{elems: []any{"2", "3"}}, expected: []any{"1", "2", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.Add(tt.args.elems...)
			assert.Equal(t, tt.expected, tt.a.data, "they should be equal")
		})
	}
}

func TestArray_Remove(t *testing.T) {
	type args struct {
		e any
	}
	tests := []struct {
		name     string
		a        *Array
		args     args
		expected []any
	}{
		{name: "int", a: &Array{data: []any{1, 2, 3}}, args: args{e: 2}, expected: []any{1, 3}},
		{name: "string", a: &Array{data: []any{"1", "2", "3"}}, args: args{e: "1"}, expected: []any{"2", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.Remove(tt.args.e)
			assert.Equal(t, tt.expected, tt.a.data, "they should be equal")
		})
	}
}

func TestArray_RemoveAll(t *testing.T) {
	type args struct {
		e interface{}
	}
	tests := []struct {
		name     string
		a        *Array
		args     args
		expected []any
	}{
		{name: "int", a: &Array{data: []any{1, 2, 3, 3, 3}}, args: args{e: 3}, expected: []any{1, 2}},
		{name: "string", a: &Array{data: []any{"1", "2", "2", "3"}}, args: args{e: "2"}, expected: []any{"1", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.RemoveAll(tt.args.e)
			assert.Equal(t, tt.expected, tt.a.data, "they should be equal")
		})
	}
}
