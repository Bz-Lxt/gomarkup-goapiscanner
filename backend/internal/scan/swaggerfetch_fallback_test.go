package scan_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alkaid/goapiscanner/internal/scan"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchSwaggerFallsBackAfterServerError(t *testing.T) {
	const document = `{"openapi":"3.0.0","paths":{}}`
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		response := &http.Response{Request: r, Header: make(http.Header)}
		switch r.URL.Path {
		case "/swagger.json":
			response.StatusCode = http.StatusServiceUnavailable
			response.Body = io.NopCloser(strings.NewReader("temporarily unavailable"))
		case "/openapi.json":
			response.StatusCode = http.StatusOK
			response.Body = io.NopCloser(strings.NewReader(document))
		default:
			response.StatusCode = http.StatusNotFound
			response.Body = http.NoBody
		}
		return response, nil
	})}

	body, source, err := scan.FetchSwagger(context.Background(), client, "https://service.example")
	if err != nil {
		t.Fatalf("FetchSwagger returned an error despite a valid fallback document: %v", err)
	}
	if source != "https://service.example/openapi.json" {
		t.Fatalf("FetchSwagger source = %q, want fallback URL", source)
	}
	if string(body) != document {
		t.Fatalf("FetchSwagger body = %q, want %q", body, document)
	}
}
