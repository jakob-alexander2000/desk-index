// Package server exposes desk-index over HTTP.
package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jakob-alexander2000/desk-index/internal/index"
)

// Server wraps an Index with HTTP handlers.
type Server struct {
	Idx *index.Index
}

// Handler returns the HTTP mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /search", s.handleSearch)
	mux.HandleFunc("POST /reindex", s.handleReindex)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	docs, terms := s.Idx.Stats()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"root":  s.Idx.Root(),
		"docs":  docs,
		"terms": terms,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing query parameter q",
		})
		return
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "limit must be an integer 1–100",
			})
			return
		}
		limit = n
	}
	hits := s.Idx.Search(q, limit)
	if hits == nil {
		hits = []index.Hit{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"query": q,
		"hits":  hits,
	})
}

func (s *Server) handleReindex(w http.ResponseWriter, r *http.Request) {
	if err := s.Idx.Build(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	docs, terms := s.Idx.Stats()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"docs":  docs,
		"terms": terms,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
