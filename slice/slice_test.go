package slice

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	s := New(1, 2, 3)
	if s.Length() != 3 {
		t.Errorf("New should create Slice with length 3, got %d", s.Length())
	}
	if s.IsEmpty() {
		t.Error("Newly created Slice should not be empty")
	}
}

func TestNewEmpty(t *testing.T) {
	s := New[int]()
	if s.Length() != 0 {
		t.Errorf("New[int]() should have length 0, got %d", s.Length())
	}
	if !s.IsEmpty() {
		t.Error("New[int]() should be empty")
	}
}

func TestPush(t *testing.T) {
	s := New[int]().Push(1, 2, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Push should add elements, expected %v, got %v", expected, s.Data())
	}
}

func TestPop(t *testing.T) {
	s := New(1, 2, 3)
	last, ok := s.Pop()
	if !ok || last != 3 || s.Length() != 2 {
		t.Errorf("Pop should return (3, true), got (%d, %v), length %d", last, ok, s.Length())
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 2}) {
		t.Errorf("Pop should leave [1, 2], got %v", s.Data())
	}
}

func TestPopEmpty(t *testing.T) {
	s := New[int]()
	val, ok := s.Pop()
	if ok {
		t.Error("Pop on empty Slice should return ok=false")
	}
	if val != 0 {
		t.Errorf("Pop on empty Slice should return zero value, got %d", val)
	}
}

func TestPopZeroValue(t *testing.T) {
	s := New(0, 1)
	val, ok := s.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop should return (1, true), got (%d, %v)", val, ok)
	}
	val, ok = s.Pop()
	if !ok || val != 0 {
		t.Errorf("Pop should return (0, true) for zero value element, got (%d, %v)", val, ok)
	}
}

func TestShift(t *testing.T) {
	s := New(1, 2, 3)
	first, ok := s.Shift()
	if !ok || first != 1 || s.Length() != 2 {
		t.Errorf("Shift should return (1, true), got (%d, %v), length %d", first, ok, s.Length())
	}
	if !reflect.DeepEqual(s.Data(), []int{2, 3}) {
		t.Errorf("Shift should leave [2, 3], got %v", s.Data())
	}
}

func TestShiftEmpty(t *testing.T) {
	s := New[int]()
	val, ok := s.Shift()
	if ok {
		t.Error("Shift on empty Slice should return ok=false")
	}
	if val != 0 {
		t.Errorf("Shift on empty Slice should return zero value, got %d", val)
	}
}

func TestUnshift(t *testing.T) {
	s := New(2, 3).Unshift(1)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Unshift should prepend 1, expected %v, got %v", expected, s.Data())
	}
}

func TestUnshiftMultiple(t *testing.T) {
	s := New(3).Unshift(1, 2)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Unshift should prepend 1,2, expected %v, got %v", expected, s.Data())
	}
}

func TestInsert(t *testing.T) {
	s := New(1, 3).Insert(1, 2)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert(1, 2) expected %v, got %v", expected, s.Data())
	}
}

func TestInsertBeginning(t *testing.T) {
	s := New(2, 3).Insert(0, 1)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert(0, 1) expected %v, got %v", expected, s.Data())
	}
}

func TestInsertEnd(t *testing.T) {
	s := New(1, 2).Insert(2, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert(2, 3) expected %v, got %v", expected, s.Data())
	}
}

func TestInsertClampNegative(t *testing.T) {
	s := New(2, 3).Insert(-1, 1)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert(-1, 1) should clamp to 0, expected %v, got %v", expected, s.Data())
	}
}

func TestInsertClampOverflow(t *testing.T) {
	s := New(1, 2).Insert(10, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert(10, 3) should clamp to end, expected %v, got %v", expected, s.Data())
	}
}

func TestInsertEmptyValues(t *testing.T) {
	s := New(1, 2)
	s.Insert(1)
	expected := []int{1, 2}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Insert with no values should be no-op, expected %v, got %v", expected, s.Data())
	}
}

func TestRemove(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Remove(1)
	if !ok || val != 2 || s.Length() != 2 {
		t.Errorf("Remove(1) should return (2, true), got (%d, %v), length %d", val, ok, s.Length())
	}
	expected := []int{1, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Remove(1) should leave %v, got %v", expected, s.Data())
	}
}

