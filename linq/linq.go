// Package linq 提供泛型 LINQ 风格的查询工具，
// 支持对切片数据进行过滤、投影、排序、分组、集合运算等链式操作。
//
// Linq 为值类型，非并发安全。
package linq

import (
	"cmp"
	"fmt"
	"sort"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// LinqError 表示 LINQ 操作中的错误。
type LinqError struct {
	Op  string
	Msg string
}

func (e *LinqError) Error() string {
	return fmt.Sprintf("linq.%s: %s", e.Op, e.Msg)
}

// errNoComparer 返回缺少比较函数的错误。
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
	Key   K
	Items []T
}

// OrderedLinq 是对 [Linq] 的包装，支持多级排序。
// 通过 [OrderBy] 或 [OrderByDescending] 创建，然后使用 [OrderedLinq.ThenBy] 追加次级排序条件。
type OrderedLinq[T any] struct {
	Linq[T]
	cmp func(T, T) int
}

// Numeric 约束支持 [Sum] 和 [Average] 的数值类型。
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// From 从切片创建一个新的 Linq 实例。
func From[T any](data []T) Linq[T] {
	return Linq[T]{data: data}
}

// Empty 创建一个空的 Linq 实例。
func Empty[T any]() Linq[T] {
	return Linq[T]{data: []T{}}
}

// Range 生成从 start 开始的 count 个连续整数。
// count <= 0 时返回空序列。
func Range(start, count int) Linq[int] {
	if count <= 0 {
		return Empty[int]()
	}
	if start > int(^uint(0)>>1)-(count-1) {
		return Linq[int]{err: &LinqError{Op: "Range", Msg: "integer overflow"}}
	}
	data := make([]int, count)
	for i := 0; i < count; i++ {
		data[i] = start + i
	}
	return From(data)
}

// Repeat 生成 count 个重复的 value。
// count <= 0 时返回空序列。
func Repeat[T any](value T, count int) Linq[T] {
	if count <= 0 {
		return Empty[T]()
	}
	data := make([]T, count)
	for i := 0; i < count; i++ {
		data[i] = value
	}
	return From(data)
}

// ---------------------------------------------------------------------------
// Error / Conversion
// ---------------------------------------------------------------------------

// Error 返回链式操作过程中发生的第一个错误。
func (l Linq[T]) Error() error {
	return l.err
}

// Results 返回底层切片的引用。
// 修改返回的切片可能影响 Linq 内部状态。如需安全副本请使用 [Linq.ToSlice]。
func (l Linq[T]) Results() []T {
	return l.data
}

// ToSlice 返回底层切片的安全副本。
func (l Linq[T]) ToSlice() []T {
	result := make([]T, len(l.data))
	copy(result, l.data)
	return result
}

// ToMap 将序列转换为 map，使用 keySelector 提取键，valueSelector 提取值。
// 键冲突时后出现的值覆盖先前的值。若序列存在错误，返回 nil。
func ToMap[T any, K comparable, V any](l Linq[T], keySelector func(T) K, valueSelector func(T) V) map[K]V {
	if l.err != nil {
		return nil
	}
	result := make(map[K]V, len(l.data))
	for _, item := range l.data {
		result[keySelector(item)] = valueSelector(item)
	}
	return result
}

// ---------------------------------------------------------------------------
// Filtering
// ---------------------------------------------------------------------------

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

