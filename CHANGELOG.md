# Changelog

## [0.9.0] - 2026-10-19

### Breaking
- **tie dependency v0.7.3.** `tie.associated()` now flattens the flat
  `[]tiedb.Row` reply internally; the Risor-visible list of
  `{key, value1, value2}` maps is unchanged, but callers embedding the
  Go client must target tie v0.7.3.
- **SQL is opt-in in headless `core`.** `core.WithSQL()` was removed;
  import `github.com/uidbz/fynerisor/core/db` and pass `db.WithSQL()`
  to opt in. Binaries that never touch SQL now link zero database
  drivers (verified against the lars static build). `gui.WithSQL()`
  is unchanged.

## [Unreleased]

### Breaking

- **`Window.Status` is now an accessor, not a field.** `Window.Status()`
  returns the last status string; writes go through `SetStatus`. The field
  was written by the Execute goroutine and read by hosts/tests without
  synchronisation (a data race). `CanvasObject.Status()` is unchanged.
  (issue #15)
- **`go(() => {...})` no longer runs concurrently with the script.** The
  closure is queued on the script's VM goroutine and runs after the current
  body or event, in call order, serialized with all other callbacks. The old
  raw-goroutine form raced the single-threaded VM (the corruption class
  documented in `docs/CONCURRENCY.md`). This is the permanent semantics
  (issue #17): the GUI stays responsive while such work runs because script
  code never runs on Fyne's UI thread; real parallelism belongs in Go (a
  host global or module that does the work and reports back through a
  callback). (issues #15, #17)
- **`gui.safeCall` is gone**; use `vmexec.NewCallback` (issue #17).
- **`risorwidget.NewMultiLineEntry` takes the window** like the other entry
  constructors; multi-line entries previously panicked in `OnSubmitted`.
  (issue #36)
- **Relative `import()` paths resolve against the importing script, not the
  process working directory.** A host that tells fynerisor where a script
  came from (`LoadScriptFrom(script, source)` — now used by both CLI and
  browser) gets imports resolved against the script's own directory, or,
  for scripts loaded over HTTP(S), against the script's URL with a
  same-origin policy: relative imports fetch from the script's own server,
  cross-origin URLs are rejected, and local filesystem paths are never
  consulted for remote scripts. Cache and cycle-detection keys are
  canonicalised (absolute path / absolute URL), so `./a.risor` and
  `a.risor` are the same module. Scripts loaded via the old `LoadScript`
  keep the CWD behaviour. (issue #8)
- **`gui/vmguard` is removed; `gui/vmexec` replaces it.** The old guard
  was one process-wide variable; a `vmexec.Run` owns one VM goroutine,
  one reentrant guard and one event queue *per script run*. Widget code
  registers script functions with `vmexec.NewCallback` and calls `Fire`
  (events, async) or `Call` (render getters, sync, may return
  `vmexec.ErrBusy`). `window.Do(fn)` and shortcut handlers now run on the
  script's VM goroutine instead of being re-dispatched onto the UI thread
  (widget methods already marshal their own UI work). (issue #13)
- **Every widget callback goes through `vmexec.Callback`.** Event handlers
  that used to run synchronously on the UI thread (`Button.OnTapped`,
  `Tree`/`List`/`GridWrap` `OnSelected`/`OnUnselected`, `Table.SetOnClick`,
  `AppTabs.OnSelected`, `Markdown.SetOnTapped`, chart hover/click) are now
  queued on the script's VM goroutine like all other events. Passing a
  non-function where a callback is expected now *raises* an error instead
  of returning an error value. (issue #12)
- **Imported modules belong to a run.** Every `Execute` starts with an
  empty module cache, so a reload (`--watch`, browser refresh) imports
  modules again instead of reusing the previous run's instances — edited
  modules are picked up, and no module state or module-registered
  callback is shared with a stopped run. On re-`Execute` the previous
  run's keyboard shortcuts, script-installed main menu and
  drop/stdout-capture handlers are removed. (issue #16)

### Fixed

- Script access to Fyne objects no longer races with the UI thread. Every
  attribute read/write, `Inspect` and method call on a gui object (widgets,
  containers, canvas objects, dialogs, menus, bindings, the window and the
  factories) now runs on Fyne's UI thread — previously e.g. reading
  `entry.Text` while the user typed was an unsynchronized read on the
  script's goroutine. Builtins that do I/O (`canvas.NewImageFromURI`,
  chart `SaveImage`) are marked `guithread.OffUI` and only marshal their
  Fyne work; image downloads honour the run's context and a 30 s timeout.
  Tests can reproduce the real UI-thread split (`startUILoop`) instead of
  relying on Fyne's inline test driver (issue #49).
- A stopped or replaced run can no longer write window state: its
  `window.SetStatus`, `window.SetContent`, `window.SetMainMenu` and final
  "Ready!"/"ERROR" status are ignored, and its widgets' callbacks are
  dropped. `Execute()` called from inside a running script (as
  `browser.Open` does) returns immediately instead of waiting up to 2 s
  for its own goroutine. `window.SetMainMenu` now runs on the UI thread
  (issue #16).

- About 30 callback sites (Tree and GridWrap getters, `Button.OnTapped`,
  `MenuItem`, `Markdown`, `Hyperlink`, `Slider`, `SelectEntry`, `DateEntry`,
  `Calendar`, `Entry`, chart callbacks, bindings) called the VM directly:
  without panic recovery (a faulty handler crashed the process) and
  without the VM guard (getters ran concurrently with handlers). Tree and
  GridWrap getters now keep last-known values while the VM is busy and
  refresh once it is idle (issue #12).

- Two windows or embedded views no longer block each other: while view A
  runs a handler, view B's render getters (List/Table) answer instead of
  being rejected as "concurrent VM access" and showing stale data. Risor's
  lazily-created type registry is now initialized once, so two VMs
  starting at the same moment no longer race inside Risor (issue #13).
- fynerisor-browser no longer enables the `os` module for pages; remote
  scripts can no longer read, delete or exfiltrate local files or run
  commands (issue #7, security).
- `Execute()` now passes its run context to `risor.Eval`: `Stop()`,
  re-`Execute()` and browser navigation actually cancel the previous VM
  (issue #9).
- The callback loop starts before the main body, and event sends never
  block Fyne's UI thread; the 11th interaction after a script error, a
  long-running body, or `Stop()` no longer freezes the app (issue #10).
- `Entry.OnChanged` and `binding` listeners no longer spawn a goroutine
  per keystroke/change into the VM (issue #11).
- Data races: `Window.Status`, `ScriptRunner.LoadScript` vs `Eval`, the
  callback channel pointer, and successive `Execute()` runs are now
  synchronized; `go test -race ./...` passes (issue #15).
- `clipboard.Read()` returns the real clipboard content when called off
  the UI thread (new `guithread.DoAndWait`, also used by chart SVG
  exports) (issue #18).
- GUI `import()` of a URL answering 404/500 now fails loudly with the
  status code; the broken duplicate fetcher was deleted in favour of the
  shared `core.FetchScript` (timeout + size limit) (issue #19).
- `require("@")` no longer panics; the "module not enabled" hint names
  the real option (`WithHTTP`, `WithSQL`, …) (issue #20).
- `widget.NewForm` list arguments expand the current argument, not
  `args[0]` (issue #21).
- The CLI no longer calls the removed `ImportScript` for analysed
  imports; a script with a top-level `import(...)` line no longer makes
  it exit(1) (issue #22).
- Small inconsistencies: `widget.NewMuliLineEntry` typo, copy-pasted
  "button" error messages, wrong `want=` counts, `NewAccordionItem`
  detail argument now renders, `NewCalendar` rejects non-time arguments
  instead of silently defaulting (issue #36).
- Creating a window or embedded view from a background goroutine no
  longer re-records that goroutine as the GUI thread. fynerisor now uses
  `guithread.Init` (first caller wins, then a one-time `fyne.Do` probe
  records Fyne's real event-loop goroutine); `guithread.SetMain` remains
  as an explicit override for hosts and tests (issue #14).

### Added

- **Typed script lifecycle.** `gui.RunState` (`StateIdle`, `StateRunning`,
  `StateReady`, `StateFailed`, `StateStopped`), `Window.State()`,
  `Window.Start(ctx) <-chan error`, `Window.Run(ctx) error` and
  `gui.WithOnStateChange(func(RunState, error))`, mirrored on
  `CanvasObject`. Hosts no longer need to parse the `"Ready!"` /
  `"ERROR: ..."` status strings, which scripts can also write with
  `window.SetStatus`; the status text remains for display. The CLI's
  `--headless` mode uses `Run` with a 30 s timeout instead of polling the
  status (exit code 0/1 as before), and fynerisor-browser shows its error
  page on `StateFailed` instead of on any "ERROR: " status (issue #39).
- **`gui/vmexec`**: the per-run script executor (`Run`, `Callback`) used
  by all widgets; see `docs/CONCURRENCY.md` (issue #37).
- **`gui.NewCanvasObject` for embedding in existing Fyne apps.** A script can
  now be hosted as a regular `fyne.CanvasObject` and placed anywhere a widget
  fits (containers, border slots, tabs, dialogs). `window.SetContent` renders
  inside the object instead of replacing the host window's content.
  `WithHostWindow(host)` opts the embedded script into window-level features
  (dialogs, shortcuts, main menu, `window.Canvas`, `window.Close`, drops);
  without it those functions return clear script errors rather than panicking.
  The object also exposes `Window()`, `LoadScript`, `CompileScript`, `Execute`,
  `Stop`, `Clear` and `Status`. See `examples/43-embedded-canvas-object`.
  `Window.Stop()` is now public for cancelling an execution from Go.

- **Interactive charts.** All five chart types (`chart.NewBarChart`,
  `NewLineChart`, `NewScatterChart`, `NewHistogram`, `NewBoxPlot`) are now
  live widgets built from Fyne canvas primitives instead of static PNG
  snapshots rendered by gonum/plot. On every chart, with no setup:
  hovering highlights the element and shows a label/value tooltip; the mouse
  wheel zooms around the cursor; dragging pans; double-click resets the view.
  New script API per chart: `OnTap/OnHover/OnSecondaryTap((i, label, value) => ...)`
  (right-click + `widget.NewPopUp` enables context menus), feature toggles
  (`ShowGrid`, `ShowValues`, histogram `ShowNormal`), zoom controls
  (`SetZoomEnabled`, `ResetZoom`), live data updates (`SetValue`, `SetData`,
  `SetPoints`, `SetBins`, `SetValues`, and `SetLabel`),
  appearance setters (`SetColor`, `SetBarColor`, `SetMarkerSize`,
  `SetLineWidth`), and `SetTitle`.
  See `examples/42-interactive-charts`.

- **OS module parity with plain risor (issue #1).** `os` gains `copy_file(src,
  dst)` (preserves mode), `remove_all(path)`, `mkdir(path[, mode])`,
  `create(path)`, `args()` (script arguments, like risor v1's `os.args`), and
  `home_dir()` (shortcut for `current_user().home_dir`). These restore names
  goto-apps expect from the v1 os module that fynerisor's trimmed module had
  dropped. `os.read_dir` entries now also carry a full `path` field next to
  `name`. Script arguments are populated by the `fynerisor` CLI automatically;
  embedding hosts use `core.WithScriptArgs(...)` / `fynerisor.WithScriptArgs(...)`
  (thin wrapper over `osmodule.SetScriptArgs`).

- **SVG chart export.** `SaveImage(path[, w, h])` and `WriteSVG([w, h])` on
  every chart serialize the exact scene being displayed (static, selectable
  text; crisp at any scale). PNG is intentionally not supported; SVG converts
  cleanly for reports (pandoc embeds it automatically). The old "chart as a
  generated image" use case is fully covered by exporting, so the static
  renderer was removed entirely; gonum/plot is no longer a dependency.

### Changed

- **Chart engine internals (`gui/chart/`).** New shared engine: `scale.go`
  (linear scales with clamped zoom/pan and nice-tick generation), `scene.go`
  (serializable primitive scene + clip-aware painter + SVG writer),
  `core.go` (widget/renderer, pooled canvas objects, axes/grid/tooltip
  chrome), `interact.go` (hover/tap/scroll/drag handling + shared Risor
  verbs). Chart types supply geometry + hit-testing only.
- **Chart errors are real Risor exceptions.** Chart builtins now return Go
  errors (which throw and are catchable with `try/catch`) instead of the
  previous silently-returned error objects. Constructor validation is
  unchanged.
- **Box plot visuals.** Now a Tukey box plot (IQR box, median, 1.5*IQR
  whiskers, mean marker, outlier points) instead of box + fitted normal
  curve overlay. The histogram normal curve remains (togglable via
  `ShowNormal`).
- **`log.Fatalf` removed from chart construction** (bad data no longer
  crashes the process; it raises a Risor error).

### Added

- **`clipboard` global** with `clipboard.Copy(text)` and `clipboard.Read()`.
  Wraps fyne's `app.Clipboard` through `guithread.Do`. Handy for copy buttons on
  non-selectable widgets (e.g. Markdown/RichText, which fyne cannot select).

- **`canvas.NewText(text, color)`.** A colored, non-wrapping text object (wraps
  fyne's `canvas.Text`, which has a `Color` field — unlike `widget.Label`, whose
  color only follows theme Importance). Settable attrs: `Text`, `Color`
  (name/hex), `Size`. Use to tint a heading to match a custom background.

- **`Selectable` attribute on `widget.Label`.** Maps to fyne's `Label.Selectable`
  (2.6+); `lbl.Selectable = true` makes label text selectable/copyable. Note
  fyne's `RichText`/`Markdown` has **no** selection support, so markdown text
  cannot be made selectable — use a selectable `Label` if selection matters more
  than markdown formatting.

- **`Wrapping` / `Truncation` attributes on the `Markdown` widget.** fyne's
  `RichText` (which Markdown wraps) defaults to `Wrapping=Off`, whose MinSize
  width is the longest paragraph on one line — inside a `container.Scroll` that
  grows a horizontal scrollbar instead of wrapping. Setting
  `md.Wrapping = constants.TextWrapWord` collapses the MinSize width so the
  scroll sizes to the viewport and markdown text autowraps.

- **`canvas.NewRectangle(fillColor, cornerRadius)`.** Rounded/filled rectangles
  for card and chat-bubble backgrounds. Colors accept names or hex
  (`#RGB`/`#RRGGBB`/`#RRGGBBAA`); `FillColor`, `StrokeColor`, `StrokeWidth` and
  `CornerRadius` are settable attributes. Combine with
  `container.NewStack(bg, container.NewPadded(content))` for a rounded card.

- **`ScrollToBottom()` / `ScrollToTop()` on `container.Scroll`.** Wraps Fyne's
  `Scroll.ScrollToBottom`/`ScrollToTop` through `guithread.Do`, so scripts can
  keep a chat or log view pinned to the newest content after appending to a
  scroll (e.g. `s.ScrollToBottom()` right after `window.SetContent(...)`).

- **Relation value types in the `@tie` module.** `db.value_types()` returns the
  value type declared for each relation in the collection (`relation -> type`
  string map; `string` is the default and is omitted), and
  `db.set_value_types({...})` declares the type of the named relations, merging
  into whatever the collection already holds — relations not named are left
  alone. Types are advisory annotations on a *relation*, not a constraint on
  stored values: the store keeps values as plain strings and never validates a
  value against its relation's declared type. This types table columns for free,
  since a column's key is a relation.
- **`numeric` option in `tie.query`.** `db.query({sort_by_value: "mass",
  numeric: true})` reads the sorted values as numbers instead of strings, so 9
  precedes 10. Values that do not parse sort last in either direction. It is the
  ordering half of the value-types feature: declare a numeric relation with
  `set_value_types`, then sort on it with `numeric`.

### Changed

- **SQL is now opt-in in the headless `core` package.** `core.WithSQL()` was
  removed: importing `core` pulled go-sql-driver/mysql, lib/pq,
  microsoft/go-mssqldb and mattn/go-sqlite3 into every binary through
  `modules/sql`, even when no database was used. Headless users now import
  `github.com/uidbz/fynerisor/core/db` and pass `db.WithSQL()`; binaries that
  do not import it link without the DB drivers (several MB smaller static
  builds). `gui.WithSQL()` is unchanged.

- Upgraded the tie dependency to the per-collection value-type registry commit
  (410474f, `v0.5.2-0.20260904115826-410474f21ff3`), which `value_types` and the
  `numeric` query option build on. The `@tie` module's public API is otherwise
  unchanged.

- **Item templates are never faked while the VM is busy (issue #50).**
  `List`/`GridWrap` `CreateItem`, `Tree` `CreateNode` and the `Table`'s
  `CreateCell` are render getters: when the script's VM was busy they
  answered with a placeholder `widget.Label`, which Fyne pools for the
  row's whole life — every later `UpdateItem` then handed the script that
  Label instead of its widget, permanently. They now hand Fyne an empty
  slot (or, for the table, leave the cell empty) that takes the script's
  real widget the first time an update getter runs with a free VM. As a
  side effect `Tree.UpdateNode` now passes non-Label node widgets (Card,
  Button, ...) as their proper wrapper type instead of `nil`.

- **The script-event queue never drops events; change events coalesce
  (issue #51).** The per-run queue was bounded at 100 with dropped events;
  fast typing during a long body could lose the last `OnChanged` forever.
  Discrete events (taps, submits, selections, dialogs) are now queued
  without bound — a backlog deeper than `vmexec.Config.QueueSize` is
  reported once instead of dropped. Change-style events (Entry,
  SelectEntry, DateEntry and Slider `OnChanged`, binding listeners) are
  coalesced per callback: at most one is pending, always carrying the
  latest value, so a burst can neither fill the queue nor lose the final
  state. `vmexec.ErrQueueFull` is gone; `Callback.FireCoalesced` /
  `Run.EnqueueCoalesced` is the new registration API for change events.

### Added

- **CI now runs the race detector (issue #45).** `.github/workflows/ci.yml`
  runs gofmt, `go vet`, `go build` and `go test -race -count=1 ./...` on
  every push and pull request. This is the acceptance criterion from the
  October 2026 review; the `-race` failures it guards against were already
  fixed on main (issue #15), so the suite is green from the first run.
- **Regression tests for run reclamation and binding thread safety
  (issue #45).** `TestReplacedRunIsReclaimed` proves a replaced run's VM
  environment and its window-side registrations (shortcuts, drop/stdout
  hooks, the current-run pointer, queued events) become garbage-collectable
  after a new `Execute` — no leak, and no finalizer or manual `Destroy`
  needed anywhere. `TestBindingConcurrentWritesKeepLatestValue` and
  `TestBindingBoundWidgetSyncFromHostGoroutines` hammer a script-visible
  binding from many host goroutines while listeners and bound widgets are
  live, pinning down the contract: hosts may read and write bindings from
  any goroutine (Fyne's storage is internally locked), while script
  handlers always run on the run's VM goroutine and always see the last
  value written.

### Changed

- **The data-binding thread-safety contract is documented**
  (`docs/CONCURRENCY.md`): the wrappers are stateless pass-throughs over
  Fyne's locked bindings, script access is UI-thread-routed like every
  other attribute access, and listener notifications coalesce through the
  run queue. This closes the old roadmap item "per-object mutexes in the
  binding layer" — the issue #49 routing made those mutexes unnecessary.
- **Stale docs refreshed (issue #46).** `docs/README.md` no longer claims
  callbacks "execute on the GUI thread via a function queue" (they run on
  the run's VM goroutine; the UI thread never blocks), the `CLAUDE.md`
  callback snippet now shows `vmexec.NewCallback` instead of raw
  `object.GetCallFunc`, and `docs/IMPROVEMENTS.md` is marked as a
  historical, superseded plan. `docs/CONCURRENCY.md` gained the "Multiple
  runs, one UI thread" and "Data bindings" sections covering the
  multi-window model (one queue + guard per run, one UI thread per
  process).

## [0.8.2] - 2026-09-03

### Added

- **Multi-row (hierarchical) table headers in the `@tie` module.**
  `db.insert_table` now accepts a list of header *rows* as well as a flat list of
  labels, so a two-row header (`Temperature (20°C)` over
  `Replicate 1 | Replicate 2`) survives storage instead of having to be flattened
  by hand, and `db.read_table` returns a new `header_levels` key holding those
  rows. Adding a map key is non-breaking: `headers` still reports the flat column
  keys every cell is stored under, and a one-row header reports one row in
  `header_levels`, so existing scripts need no change. See
  [TIE_MODULE.md](docs/TIE_MODULE.md).
- **`table.HeaderLevels(fn)`** renders a multi-row header above the columns, so a
  hierarchical table shows its levels instead of the joined column keys. The
  callback returns the header rows row-major — the shape `db.read_table` reports
  `header_levels` in — and a parent repeated across the columns it covers is drawn
  once across the run, restoring the merged cell of the source sheet. Display only:
  `Columns` stays the identity used for sorting, filtering and export, and the
  header row grows taller by itself. A table that sets no levels renders exactly as
  before. See `examples/41-table-header-levels`.

### Changed

- Upgraded the tie dependency to v0.5.1, which stores a table's header levels and
  reads them back with its columns in one call — the storage half of the two
  features above. v0.5.0 migrated the module path from `git.sr.ht/~uid/tie` to
  `github.com/uidbz/tie`; the `@tie` module's public API is unchanged. Note:
  upstream renamed the server binary `tie-daemon` to `tie-triplestore`.

## [0.8.1] - 2026-08-28

### Added

- `tie.delete_table` to drop a table via the `@tie` module
- `tie.query` now accepts `sort_by_value` and `descending` options
- Dialog `Resize(width, height)` on `CustomDialog`, `ConfirmDialog` and `FormDialog`,
  plus a `window.Size` getter (returns `{width, height}`) so scripts can size dialogs
  relative to the window
- `ProgressBarInfinite` now starts idle (created hidden); `Start()` shows and spins it,
  `Stop()` stops and hides it, so it is visible only while running

### Fixed

- FlexTable now resets its scroll offset on `SetData`
- Table getters stay resilient under VM contention
- List stays resilient under VM contention; `ErrConcurrentAccess` is now exported

### Changed

- Upgraded Fyne to v2.8.1

## [0.8.0] - 2026-08-21

### Added

**Parallel batch execution (`core.EvalBatch`):**
- `core.EvalBatch(ctx, script, inputs, cfg, opts...)` compiles a Risor script once
  and runs it over many inputs concurrently, each on its own isolated `Context`+VM
  ("compile once, run many"). Real multi-core throughput for headless data work.
- Each input `map[string]any` is exposed to the script as global variables; results
  are index-aligned with inputs via `[]core.BatchResult{Value, Err}`. Per-input
  failures are isolated in `BatchResult.Err`; the returned error covers only setup
  (e.g. compilation) failures.
- `core.BatchConfig{Concurrency}` bounds parallelism (defaults to `GOMAXPROCS`).
- Headless only (no GUI dependency). See [docs/CONCURRENCY.md](docs/CONCURRENCY.md)
  for the isolation model and why a script-level `parallel()` builtin was not added.
- Example: `40-concurrent-batch`

**Tie Module (`@tie`):**
- Client for the [tie](https://github.com/uidbz/tie) triple store, storing and querying
  `(key, relation, value)` triples. Enable with `core.WithTie()` / `gui.WithTie()`;
  require with `require(["@tie"])`.
- `tie.connect(url, opts?)` - Connect to a tie-daemon (optional `{username, password}`)
- Writes: `db.add`, `db.set`, `db.update`, `db.delete`, `db.sync`
- Reads: `db.get`, `db.query` (returns `{rows, total_count}` for pagination),
  `db.expand([keys])`, `db.associated`, `db.exists`
- Bulk: `db.batch()` with `batch.add/set/update/delete/run()`
- Tables: `db.insert_table` / `db.read_table`
- Maintenance: `db.dump`, `db.dump_stream(callback)` (memory-efficient streaming),
  `db.restore`, `db.drop`
- Documentation: [docs/TIE_MODULE.md](docs/TIE_MODULE.md)
- Examples: `36-tie-headless` (core) and `37-tie-gui`

**JSON Module (`@json`):**
- `json.parse(text)` - Parse a JSON string into a Risor value
- `json.marshal(obj)` - Encode a Risor value to a compact JSON string
- `json.marshal_indent(obj, indent="  ")` - Encode to a pretty-printed JSON string
- `json.valid(text)` - Report whether a string is valid JSON
- `json.read(path)` / `json.write(path, obj, indent?)` - Read/write JSON files
- Enable with `core.WithJSON()` / `gui.WithJSON()`; require with `require(["@json"])`

**CSV Module (`@csv`):**
- `csv.parse(text, opts?)` - Parse CSV. By default row 1 is treated as headers and
  rows are returned as a list of maps; pass `{header: false}` for a list of lists.
  Supports `{delimiter: ";"}`.
- `csv.format(rows, opts?)` - Encode a list of maps or list of lists to CSV. For map
  rows, columns default to alphabetical order; pass `{columns: [...]}` to control order,
  `{header: false}` to omit the header row, and `{delimiter: ";"}`.
- `csv.read(path, opts?)` / `csv.write(path, rows, opts?)` - Read/write CSV files
- Enable with `core.WithCSV()` / `gui.WithCSV()`; require with `require(["@csv"])`
- Examples: `38-csv-headless` and `39-json-headless` (both core, headless)

**Widgets & Containers:**
- `AppTabs.OnSelected` callback, invoked when the selected tab changes
- Alternating row colors in tables, adapting to light and dark themes

### Changed

- Upgraded Risor language runtime to v2.2.0
- Depend on published `tie` v0.4.3 (dropped the local `replace` directive)

### Fixed

- Correct `file://` URL handling for Windows filesystem paths
- Route status bar updates onto the GUI thread (opt into `fyne.Do`)
- Guard `TableData.Sort` against a missing sort column
- Enable text clipping in table cells to prevent overflow

## [0.7.0] - 2026-07-13

### Added

**Browser Package:**
- **browser/** - Generic, reusable browser UI for fynerisor applications
  - Navigation: address bar, back/forward/refresh/home buttons
  - History management with browser-style back/forward stacks
  - URL handling: automatic normalization and relative path resolution
  - Source view toggle: split layout showing content and source code side-by-side
  - **Programmatic navigation from scripts:**
    - `browser.Open(url)` - Navigate to a URL from Risor scripts
    - `browser.GetURL()` - Get current URL
    - `browser.SetStatus(text)` - Set status bar text
  - **URL query parameters:**
    - `browser.params` - Map of the current URL's query parameters, e.g. `https://server/page?myarg=value` exposes `browser.params["myarg"]`
    - `browser.GetParam(name)` - Get a single query parameter value (empty string if absent)
    - Available to external apps via `Browser.GetParams()` and `Browser.GetParam(name)`
  - Hybrid plugin system:
    - MenuProvider interface for custom menu items
    - AuthProvider interface for custom authentication UI
    - SourceViewProvider interface for source code display
    - Callbacks for navigation lifecycle (OnNavigate, OnError, etc.)
  - Built on fynerisor's gui.Window for seamless integration

**Reference Browser:**
- **cmd/fynerisor-browser/** - Reference implementation demonstrating browser package usage
  - Loads Risor scripts from HTTP(S) or file:// URLs
  - Shows plugin architecture and integration patterns
  - Programmatic navigation enabled via `browser` global object
  - Serves as starting point for custom browser applications
  - ~350 lines with clear comments
  - **Android/mobile packaging support:**
    - Uses `app.NewWithID` so preferences and storage work on mobile
    - Startup (home) URL resolved from `HomeURL` custom app metadata, since
      mobile apps receive no command-line arguments; set it in `FyneApp.toml`
      or via `fyne package --metadata HomeURL=...`
    - `FyneApp.toml` + `Icon.png` added; `fyne package -os android` builds an APK
      (auto-generated manifest already grants the `INTERNET` permission)
    - Table widget clipboard copy is now behind build tags: on mobile the
      `golang.design/x/clipboard` dependency pulls in `golang.org/x/mobile/app`,
      which clashes with Fyne's mobile driver (duplicate C symbols). It is
      excluded on android/ios, where table clipboard copy is a no-op.

**Charts:**
- **New chart types** - Added to the `gui/chart` package alongside the existing bar chart
  - **chart.NewLineChart** - Show trends over continuous data
  - **chart.NewScatterChart** - Visualize individual data points
  - **chart.NewHistogram** - Distribution of numerical data with an automatic normal-distribution overlay (fitted from the data's mean and standard deviation)
  - **chart.NewBoxPlot** - Compare distributions across groups (five-number summary plus fitted normal-distribution curves)
  - Statistical helpers (mean, standard deviation, normal PDF) in `gui/chart/stats.go`
  - Example: 35-charts

**Concurrency & Safety:**
- **VM access guard (`gui/vmguard`)** - Detects concurrent access to the single-threaded Risor VM
  - Non-blocking guard that surfaces concurrent-access bugs instead of corrupting the VM stack
  - Reentrant per-goroutine, so nested calls on the same goroutine are allowed
  - Recovers from panics in script callbacks and reports them as errors

**Time Module:**
- **time.unix(seconds)** - Construct a time value from a Unix timestamp (seconds since epoch)

**Window Enhancements:**
- **window.RegisterGlobal(name, value)** - Add global objects after window creation
  - Useful for objects that need to reference window or components created after initialization
  - Example: registering browser object for programmatic navigation

**Error Handling:**
- **Script execution error display** - Errors now show in clean UI with "Copy Error to Clipboard" button
- **browser.CopyToClipboard(text)** - Copy text to system clipboard from scripts
- **Terminal error logging** - All errors logged to terminal for debugging

**Widget Enhancements:**
- **widget.Select** - Added `Selected` and `SelectedIndex` getters

### Fixed

**Canvas Image Loading:**
- **canvas.NewImageFromURI()** - Better error messages for invalid URIs
  - Relative paths now return clear error: "relative paths are not supported, use absolute file:// URI or HTTP(S) URL"
  - Panic recovery for malformed URIs prevents crashes
  - Empty URI check with helpful error message

**Browser URL Handling:**
- **file:// directory URLs** - Now append `/index.risor` to file:// directory paths
  - Previously only HTTP(S) URLs got `/index.risor` appended
  - Now `file:///path/to/dir` → `file:///path/to/dir/index.risor`
  - Consistent behavior across all URL schemes

### Changed
- **Fyne upgraded to v2.8.0**
- **Table column resize performance** - Resizing a column no longer performs a
  per-cell `UpdateCell` VM round-trip, making resizes smoother on large tables

### Security
- **IO module disabled by default in fynerisor-browser** - The reference browser
  no longer enables the IO module (filesystem write access) by default, reducing
  the attack surface when loading untrusted remote scripts

## [0.6.1] - 2026-06-23

### Fixed

**Goroutine Safety:**
- **Fixed concurrent VM access in goroutines** - go() spawned goroutines that call GUI
  functions (e.g., `label.SetText()`) no longer cause concurrent VM access panics.
  GUI updates from goroutines are now properly routed through fyne.Do() to ensure
  thread-safe execution on the main UI thread.
- **Fixed stdout capture deadlock** - Large output from spawned goroutines no longer
  causes deadlocks when stdout capture is enabled. Output lines longer than buffer
  size are now handled gracefully.
- **Panic recovery in callbacks** - Script callbacks (button clicks, form submits, etc.)
  that panic are now caught and logged instead of crashing the application.

### Changed
- Goroutine stdout capture now routes through Go channel to avoid concurrent VM access
- Long output lines are truncated with warning instead of blocking

## [0.6.0] - 2026-06-15

### Added

**Keyboard Shortcuts:**
- **window.AddShortcut(shortcut, callback)** - Register global keyboard shortcuts
  - Simple string-based API: `"Ctrl+S"`, `"Alt+Shift+N"`, etc.
  - Works without menus - shortcuts are always active
  - Supports all modifier keys: Ctrl, Alt, Shift, Super/Cmd
  - Cross-platform modifier aliases: Ctrl/Control, Alt/Option, Super/Cmd/Command
  - Supports letters (A-Z), numbers (0-9), function keys (F1-F12)
  - Special keys: Return, Escape, Tab, Space, Backspace, Delete, arrows, etc.
  - Multi-modifier shortcuts: `"Ctrl+Shift+S"`, `"Alt+Shift+F1"`
- **window.RemoveShortcut(shortcut)** - Remove registered shortcuts dynamically
- **MenuItem.Shortcut** - Display shortcut hints in menu items
  - Shows shortcut text like "Ctrl+S" next to menu item labels
  - Display-only property (actual shortcuts registered via window.AddShortcut)
  - Example: 34-keyboard-shortcuts with menu integration

### Fixed

**Module Imports (`import()`):**
- **Module-level references now work** - Functions in an imported module can
  reference other module-level variables and functions (e.g.
  `let circleArea = (r) => PI * square(r)`). Previously such references resolved
  against the wrong globals array, returning incorrect values or crashing with
  `index out of range`. Exported functions are now executed inside their own
  module VM where global references resolve correctly.
- **Modules receive host globals** - Imported module VMs are now seeded with the
  same merged environment as the main script, so module functions can use
  `widget`, `http`, `os`, `sql`, etc. Use `WithGlobal("name", value)` (singular)
  for custom globals that imported modules should also access; opaque
  `WithRisorOptions(risor.WithEnv(...))` values are not forwarded into module VMs.

## [0.5.1] - 2026-06-12

### Added

**Widget Table Cells:**
- **Widget mode for tables** - Use any Fyne widget type in table cells
  - CreateCell callback for widget creation (called once per cell position)
  - UpdateCell callback for data binding (maps filtered rows automatically)
  - Supports all widget types: Button, Icon, Label, Entry, Check, Select, etc.
  - Canvas image support for displaying images in table cells
  - Example: 32-table-widgets with interactive buttons and icons
  - Example: 33-image-gallery with thumbnail images in tables

**Table Export Enhancements:**
- **Multi-format export** - Export table data to CSV, XLSX, or JSON
  - CSV: Standard comma-separated format
  - XLSX: Excel-compatible spreadsheet format
  - JSON: Structured data with column names as keys
  - Configurable export path and filename
  - Column selection for export
  - Export current page or all data
  - Default export directory: ./exports (current working directory)

**OS Module Enhancements:**
- **os.read_dir(path)** - List directory contents
  - Returns list of maps with entry.name and entry.is_dir fields
  - Available via gui.WithOS() option
- **os.getwd()** - Get current working directory
- **os.exec(command, args)** - Execute external commands

### Fixed

**Table Filtering with Widgets:**
- **Automatic row mapping** - Filtering now works correctly with widget mode
  - Filtered rows automatically map to original data indices
  - Widget state preserved during filtering
  - No script changes required - mapping is transparent
  - Widget placeholder columns (e.g., "[Icon]", "[Button]") excluded from filter search

### Changed
- Export functionality integrated into table widget footer
- Widget cells cached and reused for better performance

### Added

**Dialog Support:**
- **Dialog package** - Full Fyne dialog support for user interactions
  - Show* convenience functions: ShowInformation, ShowError, ShowConfirm, ShowFileOpen, ShowFileSave, ShowFolderOpen, ShowColorPicker, ShowForm, ShowCustom, ShowCustomConfirm
  - New* constructors for advanced control: NewFileOpen, NewFileSave, NewConfirm, NewCustom, NewColorPicker, NewForm
  - Dialog wrapper types with full API: FileDialog (SetFileName, SetFilter, SetLocation), ConfirmDialog (SetConfirmImportance), CustomDialog (SetButtons), ColorPickerDialog (Advanced mode, SetColor), FormDialog (Submit)
  - All callbacks properly marshalled to GUI thread
  - File dialogs return paths as strings, color picker returns RGB map
  - Example: 30-dialogs with both basic and advanced usage

**Widget Visibility Control:**
- **Hide() and Show() methods** - Added to all 48 widgets and containers
  - Dynamic visibility control for all UI elements
  - Thread-safe via fyne.Do()
  - Consistent API across buttons, labels, entries, forms, tables, containers, etc.

**Command Execution:**
- **Exec module** - Run external commands from Risor scripts
  - exec(["command", "args"]) - Execute and return result with stdout/stderr/pid
  - exec.command(["ls", "-la"]) - Create command object with control methods
  - exec.look_path("cmd") - Find command path in PATH
  - Result and Command types with methods: run(), output(), json()
  - Configuration via options map: dir, env, stdin, stdout, stderr
  - Available in both core and gui packages via WithExec()

## [0.5.0] - 2026-05-29

### Breaking Changes

**Package Restructure for Static Compilation:**
- Split fynerisor into `core` (headless) and `gui` (GUI) packages
- **All imports must be updated** - root package no longer provides functionality
- Use explicit package imports: `github.com/uidbz/fynerisor/gui` or `github.com/uidbz/fynerisor/core`

### Added

**Headless Mode with Static Compilation:**
- **`core` package** - Risor execution with zero Fyne dependencies
  - `core.NewContext()` for headless script execution
  - All modules available: HTTP, SQL, OS, Strings, Filepath, Time, IO
  - Enables static compilation without OpenGL/X11 dependencies
  - Perfect for CLI tools, batch processing, server-side scripts
- **`gui` package** - Full GUI capabilities with Fyne framework
  - All original functionality preserved
  - `gui.NewWindow()` and `gui.NewApp()` for GUI applications
  - Subpackages: widget, binding, container, canvas, chart
- **Backward compatibility** - Root package re-exports gui functionality
  - Existing imports continue to work: `import "github.com/uidbz/fynerisor"`
  - No code changes required for existing applications

### Changed
- Moved NAMING_CONVENTIONS.md to docs/ directory
- Updated documentation for new package structure
- Updated example 16-context-builder to use `core` package

### Migration Guide

**Update all imports - REQUIRED:**

```go
// OLD (v0.4.x)
import "github.com/uidbz/fynerisor"
fw := fynerisor.NewApp("My App", fynerisor.WithHTTP())
core.SetAppVersion("1.0.0")

// NEW (v0.5.0) - GUI applications
import (
    "github.com/uidbz/fynerisor/core"
    "github.com/uidbz/fynerisor/gui"
)
fw := gui.NewApp("My App", gui.WithHTTP())
core.SetAppVersion("1.0.0")
```

**For headless/static compilation:**
```go
// Use core package for CLI tools and server-side scripts
import "github.com/uidbz/fynerisor/core"
ctx := core.NewContext(core.WithHTTP(), core.WithSQL())
result, err := ctx.Eval(`print("Hello!")`)
```

## [0.4.2] - 2026-05-19

### Fixed
- **CLI installation** - Fixed `go install github.com/uidbz/fynerisor/cmd/fynerisor@latest` 
  - Previous v0.4.1 tag was cached incorrectly by Go proxy
  - CLI tool now properly installable via go install
- **Missing example** - Added 01-hello-world example that was blocked by .gitignore
- **Documentation** - Added syntax highlighting to all Risor code blocks (using js highlighting)
- **Git author** - Fixed commit attribution for GitHub

### Added
- **CLI theme flag** - Added `--theme` flag to CLI for dark/light theme selection
- **Logo and screenshot** - Added visual assets to README
- **Prerequisites documentation** - Added link to Fyne prerequisites in installation section

### Summary
**Patch Release - CLI Installation Fix**
- ✅ CLI tool now works with `go install`
- ✅ All 29 examples included
- ✅ Complete and production-ready

## [0.4.1] - 2026-05-07 [YANKED]

**Note:** This version had issues with Go module proxy caching. Use v0.4.2 instead.

### Added

**Custom Struct Support:**
- **WithGlobal(name, object)** - Expose custom Go types with methods as global variables in Risor scripts
  - Implement `object.Object` interface on your Go types
  - Scripts can call methods on your custom objects (e.g., `db.query("SELECT * FROM users")`)
  - Example: 28-custom-struct - Complete pattern with UserDatabase

**Application Versioning:**
- **SetAppVersion(version)** - Allow embedding applications to control their own version checking
  - Separates application version from fynerisor library version
  - Scripts can use `require(["v2.5"])` to check compatibility with the host application
  - Example: 27-app-versioning

### Fixed
- Fixed custom-struct example to use Risor v2 syntax (`.each()` instead of for loops)
- Fixed example numbering - moved 28-imports to 14-imports to fill gap
- Added status callback to custom-struct example for better error visibility

### Documentation
- Added comprehensive custom struct integration guide
- Documented Risor v2 functional iteration patterns
- Added debugging section with status callback examples
- Updated WithGlobal() documentation in options.go

### Examples
- All 29 examples tested and working
- 28 widget/feature examples
- 1 custom struct integration example

### Summary
**Backward Compatible Release**
- ✅ Custom Go type integration via WithGlobal()
- ✅ Application versioning support via SetAppVersion()
- ✅ All examples updated and verified working
- ✅ Fully backward compatible with v0.4.0

### Added

**Custom Struct Support:**
- **WithGlobal(name, object)** - Expose custom Go types with methods as global variables in Risor scripts
  - Implement `object.Object` interface on your Go types
  - Scripts can call methods on your custom objects (e.g., `db.query("SELECT * FROM users")`)
  - Example: 28-custom-struct - Complete pattern with UserDatabase

**Application Versioning:**
- **SetAppVersion(version)** - Allow embedding applications to control their own version checking
  - Separates application version from fynerisor library version
  - Scripts can use `require(["v2.5"])` to check compatibility with the host application
  - Example: 27-app-versioning

### Fixed
- Fixed custom-struct example to use Risor v2 syntax (`.each()` instead of for loops)
- Fixed example numbering - moved 28-imports to 14-imports to fill gap
- Added status callback to custom-struct example for better error visibility

### Documentation
- Added comprehensive custom struct integration guide
- Documented Risor v2 functional iteration patterns
- Added debugging section with status callback examples
- Updated WithGlobal() documentation in options.go

### Examples
- All 29 examples tested and working
- 28 widget/feature examples
- 1 custom struct integration example

### Summary
**Backward Compatible Release**
- ✅ Custom Go type integration via WithGlobal()
- ✅ Application versioning support via SetAppVersion()
- ✅ All examples updated and verified working
- ✅ Fully backward compatible with v0.4.0

## [0.4.0] - 2026-05-06

### Added

**🎉 Data Binding System (Complete):**
- **binding.NewString()** - String data binding with optional initial value
- **binding.NewBool()** - Boolean data binding with optional initial value
- **binding.NewInt()** - Integer data binding with optional initial value
- **binding.NewFloat()** - Float data binding with optional initial value
- All bindings support: `Get()`, `Set(value)`, `AddListener(callback)`

**Widgets with Data Binding:**
- **widget.NewLabelWithData(binding.String)** - Auto-updating labels
- **widget.NewCheckWithData(label, binding.Bool)** - Bound checkboxes
- **widget.NewSliderWithData(min, max, binding.Float)** - Bound sliders
- Enables automatic two-way synchronization between data and UI

**Container Types (Complete 100% Coverage):**
- **container.NewCenter** - Center-aligned content layout
- **container.NewMax** - Maximum size layout (fills available space)
- **container.NewStack** - Layered stack of widgets
- **container.NewPadded** - Padded container with standard spacing
- **container.NewGridWithColumns** - Fixed column grid layout
- **container.NewGridWithRows** - Fixed row grid layout
- All 10 standard Fyne container types now available

**Widget Enhancements:**
- **Label.SetText()** - Method variant for setting label text
- **Button.SetText()** - Method variant for setting button text
- **Entry.OnChanged()** - Callback fires on every text change
- **Label Properties** - Wrapping, Truncation, Alignment, Importance now fully supported

**Constants & Enums:**
- **TextAlign** - Leading, Center, Trailing constants
- Complete constants example (example 24-constants)

**Examples:**
- 25-data-binding: Basic data binding with Label + Entry
- 26-data-binding-types: Comprehensive demo of all binding types

### Fixed
- Menu example (23-menu): Simplified to use button trigger instead of right-click
- Popup example (22-popup): Fixed variable scope issues in closures
- Constants example (24-constants): Added missing Label properties
- GridWrap example (17-gridwrap): Fixed overlapping text with proper Button sizing
- All 26 examples now working correctly

### Summary
**100% Feature Complete!**
- ✅ All high, medium, and low priority widgets (37 widgets)
- ✅ All container types (10/10)
- ✅ Complete data binding system (4 types)
- ✅ 26 working examples
- ✅ Production ready!

## [0.3.0] - 2026-05-06

### Added

**Advanced Widgets (95% Coverage Complete):**
- **GridWrap Widget** - Grid layout with virtualized rendering
  - `Length()`, `CreateItem()`, `UpdateItem()` callbacks
  - `Select()`, `Unselect()`, `UnselectAll()` methods
  - `OnSelected()`, `OnUnselected()` event handlers
  - Example: 17-gridwrap
  
- **TextGrid Widget** - Monospace text grid for code/tabular data
  - Line numbers and whitespace display options
  - Row operations: `SetRow()`, `Row()`, `RowCount()`
  - Configurable tab width
  - Example: 18-textgrid
  
- **RichText Widget** - Formatted text with markdown support
  - `ParseMarkdown()` for dynamic content
  - Text wrapping and truncation control
  - Scrollable content
  - Note: Direct segment manipulation not exposed (use markdown)
  - Example: 19-richtext

**Widget Enhancements:**
- **Button.Importance** - Visual hierarchy and semantic meaning
  - Standard levels: `ImportanceHigh`, `ImportanceMedium`, `ImportanceLow`
  - Semantic levels: `SuccessImportance`, `WarningImportance`, `DangerImportance`
  - Example: 20-button-importance
  
- **Button.Disabled** - Enable/disable button state programmatically
  
- **Entry.SetValidator()** - Custom validation with visual feedback
  - Return error string for invalid input
  - Return `nil` for valid input
  - Fyne shows red border automatically
  - Example: 21-form-validation

**Constants Global:**
- New global object `constants` for Fyne enums
- Currently provides button importance values
- Extensible for future constants

### Changed
- CLI module renamed from `fynerisor-cli` to `fynerisor`
- Installation instructions updated with `go install` command
- Documentation examples fixed for current API

### Fixed
- Module options (WithOS, WithHTTP, etc.) now work correctly in NewContext
- SQL query example no longer uses `.collect()` incorrectly
- Non-GUI usage examples corrected

## [0.2.1] - 2026-05-05

### Added
- **IO Module**: File operations support
  - `io.cp(src, dst)` - Copy files
  - Enabled via `WithIO()` option
  - Require with `@io` in scripts
  
- **App Metadata Object**: `app.name` global
  - Exposes embedding application name to scripts
  - Set via `WithAppName("myapp")` option
  - Default value: `"fynerisor"`
  - Enables conditional logic based on runtime environment
  
- **Go Function**: Spawn goroutines from scripts
  - `go(() => { ... })` - Run closure in background
  - Replaces removed `go` keyword from Risor v1
  - Safe for long-running operations
  
- **Window.Do()**: Queue GUI updates from background threads
  - `window.Do(() => { ... })` - Execute on GUI thread
  - Matches Fyne's threading model
  - Safe alternative to direct widget updates from goroutines
  
- **Window.Clear()**: Clear accumulated scripts
  - Clears all loaded scripts and imports
  - Critical for navigation-based apps (like goto)
  - Prevents script accumulation when loading new pages

### Fixed
- **window.OnDropped()**: Now properly registers with Fyne window
  - Fixed missing `SetOnDropped()` call
  - Drag-and-drop now works correctly
  
- **Options application order**: Fixed initialization sequence
  - AppName and callbacks applied before globals
  - Module options applied after base globals
  - Fixes issue where custom globals weren't available

### Changed
- Documentation restructured to `docs/` directory
- Updated README with current API and examples

## [0.2.0] - 2026-05-04

### Breaking Changes
- **Risor v2 Migration**: All scripts must use arrow function syntax
  - `func()` keyword removed, use `() => {}` or `(x) => x`
  - No `for`/`while` loops, use `.map()`, `.each()`, `.filter()`, `.reduce()`

### Added
- **ContextBuilder for non-GUI usage**
  - `NewContext()` creates Risor contexts without GUI dependencies
  - Access to all modules (HTTP, SQL, OS, etc.) and import system
  - `.Eval()` and `.EvalWithImports()` methods
  - Context-specific options: `WithContextHTTP()`, `WithContextSQL()`, etc.
  - Enables headless scripts, CLI tools, and server applications
  - Example: 16-context-builder
- **SQL Module**: Database connectivity support
  - MySQL, PostgreSQL, SQLite, SQL Server drivers
  - Connection via `sql.connect(dsn)`
  - Query execution: `conn.query()`, `conn.exec()`
  - Row iterator with `.collect()` method for list conversion
  - Example: 15-sql-test
  - Integrated into CLI with `fynerisor.WithSQL()` option

- **Tree widget** with hierarchical data support
  - ChildUIDs, IsBranch, CreateNode, UpdateNode callbacks
  - OpenBranch, CloseBranch, Select, Unselect methods
  - Example: 12-tree

- **List widget** for scrolling lists
  - Length, CreateItem, UpdateItem, OnSelected callbacks
  - Example: 13-list

- **SQL row iterator streaming methods**
  - `.map(fn)` - Transform rows in single pass (efficient)
  - `.each(fn)` - Iterate with side effects
  - `.collect()` - Load all rows into list (when needed)

- **Enhanced require() function**
  - Accepts string or list of requirements
  - Version requirements: `require("v0.2.0")` or `require("==v0.2.0")`
  - Module requirements: `require("@sql")`, `require("@http")`, etc.
  - Combined: `require(["v0.2", "@sql", "@http"])`
  - Validates enabled modules at script startup

- **AnalyzeRequirements() function**
  - Parse script source to extract requirements
  - Returns Requirements struct with version and module info
  - Useful for external tools and script validation
  - Example: `reqs, _ := fynerisor.AnalyzeRequirements(script)`

- **NewApp() convenience function**
  - One-line creation of Fyne app and fynerisor window
  - Simplifies common use case of creating a new application
  - Example: `fw := fynerisor.NewApp("My App", fynerisor.WithHTTP())`
  - Use NewWindow() for advanced scenarios requiring app customization

- **Window.Resize() method**
  - Set custom window size after creation
  - Example: `fw.Resize(800, 600)`
  - Works with both NewApp and NewWindow

- **import() global function**
  - Scripts can declare imports: `import("utils.risor")` or `import(["a.risor", "b.risor"])`
  - Supports local files and HTTP URLs
  - No-op at runtime (marker for static analysis)
  - CLI automatically handles imports via AnalyzeRequirements
  - Imports loaded in declaration order before script execution

- **AnalyzeRequirements() returns imports**
  - Requirements.Imports field lists all import paths/URLs
  - Enables automatic dependency resolution
  - Example: `reqs.Imports = ["utils.risor", "http://..."]`

- CLI error reporting without --verbose flag
  - All ERROR messages now display by default
  - --verbose still shows full execution status

- Comprehensive llms.txt for LLM code generation
  - Complete widget reference
  - Risor v2 syntax examples
  - Example prompts for GUI creation
  - Common patterns and constraints

### Fixed
- **CRITICAL**: Fixed race condition in Tree and List widgets causing VM stack corruption
  - Changed UpdateNode, OnSelected, OnUnselected in Tree to synchronous calls
  - Changed UpdateItem, OnSelected in List to synchronous calls
  - Issue: Mixed sync/async callbacks caused concurrent VM access
  - Symptom: `panic: index out of range [-1]` at `vm.pop()` after 30-100 interactions
  - See CONCURRENCY.md for details and guidelines
  
- Fixed Tree widget map access for Risor v2
  - Use `.get()` method instead of direct indexing (which throws errors)
  - Risor v2 throws errors on missing keys, unlike Risor v1
  
- Fixed accordion example (08-accordion)
  - Removed `while` loops (not supported in Risor)
  - Risor only supports functional iteration (for...in, map, filter)
  - Simplified example to focus on core accordion features

### Changed
- All Tree and List widget callbacks now use synchronous pattern
- Updated version requirement system
  - `require("v0.2")` for minimum version
  - `require("==v0.2.0")` for exact match
  - Added clear error messages for version mismatches

### Documentation
- Added CONCURRENCY.md with threading guidelines for widget development
- Added comprehensive widget audit (all 28 widgets checked)
- Updated examples with Risor v2 syntax
- Added version checking to relevant examples

## Previous Releases

See git history for earlier changes.
