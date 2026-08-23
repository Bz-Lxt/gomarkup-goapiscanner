package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
	"github.com/google/uuid"
)

func TestOrchestratorHitsLab(t *testing.T) {
	logger.Init("error")
	lab := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/swagger.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"openapi":"3.0.0","paths":{
				"/api/users":{"get":{"parameters":[{"name":"id","in":"query","schema":{"type":"string"}}]}},
				"/api/users/blind":{"get":{"parameters":[{"name":"id","in":"query","schema":{"type":"string"}}]}},
				"/api/search":{"get":{"parameters":[{"name":"q","in":"query","schema":{"type":"string"}}]}},
				"/api/admin/secret":{"get":{"parameters":[{"name":"Authorization","in":"header","schema":{"type":"string"}}]}},
				"/api/files":{"get":{"parameters":[{"name":"path","in":"query","schema":{"type":"string"}}]}},
				"/api/ping":{"get":{"parameters":[{"name":"host","in":"query","schema":{"type":"string"}}]}}
			}}`))
		case r.URL.Path == "/api/users":
			id := r.URL.Query().Get("id")
			if strings.Contains(id, "'") || strings.Contains(strings.ToUpper(id), "OR") {
				_, _ = w.Write([]byte(`{"sql_error":"You have an error in your SQL syntax"}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"ok"}`))
		case r.URL.Path == "/api/users/blind":
			if strings.Contains(strings.ToUpper(r.URL.Query().Get("id")), "SLEEP") {
				time.Sleep(3 * time.Second)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		case r.URL.Path == "/api/search":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>Results for: " + r.URL.Query().Get("q") + "</body></html>"))
		case r.URL.Path == "/api/admin/secret":
			w.Header().Set("X-Lab-Secret", "exposed")
			_, _ = w.Write([]byte(`{"admin_token":"lab-root-token"}`))
		case r.URL.Path == "/api/files":
			p := r.URL.Query().Get("path")
			if strings.Contains(p, "..") || strings.Contains(p, "etc/passwd") {
				_, _ = w.Write([]byte("root:x:0:0:lab_passwd"))
				return
			}
			_, _ = w.Write([]byte("ok"))
		case r.URL.Path == "/api/ping":
			h := r.URL.Query().Get("host")
			if strings.ContainsAny(h, ";|`$") {
				_, _ = w.Write([]byte("uid=0 lab_cmd_ok"))
				return
			}
			_, _ = w.Write([]byte("pong"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer lab.Close()

	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "d"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	cfg := config.Config{
		LabPublicURL:   lab.URL,
		LabInternalURL: lab.URL,
		ScanMode:       "lab",
		MaxJobs:        8000,
		DefaultConc:    10,
		DefaultTimeout: 8000,
	}
	o := NewOrchestrator(cfg, st, ws.NewHub())
	task := model.Task{
		ID: uuid.NewString(), BaseURL: lab.URL, Status: model.TaskPending,
		Concurrency: 10, TimeoutMS: 8000, Authorized: true,
		CreatedAt: clock.NowString(), UpdatedAt: clock.NowString(),
	}
	if err := st.InsertTask(task); err != nil {
		t.Fatal(err)
	}
	o.Start(task, nil)
	deadline := time.Now().Add(75 * time.Second)
	for time.Now().Before(deadline) {
		got, err := st.GetTask(task.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status == model.TaskSucceeded {
			fs, err := st.ListFindings(task.ID)
			if err != nil {
				t.Fatal(err)
			}
			seen := map[model.VulnClass]bool{}
			for _, f := range fs {
				seen[f.Class] = true
			}
			for _, c := range model.KnownClasses() {
				if !seen[c] {
					t.Fatalf("missing class %s findings=%d", c, len(fs))
				}
			}
			return
		}
		if got.Status == model.TaskFailed {
			t.Fatal(got.Error)
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatal("timeout waiting scan")
}

func TestGuardBlocksParallelStart(t *testing.T) {
	_ = context.Background()
	_ = os.TempDir()
	g := NewGuard()
	if !g.Acquire("x") || g.Acquire("x") {
		t.Fatal("reentry")
	}
}
