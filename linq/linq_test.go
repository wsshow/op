package linq

import (
	"cmp"
	"fmt"
	"reflect"
	"testing"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

func TestFrom(t *testing.T) {
	l := From([]int{1, 2, 3})
	if len(l.data) != 3 || l.compare != nil {
		t.Errorf("From should create Linq with correct data and nil comparer, got %v", l)
	}
}

func TestEmpty(t *testing.T) {
	l := Empty[int]()
	if l.Count() != 0 {
		t.Errorf("Empty should have 0 elements, got %d", l.Count())
	}
}

func TestRange(t *testing.T) {
	l := Range(5, 3)
	if !reflect.DeepEqual(l.Results(), []int{5, 6, 7}) {
		t.Errorf("Range(5, 3) = %v", l.Results())
	}
	if Range(0, 0).Count() != 0 {
		t.Error("Range(0, 0) should be empty")
	}
	if Range(0, -1).Count() != 0 {
		t.Error("Range with negative count should be empty")
	}
	maxInt := int(^uint(0) >> 1)
	overflow := Range(maxInt, 2)
	if overflow.Error() == nil || overflow.Count() != 0 {
		t.Fatalf("overflowing Range should return an errored sequence, got error=%v count=%d", overflow.Error(), overflow.Count())
	}
}

func TestRepeat(t *testing.T) {
	l := Repeat("x", 3)
	if !reflect.DeepEqual(l.Results(), []string{"x", "x", "x"}) {
		t.Errorf("Repeat('x', 3) = %v", l.Results())
	}
	if Repeat(1, 0).Count() != 0 {
		t.Error("Repeat with count 0 should be empty")
	}
}

// ---------------------------------------------------------------------------
// Filtering
// ---------------------------------------------------------------------------

func TestWhere(t *testing.T) {
	l := From([]int{1, 2, 3, 4}).Where(func(x int) bool { return x%2 == 0 })
	expected := []int{2, 4}
	if !reflect.DeepEqual(l.Results(), expected) {
		t.Errorf("Where expected %v, got %v", expected, l.Results())
	}
}

func TestDistinctComparable(t *testing.T) {
	result := DistinctComparable(From([]int{1, 2, 2, 3, 1})).Results()
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("DistinctComparable expected %v, got %v", expected, result)
	}
}

func TestDistinct(t *testing.T) {
	l := From([]string{"b", "A", "a", "B", "c"}).WithComparer(func(a, b string) int {
		return cmp.Compare(stringsToLower(a), stringsToLower(b))
	})
	result := l.Distinct().Results()
	if len(result) != 3 {
		t.Errorf("Distinct should return 3 elements, got %d: %v", len(result), result)
	}
	// 验证保留首次出现顺序
	if result[0] != "b" {
		t.Errorf("Distinct: first element should be 'b' (first occurrence), got %q", result[0])
	}
	if result[1] != "A" {
		t.Errorf("Distinct: second element should be 'A' (first occurrence), got %q", result[1])
	}
	if result[2] != "c" {
		t.Errorf("Distinct: third element should be 'c', got %q", result[2])
	}

	// 未设置比较函数时返回错误
	l2 := From([]int{1, 2, 2})
	result2 := l2.Distinct()
	if result2.Error() == nil {
		t.Error("Distinct without comparer should set error")
	}
}

func TestDistinctBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	people := From([]Person{{"Alice", 25}, {"Bob", 30}, {"Alice", 35}})
	result := DistinctBy(people, func(p Person) string { return p.Name }).Results()
	if len(result) != 2 {
		t.Fatalf("DistinctBy should return 2 elements, got %d", len(result))
	}
	// 保留首次出现的 "Alice", 25
	if result[0].Name != "Alice" || result[0].Age != 25 {
		t.Errorf("DistinctBy should keep first occurrence, got %v", result[0])
	}
}

