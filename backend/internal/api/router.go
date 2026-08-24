package api

import (
	"net/http"

	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
)

func NewRouter(cfg config.Config, st *store.Store, orch *scan.Orchestrator, hub *ws.Hub, font string) http.Handler {
	api := &ScanAPI{Cfg: cfg, Orch: orch, Store: st, Font: font}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", Health(cfg))
	mux.HandleFunc("GET /api/v1/meta", Meta(cfg))
	mux.HandleFunc("POST /api/v1/scans", api.Create)
	mux.HandleFunc("GET /api/v1/scans", api.List)
	mux.HandleFunc("GET /api/v1/scans/{id}", api.Get)
	mux.HandleFunc("GET /api/v1/scans/{id}/findings", api.Findings)
	mux.HandleFunc("POST /api/v1/scans/{id}/cancel", api.Cancel)
	mux.HandleFunc("GET /api/v1/scans/{id}/report", api.Report)
	mux.HandleFunc("GET /api/v1/scans/{id}/report.pdf", api.ReportPDF)
	mux.HandleFunc("GET /api/v1/ws", hub.ServeWS)
	return CORS(cfg.CORSOrigin, AccessLog(mux))
}
