# Risor VM Concurrency

**Read this before touching widget callbacks.**

## The rule

A Risor VM is single-threaded. Two goroutines running bytecode in the same VM
corrupt its stack and frame pointer. This shows up as misleading errors
(`function takes 1 argument (2 given)`, `unknown opcode: 0`,
`index out of range [-1]` in `vm.pop()`) or as crashes.

A fynerisor script is entered from many places: its main body, widget events
(taps, changes, selections, dialogs, menus, shortcuts, drops), Fyne's
render-time getters (`List.Length`, `Table.Data`, `Tree.ChildUIDs`, ...) and
`go()`. Every one of them goes through **`gui/vmexec`**, which guarantees
that exactly one goroutine is inside a run's VM at a time.

## The model: one run, one VM goroutine

Each `Window.Execute` / `Start` / `Run` (and `CanvasObject` equivalents) creates
a **`vmexec.Run`**:

```
                 ┌──────────── vmexec.Run (one per Execute) ────────────┐
 Fyne UI thread  │  context (cancelled by Stop / newer Execute / ctx)    │
 ──────────────  │  reentrant guard (one per VM — never per process)     │
 tap, change,  ──┼─▶ event queue ──▶ VM goroutine: main body, then       │
 dialog, menu    │   (Fire, never     events one at a time               │
                 │    blocks)                                            │
 List.Length,  ──┼─▶ TryCall: runs inline if the VM is free,             │
 Table.Data ...  │   otherwise ErrBusy → last-known value + refresh later│
                 └───────────────────────────────────────────────────────┘
```

- **Events run on the VM goroutine, never on the UI thread.** A widget
  handler is queued with `Callback.Fire`. The queue is bounded (100). The
  UI thread never blocks on it: when it is full, the event is dropped with a
  stderr diagnostic. Events queued while the main body runs execute after
  it, in order.
- **Render getters run synchronously** because Fyne needs an answer during
  layout. They use `Callback.Call`. If the VM is free (or already owned by the
  calling goroutine: a handler that calls `Table.Refresh()` re-enters its
  own getters), the getter runs inline. Otherwise it gets `vmexec.ErrBusy`
  at once and returns its last-known value.
  `Callback.RefreshWhenBusy(widget.Refresh)` re-renders the widget once the
  VM is idle again.
- **No deadlock between the two sides.** The UI thread never waits for the
  VM: it only enqueues or try-locks. The VM goroutine may wait for the UI
  thread (`guithread.DoAndWaitContext`), and gives up when its run is stopped.
- **Script code never runs on the UI thread**, so a slow handler or a long
  main body never freezes the window. It keeps repainting and accepting
  input; events wait their turn.
- **One guard per VM.** Two windows or embedded `CanvasObject`s never block
  each other.
- **Stopped runs are inert.** `Stop()`, a cancelled `Run(ctx)` context, or a
  newer `Execute` cancels the run's context. The VM halts; queued and later
  events of that run are dropped; `window.SetStatus`/`SetContent`/
  `SetMainMenu` from it are ignored. Its shortcuts, script-set main menu,
  drop/stdout handlers and imported modules are discarded. Callbacks find the
  run that registered them through the context Risor passes to builtins
  (`vmexec.FromContext`), so an old widget can never enter a newer run's VM.

### `go(fn)`

