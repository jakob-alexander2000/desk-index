package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jakob-alexander2000/desk-index/internal/index"
	"github.com/jakob-alexander2000/desk-index/internal/server"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) > 1 && os.Args[1] == "search" {
		os.Exit(runSearch(os.Args[2:]))
	}

	dir := flag.String("dir", "./sample-notes", "folder of Markdown files to index")
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	idx := index.New(*dir)
	if err := idx.Build(); err != nil {
		log.Fatalf("index: %v", err)
	}
	docs, terms := idx.Stats()
	log.Printf("indexed %d docs (%d terms) from %s", docs, terms, *dir)

	srv := &server.Server{Idx: idx}
	log.Printf("listening on http://%s", *addr)
	log.Printf("  GET  /health")
	log.Printf("  GET  /search?q=...")
	log.Printf("  POST /reindex")
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	dir := fs.String("dir", "./sample-notes", "folder of Markdown files")
	limit := fs.Int("limit", 5, "max hits")
	_ = fs.Parse(args)
	query := fs.Arg(0)
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: desk-index search [-dir DIR] [-limit N] QUERY")
		return 2
	}
	idx := index.New(*dir)
	if err := idx.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "index: %v\n", err)
		return 1
	}
	hits := idx.Search(query, *limit)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{"query": query, "hits": hits})
	return 0
}
