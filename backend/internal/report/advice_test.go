package report

import (
	"testing"

	"github.com/alkaid/goapiscanner/internal/model"
)

func TestAdviceListDedup(t *testing.T) {
	fs := []model.Finding{
		{Class: model.ClassSQLi},
		{Class: model.ClassSQLi},
		{Class: model.ClassXSS},
	}
	got := AdviceList(fs)
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	empty := AdviceList(nil)
	if len(empty) != 1 {
		t.Fatal(empty)
	}
}
