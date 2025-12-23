package deque

import (
	"slices"
	"testing"
	"unicode"
)

// TestEmpty 测试空队列的行为
func TestEmpty(t *testing.T) {
	q := new(Deque[string])
	if q.Size() != 0 {
		t.Errorf("q.Size() = %d, expected 0", q.Size())
	}
	if q.Capacity() != 0 {
		t.Errorf("expected q.Capacity() == 0, got %d", q.Capacity())
	}
	assertIndex(t, q, func(item string) bool { return true }, -1, "empty deque should return -1")
}

// TestNil 测试nil队列的行为
func TestNil(t *testing.T) {
	var q *Deque[int]
	if q.Size() != 0 {
		t.Errorf("expected q.Size() == 0, got %d", q.Size())
	}
	if q.Capacity() != 0 {
		t.Errorf("expected q.Capacity() == 0, got %d", q.Capacity())
	}
	q.Rotate(5) // nil队列旋转应无操作
	assertIndex(t, q, func(item int) bool { return true }, -1, "nil deque should return -1")
}

// TestFrontBack 测试队列前后端操作
func TestFrontBack(t *testing.T) {
	var q Deque[string]
	q.PushBack("foo")
	q.PushBack("bar")
	q.PushBack("baz")
	assertEqual(t, q.Front(), "foo", "wrong value at front")
	assertEqual(t, q.Back(), "baz", "wrong value at back")

	assertEqual(t, q.PopFront(), "foo", "wrong value removed from front")
	assertEqual(t, q.Front(), "bar", "wrong value remaining at front")
	assertEqual(t, q.Back(), "baz", "wrong value remaining at back")

	assertEqual(t, q.PopBack(), "baz", "wrong value removed from back")
	assertEqual(t, q.Front(), "bar", "wrong value remaining at front")
	assertEqual(t, q.Back(), "bar", "wrong value remaining at back")
}

// TestGrowShrinkBack 测试从尾部添加和移除时的容量调整
func TestGrowShrinkBack(t *testing.T) {
	var q Deque[int]
	size := minCapacity * 2
	values := make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = i // 值从 0 到 size-1
	}
	testGrowShrink(t, &q, size, q.PushBack, q.PopBack, values)
}

// TestGrowShrinkFront 测试从头部添加和移除时的容量调整
func TestGrowShrinkFront(t *testing.T) {
	var q Deque[int]
	size := minCapacity * 2
	values := make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = i // 值从 0 到 size-1
	}
	testGrowShrink(t, &q, size, q.PushFront, q.PopFront, values)
}

// TestSimple 测试简单的队列操作
func TestSimple(t *testing.T) {
	var q Deque[int]
	for i := 0; i < minCapacity; i++ {
		q.PushBack(i)
	}
	assertEqual(t, q.Front(), 0, "expected 0 at front")
	assertEqual(t, q.Back(), minCapacity-1, "expected %d at back", minCapacity-1)

	for i := 0; i < minCapacity; i++ {
		assertEqual(t, q.PopFront(), i, "wrong value at index %d", i)
	}

	q.Clear()
	for i := 0; i < minCapacity; i++ {
		q.PushFront(i)
	}
	for i := minCapacity - 1; i >= 0; i-- {
		assertEqual(t, q.PopFront(), i, "wrong value at index %d", i)
	}
}

// TestBufferWrap 测试缓冲区环绕（尾部操作）
func TestBufferWrap(t *testing.T) {
	var q Deque[int]
	for i := 0; i < minCapacity; i++ {
		q.PushBack(i)
	}
	for i := 0; i < 3; i++ {
		q.PopFront()
		q.PushBack(minCapacity + i)
	}
	for i := 0; i < minCapacity; i++ {
		assertEqual(t, q.PopFront(), i+3, "wrong value at index %d", i)
	}
}

// TestBufferWrapReverse 测试缓冲区环绕（头部操作）
func TestBufferWrapReverse(t *testing.T) {
	var q Deque[int]
	for i := 0; i < minCapacity; i++ {
		q.PushFront(i)
	}
	for i := 0; i < 3; i++ {
		q.PopBack()
		q.PushFront(minCapacity + i)
	}
	for i := 0; i < minCapacity; i++ {
		assertEqual(t, q.PopBack(), i+3, "wrong value at index %d", i)
	}
}

// TestSize 测试队列大小
func TestSize(t *testing.T) {
	var q Deque[int]
	assertEqual(t, q.Size(), 0, "empty queue size not 0")

	// 准备测试数据
	values := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = i
	}

	// 调用 testSize，传入所有必要参数
	testSize(t, &q, 1000, q.PushBack, q.PopFront, values)
}

