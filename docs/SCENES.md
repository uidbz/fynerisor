# Scene documents

A **scene** is a declarative, data-only description of a page: a tree of
components with positions, sizes, properties and behaviour. The designer edits
the document; the renderer turns it into live Fyne widgets; an optional sidecar
Risor script adds dynamic behaviour by node ID.

Scenes are the format behind the point-and-click idea in the README: a
background image with clickable hotspots that navigate to other scenes, media
taken from tie collections, and an Edit button on every page. The same format is
the page format of **tie-browser**, where documents and their media are stored
content-addressed in tie.

```
scene JSON ──► scene/render (Fyne) ──► live widgets ──► browser / tie-browser
     ▲                ▲
     │                └── optional sidecar Risor script (behaviour)
     └── designer (in-place WYSIWYG editor)
```

A page containing only a scene runs **no code at all**. That makes data-only
scenes the safe default for untrusted or remote content; scripts are opt-in per
host and per page.

## Page kinds

The browser accepts two page kinds:

| Kind | Loaded from | Rendered by |
|---|---|---|
| Risor page | any URL (`page.risor`, `index.risor`, …) | the classic `gui.Window` script run |
| Scene page | `*.scene.json` (or a body whose `format` starts with `fynerisor.scene/`) | `scene/render.View` + optional sidecar script |

`browser.Config.LoadPage` returns a `browser.Page` with the kind and bytes;
`browser.Config.RenderScene` displays it. Without a scene renderer a scene page
reports an error instead of rendering.

## Document format

The current format is `fynerisor.scene/1`.

```json
{
  "format": "fynerisor.scene/1",
  "id": "c1f0…",
  "title": "Living room",
  "script": "living-room.risor",
  "root": {
    "id": "root",
    "type": "free",
    "place": {"x": 0, "y": 0, "w": 1600, "h": 900},
    "props": {
      "designW": 1600,
      "designH": 900,
      "scale": "fit",
      "background": "tie:9a1f…",
      "backgroundFill": "stretch",
      "letterbox": "#000000"
    },
    "children": [
      {
        "id": "door",
        "type": "hotspot",
        "name": "Door to kitchen",
        "place": {"x": 1210, "y": 300, "w": 140, "h": 380},
        "props": {"shape": "rect", "hoverFill": "#ffffff30", "label": "Kitchen"},
        "on": {"tapped": [{"do": "navigate", "to": "../kitchen"}]}
      },
      {
        "id": "photo-wall",
        "type": "vbox",
        "name": "Photo wall",
        "place": {"x": 80, "y": 120, "w": 520, "h": 600},
        "children": [
          {"id": "wall-title", "type": "label", "props": {"text": "Summer 2014"}},
          {"id": "wall-gallery", "type": "tie.gallery", "place": {"minH": 480},
           "props": {"tags": ["family", "2014"]}, "on": {"itemTapped": [{"do": "open"}]}}
        ]
      }
    ]
  }
}
```

### Canonical encoding

Scenes are encoded canonically by `scene.Encode`:

