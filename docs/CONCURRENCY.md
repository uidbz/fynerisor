# Risor VM Concurrency Guidelines

## Critical: Risor VM is NOT Thread-Safe

The Risor VM **does not support concurrent access** from multiple threads or goroutines. All VM operations (including `callFunc()` invocations) must be serialized.

## The Bug We Fixed

### Affected Widgets
- ✅ **Tree widget** - Fixed in commit 652a650
- ✅ **List widget** - Fixed in commit 5721fc4
- ✅ **Accordion widget** - Safe (no callbacks)
- ✅ **All other widgets** - Audited and safe

### Root Cause

Widgets that use Risor callbacks had a **mixed threading pattern**:

```go
// ❌ DANGEROUS: Mixed sync/async pattern

// Called synchronously from Fyne's GUI thread during layout/render
widget.PropertyGetter = func() T {
    result := callFunc(ctx, closure, args)  // VM access from thread 1
    return convert(result)
}

// Called asynchronously via channel from VM goroutine
widget.EventCallback = func() {
    w.Do(func() {
        callFunc(ctx, closure, args)  // VM access from thread 2
    })
}
```

**Problem:** When Fyne refreshed the widget (e.g., during tree.walk()), both threads would access the VM concurrently:
- **Thread 1 (GUI):** Calling IsBranch/ChildUIDs synchronously
- **Thread 2 (VM goroutine):** Processing queued UpdateNode/OnSelected callbacks

**Result:** VM stack corruption → `panic: index out of range [-1]` at `vm.pop()`

## The Fix

Changed all callbacks to use **consistent synchronous pattern**:

```go
// ✅ SAFE: All synchronous

widget.PropertyGetter = func() T {
    result := callFunc(ctx, closure, args)
    return convert(result)
}

widget.EventCallback = func() {
    callFunc(ctx, closure, args)  // Direct call, no queue
}
```

