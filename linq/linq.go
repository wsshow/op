// Package linq 提供了一个泛型 LINQ 风格的查询工具，
// 支持对切片数据进行过滤、投影、排序、分组、集合运算等链式操作。
package linq

import (
	"fmt"
	"sort"
)

// LinqError 表示 LINQ 操作中的错误。
type LinqError struct {
	Op  string // 触发错误的操作名称
	Msg string // 错误描述
}

func (e *LinqError) Error() string {
	return fmt.Sprintf("linq.%s: %s", e.Op, e.Msg)
}

// errNoComparer 返回一个缺少比较函数的错误。
func errNoComparer(op string) error {
	return &LinqError{Op: op, Msg: "requires a comparer, use WithComparer"}
}

// Linq 是一个泛型查询工具，用于对切片数据进行链式操作。
// 零值不可直接使用，请通过 [From] 创建实例。
type Linq[T any] struct {
	data    []T
	compare func(T, T) int
	err     error
}

// Group 代表按键分组后的结果。
type Group[K comparable, T any] struct {
	Key   K   // 分组的键
	Items []T // 该键对应的元素集合
}

// From 从切片创建一个新的 Linq 实例。
// 传入的切片不会被复制，后续操作会生成新的切片。
func From[T any](data []T) Linq[T] {
	return Linq[T]{data: data}
}

// Error 返回链式操作过程中发生的第一个错误。
func (l Linq[T]) Error() error {
	return l.err
}