func TestTakeWhile(t *testing.T) {
	result := From([]int{1, 2, 3, 4}).TakeWhile(func(x int) bool { return x < 3 }).Results()
	if !reflect.DeepEqual(result, []int{1, 2}) {
		t.Errorf("TakeWhile expected [1,2], got %v", result)
	}
	// 全部满足
	result = From([]int{1, 2}).TakeWhile(func(x int) bool { return x < 10 }).Results()
	if !reflect.DeepEqual(result, []int{1, 2}) {
		t.Errorf("TakeWhile all match: got %v", result)
	}
	// 首元素不满足
	result = From([]int{5, 1, 2}).TakeWhile(func(x int) bool { return x < 3 }).Results()
	if len(result) != 0 {
		t.Errorf("TakeWhile none match: got %v", result)
	}
}

func TestSkipWhile(t *testing.T) {
	result := From([]int{1, 2, 3, 4, 1}).SkipWhile(func(x int) bool { return x < 3 }).Results()
	if !reflect.DeepEqual(result, []int{3, 4, 1}) {
		t.Errorf("SkipWhile expected [3,4,1], got %v", result)
	}
	// 全部满足
	result = From([]int{1, 2, 3}).SkipWhile(func(x int) bool { return x < 10 }).Results()
	if len(result) != 0 {
		t.Errorf("SkipWhile all match: got %v", result)
	}
	// 首元素不满足
	result = From([]int{5, 1, 2}).SkipWhile(func(x int) bool { return x < 3 }).Results()
	if !reflect.DeepEqual(result, []int{5, 1, 2}) {
		t.Errorf("SkipWhile first not match: got %v", result)
	}
}

// ---------------------------------------------------------------------------
// Projection
// ---------------------------------------------------------------------------

func TestSelect(t *testing.T) {
	l := From([]int{1, 2, 3}).Select(func(x int) int { return x * x })
	if !reflect.DeepEqual(l.Results(), []int{1, 4, 9}) {
		t.Errorf("Select expected [1,4,9], got %v", l.Results())
	}
}

func TestSelectT(t *testing.T) {
	result := SelectT(From([]int{1, 2, 3}), func(x int) string {
		return string(rune('A' + x - 1))
	}).Results()
	if !reflect.DeepEqual(result, []string{"A", "B", "C"}) {
		t.Errorf("SelectT expected [A,B,C], got %v", result)
	}
}

func TestSelectMany(t *testing.T) {
	l := From([][]int{{1, 2}, {3, 4}})
	result := SelectMany(l, func(x []int) []int { return x }).Results()
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4}) {
		t.Errorf("SelectMany expected [1,2,3,4], got %v", result)
	}
}

// ---------------------------------------------------------------------------
// Sorting
// ---------------------------------------------------------------------------

func TestSort(t *testing.T) {
	l := From([]int{3, 1, 2}).Sort(func(a, b int) bool { return a < b })
	if !reflect.DeepEqual(l.Results(), []int{1, 2, 3}) {
		t.Errorf("Sort expected [1,2,3], got %v", l.Results())
	}
}

func TestSortStable(t *testing.T) {
	type pair struct{ x, y int }
	l := From([]pair{{1, 2}, {3, 1}, {2, 2}}).
		Sort(func(a, b pair) bool { return a.y < b.y }).
		Sort(func(a, b pair) bool { return a.x < b.x })
	// 二次排序后，x 相同的元素应保留首次排序的顺序
	result := l.Results()
	if result[0].x != 1 || result[0].y != 2 {
		t.Errorf("sort stability broke: %v", result)
	}
}

func TestWithComparer(t *testing.T) {
	compare := func(a, b int) int { return a - b }
	l := From([]int{1, 2}).WithComparer(compare)
	if l.compare == nil || l.compare(1, 2) != -1 || l.compare(2, 1) != 1 || l.compare(1, 1) != 0 {
		t.Errorf("WithComparer failed to set valid compare function")
	}
}

func TestOrderBy(t *testing.T) {
	result := OrderBy(From([]int{3, 1, 2}), func(x int) int { return x }).Results()
	if !reflect.DeepEqual(result, []int{1, 2, 3}) {
		t.Errorf("OrderBy ascending: got %v", result)
	}
}

