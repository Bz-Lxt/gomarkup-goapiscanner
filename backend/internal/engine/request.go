package engine

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/alkaid/goapiscanner/internal/payload"
)

type Result struct {
	Job        payload.Job
	StatusCode int
	Header     http.Header
	Body       []byte
	Latency    time.Duration
	Err        error
}

func Do(ctx context.Context, client *http.Client, job payload.Job) Result {
	start := time.Now()
	var rdr io.Reader
	if len(job.Body) > 0 {
		rdr = bytes.NewReader(job.Body)
	}
	req, err := http.NewRequestWithContext(ctx, job.Method, job.URL, rdr)
	if err != nil {
		return Result{Job: job, Err: err, Latency: time.Since(start)}
	}
	for k, v := range job.Headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "GoAPIScanner/1.0 (+authorized-lab)")
	}
	resp, err := client.Do(req)
	lat := time.Since(start)
	if err != nil {
		return Result{Job: job, Err: err, Latency: lat}
	}
	// Read the body BEFORE closing it. A previous version returned an
	// io.LimitReader wrapping resp.Body and closed resp.Body via defer when
	// returning from responseReader. Because Close() on a real
	// http.Response.Body invalidates the underlying reader, the caller's
	// subsequent io.ReadAll saw "http: read on closed body" and returned an
	// empty slice. That wiped out SQL-error fingerprints and reflected-XSS /
	// reflection hits which all depend on Result.Body. Read eagerly here so
	// the body is captured regardless of how the response stream behaves on
	// close.
	body, _ := readBody(resp)
	return Result{
		Job:        job,
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
		Latency:    lat,
	}
}

// readBody drains up to 256 KiB of the response body and closes the underlying
// stream. Reading must happen before Close() to avoid losing the body on
// transports whose Close() invalidates the reader (e.g. net/http transport).
func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 256<<10))
}
