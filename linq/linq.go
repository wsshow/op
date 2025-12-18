package linq

import (
	"fmt"
	"sort"
)

// LinqError 表示 LINQ 操作中的错误
type LinqError struct {
	Op  string // 操作名称
	Msg string // 错误消息
}

func (e *LinqError) Error() string {
	return fmt.Sprintf("linq.%s: %s", e.Op, e.Msg)
}

// Linq 是一个泛型查询工具，用于操作切片数据，支持链式调用
type Linq[T any] struct {
	data    []T            // 存储数据的切片
	compare func(T, T) int // 可选的比较函数，用于排序等操作
	err     error          // 存储操作过程中的错误
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

// Error 返回链式操作过程中发生的错误
func (l Linq[T]) Error() error {
	return l.err
}

// Where 过滤数据，只保留满足条件的元素
func (l Linq[T]) Where(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
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
	if l.err != nil {
		return l
	}
	result := make([]T, len(l.data))
	for i, item := range l.data {
		result[i] = selector(item)
	}
	return Linq[T]{data: result, compare: l.compare}
}

// Sort 对数据进行排序，使用自定义比较函数
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
// 如果未设置 compare 函数，将设置错误并返回空结果
func (l Linq[T]) Distinct() Linq[T] {
	if l.err != nil {
		return l
	}
	if l.compare == nil {
		return Linq[T]{err: &LinqError{Op: "Distinct", Msg: "requires a comparer, use WithComparer or DistinctComparable for comparable types"}}
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

// Skip 跳过前 n 个元素
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

// Reverse 反转数据顺序
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

// Min 返回最小元素，要求设置 compare 函数
func (l Linq[T]) Min() (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	if l.compare == nil {
		return zero, false
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
	var zero T
	if l.err != nil {
		return zero, false
	}
	if len(l.data) == 0 {
		return zero, false
	}
	if l.compare == nil {
		return zero, false
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

// Count 返回元素数量
func (l Linq[T]) Count() int {
	return len(l.data)
}

// CountBy 返回满足条件的元素数量
func (l Linq[T]) CountBy(predicate func(T) bool) int {
	count := 0
	for _, item := range l.data {
		if predicate(item) {
			count++
		}
	}
	return count
}

// First 返回第一个元素，如果为空则返回零值并设置错误
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

// FirstOrDefault 返回第一个元素，如果为空则返回零值和 false
func (l Linq[T]) FirstOrDefault() (T, bool) {
	if len(l.data) == 0 {
		var zero T
		return zero, false
	}
	return l.data[0], true
}

// FirstBy 返回第一个满足条件的元素，如果没有则返回零值和 false
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

// FirstByOrDefault 返回第一个满足条件的元素，如果没有则返回零值和 false
func (l Linq[T]) FirstByOrDefault(predicate func(T) bool) (T, bool) {
	for _, item := range l.data {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Last 返回最后一个元素，如果为空则返回零值和 false
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

// LastOrDefault 返回最后一个元素，如果为空则返回零值和 false
func (l Linq[T]) LastOrDefault() (T, bool) {
	if len(l.data) == 0 {
		var zero T
		return zero, false
	}
	return l.data[len(l.data)-1], true
}

// LastBy 返回最后一个满足条件的元素，如果没有则返回零值和 false
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

// LastByOrDefault 返回最后一个满足条件的元素，如果没有则返回零值和 false
func (l Linq[T]) LastByOrDefault(predicate func(T) bool) (T, bool) {
	for i := len(l.data) - 1; i >= 0; i-- {
		if predicate(l.data[i]) {
			return l.data[i], true
		}
	}
	var zero T
	return zero, false
}

// All 检查是否所有元素都满足条件
func (l Linq[T]) All(predicate func(T) bool) bool {
	for _, item := range l.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Contains 检查是否包含指定元素（需要 comparable 类型）
func Contains[T comparable](l Linq[T], value T) bool {
	for _, item := range l.data {
		if item == value {
			return true
		}
	}
	return false
}

// ElementAt 返回指定索引的元素，如果索引越界则返回零值和 false
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

// ElementAtOrDefault 返回指定索引的元素，如果索引越界则返回零值和 false
func (l Linq[T]) ElementAtOrDefault(index int) (T, bool) {
	if index < 0 || index >= len(l.data) {
		var zero T
		return zero, false
	}
	return l.data[index], true
}

// Append 在末尾添加元素
func (l Linq[T]) Append(elements ...T) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, 0, len(l.data)+len(elements))
	data = append(data, l.data...)
	data = append(data, elements...)
	return Linq[T]{data: data, compare: l.compare}
}

// Prepend 在开头添加元素
func (l Linq[T]) Prepend(elements ...T) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, 0, len(l.data)+len(elements))
	data = append(data, elements...)
	data = append(data, l.data...)
	return Linq[T]{data: data, compare: l.compare}
}

// Union 返回两个集合的并集（去重，需要 comparable 类型）
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

// Intersect 返回两个集合的交集（需要 comparable 类型）
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

// Except 返回第一个集合中不在第二个集合中的元素（需要 comparable 类型）
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

// SelectMany 扁平化映射，将每个元素映射为一个切片，然后扁平化
func SelectMany[T, R any](l Linq[T], selector func(T) []R) Linq[R] {
	result := make([]R, 0)
	for _, item := range l.data {
		result = append(result, selector(item)...)
	}
	return Linq[R]{data: result}
}

// Chunk 将数据分块，每块最多包含 size 个元素
// 如果 size <= 0，返回空切片
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

// DefaultIfEmpty 如果序列为空则返回包含默认值的序列
func (l Linq[T]) DefaultIfEmpty(defaultValue T) Linq[T] {
	if l.err != nil {
		return l
	}
	if len(l.data) == 0 {
		return Linq[T]{data: []T{defaultValue}, compare: l.compare}
	}
	return l
}

// TakeWhile 从开头获取元素直到条件不满足
func (l Linq[T]) TakeWhile(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	result := make([]T, 0)
	for _, item := range l.data {
		if !predicate(item) {
			break
		}
		result = append(result, item)
	}
	return Linq[T]{data: result, compare: l.compare}
}

// SkipWhile 从开头跳过元素直到条件不满足
func (l Linq[T]) SkipWhile(predicate func(T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	start := 0
	for i, item := range l.data {
		if !predicate(item) {
			start = i
			break
		}
	}

	if start == 0 && len(l.data) > 0 && predicate(l.data[len(l.data)-1]) {
		return Linq[T]{compare: l.compare}
	}

	data := make([]T, len(l.data)-start)
	copy(data, l.data[start:])
	return Linq[T]{data: data, compare: l.compare}
}

// Sum 计算数值序列的总和（仅支持数值类型）
func Sum[T interface {
	int | int64 | float64 | float32
}](l Linq[T]) T {
	var sum T
	for _, item := range l.data {
		sum += item
	}
	return sum
}

// Average 计算数值序列的平均值（仅支持数值类型）
func Average[T interface {
	int | int64 | float64 | float32
}](l Linq[T]) float64 {
	if len(l.data) == 0 {
		return 0
	}
	var sum T
	for _, item := range l.data {
		sum += item
	}
	return float64(sum) / float64(len(l.data))
}

// ForEach 对每个元素执行操作（不返回新的 Linq）
func (l Linq[T]) ForEach(action func(T)) {
	for _, item := range l.data {
		action(item)
	}
}

// ToSlice 返回底层切片的副本（Results 的别名）
func (l Linq[T]) ToSlice() []T {
	result := make([]T, len(l.data))
	copy(result, l.data)
	return result
}
