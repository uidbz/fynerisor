# Scroll container — scrolling bindings

## ScrollToBottom / ScrollToTop (v0.8.2+)

Wraps Fyne's `Scroll.ScrollToBottom`/`ScrollToTop` through `guithread.Do`.

```risor
let scroll = container.NewScroll(content)
scroll.ScrollToBottom()  // pin chat/log to newest content
scroll.ScrollToTop()
```

Guard for older builds (attribute resolves to nil when missing):

```risor
let f = scroll.ScrollToBottom
if (f != nil) { f() }
```

Call it right after `window.SetContent(...)` — SetContent marshals onto the
GUI thread and runs inline, so content/viewport sizes are settled by the time
the next statement executes. No delay/sleep needed.

## No sleep in scripts

The fynerisor `time` module (modules/time) exposes only now/date/parse/unix
+ Time objects — **no sleep**, and Risor v2 has no `sleep()` builtin. Never
script delays; sequence GUI calls inline instead.
