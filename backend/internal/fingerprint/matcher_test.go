package fingerprint

import (
	"net/http"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/model"
)

func TestMatcherHitsAndDedup(t *testing.T) {
	m := New()
	p := Probe{
		Class:      model.ClassSQLi,
		Payload:    "' OR 1=1--",
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       `{"sql_error":"You have an error in your SQL syntax"}`,
		Endpoint:   "/api/users",
		Method:     "GET",
		ParamName:  "id",
	}
	hit, ok := m.Match(p)
	if !ok || hit.Severity != model.SeverityCritical {
		t.Fatalf("expected hit, got %+v ok=%v", hit, ok)
	}
	if _, ok := m.Match(p); ok {
		t.Fatal("duplicate should be suppressed")
	}
}

func TestTimingRule(t *testing.T) {
	m := New()
	p := Probe{
		Class:      model.ClassTimeBlind,
		Payload:    "1' AND SLEEP(3)--",
		StatusCode: 200,
		Body:       `{"ok":true}`,
		Latency:    3100 * time.Millisecond,
		Endpoint:   "/api/users/blind",
		Method:     "GET",
		ParamName:  "id",
	}
	if _, ok := m.Match(p); !ok {
		t.Fatal("expected timing hit")
	}
	fast := p
	fast.Latency = 20 * time.Millisecond
	fast.ParamName = "id2"
	if _, ok := m.Match(fast); ok {
		t.Fatal("fast response should not hit")
	}
}

func TestXSSNeedsReflection(t *testing.T) {
	m := New()
	payload := "<script>alert(1)</script>"
	p := Probe{
		Class:      model.ClassXSS,
		Payload:    payload,
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       "<html><body>Results for: " + payload + "</body></html>",
		Endpoint:   "/api/search",
		Method:     "GET",
		ParamName:  "q",
	}
	if _, ok := m.Match(p); !ok {
		t.Fatal("expected xss")
	}
}

func TestIsTimingAnomaly(t *testing.T) {
	if IsTimingAnomaly(time.Second) {
		t.Fatal("1s should not be anomaly")
	}
	if !IsTimingAnomaly(3 * time.Second) {
		t.Fatal("3s should be anomaly")
	}
}
