package engine_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/engine"
	"github.com/alkaid/goapiscanner/internal/fingerprint"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/payload"
)

type redirectRoundTripper func(*http.Request) (*http.Response, error)

func (f redirectRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDoPreservesLastResponseAtRedirectLimit(t *testing.T) {
	next := map[string]string{
		"/start": "/one",
		"/one":   "/two",
		"/two":   "/three",
	}
	var requested []string
	client := engine.NewClient(time.Second)
	client.Transport = redirectRoundTripper(func(req *http.Request) (*http.Response, error) {
		requested = append(requested, req.URL.Path)
		body := "continue"
		if req.URL.Path == "/two" {
			body = `{"error":"sql_error"}`
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Status:     "302 Found",
			Header: http.Header{
				"Location":             []string{next[req.URL.Path]},
				"X-Gateway-Diagnostic": []string{"available"},
			},
			Body:    io.NopCloser(strings.NewReader(body)),
			Request: req,
		}, nil
	})

	job := payload.Job{
		Method:    http.MethodGet,
		URL:       "http://target.test/start",
		Headers:   map[string]string{},
		Payload:   "' OR '1'='1",
		Class:     model.ClassSQLi,
		Endpoint:  "/users",
		ParamName: "id",
	}
	res := engine.Do(context.Background(), client, job)
	if res.Err != nil {
		t.Fatalf("redirect-limit response became a request error: %v", res.Err)
	}
	if res.StatusCode != http.StatusFound {
		t.Fatalf("status=%d, want %d", res.StatusCode, http.StatusFound)
	}
	if got := res.Header.Get("X-Gateway-Diagnostic"); got != "available" {
		t.Fatalf("diagnostic header=%q", got)
	}
	if got := string(res.Body); got != `{"error":"sql_error"}` {
		t.Fatalf("body=%q", got)
	}
	if want := []string{"/start", "/one", "/two"}; !reflect.DeepEqual(requested, want) {
		t.Fatalf("requested=%v, want %v", requested, want)
	}

	probe := fingerprint.FromJob(res.Job, res.StatusCode, res.Header, res.Body, res.Latency)
	if _, ok := fingerprint.New().Match(probe); !ok {
		t.Fatal("preserved redirect response did not reach fingerprint matching")
	}
}
