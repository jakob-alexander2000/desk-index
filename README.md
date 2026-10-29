# desk-index

**Tiny Go search API over a folder of Markdown notes.** Point it at a corpus, get ranked hits with snippets. Complements [Desk](https://github.com/jakob-alexander2000/desk) and [CiteBot](https://github.com/jakob-alexander2000/citebot): local retrieval without a model.

```
folder of .md  →  inverted index  →  GET /search?q=…
```

No database, no frameworks — stdlib HTTP + an in-memory index.

## Install / run

```bash
git clone https://github.com/jakob-alexander2000/desk-index.git
cd desk-index
go test ./...
go run ./cmd/desk-index -dir ./sample-notes -addr 127.0.0.1:8080
```

## Demo

One-shot CLI search (no server):

```bash
go run ./cmd/desk-index search -dir ./sample-notes "What are core hours?"
```

Expected shape:

```json
{
  "query": "What are core hours?",
  "hits": [
    {
      "path": "core-hours.md",
      "title": "Core hours",
      "score": …,
      "snippet": "…"
    }
  ]
}
```

HTTP:

```bash
# terminal A
go run ./cmd/desk-index -dir ./sample-notes

# terminal B
curl -s 'http://127.0.0.1:8080/health' | jq .
curl -s --get --data-urlencode 'q=What are core hours?' 'http://127.0.0.1:8080/search' | jq .
curl -s 'http://127.0.0.1:8080/search?q=paid+time+off' | jq .
curl -s 'http://127.0.0.1:8080/search?q=team+offsite' | jq .
curl -s 'http://127.0.0.1:8080/search?q=expense+policy' | jq .
curl -s -X POST 'http://127.0.0.1:8080/reindex' | jq .
```

### Endpoints

| Method | Path | Notes |
|--------|------|--------|
| `GET` | `/health` | docs + term counts |
| `GET` | `/search?q=&limit=` | ranked hits (limit default 10, max 100) |
| `POST` | `/reindex` | rebuild from `-dir` |

## Layout

```
cmd/desk-index/     CLI + HTTP server
internal/index/     walk .md → TF-IDF-ish inverted index
internal/server/    JSON handlers
sample-notes/       tiny demo corpus
```

## Tests

```bash
go test ./...
```

## License

MIT
