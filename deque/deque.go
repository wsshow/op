package deque

import "fmt"

// minCapacity 是双端队列的最小容量，必须是2的幂，用于位运算取模
const minCapacity = 16

// Deque 表示一个双端队列实例，支持泛型类型T
type Deque[T any] struct {
	buffer  []T // 存储元素的缓冲区
	headIdx int // 头部索引
	tailIdx int // 尾部索引
	size    int // 当前元素数量
	baseCap int // 基础容量
}

// New 创建并返回一个新的双端队列实例
func New[T any]() *Deque[T] {
	return &Deque[T]{baseCap: minCapacity}
}

// Capacity 返回当前缓冲区的容量，若队列为nil则返回0
func (d *Deque[T]) Capacity() int {
	if d == nil {
		return 0
	}
	return len(d.buffer)
}

// Size 返回当前存储的元素数量，若队列为nil则返回0
func (d *Deque[T]) Size() int {
	if d == nil {
		return 0
	}
	return d.size
}

// PushBack 在队列尾部添加元素，支持FIFO和LIFO操作
func (d *Deque[T]) PushBack(elem T) {
	d.ensureCapacity()
	d.buffer[d.tailIdx] = elem
	d.tailIdx = d.nextIndex(d.tailIdx)
	d.size++
}

// PushFront 在队列头部添加元素
func (d *Deque[T]) PushFront(elem T) {
	d.ensureCapacity()
	d.headIdx = d.prevIndex(d.headIdx)
	d.buffer[d.headIdx] = elem
	d.size++
}

// PopFront 从队列头部移除并返回元素，若队列为空则panic
func (d *Deque[T]) PopFront() T {
	if d.size == 0 {
		panic("deque: PopFront() called on empty queue")
	}
	elem := d.buffer[d.headIdx]
	d.buffer[d.headIdx] = *new(T) // 清空元素
	d.headIdx = d.nextIndex(d.headIdx)
	d.size--
	d.shrinkIfNeeded()
	return elem
}

// PopBack 从队列尾部移除并返回元素，若队列为空则panic
func (d *Deque[T]) PopBack() T {
	if d.size == 0 {
		panic("deque: PopBack() called on empty queue")
	}
	d.tailIdx = d.prevIndex(d.tailIdx)
	elem := d.buffer[d.tailIdx]
	d.buffer[d.tailIdx] = *new(T) // 清空元素
	d.size--
	d.shrinkIfNeeded()
	return elem
}

// Front 返回队列头部元素，若队列为空则panic
func (d *Deque[T]) Front() T {
	if d.size == 0 {
		panic("deque: Front() called when empty")
	}
	return d.buffer[d.headIdx]
}

// Back 返回队列尾部元素，若队列为空则panic
func (d *Deque[T]) Back() T {
	if d.size == 0 {
		panic("deque: Back() called when empty")
	}
	return d.buffer[d.prevIndex(d.tailIdx)]
}

// At 返回指定索引处的元素，不移除元素，若索引无效则panic
func (d *Deque[T]) At(index int) T {
	d.checkIndex(index)
	return d.buffer[d.realIndex(index)]
}

// Set 将指定索引处的值设置为item，若索引无效则panic
func (d *Deque[T]) Set(index int, item T) {
	d.checkIndex(index)
	d.buffer[d.realIndex(index)] = item
}

// Clear 清空队列但保留当前容量
func (d *Deque[T]) Clear() {
	if d.size == 0 {
		return
	}
	// 清空整个缓冲区内容（避免内存泄漏）
	for i := range d.buffer {
		d.buffer[i] = *new(T)
	}
	d.headIdx = 0
	d.tailIdx = 0
	d.size = 0
}

// Grow 确保队列有足够空间容纳n个额外元素，若n为负则panic
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

// Rotate 将队列旋转n步，正数向后，负数向前，若元素少于2则无操作
func (d *Deque[T]) Rotate(steps int) {
	if d == nil || d.size <= 1 {
		return
	}
	steps %= d.size
	if steps == 0 {
		return
	}

	// 正向旋转(steps > 0): 将前面的元素移到后面
	// 负向旋转(steps < 0): 将后面的元素移到前面
	if steps > 0 {
		// 将前 steps 个元素移到后面
		for i := 0; i < steps; i++ {
			elem := d.PopFront()
			d.PushBack(elem)
		}
	} else {
		// 将后 -steps 个元素移到前面
		for i := 0; i < -steps; i++ {
			elem := d.PopBack()
			d.PushFront(elem)
		}
	}
}

// Index 返回第一个满足条件的元素索引，从头开始搜索，未找到返回-1
func (d *Deque[T]) Index(match func(T) bool) int {
	if d == nil || d.size == 0 {
		return -1
	}
	return d.search(match, true)
}

// RIndex 从尾部开始搜索第一个满足条件的元素索引，返回从头部计算的索引
func (d *Deque[T]) RIndex(match func(T) bool) int {
	if d == nil || d.size == 0 {
		return -1
	}
	return d.search(match, false)
}

// Insert 在指定位置插入元素，若索引超出范围则添加到头部或尾部
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

// Remove 移除并返回指定索引处的元素，若索引无效则panic
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

// SetBaseCap 设置基础容量，确保至少能存储指定数量的元素
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

// Swap 交换两个索引处的值，若索引无效则panic
func (d *Deque[T]) Swap(idxA, idxB int) {
	d.checkIndex(idxA)
	d.checkIndex(idxB)
	if idxA != idxB {
		a, b := d.realIndex(idxA), d.realIndex(idxB)
		d.buffer[a], d.buffer[b] = d.buffer[b], d.buffer[a]
	}
}

// 以下为内部辅助方法

// checkIndex 检查索引是否有效
func (d *Deque[T]) checkIndex(i int) {
	if i < 0 || i >= d.size {
		panic(fmt.Sprintf("deque: index out of range %d with length %d", i, d.size))
	}
}

// realIndex 计算实际缓冲区索引
func (d *Deque[T]) realIndex(i int) int {
	return (d.headIdx + i) & (len(d.buffer) - 1)
}

// prevIndex 计算前一个索引位置
func (d *Deque[T]) prevIndex(i int) int {
	return (i - 1) & (len(d.buffer) - 1)
}

// nextIndex 计算下一个索引位置
func (d *Deque[T]) nextIndex(i int) int {
	return (i + 1) & (len(d.buffer) - 1)
}

// ensureCapacity 在队列满时扩展容量
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

// shrinkIfNeeded 在队列占用小于1/4时缩减容量
func (d *Deque[T]) shrinkIfNeeded() {
	if len(d.buffer) > d.baseCap && (d.size<<2) <= len(d.buffer) {
		d.resize(d.size << 1)
	}
}

// resize 调整队列到指定大小
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

// calculateNewCapacity 计算需要的新容量
func (d *Deque[T]) calculateNewCapacity(n int) int {
	cap := max(d.Capacity(), minCapacity)
	for cap < d.size+n {
		cap <<= 1
	}
	return cap
}

// search 执行线性搜索，正向或反向
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

// insertAtMiddle 在中间位置插入元素
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

// removeFromMiddle 从中间移除元素
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

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
