package engine_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/alkaid/goapiscanner/internal/engine"
	"github.com/alkaid/goapiscanner/internal/payload"
	"github.com/alkaid/goapiscanner/internal/swagger"
)

type recordingTransport struct {
	mu        sync.Mutex
	received  []string
	decodeErr error
}

func (t *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	defer r.Body.Close()
	var body map[string]string
	err := json.NewDecoder(r.Body).Decode(&body)
	t.mu.Lock()
	if err != nil && t.decodeErr == nil {
		t.decodeErr = err
	}
	t.received = append(t.received, body["query"])
	t.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    r,
	}, nil
}

func TestMutatedRequestBodiesRemainIndependent(t *testing.T) {
	jobs := payload.Mutate([]swagger.Endpoint{{
		Method: http.MethodPost,
		Path:   "/search",
		Params: []swagger.Param{{Name: "query", In: swagger.InBody, Type: "string"}},
	}}, payload.MutateOptions{BaseURL: "http://scanner-target.test", MaxJobs: 2})
	if len(jobs) != 2 {
		t.Fatalf("generated %d jobs, want 2", len(jobs))
	}
	want := map[string]int{}
	for _, job := range jobs {
		want[job.Payload]++
	}

	transport := &recordingTransport{}
	client := &http.Client{Transport: transport}
	for _, job := range jobs {
		if result := engine.Do(context.Background(), client, job); result.Err != nil {
			t.Fatal(result.Err)
		}
	}

	transport.mu.Lock()
	defer transport.mu.Unlock()
	if transport.decodeErr != nil {
		t.Fatalf("target received malformed JSON: %v", transport.decodeErr)
	}
	got := map[string]int{}
	for _, value := range transport.received {
		got[value]++
	}
	if len(transport.received) != len(jobs) {
		t.Fatalf("target received %d requests, want %d", len(transport.received), len(jobs))
	}
	for value, count := range want {
		if got[value] != count {
			t.Fatalf("target received payload %q %d times, want %d; all payloads: %v", value, got[value], count, transport.received)
		}
	}
}