func TestRemoveOutOfBounds(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Remove(5)
	if ok {
		t.Error("Remove(5) should return ok=false")
	}
	if val != 0 {
		t.Errorf("Remove(5) should return zero, got %d", val)
	}
}

func TestRemoveNegative(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Remove(-1)
	if ok {
		t.Error("Remove(-1) should return ok=false")
	}
	if val != 0 {
		t.Errorf("Remove(-1) should return zero, got %d", val)
	}
}

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

func TestCap(t *testing.T) {
	s := New[int]()
	if s.Cap() < 0 {
		t.Errorf("Cap should be >= 0, got %d", s.Cap())
	}
	s2 := New(1, 2, 3)
	if s2.Cap() < 3 {
		t.Errorf("Cap should be >= 3, got %d", s2.Cap())
	}
}

func TestForEach(t *testing.T) {
	s := New(1, 2, 3)
	sum := 0
	s.ForEach(func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach should sum to 6, got %d", sum)
	}
}

func TestForEachEmpty(t *testing.T) {
	s := New[int]()
	count := 0
	s.ForEach(func(v int) { count++ })
	if count != 0 {
		t.Errorf("ForEach on empty should not call fn, got %d calls", count)
	}
}

func TestForEachIndex(t *testing.T) {
	s := New(10, 20, 30)
	var indices []int
	var values []int
	s.ForEachIndex(func(i, v int) {
		indices = append(indices, i)
		values = append(values, v)
	})
	if !reflect.DeepEqual(indices, []int{0, 1, 2}) {
		t.Errorf("ForEachIndex indices should be [0, 1, 2], got %v", indices)
	}
	if !reflect.DeepEqual(values, []int{10, 20, 30}) {
		t.Errorf("ForEachIndex values should be [10, 20, 30], got %v", values)
	}
}

func TestForEachIndexEmpty(t *testing.T) {
	s := New[int]()
	count := 0
	s.ForEachIndex(func(i, v int) { count++ })
	if count != 0 {
		t.Errorf("ForEachIndex on empty should not call fn, got %d calls", count)
	}
}

func TestMap(t *testing.T) {
	original := New(1, 2, 3)
	mapped := original.Map(func(v int) int { return v * 2 })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(mapped.Data(), expected) {
		t.Errorf("Map should return [2, 4, 6], got %v", mapped.Data())
	}
	if !reflect.DeepEqual(original.Data(), []int{1, 2, 3}) {
		t.Errorf("Map should not modify original, expected [1, 2, 3], got %v", original.Data())
	}
}

func TestMapEmpty(t *testing.T) {
	s := New[int]()
	result := s.Map(func(v int) int { return v * 2 })
	if result.Length() != 0 {
		t.Errorf("Map on empty should return empty, got length %d", result.Length())
	}
}

func TestMapTo(t *testing.T) {
	s := New(1, 2, 3)
	result := MapTo(s, func(v int) string {
		if v == 1 {
			return "one"
		}
		return "other"
	})
	expected := []string{"one", "other", "other"}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("MapTo expected %v, got %v", expected, result.Data())
	}
}

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

func TestFilterEmptyResult(t *testing.T) {
	s := New(1, 3, 5)
	result := s.Filter(func(v int) bool { return v%2 == 0 })
	if result.Length() != 0 {
		t.Errorf("Filter with no match should return empty, got length %d", result.Length())
	}
}

func TestFind(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Find(func(v int) bool { return v > 1 })
	if !ok || val != 2 {
		t.Errorf("Find should return (2, true), got (%d, %v)", val, ok)
	}
	_, ok = s.Find(func(v int) bool { return v > 3 })
	if ok {
		t.Error("Find should return ok=false for no match")
	}
}

func TestFindIndex(t *testing.T) {
	s := New(1, 2, 3, 2)
	if idx := s.FindIndex(func(v int) bool { return v == 2 }); idx != 1 {
		t.Errorf("FindIndex for 2 should return 1, got %d", idx)
	}
	if idx := s.FindIndex(func(v int) bool { return v > 3 }); idx != -1 {
		t.Errorf("FindIndex for non-match should return -1, got %d", idx)
	}
}

