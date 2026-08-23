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
	body, _ := io.ReadAll(responseReader(resp))
	return Result{
		Job:        job,
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
		Latency:    lat,
	}
}

func responseReader(resp *http.Response) io.Reader {
	defer resp.Body.Close()
	return io.LimitReader(resp.Body, 256<<10)
}
