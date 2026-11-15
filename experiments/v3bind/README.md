# v3bind: Risor v3 for page bindings

Status: small experiment. No Fyne, no scene document, no event loop. It exists
to test the one role v3 keeps in fynerisor after the editor decision:

> **Handlers stay on Risor v2. v3 is for checked, bounded binding
> expressions.**

This replaces the earlier `v3scene` prototype. That prototype proved the
host-dispatched event model works, but it also showed that a VB-style editor
wants mutable handlers, which is exactly what v3 does not do. The event code
was removed rather than kept as a second, more complicated way to build pages.

## What it does

`Compile[T]` derives the state type from a Go struct with `risor.TypeOf` and
declares each record field as a top-level variable. The binding text is the
same bare-name form as `${count}` in a scene property:

```go
type State struct {
    Count int
    Name  string
    Items []string
}

set, err := v3bind.Compile[State](map[string]string{
    "count":    "count",
    "title":    `"Count: " + string(count)`,
    "hasItems": "len(items) > 0",
}, v3bind.Config{})

out, err := set.Run(ctx, State{Count: 2, Items: []string{"a"}})
// out["count"] = 2
// out["title"] = "Count: 2"
// out["hasItems"] = true
```

Each binding is parsed and checked when `Compile` runs. A typo is a load error:

```
v3bind: bad: check undefined-name at [1, 1, 1, 5]: nope is not defined
```

The compiled programs are immutable and safe to run from many goroutines at
once. `RunOne` evaluates a single binding when a host only needs to refresh
one property.

## What it does not do

- No callbacks, handlers or stored closures.
- No scene document parsing.
- No Fyne dependency.
- No automatic renderer updates; that is the host's job in fynerisor.

That is deliberate. The editor's event workflow is v2 (`on_<node>_<event>` +
`page.Bind`), and `page.BindValues(() => ({...}))` is the v2 implementation
used today. This package checks whether the same binding text can move to v3
later.

## Run

```sh
cd experiments/v3bind
go test ./...
go test -race ./...
```

The module is pinned to the v3 Go implementation at
`github.com/deepnoodle-ai/risor/go/v3@v3.0.0-20261112045619-aeaaf1fd6c13`.
There is no tagged release for that module path yet.
