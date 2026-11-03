# 44 — Scene page

A scene page is a JSON document rendered by `scene/render`; the optional
`page.risor` sidecar adds behaviour by node ID. This example hosts the page
directly from Go, which is the same flow the browser uses.

```bash
go run ./examples/44-scene-page
# or, from this directory:
go run .
```

Then open the page in the designer:

```bash
go run ./cmd/fynerisor-designer examples/44-scene-page/page.scene.json
```

What to try:

- Click **Call the script** — the sidecar handler changes the title and the
  light color.
- Click the circle — declarative `toggle` and `set` actions run without script.
- Click the hotspot — another declarative `set` changes the hint text.
- Hover the hotspot to see the hover fill.

The document itself contains no code. `page.risor` is resolved relative to
`page.scene.json` and receives the `page` object when a browser (or this
example) opens the page. `window.SetContent(page)` is appended by the host so a
sidecar that only binds handlers still displays the page.

Files:

- `page.scene.json` — layout, components, actions.
- `page.risor` — sidecar script (`page.Bind`, `page.Node`).
- `main.go` — decode, build the page, execute the sidecar, show it.

See [docs/SCENES.md](../../docs/SCENES.md) for the format and
[docs/DESIGNER.md](../../docs/DESIGNER.md) for editing.
