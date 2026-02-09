package linq

import (
	"reflect"
	"strings"
	"testing"
)

// TestFrom 测试创建 Linq 实例
func TestFrom(t *testing.T) {
	data := []int{1, 2, 3}
	l := From(data)
	if len(l.data) != 3 || l.compare != nil {
		t.Errorf("From should create Linq with correct data and nil comparer, got %v", l)
	}
}

// TestWhere 测试过滤功能
func TestWhere(t *testing.T) {
	l := From([]int{1, 2, 3, 4}).Where(func(x int) bool { return x%2 == 0 })
	expected := []int{2, 4}
	if !reflect.DeepEqual(l.Results(), expected) {
		t.Errorf("Where should filter even numbers, expected %v, got %v", expected, l.Results())
	}
}

// TestSelect 测试投影功能
func TestSelect(t *testing.T) {
	l := From([]int{1, 2, 3}).Select(func(x int) int { return x * x })
	expected := []int{1, 4, 9}
	if !reflect.DeepEqual(l.Results(), expected) {
		t.Errorf("Select should square numbers, expected %v, got %v", expected, l.Results())
	}
}

// TestSort 测试排序功能
func TestSort(t *testing.T) {
	l := From([]int{3, 1, 2}).Sort(func(a, b int) bool { return a < b })
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(l.Results(), expected) {
		t.Errorf("Sort should order ascending, expected %v, got %v", expected, l.Results())
	}
}

// TestWithComparer 测试设置比较函数
func TestWithComparer(t *testing.T) {
	compare := func(a, b int) int { return a - b }
	l := From([]int{1, 2}).WithComparer(compare)

	// 验证比较函数是否设置正确
	if l.compare == nil || l.compare(1, 2) != -1 || l.compare(2, 1) != 1 || l.compare(1, 1) != 0 {
		t.Errorf("WithComparer failed to set valid compare function")
	}
}

// TestAny 测试是否存在满足条件的元素
func TestAny(t *testing.T) {
	l := From([]int{1, 2, 3})
	if !l.Any(func(x int) bool { return x > 1 }) {
		t.Error("Any should return true for numbers > 1")
	}
	if l.Any(func(x int) bool { return x > 3 }) {
		t.Error("Any should return false for numbers > 3")
	}
}

// TestDistinctComparable 测试 comparable 类型的去重
func TestDistinctComparable(t *testing.T) {
	l := From([]int{1, 2, 2, 3, 1})
	result := DistinctComparable(l).Results()
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("DistinctComparable should remove duplicates, expected %v, got %v", expected, result)
	}
}

// TestDistinct 测试自定义比较函数去重
func TestDistinct(t *testing.T) {
	// 使用大小写不敏感的比较器
	l := From([]string{"a", "A", "b"}).WithComparer(func(a, b string) int {
		aLower, bLower := strings.ToLower(a), strings.ToLower(b)
		if aLower < bLower {
			return -1
		} else if aLower > bLower {
			return 1
		}
		return 0
	})
	result := l.Distinct().Results()
	// 去重后应该有 2 个元素，a/A 算作同一个
	if len(result) != 2 {
		t.Errorf("Distinct should return 2 elements, got %d: %v", len(result), result)
	}
	// 验证包含 b
	hasB := false
	hasAorA := false
	for _, v := range result {
		if v == "b" {
			hasB = true
		}
		if v == "a" || v == "A" {
			hasAorA = true
		}
	}
	if !hasB || !hasAorA {
		t.Errorf("Distinct should contain 'b' and either 'a' or 'A', got %v", result)
	}

	// 测试未设置比较函数时返回错误
	l2 := From([]int{1, 2, 2})
	result2 := l2.Distinct()
	if result2.Error() == nil {
		t.Error("Distinct without comparer should set error")
	}
}

// TestTake 测试获取前 n 个元素
func TestTake(t *testing.T) {
	l := From([]int{1, 2, 3, 4})
	result := l.Take(2).Results()
	expected := []int{1, 2}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Take(2) should return first 2 elements, expected %v, got %v", expected, result)
	}

	if len(l.Take(0).Results()) != 0 {
		t.Error("Take(0) should return empty slice")
	}
}

// TestSkip 测试跳过前 n 个元素
func TestSkip(t *testing.T) {
	l := From([]int{1, 2, 3, 4})
	result := l.Skip(2).Results()
	expected := []int{3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Skip(2) should skip first 2 elements, expected %v, got %v", expected, result)
	}

	if len(l.Skip(4).Results()) != 0 {
		t.Error("Skip(4) should return empty slice for full length")
	}
}

