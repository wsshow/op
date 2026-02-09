// Package deque 提供了一个基于环形缓冲区的高性能泛型双端队列实现。
// 支持 O(1) 的头尾插入/删除操作，以及自动扩缩容。
package deque

import "fmt"

// minCapacity 是双端队列的最小容量，必须是 2 的幂以支持位运算取模。
const minCapacity = 16

// Deque 是一个基于环形缓冲区的双端队列，支持泛型类型 T。
// 零值可直接使用，无需显式初始化。
type Deque[T any] struct {
	buffer  []T
	headIdx int
	tailIdx int
	size    int
	baseCap int
}

// New 创建并返回一个新的双端队列实例。
func New[T any]() *Deque[T] {
	return &Deque[T]{baseCap: minCapacity}
}

// Capacity 返回当前缓冲区的容量。若队列为 nil 则返回 0。
func (d *Deque[T]) Capacity() int {
	if d == nil {
		return 0
	}
	return len(d.buffer)
}

// Size 返回当前存储的元素数量。若队列为 nil 则返回 0。
func (d *Deque[T]) Size() int {
	if d == nil {
		return 0
	}
	return d.size
}

// PushBack 在队列尾部添加元素。
func (d *Deque[T]) PushBack(elem T) {
	d.ensureCapacity()
	d.buffer[d.tailIdx] = elem
	d.tailIdx = d.nextIndex(d.tailIdx)
	d.size++
}

// PushFront 在队列头部添加元素。
func (d *Deque[T]) PushFront(elem T) {
	d.ensureCapacity()
	d.headIdx = d.prevIndex(d.headIdx)
	d.buffer[d.headIdx] = elem
	d.size++
}

// PopFront 从队列头部移除并返回元素。若队列为空则 panic。
func (d *Deque[T]) PopFront() T {
	if d.size == 0 {
		panic("deque: PopFront() called on empty queue")
	}
	var zero T
	elem := d.buffer[d.headIdx]
	d.buffer[d.headIdx] = zero
	d.headIdx = d.nextIndex(d.headIdx)
	d.size--
	d.shrinkIfNeeded()
	return elem
}

// PopBack 从队列尾部移除并返回元素。若队列为空则 panic。
func (d *Deque[T]) PopBack() T {
	if d.size == 0 {
		panic("deque: PopBack() called on empty queue")
	}
	var zero T
	d.tailIdx = d.prevIndex(d.tailIdx)
	elem := d.buffer[d.tailIdx]
	d.buffer[d.tailIdx] = zero
	d.size--
	d.shrinkIfNeeded()
	return elem
}

// Front 返回队列头部元素。若队列为空则 panic。
func (d *Deque[T]) Front() T {
	if d.size == 0 {
		panic("deque: Front() called when empty")
	}
	return d.buffer[d.headIdx]
}

// Back 返回队列尾部元素。若队列为空则 panic。
func (d *Deque[T]) Back() T {
	if d.size == 0 {
		panic("deque: Back() called when empty")
	}
	return d.buffer[d.prevIndex(d.tailIdx)]
}

// At 返回指定索引处的元素（不移除）。若索引无效则 panic。
func (d *Deque[T]) At(index int) T {
	d.checkIndex(index)
	return d.buffer[d.realIndex(index)]
}

// Set 将指定索引处的值设置为 item。若索引无效则 panic。
func (d *Deque[T]) Set(index int, item T) {
	d.checkIndex(index)
	d.buffer[d.realIndex(index)] = item
}

// Clear 清空队列中的所有元素，但保留当前缓冲区容量。
func (d *Deque[T]) Clear() {
	if d.size == 0 {
		return
	}
	clear(d.buffer)
	d.headIdx = 0
	d.tailIdx = 0
	d.size = 0
}

// Grow 确保队列有足够空间容纳 n 个额外元素。若 n 为负则 panic。
func (d *Deque[T]) Grow(n int) {
	if n < 0 {
		panic("deque.Grow: negative count")
	}
	if d.Capacity()-d.size >= n {
		return
	}
	newCap := d.calculateNewCapacity(n)
	d.resize(newCap)
}

// Rotate 将队列元素旋转 steps 步。正数向前旋转，负数向后旋转。
// 元素少于 2 个时无操作。
func (d *Deque[T]) Rotate(steps int) {
	if d == nil || d.size <= 1 {
		return
	}
	steps %= d.size
	if steps == 0 {
		return
	}

	if steps > 0 {
		for i := 0; i < steps; i++ {
			d.PushBack(d.PopFront())
		}
	} else {
		for i := 0; i < -steps; i++ {
			d.PushFront(d.PopBack())
		}
	}
}

// Index 返回第一个满足条件的元素索引（从头部开始搜索）。未找到返回 -1。
func (d *Deque[T]) Index(match func(T) bool) int {
	if d == nil || d.size == 0 {
		return -1
	}
	return d.search(match, true)
}

// RIndex 从尾部开始搜索第一个满足条件的元素索引。返回从头部计算的索引，未找到返回 -1。
func (d *Deque[T]) RIndex(match func(T) bool) int {
	if d == nil || d.size == 0 {
		return -1
	}
	return d.search(match, false)
}

// Insert 在指定位置插入元素。若索引 <= 0 则添加到头部，>= Size 则添加到尾部。
func (d *Deque[T]) Insert(at int, item T) {
	if at <= 0 {
		d.PushFront(item)
		return
	}
	if at >= d.size {
		d.PushBack(item)
		return
	}
	d.insertAtMiddle(at, item)
}

