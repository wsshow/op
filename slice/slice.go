package slice

import (
	"sort"
)

// Slice 是一个泛型切片包装器，提供丰富的操作方法。
type Slice[T any] struct {
	data []T
}

// New 创建一个新的 Slice 实例，可传入初始值。
func New[T any](values ...T) *Slice[T] {
	s := &Slice[T]{data: make([]T, 0, len(values))}
	s.data = append(s.data, values...)
	return s
}

// Push 将一个或多个元素添加到切片末尾，返回自身以支持链式调用。
func (s *Slice[T]) Push(values ...T) *Slice[T] {
	s.data = append(s.data, values...)
	return s
}

// Pop 移除并返回切片最后一个元素，若切片为空则返回零值和 false。
func (s *Slice[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	pos := s.Length() - 1
	result := s.data[pos]
	var zero T
	s.data[pos] = zero
	s.data = s.data[:pos]
	return result, true
}

// Shift 移除并返回切片第一个元素，若切片为空则返回零值和 false。
func (s *Slice[T]) Shift() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	result := s.data[0]
	var zero T
	s.data[0] = zero
	s.data = s.data[1:]
	return result, true
}

// Unshift 在切片开头添加一个或多个元素，返回自身。
func (s *Slice[T]) Unshift(values ...T) *Slice[T] {
	s.data = append(values, s.data...)
	return s
}

// Insert 在指定索引处插入元素，返回自身。
// index < 0 时从开头插入，index > Length 时追加到末尾。
func (s *Slice[T]) Insert(index int, values ...T) *Slice[T] {
	if index < 0 {
		index = 0
	}
	if index > s.Length() {
		index = s.Length()
	}
	if len(values) == 0 {
		return s
	}
	n := len(values)
	s.data = append(s.data, values...)
	copy(s.data[index+n:], s.data[index:])
	copy(s.data[index:], values)
	return s
}

// Remove 移除指定索引处的元素并返回。若索引越界则返回零值和 false。
func (s *Slice[T]) Remove(index int) (T, bool) {
	if index < 0 || index >= s.Length() {
		var zero T
		return zero, false
	}
	result := s.data[index]
	var zero T
	copy(s.data[index:], s.data[index+1:])
	s.data[s.Length()-1] = zero
	s.data = s.data[:s.Length()-1]
	return result, true
}

// Length 返回切片长度。
func (s *Slice[T]) Length() int {
	return len(s.data)
}

