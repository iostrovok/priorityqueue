package priorityqueue

/*
The packet priorityqueue implements a priority queue with a high priority return probability:
A high-level item has a better chance of being returned from the queue faster, but this condition is not very strict.
*/

import (
	"context"
	"sort"
)

type IQueue interface {
	Push(data interface{}, priority int)
}

// Queue is main top structure.
type Queue struct {
	priority int
	length   int
	child    *Queue
	ch       chan interface{} // internal channel
}

// New creates a new priority queue and returns channel to get result.
// - ctx context.Context for final canceling
// - length is length internal channel for each level. Total memory usage is length * len(priority).
// - priority is a list of priorities.
func New(ctx context.Context, length int, priority []int) (IQueue, chan interface{}) {
	// Let's not make a big deal out of nothing
	if len(priority) < 2 {
		q := &SimpleQueue{
			ch: make(chan interface{}, length),
		}
		return q, q.ch
	}

	notDuplicates := map[int]bool{} // use only unique levels

	sort.Ints(priority)

	var queue *Queue
	for i := 0; i < len(priority); i++ {
		if notDuplicates[priority[i]] {
			continue
		}
		notDuplicates[priority[i]] = true

		q := &Queue{
			priority: priority[i],
			length:   length,
			ch:       make(chan interface{}, length),
		}

		if queue == nil {
			// the smallest level has not sub-levels
			queue = q
			continue
		}

		q.child = queue
		queue = q
	}

	out := queue.run(ctx)

	return queue, out
}

// Push pushes a value with a given priority into the queue.
func (q *Queue) Push(data interface{}, priority int) {
	if q.child == nil || q.priority <= priority {
		q.ch <- data
	} else {
		q.child.Push(data, priority)
	}
}

// Run starts the whole internal and return the result data channel
func (q *Queue) run(ctx context.Context) chan interface{} {
	out := make(chan interface{}, q.length)

	if q.child == nil {
		go func() {
			defer close(out)

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-q.ch:
					if !ok {
						return
					}
					out <- job
				}
			}
		}()

		return out
	}

	childChan := q.child.run(ctx)
	go func() {
		defer close(out)

		for {
			var job interface{}
			var ok bool

			select {
			case job = <-q.ch:
			default:
				select {
				case job = <-q.ch:
				case job, ok = <-childChan:
					if !ok {
						return
					}
				}
			}

			out <- job
		}
	}()

	return out
}
