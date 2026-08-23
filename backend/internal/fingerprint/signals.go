package fingerprint

import (
	"net/http"
	"strings"
	"time"

	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/payload"
)

type Probe struct {
	Class      model.VulnClass
	Payload    string
	StatusCode int
	Header     http.Header
	Body       string
	Latency    time.Duration
	URL        string
	Method     string
	ParamName  string
	Endpoint   string
}

func FromJob(job payload.Job, status int, header http.Header, body []byte, latency time.Duration) Probe {
	return Probe{
		Class:      job.Class,
		Payload:    job.Payload,
		StatusCode: status,
		Header:     header,
		Body:       string(body),
		Latency:    latency,
		URL:        job.URL,
		Method:     job.Method,
		ParamName:  job.ParamName,
		Endpoint:   job.Endpoint,
	}
}

func headerGet(h http.Header, key string) string {
	if h == nil {
		return ""
	}
	return h.Get(key)
}

func bodyHasAny(body string, needles []string) string {
	low := strings.ToLower(body)
	for _, n := range needles {
		if n == "" {
			continue
		}
		if strings.Contains(body, n) || strings.Contains(low, strings.ToLower(n)) {
			return n
		}
	}
	return ""
}

func payloadHasAny(payload string, needles []string) bool {
	up := strings.ToUpper(payload)
	for _, n := range needles {
		if strings.Contains(up, strings.ToUpper(n)) {
			return true
		}
	}
	return false
}