// TestBack 测试尾部元素访问
func TestBack(t *testing.T) {
	var q Deque[int]
	for i := 0; i < minCapacity+5; i++ {
		q.PushBack(i)
		assertEqual(t, q.Back(), i, "Back returned wrong value")
	}
}

// TestGrow 测试容量增长
func TestGrow(t *testing.T) {
	var q Deque[int]
	assertPanics(t, "should panic with negative size", func() { q.Grow(-1) })

	testGrowCapacity(t, &q, 35, 64)
	testGrowCapacity(t, &q, 55, 64)
	testGrowCapacity(t, &q, 77, 128)

	for i := 0; i < 127; i++ {
		q.PushBack(i)
	}
	assertEqual(t, q.Capacity(), 128, "expected no capacity change")
	testGrowCapacity(t, &q, 2, 256)
}

// TestNew 测试队列初始化
func TestNew(t *testing.T) {
	minCap := 64
	q := &Deque[string]{}
	q.SetBaseCap(minCap)
	assertEqual(t, q.Capacity(), 0, "should not have allocated memory yet")
	q.PushBack("foo")
	q.PopFront()
	assertEqual(t, q.Size(), 0, "Size() should return 0")
	assertEqual(t, q.Capacity(), minCap, "wrong capacity, expected %d", minCap)

	curCap := 128
	q = new(Deque[string])
	q.SetBaseCap(minCap)
	q.Grow(curCap)
	assertEqual(t, q.Capacity(), curCap, "wrong capacity, expected %d", curCap)
	assertEqual(t, q.Size(), 0, "Size() should return 0")
	q.PushBack("foo")
	assertEqual(t, q.Capacity(), curCap, "wrong capacity, expected %d", curCap)
}

// TestRotate 测试队列旋转
func TestRotate(t *testing.T) {
	testRotate(t, 10)
	testRotate(t, minCapacity)
	testRotate(t, minCapacity+minCapacity/2)

	var q Deque[int]
	for i := 0; i < 10; i++ {
		q.PushBack(i)
	}
	q.Rotate(11)
	assertEqual(t, q.Front(), 1, "rotating 11 places should be same as 1")
	q.Rotate(-21)
	assertEqual(t, q.Front(), 0, "rotating -21 places should be same as -1")
	q.Rotate(q.Size())
	assertEqual(t, q.Front(), 0, "should not have rotated")
	q.Clear()
	q.PushBack(0)
	q.Rotate(13)
	assertEqual(t, q.Front(), 0, "should not have rotated")
}

// TestAt 测试元素访问
func TestAt(t *testing.T) {
	var q Deque[int]
	for i := 0; i < 1000; i++ {
		q.PushBack(i)
	}
	for i := 0; i < q.Size(); i++ {
		assertEqual(t, q.At(i), i, "index %d should contain %d", i, i)
	}
}

// TestSet 测试元素设置
func TestSet(t *testing.T) {
	var q Deque[int]
	for i := 0; i < 1000; i++ {
		q.PushBack(i)
		q.Set(i, i+50)
	}
	for i := 0; i < q.Size(); i++ {
		assertEqual(t, q.At(i), i+50, "index %d should contain %d", i, i+50)
	}
}

// TestClear 测试队列清空
func TestClear(t *testing.T) {
	var q Deque[int]
	for i := 0; i < 100; i++ {
		q.PushBack(i)
	}
	cap := q.Capacity()
	q.Clear()
	assertEqual(t, q.Size(), 0, "empty queue size not 0 after clear")
	assertEqual(t, q.Capacity(), cap, "capacity changed after clear")
	assertBufferCleared(t, &q)

	for i := 0; i < 128; i++ {
		q.PushBack(i)
	}
	q.Clear()
	assertBufferCleared(t, &q)
}

// TestIndex 测试正向索引搜索
func TestIndex(t *testing.T) {
	var q Deque[rune]
	for _, x := range "Hello, 世界" {
		q.PushBack(x)
	}
	assertIndex(t, &q, func(r rune) bool { return unicode.Is(unicode.Han, r) }, 7, "expected index 7")
	assertIndex(t, &q, func(r rune) bool { return r == 'H' }, 0, "expected index 0")
	assertIndex(t, &q, func(r rune) bool { return false }, -1, "expected index -1")
}

