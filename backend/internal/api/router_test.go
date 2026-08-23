package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	logger.Init("error")
	st, err := store.Open(filepath.Join(t.TempDir(), "d"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cfg := config.Config{
		CORSOrigin:     "*",
		LabPublicURL:   "http://localhost:28483",
		LabInternalURL: "http://target-lab:8090",
		ScanMode:       "lab",
		MaxBodyBytes:   1 << 20,
		DefaultConc:    8,
		DefaultTimeout: 3000,
	}
	hub := ws.NewHub()
	orch := scan.NewOrchestrator(cfg, st, hub)
	return NewRouter(cfg, st, orch, hub, "")
}

func TestHealthAndReject(t *testing.T) {
	h := testRouter(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	body := []byte(`{"base_url":"http://localhost:28483","authorized":false}`)
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestTaskNotFound(t *testing.T) {
	h := testRouter(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/scans/nope", nil))
	if rr.Code != 404 {
		t.Fatal(rr.Code)
	}
}

func TestListEmpty(t *testing.T) {
	h := testRouter(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/scans", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	var env struct {
		OK   bool
		Data []model.Task
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK || env.Data == nil {
		t.Fatalf("%+v", env)
	}
	_ = clock.NowString()
}