func TestIndexOf(t *testing.T) {
	s := New(1, 2, 3)
	if idx := IndexOf(s, 2); idx != 1 {
		t.Errorf("IndexOf 2 should return 1, got %d", idx)
	}
	if idx := IndexOf(s, 4); idx != -1 {
		t.Errorf("IndexOf 4 should return -1, got %d", idx)
	}
}

func TestContains(t *testing.T) {
	s := New(1, 2, 3)
	if !Contains(s, 2) {
		t.Error("Contains(2) should be true")
	}
	if Contains(s, 4) {
		t.Error("Contains(4) should be false")
	}
}

func TestFirst(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.First()
	if !ok || val != 1 {
		t.Errorf("First should return (1, true), got (%d, %v)", val, ok)
	}
}

func TestFirstEmpty(t *testing.T) {
	s := New[int]()
	val, ok := s.First()
	if ok {
		t.Error("First on empty should return ok=false")
	}
	if val != 0 {
		t.Errorf("First on empty should return zero, got %d", val)
	}
}

func TestLast(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Last()
	if !ok || val != 3 {
		t.Errorf("Last should return (3, true), got (%d, %v)", val, ok)
	}
}

func TestLastEmpty(t *testing.T) {
	s := New[int]()
	val, ok := s.Last()
	if ok {
		t.Error("Last on empty should return ok=false")
	}
	if val != 0 {
		t.Errorf("Last on empty should return zero, got %d", val)
	}
}

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

func TestEveryEmpty(t *testing.T) {
	s := New[int]()
	if !s.Every(func(v int) bool { return false }) {
		t.Error("Every on empty slice should return true")
	}
}

func TestSome(t *testing.T) {
	s := New(1, 3, 4)
	if !s.Some(func(v int) bool { return v%2 == 0 }) {
		t.Error("Some should return true for even number")
	}
	if s.Some(func(v int) bool { return v > 5 }) {
		t.Error("Some should return false for no numbers > 5")
	}
}

func TestSomeEmpty(t *testing.T) {
	s := New[int]()
	if s.Some(func(v int) bool { return true }) {
		t.Error("Some on empty slice should return false")
	}
}

func TestReduce(t *testing.T) {
	s := New(1, 2, 3)
	sum := s.Reduce(func(acc, v int) int { return acc + v }, 0)
	if sum != 6 {
		t.Errorf("Reduce should sum to 6, got %d", sum)
	}
}

func TestReduceEmpty(t *testing.T) {
	s := New[int]()
	result := s.Reduce(func(acc, v int) int { return acc + v }, 42)
	if result != 42 {
		t.Errorf("Reduce on empty should return initial value 42, got %d", result)
	}
}

func TestReduceTo(t *testing.T) {
	s := New(1, 2, 3)
	result := ReduceTo(s, func(acc string, v int) string {
		if acc == "" {
			return string(rune('a' + v - 1))
		}
		return acc + string(rune('a'+v-1))
	}, "")
	if result != "abc" {
		t.Errorf("ReduceTo expected \"abc\", got %q", result)
	}
}

func TestSort(t *testing.T) {
	s := New(3, 1, 2).Sort(func(a, b int) bool { return a < b })
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Sort should order ascending, expected %v, got %v", expected, s.Data())
	}
}

func TestReverse(t *testing.T) {
	s := New(1, 2, 3).Reverse()
	expected := []int{3, 2, 1}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Reverse should invert order, expected %v, got %v", expected, s.Data())
	}
}

func TestReverseSingle(t *testing.T) {
	s := New(1).Reverse()
	expected := []int{1}
	if !reflect.DeepEqual(s.Data(), expected) {
		t.Errorf("Reverse of single element should be unchanged, got %v", s.Data())
	}
}

func TestReverseEmpty(t *testing.T) {
	s := New[int]().Reverse()
	if s.Length() != 0 {
		t.Errorf("Reverse of empty should be empty, got length %d", s.Length())
	}
}

