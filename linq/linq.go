package linq

import (
	"sort"
)

type Linq[T comparable] struct {
	arr []T
}

func From[T comparable](arr []T) (l Linq[T]) {
	l.arr = arr
	return l
}

func (l Linq[T]) Where(predicate func(T) bool) Linq[T] {
	var res []T
	for _, v := range l.arr {
		if predicate(v) {
			res = append(res, v)
		}
	}
	l.arr = res
	return l
}

func (l Linq[T]) Sort(compareFn func(a, b T) bool) Linq[T] {
	sort.Slice(l.arr, func(i, j int) bool {
		return compareFn(l.arr[i], l.arr[j])
	})
	return l
}

func (l Linq[T]) Results() []T {
	return l.arr
}
