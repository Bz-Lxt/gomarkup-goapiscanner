package payload

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/swagger"
)

type Job struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	Payload   string
	Class     model.VulnClass
	Endpoint  string
	ParamName string
	Timeout   bool
}

type MutateOptions struct {
	BaseURL string
	MaxJobs int
}

func Mutate(eps []swagger.Endpoint, opts MutateOptions) []Job {
	items := Catalog()
	var jobs []Job
	base := strings.TrimRight(opts.BaseURL, "/")
	for _, ep := range eps {
		for _, p := range ep.Params {
			for _, it := range items {
				if !compatible(p, it) {
					continue
				}
				j := apply(base, ep, p, it)
				jobs = append(jobs, j)
				if opts.MaxJobs > 0 && len(jobs) >= opts.MaxJobs {
					return jobs
				}
			}
		}
	}
	return jobs
}

func compatible(p swagger.Param, it Item) bool {
	if it.Class == model.ClassUnauth {
		return true
	}
	if p.In == swagger.InHeader && !strings.EqualFold(p.Name, "authorization") {
		return it.Class == model.ClassSQLi || it.Class == model.ClassXSS
	}
	return true
}

func apply(base string, ep swagger.Endpoint, p swagger.Param, it Item) Job {
	path := ep.Path
	q := url.Values{}
	headers := map[string]string{"Accept": "application/json, text/html;q=0.8"}
	bodyObj := map[string]any{}

	inject := func(target swagger.Param, val string) {
		switch target.In {
		case swagger.InPath:
			path = strings.ReplaceAll(path, "{"+target.Name+"}", url.PathEscape(val))
			path = strings.ReplaceAll(path, ":"+target.Name, url.PathEscape(val))
		case swagger.InQuery:
			q.Set(target.Name, val)
		case swagger.InHeader:
			headers[target.Name] = val
		default:
			bodyObj[target.Name] = val
		}
	}

	for _, other := range ep.Params {
		if other.Name == p.Name && other.In == p.In {
			continue
		}
		inject(other, baseline(other))
	}

	if it.Class == model.ClassUnauth {
		delete(headers, "Authorization")
		delete(headers, "authorization")
		if it.Value != "" {
			headers["Authorization"] = it.Value
		}
	} else {
		inject(p, it.Value)
	}

	var body []byte
	if len(bodyObj) > 0 {
		body, _ = json.Marshal(bodyObj)
		headers["Content-Type"] = "application/json"
	}
	u := base + path
	if enc := q.Encode(); enc != "" {
		u += "?" + enc
	}
	return Job{
		Method:    ep.Method,
		URL:       u,
		Headers:   headers,
		Body:      body,
		Payload:   it.Value,
		Class:     it.Class,
		Endpoint:  ep.Path,
		ParamName: p.Name,
		Timeout:   it.Timeout,
	}
}

func baseline(p swagger.Param) string {
	switch p.Type {
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	default:
		if p.In == swagger.InPath {
			return "1"
		}
		return "ok"
	}
}
