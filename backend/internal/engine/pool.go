package engine

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/alkaid/goapiscanner/internal/payload"
)

type Pool struct {
	workers int
	client  *http.Client
	sem     *Semaphore
}

func NewPool(workers int, timeout time.Duration) *Pool {
	if workers < 1 {
		workers = 1
	}
	if workers > 64 {
		workers = 64
	}
	return &Pool{
		workers: workers,
		client:  NewClient(timeout),
		sem:     NewSemaphore(workers),
	}
}

func (p *Pool) Run(ctx context.Context, jobs []payload.Job, onResult func(Result)) error {
	if onResult == nil {
		onResult = func(Result) {}
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	timingCtx, cancelTiming := context.WithTimeout(ctx, p.client.Timeout)
	defer cancelTiming()

	for _, job := range jobs {
		if err := p.sem.Acquire(ctx); err != nil {
			wg.Wait()
			return err
		}
		wg.Add(1)
		j := job
		go func() {
			defer wg.Done()
			defer p.sem.Release()
			requestCtx := ctx
			if j.Timeout {
				requestCtx = timingCtx
			}
			res := Do(requestCtx, p.client, j)
			if j.Timeout {
				cancelTiming()
			}
			onResult(res)
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		<-done
		return ctx.Err()
	}
}

func (p *Pool) Workers() int { return p.workers }
