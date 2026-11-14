# v3scene: a Risor v3 scene-page prototype

Status: experiment. It is not imported by fynerisor's main module and has not
been reviewed for production use. It exists to answer the questions in
[deepnoodle-ai/risor#641](https://github.com/deepnoodle-ai/risor/issues/641)
with working code.

## The model

`v3scene` is a **host-dispatched event runtime**, the model the Risor author
confirmed in #641:

```
Fyne event
    -> View.On listener (UI thread)
    -> event queue
    -> App.Step (one goroutine)
        -> Program.Run({state, event})      immutable v3 program, compile once
        -> result {state, actions}
        -> apply actions on the UI thread
```

There are no script callbacks. A script cannot hand a closure to the host,
because v3 does not allow it. Instead, every interaction is an **event** and
every effect is a **return value**. The host owns the state; the program takes
a snapshot and returns the next snapshot.

The v3 program is checked at load against these types:

```risor
type Customer = {name: string, city: string, active: bool}
type State    = {count: int, name: string, city: string,
                 customers: Customer[], log: string[]}
type Event    = {type: string, node?: string,
                 values?: {name?: string, city?: string, ...},
                 data?: Customer[], ...}
type Action   = {do: "set", node: string, prop: string, value: any}
              | {do: "show", node: string}
              | {do: "hide", node: string}
              | {do: "toggle", node: string}
              | {do: "rows", node: string, rows: string[][]}
              | {do: "append", node: string, text: string}
              | {do: "status", text: string}
              | {do: "fetch", node: string, target: string}
              | {do: "save", customer: Customer}
type Result   = {state: State, actions: Action[]}
```

A misspelled field, a wrong node property or an incomplete action arm is a
**check error before the page renders**. The tests below prove it.

## Two effects models

One page can use either model. The mode is explicit in `Config.Mode`; the
runtime does not mix them in one demo.

| | `ModeActions` | `ModeHostFunctions` |
|---|---|---|
| Read | program returns `{do: "fetch"}`; host does I/O and sends a `loaded` event | program calls `store.customers()` during the run |
| Write | program returns `{do: "save", customer}`; host writes | program calls `store.add(name, city)` during the run |
| Blocking | no (the run ends until the host resumes it) | yes (the run waits for the host function) |
| Failure | an action list is applied in full, or not at all | a host function can mutate before a later failure |
| Testing | give state + event, assert the returned actions | give state + event, assert state after the host calls |
| Replay | possible: host calls are data | possible only if the host journals calls |
| Code | `testdata/customers/actions.risor` (54 lines) | `testdata/customers/host-functions.risor` (48 lines) |

Recommendation from the prototype: **actions for writes and slow/remote I/O;
host functions for cheap reads.** Mixing both in one page is possible at the
host level, but it makes the failure model harder to explain; the mode switch
is deliberately per page.

## Run it

```sh
cd experiments/v3scene

# Tests: Fyne test driver, no window needed.
go test ./...

# Headless check (no Fyne app needed): compile the page and exit.
go run ./cmd/v3scene -check \
  -page testdata/customers/customers.scene.json \
  -script testdata/customers/actions.risor -mode actions

# Interactive.
go run ./cmd/v3scene \
  -page testdata/customers/customers.scene.json \
  -script testdata/customers/actions.risor -mode actions

# Same page and host, other effects model.
go run ./cmd/v3scene \
  -page testdata/customers/customers.scene.json \
  -script testdata/customers/host-functions.risor -mode host-functions
```

## What works

- **The page and the program are independent.** The document is the existing
  `fynerisor.scene/1` format; the program is a v3 file. No change was needed
  in `scene`, `scene/render` or the Fyne layer.
- **Compile once, run per event.** `Program` is immutable; the host does not
  need `vmexec`, a VM goroutine or a reentrant guard.
- **Static checking catches real mistakes at load.** Tested:
  - `state.nope` → `no-field`;
  - an action missing `rows` → check error;
  - `store.customers()` in `ModeActions` → undefined function.
- **Actions are data, so the runtime is testable without a GUI or a database.**
  `TestDataToolActions` asserts the `fetch → loaded → rows` sequence, the save
  effect, and invalid input, all synchronously.
- **Getters become state.** The table and log are fed by `rows`/`append`
  actions; nothing calls a script during Fyne layout.
- **The event payload can be typed.** `values` carries the input fields, and
  the program reads it as a record. Invalid optional-field reads are caught
  by the checker.

## Measurements

Machine: AMD Ryzen 9 9950X, Linux, Go 1.27.2, Fyne 2.8.2, Risor `main`
`aeaaf1f` (Go module `github.com/deepnoodle-ai/risor/go/v3`, pinned as
`v3.0.0-20261112045619-aeaaf1fd6c13`).

| Operation | Time | Memory | Allocs |
|---|---:|---:|---:|
| Compile the counter page (`Check`) | 57 µs | 176 KB | 1,391 |
| Counter event: validate, run, set one label | 16.6 µs | 12 KB | 168 |
| Customers refresh (host function + 3-row table rebuild) | 463 µs | 726 KB | 20,574 |

The customer refresh is dominated by rebuilding the Fyne table, not by v3;
the counter event is the better measure of the language + host overhead.
Compilation is a one-time cost.

## Friction found

These are specifically useful for #640/#641:

1. **Optional fields need `?.`, not just `??`.** `event.node == "inc"` is
   `maybe-absent`; `event.node ?? ""` is still `maybe-absent`. The working form
   is `event?.node ?? ""`. This matches the complaint in #640 and is the first
   error a tool author meets.
2. **No `sprintf`.** The core library has `string`, `fixed`, `padStart` and
   template strings. `sprintf` is only a "wrong" hint, not an entry. Templates
   (`${...}`) cover the formatting needs here.
3. **Tagged-union inference is sensitive to spreads.** A list built from
   `[...effects, ...logActions, rowsAction]` produced a result-mismatch; the
   same actions as branch-local literals check cleanly. This is a style
   issue, not a blocker, but worth a diagnostic hint.
4. **Node IDs are strings.** A misspelled `node` in `{do: "set"}` is a
   check error only if the host declares the node IDs as a literal type. The
   prototype does not yet generate that type from the document; it is the
   obvious next step.
5. **No imports.** The program is one file. Fine for a page; a real sidecar
   will want language imports or host-level composition.

## Limitations (deliberate)

- The event queue drops events when full; a real host should reuse the
  unbounded queue design in `gui/vmexec`.
- `fetch` is performed synchronously inside `Step` for the demo. A real host
  should do it on a worker and call `Dispatch({type:"loaded", data})`.
- `readValues` understands `Entry`, `Check`, `Select`, `RadioGroup` and
  `Slider`; other input components need adding.
- One page, one program, no persistence or replay.
- The v3 Go module has no tag for the `go/v3` path yet; the module is pinned
  to a commit. Do not promote this dependency to fynerisor's main `go.mod`
  until there is a tag.

## Next experiment

If this direction is approved:

1. Generate the node-ID literal type and the property set from the scene
   specs, so action node/property mistakes are compile errors.
2. Add a real async effect host (HTTP or tie) with a worker, a journal and
   `Dispatch` on completion.
3. Write the same data-tool page as a v2 script and compare lines, errors
   caught before run, and event latency.
4. Report the results on #641 and link this branch.
