package scan

import (
	"context"
	"sync"
)

type Canceller struct {
	mu    sync.Mutex
	funcs map[string]context.CancelFunc
}

func NewCanceller() *Canceller {
	return &Canceller{funcs: make(map[string]context.CancelFunc)}
}

func (c *Canceller) Bind(id string, cancel context.CancelFunc) {
	c.mu.Lock()
	c.funcs[id] = cancel
	c.mu.Unlock()
}

func (c *Canceller) Cancel(id string) bool {
	c.mu.Lock()
	fn, ok := c.funcs[id]
	c.mu.Unlock()
	if !ok {
		return false
	}
	fn()
	return true
}

func (c *Canceller) Drop(id string) {
	c.mu.Lock()
	delete(c.funcs, id)
	c.mu.Unlock()
}
