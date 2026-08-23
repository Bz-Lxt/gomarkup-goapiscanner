package scan

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func FetchSwagger(ctx context.Context, client *http.Client, base string) ([]byte, string, error) {
	candidates := []string{
		base + "/swagger.json",
		base + "/openapi.json",
		base + "/api/swagger.json",
		base + "/v2/api-docs",
	}
	var last error
	for _, u := range candidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			last = err
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			last = err
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if resp.StatusCode != 200 {
			last = fmt.Errorf("%s -> HTTP %d", u, resp.StatusCode)
			continue
		}
		if !looksLikeSpec(body) {
			last = fmt.Errorf("%s: not an OpenAPI document", u)
			continue
		}
		return body, u, nil
	}
	if last == nil {
		last = fmt.Errorf("no swagger candidate")
	}
	return nil, "", last
}

func looksLikeSpec(b []byte) bool {
	s := strings.ToLower(string(b))
	return strings.Contains(s, `"paths"`) && (strings.Contains(s, `"swagger"`) || strings.Contains(s, `"openapi"`))
}

func ProbeClient(timeout time.Duration) *http.Client {
	c := &http.Client{Timeout: timeout}
	return c
}
