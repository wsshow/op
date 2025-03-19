package slice

import (
	"reflect"
	"testing"
)

// TestNew 测试创建 Slice 实例
func TestNew(t *testing.T) {
	s := New(1, 2, 3)
	if s.Length() != 3 {
		t.Errorf("New should create Slice with length 3, got %d", s.Length())
	}
	if s.IsEmpty() {
		t.Error("Newly created Slice should not be empty")
	}
}

// TestPush 测试添加元素
func TestPush(t *testing.T) {
	s := New[int]().Push(1, 2, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Push should add elements, expected %v, got %v", expected, s.Data())
	}
}

// TestPop 测试移除最后一个元素
func TestPop(t *testing.T) {
	s := New(1, 2, 3)
	last := s.Pop()
	if last != 3 || s.Length() != 2 {
		t.Errorf("Pop should remove and return 3, got %d, length %d", last, s.Length())
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 2}) {
		t.Errorf("Pop should leave [1, 2], got %v", s.Data())
	}

	// 测试空切片
	s = New[int]()
	if pop := s.Pop(); pop != 0 {
		t.Errorf("Pop on empty Slice should return zero value, got %d", pop)
	}
}

// TestShift 测试移除第一个元素
func TestShift(t *testing.T) {
	s := New(1, 2, 3)
	first := s.Shift()
	if first != 1 || s.Length() != 2 {
		t.Errorf("Shift should remove and return 1, got %d, length %d", first, s.Length())
	}
	if !reflect.DeepEqual(s.Data(), []int{2, 3}) {
		t.Errorf("Shift should leave [2, 3], got %v", s.Data())
	}

	// 测试空切片
	s = New[int]()
	if shift := s.Shift(); shift != 0 {
		t.Errorf("Shift on empty Slice should return zero value, got %d", shift)
	}
}

// TestUnshift 测试在开头添加元素
func TestUnshift(t *testing.T) {
	s := New(2, 3).Unshift(1)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Unshift should prepend 1, expected %v, got %v", expected, s.Data())
	}
}

// TestLength 测试长度计算
func TestLength(t *testing.T) {
	s := New(1, 2, 3)
	if s.Length() != 3 {
		t.Errorf("Length should be 3, got %d", s.Length())
	}
	s = New[int]()
	if s.Length() != 0 {
		t.Errorf("Empty Slice length should be 0, got %d", s.Length())
	}
}

// TestIsEmpty 测试空切片检查
func TestIsEmpty(t *testing.T) {
	s := New[int]()
	if !s.IsEmpty() {
		t.Error("New empty Slice should be empty")
	}
	s.Push(1)
	if s.IsEmpty() {
		t.Error("Slice with elements should not be empty")
	}
}

// TestForeach 测试遍历功能
func TestForeach(t *testing.T) {
	s := New(1, 2, 3)
	sum := 0
	s.Foreach(func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("Foreach should sum to 6, got %d", sum)
	}
}

// TestMap 测试映射功能
func TestMap(t *testing.T) {
	s := New(1, 2, 3).Map(func(v int) int { return v * 2 })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Map should double values, expected %v, got %v", expected, s.Data())
	}
}

// TestFilter 测试过滤功能
func TestFilter(t *testing.T) {
	s := New(1, 2, 3, 4)
	result := s.Filter(func(v int) bool { return v%2 == 0 })
	expected := []int{2, 4}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Filter should return even numbers, expected %v, got %v", expected, result.Data())
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 2, 3, 4}) {
		t.Errorf("Filter should not modify original, expected %v, got %v", []int{1, 2, 3, 4}, s.Data())
	}
}

// TestFind 测试查找功能
func TestFind(t *testing.T) {
	s := New(1, 2, 3)
	if val, ok := s.Find(func(v int) bool { return v > 1 }); !ok || val != 2 {
		t.Errorf("Find should return 2, got %d, ok=%v", val, ok)
	}
	if _, ok := s.Find(func(v int) bool { return v > 3 }); ok {
		t.Error("Find should return false for no match")
	}
}

