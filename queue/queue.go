package queue

import "sync"

type Queue struct {
	mu    sync.Locker
	items []interface{}
}

func (q *Queue) Enqueue(items ...interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, items...)
}

func (q *Queue) Dequeue() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
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
