package engine_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/alkaid/goapiscanner/internal/engine"
	"github.com/alkaid/goapiscanner/internal/payload"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type trackedBody struct {
	reader *strings.Reader
	closed bool
}

func (b *trackedBody) Read(p []byte) (int, error) {
	if b.closed {
		return 0, errors.New("read after close")
	}
	return b.reader.Read(p)
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func TestDoReturnsResponseBody(t *testing.T) {
	const responseBody = `{"message":"probe reflected","token":"scanner-marker"}`
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       &trackedBody{reader: strings.NewReader(responseBody)},
		}, nil
	})}

	result := engine.Do(context.Background(), client, payload.Job{
		Method:  http.MethodGet,
		URL:     "http://scanner.test/probe",
		Headers: map[string]string{},
	})
	if result.Err != nil {
		t.Fatalf("Do returned error: %v", result.Err)
	}
	if got := string(result.Body); got != responseBody {
		t.Fatalf("response body = %q, want %q", got, responseBody)
	}
}
