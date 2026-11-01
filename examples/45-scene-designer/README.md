# 45 — Embedding the scene designer

`designer.Editor` is an ordinary Fyne canvas object. This example opens an
editor for a scene file (creating a small demo on first save) and wires the
host callbacks a real application needs: status, dirty state and close.

```bash
go run ./examples/45-scene-designer
# edit a specific file
go run ./examples/45-scene-designer /tmp/my.scene.json
```

Try:

- Click **Palette → Hotspot**, then drag a rectangle on the canvas.
- Select it and edit X/Y/W/H, props and events in the Inspector.
- Drag it, resize it with a handle, Ctrl+Z, Ctrl+D, Delete.
- Open **Outline** to hide/lock nodes; locked nodes are click-through.
- **Preview** runs declarative actions (there is no navigator in this example).
- **Done** asks to save; **Save** writes `designer-demo.scene.json` next to the
  working directory.

The standalone binary (`go run ./cmd/fynerisor-designer file.scene.json`) is the
same editor without the surrounding example.

See [docs/DESIGNER.md](../../docs/DESIGNER.md) for the full guide and
[docs/SCENES.md](../../docs/SCENES.md) for the document format.
