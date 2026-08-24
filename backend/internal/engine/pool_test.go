package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/payload"
)

func TestPoolCompletesAndCancel(t *testing.T) {
	var n atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	jobs := make([]payload.Job, 40)
	for i := range jobs {
		jobs[i] = payload.Job{Method: http.MethodGet, URL: srv.URL, Headers: map[string]string{}}
	}
	p := NewPool(8, 2*time.Second)
	var got atomic.Int64
	if err := p.Run(context.Background(), jobs, func(Result) { got.Add(1) }); err != nil {
		t.Fatal(err)
	}
	if got.Load() != 40 || n.Load() != 40 {
		t.Fatalf("got=%d n=%d", got.Load(), n.Load())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Run(ctx, jobs, nil); err == nil {
		t.Fatal("expected cancel error")
	}
}

func TestSemaphoreCancel(t *testing.T) {
	s := NewSemaphore(1)
	_ = s.Acquire(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := s.Acquire(ctx); err == nil {
		t.Fatal("expected ctx error")
	}
}