// TestRIndex 测试反向索引搜索
func TestRIndex(t *testing.T) {
	var q Deque[rune]
	for _, x := range "Hello, 世界" {
		q.PushBack(x)
	}
	assertIndexReverse(t, &q, func(r rune) bool { return unicode.Is(unicode.Han, r) }, 8, "expected index 8")
	assertIndexReverse(t, &q, func(r rune) bool { return r == 'H' }, 0, "expected index 0")
	assertIndexReverse(t, &q, func(r rune) bool { return false }, -1, "expected index -1")
}

// TestInsert 测试元素插入
func TestInsert(t *testing.T) {
	q := new(Deque[rune])
	for _, x := range "ABCDEFG" {
		q.PushBack(x)
	}
	q.Insert(4, 'x')
	assertEqual(t, q.At(4), 'x', "expected x at position 4")
	q.Insert(2, 'y')
	assertEqual(t, q.At(2), 'y', "expected y at position 2")
	assertEqual(t, q.At(5), 'x', "expected x at position 5")
	q.Insert(0, 'b')
	assertEqual(t, q.Front(), 'b', "expected b at front")
	q.Insert(q.Size(), 'e')
	assertSequence(t, q, []rune("bAByCDxEFGe"), "wrong sequence after inserts")
}

// TestRemove 测试元素移除
func TestRemove(t *testing.T) {
	q := new(Deque[rune])
	for _, x := range "ABCDEFG" {
		q.PushBack(x)
	}
	assertEqual(t, q.Remove(4), 'E', "expected E from position 4")
	assertEqual(t, q.Remove(2), 'C', "expected C from position 2")
	assertEqual(t, q.Back(), 'G', "expected G at back")
	assertEqual(t, q.Remove(0), 'A', "expected A from front")
	assertEqual(t, q.Remove(q.Size()-1), 'G', "expected G from back")
	assertEqual(t, q.Size(), 3, "wrong length")
}

// TestSwap 测试元素交换
func TestSwap(t *testing.T) {
	var q Deque[string]
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		q.PushBack(s)
	}
	q.Swap(0, 4)
	assertEqual(t, q.Front(), "e", "wrong value at front")
	assertEqual(t, q.Back(), "a", "wrong value at back")
	q.Swap(3, 1)
	assertEqual(t, q.At(1), "d", "wrong value at index 1")
	assertEqual(t, q.At(3), "b", "wrong value at index 3")
	q.Swap(2, 2) // 同位置交换
	assertEqual(t, q.At(2), "c", "wrong value at index 2")
	assertPanics(t, "should panic with out-of-range index", func() { q.Swap(1, 5) })
}

// TestFrontBackOutOfRangePanics 测试空队列前后访问的panic
func TestFrontBackOutOfRangePanics(t *testing.T) {
	var q Deque[int]
	assertPanics(t, "should panic when peeking empty queue", func() { q.Front() })
	assertPanics(t, "should panic when peeking empty queue", func() { q.Back() })
	q.PushBack(1)
	q.PopFront()
	assertPanics(t, "should panic when peeking emptied queue", func() { q.Front() })
	assertPanics(t, "should panic when peeking emptied queue", func() { q.Back() })
}

// TestPopFrontOutOfRangePanics 测试空队列PopFront的panic
func TestPopFrontOutOfRangePanics(t *testing.T) {
	var q Deque[int]
	assertPanics(t, "should panic when popping empty queue", func() { q.PopFront() })
	q.PushBack(1)
	q.PopFront()
	assertPanics(t, "should panic when popping emptied queue", func() { q.PopFront() })
}

// TestPopBackOutOfRangePanics 测试空队列PopBack的panic
func TestPopBackOutOfRangePanics(t *testing.T) {
	var q Deque[int]
	assertPanics(t, "should panic when popping empty queue", func() { q.PopBack() })
	q.PushBack(1)
	q.PopBack()
	assertPanics(t, "should panic when popping emptied queue", func() { q.PopBack() })
}

// TestAtOutOfRangePanics 测试越界At的panic
func TestAtOutOfRangePanics(t *testing.T) {
	var q Deque[int]
	q.PushBack(1)
	q.PushBack(2)
	q.PushBack(3)
	assertPanics(t, "should panic with negative index", func() { q.At(-4) })
	assertPanics(t, "should panic with out-of-range index", func() { q.At(4) })
}

// TestSetOutOfRangePanics 测试越界Set的panic
func TestSetOutOfRangePanics(t *testing.T) {
	var q Deque[int]
	q.PushBack(1)
	q.PushBack(2)
	q.PushBack(3)
	assertPanics(t, "should panic with negative index", func() { q.Set(-4, 1) })
	assertPanics(t, "should panic with out-of-range index", func() { q.Set(4, 1) })
}

