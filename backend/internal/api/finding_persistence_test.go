package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/api"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
)

func TestRejectedFindingDoesNotInflateTaskSummary(t *testing.T) {
	logger.Init("error")
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sql_error":"You have an error in your SQL syntax"}`))
	}))
	defer target.Close()

	dataDir := filepath.Join(t.TempDir(), "data")
	st, err := store.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "scanner.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TRIGGER reject_finding_writes
		BEFORE INSERT ON findings
		BEGIN
			SELECT RAISE(ABORT, 'finding storage unavailable');
		END;
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		CORSOrigin:     "*",
		LabPublicURL:   target.URL,
		LabInternalURL: target.URL,
		ScanMode:       "lab",
		MaxBodyBytes:   1 << 20,
		MaxJobs:        1,
		DefaultConc:    1,
		DefaultTimeout: 3000,
	}
	hub := ws.NewHub()
	orch := scan.NewOrchestrator(cfg, st, hub)
	handler := api.NewRouter(cfg, st, orch, hub, "")

	var form bytes.Buffer
	mw := multipart.NewWriter(&form)
	_ = mw.WriteField("base_url", target.URL)
	_ = mw.WriteField("authorized", "true")
	_ = mw.WriteField("concurrency", "1")
	part, err := mw.CreateFormFile("swagger", "openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(`{
		"openapi":"3.0.0",
		"paths":{"/search":{"get":{"parameters":[
			{"name":"q","in":"query","schema":{"type":"string"}}
		]}}}
	}`))
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	create := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", &form)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	handler.ServeHTTP(create, req)
	if create.Code != http.StatusCreated {
		t.Fatalf("create code=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		OK   bool       `json:"ok"`
		Data model.Task `json:"data"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if !created.OK || created.Data.ID == "" {
		t.Fatalf("unexpected create response: %s", create.Body.String())
	}

	var detail model.Task
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/scans/"+created.Data.ID, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("detail code=%d body=%s", rr.Code, rr.Body.String())
		}
		var got struct {
			OK   bool       `json:"ok"`
			Data model.Task `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		detail = got.Data
		if detail.Status == model.TaskSucceeded || detail.Status == model.TaskFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if detail.Status != model.TaskSucceeded {
		t.Fatalf("scan did not succeed: status=%s error=%q", detail.Status, detail.Error)
	}

	findings := httptest.NewRecorder()
	handler.ServeHTTP(findings, httptest.NewRequest(http.MethodGet, "/api/v1/scans/"+created.Data.ID+"/findings", nil))
	if findings.Code != http.StatusOK {
		t.Fatalf("findings code=%d body=%s", findings.Code, findings.Body.String())
	}
	var listed struct {
		OK   bool `json:"ok"`
		Data struct {
			Findings []model.Finding `json:"findings"`
		} `json:"data"`
	}
	if err := json.Unmarshal(findings.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if !listed.OK {
		t.Fatalf("unexpected findings response: %s", findings.Body.String())
	}
	if detail.Hits != len(listed.Data.Findings) {
		t.Fatalf("task reports %d hits, but findings endpoint returned %d records", detail.Hits, len(listed.Data.Findings))
	}
}