func TestConcat(t *testing.T) {
	s1 := New(1, 2)
	s2 := New(3, 4)
	s3 := New(5, 6)
	result := s1.Concat(s2, s3)
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Concat should merge slices, expected %v, got %v", expected, result.Data())
	}
	// 原切片不变
	if !reflect.DeepEqual(s1.Data(), []int{1, 2}) {
		t.Errorf("Concat should not modify s1, got %v", s1.Data())
	}
}

func TestConcatSingle(t *testing.T) {
	s1 := New(1, 2)
	s2 := New(3, 4)
	result := s1.Concat(s2)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Concat expected %v, got %v", expected, result.Data())
	}
}

func TestConcatNone(t *testing.T) {
	s := New(1, 2)
	result := s.Concat()
	expected := []int{1, 2}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Concat with no args should return copy, expected %v, got %v", expected, result.Data())
	}
}

func TestSub(t *testing.T) {
	s := New(1, 2, 3, 4)
	result := s.Sub(1, 3)
	expected := []int{2, 3}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Sub(1, 3) should return [2, 3], got %v", result.Data())
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 2, 3, 4}) {
		t.Errorf("Sub should not modify original, got %v", s.Data())
	}
}

func TestSubClamp(t *testing.T) {
	s := New(1, 2, 3)
	result := s.Sub(-1, 10)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Sub(-1, 10) should return all elements, got %v", result.Data())
	}
}

func TestSubStartAfterEnd(t *testing.T) {
	s := New(1, 2, 3)
	result := s.Sub(2, 1)
	if result.Length() != 0 {
		t.Errorf("Sub(2, 1) should return empty, got length %d", result.Length())
	}
}

func TestGet(t *testing.T) {
	s := New(1, 2, 3)
	val, ok := s.Get(1)
	if !ok || val != 2 {
		t.Errorf("Get(1) should return (2, true), got (%d, %v)", val, ok)
	}
	_, ok = s.Get(3)
	if ok {
		t.Error("Get(3) should return ok=false for out of bounds")
	}
}

func TestGetNegative(t *testing.T) {
	s := New(1, 2, 3)
	_, ok := s.Get(-1)
	if ok {
		t.Error("Get(-1) should return ok=false")
	}
}

func TestSet(t *testing.T) {
	s := New(1, 2, 3)
	if !s.Set(1, 10) {
		t.Error("Set(1, 10) should succeed")
	}
	if !reflect.DeepEqual(s.Data(), []int{1, 10, 3}) {
		t.Errorf("Set should update value, expected [1, 10, 3], got %v", s.Data())
	}
	if s.Set(3, 4) {
		t.Error("Set(3, 4) should return false for out of bounds")
	}
}

func TestData(t *testing.T) {
	s := New(1, 2, 3)
	d := s.Data()
	d[0] = 10
	if !reflect.DeepEqual(s.Data(), []int{1, 2, 3}) {
		t.Errorf("Data should return copy, original should remain [1, 2, 3], got %v", s.Data())
	}
}

func TestRaw(t *testing.T) {
	s := New(1, 2, 3)
	raw := s.Raw()
	if !reflect.DeepEqual(raw, []int{1, 2, 3}) {
		t.Errorf("Raw should return [1, 2, 3], got %v", raw)
	}
	// Raw 返回的是底层切片，修改会影响 Slice 内部
	raw[0] = 10
	if s.Raw()[0] != 10 {
		t.Error("Modifying Raw should affect internal data")
	}
}

func TestClear(t *testing.T) {
	s := New(1, 2, 3)
	s.Clear()
	if s.Length() != 0 {
		t.Errorf("Clear should result in length 0, got %d", s.Length())
	}
	if !s.IsEmpty() {
		t.Error("Clear should result in empty Slice")
	}
}

func TestClone(t *testing.T) {
	s := New(1, 2, 3)
	clone := s.Clone()
	if !reflect.DeepEqual(clone.Data(), []int{1, 2, 3}) {
		t.Errorf("Clone should return [1, 2, 3], got %v", clone.Data())
	}
	clone.Set(0, 10)
	if reflect.DeepEqual(s.Data(), clone.Data()) {
		t.Error("Modifying clone should not affect original")
	}
}

