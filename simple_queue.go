package priorityqueue

// SimpleQueue is simple case of main top structure.
type SimpleQueue struct {
	ch chan interface{} // internal channel
}

// Push pushes a value with a given priority into the queue.
func (q *SimpleQueue) Push(data interface{}, _ int) {
	q.ch <- data
}
