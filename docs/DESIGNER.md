# The scene designer

`designer` is the in-place WYSIWYG editor for [scene documents](SCENES.md). It
shows the real Fyne widgets, not a mock-up, under a single input-shielding
overlay and edits a document through undoable operations. The same editor is
used by:

- the standalone tool `fynerisor-designer page.scene.json` (or
  `go run ./cmd/fynerisor-designer page.scene.json` from a checkout);
- the browser: press **F4** or the pencil button on a writable scene page.

A missing file starts a blank scene; the first Save creates it.

## Quick tour

```
┌ ↶ ↷ │ 💾 │ ⧉ 🗑 │ ⬆ ⬇ │ ▶  grid ┐                              Done
├ Palette ─┬─────────────── canvas ───────────────┬ Inspector ─────────────┤
│ Layout   │                                      │ Placement in free      │
│ Display  │        [ live scene widgets ]        │ x 1210  y 300          │
│ Shapes   │        [ selection + handles ]       │ 140 × 380  anchor …    │
│ Input    │                                      │ Content / Look / Data  │
│ Nav      │                                      │ Events                 │
├ Outline ─┤                                      │ Problems               │
└──────────┴──── status: hotspot · door · x … ────┴────────────────────────┘
```

- **Palette** — click a component, then click or drag a rectangle on the canvas.
  The tool stays armed until you place something or press Esc.
- **Canvas** — the live scene. Click to select, drag to move, drag a handle to
  resize, drag on empty space to marquee-select.
- **Outline** — the node tree with hide/lock markers and up/down/delete
  buttons. Selecting here and on the canvas stay in sync.
- **Inspector** — schema-driven fields for the selected node, grouped Content,
  Layout, Look, Data; placement adapts to the parent container; event actions
  are edited at the bottom.
- **Problems** — validation errors and warnings; clicking an issue selects its
  node. Saving refuses to proceed while there are errors.
- **Preview** — hides the editor chrome, runs the page in play mode and lets
  declarative actions and navigation run. Press the preview button again to
  return to editing.

## Selection and gestures

- Click selects the topmost unlocked, visible node.
- Shift/Ctrl/Cmd-click toggles selection.
- Dragging a selected node moves all selected free-positioned siblings together.
- Dragging a node that lives in an ordered container reorders it; dragging it
  over another container reparents it at the highlighted insertion point.
- Eight handles resize a node in a free container. Shift keeps the aspect ratio
  on corner handles. Resizing is disabled for layout-managed children (their
  parent owns the size); use `minW`/`minH` hints instead.
- Drag on empty canvas draws a marquee; nodes it touches are selected on
  release.
- Alt disables snapping while moving/resizing (held at interaction time).
- Esc selects the parent; Esc with no selection clears; Esc with a palette tool
  disarms it.
- Right-click opens duplicate/delete/z-order/select-parent actions.
- Lock (`outline` lock button) makes a node click-through, which is how a
  background image stays out of the way of its hotspots.
- Hidden nodes are skipped by hit-testing; use the outline to select them.

### Snapping and guides

- Grid: 8 design units by default; the toolbar toggles it off/on.
- Sibling edges and centers: left/center/right and top/middle/bottom of siblings
  snap within 6 screen pixels. Matched guides are drawn across the canvas.
- The grid is expressed in the parent's design units; on a scaled scene the
  on-screen step follows the scale.
- Arrow keys nudge by 1 design unit (grid size when the grid is on); Shift+arrow
  nudges 8×.

## Containers

The palette has four container families:

| Family | Components | Placement of children |
|---|---|---|
| Free | `free` | x/y/w/h, z-order; the point-and-click canvas |
| Ordered | `vbox`, `hbox`, `grid`, `gridwrap`, `stack`, `center`, `padded` | document order; insertion bar while dragging |
| Slots | `border`, `hsplit`, `vsplit`, `scroll`, `card` | named slot (`top`, `center`, `leading`, …); empty slots are highlighted |
| Tabs | `tabs`, `accordion` | one child per tab; the child's `title` names the tab |

Nesting works in both directions: a `vbox` can sit at an x/y position on a free
canvas, and a free canvas can fill a tab. A free canvas inside a vertical box
needs a `minH` hint because Fyne has no height-for-width pass.

## Inspector

### Placement

The placement section depends on the parent:

- free parent: X, Y, Width, Height, horizontal/vertical anchors;
- ordered/tabs parent: minimum width/height; tab title/icon for tab parents;
- slots parent: slot selector and minimum size.

Values commit as you type (coalesced into one undo step per field).

### Properties

Fields are generated from the component spec:

| Kind | Editor |
|---|---|
| string/asset/url/scene ref/icon | entry (+ host asset picker when configured) |
| text | multi-line entry |
| number/int | entry (validated on change) |
| bool | checkbox |
| enum | select |
| color | entry + color picker dialog |
| options/tags | comma-separated list |
| node ref | select of node IDs |

### Events

