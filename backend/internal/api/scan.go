package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/report"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/google/uuid"
)

type ScanAPI struct {
	Cfg  config.Config
	Orch *scan.Orchestrator
	Store *store.Store
	Font string
}

func (s *ScanAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateScanRequest
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(io.LimitReader(r.Body, s.Cfg.MaxBodyBytes)).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "JSON 无法解析")
			return
		}
	} else {
		if err := r.ParseMultipartForm(s.Cfg.MaxBodyBytes); err != nil {
			_ = r.ParseForm()
		}
		req.BaseURL = r.FormValue("base_url")
		req.Authorized = r.FormValue("authorized") == "true" || r.FormValue("authorized") == "1"
		req.Concurrency, _ = strconv.Atoi(r.FormValue("concurrency"))
		req.TimeoutMS, _ = strconv.Atoi(r.FormValue("timeout_ms"))
	}
	if !req.Authorized {
		writeErr(w, http.StatusBadRequest, "必须勾选授权声明后才能启动扫描")
		return
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		req.BaseURL = s.Cfg.LabPublicURL
	}
	if req.Concurrency <= 0 {
		req.Concurrency = s.Cfg.DefaultConc
	}
	if req.Concurrency > 64 {
		req.Concurrency = 64
	}
	if req.TimeoutMS <= 0 {
		req.TimeoutMS = s.Cfg.DefaultTimeout
	}
	var spec []byte
	var swaggerName string
	if r.MultipartForm != nil {
		if fhs := r.MultipartForm.File["swagger"]; len(fhs) > 0 {
			f, err := fhs[0].Open()
			if err == nil {
				spec, _ = io.ReadAll(io.LimitReader(f, s.Cfg.MaxBodyBytes))
				_ = f.Close()
				swaggerName = fhs[0].Filename
			}
		}
	}
	now := clock.NowString()
	task := model.Task{
		ID:          uuid.NewString(),
		BaseURL:     strings.TrimSpace(req.BaseURL),
		Status:      model.TaskPending,
		Concurrency: req.Concurrency,
		TimeoutMS:   req.TimeoutMS,
		Authorized:  true,
		SwaggerName: swaggerName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.Store.InsertTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, "任务落库失败")
		return
	}
	s.Orch.Start(task, spec)
	writeJSON(w, http.StatusCreated, task)
}

func (s *ScanAPI) List(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.ListTasks()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []model.Task{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *ScanAPI) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := s.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *ScanAPI) Findings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.Store.GetTask(id); err != nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	fs, err := s.Store.ListFindings(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if fs == nil {
		fs = []model.Finding{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"findings": fs,
		"tree":     store.BuildTree(fs),
		"stats":    store.StatsOf(fs),
	})
}

func (s *ScanAPI) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.Store.GetTask(id); err != nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	if !s.Orch.Cancel(id) {
		writeErr(w, http.StatusConflict, "任务未在运行")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelling"})
}

func (s *ScanAPI) Report(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := s.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	fs, err := s.Store.ListFindings(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if fs == nil {
		fs = []model.Finding{}
	}
	writeJSON(w, http.StatusOK, model.ReportPreview{
		Task:        t,
		Findings:    fs,
		Tree:        store.BuildTree(fs),
		Stats:       store.StatsOf(fs),
		Advice:      report.AdviceList(fs),
		GeneratedAt: clock.NowString(),
	})
}

func (s *ScanAPI) ReportPDF(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := s.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	fs, err := s.Store.ListFindings(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	b, err := report.RenderPDF(t, fs, s.Font)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "PDF 生成失败")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="scan-`+id+`.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
