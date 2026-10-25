# 43 - Embedded CanvasObject

Shows how to run a Risor script inside an existing Fyne application instead of
letting fynerisor own the whole window.

The Go program builds the window, the header and a native "Go button". The
script is hosted by a `gui.CanvasObject`, which is added to a normal
`container.NewBorder` layout. Script content created with `window.SetContent`
is rendered inside that object only.

```go
view := gui.NewCanvasObject(script, gui.WithHostWindow(host))
view.Execute()
host.SetContent(container.NewBorder(header, nil, nil, nil, view))
```

`WithHostWindow` is optional. It enables window-level script features
(`dialog.*`, `window.AddShortcut`, `window.SetMainMenu`, `window.Canvas`,
`window.Close` and `window.OnDropped`). Without it the view is fully
self-contained, but those functions return script errors.

Run from this directory:

```bash
go run main.go
```
