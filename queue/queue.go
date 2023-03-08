package queue

import "sync"

type Queue struct {
	mu    sync.Mutex
	items []interface{}
}

func NewQueue() *Queue {
	return new(Queue)
}

func (q *Queue) Enqueue(items ...interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, items...)
}

func (q *Queue) Dequeue() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.IsEmpty() {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) Peek() interface{} {
	return q.items[0]
}

func (q *Queue) Count() int {
	return len(q.items)
}

func (q *Queue) Contains(item interface{}) bool {
	for _, qItem := range q.items {
		if qItem == item {
			return true
		}
	}
	return false
}

func (q *Queue) ToSlice() []interface{} {
	return q.items
}

func (q *Queue) IsEmpty() bool {
	return q.Count() == 0
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = nil
}

func (q *Queue) ForEach(f func(interface{})) {
	for _, qItem := range q.items {
		f(qItem)
	}
}

func (q *Queue) Map(f func(interface{}) interface{}) *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()
	nq := NewQueue()
	for _, qItem := range q.items {
		nq.Enqueue(f(qItem))
	}
	return nq
}
