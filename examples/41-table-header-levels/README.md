# 41-table-header-levels

A table with a two-row (hierarchical) header, read out of a tie collection:
`Temperature (20°C)` and `Temperature (37°C)` each spanning two replicate columns.

## Requirements

**A running tie triplestore**, with the namespace/collection this script connects
to (`testing`/`testing` on `http://localhost:2161` — the tie repo's `test-env`):

```bash
cd /home/johan/go/src/github/tie
./test-env/build.sh
./test-env/start.sh
```

## What it demonstrates

- `db.insert_table` with header **rows** instead of labels, and `db.read_table`
  returning them as `header_levels`
- `table.HeaderLevels(fn)` rendering those rows above the columns, drawing a parent
  once across the columns it covers rather than repeating it
- `table.Columns(fn)` still carrying the column keys, so header-click sorting and
  export are unaffected by the levels
- The **Toggle header rows** button switching the levels off, which falls back to
  one header row of `\x1f`-joined keys — what the same table looked like before
  header levels existed, and why they are worth having

## Run

```bash
go run main.go
```
