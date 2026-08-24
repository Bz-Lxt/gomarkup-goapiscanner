package scan

import (
	"testing"

	"github.com/alkaid/goapiscanner/internal/config"
)

func TestGuardReentry(t *testing.T) {
	g := NewGuard()
	if !g.Acquire("a") {
		t.Fatal("first acquire")
	}
	if g.Acquire("a") {
		t.Fatal("reentry must fail")
	}
	g.Release("a")
	if !g.Acquire("a") {
		t.Fatal("after release")
	}
}

func TestResolveTargetLabOnly(t *testing.T) {
	cfg := config.Config{
		LabPublicURL:   "http://localhost:28483",
		LabInternalURL: "http://target-lab:8090",
		ScanMode:       "lab",
	}
	got, err := ResolveTarget("http://localhost:28483", cfg)
	if err != nil || got != "http://target-lab:8090" {
		t.Fatalf("got=%s err=%v", got, err)
	}
	if _, err := ResolveTarget("http://evil.example", cfg); err == nil {
		t.Fatal("lab mode must reject foreign host")
	}
	cfg.ScanMode = "authorized"
	if _, err := ResolveTarget("http://evil.example", cfg); err != nil {
		t.Fatal(err)
	}
}
