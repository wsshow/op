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

// TestWithComparer 测试设置比较函数// TestWithComparer 测试设置比较函数
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

	// 测试未设置比较函数时 panic
	l2 := From([]int{1, 2, 2})
	assertPanics(t, "Distinct without comparer should panic", func() {
		l2.Distinct()
	})
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
	expected := []Group[rune, string]{
		{Key: 'a', Items: []string{"apple", "apricot"}},
		{Key: 'b', Items: []string{"banana"}},
	}
	if len(groups) != len(expected) {
		t.Errorf("GroupBy should create 2 groups, got %d", len(groups))
	}
	for i, g := range groups {
		if g.Key != expected[i].Key || !reflect.DeepEqual(g.Items, expected[i].Items) {
			t.Errorf("GroupBy result mismatch, expected %v, got %v", expected[i], g)
		}
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

	// 测试未设置比较函数时 panic
	l2 := From([]int{1, 2})
	assertPanics(t, "Min without comparer should panic", func() {
		l2.Min()
	})
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
