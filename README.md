## Probabilistic priority queue

The packet implements a priority queue with a high priority return probability:

1. A high-level item has a better chance of returning faster from the queue, but this condition is not very strict.
2. Within each level, the queue works according to the FIFO principle.
3. The queue only uses the standard golang "channel" data type.


## Example

see github.com/iostrovok/priorityqueue/console

```go

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iostrovok/priorityqueue"
)

const (
	countTest = 100 // count test for each level
	lengthCh  = 25  // length internal channel for each level
)

var (
	levels = []int{1, 10, 20, 30, 100}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	q, out := priorityqueue.New(ctx, lengthCh, levels)

	// call before pushing to avoid the deadlock
	go reader(out)

	w := sync.WaitGroup{}
	for _, pr := range levels {
		w.Add(1)
		go func(pr int) {
			// push N test with selected priority
			for i := 0; i < countTest; i++ {
				q.Push(pr, pr)
			}

			w.Done()
		}(pr)
	}

	// this will be processed with priority 30
	q.Push(55, 55)
	// this will be processed with priority 100
	q.Push(100, 110)
	// this will be processed with priority 1
	q.Push(0, 0)

	// here we just are waiting for all writers are done
	w.Wait()

	// here we just are waiting for reader is done
	time.Sleep(1 * time.Second)
	// correct stop
	cancel()
}

func reader(out chan interface{}) {
	for {
		select {
		case res, ok := <-out:
			if !ok {
				return
			}
			fmt.Printf("%d - ", res)
		}
	}
}


```