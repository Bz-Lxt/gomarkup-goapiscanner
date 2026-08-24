package api

import (
	"net/http"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
)

func Health(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":     "ok",
			"time":       clock.NowString(),
			"scan_mode":  cfg.ScanMode,
			"lab_public": cfg.LabPublicURL,
		})
	}
}

func Meta(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"scan_mode":        cfg.ScanMode,
			"lab_public_url":   cfg.LabPublicURL,
			"default_base_url": cfg.LabPublicURL,
			"timezone":         "Asia/Shanghai",
		})
	}
}
