package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jakob-alexander2000/desk-index/internal/index"
)

func TestHealthSearchReindex(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "notes.md"), "# Notes\n\nDesk index finds markdown passages fast.\n")

	idx := index.New(root)
	if err := idx.Build(); err != nil {
		t.Fatal(err)
	}
	srv := &Server{Idx: idx}
	h := srv.Handler()

	t.Run("health", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["ok"] != true {
			t.Fatalf("body=%v", body)
		}
		if body["docs"].(float64) != 1 {
			t.Fatalf("docs=%v", body["docs"])
		}
	})

	t.Run("search", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/search?q=markdown", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		var body struct {
			Hits []index.Hit `json:"hits"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Hits) != 1 || body.Hits[0].Path != "notes.md" {
			t.Fatalf("hits=%v", body.Hits)
		}
	})

	t.Run("search missing q", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/search", nil))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rr.Code)
		}
	})

	t.Run("reindex", func(t *testing.T) {
		mustWrite(t, filepath.Join(root, "second.md"), "# Second\n\nAnother file about retrieval.\n")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/reindex", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		docs, _ := idx.Stats()
		if docs != 2 {
			t.Fatalf("docs=%d after reindex", docs)
		}
	})
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
