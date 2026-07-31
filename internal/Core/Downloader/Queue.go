package downloader

import "sync"

type Queue struct {
	worker chan struct{}
	wg     sync.WaitGroup
}

func NewQueue(maxConcurrent int) *Queue {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	return &Queue{
		worker: make(chan struct{}, maxConcurrent),
	}
}

func (q *Queue) Add(fn func()) {
	q.wg.Add(1)
	q.worker <- struct{}{}

	go func() {
		defer q.wg.Done()
		defer func() { <-q.worker }()
		fn()
	}()
}

func (q *Queue) Wait() {
	q.wg.Wait()
}
