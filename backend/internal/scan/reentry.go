package scan

import "sync"

// Guard prevents the same task id from being executed twice.
// Lock order: only this mutex; never acquire while holding a store lock.
type Guard struct {
	mu      sync.Mutex
	running map[string]struct{}
}

func NewGuard() *Guard {
	return &Guard{running: make(map[string]struct{})}
}

func (g *Guard) Acquire(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.running[id]; ok {
		return false
	}
	g.running[id] = struct{}{}
	return true
}

func (g *Guard) Release(id string) {
	g.mu.Lock()
	delete(g.running, id)
	g.mu.Unlock()
}

func (g *Guard) Running() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]string, 0, len(g.running))
	for id := range g.running {
		out = append(out, id)
	}
	return out
}
