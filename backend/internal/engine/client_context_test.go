package engine

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alkaid/goapiscanner/internal/payload"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDoCancelStopsRedirectedRequest(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	defer close(release)

	client := NewClient(2 * time.Second)
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/start":
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{"/blocked"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		case "/blocked":
			started <- struct{}{}
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-release:
				return nil, errors.New("request released by test")
			}
		default:
			return nil, errors.New("unexpected request path")
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Result, 1)
	go func() {
		done <- Do(ctx, client, payload.Job{Method: http.MethodGet, URL: "http://target.test/start"})
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("redirected request did not reach target")
	}
	cancel()

	select {
	case res := <-done:
		if !errors.Is(res.Err, context.Canceled) {
			t.Fatalf("Do returned err=%v, want context.Canceled", res.Err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("redirected request remained active after cancellation")
	}
}