func TestOrderByDescending(t *testing.T) {
	result := OrderByDescending(From([]int{1, 3, 2}), func(x int) int { return x }).Results()
	if !reflect.DeepEqual(result, []int{3, 2, 1}) {
		t.Errorf("OrderByDescending: got %v", result)
	}
}

func TestOrderByThenBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	people := From([]Person{
		{"Bob", 30}, {"Alice", 25}, {"Alice", 35}, {"Bob", 20},
	})
	result := OrderBy(people, func(p Person) string { return p.Name }).
		ThenBy(func(a, b Person) int { return cmp.Compare(a.Age, b.Age) }).
		Results()
	// 按 Name 升序，同 Name 按 Age 升序
	expected := []Person{{"Alice", 25}, {"Alice", 35}, {"Bob", 20}, {"Bob", 30}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("OrderBy + ThenBy: got %v", result)
	}
}

func TestThenByDoesNotMutatePreviousOrdering(t *testing.T) {
	type item struct{ group, value int }
	primary := OrderBy(
		From([]item{{1, 3}, {1, 1}, {1, 2}, {2, 0}}),
		func(x item) int { return x.group },
	)
	before := primary.ToSlice()
	secondary := primary.ThenBy(func(a, b item) int { return cmp.Compare(a.value, b.value) })

	if !reflect.DeepEqual(primary.Results(), before) {
		t.Fatalf("ThenBy mutated the previous ordering: got %v, want %v", primary.Results(), before)
	}
	if got := secondary.Results(); got[0].value != 1 || got[1].value != 2 || got[2].value != 3 {
		t.Fatalf("ThenBy result = %v, want values [1 2 3 0]", got)
	}
}

// ---------------------------------------------------------------------------
// Aggregation
// ---------------------------------------------------------------------------

func TestCountBy(t *testing.T) {
	count := From([]int{1, 2, 3, 4}).CountBy(func(x int) bool { return x%2 == 0 })
	if count != 2 {
		t.Errorf("CountBy even: expected 2, got %d", count)
	}
}

func TestSum(t *testing.T) {
	if got := Sum(From([]int{1, 2, 3})); got != 6 {
		t.Errorf("Sum expected 6, got %d", got)
	}
}

func TestAverage(t *testing.T) {
	avg := Average(From([]float64{1.0, 2.0, 3.0}))
	if avg != 2.0 {
		t.Errorf("Average expected 2.0, got %f", avg)
	}
	if Average(From([]float64{})) != 0 {
		t.Error("Average of empty should be 0")
	}
}

func TestNumericAggregatesShortCircuitErrors(t *testing.T) {
	errLinq := From([]int{1, 2, 3}).Distinct()
	if errLinq.Error() == nil {
		t.Fatal("expected Distinct error")
	}
	if got := Sum(errLinq); got != 0 {
		t.Fatalf("Sum on errored sequence = %d, want 0", got)
	}
	if got := Average(errLinq); got != 0 {
		t.Fatalf("Average on errored sequence = %v, want 0", got)
	}
}

func TestMin(t *testing.T) {
	l := From([]int{3, 1, 2}).WithComparer(func(a, b int) int { return a - b })
	minVal, ok := l.Min()
	if !ok || minVal != 1 {
		t.Errorf("Min should return (1, true), got (%d, %v)", minVal, ok)
	}

	// 未设置比较函数
	_, ok = From([]int{1, 2}).Min()
	if ok {
		t.Error("Min without comparer should return false")
	}

	// 空序列
	_, ok = From([]int{}).WithComparer(func(a, b int) int { return a - b }).Min()
	if ok {
		t.Error("Min on empty should return false")
	}
}

