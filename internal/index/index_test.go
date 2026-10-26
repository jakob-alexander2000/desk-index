package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildAndSearch(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "roadmap.md"), "# Product priorities\n\nQ4 focus is applied AI tooling and agent workflows.\n")
	mustWrite(t, filepath.Join(root, "ops", "remote.md"), "# Remote work\n\nCore hours are 10:00–16:00 Berlin time.\n")
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

	hits := idx.Search("applied AI agent", 5)
	if len(hits) == 0 {
		t.Fatal("expected hits for applied AI agent")
	}
	if hits[0].Path != "roadmap.md" {
		t.Fatalf("top hit path=%q want roadmap.md", hits[0].Path)
	}
	if hits[0].Title != "Product priorities" {
		t.Fatalf("title=%q", hits[0].Title)
	}
	if hits[0].Snippet == "" {
		t.Fatal("expected snippet")
	}

	remote := idx.Search("Berlin core hours", 5)
	if len(remote) == 0 || remote[0].Path != "ops/remote.md" {
		t.Fatalf("remote hits=%v", remote)
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