// TestInsertOutOfRangePanics 测试越界插入
func TestInsertOutOfRangePanics(t *testing.T) {
	q := new(Deque[string])
	q.Insert(1, "A")
	assertEqual(t, q.Front(), "A", "expected A at front")
	q.Insert(-1, "B")
	assertEqual(t, q.Front(), "B", "expected B at front")
	q.Insert(999, "C")
	assertEqual(t, q.Back(), "C", "expected C at back")
}

// TestRemoveOutOfRangePanics 测试越界移除的panic
func TestRemoveOutOfRangePanics(t *testing.T) {
	q := new(Deque[string])
	assertPanics(t, "should panic when removing from empty queue", func() { q.Remove(0) })
	q.PushBack("A")
	assertPanics(t, "should panic with negative index", func() { q.Remove(-1) })
	assertPanics(t, "should panic with out-of-range index", func() { q.Remove(1) })
}

// TestSetBaseCapacity 测试设置基础容量
func TestSetBaseCapacity(t *testing.T) {
	var q Deque[string]
	q.SetBaseCap(200)
	q.PushBack("A")
	assertEqual(t, q.baseCap, 256, "wrong minimum capacity")
	assertEqual(t, q.Capacity(), 256, "wrong buffer size")
	q.PopBack()
	assertEqual(t, q.baseCap, 256, "wrong minimum capacity")
	assertEqual(t, q.Capacity(), 256, "wrong buffer size")
	q.SetBaseCap(0)
	assertEqual(t, q.baseCap, minCapacity, "wrong minimum capacity")
}

// 以下为基准测试

// BenchmarkPushFront 基准测试头部添加性能
func BenchmarkPushFront(b *testing.B) {
	var q Deque[int]
	for i := 0; i < b.N; i++ {
		q.PushFront(i)
	}
}

// BenchmarkPushBack 基准测试尾部添加性能
func BenchmarkPushBack(b *testing.B) {
	var q Deque[int]
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
}

// BenchmarkSerial 基准测试顺序添加和移除性能
func BenchmarkSerial(b *testing.B) {
	var q Deque[int]
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	for i := 0; i < b.N; i++ {
		q.PopFront()
	}
}

// BenchmarkSerialReverse 基准测试逆序添加和移除性能
func BenchmarkSerialReverse(b *testing.B) {
	var q Deque[int]
	for i := 0; i < b.N; i++ {
		q.PushFront(i)
	}
	for i := 0; i < b.N; i++ {
		q.PopBack()
	}
}

// BenchmarkRotate 基准测试旋转性能
func BenchmarkRotate(b *testing.B) {
	q := new(Deque[int])
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Rotate(b.N - 1)
	}
}

// BenchmarkInsert 基准测试中间插入性能
func BenchmarkInsert(b *testing.B) {
	q := new(Deque[int])
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Insert(q.Size()/2, -i)
	}
}

// BenchmarkRemove 基准测试中间移除性能
func BenchmarkRemove(b *testing.B) {
	q := new(Deque[int])
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Remove(q.Size() / 2)
	}
}

// BenchmarkYoyo 基准测试反复添加和移除性能
func BenchmarkYoyo(b *testing.B) {
	var q Deque[int]
	for i := 0; i < b.N; i++ {
		for j := 0; j < 65536; j++ {
			q.PushBack(j)
		}
		for j := 0; j < 65536; j++ {
			q.PopFront()
		}
	}
}

// BenchmarkYoyoFixed 基准测试固定容量下的反复添加和移除性能
func BenchmarkYoyoFixed(b *testing.B) {
	var q Deque[int]
	q.SetBaseCap(64000)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 65536; j++ {
			q.PushBack(j)
		}
		for j := 0; j < 65536; j++ {
			q.PopFront()
		}
	}
}

// BenchmarkClearContiguous 基准测试连续缓冲区清空性能
func BenchmarkClearContiguous(b *testing.B) {
	var src Deque[int]
	for i := 0; i < 1<<15+1<<14; i++ {
		src.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := src
		q.buffer = slices.Clone(q.buffer)
		q.Clear()
	}
}