func TestMax(t *testing.T) {
	l := From([]int{3, 1, 2}).WithComparer(func(a, b int) int { return a - b })
	maxVal, ok := l.Max()
	if !ok || maxVal != 3 {
		t.Errorf("Max should return (3, true), got (%d, %v)", maxVal, ok)
	}

	// 空序列
	_, ok = From([]int{}).WithComparer(func(a, b int) int { return a - b }).Max()
	if ok {
		t.Error("Max on empty should return false")
	}
}

func TestMinBy(t *testing.T) {
	type item struct {
		name string
		val  int
	}
	l := From([]item{{"a", 3}, {"b", 1}, {"c", 2}})
	result, ok := MinBy(l, func(x item) int { return x.val })
	if !ok || result.name != "b" {
		t.Errorf("MinBy should return {b,1}, got %v, ok=%v", result, ok)
	}

	_, ok = MinBy(Empty[item](), func(x item) int { return x.val })
	if ok {
		t.Error("MinBy on empty should return false")
	}
}

func TestMaxBy(t *testing.T) {
	type item struct {
		name string
		val  int
	}
	l := From([]item{{"a", 3}, {"b", 1}, {"c", 2}})
	result, ok := MaxBy(l, func(x item) int { return x.val })
	if !ok || result.name != "a" {
		t.Errorf("MaxBy should return {a,3}, got %v, ok=%v", result, ok)
	}

	_, ok = MaxBy(Empty[item](), func(x item) int { return x.val })
	if ok {
		t.Error("MaxBy on empty should return false")
	}
}

func TestAggregate(t *testing.T) {
	result := Aggregate(From([]int{1, 2, 3, 4}), 0, func(acc, x int) int { return acc + x })
	if result != 10 {
		t.Errorf("Aggregate sum: expected 10, got %d", result)
	}

	strResult := Aggregate(From([]string{"a", "b", "c"}), "", func(acc, s string) string { return acc + s })
	if strResult != "abc" {
		t.Errorf("Aggregate concat: expected 'abc', got '%s'", strResult)
	}

	// 种子类型可不同
	result2 := Aggregate(From([]int{1, 2, 3}), 1.0, func(acc float64, x int) float64 { return acc * float64(x) })
	if result2 != 6.0 {
		t.Errorf("Aggregate product: expected 6.0, got %f", result2)
	}
}

// ---------------------------------------------------------------------------
// Element Access
// ---------------------------------------------------------------------------

func TestAny(t *testing.T) {
	l := From([]int{1, 2, 3})
	if !l.Any(func(x int) bool { return x > 1 }) {
		t.Error("Any should return true")
	}
	if l.Any(func(x int) bool { return x > 3 }) {
		t.Error("Any should return false")
	}
}

func TestAll(t *testing.T) {
	if !From([]int{2, 4, 6}).All(func(x int) bool { return x%2 == 0 }) {
		t.Error("All should return true for all even")
	}
	if From([]int{2, 3, 6}).All(func(x int) bool { return x%2 == 0 }) {
		t.Error("All should return false when not all match")
	}
}

func TestFirst(t *testing.T) {
	v, ok := From([]int{10, 20}).First()
	if !ok || v != 10 {
		t.Errorf("First: got (%d, %v)", v, ok)
	}
	_, ok = Empty[int]().First()
	if ok {
		t.Error("First on empty should return false")
	}
}

func TestFirstBy(t *testing.T) {
	v, ok := From([]int{1, 2, 3}).FirstBy(func(x int) bool { return x > 1 })
	if !ok || v != 2 {
		t.Errorf("FirstBy: got (%d, %v)", v, ok)
	}
	_, ok = From([]int{1, 2}).FirstBy(func(x int) bool { return x > 5 })
	if ok {
		t.Error("FirstBy no match should return false")
	}
}

func TestLast(t *testing.T) {
	v, ok := From([]int{1, 2, 3}).Last()
	if !ok || v != 3 {
		t.Errorf("Last: got (%d, %v)", v, ok)
	}
	_, ok = Empty[int]().Last()
	if ok {
		t.Error("Last on empty should return false")
	}
}

