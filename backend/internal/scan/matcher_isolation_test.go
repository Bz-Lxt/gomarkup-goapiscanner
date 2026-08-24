package scan_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
)

func TestParallelScansKeepFindingStateIsolated(t *testing.T) {
	logger.Init("error")
	st, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	arrived := make(chan string, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }

	newTarget := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			arrived <- name
			<-release
			_, _ = w.Write([]byte(`{"sql_error":"You have an error in your SQL syntax"}`))
		}))
	}
	targetA := newTarget("a")
	t.Cleanup(targetA.Close)
	targetB := newTarget("b")
	t.Cleanup(targetB.Close)
	t.Cleanup(unblock)

	cfg := config.Config{ScanMode: "authorized", MaxJobs: 1}
	orch := scan.NewOrchestrator(cfg, st, ws.NewHub())
	spec := []byte(`{"openapi":"3.0.0","paths":{"/probe":{"get":{"parameters":[{"name":"id","in":"query","schema":{"type":"string"}}]}}}}`)
	now := clock.NowString()
	tasks := []model.Task{
		{ID: "parallel-a", BaseURL: targetA.URL, Status: model.TaskPending, Concurrency: 1, TimeoutMS: 8000, Authorized: true, CreatedAt: now, UpdatedAt: now},
		{ID: "parallel-b", BaseURL: targetB.URL, Status: model.TaskPending, Concurrency: 1, TimeoutMS: 8000, Authorized: true, CreatedAt: now, UpdatedAt: now},
	}
	for _, task := range tasks {
		if err := st.InsertTask(task); err != nil {
			t.Fatal(err)
		}
		orch.Start(task, spec)
	}

	seenTargets := map[string]bool{}
	for len(seenTargets) < len(tasks) {
		select {
		case name := <-arrived:
			seenTargets[name] = true
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for both scans to reach their targets")
		}
	}
	unblock()

	for _, task := range tasks {
		deadline := time.Now().Add(5 * time.Second)
		for {
			got, err := st.GetTask(task.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status == model.TaskSucceeded {
				break
			}
			if got.Status == model.TaskFailed || got.Status == model.TaskCancelled {
				t.Fatalf("task %s finished with status %s: %s", task.ID, got.Status, got.Error)
			}
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for task %s", task.ID)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	for _, task := range tasks {
		findings, err := st.ListFindings(task.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(findings) != 1 {
			t.Errorf("task %s findings=%d, want 1", task.ID, len(findings))
		}
	}
}
