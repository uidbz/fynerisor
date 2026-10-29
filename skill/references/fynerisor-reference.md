# Fynerisor - Operational Guide

Distilled operational knowledge: semantics, gotchas, and patterns that the
API listing doesn't tell you. **Not an API reference** — for exact attribute
lists use `docs/API.md` (generated from the compiled registries, CI-pinned;
`fynerisor docs [--json]` regenerates it). For tutorials and full module
docs use `docs/README.md`.

Version: 0.9.0 · Risor v2.2+ · Fyne v2.8+ · `github.com/uidbz/fynerisor`
(host code: `core` = headless, `gui` = Fyne; the root package is docs-only.)

## Script Model (most mistakes come from these)

- Scripts get a **`window` global**; the host (or `fynerisor` CLI) owns the
  app lifecycle. Never create apps/windows from script code.
- **All script code** — top-level body and every callback (button taps,
  menu items, shortcuts, table `Data`, chart `OnTap`, ...) — runs on the
  script's own **VM goroutine**, never on Fyne's UI thread. One run is
  effectively single-threaded: callbacks never run concurrently with each
  other, and the UI keeps repainting while script code runs.
- **Widget methods are safe to call from anywhere** (inside handlers, inside
  `go()`, top level): each marshals its own UI work to the UI thread. There
  is **no `window.Do`** — it was removed; nothing needs a GUI wrapper.
- `go(fn)` is a **deferral, not parallelism**: `fn` is queued on the same VM
  goroutine and runs after the current body/event, in queue order. A long
  `go()` callback delays that run's later callbacks. Use it to sequence
  updates, not for speed. Real parallel work belongs in host Go (a
  `WithGlobal` function spawning goroutines that reports back via callback).
- `import("path.risor")` is **live at runtime** and returns a module
  namespace object (`let utils = import("utils.risor")`, then `utils.f(x)`).
  Relative paths resolve against the script's own location; modules are
  cached per run; circular imports error; each `Execute`/`--watch` reload
  re-imports fresh (no state leaks across runs). URL imports work.
- `require(["v0.9", "@sql", "@http"])` is optional gating: fails fast if the
  host hasn't enabled a module or the version is too low. Module names:
  `@http @sql @os @io @time @strings @filepath @json @csv @exec @tie`
  (`@gui` is implicit in GUI runs).

## Critical Risor v2 Syntax Rules

**Arrow functions only** - no `func` keyword:
```risor
let onClick = () => { print("clicked") }
let process = (x) => x * 2
```

**No for/while loops** - functional methods only:
```risor
[1, 2, 3].map(x => x * 2)
list.each(item => print(item))
```
Also: `.filter()`, `.reduce()`. See SKILL.md for the `match`-arms gotcha.

**No `sleep()`** - the `time` module has no sleep and there is no builtin.
Never script delays; sequence with `go()` / callbacks.

**`else` on the same line** as the preceding `}` (`} else {`).

**SQL iterators** stream — `.map()`/`.each()` are single-pass, `.collect()`
loads the full result:
```risor
conn.query("SELECT * FROM users").map(row => row["name"])   // streaming
conn.query("SELECT * FROM users").collect()                 // full list
```

## Data Binding Semantics

Bindings hold one value (`String`/`Bool`/`Int`/`Float`) and expose **only**
`Get()`, `Set(v)`, `AddListener(fn)`. Fyne 2.8 removed binding *mappers* —
there is **no `binding.map()`**; derive displayed values through a listener:

```risor
let temperature = binding.NewFloat(20.5)
let slider = widget.NewSliderWithData(0, 100, temperature)
let readout = widget.NewLabel("")
temperature.AddListener(() => {
    readout.SetText(sprintf("%.1f C", temperature.Get()))
})
```

Bound widgets: `NewLabelWithData`, `NewEntryWithData` (bi-directional),
`NewCheckWithData`, `NewSliderWithData`, `NewProgressBarWithData`.
Bindings are created via the **`binding`** namespace — `binding.NewString()`,
never `widget.NewString()`.

## Gotchas Worth Their Weight

- `window.Size` is a **map property**, not a method: `window.Size.width` /
  `.height`. `window.Size()` throws "object is not callable (got map)".
- `container.NewStack(...)` exposes only `Show`/`Hide` — no `.Objects`,
  `Add`, or `Remove`. To swap content: rebuild the layout and
  `window.SetContent(...)` (safe from anywhere, including inside `go()`).
- Canvas images load from a **URI only**: `canvas.NewImageFromURI("file:///abs/path.png")`.
  `canvas.NewImageFromFile` is gone.
- Every widget **and** canvas object shares these verbs: `Show()`, `Hide()`,
  `Refresh()`, `Resize(w, h)`, `Move(x, y)`, `MinSize()` → `{width, height}`,
  `Visible()` → bool.