Each event lists its actions. Every action has a verb, a target (URL for
`navigate`, handler for `call`, node for the rest), and a delete button. “+ action”
adds one. The full verb list is in [SCENES.md](SCENES.md#actions).

## Clipboard, undo and saving

- Ctrl+C/X/V — canonical JSON fragments on the system clipboard, so nodes can be
  pasted between designer windows. Pasted nodes get fresh IDs (`…-copy`), are
  offset by +10 in free containers, and select themselves.
- Ctrl+D — duplicate the selection with a +10 offset.
- Ctrl+Z / Ctrl+Y — undo/redo. Consecutive changes to the same field or drag are
  coalesced into one step (one second window).
- Ctrl+S — validate, encode canonically and save. The title shows `•` while
  dirty.
- Saving is asynchronous; conflict detection uses the store's version. On
  `ErrConflict` the status reports the conflict (Reload/Overwrite/Save-as-copy
  UI is on the roadmap).
- Leaving with unsaved changes asks Save/Discard. The browser additionally
  asks before a navigation would replace an open editor (`BeforeLeave`).

Keyboard shortcuts work while the overlay has focus (click the canvas once).
Ctrl+S is also available from a main menu shortcut; a focused inspector entry
forwards it to the editor.

## Architecture

```
Editor
├── document + scene.History          pure Go model, ops/Tx, undo
├── render.View                       live Fyne tree (same as play mode)
├── overlay                           the only input receiver in edit mode
├── palette / outline / inspector     document views and commands
└── store                             FileStore / HTTPStore / tie store
```

**Why one overlay.** Fyne hit-tests the last child of a container first and
ignores parent bounds. A single transparent overlay as the last child of a
`Stack{View, overlay}` therefore receives every tap, drag, hover, wheel and
cursor query. The live widgets underneath never act on input, and the tree is
exactly the play-mode tree — true WYSIWYG with no per-node shields.

**Coordinates.** Overlay-local pixels are converted to a free parent's design
units with the parent's `render.Transform` (scale, offset, zoom). Move math adds
`DragEvent.Dragged` deltas to the geometry captured at press time; it never
subtracts absolute screen positions, which is what made early absolute-layout
experiments drift and catch toolbar offsets.

**Previews.** During a drag the overlay asks the parent's `ContainerLive` to
preview a placement. The model is not touched until release; a layout pass in
the middle of a drag (window resize) therefore cannot snap the node back.

**Operations.** Every edit becomes a `scene.Tx` with invertible `scene.Op`s.
`Document.Apply` is atomic (clone, apply, swap), returns `[]scene.Change`, and
`History` stores the undo steps. The same change stream drives `View.Apply`,
the outline, the inspector and the problems panel. A drag is one transaction,
so one Ctrl+Z undoes the whole gesture.

**Threading.** The editor runs entirely on Fyne's UI thread: gestures, document
mutation, rendering and panel updates. Asset loads run in goroutines and return
through `guithread.Do`; saves run in a goroutine with a version check and report
back on the UI thread. Entering edit mode calls `Window.Release()`, stopping the
page script and removing the run's shortcuts/menu/handlers; leaving reloads the
page. The VM and the editor never run at the same time.

## Embedding the editor

```go
doc, err := scene.Decode(data)
if err != nil { /* show the parse error */ }

ed := designer.New(doc, designer.Options{
    Registry: renderbuiltin.NewRegistry(), // or your registry with tie.* components
    Store:    scene.FileStore{},
    Ref:      "file:///srv/scenes/house.scene.json",
    Version:  version,                     // from Store.Load
    Window:   win,                         // fyne.Window: dialogs + focus
    Navigator: nav,                        // preview-mode navigation
    Opener:    opener,                     // preview-mode media/open
    Status:    func(s string) { log.Println(s) },
    OnClose:   func(reload bool) { /* remove the editor or reload the page */ },
    OnDirty:   func(dirty bool) { /* window title / menu state */ },
})
win.SetContent(ed.Object())
```

`designer.Options.Registry` is the extension point: a host registers its own
components (tie-browser registers `tie.image`, `tie.gallery`, …) with
`render.Registry.Register(spec, factory)` or `Factory(type, factory)` when the
spec already exists. The designer palette, inspector, validation and problem
panel pick them up automatically.

The browser integration is a `browser.EditorProvider`: `CanEdit` only writable
scene pages, `Open` returns `ed.Object()`, and `BeforeLeave` delegates to
`Editor.BeforeLeave`. The browser supplies a toolbar pencil and F4.

## Status

Implemented in v0.12.0:

- free/ordered/slot/tab containers, selection, move, resize, marquee, z-order;
- reparenting, ordered reordering and palette drag-and-drop (click-to-place
  remains for touch);
- grid and sibling-edge snapping with guides; arrow nudging; larger touch
  handles on mobile;
- inline text editing on double-click (an entry over the node, commit on
  Enter/focus loss, Esc cancels);
- wrap/unwrap and free↔layout conversion from the context menu;
- schema-driven inspector, name/placement/props/events, polygon point handles
  with add/remove point, outline with hide/lock;
- problems panel, save blocking on errors;
- clipboard fragments, duplicate, delete, undo/redo/coalescing;
- save with version conflicts (Overwrite/Reload), dirty state, unsaved-changes
  prompt, 30-second autosaved drafts with restore;
- preview mode, device frames (Desktop/Tablet/Phone), 50–200% zoom with
  scrolling and theme text scaling;
- multi-document story map (loads linked scenes through the host resolver);
- style overrides (foreground/background/text size/bold/italic) applied through
  a Fyne theme override, host-resolved `embed`, file/tie scene thumbnails, and
  in-place browser editing (F4) with page reload.

Roadmap:

- native media playback in tie-browser (the current opener uses the system
  handler; pwplay/mpv integration is the next step);
- further component and mobile polish (the editor now uses a bottom-tab layout
  on touch drivers).
