package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := os.Getenv("LAB_LISTEN")
	if addr == "" {
		addr = ":8090"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /swagger.json", serveSwagger)
	mux.HandleFunc("GET /openapi.json", serveSwagger)
	mux.HandleFunc("GET /api/users", usersSQLi)
	mux.HandleFunc("GET /api/users/blind", usersBlind)
	mux.HandleFunc("GET /api/search", searchXSS)
	mux.HandleFunc("GET /api/admin/secret", adminSecret)
	mux.HandleFunc("GET /api/files", filesTraversal)
	mux.HandleFunc("GET /api/ping", pingCMDi)

	srv := &http.Server{
		Addr:              addr,
		Handler:           logRequest(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("target-lab listen %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","lab":true}`))
}
