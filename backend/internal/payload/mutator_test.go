package payload

import (
	"testing"

	"github.com/alkaid/goapiscanner/internal/swagger"
)

func TestMutateGeneratesJobs(t *testing.T) {
	eps := []swagger.Endpoint{{
		Method: "GET",
		Path:   "/api/users",
		Params: []swagger.Param{{Name: "id", In: swagger.InQuery, Type: "string"}},
	}}
	jobs := Mutate(eps, MutateOptions{BaseURL: "http://lab", MaxJobs: 5000})
	if len(jobs) < 20 {
		t.Fatalf("too few jobs: %d", len(jobs))
	}
	seen := map[string]bool{}
	for _, j := range jobs {
		seen[string(j.Class)] = true
	}
	if !seen["sql_injection"] || !seen["xss"] {
		t.Fatalf("missing classes %#v", seen)
	}
}
