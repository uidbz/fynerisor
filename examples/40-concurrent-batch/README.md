# 40-concurrent-batch

Headless (no GUI) example demonstrating **parallel batch execution** with
`core.EvalBatch`. Uses only `core` — no Fyne/GUI dependencies.

## What it demonstrates

The "compile once, run many" pattern: a single Risor script is compiled once
into immutable bytecode, then run over many inputs **concurrently**, each on its
own isolated VM. This gives real multi-core throughput for headless data work
(processing many files, records, or requests) without any GUI-thread concerns.

- `core.EvalBatch(ctx, script, inputs, cfg, opts...)` — compile once, run each
  input on a fresh, isolated `Context`+VM with bounded concurrency.
- Each input map is exposed to the script as global variables. Here each input
  is `{"file": "data/xxx.csv"}`, so the script reads the global `file`.
- Results are **index-aligned** with inputs. Per-input failures are reported in
  `BatchResult.Err`, so one bad input does not fail the whole batch; the returned
  error is only for setup failures such as compilation.
- `BatchConfig{Concurrency: 4}` bounds how many inputs run at once (defaults to
  `GOMAXPROCS` when `<= 0`).

`worker.risor` is the unit of work: it reads the CSV file named by `file`, sums
the `amount` column, and returns a summary map.

## Isolation & safety

Each worker builds its own `Context`, so its module objects (`csv`, ...), module
cache, and import stack are independent — no shared mutable interpreter state.
Stateless modules like `csv`/`json`/`strings` are safe to use across workers.
Values passed via `WithGlobal(name, value)` are shared by pointer; if such a
value is stateful, the caller must make it concurrency-safe.

`EvalBatch` is headless only — do not use it to drive Fyne widgets. See
`docs/CONCURRENCY.md` for the full model and for why a script-level `parallel()`
builtin was deliberately not added.

## Run

```bash
cd examples/40-concurrent-batch
go run main.go
```

Expected output (order-stable, one line per file):

```
data/east.csv    map[rows:4 total:400]
data/north.csv   map[rows:3 total:425]
data/south.csv   map[rows:2 total:625]
data/west.csv    map[rows:2 total:355]
```