func TestLastBy(t *testing.T) {
	v, ok := From([]int{1, 2, 3, 4}).LastBy(func(x int) bool { return x%2 == 0 })
	if !ok || v != 4 {
		t.Errorf("LastBy: got (%d, %v)", v, ok)
	}
}

func TestElementAt(t *testing.T) {
	l := From([]int{10, 20, 30})
	v, ok := l.ElementAt(1)
	if !ok || v != 20 {
		t.Errorf("ElementAt(1): got (%d, %v)", v, ok)
	}
	_, ok = l.ElementAt(-1)
	if ok {
		t.Error("ElementAt(-1) should return false")
	}
	_, ok = l.ElementAt(3)
	if ok {
		t.Error("ElementAt(3) out of range should return false")
	}
}

func TestSingle(t *testing.T) {
	v, ok := From([]int{42}).Single()
	if !ok || v != 42 {
		t.Errorf("Single: got (%d, %v)", v, ok)
	}
	_, ok = Empty[int]().Single()
	if ok {
		t.Error("Single on empty should return false")
	}
	_, ok = From([]int{1, 2}).Single()
	if ok {
		t.Error("Single on multiple should return false")
	}
}

func TestSingleBy(t *testing.T) {
	v, ok := From([]int{1, 2, 3}).SingleBy(func(x int) bool { return x == 2 })
	if !ok || v != 2 {
		t.Errorf("SingleBy: got (%d, %v)", v, ok)
	}
	_, ok = From([]int{1, 2, 3}).SingleBy(func(x int) bool { return x > 0 })
	if ok {
		t.Error("SingleBy multiple matches should return false")
	}
	_, ok = From([]int{1, 2, 3}).SingleBy(func(x int) bool { return x > 5 })
	if ok {
		t.Error("SingleBy no match should return false")
	}
}

func TestContains(t *testing.T) {
	if !Contains(From([]int{1, 2, 3}), 2) {
		t.Error("Contains should return true")
	}
	if Contains(From([]int{1, 2, 3}), 4) {
		t.Error("Contains should return false")
	}
}

// ---------------------------------------------------------------------------
// Partitioning
// ---------------------------------------------------------------------------

func TestTake(t *testing.T) {
	l := From([]int{1, 2, 3, 4})
	if !reflect.DeepEqual(l.Take(2).Results(), []int{1, 2}) {
		t.Errorf("Take(2) failed")
	}
	if len(l.Take(0).Results()) != 0 {
		t.Error("Take(0) should return empty")
	}
}

func TestSkip(t *testing.T) {
	l := From([]int{1, 2, 3, 4})
	if !reflect.DeepEqual(l.Skip(2).Results(), []int{3, 4}) {
		t.Errorf("Skip(2) failed")
	}
	if len(l.Skip(4).Results()) != 0 {
		t.Error("Skip(4) should return empty")
	}
}

func TestChunk(t *testing.T) {
	chunks := From([]int{1, 2, 3, 4, 5}).Chunk(2)
	if len(chunks) != 3 {
		t.Fatalf("Chunk(2): expected 3 chunks, got %d", len(chunks))
	}
	if !reflect.DeepEqual(chunks[0], []int{1, 2}) {
		t.Errorf("chunk[0] = %v", chunks[0])
	}
	if !reflect.DeepEqual(chunks[2], []int{5}) {
		t.Errorf("last chunk = %v", chunks[2])
	}
	if len(From([]int{1, 2}).Chunk(0)) != 0 {
		t.Error("Chunk(0) should return empty")
	}
	maxInt := int(^uint(0) >> 1)
	if got := From([]int{1, 2}).Chunk(maxInt); !reflect.DeepEqual(got, [][]int{{1, 2}}) {
		t.Fatalf("Chunk(maxInt) = %v, want [[1 2]]", got)
	}
}

// ---------------------------------------------------------------------------
// Concatenation
// ---------------------------------------------------------------------------

func TestConcat(t *testing.T) {
	l1 := From([]int{1, 2})
	l2 := From([]int{3, 4})
	if !reflect.DeepEqual(l1.Concat(l2).Results(), []int{1, 2, 3, 4}) {
		t.Errorf("Concat failed")
	}
}