// TestIndexOf 测试索引查找
func TestIndexOf(t *testing.T) {
	s := New(1, 2, 3)
	if idx := IndexOf(s, 2); idx != 1 {
		t.Errorf("IndexOf 2 should return 1, got %d", idx)
	}
	if idx := IndexOf(s, 4); idx != -1 {
		t.Errorf("IndexOf 4 should return -1, got %d", idx)
	}
}

// TestEvery 测试所有元素满足条件
func TestEvery(t *testing.T) {
	s := New(2, 4, 6)
	if !s.Every(func(v int) bool { return v%2 == 0 }) {
		t.Error("Every should return true for all even numbers")
	}
	s.Push(7)
	if s.Every(func(v int) bool { return v%2 == 0 }) {
		t.Error("Every should return false with odd number")
	}
}

// TestSome 测试存在满足条件的元素
func TestSome(t *testing.T) {
	s := New(1, 3, 4)
	if !s.Some(func(v int) bool { return v%2 == 0 }) {
		t.Error("Some should return true for even number")
	}
	if s.Some(func(v int) bool { return v > 5 }) {
		t.Error("Some should return false for no numbers > 5")
	}
}

// TestReduce 测试归约功能
func TestReduce(t *testing.T) {
	s := New(1, 2, 3)
	sum := s.Reduce(func(acc, v int) int { return acc + v }, 0)
	if sum != 6 {
		t.Errorf("Reduce should sum to 6, got %d", sum)
	}
}

// TestSort 测试排序功能
func TestSort(t *testing.T) {
	s := New(3, 1, 2).Sort(func(a, b int) bool { return a < b })
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Sort should order ascending, expected %v, got %v", expected, s.Data())
	}
}

// TestReverse 测试反转功能
func TestReverse(t *testing.T) {
	s := New(1, 2, 3).Reverse()
	expected := []int{3, 2, 1}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Reverse should invert order, expected %v, got %v", expected, s.Data())
	}
}

// TestConcat 测试合并功能
func TestConcat(t *testing.T) {
	s1 := New(1, 2)
	s2 := New(3, 4)
	result := s1.Concat(s2)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Concat should merge slices, expected %v, got %v", expected, result.Data())
	}
}

// TestSlice 测试提取子集
func TestSlice(t *testing.T) {
	s := New(1, 2, 3, 4)
	result := s.Slice(1, 3)
	expected := []int{2, 3}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Slice(1, 3) should return [2, 3], expected %v, got %v", expected, result.Data())
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 2, 3, 4}) {
		t.Errorf("Slice should not modify original, expected %v, got %v", []int{1, 2, 3, 4}, s.Data())
	}
}

// TestGet 测试获取元素
func TestGet(t *testing.T) {
	s := New(1, 2, 3)
	if val, ok := s.Get(1); !ok || val != 2 {
		t.Errorf("Get(1) should return 2, got %d, ok=%v", val, ok)
	}
	if _, ok := s.Get(3); ok {
		t.Error("Get(3) should return false for out of bounds")
	}
}

// TestSet 测试设置元素
func TestSet(t *testing.T) {
	s := New(1, 2, 3)
	if !s.Set(1, 10) {
		t.Error("Set(1, 10) should succeed")
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 10, 3}) {
		t.Errorf("Set should update value, expected [1, 10, 3], got %v", s.Data())
	}
	if s.Set(3, 4) {
		t.Error("Set(3, 4) should fail for out of bounds")
	}
}

// TestData 测试获取数据副本
func TestData(t *testing.T) {
	s := New(1, 2, 3)
	data := s.Data()
	data[0] = 10
	if !reflect.DeepEqual(s.Data(), []int{1, 2, 3}) {
		t.Errorf("Data should return copy, original should remain [1, 2, 3], got %v", s.Data())
	}
}