// TestGroupBy 测试分组功能
func TestGroupBy(t *testing.T) {
	l := From([]string{"apple", "banana", "apricot"})
	groups := GroupBy(l, func(s string) rune { return []rune(s)[0] })
	if len(groups) != 2 {
		t.Fatalf("GroupBy should create 2 groups, got %d", len(groups))
	}

	// 不依赖顺序，按键查找分组
	groupMap := make(map[rune][]string, len(groups))
	for _, g := range groups {
		groupMap[g.Key] = g.Items
	}

	if !reflect.DeepEqual(groupMap['a'], []string{"apple", "apricot"}) {
		t.Errorf("group 'a' mismatch, got %v", groupMap['a'])
	}
	if !reflect.DeepEqual(groupMap['b'], []string{"banana"}) {
		t.Errorf("group 'b' mismatch, got %v", groupMap['b'])
	}
}

// TestJoin 测试连接功能
func TestJoin(t *testing.T) {
	outer := From([]struct{ ID int }{{1}, {2}})
	inner := From([]struct {
		OrderID int
		Name    string
	}{{1, "A"}, {2, "B"}})
	result := Join(outer, inner,
		func(o struct{ ID int }) int { return o.ID },
		func(i struct {
			OrderID int
			Name    string
		}) int {
			return i.OrderID
		},
		func(o struct{ ID int }, i struct {
			OrderID int
			Name    string
		}) string {
			return i.Name
		}).Results()
	expected := []string{"A", "B"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Join should match IDs, expected %v, got %v", expected, result)
	}
}

// TestConcat 测试合并功能
func TestConcat(t *testing.T) {
	l1 := From([]int{1, 2})
	l2 := From([]int{3, 4})
	result := l1.Concat(l2).Results()
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Concat should merge slices, expected %v, got %v", expected, result)
	}
}

// TestReverse 测试反转功能
func TestReverse(t *testing.T) {
	l := From([]int{1, 2, 3})
	result := l.Reverse().Results()
	expected := []int{3, 2, 1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Reverse should invert order, expected %v, got %v", expected, result)
	}
}

// TestMin 测试最小值功能
func TestMin(t *testing.T) {
	l := From([]int{3, 1, 2}).WithComparer(func(a, b int) int { return a - b })
	min, ok := l.Min()
	if !ok || min != 1 {
		t.Errorf("Min should return 1, got %d, ok=%v", min, ok)
	}

	// 测试未设置比较函数时返回错误
	l2 := From([]int{1, 2})
	_, ok = l2.Min()
	if ok {
		t.Error("Min without comparer should return false")
	}
}

// TestMax 测试最大值功能
func TestMax(t *testing.T) {
	l := From([]int{3, 1, 2}).WithComparer(func(a, b int) int { return a - b })
	max, ok := l.Max()
	if !ok || max != 3 {
		t.Errorf("Max should return 3, got %d, ok=%v", max, ok)
	}

	// 测试空切片
	l2 := From([]int{}).WithComparer(func(a, b int) int { return a - b })
	_, ok = l2.Max()
	if ok {
		t.Error("Max on empty slice should return false")
	}
}

// assertPanics 检查函数是否引发 panic
func assertPanics(t *testing.T, msg string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: didn't panic as expected", msg)
		}
	}()
	f()
}

// TestSkipWhile 测试跳过开头满足条件的元素
func TestSkipWhile(t *testing.T) {
	// 基本用例：跳过前面小于 3 的元素
	result := From([]int{1, 2, 3, 4, 1}).SkipWhile(func(x int) bool { return x < 3 }).Results()
	expected := []int{3, 4, 1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("SkipWhile expected %v, got %v", expected, result)
	}

	// 所有元素满足条件：返回空
	result = From([]int{1, 2, 3}).SkipWhile(func(x int) bool { return x < 10 }).Results()
	if len(result) != 0 {
		t.Errorf("SkipWhile all match should return empty, got %v", result)
	}

	// 首元素不满足：返回全部
	result = From([]int{5, 1, 2}).SkipWhile(func(x int) bool { return x < 3 }).Results()
	expected = []int{5, 1, 2}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("SkipWhile first not match should return all, got %v", result)
	}

	// 空序列
	result = From([]int{}).SkipWhile(func(x int) bool { return true }).Results()
	if len(result) != 0 {
		t.Errorf("SkipWhile on empty should return empty, got %v", result)
	}
}

// TestTakeWhile 测试获取开头满足条件的元素
func TestTakeWhile(t *testing.T) {
	result := From([]int{1, 2, 3, 4}).TakeWhile(func(x int) bool { return x < 3 }).Results()
	expected := []int{1, 2}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TakeWhile expected %v, got %v", expected, result)
	}

	// 所有满足
	result = From([]int{1, 2}).TakeWhile(func(x int) bool { return x < 10 }).Results()
	if !reflect.DeepEqual(result, []int{1, 2}) {
		t.Errorf("TakeWhile all match should return all, got %v", result)
	}

	// 首元素不满足
	result = From([]int{5, 1, 2}).TakeWhile(func(x int) bool { return x < 3 }).Results()
	if len(result) != 0 {
		t.Errorf("TakeWhile first not match should return empty, got %v", result)
	}
}

