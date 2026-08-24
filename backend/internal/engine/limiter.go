package engine

import "context"

// Semaphore is a buffered-channel limiter. Acquire is cancellable so a
// shutting-down pool cannot deadlock waiters behind a full queue.
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	if n < 1 {
		n = 1
	}
	return &Semaphore{ch: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.ch <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}

func (s *Semaphore) Cap() int {
	return cap(s.ch)
}
