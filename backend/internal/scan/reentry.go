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

// Acquire atomically claims id for execution. It returns true only for the
// first caller; every concurrent caller for an already-running id observes
// false. The check-then-set is performed under a single Lock so there is no
// window in which two goroutines can both see "absent" and both proceed.
func (g *Guard) Acquire(id string) bool {
	g.mu.Lock()
	if _, ok := g.running[id]; ok {
		g.mu.Unlock()
		return false
	}
	g.running[id] = struct{}{}
	g.mu.Unlock()
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
