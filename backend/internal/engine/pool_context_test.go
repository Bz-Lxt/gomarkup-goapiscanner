package engine_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/engine"
	"github.com/alkaid/goapiscanner/internal/payload"
)

func TestPoolKeepsTimingProbeContextsIndependent(t *testing.T) {
	slowStarted := make(chan struct{})
	var startedOnce sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow":
			startedOnce.Do(func() { close(slowStarted) })
			select {
			case <-time.After(250 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("slow response"))
			case <-r.Context().Done():
				return
			}
		case "/fast":
			select {
			case <-slowStarted:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("fast response"))
			case <-time.After(time.Second):
				http.Error(w, "slow probe did not start", http.StatusServiceUnavailable)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	jobs := []payload.Job{
		{Method: http.MethodGet, URL: srv.URL + "/slow", Timeout: true},
		{Method: http.MethodGet, URL: srv.URL + "/fast", Timeout: true},
	}
	pool := engine.NewPool(2, time.Second)
	var mu sync.Mutex
	results := make([]engine.Result, 0, len(jobs))
	if err := pool.Run(context.Background(), jobs, func(result engine.Result) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}); err != nil {
		t.Fatalf("run timing probes: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(results) != len(jobs) {
		t.Fatalf("got %d results, want %d", len(results), len(jobs))
	}
	for _, result := range results {
		if result.Err != nil {
			t.Errorf("%s ended with an unrelated probe's context: %v", result.Job.URL, result.Err)
			continue
		}
		if result.StatusCode != http.StatusOK || len(result.Body) == 0 {
			t.Errorf("%s returned status=%d body=%q", result.Job.URL, result.StatusCode, result.Body)
		}
	}
}