func TestChain(t *testing.T) {
	result := New(3, 1, 2).
		Push(4, 5).
		Sort(func(a, b int) bool { return a < b }).
		Map(func(v int) int { return v * 2 })

	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result.Data(), expected) {
		t.Errorf("Chain result expected %v, got %v", expected, result.Data())
	}
}

func TestMemoryLeakPop(t *testing.T) {
	// 创建 cap > len 的切片，使 Pop 后仍能访问原位置
	type holder struct{ p *int }
	x := 42
	s := &Slice[holder]{data: make([]holder, 1, 3)}
	s.data[0] = holder{p: &x}
	_, ok := s.Pop()
	if !ok {
		t.Fatal("Pop should succeed")
	}
	// 扩展到底层数组容量以检查清零
	backing := s.Raw()[:cap(s.Raw())]
	if backing[0].p != nil {
		t.Error("Pop should clear the reference in underlying array at the popped index")
	}
}

func TestMemoryLeakShift(t *testing.T) {
	type holder struct{ p *int }
	x := 42
	y := 99
	s := &Slice[holder]{data: make([]holder, 2, 3)}
	s.data[0] = holder{p: &x}
	s.data[1] = holder{p: &y}
	_, ok := s.Shift()
	if !ok {
		t.Fatal("Shift should succeed")
	}
	if s.Length() != 1 {
		t.Fatalf("Shift should leave 1 element, got %d", s.Length())
	}
	// Shift 只移除 s.data[0]，不会清 s.data[1]，只验证剩余元素正确
	if s.Raw()[0].p != &y {
		t.Error("Shift should leave second element at index 0")
	}
}

func TestMemoryLeakRemove(t *testing.T) {
	type holder struct{ p *int }
	x := 42
	y := 99
	s := &Slice[holder]{data: make([]holder, 3, 4)}
	s.data[0] = holder{p: &x}
	s.data[1] = holder{p: &y}
	s.data[2] = holder{p: &y}
	_, ok := s.Remove(1)
	if !ok {
		t.Fatal("Remove should succeed")
	}
	if s.Length() != 2 {
		t.Fatalf("Remove should leave 2 elements, got %d", s.Length())
	}
	// 扩展到底层数组容量以检查清零位置（原始索引2）
	backing := s.Raw()[:cap(s.Raw())]
	if backing[2].p != nil {
		t.Error("Remove should clear the reference at the vacated position")
	}
}

func TestSortStable(t *testing.T) {
	type pair struct{ key, val int }
	s := New[pair]()
	s.Push(pair{2, 200}, pair{1, 100}, pair{2, 201}, pair{1, 101}, pair{3, 300})
	s.SortStable(func(a, b pair) bool { return a.key < b.key })
	raw := s.Raw()
	for i := 1; i < len(raw); i++ {
		if raw[i-1].key > raw[i].key {
			t.Errorf("SortStable: elements not sorted at index %d: %v > %v", i, raw[i-1], raw[i])
		}
	}
	var vals []int
	for _, p := range raw {
		if p.key == 2 {
			vals = append(vals, p.val)
		}
	}
	if len(vals) != 2 || vals[0] != 200 || vals[1] != 201 {
		t.Errorf("SortStable: expected [200, 201] (stable), got %v", vals)
	}
}

func ExampleNew() {
	s := New(1, 2, 3)
	s.ForEach(func(v int) {
		// process v
	})
	// Output:
}

func ExampleSlice_Push() {
	s := New[int]()
	s.Push(1, 2, 3)
	fmt.Println(s.Length())
	// Output: 3
}

func ExampleSlice_Filter() {
	s := New(1, 2, 3, 4, 5)
	filtered := s.Filter(func(v int) bool { return v%2 == 0 })
	fmt.Println(filtered.Raw())
	// Output: [2 4]
}

func ExampleSlice_Map() {
	s := New(1, 2, 3)
	mapped := s.Map(func(v int) int { return v * 2 })
	fmt.Println(mapped.Raw())
	// Output: [2 4 6]
}

func ExampleSlice_Reduce() {
	s := New(1, 2, 3, 4, 5)
	sum := s.Reduce(func(acc, cur int) int { return acc + cur }, 0)
	fmt.Println(sum)
	// Output: 15
}
