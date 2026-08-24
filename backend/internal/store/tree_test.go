package store

import (
	"testing"

	"github.com/alkaid/goapiscanner/internal/model"
)

func TestBuildTreeAndStats(t *testing.T) {
	fs := []model.Finding{
		{ID: "1", Method: "GET", Endpoint: "/api/users", Severity: model.SeverityCritical, Title: "sqli"},
		{ID: "2", Method: "GET", Endpoint: "/api/users", Severity: model.SeverityHigh, Title: "x"},
		{ID: "3", Method: "GET", Endpoint: "/api/search", Severity: model.SeverityHigh, Title: "xss"},
	}
	tree := BuildTree(fs)
	if len(tree) != 2 {
		t.Fatalf("tree=%d", len(tree))
	}
	if tree[0].Severity != model.SeverityCritical || tree[0].Count != 2 {
		t.Fatalf("%+v", tree[0])
	}
	st := StatsOf(fs)
	if st.Critical != 1 || st.High != 2 {
		t.Fatalf("%+v", st)
	}
	if BuildTree(nil) == nil {
		t.Fatal("empty tree must be non-nil slice")
	}
}