`go(fn)` queues `fn` on the VM goroutine. It runs after the current body or
event, like any other event (issue #17). It is **not** a goroutine: running a
closure in parallel with the script is exactly the corruption described
above. The window stays responsive while `fn` works, because script code
never runs on the UI thread. Real parallelism belongs in Go: a host global or
module does the work on its own goroutines and reports back with
`Callback.Fire`.

## Writing a widget callback

Use `vmexec.NewCallback`. Never call the `object.CallFunc` from
`object.GetCallFunc` yourself, and never start a goroutine that enters the VM.

```go
case "OnChanged":
    return object.NewBuiltin("widget.Slider.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
        if len(args) != 1 {
            return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
        }
        cb, err := vmexec.NewCallback(ctx, "widget.Slider.OnChanged", args[0])
        if err != nil {
            return nil, err
        }
        // Event: async, ordered, recovered, dropped once the run stops.
        obj.instance.OnChanged = func(v float64) { cb.Fire(object.NewFloat(v)) }
        return object.Nil, nil
    }), true
```

Getter (Fyne calls it during layout and needs a value back):

```go
cb, err := vmexec.NewCallback(ctx, "widget.List.Length", args[0])
if err != nil {
    return nil, err
}
cb.RefreshWhenBusy(obj.refresh) // func() { guithread.Do(obj.instance.Refresh) }

obj.instance.Length = func() int {
    result, err := cb.Call()
    if err != nil {
        cb.Report(err)   // silent for ErrBusy / ErrStopped
        return obj.lastLen
    }
    n, err := object.AsInt(result)
    if err != nil {
        cb.Report(fmt.Errorf("expected int, got %s", result.Type()))
        return obj.lastLen
    }
    obj.lastLen = int(n)
    return obj.lastLen
}
```

Checklist:

- [ ] Every script function goes through `vmexec.NewCallback`.
- [ ] Fyne events → `Fire`. Fyne getters (called during layout or render) →
      `Call` with a last-known fallback and `RefreshWhenBusy`.
- [ ] UI mutations from builtins go through `guithread.Do`. Reads that must
      return a value use `guithread.DoAndWaitContext(ctx, ...)`.
- [ ] Builtins that change *window-level* state (content, status, menu)
      check `ctx.Err()` so a stopped run cannot change it.
- [ ] Add a `-race` test. Fyne's test driver runs `fyne.Do` inline on the
      calling goroutine, so tests must get a happens-before edge with the VM
      goroutine before reading widget state: wait for `Status() == "Ready!"`,
      `State()`, `Run(ctx)`, or a recorder builtin called by the script.

## Lifecycle

`Window.State()` returns `StateIdle`, `StateRunning`, `StateReady`,
`StateFailed` (plus the body's error) or `StateStopped`.
`Run(ctx)` blocks until the main body has ended. `Start(ctx)` returns a
channel with that result. `WithOnStateChange` reports transitions of the
current run. Hosts should use these, not the `"Ready!"` / `"ERROR: ..."`
status text, which is for display and can be set by scripts.

## Headless Parallelism: `core.EvalBatch`

Everything above concerns the GUI, where one VM per run serves all callbacks.
The headless `core` package offers **safe, real multi-core parallelism** via
`core.EvalBatch`.

### The model: compile once, run many

`EvalBatch` compiles the script once into an immutable `*bytecode.Code`,
then runs it over each input on its **own fresh VM**. Compiled bytecode is
read-only and safe to share across goroutines; a running VM is never shared.
This is the pattern from Risor's own `examples/go/concurrent_vms`.

```go
results, err := core.EvalBatch(ctx, script, inputs,
    core.BatchConfig{Concurrency: 4}, core.WithCSV())
// results[i] corresponds to inputs[i]; results[i].Err isolates per-input failures.
```

### Why it is safe

- **Fresh `Context` per worker.** Each goroutine calls `NewContext(opts...)`,
  giving it its own `env`, module objects, module cache and import stack.
- **Immutable shared code.** All workers execute the same compiled bytecode.
  Each VM keeps its mutable globals per instance and never mutates the
  shared `Code`.
- **Warm-up before fan-out.** The first input runs synchronously to
  initialize Risor's lazily-created package globals (the default type
  registry). `vmexec.New` does the same for GUI runs, so two windows can
  start at once.

### Caveats

- **Stateful shared globals.** `WithGlobal(name, value)` passes one pointer to
  all workers. Stateless modules (`csv`, `json`, `strings`) are safe. A
  stateful custom global (DB handle, mutable map) must be made
  concurrency-safe by the caller.
- **Headless only.** Do not use `EvalBatch` to drive Fyne widgets.

### Why not a script-level `parallel(items, fn)` builtin?

To run a closure on a fresh VM, that VM must first `Run()` its root program,
which re-executes the main script's entire top level once per worker. Closures
also capture free variables (`*Cell`) that alias the defining VM's frame, so
invoking one from several goroutines is a data race. Parallelism belongs at
the Go/embedder boundary, where each unit of work is a whole script on an
isolated VM.

## History

These incidents shaped the model:

- **Widget callback stack underflow** (commits 652a650, 5721fc4). After 30–100
  interactions: `index out of range [-1]` in `vm.pop()`. Tree/List getters ran
  on the UI thread while queued handlers ran on the callback goroutine.
- **`go()` + stdout capture** (commit 9fa2634). `unknown opcode: 0` while a
  `go()` worker printed. `CaptureStdout` re-entered the VM per output line,
  concurrently with the worker. Captured output now goes through a pure-Go
  sink (`Log.AppendGo`).
- **Code review, October 2026** (issues #10–#17, #37, #39). The callback queue
  could freeze the UI. `Entry.OnChanged` and bindings started a goroutine per
  change. About 30 sites bypassed recovery and the guard. The guard was
  process-wide. Stopped runs kept running. `go()` raced the VM. All of these
  were replaced by the per-run `vmexec` model above.
