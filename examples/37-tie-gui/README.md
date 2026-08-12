# 37-tie-gui

GUI example demonstrating the tie triple store module with a Fyne interface.

## Requirements

**This example requires a running tie-daemon.** Start one from the tie repository:

```bash
cd /home/johan/go/src/sourcehut/tie
# Option 1: use the test environment
./test-env/build.sh
./test-env/start.sh

# Option 2: run the daemon directly
go run ./cmd/tie-daemon
```

The daemon listens on `http://localhost:1161` by default.

## What it demonstrates

- A complete tie triple browser GUI built with Fyne
- Connecting to a tie daemon
- **Reverse query**: search for a value (e.g., "cheese") and find all keys that have it
  - Status bar shows total count from `result.total_count` (useful for pagination)
- **Add triples**: form to add `key | relation | value` triples
- **Table display**: showing query results with key and attributes
- Real-time updates when adding new triples

## Run

```bash
go run main.go
```

## How to use

1. The app seeds sample documents on startup (programming guides, personal notes, project docs with various tags)
2. **Query**: enter a tag like `programming` and click "Search" — you'll see all documents that have that tag (reverse association lookup)
3. **Add**: fill in the document name, relation (defaults to "tag"), and tag value (e.g., `my-doc | tag | urgent`) and click "Add"
4. Query again to see your new data appear

## Data Model

Tie stores **triples**: `(key, relation, value)`. This GUI performs **reverse queries** — given a tag value, find which documents have it. For example, querying for `programming` returns `rust-guide`, `go-concurrency`, and `python-async` because they all have `tag | programming`.

**Important**: Reverse queries only work for relations in the daemon's `ReverseRelations` config. The relation `"tag"` is in the default list, which is why this example uses tags. If you add triples with other relations (like `author`), they won't appear in reverse queries unless you configure the daemon.