- JSON object keys sorted (Go's encoder does this);
- struct fields in a fixed order, zero values omitted;
- numbers rounded to `0.01` design units;
- two-space indentation, LF line endings, trailing newline.

The same document always produces the same bytes. That keeps diffs stable and
makes content-addressed storage (tie, IPFS-style) meaningful: a document's
version is its hash.

### Document fields

| Field | Meaning |
|---|---|
| `format` | Format version. Unknown major versions are refused with a clear error. |
| `id` | Stable identity. In tie-browser this is the scene directory's `DirUID`; it survives edits and moves. |
| `title` | Human title (browser, tab, outline). |
| `script` | Optional sidecar Risor script, resolved relative to the document. |
| `meta` | Free-form string map for hosts. |
| `root` | The root node. It must be a container. |

### Node fields

| Field | Meaning |
|---|---|
| `id` | Unique within the document, `[a-z0-9][a-z0-9-]*`. Scripts and actions address nodes by ID. |
| `type` | Component type (see [API.md](API.md#scene-components)). |
| `name` | Display name in the outline; scripts see it as `evt.name`. |
| `place` | Placement, interpreted by the **parent's** container kind. |
| `props` | Component properties, validated against the component spec. |
| `style` | Reserved subtree style override (foreground, background, text size, bold/italic). Stored and validated; application to widgets is on the roadmap. |
| `on` | Event → list of declarative actions. |
| `hidden` | Hidden at start. |
| `locked` | Editor-only: clicks pass through the node, so a background image can be locked while its hotspots stay selectable. |
| `children` | Child nodes. Free parents use document order as z-order; ordered parents use it as layout order. |

Property names are lowerCamelCase and must not be Risor keywords (`in`, `for`,
`let`, `if`, `from`, `import`, …), because scripts receive them as maps.
Keys starting with `$` are reserved for future expressions/bindings.

## Placement

| Parent kind | Components | `place` fields | Children order |
|---|---|---|---|
| `free` | `free` | `x`, `y`, `w`, `h`, optional `anchor` | z-order (later = on top) |
| ordered | `vbox hbox grid gridwrap stack center padded` | `minW`, `minH` | layout order |
| slots | `border hsplit vsplit scroll card` | `slot` (+ `minW`, `minH`) | one per slot |
| tabs | `tabs accordion` | `title`, `icon` | one per tab/section |

The editor validates place fields against the parent: a `slot` inside a `vbox`
is reported as a warning, a child of a `border` without a slot is an error, and
duplicate slots are errors.

### Free canvases and scaling

`free` is the point-and-click canvas. Its props:

| Prop | Default | Meaning |
|---|---|---|
| `designW`, `designH` | 1600, 900 | Design-unit coordinate space. |
| `scale` | `fit` | `none` (1:1, anchors), `fit` (uniform, letterboxed), `fill` (uniform, cropped), `stretch` (non-uniform). |
| `background` | — | Asset reference drawn behind the children. |
| `backgroundFill` | `stretch` | `stretch`, `contain`, `cover`, `original`. |
| `letterbox` | `#000000` | Color behind/around the design rectangle. |
| `clip` | false | Clip children to the container. Changing it rebuilds the node. |
| `scaleText` | true | Canvas text follows the scale (widget text scaling is on the roadmap). |

`render.FreeLayout` computes a `Transform` from the window size:

```
fit:     s = min(W/dw, H/dh);  ox = (W - dw*s)/2;  oy = (H - dh*s)/2
fill:    s = max(W/dw, H/dh);  ox/oy centered (negative = crop)
stretch: sx = W/dw; sy = H/dh
none:    sx = sy = 1
```

Children (and the background) are placed with the **same transform**, so a
hotspot stays on the right pixel of the background at every window size. This is
the single most important invariant of point-and-click scenes. Use "fit scene to
background" (planned designer command) or set `designW/H` to the image's natural
size.

With `scale: none`, anchors keep edges fixed on resize:

| `anchor.h` | Effect |
|---|---|
| `left` | keep the left offset |
| `right` | keep the right offset |
| `center` | keep the offset from the center |
| `stretch` | keep both margins; the node grows |

`anchor.v` is `top`, `bottom`, `center` or `stretch`. Anchors are ignored in
scaled modes (scaling already keeps the composition).

### Nesting

Free and layout containers nest in both directions:

```
free root (fit)
├── hotspot                       x/y over the background
├── vbox "photo wall"             a layout container placed at x/y
│   ├── label
│   └── slot  (script fills it)
└── tabs "drawer"                 a tab container placed at x/y
    ├── free "map"                a free canvas inside a tab
    └── vbox "list"
```

Fyne has no height-for-width pass. A free container inside a `vbox`/`hbox` only
sees a minimum size along the main axis, so give it `place.minH` or use a
`border` center, `stack`, `split` or tab (they hand it the remaining space).

## Components

The authoritative component list, with every property, default, event and slot,
is generated into [API.md](API.md#scene-components) from
`scene/builtin`. The families are:

| Category | Components |
|---|---|
| Layout | `free`, `vbox`, `hbox`, `grid`, `gridwrap`, `stack`, `center`, `padded`, `border`, `hsplit`, `vsplit`, `scroll`, `card`, `tabs`, `accordion` |
| Display | `label`, `text`, `markdown`, `image`, `icon`, `separator`, `progress`, `activity` |
| Shapes | `rect`, `circle`, `line` |
| Input | `button`, `hyperlink`, `entry`, `check`, `select`, `radio`, `slider` |
| Navigation | `hotspot`, `link`, `embed`, `slot` |
| tie (tie-browser) | `tie.image`, `tie.gallery`, `tie.collection`, `tie.audio`, `tie.video`, `tie.document`, `tie.tags`, `tie.related` |

Notes:

- `hotspot` is the point-and-click primitive. It emits `tapped`,
  `doubleTapped`, `secondaryTapped`, `hoverIn`, `hoverOut`, shows a pointer
  cursor, and is invisible in play mode unless `showInPlay` is set. With
  `shape: polygon`, `points` is a list of normalized `{x,y}` vertices (0..1);
  the designer shows draggable point handles and add/remove commands.
- `link` is a button that navigates to its `to` prop when it has no `tapped`
  action. It is the terse form of a scene link.
- `slot` is a runtime hole. In edit mode it shows its `hint`; in play mode it is
  empty until the page script calls `node.SetContent(widget)`.
- `embed` hosts a nested scene document. The host supplies a
  `render.EmbedResolver`; the default `StoreResolver` loads file/http docs
  (with a remote-page origin check), and tie-browser passes the tie store.
- Images, shapes and icons have an invisible event surface when they carry
  actions; entering or leaving a shape emits hover events.

## Actions

Actions are declarative behaviour stored in the document. They run on the UI
thread, in order, before script listeners for the same event.

```json
"on": {
  "tapped": [
    {"do": "navigate", "to": "../kitchen"},
    {"do": "hide", "node": "help"},
    {"do": "set", "node": "title", "prop": "text", "value": "Kitchen"},
    {"do": "call", "handler": "toggleMusic"}
  ]
}
```

| `do` | Fields | Effect |
|---|---|---|
| `navigate` | `to` | Resolve `to` against the document URL and ask the host navigator to open it. |
| `back` | — | Ask the host navigator to go back. |
| `show`, `hide`, `toggle` | `node` (empty = self) | Change runtime visibility. |
| `set` | `node`, `prop`, `value` | Set a property at runtime (not saved to the document). |
| `call` | `handler` | Invoke a handler registered with `page.Bind`. |
| `open` | `ref` (default: event item) | Ask the host opener to open media/external content. |
| `play`, `pause` | `ref` for play | Media control through the host opener. |
| `select` | — | Reserved for editor selection; a no-op in play mode. |

Validation rejects unknown verbs and missing `to`/`handler`/`prop` fields.
Actions that run before `page.Bind` are buffered by the page and flushed when
the handlers are registered.

## Events

| Event | Emitted by |
|---|---|
| `tapped` | buttons, links, hotspots, images, shapes, icons, hyperlinks |
| `doubleTapped`, `secondaryTapped` | hotspots, images, shapes |
| `hoverIn`, `hoverOut` | hotspots, images, shapes (desktop pointer) |
| `changed` | entry, check, select, radio, slider (coalesced by the widget) |
| `submitted` | entry (Enter / Shift+Enter) |
| `itemTapped` | tie components (tie-browser) |
| `selected` | tabs/collections that expose a selection |

Every event payload contains **every** key, because Risor treats a missing map
key as an error:

```json
{"id": "door", "name": "Door", "type": "hotspot", "event": "tapped",
 "x": 0, "y": 0, "item": "", "value": ""}
```

Components fill `x`/`y` from the tap, `value` from inputs, and `item` from media
items. Scripts must not assume that unused keys are absent.

## Rendering and runtime

`scene/render` builds a `View` (a Fyne widget) from a document:

- one `Live` component per node; containers implement `Insert`, `Remove`,
  `Reorder`, `SetPlace`, `Preview`, `DropSpot` and `Transform`;
- `View.Apply([]scene.Change)` applies document changes incrementally
  (property updates, subtree rebuilds, insert/remove/move);
- `View.RectOf(id)` returns cached geometry for hit-testing and editor chrome;
- async assets (images, backgrounds) load off the UI thread and apply through
  `guithread.Do` with a generation guard;
- unknown component types render as a labeled placeholder and keep their data,
  so a newer document degrades instead of failing in an older renderer.

Modes:

- **play** — normal interaction; hotspots are invisible unless configured.
- **edit** — hotspots show outlines, slots show hints; input is supplied by the
  designer overlay, not by the widgets.

## Scripting

A sidecar script is an ordinary Risor page script. The browser/CLI injects a
`page` global for the run and, in the browser, a `scene` namespace scoped to the
page URL:

```risor
require(["v0.10", "@gui", "@scene"])

let musicOn = false

page.Bind({
    toggleMusic: (evt) => {
        musicOn = !musicOn
        if (musicOn) {
            page.Node("radio-glow").Show()
        } else {
            page.Node("radio-glow").Hide()
        }
    }
})

page.Node("wall-title").SetText(sprintf("Summer %d", 2014))
page.Find({type: "hotspot"}).each((n) => {
    n.On("hoverIn", (evt) => { browser.SetStatus(evt.name) })
})
```

### Script API

Types and attributes are listed in [API.md](API.md#scene). Summary:

| Object | Members |
|---|---|
| `scene` | `Load(ref)`, `Parse(json)`, `New()` |
| `scene.Page` | `Root`, `Title`, `Node(id)`, `Find(filter)`, `Bind(handlers)`, `Values()` |
| `scene.Node` | `ID`, `Type`, `Name`, `Props`, `Place`, `Children`, `Parent`, `Prop(name)`, `Set(name, value)`, `SetPlace(map)`, `SetName(name)`, `Show()`, `Hide()`, `On(event, fn)`, `Off(event)`, `SetContent(object)` |
| `scene.Doc` | `Title`, `Encode()`, `Validate()`, `Page()` |

`page.Node(id)` returns a **composite handle**: scene attributes win
(`SetPlace`, `On`, `Prop`), unknown attributes delegate to the familiar
fynerisor wrapper, so `page.Node("title").SetText("x")` works for a label and
`page.Node("img").FillMode` works for an image.

`page.Values()` reads the live input widgets, which makes forms easy:

```risor
page.Bind({
    run: (evt) => {
        let v = page.Values()
        print(sprintf("sample=%s threshold=%d", v.sample, v.threshold))
    }
})
```

Threading: UI events enter the Go action engine on the UI thread; script
listeners are queued on the page's VM goroutine (`vmexec.Callback`); node
mutations come back through `guithread.Do`. The VM is never entered from the UI
thread. The script is **stopped while the designer is open**; the editor renders
the document, not runtime state.

### Dynamic content

Use `slot` nodes and `node.SetContent(widget)` for content a script builds with
the full fynerisor widget API (tables, chart, list, custom controls). Use
`node.Set(prop, value)` for scene components. Runtime changes are not written
back to the document; the designer edits the saved document.

## Assets

Asset references (image/background, `scene.Load` arguments) accept:

- `tie:<hash>` — content-addressed tie blob (host resolver; tie-browser streams
  from the filehost and can serve thumbnails);
- `https://…` / `http://…` — fetched with a 30 s timeout and a size cap;
- `file:///…` — local file;
- a path relative to the document (resolved with `scene.ResolveRef`).

`render.LocalAssets` caches successful loads per run. The viewer never blocks the
UI thread on I/O. For `tie:` references, `scene/render/tie`
(`tieassets.Assets`) downloads through the tie filehost, preferring the item's
server-cached `thumbnail` relation and falling back to the original blob. Hosts
choose the resolver when they build the render options.

## Stores and versions

| Store | Use |
|---|---|
| `scene.FileStore` | local files; atomic temp-file + rename; conflict detection by content hash |
| `scene.HTTPStore` | GET, and WebDAV-style PUT when writable; `If-Match`/ETag conflicts |
| `scene.MemoryStore` | tests and previews |
| `scene/store/tie` (`tiestore.Store`) | a scene directory in a tie collection: stable DirUID, `scene-doc` pointer, `scene.json` versions |

`Store.Load` returns a `Snapshot{Ref, Version, Data, Writable}`.
`Store.Save(ref, data, baseVersion)` refuses to overwrite a newer version and
returns `ErrConflict`. `scene.ResolveRef` implements browser-style relative
resolution for file/http/tie URLs, including same-origin checks in the HTTP
store.

Document validation is available without Fyne:

```sh
fynerisor check page.scene.json      # validates props, slots, actions, sidecar
fynerisor scene page.scene.json      # render it
fynerisor-designer page.scene.json   # edit it
```

## tie-browser storage model (follow-up)

The plan for `tie-gui/cmd/tie-browser` keeps a scene a first-class tie subject:

```
tie:/scenes/house/living-room/        DirUID D (dir-type "scene")
  scene.json                          H1 (versions in _prev)
  page.risor                          H2 (optional sidecar)
  thumbnail.png                       H3 (rendered offscreen on save)

(D, scene-doc, H1)                    pointer read with one Get
(D, scene-script, H2)
(D, thumbnail, H3)
(D, name, "Living room")  (D, tag, family)  (D, added-at, …)
(M, member-of, D)                     every media item the scene presents
(D, links-to, D2)                     derived from navigate actions
("scenes", "home", D)                 start page
```

- The directory DirUID is the stable identity for tags, `member-of` and
  `links-to`; the document hash is the version.
- `member-of` is already reverse-indexed by tie, so "scenes with this photo" is
  free.
- Saves write a new `scene.json` (version history from tie's `_prev`), update
  index edges in one batch, and write the `scene-doc` pointer last.
- `tie:` URLs are paths (`tie:/scenes/house/living-room`); navigation resolves
  relative links like files, and media refs are `tie:<hash>`.
- The same design works for a decentralized backend: the document graph is
  content-addressed; a name→hash pointer (`scene-doc`) is the mutable part.
- `scene/store/tie` implements `scene.Store` for this layout (load pointer and
  blob, atomic new file, conflict check, pointer update last), best-effort
  `member-of`/`links-to` indexing and the `thumbnail.png` relation.
- `scene/render/tie` implements the asset resolver and the `tie.*` component
  family (`tie.image`, `tie.gallery`, `tie.collection`, `tie.tags`,
  `tie.related`, `tie.document`, `tie.audio`, `tie.video`).
- `tie-gui/cmd/tie-browser` is the app shell: home scene creation, F4 editor,
  source view, save/reload and system media opening.

## Security

- Data-only scenes execute no code.
- A host decides whether sidecar scripts run at all. A foreign/decentralized
  scene should load data-only, or run with the `Safe()` capability preset.
- Script `scene.Load` resolves relative references against the page and applies
  the same-origin policy (HTTP pages cannot read local files through it).
- `open` on external URLs is a host policy hook; the reference browser opens in
  the system browser, tie-browser can confirm or block.
- Document limits: 2 MiB, 5000 nodes, depth 32, unknown major versions refused.

## Status

Implemented in v0.11.1: document model/specs/registry, canonical codec with
limits, validation, invertible ops and undo/redo, file/http/memory/tie stores,
`FreeLayout` with fit/fill/stretch and anchors, the standard component set,
hotspot/action engine (including polygon hotspots), incremental rendering,
style overrides, host-resolved embeds, designer overlay/palette/outline/
inspector/problems with inline text, wrap/convert, autosave/drafts,
device/zoom previews, palette drag, story map and touch-sized handles, Risor
binding, browser page kinds and F4 editing, CLI `check`/`scene`, standalone
designer, tie store indexing/thumbnails, tie components and tie-browser.

Roadmap: multi-document story map, native pwplay/mpv playback and tag/thumbnail
pickers in tie-browser, Android/touch-first panel layout, thumbnail generation
for file-hosted scenes, and further component polish.