// IsEmpty 检查切片是否为空。
func (s *Slice[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// ForEach 对每个元素执行回调函数，返回自身以支持链式调用。
func (s *Slice[T]) ForEach(fn func(T)) *Slice[T] {
	for _, v := range s.data {
		fn(v)
	}
	return s
}

// ForEachIndex 对每个元素执行回调函数，回调接收索引和值，返回自身。
func (s *Slice[T]) ForEachIndex(fn func(int, T)) *Slice[T] {
	for i, v := range s.data {
		fn(i, v)
	}
	return s
}

// Map 对每个元素应用映射函数，返回新 Slice，不修改原切片。
func (s *Slice[T]) Map(fn func(T) T) *Slice[T] {
	result := &Slice[T]{data: make([]T, len(s.data))}
	for i, v := range s.data {
		result.data[i] = fn(v)
	}
	return result
}

// Filter 过滤切片，返回一个新 Slice 包含满足条件的元素。
func (s *Slice[T]) Filter(predicate func(T) bool) *Slice[T] {
	result := New[T]()
	for _, v := range s.data {
		if predicate(v) {
			result.data = append(result.data, v)
		}
	}
	return result
}

// Find 查找第一个满足条件的元素，返回该元素和是否存在标志。
func (s *Slice[T]) Find(predicate func(T) bool) (T, bool) {
	for _, v := range s.data {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex 返回第一个满足条件的元素的索引，若未找到则返回 -1。
func (s *Slice[T]) FindIndex(predicate func(T) bool) int {
	for i, v := range s.data {
		if predicate(v) {
			return i
		}
	}
	return -1
}

// First 返回第一个元素。若切片为空则返回零值和 false。
func (s *Slice[T]) First() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.data[0], true
}

// Last 返回最后一个元素。若切片为空则返回零值和 false。
func (s *Slice[T]) Last() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.data[s.Length()-1], true
}

// IndexOf 返回第一个匹配元素的索引，若未找到则返回 -1。
func IndexOf[T comparable](s *Slice[T], value T) int {
	for i, v := range s.data {
		if v == value {
			return i
		}
	}
	return -1
}

// Contains 检查切片中是否包含指定元素。
func Contains[T comparable](s *Slice[T], value T) bool {
	return IndexOf(s, value) != -1
}

// Every 检查是否所有元素都满足条件。空切片返回 true。
func (s *Slice[T]) Every(predicate func(T) bool) bool {
	for _, v := range s.data {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// Some 检查是否至少有一个元素满足条件。空切片返回 false。
func (s *Slice[T]) Some(predicate func(T) bool) bool {
	for _, v := range s.data {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Reduce 从左到右对切片元素进行归约，返回最终结果。
func (s *Slice[T]) Reduce(fn func(acc, cur T) T, initialValue T) T {
	acc := initialValue
	for _, v := range s.data {
		acc = fn(acc, v)
	}
	return acc
}

// Sort 对切片进行原地排序，使用自定义比较函数，返回自身。
func (s *Slice[T]) Sort(less func(a, b T) bool) *Slice[T] {
	sort.Slice(s.data, func(i, j int) bool {
		return less(s.data[i], s.data[j])
	})
	return s
}

// Reverse 原地反转切片顺序，返回自身。
func (s *Slice[T]) Reverse() *Slice[T] {
	for i, j := 0, len(s.data)-1; i < j; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
	return s
}

// Concat 合并当前切片与一个或多个其他切片，返回新 Slice。
func (s *Slice[T]) Concat(others ...*Slice[T]) *Slice[T] {
	total := len(s.data)
	for _, o := range others {
		total += o.Length()
	}
	result := &Slice[T]{data: make([]T, 0, total)}
	result.data = append(result.data, s.data...)
	for _, o := range others {
		result.data = append(result.data, o.data...)
	}
	return result
}

// Sub 返回切片的一个子集，返回新 Slice。
// start: 开始索引（包含），end: 结束索引（不包含）。
func (s *Slice[T]) Sub(start, end int) *Slice[T] {
	if start < 0 {
		start = 0
	}
	if end > s.Length() {
		end = s.Length()
	}
	if start >= end {
		return New[T]()
	}
	result := &Slice[T]{data: make([]T, end-start)}
	copy(result.data, s.data[start:end])
	return result
}

// Get 返回指定索引处的元素，若越界则返回零值和 false。
func (s *Slice[T]) Get(index int) (T, bool) {
	if index < 0 || index >= s.Length() {
		var zero T
		return zero, false
	}
	return s.data[index], true
}

// Set 设置指定索引处的值，若越界则不操作，返回是否成功。
func (s *Slice[T]) Set(index int, value T) bool {
	if index < 0 || index >= s.Length() {
		return false
	}
	s.data[index] = value
	return true
}

// Data 返回底层切片数据的副本。
func (s *Slice[T]) Data() []T {
	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

// Raw 返回底层切片（不拷贝）。调用者不应修改返回的切片。
func (s *Slice[T]) Raw() []T {
	return s.data
}

// Clear 清空切片中的所有元素，返回自身。
func (s *Slice[T]) Clear() *Slice[T] {
	var zero T
	for i := range s.data {
		s.data[i] = zero
	}
	s.data = s.data[:0]
	return s
}

// Clone 创建切片的深拷贝。
func (s *Slice[T]) Clone() *Slice[T] {
	result := &Slice[T]{data: make([]T, len(s.data))}
	copy(result.data, s.data)
	return result
}

// MapTo 对每个元素应用映射函数，转换为新类型，返回新 Slice。
func MapTo[T, U any](s *Slice[T], fn func(T) U) *Slice[U] {
	result := &Slice[U]{data: make([]U, len(s.data))}
	for i, v := range s.data {
		result.data[i] = fn(v)
	}
	return result
}

// ReduceTo 从左到右对切片元素进行归约，累加器类型可不同于元素类型。
func ReduceTo[T, U any](s *Slice[T], fn func(acc U, val T) U, initial U) U {
	acc := initial
	for _, v := range s.data {
		acc = fn(acc, v)
	}
	return acc
}
