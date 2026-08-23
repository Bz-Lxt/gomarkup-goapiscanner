package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu   sync.RWMutex
	inst *slog.Logger
)

func Init(level string) {
	lv := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lv = slog.LevelDebug
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	mu.Lock()
	inst = slog.New(h)
	mu.Unlock()
}

func L() *slog.Logger {
	mu.RLock()
	l := inst
	mu.RUnlock()
	if l != nil {
		return l
	}
	Init("info")
	return L()
}

func SetOutput(w io.Writer, level slog.Level) {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	mu.Lock()
	inst = slog.New(h)
	mu.Unlock()
}
