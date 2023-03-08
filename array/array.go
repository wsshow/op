package array

import (
	"sort"
)

type Array struct {
	data []any
}

func NewArray() *Array {
	return new(Array)
}

func (a *Array) Add(elems ...any) {
	a.data = append(a.data, elems...)
}

func (a *Array) Remove(e any) {
	d := a.data
	for i, cnt := 0, len(d); i < cnt; i++ {
		if d[i] == e {
			d = append(d[:i], d[i+1:]...)
			break
		}
	}
	a.data = d
}

func (a *Array) RemoveAll(e any) {
	d := a.data
	for i := 0; i < len(d); {
		if d[i] == e {
			d = append(d[:i], d[i+1:]...)
		} else {
			i++
		}
	}
	a.data = d
}

func (a *Array) Contain(e any) bool {
	for _, v := range a.data {
		if v == e {
			return true
		}
	}
	return false
}

func (a *Array) Count() int {
	return len(a.data)
}

func (a *Array) ForEach(f func(any)) {
	for _, v := range a.data {
		f(v)
	}
}

func (a *Array) Clear() {
	a.data = nil
}

func (a *Array) Data() []any {
	return a.data
}

func (a *Array) Sort(less func(i, j int) bool) {
	sort.Slice(a.data, less)
}

func (a *Array) Filter(f func(any) bool) *Array {
	na := NewArray()
	for _, v := range a.data {
		if f(v) {
			na.Add(f(v))
		}
	}
	return na
}

func (a *Array) Map(f func(any) any) *Array {
	na := NewArray()
	for _, v := range a.data {
		na.Add(f(v))
	}
	return na
}
