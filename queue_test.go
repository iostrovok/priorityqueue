package priorityqueue

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	. "github.com/iostrovok/check"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

type Final struct {
	sync.RWMutex
	Res []int
}

func (f *Final) Out() []int {
	f.Lock()
	out := make([]int, len(f.Res))
	copy(out, f.Res)
	f.Unlock()

	return out
}

func (f *Final) Add(i int) {
	f.Lock()
	f.Res = append(f.Res, i)
	f.Unlock()
}

func (s *testSuite) Test_Simple(c *C) {
	c.Assert(nil, IsNil)
}

func (s *testSuite) TestRun1(c *C) {
	RunOne(c, 100, 10, []int{1, 5, 2, 3, 4})
}

func (s *testSuite) TestRun2(c *C) {
	RunOne(c, 1000, 10, []int{1, 5})
}

func (s *testSuite) TestRun3(c *C) {
	RunOne(c, 10000, 100, []int{1, 5, 2, 3, 4, 6, 7})
}

func (s *testSuite) TestRunSimple1(c *C) {
	RunOne(c, 100, 10, []int{1})
}

func RunOne(c *C, countTest, lengthCh int, levels []int) {
	ctx, cancel := context.WithCancel(context.Background())
	q, out := New(ctx, lengthCh, levels)
	w := sync.WaitGroup{}
	waitCh := make(chan struct{}, 10*len(levels))

	distribution := map[int]int{}
	expectedLen := 0
	for i := 0; i < len(levels); i++ {
		level := levels[i]
		distribution[level] = 0
		expectedLen += countTest

		w.Add(1)
		go func(pr int) {
			defer w.Done()

			// waiting the start
			select {
			case <-waitCh:
			}

			// run
			for i := 0; i < countTest; i++ {
				q.Push(pr, pr)
			}
		}(level)
	}

	final := &Final{
		Res: make([]int, 0),
	}
	go checkCh(out, final)

	// delay before start
	time.Sleep(1 * time.Second)

	for i := 0; i <= len(levels); i++ {
		waitCh <- struct{}{}
		waitCh <- struct{}{}
	}

	w.Wait()

	// here we are waiting for all channels are done
	time.Sleep(1 * time.Second)

	// correct stop
	cancel()

	finalResult := final.Out()
	c.Logf("expected total tests: %d, found: ", expectedLen, len(finalResult))
	c.Assert(len(finalResult), Equals, expectedLen)

	// simple check distribution by priorities
	for i := range finalResult {
		distribution[finalResult[i]] += i
	}

	// it's not strict distribution!
	sort.Ints(levels)
	c.Logf("distribution: %+v", distribution)
	c.Logf("levels: %+v", levels)
	for j := len(levels) - 1; j > 0; j-- {
		s := (len(levels) - j) * countTest * countTest
		c.Logf("[%d, %d] %d < %d", j, levels[j], distribution[levels[j]], s)
		c.Assert(distribution[levels[j]], LessThan, s) // priority level is distributed around 1/level from total case
	}
}

func checkCh(out chan interface{}, final *Final) {
	for {
		select {
		case res, ok := <-out:
			if !ok {
				return
			}
			final.Add(res.(int))
		}
	}
}
