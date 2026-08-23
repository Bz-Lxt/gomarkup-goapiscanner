package swagger

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Document is a loosely typed OpenAPI 2/3 tree. Validation is structural, not schema-complete.
type Document struct {
	Swagger     string                     `json:"swagger"`
	OpenAPI     string                     `json:"openapi"`
	BasePath    string                     `json:"basePath"`
	Servers     []map[string]any           `json:"servers"`
	Paths       map[string]json.RawMessage `json:"paths"`
	Definitions map[string]json.RawMessage `json:"definitions"`
	Components  *components                `json:"components"`
}

type components struct {
	Schemas map[string]json.RawMessage `json:"schemas"`
}

type ParseResult struct {
	Version   string
	BasePath  string
	Endpoints []Endpoint
}

var methods = []string{"get", "post", "put", "patch", "delete", "head", "options"}

func Parse(raw []byte) (*ParseResult, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("swagger: empty document")
	}
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("swagger: invalid json: %w", err)
	}
	if len(doc.Paths) == 0 {
		return nil, fmt.Errorf("swagger: missing paths object")
	}
	ver := doc.OpenAPI
	if ver == "" {
		ver = doc.Swagger
	}
	if ver == "" {
		return nil, fmt.Errorf("swagger: missing swagger/openapi version")
	}
	base := doc.BasePath
	if base == "" && len(doc.Servers) > 0 {
		if u, _ := doc.Servers[0]["url"].(string); u != "" {
			base = extractPath(u)
		}
	}
	out := &ParseResult{Version: ver, BasePath: base}
	for path, rawItem := range doc.Paths {
		if path == "" || rawItem == nil {
			continue
		}
		item, err := decodePathItem(rawItem)
		if err != nil {
			return nil, fmt.Errorf("swagger: path %s: %w", path, err)
		}
		shared := item.Parameters
		for _, m := range methods {
			opRaw, ok := item.Ops[m]
			if !ok || len(opRaw) == 0 {
				continue
			}
			ep, err := decodeOperation(m, JoinBase(base, path), opRaw, shared, doc)
			if err != nil {
				return nil, fmt.Errorf("swagger: %s %s: %w", m, path, err)
			}
			out.Endpoints = append(out.Endpoints, ep)
		}
	}
	if len(out.Endpoints) == 0 {
		return nil, fmt.Errorf("swagger: no operations found")
	}
	return out, nil
}

func extractPath(u string) string {
	u = strings.TrimSpace(u)
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
		if j := strings.Index(u, "/"); j >= 0 {
			return u[j:]
		}
		return ""
	}
	if strings.HasPrefix(u, "/") {
		return u
	}
	return ""
}