All VM calls now happen on the same thread (Fyne's GUI thread), eliminating concurrent access.

## Guidelines for Widget Development

### Rule 1: Identify All Callbacks

When adding Risor callbacks to a widget, list them all:
- Property getters (called by Fyne during layout/render)
- Event handlers (called by user interaction)
- Update callbacks (called during refresh)

### Rule 2: Determine Threading Pattern

**If widget has BOTH property getters AND event/update callbacks:**
- ✅ Use all-synchronous pattern (direct `callFunc()` calls)
- ❌ Never mix sync and async

**If widget has ONLY event callbacks:**
- ✅ Can use async pattern (`w.Do()`) consistently
- Examples: Button, Entry, Slider, Calendar

**If widget has NO callbacks:**
- ✅ Safe by default
- Examples: Label, Icon, Separator, Accordion

### Rule 3: Use This Template

For widgets with callbacks:

```go
case "CallbackName":
    return object.NewBuiltin("widget.CallbackName", func(ctx context.Context, args ...object.Object) (object.Object, error) {
        // Validate args...
        
        fn, ok := args[0].(*object.Closure)
        if !ok {
            return nil, fmt.Errorf("expected closure")
        }
        
        callFunc, ok := object.GetCallFunc(ctx)
        if !ok {
            return nil, fmt.Errorf("unable to get call function")
        }
        
        // For property getters or any callback that might be called during render:
        widget.Callback = func(params) ReturnType {
            // Call synchronously - no w.Do()!
            result, err := callFunc(ctx, fn, []object.Object{...})
            if err != nil {
                // Handle error
                return defaultValue
            }
            return convertResult(result)
        }
        
        return object.Nil, nil
    }), true
```

### Rule 4: Test Thoroughly

For any widget with callbacks:
- Test with 100+ rapid interactions
- Monitor for stack underflow panics
- If crashes occur → check for mixed threading pattern

## The Second Bug We Fixed: go() + stdout capture (commit 9fa2634)

The widget-callback rules above assume VM calls happen on the GUI thread. But
there are actually **two independent entry points into the VM**, and only one is
serialized:

1. **The `functionCalls` channel** — consumed serially by the `Execute()` loop in
   `window.go`. Button taps, `window.Do`, `OnDropped`, and shortcut handlers all
   funnel through here. This path is single-threaded. Safe.
2. **The `go(() => {...})` builtin** (`newGoBuiltin` in `window.go`) — spawns a
   **raw goroutine** that calls the closure via `safeCall` directly, bypassing
   `functionCalls`. This path is NOT serialized against anything.

### Root Cause

`window.CaptureStdout` originally took a Risor closure and invoked it once per
captured output line, through the VM:

```go
// ❌ DANGEROUS: capture callback re-enters the VM
w.onNewStdout = func(text string) {
    safeCall(callFunc, ctx, fn, []object.Object{object.NewString(text)})
}
```

When a script ran a worker via `go()` that produced output:

- **Goroutine 1 (go worker):** executing VM bytecode for the job closure.
- **Goroutine 2 (Execute loop):** running the stdout-capture closure in the SAME
  VM, triggered by a `print()` from goroutine 1.

Two `callFunction` invocations, one VM, no lock → corrupted `vm.fp` / stack →

```
go routine error: runtime error: unknown opcode: 0
recovered panic in script callback: runtime error: index out of range [21] with length 1
```

This is the same class of bug as the widget one (concurrent VM access), but the
second goroutine is the `go()` worker rather than Fyne's render thread, so the
"all-synchronous callback" fix does not apply.

### The Fix

Captured stdout no longer enters the VM at all. `CaptureStdout` accepts a `Log`
widget directly and appends lines via a **pure-Go sink** (`Log.AppendGo`):

```go
// ✅ SAFE: Go sink, never touches the VM
if logWidget, ok := args[0].(*risorwidget.Log); ok {
    w.stdoutSink = logWidget.AppendGo   // plain Go append + guithread.Do refresh
    w.captureStdout()
    return object.Nil, nil
}
// (legacy closure form retained but documented as unsafe with go()+output)
```

Risor script side:

```risor
let log = widget.NewLog(1000)
window.CaptureStdout(log)   // pass the widget, NOT a closure
```

### Rule 5: Never re-enter the VM from a goroutine that can race a go() worker

Any Go callback that may fire *while a `go()` closure is executing* (stdout
capture, timers, network completion handlers, anything not gated by
`functionCalls`) must NOT call back into Risor. Do the work in Go, or marshal it
onto the `functionCalls` channel — but be aware that `go()` workers run
concurrently with `functionCalls` consumption, so even the channel does not
serialize against an in-flight `go()` closure. When in doubt, keep it in Go.

## Safe Patterns Summary

| Pattern | Use When | Example |
|---------|----------|---------|
| **All Sync** | Property getters + events | Tree, List |
| **All Async** | Event callbacks only | Button, Slider, Entry |
| **No Callbacks** | Display/control only | Label, Icon, Accordion |
| **Go-only (no VM)** | Callback can fire during a `go()` worker | stdout capture (Log.AppendGo) |

## What About Performance?

**Q: Won't synchronous calls block the GUI?**

**A: No, for these reasons:**

1. **Callbacks are fast** - Typically just map lookups and simple logic (~microseconds)
2. **GUI already blocks** - Tree/List refresh was already blocking while waiting for results
3. **Matches Fyne design** - Built-in Fyne widgets use synchronous property getters
4. **No observable latency** - Tested with 100+ rapid clicks, no performance issues

## Testing Checklist

When adding/modifying widgets with Risor callbacks:

- [ ] Identify all callbacks (getters, events, updaters)
- [ ] Determine which pattern to use (all-sync vs all-async)
- [ ] Implement callbacks using consistent pattern
- [ ] Test with 100+ rapid interactions
- [ ] Verify no panics or stack corruption
- [ ] Document threading pattern in code comments

## Example: Correct Implementation

```go
// Tree widget - all synchronous (correct)
case "IsBranch":
    // ...
    t.instance.IsBranch = func(uid TreeNodeID) bool {
        result, _ := callFunc(ctx, fn, []object.Object{...})
        return result.IsTruthy()
    }

case "UpdateNode":
    // ...
    t.instance.UpdateNode = func(uid TreeNodeID, branch bool, obj CanvasObject) {
        // Synchronous - matches IsBranch pattern
        callFunc(ctx, fn, []object.Object{...})
    }
```

## References

### Widget callback bug
- **Issue:** Stack underflow after 30-100 widget interactions
- **Error:** `panic: runtime error: index out of range [-1]` at `vm.(*VirtualMachine).pop()`
- **Root cause:** Concurrent VM access from GUI thread and VM goroutine
- **Solution:** All-synchronous callback pattern
- **Commits:** 652a650 (Tree), 5721fc4 (List)
- **Audit:** WIDGET_AUDIT.md in bug report directory

### go() + stdout capture bug
- **Issue:** GUI froze (spinner stuck) when a `go()` worker produced output via `print()`; only seen in the goto pdf-table-extract app, never in headless CLI runs
- **Error:** `runtime error: unknown opcode: 0` / `index out of range [21] with length 1`, panicking from the stdout-capture callback while the go() worker stepped the VM
- **Root cause:** `CaptureStdout` invoked a Risor closure per output line, re-entering the single non-threadsafe VM concurrently with the `go()` worker
- **Solution:** Route captured stdout through a pure-Go sink (`Log.AppendGo`); `CaptureStdout(log)` takes the widget directly
- **Commit:** 9fa2634
