package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alkaid/goapiscanner/internal/api"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/scan"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/ws"
)

func main() {
	cfg := config.FromEnv()
	logger.Init(cfg.LogLevel)
	st, err := store.Open(cfg.DataDir)
	if err != nil {
		logger.L().Error("store open", "err", err.Error())
		os.Exit(1)
	}
	defer st.Close()

	hub := ws.NewHub()
	orch := scan.NewOrchestrator(cfg, st, hub)
	font := os.Getenv("PDF_FONT")
	if font == "" {
		font = "/usr/share/fonts/SimHei.ttf"
	}
	h := api.NewRouter(cfg, st, orch, hub, font)
	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           h,
		ReadHeaderTimeout: 8 * time.Second,
	}
	go func() {
		logger.L().Info("listen", "addr", cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("serve", "err", err.Error())
			os.Exit(1)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