// Distinct 使用 [Linq.WithComparer] 设置的比较函数移除重复元素，保留首次出现的顺序。
// 若未设置比较函数，将通过 [Linq.Error] 返回错误。
// 对于 comparable 类型，推荐使用 [DistinctComparable]；按 key 去重请使用 [DistinctBy]。
func (l Linq[T]) Distinct() Linq[T] {
	if l.err != nil {
		return l
	}
	if l.compare == nil {
		return Linq[T]{data: l.data, compare: l.compare, err: errNoComparer("Distinct")}
	}
	if len(l.data) <= 1 {
		return Linq[T]{data: l.data, compare: l.compare}
	}

	result := make([]T, 0, len(l.data))
	seen := make([]T, 0, len(l.data))
	for _, item := range l.data {
		idx := sort.Search(len(seen), func(i int) bool {
			return l.compare(seen[i], item) >= 0
		})
		if idx < len(seen) && l.compare(seen[idx], item) == 0 {
			continue
		}
		seen = append(seen, item)
		copy(seen[idx+1:], seen[idx:])
		seen[idx] = item
		result = append(result, item)
	}
	return Linq[T]{data: result, compare: l.compare}
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

// DistinctBy 按键去重，保留每个键首次出现的元素。
func DistinctBy[T any, K comparable](l Linq[T], keySelector func(T) K) Linq[T] {
	if l.err != nil {
		return l
	}
	seen := make(map[K]struct{}, len(l.data))
	result := make([]T, 0, len(l.data))
	for _, item := range l.data {
		key := keySelector(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return Linq[T]{data: result, compare: l.compare}
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

// SkipWhile 从序列开头跳过满足 predicate 的元素，返回剩余部分。
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

// ---------------------------------------------------------------------------
// Projection
// ---------------------------------------------------------------------------

// Select 将每个元素通过 selector 转换为同类型的新值。
// 如需变换类型（T → R），请使用 [SelectT]。
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

// SelectT 将每个元素通过 selector 转换为新类型 R。
func SelectT[T, R any](l Linq[T], selector func(T) R) Linq[R] {
	if l.err != nil {
		return Linq[R]{err: l.err}
	}
	result := make([]R, len(l.data))
	for i, item := range l.data {
		result[i] = selector(item)
	}
	return Linq[R]{data: result}
}

// SelectMany 将每个元素通过 selector 映射为切片，并将结果扁平化。
func SelectMany[T, R any](l Linq[T], selector func(T) []R) Linq[R] {
	if l.err != nil {
		return Linq[R]{err: l.err}
	}
	result := make([]R, 0, len(l.data))
	for _, item := range l.data {
		result = append(result, selector(item)...)
	}
	return Linq[R]{data: result}
}

// ---------------------------------------------------------------------------
// Sorting
// ---------------------------------------------------------------------------

// Sort 使用自定义比较函数（less 语义）对数据排序，不修改原始数据。
func (l Linq[T]) Sort(less func(a, b T) bool) Linq[T] {
	if l.err != nil {
		return l
	}
	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.SliceStable(data, func(i, j int) bool {
		return less(data[i], data[j])
	})
	return Linq[T]{data: data, compare: l.compare}
}

// WithComparer 设置自定义比较函数（三值语义），供 [Linq.Distinct]、[Linq.Min]、[Linq.Max] 使用。
func (l Linq[T]) WithComparer(compare func(a, b T) int) Linq[T] {
	return Linq[T]{data: l.data, compare: compare, err: l.err}
}

// OrderBy 按键升序排序。返回 [OrderedLinq] 以支持 [OrderedLinq.ThenBy] 链式多级排序。
func OrderBy[T any, K cmp.Ordered](l Linq[T], keySelector func(T) K) OrderedLinq[T] {
	if l.err != nil {
		return OrderedLinq[T]{Linq: l}
	}
	cmpFn := func(a, b T) int {
		return cmp.Compare(keySelector(a), keySelector(b))
	}
	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.SliceStable(data, func(i, j int) bool { return cmpFn(data[i], data[j]) < 0 })
	return OrderedLinq[T]{
		Linq: Linq[T]{data: data, compare: l.compare},
		cmp:  cmpFn,
	}
}

// OrderByDescending 按键降序排序。返回 [OrderedLinq] 以支持 [OrderedLinq.ThenBy] 链式多级排序。
func OrderByDescending[T any, K cmp.Ordered](l Linq[T], keySelector func(T) K) OrderedLinq[T] {
	if l.err != nil {
		return OrderedLinq[T]{Linq: l}
	}
	cmpFn := func(a, b T) int {
		return cmp.Compare(keySelector(b), keySelector(a))
	}
	data := make([]T, len(l.data))
	copy(data, l.data)
	sort.SliceStable(data, func(i, j int) bool { return cmpFn(data[i], data[j]) < 0 })
	return OrderedLinq[T]{
		Linq: Linq[T]{data: data, compare: l.compare},
		cmp:  cmpFn,
	}
}

// ThenBy 在已有排序基础上追加次级升序比较条件。可多次链式调用，优先级与调用顺序一致。
func (ol OrderedLinq[T]) ThenBy(cmpFn func(a, b T) int) OrderedLinq[T] {
	if ol.err != nil {
		return ol
	}
	data := make([]T, len(ol.data))
	copy(data, ol.data)
	ol.data = data
	prev := ol.cmp
	ol.cmp = func(a, b T) int {
		if c := prev(a, b); c != 0 {
			return c
		}
		return cmpFn(a, b)
	}
	sort.SliceStable(ol.data, func(i, j int) bool { return ol.cmp(ol.data[i], ol.data[j]) < 0 })
	return ol
}

// ThenByDescending 在已有排序基础上追加次级降序比较条件。
func (ol OrderedLinq[T]) ThenByDescending(cmpFn func(a, b T) int) OrderedLinq[T] {
	if ol.err != nil {
		return ol
	}
	// 交换参数反转顺序，避免对 math.MinInt 取负仍为负数的溢出。
	return ol.ThenBy(func(a, b T) int { return cmpFn(b, a) })
}

// ---------------------------------------------------------------------------
// Aggregation
// ---------------------------------------------------------------------------

// Count 返回序列中的元素数量。序列存在错误时返回 0。
func (l Linq[T]) Count() int {
	if l.err != nil {
		return 0
	}
	return len(l.data)
}

// CountBy 返回满足 predicate 的元素数量。序列存在错误时返回 0。
func (l Linq[T]) CountBy(predicate func(T) bool) int {
	if l.err != nil {
		return 0
	}
	count := 0
	for _, item := range l.data {
		if predicate(item) {
			count++
		}
	}
	return count
}

// Sum 计算数值序列的总和。序列存在错误时返回零值。
func Sum[T Numeric](l Linq[T]) T {
	var sum T
	if l.err != nil {
		return sum
	}
	for _, item := range l.data {
		sum += item
	}
	return sum
}

// Average 计算数值序列的平均值，返回 float64。
// 对于 int64/uint64 等大整数类型，float64 转换可能丢失精度。
// 空序列或存在错误时返回 0。
func Average[T Numeric](l Linq[T]) float64 {
	if l.err != nil || len(l.data) == 0 {
		return 0
	}
	return float64(Sum(l)) / float64(len(l.data))
}

// Min 返回序列中的最小元素。需先通过 [Linq.WithComparer] 设置比较函数。
// 序列为空、存在错误或未设置比较函数时返回 (零值, false)。
func (l Linq[T]) Min() (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 || l.compare == nil {
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

// Max 返回序列中的最大元素。需先通过 [Linq.WithComparer] 设置比较函数。
// 序列为空、存在错误或未设置比较函数时返回 (零值, false)。
func (l Linq[T]) Max() (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 || l.compare == nil {
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

// MinBy 返回使 keySelector 值最小的元素。空序列或存在错误时返回 (零值, false)。
func MinBy[T any, K cmp.Ordered](l Linq[T], keySelector func(T) K) (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 {
		return zero, false
	}
	minIdx := 0
	minKey := keySelector(l.data[0])
	for i := 1; i < len(l.data); i++ {
		if key := keySelector(l.data[i]); key < minKey {
			minIdx = i
			minKey = key
		}
	}
	return l.data[minIdx], true
}

// MaxBy 返回使 keySelector 值最大的元素。空序列或存在错误时返回 (零值, false)。
func MaxBy[T any, K cmp.Ordered](l Linq[T], keySelector func(T) K) (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 {
		return zero, false
	}
	maxIdx := 0
	maxKey := keySelector(l.data[0])
	for i := 1; i < len(l.data); i++ {
		if key := keySelector(l.data[i]); key > maxKey {
			maxIdx = i
			maxKey = key
		}
	}
	return l.data[maxIdx], true
}

// MinVal 返回序列中的最小元素，专用于 cmp.Ordered 类型。
// 空序列或存在错误时返回 (零值, false)。
func MinVal[T cmp.Ordered](l Linq[T]) (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 {
		return zero, false
	}
	result := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.data[i] < result {
			result = l.data[i]
		}
	}
	return result, true
}

// MaxVal 返回序列中的最大元素，专用于 cmp.Ordered 类型。
// 空序列或存在错误时返回 (零值, false)。
func MaxVal[T cmp.Ordered](l Linq[T]) (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 {
		return zero, false
	}
	result := l.data[0]
	for i := 1; i < len(l.data); i++ {
		if l.data[i] > result {
			result = l.data[i]
		}
	}
	return result, true
}

// Aggregate 对序列进行累积运算。
// seed 为初始值，fn 接收当前累积值和每个元素，返回新的累积值。
// 若序列存在错误，直接返回 seed。
func Aggregate[T, R any](l Linq[T], seed R, fn func(R, T) R) R {
	if l.err != nil {
		return seed
	}
	result := seed
	for _, item := range l.data {
		result = fn(result, item)
	}
	return result
}

// ---------------------------------------------------------------------------
// Element Access
// ---------------------------------------------------------------------------

// First 返回第一个元素。若序列为空或存在错误，返回 (零值, false)。
func (l Linq[T]) First() (T, bool) {
	var zero T
	if l.err != nil || len(l.data) == 0 {
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
	if l.err != nil || len(l.data) == 0 {
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

// ElementAt 返回指定索引处的元素。若索引越界或存在错误，返回 (零值, false)。
func (l Linq[T]) ElementAt(index int) (T, bool) {
	var zero T
	if l.err != nil || index < 0 || index >= len(l.data) {
		return zero, false
	}
	return l.data[index], true
}

// Single 返回序列中唯一的元素。若序列为空或包含多个元素，返回 (零值, false)。
func (l Linq[T]) Single() (T, bool) {
	var zero T
	if l.err != nil || len(l.data) != 1 {
		return zero, false
	}
	return l.data[0], true
}

// SingleBy 返回序列中唯一满足 predicate 的元素。
// 若无匹配或有多个匹配，返回 (零值, false)。
func (l Linq[T]) SingleBy(predicate func(T) bool) (T, bool) {
	var zero T
	if l.err != nil {
		return zero, false
	}
	var found bool
	var result T
	for _, item := range l.data {
		if predicate(item) {
			if found {
				return zero, false // 多个匹配
			}
			found = true
			result = item
		}
	}
	if !found {
		return zero, false
	}
	return result, true
}

// Contains 检查序列中是否包含指定元素。
func Contains[T comparable](l Linq[T], value T) bool {
	if l.err != nil {
		return false
	}
	for _, item := range l.data {
		if item == value {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Partitioning
// ---------------------------------------------------------------------------

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

// Chunk 将序列分割为多个块，每块最多 size 个元素。
// size <= 0 返回空。存在前置错误时返回 nil。
func (l Linq[T]) Chunk(size int) [][]T {
	if l.err != nil {
		return nil
	}
	if size <= 0 {
		return [][]T{}
	}
	chunkCount := 0
	if len(l.data) > 0 {
		chunkCount = (len(l.data)-1)/size + 1
	}
	result := make([][]T, 0, chunkCount)
	for start := 0; start < len(l.data); {
		end := len(l.data)
		if size < len(l.data)-start {
			end = start + size
		}
		chunk := make([]T, end-start)
		copy(chunk, l.data[start:end])
		result = append(result, chunk)
		if end == len(l.data) {
			break
		}
		start = end
	}
	return result
}

// ---------------------------------------------------------------------------
// Concatenation
// ---------------------------------------------------------------------------

// Concat 将 other 中的元素追加到当前序列之后。
// 若任一方存在错误，返回携带该错误的 Linq。
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

// ---------------------------------------------------------------------------
// Set Operations
// ---------------------------------------------------------------------------

// Union 返回两个序列的并集（去重），保持 l1 元素在前、l2 元素在后的顺序。
func Union[T comparable](l1, l2 Linq[T]) Linq[T] {
	if l1.err != nil {
		return l1
	}
	if l2.err != nil {
		return l2
	}
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
	return Linq[T]{data: result, compare: l1.compare}
}

// Intersect 返回两个序列的交集（去重），保留 l1 中首次出现的顺序。
func Intersect[T comparable](l1, l2 Linq[T]) Linq[T] {
	if l1.err != nil {
		return l1
	}
	if l2.err != nil {
		return l2
	}
	set := make(map[T]struct{}, len(l2.data))
	for _, item := range l2.data {
		set[item] = struct{}{}
	}
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, item := range l1.data {
		if _, inSet := set[item]; inSet {
			if _, added := seen[item]; !added {
				seen[item] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return Linq[T]{data: result, compare: l1.compare}
}

// Except 返回 l1 中不存在于 l2 的元素（去重），保留 l1 中的出现顺序。
func Except[T comparable](l1, l2 Linq[T]) Linq[T] {
	if l1.err != nil {
		return l1
	}
	if l2.err != nil {
		return l2
	}
	set := make(map[T]struct{}, len(l2.data))
	for _, item := range l2.data {
		set[item] = struct{}{}
	}
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, item := range l1.data {
		if _, inSet := set[item]; !inSet {
			if _, added := seen[item]; !added {
				seen[item] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return Linq[T]{data: result, compare: l1.compare}
}

// SequenceEqual 判断两个序列元素逐个相等（使用 == 比较）。
// 长度不同直接返回 false。任一方存在错误时返回 false。
func SequenceEqual[T comparable](l1, l2 Linq[T]) bool {
	if l1.err != nil || l2.err != nil {
		return false
	}
	if len(l1.data) != len(l2.data) {
		return false
	}
	for i := range l1.data {
		if l1.data[i] != l2.data[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Grouping & Joining
// ---------------------------------------------------------------------------

// GroupBy 按 keySelector 返回的键对数据分组。
// 分组顺序与各键首次出现的顺序一致。
func GroupBy[K comparable, T any](l Linq[T], keySelector func(T) K) []Group[K, T] {
	if l.err != nil {
		return nil
	}
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
func Join[T, U, K comparable, R any](outer Linq[T], inner Linq[U],
	outerKeySelector func(T) K, innerKeySelector func(U) K,
	resultSelector func(T, U) R) Linq[R] {
	if outer.err != nil {
		return Linq[R]{err: outer.err}
	}
	if inner.err != nil {
		return Linq[R]{err: inner.err}
	}
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

// LeftJoin 对 outer 和 inner 进行左外连接。
// 对于每个 outer 元素，若 inner 中存在匹配，调用 onMatch；否则调用 onNoMatch。
func LeftJoin[T, U, K comparable, R any](outer Linq[T], inner Linq[U],
	outerKeySelector func(T) K, innerKeySelector func(U) K,
	onMatch func(T, U) R,
	onNoMatch func(T) R) Linq[R] {
	if outer.err != nil {
		return Linq[R]{err: outer.err}
	}
	if inner.err != nil {
		return Linq[R]{err: inner.err}
	}
	innerIndex := make(map[K][]U, len(inner.data))
	for _, i := range inner.data {
		key := innerKeySelector(i)
		innerIndex[key] = append(innerIndex[key], i)
	}
	result := make([]R, 0, len(outer.data))
	for _, o := range outer.data {
		key := outerKeySelector(o)
		if matches, ok := innerIndex[key]; ok {
			for _, m := range matches {
				result = append(result, onMatch(o, m))
			}
		} else {
			result = append(result, onNoMatch(o))
		}
	}
	return Linq[R]{data: result}
}

// Zip 将两个序列按位置配对，对每对元素调用 resultSelector。
// 结果长度等于较短序列的长度。
func Zip[T, U, R any](l1 Linq[T], l2 Linq[U], resultSelector func(T, U) R) Linq[R] {
	if l1.err != nil {
		return Linq[R]{err: l1.err}
	}
	if l2.err != nil {
		return Linq[R]{err: l2.err}
	}
	length := len(l1.data)
	if len(l2.data) < length {
		length = len(l2.data)
	}
	result := make([]R, length)
	for i := 0; i < length; i++ {
		result[i] = resultSelector(l1.data[i], l2.data[i])
	}
	return Linq[R]{data: result}
}

// ---------------------------------------------------------------------------
// Conditional / Execution
// ---------------------------------------------------------------------------

// Any 检查是否存在满足 predicate 的元素。序列存在错误时返回 false。
func (l Linq[T]) Any(predicate func(T) bool) bool {
	if l.err != nil {
		return false
	}
	for _, item := range l.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All 检查是否所有元素都满足 predicate。空序列返回 true，序列存在错误时返回 false。
func (l Linq[T]) All(predicate func(T) bool) bool {
	if l.err != nil {
		return false
	}
	for _, item := range l.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// ForEach 对序列中的每个元素执行 action。若序列存在错误则不执行。
func (l Linq[T]) ForEach(action func(T)) {
	if l.err != nil {
		return
	}
	for _, item := range l.data {
		action(item)
	}
}

// DefaultIfEmpty 若序列为空，返回仅包含 defaultValue 的序列；否则返回原序列。
func (l Linq[T]) DefaultIfEmpty(defaultValue T) Linq[T] {
	if l.err != nil {
		return l
	}
	if len(l.data) == 0 {
		return Linq[T]{data: []T{defaultValue}, compare: l.compare}
	}
	return l
}
