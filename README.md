# desk-index

**A small Go search service for a folder of Markdown notes.** Point it at a folder and get ranked matches with a short excerpt from each. It finds passages without using a model, and works alongside [Desk](https://github.com/jakob-alexander2000/desk) and [CiteBot](https://github.com/jakob-alexander2000/citebot).

```
folder of .md  →  inverted index  →  GET /search?q=…
```

No database, no frameworks — stdlib HTTP + an in-memory index.

## Install / run

```bash
git clone https://github.com/jakob-alexander2000/desk-index.git
cd desk-index
go test -count=1 ./...
go run ./cmd/desk-index -dir ./sample-notes -addr 127.0.0.1:8080
```

## Demo

One-shot CLI search (no server):

```bash
go run ./cmd/desk-index search -dir ./sample-notes "What are core hours?"
```

Output:

```json
{
  "hits": [
    {
      "path": "core-hours.md",
      "title": "Core hours",
      "score": 39,
      "snippet": "# Core hours Core hours are **10:00–16:00** local time, Monday through Thursday. During core hours, be reachable in the team channel. Outside that window, a…"
    },
    {
      "path": "offsite.md",
      "title": "Team offsite",
      "score": 7.5,
      "snippet": "…lakeside meeting house. Plan: - Review what shipped in the last quarter - Set milestones for the next quarter - One afternoon with no laptops Travel to the of…"
    },
    {
      "path": "pto.md",
      "title": "Paid time off",
      "score": 3,
      "snippet": "…ic holidays. Request time off in the shared notes folder at least five weekdays ahead when the dates are known. A teammate confirms coverage before the dates a…"
    },
    {
      "path": "expense-policy.md",
      "title": "Expense policy",
      "score": 1.5,
      "snippet": "…thin 30 days. Meals at the team offsite are covered. Personal entertainment is not covered. A teammate who is not the spender approves each expense before it i…"
    }
  ],
  "query": "What are core hours?"
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
internal/index/     reads .md files into a word index; rarer words count more
internal/server/    JSON handlers
sample-notes/       sample notes for the demo
```

## Tests

```bash
go test -count=1 ./...
```

## License

MIT
