package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSwaggerFallsBackWhenFirstCandidateFails(t *testing.T) {
	hits := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits[r.URL.Path]++
		switch r.URL.Path {
		case "/swagger.json":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("nginx 503"))
		case "/openapi.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"openapi":"3.0.0","paths":{"/healthz":{"get":{}}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := ProbeClient(5 * time.Second)
	body, src, err := FetchSwagger(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("expected fallback to openapi.json, got error: %v", err)
	}
	if src != srv.URL+"/openapi.json" {
		t.Fatalf("expected source to be openapi.json, got %q", src)
	}
	if string(body) == "" {
		t.Fatal("expected non-empty body")
	}
	if hits["/swagger.json"] != 1 {
		t.Fatalf("expected to probe swagger.json once, got %d", hits["/swagger.json"])
	}
	if hits["/openapi.json"] != 1 {
		t.Fatalf("expected to fall back to openapi.json, got %d", hits["/openapi.json"])
	}
}

func TestFetchSwaggerAllCandidatesFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := ProbeClient(5 * time.Second)
	_, _, err := FetchSwagger(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected error when all candidates fail")
	}
}

func TestFetchSwaggerSkipsNonSpecBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html>not a spec</html>`))
		case "/openapi.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"openapi":"3.0.0","paths":{"/healthz":{"get":{}}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := ProbeClient(5 * time.Second)
	_, src, err := FetchSwagger(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("expected fallback after non-spec body, got error: %v", err)
	}
	if src != srv.URL+"/openapi.json" {
		t.Fatalf("expected source to be openapi.json, got %q", src)
	}
}
