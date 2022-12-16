package array

import (
	"sort"
)

type Array struct {
	data []interface{}
}

func NewArray() *Array {
	return new(Array)
}

func (a *Array) Add(elems ...interface{}) {
	a.data = append(a.data, elems...)
}

func (a *Array) Remove(e interface{}) {
	d := a.data
	cnt := len(d)
	for i := 0; i < cnt; i++ {
		if d[i] == e {
			d = append(d[:i], d[i+1:]...)
			break
		}
	}
	a.data = d
}

func (a *Array) RemoveAll(e interface{}) {
	d := a.data
	cnt := len(d)
	for i := 0; i < cnt; i++ {
		if d[i] == e {
			d = append(d[:i], d[i+1:]...)
			continue
		}
	}
	a.data = d
}

func (a *Array) Contain(e interface{}) bool {
	d := a.data
	cnt := len(d)
	for i := 0; i < cnt; i++ {
		if d[i] == e {
			return true
		}
	}
	return false
}

func (a *Array) Count() int {
	return len(a.data)
}

func (a *Array) ForEach(f func(e interface{})) {
	d := a.data
	cnt := len(d)
	for i := 0; i < cnt; i++ {
		f(d[i])
	}
}

func (a *Array) Clear() {
	a.data = nil
}

func (a *Array) Data() []interface{} {
	return a.data
}

func (a *Array) Sort(less func(i, j int) bool) {
	sort.Slice(a.data, less)
}

func (a *Array) Filter(f func(e interface{}) bool) *Array {
	newArr := NewArray()
	d := a.data
	cnt := len(d)
	for i := 0; i < cnt; i++ {
		if f(d[i]) {
			newArr.Add(d[i])
		}
	}
	return newArr
}