// BenchmarkClearSplit 基准测试分割缓冲区清空性能
func BenchmarkClearSplit(b *testing.B) {
	var src Deque[int]
	for i := 0; i < 1<<15+1<<14; i++ {
		src.PushFront(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := src
		q.buffer = slices.Clone(q.buffer)
		q.Clear()
	}
}

// 以下为测试辅助函数

// assertEqual 检查实际值与期望值是否相等
func assertEqual[T comparable](t *testing.T, got, want T, msg string, args ...interface{}) {
	if got != want {
		t.Errorf(msg+", got %v, want %v", append(args, got, want)...)
	}
}

// assertPanics 检查函数是否引发panic
func assertPanics(t *testing.T, msg string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: didn't panic as expected", msg)
		}
	}()
	f()
}

// assertIndex 检查正向索引搜索结果
func assertIndex[T any](t *testing.T, q *Deque[T], match func(T) bool, want int, msg string) {
	got := q.Index(match)
	if got != want {
		t.Errorf("%s: got index %d, want %d", msg, got, want)
	}
}

// assertIndexReverse 检查反向索引搜索结果
func assertIndexReverse[T any](t *testing.T, q *Deque[T], match func(T) bool, want int, msg string) {
	got := q.RIndex(match)
	if got != want {
		t.Errorf("%s: got index %d, want %d", msg, got, want)
	}
}

// assertBufferCleared 检查缓冲区是否被清空
func assertBufferCleared[T comparable](t *testing.T, q *Deque[T]) {
	var zero T
	for i := 0; i < len(q.buffer); i++ {
		if q.buffer[i] != zero {
			t.Errorf("buffer has non-zero elements after Clear() at index %d", i)
			break
		}
	}
}

// assertSequence 检查队列内容是否与预期序列一致
func assertSequence[T comparable](t *testing.T, q *Deque[T], want []T, msg string) {
	for i, x := range want {
		if got := q.PopFront(); got != x {
			t.Errorf("%s: at position %d, got %v, want %v", msg, i, got, x)
		}
	}
}

// testGrowShrink 测试队列增长和缩减
func testGrowShrink[T comparable](t *testing.T, q *Deque[T], size int, push func(T), pop func() T, values []T) {
	if len(values) != size {
		t.Fatalf("values length %d does not match size %d", len(values), size)
	}

	// 测试增长
	for i := 0; i < size; i++ {
		if q.Size() != i {
			t.Errorf("Size() = %d, expected %d", q.Size(), i)
		}
		push(values[i])
	}

	// 记录初始容量
	initialCap := q.Capacity()

	// 测试缩减
	for i := size - 1; i >= 0; i-- {
		if q.Size() != i+1 {
			t.Errorf("Size() = %d, expected %d", q.Size(), i+1)
		}
		x := pop()
		if x != values[i] {
			t.Errorf("Pop returned %v, expected %v", x, values[i])
		}
	}

	// 检查最终状态
	if q.Size() != 0 {
		t.Errorf("Size() = %d, expected 0", q.Size())
	}
	if q.Capacity() >= initialCap {
		t.Errorf("buffer did not shrink, capacity = %d, initial = %d", q.Capacity(), initialCap)
	}
}

// testSize 测试队列大小变化
func testSize[T comparable](t *testing.T, q *Deque[T], n int, push func(T), pop func() T, values []T) {
	if len(values) != n {
		t.Fatalf("values length %d does not match n %d", len(values), n)
	}

	// 测试添加时的size变化
	for i := 0; i < n; i++ {
		push(values[i])
		if q.Size() != i+1 {
			t.Errorf("adding: size = %d, expected %d", q.Size(), i+1)
		}
	}

	// 测试移除时的size变化
	for i := 0; i < n; i++ {
		pop()
		if q.Size() != n-i-1 {
			t.Errorf("removing: size = %d, expected %d", q.Size(), n-i-1)
		}
	}
}

// testGrowCapacity 测试容量增长到指定值
func testGrowCapacity(t *testing.T, q *Deque[int], growSize, expectedCap int) {
	q.Grow(growSize)
	if q.Capacity() != expectedCap {
		t.Errorf("did not grow to expected capacity, got %d, want %d", q.Capacity(), expectedCap)
	}
}

// testRotate 测试旋转逻辑
func testRotate(t *testing.T, size int) {
	var q Deque[int]
	for i := 0; i < size; i++ {
		q.PushBack(i)
	}
	for i := 0; i < q.Size(); i++ {
		q.Rotate(1)
		assertEqual(t, q.Back(), i, "wrong value during rotation")
	}
	for i := q.Size() - 1; i >= 0; i-- {
		q.Rotate(-1)
		assertEqual(t, q.Front(), i, "wrong value during reverse rotation")
	}
}
