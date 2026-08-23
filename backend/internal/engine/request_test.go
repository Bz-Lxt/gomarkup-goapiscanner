package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkaid/goapiscanner/internal/fingerprint"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/payload"
)

// TestDoCapturesBody guards against a regression where Result.Body was always
// empty: responseReader closed resp.Body via defer before the caller read it,
// so io.ReadAll returned "http: read on closed body". This broke every
// body-based fingerprint (SQL errors, reflected XSS, traversal, commandi).
func TestDoCapturesBody(t *testing.T) {
	const want = `{"sql_error":"You have an error in your SQL syntax"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(want))
	}))
	defer srv.Close()

	res := Do(context.Background(), http.DefaultClient, payload.Job{Method: "GET", URL: srv.URL + "/x"})
	if res.Err != nil {
		t.Fatalf("unexpected err: %v", res.Err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", res.StatusCode, http.StatusOK)
	}
	if string(res.Body) != want {
		t.Fatalf("body: got %q (len=%d) want %q", string(res.Body), len(res.Body), want)
	}
	if res.Header.Get("Content-Type") == "" {
		t.Fatalf("header missing")
	}
}

// TestDoFeedsSQLFingerprintThroughMatcher verifies the real detection path
// end-to-end: Do -> Result.Body -> FromJob -> Matcher.Match hits the SQLi rule.
func TestDoFeedsSQLFingerprintThroughMatcher(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"err":"You have an error in your SQL syntax near "'"}`))
	}))
	defer srv.Close()

	m := fingerprint.New()
	res := Do(context.Background(), http.DefaultClient, payload.Job{
		Method:    "GET",
		URL:       srv.URL + "/q?id=1'",
		Class:     model.ClassSQLi,
		Payload:   "1'",
		ParamName: "id",
		Endpoint:  "/q",
	})
	if res.Err != nil {
		t.Fatalf("unexpected err: %v", res.Err)
	}
	probe := fingerprint.FromJob(res.Job, res.StatusCode, res.Header, res.Body, res.Latency)
	hit, ok := m.Match(probe)
	if !ok || hit.Class != model.ClassSQLi {
		t.Fatalf("expected SQLi hit, got ok=%v hit=%+v", ok, hit)
	}
}

// TestDoFeedsReflectionFingerprintThroughMatcher verifies reflected-XSS
// detection, which depends entirely on Result.Body containing the payload.
func TestDoFeedsReflectionFingerprintThroughMatcher(t *testing.T) {
	const xss = "<script>alert(1)</script>"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Results for: " + xss + "</body></html>"))
	}))
	defer srv.Close()

	m := fingerprint.New()
	res := Do(context.Background(), http.DefaultClient, payload.Job{
		Method:    "GET",
		URL:       srv.URL + "/search?q=" + xss,
		Class:     model.ClassXSS,
		Payload:   xss,
		ParamName: "q",
		Endpoint:  "/search",
	})
	if res.Err != nil {
		t.Fatalf("unexpected err: %v", res.Err)
	}
	probe := fingerprint.FromJob(res.Job, res.StatusCode, res.Header, res.Body, res.Latency)
	hit, ok := m.Match(probe)
	if !ok || hit.Class != model.ClassXSS {
		t.Fatalf("expected XSS hit, got ok=%v hit=%+v", ok, hit)
	}
}