- Attribute/arity errors from the registries carry the type and method name
  and are **catchable**: `widget.Label.SetText: expected 1 argument, got 0`.
- `strings` uses `to_upper`/`to_lower` (not `upper`/`lower`).
- `io.cp` is deprecated — use `io.copy(src, dst)` (preserves mode).
- `os.exec(...)` just launches (non-blocking, no captured output); use the
  `exec` module (`exec.run` → `{stdout, stderr, combined_output}`,
  `exec.start(...).wait()`) to capture output.
- `widget.NewProgressBarInfinite()` starts **idle** (v0.8.1+): call
  `Start()`/`Stop()`.
- Charts: all chart errors are catchable Risor errors; `SaveImage` exports
  **SVG only** (no PNG). Details: `docs/API.md` § chart,
  `examples/42-interactive-charts`.
- Table `Data`/`RowCount`, `List.Length`, `Tree.ChildUIDs` callbacks are
  invoked *from the UI* but execute on the run's VM goroutine — keep them
  fast and side-effect-light; refresh via `table.Refresh()`.
- `print()` output can be redirected to a Log widget:
  `window.SetStdout(logWidget)`.

## Patterns

### Update UI from Callback
```risor
let label = widget.NewLabel("Click the button")
let btn = widget.NewButton("Update", () => {
    label.Text = "Button was clicked!"    // no wrapper needed
})
```

### Deferred Work (go)
```risor
let btn = widget.NewButton("Compute", () => {
    window.SetStatus("working...")
    go(() => {
        // runs after the current event finishes, on the same VM goroutine
        let answer = expensive_pure_risor_computation()
        resultLabel.SetText(string(answer))
        window.SetStatus("done")
    })
})
```

### Build Dynamic List
```risor
let items = []
let list = widget.NewList()
list.Length(() => len(items))
list.CreateItem(() => widget.NewLabel(""))
list.UpdateItem((id, obj) => { obj.Text = items[id] })

let addBtn = widget.NewButton("Add Item", () => {
    items = items + ["Item " + string(len(items) + 1)]
    list.Refresh()
})
```

### Form with Validation
```risor
let entry = widget.NewEntry()
let errorLabel = widget.NewLabel("")
let submit = widget.NewButton("Submit", () => {
    if (entry.Text == "") {
        errorLabel.Text = "Field is required"
    } else {
        errorLabel.Text = ""
        // process entry.Text
    }
})
```

### Paginated Table from SQL
```risor
let conn = sql.connect("sqlite3::memory:")
conn.exec("CREATE TABLE IF NOT EXISTS items (id INT, name TEXT)")

let table = widget.NewTable("items", 50)
table.Columns(() => ["ID", "Name"])
table.RowCount(() => conn.query("SELECT COUNT(*) FROM items").collect()[0][0])
table.Data((offset, limit) => {
    return conn.query("SELECT id, name FROM items LIMIT " + string(limit) +
        " OFFSET " + string(offset)).map(row => [row["id"], row["name"]])
})
// refresh after writes: table.Refresh()
```

### Error Handling
```risor
try {
    conn.query("SELECT * FROM nope")
} catch (e) {
    window.SetStatus(sprintf("Query failed: %v", e))
}
```
Parse/compile errors abort before any code runs; runtime errors are catchable
Risor errors. Uncaught callback errors go to the run's error sink/log.

## Go Host Entry Points (names only — see README "Usage as a Library")

- `gui.NewApp(title, opts...)` → standalone window app; `fw.LoadScript`,
  `fw.Execute`, `fw.ShowAndRun`.
- `gui.NewWindow(fyneWindow, opts...)` → script bound to your Fyne window.
- `gui.NewCanvasObject(script, opts...)` → embed a script view as a Fyne
  widget; `gui.WithHostWindow(host)` unlocks dialogs/shortcuts.
- `core.NewContext(opts...)` + `ctx.Eval(script)` → headless, no GUI deps.
- Options: `WithHTTP/WithOS/WithIO/WithTime/WithStrings/WithFilepath/
  WithJSON/WithCSV/WithExec/WithTie/WithAppName/WithGlobal/...`; capability
  presets `Safe()`, `Files(readOnly bool)`, `Exec()`, `HTTP()`, `Full()`.
- SQL is opt-in: `guidb.WithSQL()` (gui/db) or `db.WithSQL()` (core/db).
- Custom globals: `WithGlobal(name, goValue)` (maps/funcs auto-convert) or
  `WithRisorOptions(risor.WithEnv(...))`.
- Host goroutines marshal into the VM with the Go method
  `Window.Do(func())` (script code needs nothing; the old script-visible
  `window.Do` attr was removed).

## CLI (see README / `fynerisor --help`)

`fynerisor [run] script.risor`, `check` (compile-only, CI gate), `analyze
[--json]`, `docs [--json] [-o file]`. Exit codes 0/1/2 (ok/script/usage).
