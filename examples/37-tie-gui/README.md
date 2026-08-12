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

1. The app seeds some sample data on startup (pizza, burger, taco with various toppings)
2. **Query**: enter a value like `cheese` and click "Search" — you'll see all items that have cheese as a value (reverse association lookup)
3. **Add**: fill in the key/relation/value fields (e.g., `sushi | topping | avocado`) and click "Add"
4. Query again to see your new data appear

## Data Model

Tie stores **triples**: `(key, relation, value)`. This GUI performs **reverse queries** — given a value, find which keys are associated with it. For example, querying for "cheese" returns `pizza`, `burger`, and `taco` because they all have `topping | cheese`.
