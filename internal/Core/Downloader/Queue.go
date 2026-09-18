package downloader

import "sync"

type Queue struct {
	mu      sync.Mutex
	cond    *sync.Cond
	running int
	limit   int
	wg      sync.WaitGroup
}

func NewQueue(maxConcurrent int) *Queue {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	q := &Queue{limit: maxConcurrent}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Add(fn func()) {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.acquire()
		defer q.release()
		fn()
	}()
}

func (q *Queue) SetMaxConcurrent(maxConcurrent int) {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	q.mu.Lock()
	q.limit = maxConcurrent
	q.cond.Broadcast()
	q.mu.Unlock()
}

func (q *Queue) acquire() {
	q.mu.Lock()
	for q.running >= q.limit {
		q.cond.Wait()
	}
	q.running++
	q.mu.Unlock()
}

func (q *Queue) release() {
	q.mu.Lock()
	q.running--
	q.cond.Signal()
	q.mu.Unlock()
}

func (q *Queue) Wait() {
	q.wg.Wait()
}
