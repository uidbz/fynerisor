---
name: fynerisor
description: Expert guidance on fynerisor (Risor v2 bindings for Fyne GUI) - widgets, data binding, syntax, and best practices
---

# Fynerisor Development Guide

This skill provides expert assistance with fynerisor development, including:
- Risor v2 syntax and functional programming patterns
- Fyne widget creation and configuration
- Data binding for reactive UIs
- Custom global objects and integration
- Layout and container management
- Event handling and callbacks
- Modules (HTTP, SQL, OS, Time, Strings, Filepath, IO, JSON, CSV, Exec, Tie)

Current: v0.9.0, Risor v2.2+, Fyne v2.8+, module path `github.com/uidbz/fynerisor`
(host code imports the `core` and `gui` packages — the root package is docs-only).

## When to Use This Skill

This is the **canonical guidance** for Risor v2 syntax and fynerisor GUI
widgets. Use this skill when the user asks about:
- Risor v2 syntax questions (arrow functions, .each(), .map(), match, etc.)
- Fyne widgets in Risor context
- Data binding (String, Bool, Int, Float)
- Layout containers (Border, VBox, HBox, Grid, etc.)
- Widget properties and methods
- Form validation and user input
- Functional programming patterns
- Charts, tables, trees, keyboard shortcuts, menus

## Knowledge Base

Docs in the fynerisor repo are the single source of truth — this skill tells
you where to look rather than duplicating them:

| Question | Source |
|---|---|
| Exact attribute list of any script type | `docs/API.md` — **generated from the compiled registries, CI-pinned**; regenerate with `fynerisor docs` |
| Full guide: modules w/ functions, options, CLI | `docs/README.md` |
| Library embedding (Go host code) | `README.md` "Usage as a Library" |
| Concurrency model deep dive | `docs/CONCURRENCY.md` |
| Runnable examples | `examples/` (per-example READMEs) + `docs/EXAMPLES.md` |
| Tie triple store | `docs/TIE_MODULE.md` |
| What changed between versions | `CHANGELOG.md` |

Refer to `references/fynerisor-reference.md` for distilled operational
knowledge: semantics, gotchas, and idiomatic patterns not obvious from the
API listing. Refer to `references/scroll-bindings.md` for scroll-container
details (ScrollToBottom/Top, no-sleep rule).

When in doubt about an attribute or function name, check `docs/API.md` (or
run `fynerisor docs --json`) instead of guessing from memory.

## Key Concepts to Emphasize

### Risor v2 Syntax
- No C-style for loops - use `.each()`, `.map()`, `.filter()`
- Arrow functions: `(x) => x * 2` (there is no `func` keyword)
- Functional iteration over arrays
- `import("file.risor")` is live at runtime and returns a module namespace;
  `require([...])` gates version and modules (optional)

### Control Flow: `match` (not `switch`)

Risor v2 has no `switch`/`case`. Use `match` expressions with `=>` arms and `_`
as the default (a default arm is **required** or it's a parse error):

```risor
let label = match status {
    "done"    => "✓ done"
    "failed"  => "✗ failed"
    _         => status
}
```

**Gotcha — arms are single expressions, not blocks.** A `{ ... }` after `=>`
is parsed as a *map literal*, not a statement block, so multi-statement arms
fail with `unexpected ... while parsing map`. Put multiple statements in a
helper function and call it from the arm:

```risor
// WRONG: { ... } is read as a map literal
match col {
    0 => { cell.SetText("Start"); cell.OnTapped(fn) }   // parse error
    _ => cell.SetText("")
}

// RIGHT: single call per arm
function setBtn(cell, label, fn) {
    cell.SetText(label)
    cell.OnTapped(fn)
}
match col {
    0 => setBtn(cell, "Start", fn)
    _ => cell.SetText("")
}
```

This is the idiomatic replacement for the `if (col == 0) {...} else if (...)`
chains common in table `UpdateCell` callbacks. Note: `else` must sit on the
same line as the preceding `}` (`} else {`), so standalone `if`s or `match` are
usually cleaner.

### Common Patterns

**Basic Window Setup** — scripts get a `window` global; never create one:
```risor
let label = widget.NewLabel("Hello")
window.SetContent(label)
```

**Data Binding:**
```risor
let textBinding = binding.NewString("")
let entry = widget.NewEntryWithData(textBinding)
let label = widget.NewLabelWithData(textBinding)
```

**Callbacks:**
```risor
let button = widget.NewButton("Click", () => {
    window.SetStatus("Button clicked!")
})
```

**Layouts:**
```risor
let content = container.NewBorder(
    topWidget,    // top
    bottomWidget, // bottom
    leftWidget,   // left
    rightWidget,  // right
    centerWidget  // center
)
```

### Common Mistakes to Catch

1. **Using for loops instead of functional iteration**
   - Wrong: `for i in items { ... }`
   - Right: `items.each(item => { ... })`

2. **Forgetting arrow syntax**
   - Wrong: `func(x) { return x * 2 }`
   - Right: `(x) => x * 2`

3. **Treating `go()` as parallelism or GUI marshalling**
   - `go(fn)` queues `fn` on the script's *own* VM goroutine (runs after the
     current body/event, in order). It does not parallelize and long `go()`
     callbacks delay later callbacks of the same run.
   - **Widget calls need no wrapper anywhere**: every widget method marshals
     its own UI work (safe inside event handlers and inside `go()`).
   - `window.Do` does not exist — it was removed; `go(fn)` replaced it.
   - Real parallelism belongs in the Go host (a `WithGlobal` function that
     spawns goroutines and reports back through callbacks).

4. **Creating windows/apps from script code**
   - Wrong: `widget.NewApp()`, `window.ShowAndRun()` in a script
   - Right: use the `window` global; the host/CLI owns the lifecycle.

5. **Wrong namespace for bindings**
   - Wrong: `widget.NewString()`
   - Right: `binding.NewString()` (also NewBool/NewInt/NewFloat)

## Response Style

- Provide complete, working code examples
- Prefer the repo's generated `docs/API.md` when citing exact widget attributes
- Show before/after for corrections
- Highlight Risor v2 specific syntax differences from other languages
- Remember widget/container/binding/etc. are **globals** — no import needed
- Explain the "why" for patterns, not just the "how"

## Example Interactions

**User:** "How do I create a button in fynerisor?"

**Response:** "In fynerisor, create a button with a callback using `widget.NewButton()`:

```risor
let button = widget.NewButton("Click Me", () => {
    print("Button was clicked!")
})
window.SetContent(button)
```

The first argument is the button text, and the second is an arrow function that runs when clicked. Callbacks run on the script's VM goroutine, and the `SetText`/`SetStatus` style calls marshal their own UI work — no wrapper needed."

**User:** "Why isn't my for loop working?"

**Response:** "Risor v2 uses functional iteration instead of traditional for loops. Replace your for loop with `.each()`:

```risor
// Instead of: for item in items { ... }
items.each(item => {
    print(item)
})
```

Other functional methods available: `.map()`, `.filter()`, `.reduce()`"

**User:** "How do I bind an entry to a label?"

**Response:** "Use data binding to create reactive connections between widgets:

```risor
// Create a string binding
let textBinding = binding.NewString("")

// Create widgets that share the binding
let entry = widget.NewEntryWithData(textBinding)
let label = widget.NewLabelWithData(textBinding)

// When user types in entry, label updates automatically
```

The binding keeps both widgets synchronized - changes in one automatically update the other. Bindings expose only `Get`/`Set`/`AddListener` (Fyne 2.8 removed mappers — derive transformed text via `AddListener`)."
