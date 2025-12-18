package linq

import (
	"sort"
)

// Linq 是一个泛型查询工具，用于操作切片数据，支持链式调用
type Linq[T any] struct {
	data    []T            // 存储数据的切片
	compare func(T, T) int // 可选的比较函数，用于排序等操作
}

// Group 代表按键分组后的结果
type Group[K comparable, T any] struct {
	Key   K   // 分组的键
	Items []T // 该键对应的元素集合
}

// From 从切片创建 Linq 实例
func From[T any](data []T) Linq[T] {
	return Linq[T]{data: data}
}

// Where 过滤数据，只保留满足条件的元素
func (l Linq[T]) Where(predicate func(T) bool) Linq[T] {
	var result []T
	for _, item := range l.data {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Select 投影数据，将每个元素转换为新值
func (l Linq[T]) Select(selector func(T) T) Linq[T] {
	result := make([]T, len(l.data))
	for i, item := range l.data {
		result[i] = selector(item)
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Sort 对数据进行排序，使用自定义比较函数
func (l Linq[T]) Sort(compareFn func(a, b T) bool) Linq[T] {
	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.Slice(data, func(i, j int) bool {
		return compareFn(data[i], data[j])
	})
	return Linq[T]{data: data, compare: l.compare}
}

// WithComparer 设置自定义比较函数
func (l Linq[T]) WithComparer(compare func(a, b T) int) Linq[T] {
	return Linq[T]{data: l.data, compare: compare}
}

// Any 检查是否存在满足条件的元素
func (l Linq[T]) Any(predicate func(T) bool) bool {
	for _, item := range l.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Distinct 移除重复元素，使用自定义比较函数
// 如果未设置 compare 函数且调用此方法，将 panic
func (l Linq[T]) Distinct() Linq[T] {
	if l.compare == nil {
		panic("Distinct requires a comparer, use WithComparer or DistinctComparable for comparable types")
	}
	return l.distinctWithComparer()
}

// DistinctComparable 移除重复元素，专用于 comparable 类型
func DistinctComparable[T comparable](l Linq[T]) Linq[T] {
	seen := make(map[T]struct{}, len(l.data))
	result := make([]T, 0, len(l.data))
	for _, item := range l.data {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return Linq[T]{data: result, compare: l.compare}
}

// distinctWithComparer 使用自定义比较函数去重
func (l Linq[T]) distinctWithComparer() Linq[T] {
	if len(l.data) <= 1 {
		return Linq[T]{data: l.data, compare: l.compare}
	}

	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.Slice(data, func(i, j int) bool {
		return l.compare(data[i], data[j]) < 0
	})

	result := make([]T, 0, len(data))
	result = append(result, data[0])
	for i := 1; i < len(data); i++ {
		if l.compare(data[i], data[i-1]) != 0 {
			result = append(result, data[i])
		}
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Take 获取前 n 个元素
func (l Linq[T]) Take(n int) Linq[T] {
	if n <= 0 {
		return Linq[T]{compare: l.compare}
	}
	if n >= len(l.data) {
		return Linq[T]{data: l.data, compare: l.compare}
	}
	data := make([]T, n)
	copy(data, l.data[:n])
	return Linq[T]{data: data, compare: l.compare}
}

// Skip 跳过前 n 个元素
func (l Linq[T]) Skip(n int) Linq[T] {
	if n <= 0 {
		return Linq[T]{data: l.data, compare: l.compare}
	}
	if n >= len(l.data) {
		return Linq[T]{compare: l.compare}
	}
	data := make([]T, len(l.data)-n)
	copy(data, l.data[n:])
	return Linq[T]{data: data, compare: l.compare}
}

// GroupBy 按键分组数据
func GroupBy[K comparable, T any](l Linq[T], keySelector func(T) K) []Group[K, T] {
	groups := make(map[K][]T)
	for _, item := range l.data {
		key := keySelector(item)
		groups[key] = append(groups[key], item)
	}

	result := make([]Group[K, T], 0, len(groups))
	for key, items := range groups {
		result = append(result, Group[K, T]{Key: key, Items: items})
	}
	return result
}

// Join 连接两个 Linq 数据集
func Join[T, U, K comparable, R any](outer Linq[T], inner Linq[U],
	outerKeySelector func(T) K, innerKeySelector func(U) K,
	resultSelector func(T, U) R) Linq[R] {
	result := make([]R, 0, len(outer.data)*len(inner.data))
	for _, o := range outer.data {
		outerKey := outerKeySelector(o)
		for _, i := range inner.data {
			if innerKeySelector(i) == outerKey {
				result = append(result, resultSelector(o, i))
			}
		}
	}
	return Linq[R]{data: result}
}

// Concat 合并两个 Linq 数据集
func (l Linq[T]) Concat(other Linq[T]) Linq[T] {
	data := make([]T, 0, len(l.data)+len(other.data))
	data = append(data, l.data...)
	data = append(data, other.data...)
	return Linq[T]{data: data, compare: l.compare}
}

// Reverse 反转数据顺序
func (l Linq[T]) Reverse() Linq[T] {
	data := make([]T, len(l.data))
	for i, j := 0, len(l.data)-1; i < len(l.data); i, j = i+1, j-1 {
		data[i] = l.data[j]
	}
	return Linq[T]{data: data, compare: l.compare}
}

// Min 返回最小元素，要求设置 compare 函数
func (l Linq[T]) Min() (T, bool) {
	if len(l.data) == 0 {
		var zero T
		return zero, false
	}
	if l.compare == nil {
		panic("Min requires a comparer, use WithComparer first")
	}
	min := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.compare(l.data[i], min) < 0 {
			min = l.data[i]
		}
	}
	return min, true
}

// Max 返回最大元素，要求设置 compare 函数
func (l Linq[T]) Max() (T, bool) {
	if len(l.data) == 0 {
		var zero T
		return zero, false
	}
	if l.compare == nil {
		panic("Max requires a comparer, use WithComparer first")
	}
	max := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.compare(l.data[i], max) > 0 {
			max = l.data[i]
		}
	}
	return max, true
}

// Results 返回最终的切片结果
func (l Linq[T]) Results() []T {
	return l.data
}
