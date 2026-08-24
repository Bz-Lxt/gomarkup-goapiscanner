package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Listen         string
	DataDir        string
	CORSOrigin     string
	LabPublicURL   string
	LabInternalURL string
	ScanMode       string // lab | authorized
	LogLevel       string
	MaxBodyBytes   int64
	MaxJobs        int
	DefaultConc    int
	DefaultTimeout int
}

func FromEnv() Config {
	return Config{
		Listen:         env("LISTEN", ":8080"),
		DataDir:        env("DATA_DIR", "./data"),
		CORSOrigin:     env("CORS_ORIGIN", "http://localhost:28481"),
		LabPublicURL:   strings.TrimRight(env("LAB_PUBLIC_URL", "http://localhost:28483"), "/"),
		LabInternalURL: strings.TrimRight(env("LAB_INTERNAL_URL", "http://target-lab:8090"), "/"),
		ScanMode:       strings.ToLower(env("SCAN_MODE", "lab")),
		LogLevel:       env("LOG_LEVEL", "info"),
		MaxBodyBytes:   int64(envInt("MAX_BODY_BYTES", 8<<20)),
		MaxJobs:        envInt("MAX_JOBS", 8000),
		DefaultConc:    envInt("DEFAULT_CONCURRENCY", 16),
		DefaultTimeout: envInt("DEFAULT_TIMEOUT_MS", 5000),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
