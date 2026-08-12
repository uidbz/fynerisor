# 36-tie-headless

Headless (no GUI) example demonstrating the tie triple store module.

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
- **Streaming**: `db.dump_stream(callback)` for memory-efficient dumps (see `dump-stream-test.risor`)

## Run

```bash
# Main example
go run main.go

# Streaming dump example
go run main.go dump-stream-test.risor
```

## Data Model

Tie stores **triples**: `(key, relation, value)`. For example, `pizza | topping | cheese` means "pizza has topping cheese". A key's forward associations (what relations and values it has) come from `get()`. Reverse associations (which keys point *to* a value) come from `query({reverse: true})`.
