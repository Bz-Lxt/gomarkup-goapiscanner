package scan_test

import (
	"sync"
	"testing"

	"github.com/alkaid/goapiscanner/internal/scan"
)

func TestProgressConcurrentHitsRemainConsistent(t *testing.T) {
	const (
		workers = 32
		perWorker = 1000
	)

	progress := &scan.Progress{}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			<-start
			for range perWorker {
				progress.AddHit(5)
			}
		}()
	}
	close(start)
	wg.Wait()

	snapshot := progress.Snapshot()
	want := workers * perWorker
	if snapshot.Hits != want || snapshot.Crit != want {
		t.Fatalf("concurrent hit totals are inconsistent: hits=%d critical=%d, want both %d", snapshot.Hits, snapshot.Crit, want)
	}
}
