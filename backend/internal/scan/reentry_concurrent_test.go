package scan_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alkaid/goapiscanner/internal/scan"
)

func TestGuardConcurrentAcquireHasSingleWinner(t *testing.T) {
	const contenders = 256

	for round := 0; round < 50; round++ {
		g := scan.NewGuard()
		start := make(chan struct{})
		var ready sync.WaitGroup
		var done sync.WaitGroup
		var winners atomic.Int64
		ready.Add(contenders)
		done.Add(contenders)

		for i := 0; i < contenders; i++ {
			go func() {
				defer done.Done()
				ready.Done()
				<-start
				if g.Acquire("same-task") {
					winners.Add(1)
				}
			}()
		}

		ready.Wait()
		close(start)
		done.Wait()
		if got := winners.Load(); got != 1 {
			t.Fatalf("round %d: concurrent acquire admitted %d starters, want exactly one", round, got)
		}
	}
}
