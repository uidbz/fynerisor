# 36-tie-headless

Headless (no GUI) example demonstrating the tie triple store module.

## Requirements

**This example requires a running tie-triplestore.** Start one from a checkout of
[tie](https://github.com/uidbz/tie):

```bash
# Option 1: use the test environment — listens on http://localhost:2161
./test-env/build.sh
./test-env/start.sh

# Option 2: run the server directly — listens on http://localhost:1161
go run ./cmd/tie-triplestore
```

Mind which one you started — the two options listen on **different ports**.
`app.risor` connects to `1161`, so with the test environment change that URL to
`2161` first.

## What it demonstrates

- Connecting to a tie daemon with `tie.connect(url, opts)`
- Adding triples with `db.add(key, relation, value)`
- Syncing writes with `db.sync()`
- Getting a key's attributes with `db.get(key)`
- Reverse association queries with `db.query({terms, reverse: true})`
  - Query returns `{rows: [...], total_count: N}` for pagination support
- Batch operations: `batch.add/set/run()`
- Expanding multiple keys in one call with `db.expand([keys])`
- Checking existence with `db.exists(key)`
- Updating triples with `db.update(key, relation, old, new)`
- **Streaming**: `db.dump_stream(callback)` for memory-efficient dumps of large collections

## Run

```bash
go run main.go
```

## Data Model

Tie stores **triples**: `(key, relation, value)`. For example, `rust-guide | tag | programming` means "rust-guide has tag programming". 

- **Forward associations**: `get(key)` returns what relations and values a key has
- **Reverse associations**: `query({terms: ["value"], reverse: true})` finds which keys have that value

**Note**: Reverse queries only work for relations in the daemon's `ReverseRelations` config. The relation `"tag"` is in the default list, which is why this example uses tags. Other relations like `"author"` won't work for reverse queries without explicit configuration.
