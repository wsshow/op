package slice

import (
	"sort"
)

// Slice 是一个泛型切片包装器，提供丰富的操作方法
type Slice[T any] struct {
	data []T // 存储数据的切片
}

// New 创建一个新的 Slice 实例，可传入初始值
func New[T any](values ...T) *Slice[T] {
	return &Slice[T]{data: append([]T{}, values...)}
}

// Push 将一个或多个元素添加到切片末尾，返回自身以支持链式调用
func (s *Slice[T]) Push(values ...T) *Slice[T] {
	s.data = append(s.data, values...)
	return s
}

// Pop 移除并返回切片最后一个元素，若切片为空则返回零值
func (s *Slice[T]) Pop() T {
	if s.IsEmpty() {
		var zero T
		return zero
	}
	pos := s.Length() - 1
	result := s.data[pos]
	s.data = s.data[:pos]
	return result
}

// Shift 移除并返回切片第一个元素，若切片为空则返回零值
func (s *Slice[T]) Shift() T {
	if s.IsEmpty() {
		var zero T
		return zero
	}
	result := s.data[0]
	s.data = s.data[1:]
	return result
}

// Unshift 在切片开头添加一个或多个元素，返回自身
func (s *Slice[T]) Unshift(values ...T) *Slice[T] {
	s.data = append(values, s.data...)
	return s
}

// Length 返回切片长度
func (s *Slice[T]) Length() int {
	return len(s.data)
}

// IsEmpty 检查切片是否为空
func (s *Slice[T]) IsEmpty() bool {
	return s.Length() == 0
}

// Foreach 对每个元素执行回调函数，返回自身以支持链式调用
func (s *Slice[T]) Foreach(callbackfn func(value T)) *Slice[T] {
	for _, v := range s.data {
		callbackfn(v)
	}
	return s
}

// Map 对每个元素应用映射函数并修改原切片，返回自身
func (s *Slice[T]) Map(callbackfn func(value T) T) *Slice[T] {
	for i, v := range s.data {
		s.data[i] = callbackfn(v)
	}
	return s
}

// Filter 过滤切片，返回一个新 Slice 包含满足条件的元素
func (s *Slice[T]) Filter(predicate func(value T) bool) *Slice[T] {
	result := New[T]()
	for _, v := range s.data {
		if predicate(v) {
			result.data = append(result.data, v)
		}
	}
	return result
}

// Find 查找第一个满足条件的元素，返回该元素和是否存在标志
func (s *Slice[T]) Find(predicate func(value T) bool) (result T, existed bool) {
	for _, v := range s.data {
		if predicate(v) {
			return v, true
		}
	}
	return result, false
}

// IndexOf 返回第一个匹配元素的索引，若未找到则返回 -1
// 需要 T 是 comparable 类型，或通过外部比较函数实现
func IndexOf[T comparable](s *Slice[T], value T) int {
	for i, v := range s.data {
		if v == value {
			return i
		}
	}
	return -1
}

// Every 检查是否所有元素都满足条件
func (s *Slice[T]) Every(predicate func(value T) bool) bool {
	for _, v := range s.data {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// Some 检查是否至少有一个元素满足条件
func (s *Slice[T]) Some(predicate func(value T) bool) bool {
	for _, v := range s.data {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Reduce 从左到右对切片元素进行归约，返回最终结果
func (s *Slice[T]) Reduce(callbackfn func(previousValue, currentValue T) T, initialValue T) T {
	acc := initialValue
	for _, v := range s.data {
		acc = callbackfn(acc, v)
	}
	return acc
}

// Sort 对切片进行排序，使用自定义比较函数，返回自身
func (s *Slice[T]) Sort(compareFn func(a, b T) bool) *Slice[T] {
	sort.Slice(s.data, func(i, j int) bool {
		return compareFn(s.data[i], s.data[j])
	})
	return s
}

// Reverse 反转切片顺序，返回自身
func (s *Slice[T]) Reverse() *Slice[T] {
	for i, j := 0, s.Length()-1; i < j; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
	return s
}

// Concat 合并当前切片与另一个切片，返回新 Slice
func (s *Slice[T]) Concat(other *Slice[T]) *Slice[T] {
	result := New[T]()
	result.data = append(result.data, s.data...)
	result.data = append(result.data, other.data...)
	return result
}

// Slice 返回切片的一个子集，返回新 Slice
// start: 开始索引（包含），end: 结束索引（不包含）
func (s *Slice[T]) Slice(start, end int) *Slice[T] {
	if start < 0 {
		start = 0
	}
	if end > s.Length() {
		end = s.Length()
	}
	if start >= end {
		return New[T]()
	}
	return &Slice[T]{data: append([]T{}, s.data[start:end]...)}
}

// Get 返回指定索引处的元素，若越界则返回零值和 false
func (s *Slice[T]) Get(index int) (T, bool) {
	if index < 0 || index >= s.Length() {
		var zero T
		return zero, false
	}
	return s.data[index], true
}

// Set 设置指定索引处的值，若越界则不操作，返回是否成功
func (s *Slice[T]) Set(index int, value T) bool {
	if index < 0 || index >= s.Length() {
		return false
	}
	s.data[index] = value
	return true
}

// Data 返回底层切片数据的副本
func (s *Slice[T]) Data() []T {
	return append([]T{}, s.data...)
}