// Remove 移除并返回指定索引处的元素。若索引无效则 panic。
func (d *Deque[T]) Remove(at int) T {
	d.checkIndex(at)
	if at == 0 {
		return d.PopFront()
	}
	if at == d.size-1 {
		return d.PopBack()
	}
	return d.removeFromMiddle(at)
}

// SetBaseCap 设置基础容量（向上取整到 2 的幂）。
// 缩容时不会缩小到基础容量以下。
func (d *Deque[T]) SetBaseCap(baseCap int) {
	newCap := minCapacity
	for newCap < baseCap {
		newCap <<= 1
	}
	d.baseCap = newCap
	// 只有在已经分配了 buffer 且容量不足时才调整大小
	if d.buffer != nil && d.Capacity() < newCap {
		d.resize(newCap)
	}
}

// Swap 交换两个索引处的值。若索引无效则 panic。
func (d *Deque[T]) Swap(idxA, idxB int) {
	d.checkIndex(idxA)
	d.checkIndex(idxB)
	if idxA != idxB {
		a, b := d.realIndex(idxA), d.realIndex(idxB)
		d.buffer[a], d.buffer[b] = d.buffer[b], d.buffer[a]
	}
}

// checkIndex 检查索引是否有效。
func (d *Deque[T]) checkIndex(i int) {
	if i < 0 || i >= d.size {
		panic(fmt.Sprintf("deque: index out of range %d with length %d", i, d.size))
	}
}

// realIndex 将逻辑索引转换为缓冲区中的实际索引。
func (d *Deque[T]) realIndex(i int) int {
	return (d.headIdx + i) & (len(d.buffer) - 1)
}

// prevIndex 返回环形缓冲区中 i 的前一个索引。
func (d *Deque[T]) prevIndex(i int) int {
	return (i - 1) & (len(d.buffer) - 1)
}

// nextIndex 返回环形缓冲区中 i 的下一个索引。
func (d *Deque[T]) nextIndex(i int) int {
	return (i + 1) & (len(d.buffer) - 1)
}

// ensureCapacity 确保缓冲区有空间容纳新元素，必要时进行扩容。
func (d *Deque[T]) ensureCapacity() {
	if d.buffer == nil {
		if d.baseCap == 0 {
			d.baseCap = minCapacity
		}
		d.buffer = make([]T, d.baseCap)
	} else if d.size == len(d.buffer) {
		d.resize(d.size << 1)
	}
}

// shrinkIfNeeded 在队列占用低于缓冲区容量的 1/4 时缩减容量。
// 缩容后的容量不会低于 baseCap。
func (d *Deque[T]) shrinkIfNeeded() {
	if len(d.buffer) > d.baseCap && (d.size<<2) <= len(d.buffer) {
		newCap := d.size << 1
		if newCap < d.baseCap {
			newCap = d.baseCap
		}
		d.resize(newCap)
	}
}

// resize 重新分配缓冲区并复制现有元素。
func (d *Deque[T]) resize(newSize int) {
	newBuffer := make([]T, newSize)
	if d.size > 0 {
		if d.tailIdx > d.headIdx {
			copy(newBuffer, d.buffer[d.headIdx:d.tailIdx])
		} else {
			n := copy(newBuffer, d.buffer[d.headIdx:])
			copy(newBuffer[n:], d.buffer[:d.tailIdx])
		}
	}
	d.buffer = newBuffer
	d.headIdx = 0
	d.tailIdx = d.size
}

// calculateNewCapacity 计算容纳 n 个额外元素所需的最小 2 的幂容量。
func (d *Deque[T]) calculateNewCapacity(n int) int {
	newCap := max(d.Capacity(), minCapacity)
	for newCap < d.size+n {
		newCap <<= 1
	}
	return newCap
}

// search 在缓冲区中执行线性搜索。forward 为 true 时从头部搜索，否则从尾部搜索。
func (d *Deque[T]) search(match func(T) bool, forward bool) int {
	modBits := len(d.buffer) - 1
	if forward {
		for i := 0; i < d.size; i++ {
			if match(d.buffer[(d.headIdx+i)&modBits]) {
				return i
			}
		}
	} else {
		for i := d.size - 1; i >= 0; i-- {
			if match(d.buffer[(d.headIdx+i)&modBits]) {
				return i
			}
		}
	}
	return -1
}

// insertAtMiddle 在中间位置插入元素，选择移动较少元素的方向。
func (d *Deque[T]) insertAtMiddle(at int, item T) {
	if at*2 < d.size {
		d.PushFront(item)
		for i := 0; i < at; i++ {
			a := d.realIndex(i)
			b := d.realIndex(i + 1)
			d.buffer[a], d.buffer[b] = d.buffer[b], d.buffer[a]
		}
	} else {
		d.PushBack(item)
		for i := d.size - 2; i >= at; i-- {
			a := d.realIndex(i)
			b := d.realIndex(i + 1)
			d.buffer[a], d.buffer[b] = d.buffer[b], d.buffer[a]
		}
	}
}

// removeFromMiddle 从中间位置移除元素，选择移动较少元素的方向。
func (d *Deque[T]) removeFromMiddle(at int) T {
	realIdx := d.realIndex(at)
	elem := d.buffer[realIdx]
	if at*2 < d.size {
		for i := at; i > 0; i-- {
			a := d.realIndex(i - 1)
			b := d.realIndex(i)
			d.buffer[b] = d.buffer[a]
		}
		d.PopFront()
	} else {
		for i := at; i < d.size-1; i++ {
			a := d.realIndex(i)
			b := d.realIndex(i + 1)
			d.buffer[a] = d.buffer[b]
		}
		d.PopBack()
	}
	return elem
}