// TestChunk 测试分块功能
func TestChunk(t *testing.T) {
	chunks := From([]int{1, 2, 3, 4, 5}).Chunk(2)
	if len(chunks) != 3 {
		t.Fatalf("Chunk(2) for 5 elements should produce 3 chunks, got %d", len(chunks))
	}
	if !reflect.DeepEqual(chunks[0], []int{1, 2}) {
		t.Errorf("chunk[0] got %v", chunks[0])
	}
	if !reflect.DeepEqual(chunks[2], []int{5}) {
		t.Errorf("last chunk got %v", chunks[2])
	}

	// size <= 0 返回空
	if len(From([]int{1, 2}).Chunk(0)) != 0 {
		t.Error("Chunk(0) should return empty")
	}
}

// TestSelectMany 测试扁平化映射
func TestSelectMany(t *testing.T) {
	l := From([][]int{{1, 2}, {3, 4}})
	result := SelectMany(l, func(x []int) []int { return x }).Results()
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("SelectMany expected %v, got %v", expected, result)
	}
}

// TestUnion 测试并集
func TestUnion(t *testing.T) {
	result := Union(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Union expected %v, got %v", expected, result)
	}
}

// TestIntersect 测试交集
func TestIntersect(t *testing.T) {
	result := Intersect(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	expected := []int{2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Intersect expected %v, got %v", expected, result)
	}
}

// TestExcept 测试差集
func TestExcept(t *testing.T) {
	result := Except(From([]int{1, 2, 3}), From([]int{2, 3, 4})).Results()
	expected := []int{1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Except expected %v, got %v", expected, result)
	}
}

// TestContains 测试包含判断
func TestContains(t *testing.T) {
	l := From([]int{1, 2, 3})
	if !Contains(l, 2) {
		t.Error("Contains should return true for existing element")
	}
	if Contains(l, 4) {
		t.Error("Contains should return false for non-existing element")
	}
}

// TestDefaultIfEmpty 测试空序列默认值
func TestDefaultIfEmpty(t *testing.T) {
	result := From([]int{}).DefaultIfEmpty(42).Results()
	if !reflect.DeepEqual(result, []int{42}) {
		t.Errorf("DefaultIfEmpty on empty should return [42], got %v", result)
	}
	result = From([]int{1}).DefaultIfEmpty(42).Results()
	if !reflect.DeepEqual(result, []int{1}) {
		t.Errorf("DefaultIfEmpty on non-empty should return original, got %v", result)
	}
}

// TestSum 测试求和
func TestSum(t *testing.T) {
	if got := Sum(From([]int{1, 2, 3})); got != 6 {
		t.Errorf("Sum expected 6, got %d", got)
	}
}

// TestAverage 测试平均值
func TestAverage(t *testing.T) {
	avg := Average(From([]float64{1.0, 2.0, 3.0}))
	if avg != 2.0 {
		t.Errorf("Average expected 2.0, got %f", avg)
	}
	if Average(From([]float64{})) != 0 {
		t.Error("Average of empty should be 0")
	}
}

// TestToSlice 测试安全副本
func TestToSlice(t *testing.T) {
	l := From([]int{1, 2, 3})
	s := l.ToSlice()
	s[0] = 99
	if l.Results()[0] == 99 {
		t.Error("ToSlice should return independent copy")
	}
}

// TestLast 测试获取最后一个元素
func TestLast(t *testing.T) {
	v, ok := From([]int{1, 2, 3}).Last()
	if !ok || v != 3 {
		t.Errorf("Last expected (3, true), got (%d, %v)", v, ok)
	}
	_, ok = From([]int{}).Last()
	if ok {
		t.Error("Last on empty should return false")
	}
}

// TestAll 测试全部满足条件
func TestAll(t *testing.T) {
	if !From([]int{2, 4, 6}).All(func(x int) bool { return x%2 == 0 }) {
		t.Error("All should return true for all even numbers")
	}
	if From([]int{2, 3, 6}).All(func(x int) bool { return x%2 == 0 }) {
		t.Error("All should return false when not all match")
	}
}

// TestElementAt 测试按索引获取
func TestElementAt(t *testing.T) {
	l := From([]int{10, 20, 30})
	v, ok := l.ElementAt(1)
	if !ok || v != 20 {
		t.Errorf("ElementAt(1) expected (20, true), got (%d, %v)", v, ok)
	}
	_, ok = l.ElementAt(-1)
	if ok {
		t.Error("ElementAt(-1) should return false")
	}
	_, ok = l.ElementAt(3)
	if ok {
		t.Error("ElementAt(3) should return false for out of range")
	}
}

// TestAppendPrepend 测试添加元素
func TestAppendPrepend(t *testing.T) {
	result := From([]int{2, 3}).Append(4, 5).Results()
	if !reflect.DeepEqual(result, []int{2, 3, 4, 5}) {
		t.Errorf("Append got %v", result)
	}
	result = From([]int{2, 3}).Prepend(0, 1).Results()
	if !reflect.DeepEqual(result, []int{0, 1, 2, 3}) {
		t.Errorf("Prepend got %v", result)
	}
}