func TestAppendPrepend(t *testing.T) {
	result := From([]int{2, 3}).Append(4, 5).Results()
	if !reflect.DeepEqual(result, []int{2, 3, 4, 5}) {
		t.Errorf("Append: got %v", result)
	}
	result = From([]int{2, 3}).Prepend(0, 1).Results()
	if !reflect.DeepEqual(result, []int{0, 1, 2, 3}) {
		t.Errorf("Prepend: got %v", result)
	}
}

func TestReverse(t *testing.T) {
	l := From([]int{1, 2, 3})
	if !reflect.DeepEqual(l.Reverse().Results(), []int{3, 2, 1}) {
		t.Errorf("Reverse failed")
	}
}

// ---------------------------------------------------------------------------
// Set Operations
// ---------------------------------------------------------------------------

func TestUnion(t *testing.T) {
	result := Union(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4}) {
		t.Errorf("Union: got %v", result)
	}
}

func TestIntersect(t *testing.T) {
	result := Intersect(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	if !reflect.DeepEqual(result, []int{2, 3}) {
		t.Errorf("Intersect: got %v", result)
	}
}

func TestExcept(t *testing.T) {
	result := Except(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	if !reflect.DeepEqual(result, []int{1}) {
		t.Errorf("Except: got %v", result)
	}
}

func TestSequenceEqual(t *testing.T) {
	if !SequenceEqual(From([]int{1, 2, 3}), From([]int{1, 2, 3})) {
		t.Error("SequenceEqual should be true")
	}
	if SequenceEqual(From([]int{1, 2}), From([]int{1, 2, 3})) {
		t.Error("SequenceEqual different lengths should be false")
	}
	if SequenceEqual(From([]int{1, 2, 3}), From([]int{1, 2, 4})) {
		t.Error("SequenceEqual different values should be false")
	}
}

// ---------------------------------------------------------------------------
// Grouping & Joining
// ---------------------------------------------------------------------------

func TestGroupBy(t *testing.T) {
	l := From([]string{"apple", "banana", "apricot"})
	groups := GroupBy(l, func(s string) rune { r, _ := utf8.DecodeRuneInString(s); return r })
	if len(groups) != 2 {
		t.Fatalf("GroupBy should create 2 groups, got %d", len(groups))
	}
	groupMap := make(map[rune][]string, len(groups))
	for _, g := range groups {
		groupMap[g.Key] = g.Items
	}
	if !reflect.DeepEqual(groupMap['a'], []string{"apple", "apricot"}) {
		t.Errorf("group 'a': got %v", groupMap['a'])
	}
	if !reflect.DeepEqual(groupMap['b'], []string{"banana"}) {
		t.Errorf("group 'b': got %v", groupMap['b'])
	}
}

func TestJoin(t *testing.T) {
	type Order struct{ ID int }
	type Customer struct {
		OrderID int
		Name    string
	}
	outer := From([]Order{{1}, {2}})
	inner := From([]Customer{{1, "A"}, {2, "B"}})
	result := Join(outer, inner,
		func(o Order) int { return o.ID },
		func(c Customer) int { return c.OrderID },
		func(o Order, c Customer) string { return c.Name },
	).Results()
	if !reflect.DeepEqual(result, []string{"A", "B"}) {
		t.Errorf("Join: got %v", result)
	}
}

func TestLeftJoin(t *testing.T) {
	type Order struct{ ID int }
	type Detail struct {
		OrderID int
		Product string
	}
	outer := From([]Order{{1}, {2}, {3}})
	inner := From([]Detail{{1, "Apple"}, {2, "Banana"}})
	result := LeftJoin(outer, inner,
		func(o Order) int { return o.ID },
		func(d Detail) int { return d.OrderID },
		func(o Order, d Detail) string { return d.Product },
		func(o Order) string { return "N/A" },
	).Results()
	if !reflect.DeepEqual(result, []string{"Apple", "Banana", "N/A"}) {
		t.Errorf("LeftJoin: got %v", result)
	}
}

func TestZip(t *testing.T) {
	result := Zip(
		From([]int{1, 2, 3}),
		From([]string{"a", "b"}),
		func(x int, s string) string { return string(rune('0'+x)) + s },
	).Results()
	if !reflect.DeepEqual(result, []string{"1a", "2b"}) {
		t.Errorf("Zip: got %v", result)
	}
}

// ---------------------------------------------------------------------------
// Conditional / Execution
// ---------------------------------------------------------------------------

func TestDefaultIfEmpty(t *testing.T) {
	result := From([]int{}).DefaultIfEmpty(42).Results()
	if !reflect.DeepEqual(result, []int{42}) {
		t.Errorf("DefaultIfEmpty on empty: got %v", result)
	}
	result = From([]int{1}).DefaultIfEmpty(42).Results()
	if !reflect.DeepEqual(result, []int{1}) {
		t.Errorf("DefaultIfEmpty on non-empty: got %v", result)
	}
}

func TestToSlice(t *testing.T) {
	l := From([]int{1, 2, 3})
	s := l.ToSlice()
	s[0] = 99
	if l.Results()[0] == 99 {
		t.Error("ToSlice should return independent copy")
	}
}

func TestToMap(t *testing.T) {
	type Person struct {
		ID   int
		Name string
	}
	people := From([]Person{{1, "Alice"}, {2, "Bob"}})
	m := ToMap(people, func(p Person) int { return p.ID }, func(p Person) string { return p.Name })
	if len(m) != 2 || m[1] != "Alice" || m[2] != "Bob" {
		t.Errorf("ToMap: got %v", m)
	}
	// 存在错误时返回 nil
	errLinq := From([]Person{{1, "A"}}).Distinct()
	if ToMap(errLinq, func(p Person) int { return p.ID }, func(p Person) string { return p.Name }) != nil {
		t.Error("ToMap on errored input should return nil")
	}
}

// ---------------------------------------------------------------------------
// Error Propagation
// ---------------------------------------------------------------------------

func TestErrorPropagation_Methods(t *testing.T) {
	// 构造一个携带错误的 Linq，然后链式调用应短路
	errLinq := From([]int{1, 2}).Distinct() // 无 comparer，携带错误
	if errLinq.Error() == nil {
		t.Fatal("expected error on Distinct without comparer")
	}
	// 后续操作应保留错误和数据
	next := errLinq.Where(func(x int) bool { return true })
	if next.Error() == nil {
		t.Error("error should propagate through Where")
	}
}

func TestErrorPropagation_Standalone(t *testing.T) {
	errLinq := From([]int{1, 2}).Distinct()
	if errLinq.Error() == nil {
		t.Fatal("expected error")
	}
	// SelectT 应传播错误
	result := SelectT(errLinq, func(x int) string { return "" })
	if result.Error() == nil {
		t.Error("SelectT should propagate error")
	}
	// SelectMany 应传播错误
	result2 := SelectMany(errLinq, func(x int) []int { return nil })
	if result2.Error() == nil {
		t.Error("SelectMany should propagate error")
	}
	// Union 应传播错误
	result3 := Union(errLinq, From([]int{}))
	if result3.Error() == nil {
		t.Error("Union should propagate error")
	}
	// Join 应传播错误
	result4 := Join(errLinq, From([]int{}), func(x int) int { return x }, func(x int) int { return x }, func(a, b int) int { return a })
	if result4.Error() == nil {
		t.Error("Join should propagate error")
	}
}

func TestErrorPropagation_SequenceEqual(t *testing.T) {
	errLinq := From([]int{1, 2}).Distinct()
	if SequenceEqual(errLinq, From([]int{1, 2})) {
		t.Error("SequenceEqual with errored l1 should return false")
	}
	if SequenceEqual(From([]int{1, 2}), errLinq) {
		t.Error("SequenceEqual with errored l2 should return false")
	}
}

func TestErrorPropagation_ForEach(t *testing.T) {
	errLinq := From([]int{1, 2}).Distinct()
	called := false
	errLinq.ForEach(func(x int) { called = true })
	if called {
		t.Error("ForEach on errored input should not execute action")
	}
}

func TestSetOperationsPreserveComparer(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	l1 := From([]int{3, 1, 2}).WithComparer(cmp)
	l2 := From([]int{4, 5})
	// Union、Intersect、Except 应保留 l1 的比较器
	for name, l := range map[string]Linq[int]{
		"Union":     Union(l1, l2),
		"Intersect": Intersect(l1, l2),
		"Except":    Except(l1, l2),
	} {
		if l.compare == nil {
			t.Errorf("%s should preserve l1.compare", name)
		}
	}
}

// ---------------------------------------------------------------------------
// MinVal / MaxVal
// ---------------------------------------------------------------------------

func TestMinVal(t *testing.T) {
	v, ok := MinVal(From([]int{3, 1, 4, 1, 5}))
	if !ok || v != 1 {
		t.Errorf("MinVal expected 1, got %v (ok=%v)", v, ok)
	}
	_, ok = MinVal(Empty[int]())
	if ok {
		t.Error("MinVal on empty should return false")
	}
}

func TestMaxVal(t *testing.T) {
	v, ok := MaxVal(From([]float64{1.0, 3.5, 2.2}))
	if !ok || v != 3.5 {
		t.Errorf("MaxVal expected 3.5, got %v (ok=%v)", v, ok)
	}
	_, ok = MaxVal(Empty[float64]())
	if ok {
		t.Error("MaxVal on empty should return false")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func stringsToLower(s string) string {
	lower := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		lower[i] = c
	}
	return string(lower)
}

func TestThenByDescending(t *testing.T) {
	type item struct{ a, b int }
	data := From([]item{{1, 3}, {1, 1}, {1, 2}, {2, 0}})
	ol := OrderBy(data, func(x item) int { return x.a })
	result := ol.ThenByDescending(func(a, b item) int { return a.b - b.b }).Results()
	if len(result) != 4 {
		t.Fatalf("expected 4 results, got %d", len(result))
	}
	if result[0].b != 3 || result[1].b != 2 || result[2].b != 1 {
		t.Errorf("ThenByDescending: expected b=[3,2,1,0], got %v", result)
	}
}

func TestThenByDescendingHandlesMinIntComparer(t *testing.T) {
	minInt := -int(^uint(0)>>1) - 1
	compare := func(a, b int) int {
		switch {
		case a < b:
			return minInt
		case a > b:
			return 1
		default:
			return 0
		}
	}
	ol := OrderBy(From([]int{1, 2}), func(int) int { return 0 })
	if got := ol.ThenByDescending(compare).Results(); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Fatalf("ThenByDescending = %v, want [2 1]", got)
	}
}

func ExampleFrom() {
	nums := From([]int{1, 2, 3, 4, 5}).
		Where(func(x int) bool { return x%2 == 0 }).
		Select(func(x int) int { return x * 2 })
	fmt.Println(nums.Results())
	// Output: [4 8]
}

func ExampleOrderBy() {
	type Person struct {
		Name string
		Age  int
	}
	people := From([]Person{{"Bob", 30}, {"Alice", 25}})
	sorted := OrderBy(people, func(p Person) int { return p.Age })
	for _, p := range sorted.Results() {
		fmt.Println(p.Name)
	}
	// Output:
	// Alice
	// Bob
}

func ExampleRange() {
	l := Range(0, 5)
	fmt.Println(l.Results())
	// Output: [0 1 2 3 4]
}

func ExampleGroupBy() {
	nums := From([]int{1, 2, 3, 4, 5})
	groups := GroupBy(nums, func(x int) int { return x % 2 })
	for _, g := range groups {
		fmt.Printf("%d: %v\n", g.Key, g.Items)
	}
	// Output:
	// 1: [1 3 5]
	// 0: [2 4]
}
