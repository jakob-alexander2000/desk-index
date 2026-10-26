// Package index builds a simple inverted index over Markdown files on disk.
package index

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
)

var wordRe = regexp.MustCompile(`[a-z0-9]+`)

// Doc is one indexed Markdown file.
type Doc struct {
	Path     string
	Title    string
	Content  string
	TermFreq map[string]int
}

// Hit is a ranked search result.
type Hit struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
}

// Index is a thread-safe in-memory Markdown search index.
type Index struct {
	mu   sync.RWMutex
	root string
	docs []Doc
	post map[string][]int
}

// New creates an empty index rooted at dir.
func New(root string) *Index {
	return &Index{
		root: root,
		post: make(map[string][]int),
	}
}

// Root returns the corpus directory.
func (idx *Index) Root() string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.root
}

// Stats returns document and unique-term counts.
func (idx *Index) Stats() (docs, terms int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs), len(idx.post)
}

// Build walks root for .md files and rebuilds the index.
func (idx *Index) Build() error {
	root := idx.root
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("corpus: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("corpus: not a directory: %s", root)
	}

	var docs []Doc
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		content := string(raw)
		docs = append(docs, Doc{
			Path:     rel,
			Title:   titleFrom(rel, content),
			Content: content,
			TermFreq: termFreq(content),
		})
		return nil
	})
	if err != nil {
		return err
	}

	post := make(map[string][]int)
	for i, doc := range docs {
		seen := make(map[string]struct{})
		for term := range doc.TermFreq {
			if _, ok := seen[term]; ok {
				continue
			}
			seen[term] = struct{}{}
			post[term] = append(post[term], i)
		}
	}

	idx.mu.Lock()
	idx.docs = docs
	idx.post = post
	idx.mu.Unlock()
	return nil
}

// Search returns up to limit ranked hits for query.
func (idx *Index) Search(query string, limit int) []Hit {
	if limit <= 0 {
		limit = 10
	}
	terms := tokenize(query)
	if len(terms) == 0 {
		return nil
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	scores := make(map[int]float64)
	nDocs := float64(len(idx.docs))
	for _, term := range terms {
		df := float64(len(idx.post[term]))
		if df == 0 {
			continue
		}
		idf := nDocs / df
		for _, di := range idx.post[term] {
			tf := float64(idx.docs[di].TermFreq[term])
			scores[di] += tf * idf
		}
	}

	type scored struct {
		i int
		s float64
	}
	ranked := make([]scored, 0, len(scores))
	for i, s := range scores {
		ranked = append(ranked, scored{i: i, s: s})
	}
	sort.Slice(ranked, func(a, b int) bool {
		if ranked[a].s == ranked[b].s {
			return idx.docs[ranked[a].i].Path < idx.docs[ranked[b].i].Path
		}
		return ranked[a].s > ranked[b].s
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	hits := make([]Hit, 0, len(ranked))
	for _, r := range ranked {
		doc := idx.docs[r.i]
		hits = append(hits, Hit{
			Path:    doc.Path,
			Title:   doc.Title,
			Score:   r.s,
			Snippet: snippet(doc.Content, terms),
		})
	}
	return hits
}

func tokenize(s string) []string {
	raw := wordRe.FindAllString(strings.ToLower(s), -1)
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{})
	for _, t := range raw {
		if len(t) < 2 {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func termFreq(content string) map[string]int {
	tf := make(map[string]int)
	for _, t := range wordRe.FindAllString(strings.ToLower(content), -1) {
		if len(t) < 2 {
			continue
		}
		tf[t]++
	}
	return tf
}

func titleFrom(rel, content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := filepath.Base(rel)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func snippet(content string, terms []string) string {
	lower := strings.ToLower(content)
	pos := -1
	for _, t := range terms {
		if i := strings.Index(lower, t); i >= 0 {
			if pos < 0 || i < pos {
				pos = i
			}
		}
	}
	if pos < 0 {
		pos = 0
	}
	start := pos - 40
	if start < 0 {
		start = 0
	}
	end := start + 160
	if end > len(content) {
		end = len(content)
	}
	s := collapseSpace(strings.TrimSpace(content[start:end]))
	if start > 0 {
		s = "…" + s
	}
	if end < len(content) {
		s = s + "…"
	}
	return s
}

func collapseSpace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}