// Where 过滤数据，只保留满足 predicate 的元素。
func (l Linq[T]) Where(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	result := make([]T, 0, len(l.data))
	for _, item := range l.data {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Select 将每个元素通过 selector 转换为新值。
func (l Linq[T]) Select(selector func(T) T) Linq[T] {
	if l.err != nil {
		return l
	}
	result := make([]T, len(l.data))
	for i, item := range l.data {
		result[i] = selector(item)
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Sort 使用自定义比较函数对数据排序，不修改原始数据。
func (l Linq[T]) Sort(compareFn func(a, b T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.Slice(data, func(i, j int) bool {
		return compareFn(data[i], data[j])
	})
	return Linq[T]{data: data, compare: l.compare}
}

// WithComparer 设置自定义比较函数，供 [Linq.Distinct]、[Linq.Min]、[Linq.Max] 使用。
func (l Linq[T]) WithComparer(compare func(a, b T) int) Linq[T] {
	return Linq[T]{data: l.data, compare: compare, err: l.err}
}

// Any 检查是否存在满足 predicate 的元素。
func (l Linq[T]) Any(predicate func(T) bool) bool {
	for _, item := range l.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Distinct 使用 WithComparer 设置的比较函数移除重复元素。
// 若未设置比较函数，将通过 [Linq.Error] 返回错误。
// 对于 comparable 类型，推荐使用 [DistinctComparable]。
func (l Linq[T]) Distinct() Linq[T] {
	if l.err != nil {
		return l
	}
	if l.compare == nil {
		return Linq[T]{data: l.data, compare: l.compare, err: errNoComparer("Distinct")}
	}
	return l.distinctWithComparer()
}

// DistinctComparable 移除重复元素，专用于 comparable 类型。
func DistinctComparable[T comparable](l Linq[T]) Linq[T] {
	if l.err != nil {
		return l
	}
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

// distinctWithComparer 使用自定义比较函数去重（先排序再相邻去重）。
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

// Take 返回前 n 个元素。n <= 0 返回空序列，n >= 长度返回全部。
func (l Linq[T]) Take(n int) Linq[T] {
	if l.err != nil {
		return l
	}
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

// Skip 跳过前 n 个元素。n <= 0 返回全部，n >= 长度返回空序列。
func (l Linq[T]) Skip(n int) Linq[T] {
	if l.err != nil {
		return l
	}
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

// GroupBy 按 keySelector 返回的键对数据分组。
// 分组结果的顺序与各键首次出现的顺序一致。
func GroupBy[K comparable, T any](l Linq[T], keySelector func(T) K) []Group[K, T] {
	groups := make(map[K][]T)
	var keys []K
	for _, item := range l.data {
		key := keySelector(item)
		if _, exists := groups[key]; !exists {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], item)
	}

	result := make([]Group[K, T], 0, len(groups))
	for _, key := range keys {
		result = append(result, Group[K, T]{Key: key, Items: groups[key]})
	}
	return result
}

// Join 对 outer 和 inner 两个数据集按键进行内连接。
// 对于每对键相同的 (outer, inner) 元素，调用 resultSelector 生成结果。
func Join[T, U, K comparable, R any](outer Linq[T], inner Linq[U],
	outerKeySelector func(T) K, innerKeySelector func(U) K,
	resultSelector func(T, U) R) Linq[R] {
	// 先为 inner 构建索引，将 O(n*m) 优化为 O(n+m)
	innerIndex := make(map[K][]U, len(inner.data))
	for _, i := range inner.data {
		key := innerKeySelector(i)
		innerIndex[key] = append(innerIndex[key], i)
	}

	result := make([]R, 0)
	for _, o := range outer.data {
		key := outerKeySelector(o)
		for _, i := range innerIndex[key] {
			result = append(result, resultSelector(o, i))
		}
	}
	return Linq[R]{data: result}
}

// Concat 将 other 中的元素追加到当前序列之后。
func (l Linq[T]) Concat(other Linq[T]) Linq[T] {
	if l.err != nil {
		return l
	}
	if other.err != nil {
		return other
	}
	data := make([]T, 0, len(l.data)+len(other.data))
	data = append(data, l.data...)
	data = append(data, other.data...)
	return Linq[T]{data: data, compare: l.compare}
}

// Reverse 返回元素顺序反转后的新序列。
func (l Linq[T]) Reverse() Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, len(l.data))
	for i, j := 0, len(l.data)-1; i < len(l.data); i, j = i+1, j-1 {
		data[i] = l.data[j]
	}
	return Linq[T]{data: data, compare: l.compare}
}

// Min 返回最小元素。需要先通过 WithComparer 设置比较函数。
// 若序列为空或存在错误，返回 (零值, false)。
// 若未设置比较函数，错误可通过 Error() 获取。
func (l Linq[T]) Min() (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	if l.compare == nil {
		l.err = errNoComparer("Min")
		return zero, false
	}
	result := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.compare(l.data[i], result) < 0 {
			result = l.data[i]
		}
	}
	return result, true
}

// Max 返回最大元素。需要先通过 WithComparer 设置比较函数。
// 若序列为空或存在错误，返回 (零值, false)。
// 若未设置比较函数，错误可通过 Error() 获取。
func (l Linq[T]) Max() (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	if l.compare == nil {
		l.err = errNoComparer("Max")
		return zero, false
	}
	result := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.compare(l.data[i], result) > 0 {
			result = l.data[i]
		}
	}
	return result, true
}

// Results 返回底层切片的引用。
// 注意：修改返回的切片可能影响 Linq 内部状态。如需安全副本请使用 [Linq.ToSlice]。
func (l Linq[T]) Results() []T {
	return l.data
}

// Count 返回序列中的元素数量。
func (l Linq[T]) Count() int {
	return len(l.data)
}

// CountBy 返回满足 predicate 的元素数量。
func (l Linq[T]) CountBy(predicate func(T) bool) int {
	count := 0
	for _, item := range l.data {
		if predicate(item) {
			count++
		}
	}
	return count
}

// First 返回第一个元素。若序列为空或存在错误，返回 (零值, false)。
func (l Linq[T]) First() (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	return l.data[0], true
}

// FirstBy 返回第一个满足 predicate 的元素。若没有匹配则返回 (零值, false)。
func (l Linq[T]) FirstBy(predicate func(T) bool) (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	for _, item := range l.data {
		if predicate(item) {
			return item, true
		}
	}
	return zero, false
}

// Last 返回最后一个元素。若序列为空或存在错误，返回 (零值, false)。
func (l Linq[T]) Last() (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	return l.data[len(l.data)-1], true
}

// LastBy 返回最后一个满足 predicate 的元素。若没有匹配则返回 (零值, false)。
func (l Linq[T]) LastBy(predicate func(T) bool) (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	for i := len(l.data) - 1; i >= 0; i-- {
		if predicate(l.data[i]) {
			return l.data[i], true
		}
	}
	return zero, false
}

// All 检查是否所有元素都满足 predicate。空序列返回 true。
func (l Linq[T]) All(predicate func(T) bool) bool {
	for _, item := range l.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Contains 检查序列中是否包含指定元素。
func Contains[T comparable](l Linq[T], value T) bool {
	for _, item := range l.data {
		if item == value {
			return true
		}
	}
	return false
}

// ElementAt 返回指定索引处的元素。若索引越界或存在错误，返回 (零值, false)。
func (l Linq[T]) ElementAt(index int) (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if index < 0 || index >= len(l.data) {
		return zero, false
	}
	return l.data[index], true
}

// Append 在序列末尾添加元素。
func (l Linq[T]) Append(elements ...T) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, 0, len(l.data)+len(elements))
	data = append(data, l.data...)
	data = append(data, elements...)
	return Linq[T]{data: data, compare: l.compare}
}

// Prepend 在序列开头添加元素。
func (l Linq[T]) Prepend(elements ...T) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, 0, len(l.data)+len(elements))
	data = append(data, elements...)
	data = append(data, l.data...)
	return Linq[T]{data: data, compare: l.compare}
}

// Union 返回两个序列的并集（去重）。
func Union[T comparable](l1, l2 Linq[T]) Linq[T] {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(l1.data)+len(l2.data))

	for _, item := range l1.data {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	for _, item := range l2.data {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return Linq[T]{data: result}
}

// Intersect 返回两个序列的交集（去重）。
func Intersect[T comparable](l1, l2 Linq[T]) Linq[T] {
	set := make(map[T]struct{})
	for _, item := range l2.data {
		set[item] = struct{}{}
	}

	seen := make(map[T]struct{})
	result := make([]T, 0)

	for _, item := range l1.data {
		if _, exists := set[item]; exists {
			if _, added := seen[item]; !added {
				seen[item] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return Linq[T]{data: result}
}

// Except 返回 l1 中不存在于 l2 的元素（去重）。
func Except[T comparable](l1, l2 Linq[T]) Linq[T] {
	set := make(map[T]struct{})
	for _, item := range l2.data {
		set[item] = struct{}{}
	}

	seen := make(map[T]struct{})
	result := make([]T, 0)

	for _, item := range l1.data {
		if _, exists := set[item]; !exists {
			if _, added := seen[item]; !added {
				seen[item] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return Linq[T]{data: result}
}

// SelectMany 将每个元素通过 selector 映射为切片，并将结果扁平化。
func SelectMany[T, R any](l Linq[T], selector func(T) []R) Linq[R] {
	result := make([]R, 0, len(l.data))
	for _, item := range l.data {
		result = append(result, selector(item)...)
	}
	return Linq[R]{data: result}
}

// Chunk 将序列分割为多个块，每块最多 size 个元素。
// 若 size <= 0，返回空切片。
func (l Linq[T]) Chunk(size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	result := make([][]T, 0, (len(l.data)+size-1)/size)
	for i := 0; i < len(l.data); i += size {
		end := i + size
		if end > len(l.data) {
			end = len(l.data)
		}
		chunk := make([]T, end-i)
		copy(chunk, l.data[i:end])
		result = append(result, chunk)
	}
	return result
}

// DefaultIfEmpty 若序列为空，返回仅包含 defaultValue 的序列。
func (l Linq[T]) DefaultIfEmpty(defaultValue T) Linq[T] {
	if l.err != nil {
		return l
	}
	if len(l.data) == 0 {
		return Linq[T]{data: []T{defaultValue}, compare: l.compare}
	}
	return l
}

// TakeWhile 从序列开头获取元素，直到 predicate 返回 false。
func (l Linq[T]) TakeWhile(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	result := make([]T, 0, len(l.data))
	for _, item := range l.data {
		if !predicate(item) {
			break
		}
		result = append(result, item)
	}
	return Linq[T]{data: result, compare: l.compare}
}

// SkipWhile 从序列开头跳过元素，直到 predicate 返回 false，返回剩余元素。
func (l Linq[T]) SkipWhile(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	start := len(l.data)
	for i, item := range l.data {
		if !predicate(item) {
			start = i
			break
		}
	}
	data := make([]T, len(l.data)-start)
	copy(data, l.data[start:])
	return Linq[T]{data: data, compare: l.compare}
}

// Numeric 约束支持 Sum 和 Average 的数值类型。
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sum 计算数值序列的总和。
func Sum[T Numeric](l Linq[T]) T {
	var sum T
	for _, item := range l.data {
		sum += item
	}
	return sum
}

// Average 计算数值序列的平均值。若序列为空，返回 0。
func Average[T Numeric](l Linq[T]) float64 {
	if len(l.data) == 0 {
		return 0
	}
	var sum T
	for _, item := range l.data {
		sum += item
	}
	return float64(sum) / float64(len(l.data))
}

// ForEach 对序列中的每个元素执行 action。
func (l Linq[T]) ForEach(action func(T)) {
	for _, item := range l.data {
		action(item)
	}
}

// ToSlice 返回底层切片的安全副本。
func (l Linq[T]) ToSlice() []T {
	result := make([]T, len(l.data))
	copy(result, l.data)
	return result
}
