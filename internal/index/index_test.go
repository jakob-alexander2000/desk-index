package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildAndSearch(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "core-hours.md"), "# Core hours\n\nCore hours are 10:00–16:00 local time, Monday through Thursday.\n")
	mustWrite(t, filepath.Join(root, "ops", "pto.md"), "# Paid time off\n\nFull-time staff receive 25 days of paid time off each calendar year.\n")
	mustWrite(t, filepath.Join(root, "readme.txt"), "ignored non-markdown")

	idx := New(root)
	if err := idx.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	docs, terms := idx.Stats()
	if docs != 2 {
		t.Fatalf("docs=%d want 2", docs)
	}
	if terms < 5 {
		t.Fatalf("terms=%d want >=5", terms)
	}

	hits := idx.Search("What are core hours?", 5)
	if len(hits) == 0 {
		t.Fatal("expected hits for core hours")
	}
	if hits[0].Path != "core-hours.md" {
		t.Fatalf("top hit path=%q want core-hours.md", hits[0].Path)
	}
	if hits[0].Title != "Core hours" {
		t.Fatalf("title=%q", hits[0].Title)
	}
	if hits[0].Snippet == "" {
		t.Fatal("expected snippet")
	}

	pto := idx.Search("paid time off", 5)
	if len(pto) == 0 || pto[0].Path != "ops/pto.md" {
		t.Fatalf("pto hits=%v", pto)
	}
	if pto[0].Title != "Paid time off" {
		t.Fatalf("title=%q", pto[0].Title)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	idx := New(t.TempDir())
	_ = idx.Build()
	if hits := idx.Search("  ", 5); hits != nil {
		t.Fatalf("want nil, got %v", hits)
	}
}

func TestBuildMissingDir(t *testing.T) {
	idx := New(filepath.Join(t.TempDir(), "nope"))
	if err := idx.Build(); err == nil {
		t.Fatal("expected error")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
