# Convergence plan: connect the document layer to production usage

Status: active plan (2026-11-12). Rationale and doctrine: [DIRECTION.md](DIRECTION.md).
The problem this plan fixes is concrete: the scene stacks and the script stacks
do not yet serve each other's flagship workloads — `scene/builtin` lacks the
components production data-tool apps live on (`form`, `table`, `log`), and
browser hosts have not all opted into scene pages even though
`fynerisor/browser` already carries both page kinds (`KindRisor`/`KindScene`).
Neither stack changes shape here; they get connected.

Owner key: **F** = fynerisor, **T** = tie-gui (tie-browser), **H** = browser
hosts in the wider ecosystem.

## M1 — Close the component gap: `form`, `field`, `table`, `log` (F)

Goal: a typical data-tool page is expressible as a scene document + sidecar.
Everything stays inside the existing schema machinery — spec in
`scene/builtin/builtin.go`, factory bound in `scene/render/builtin/registry.go`
(`RegisterFactories`), inspector editors driven by `PropKind`.

### `form` + `field`

- `form` — ordered container; children render as Fyne `widget.Form` rows.
- `field` — new single-child container, `Kind: Ordered`, `MaxChildren: 1`,
  props `label` (string), `help` (string). Renders as `widget.FormItem`; the
  single child (entry/select/check/slider/…) becomes the row's widget.
- Decision taken: composition via child nodes, **not** a `fields` array prop.
  A structured-data prop would need a new `PropKind` for object lists and a
  bespoke inspector editor; child nodes reuse existing nesting, palette and
  drag behaviour. Revisit only if the designer shows the wrapper is painful.

### `table`

- Props (v1): `columns` (`PropOptions`, header row). Rows are script-fed;
  declarative static `data` is deferred until the designer needs previews —
  and when it lands it must not invent an object-list `PropKind` without a
  real editor use-case.
- Script feed: extend `gui/scene` node wrappers with component-specific
  methods routed through `View.LiveObject` (DIRECTION rule 4):
  `table.SetRows(rows)` / `table.SetCell(r, c, v)` (coalesced refresh via
  `vmexec`), `log.Append(text)`. Generated `docs/API.md` via `go generate .`.

### `log`

- Leaf; props: `text` (initial content), `autoScroll` (bool, default true).
  Factory wraps `widget.Log`; script API `Append` (see above).

### M1 work items

| # | Repo | Task |
|---|---|---|
| 1.1 | F | Specs for `form`/`field`/`table`/`log` in `scene/builtin/builtin.go` (categories "Input"/"Data") |
| 1.2 | F | Factories in `scene/render/builtin/` (new `data.go`), wire in `registry.go` |
| 1.3 | F | `gui/scene`: `Log.Append`, `Table.SetRows`/`SetCell` via `LiveObject`; `objattrs` method defs |
| 1.4 | F | Tests: spec validation (scene package), render smoke (fyne test app), node-method round-trip; `go test -race ./...` |
| 1.5 | F | Regenerate `docs/API.md`; update `docs/SCENES.md` component tables; CHANGELOG |
| 1.6 | F | Designer check: palette auto-lists new types (spec-driven); verify `field` single-child drag-in UX; fix gaps |
| 1.7 | T | tie-browser dogfood: rebuild against new registry, sample page with a table + log |

Acceptance: a scene document with form/table/log renders in tie-browser and in
`fynerisor-designer`; a sidecar script fills a table and appends to a log
without touching raw widgets.

## M2 — Browser hosts render scene pages (H, F)

Goal: any browser-style host can open a `.scene.json` page from wherever it
already fetches script pages — same transport, same auth — while its existing
script-page URLs keep working **forever, unchanged**.

Library support exists: `browser.Config.LoadPage`/`RenderScene`,
`browser.Page{URL, Kind, Body}` (`browser/page.go`); kind sniffing
(`isSceneURL`/`looksLikeScene`) is in `browser/editor.go`. tie-browser is the
reference implementation (`cmd/tie-browser/main.go`: `LoadPage: t.loadPage,
RenderScene: t.renderScene`).

| # | Repo | Task |
|---|---|---|
| 2.1 | F | Factor the reusable parts of tie-browser's scene wiring into a `browser` helper (view construction, navigation/status plumbing) if the duplication is non-trivial |
| 2.2 | F | Document the host recipe: `LoadPage` (fetch + kind sniff) and `RenderScene` (`render.NewView` with a host `AssetResolver`, `Navigator`, `Status`) in SCENES.md / DESIGNER.md |
| 2.3 | H | Each browser host switches `LoadContent` → `LoadPage` and implements `RenderScene`; script pages keep any script-specific validation they have |
| 2.4 | H | `AssetResolver`: resolve scene `asset` props relative to the page URL through the host's own transport/auth (hosts without tie use plain URLs; per-page cache lifetime is fine for v1) |
| 2.5 | H | Ship a demo scene app per host; regression-pass the existing script catalogue |

Acceptance: a host opens a scene page end-to-end; script pages behave exactly
as before; scene↔script navigation works both ways.

Effort note: M1 and the fynerisor side of M2 are each roughly a focused
evening in this codebase; the risk concentrates in host asset fetching (2.4)
and designer UX for `field` (1.6).

## M3 — Scene-first as the default for new pages (docs, hosts)

Policy, not new machinery:

- **New pages are scene documents** (`index.scene.json` + optional
  `page.risor`) as the default convention wherever fynerisor pages are
  authored or generated.
- **Generated pages (AI or scaffolding pipelines) emit scenes**, not raw
  widget scripts — the generated page is then *human-fixable in the
  designer*, the AI+designer loop DIRECTION.md promises.
- `docs/EXAMPLES.md` and `examples/`: label the scene-first path canonical
  for pages; script pages stay documented as "classic / escape hatch". No
  example deletions.
- **Migration of existing script apps: opportunistic only.** When someone
  revisits an app for a feature change, offer the scene rewrite. No bulk
  port; no date. A rewrite-recipe doc comes with the first real rewrite.
- Optional, decide at M2 exit: embed the designer inside browser hosts
  (save-back requires a host-side write endpoint; defer unless a real user
  wants it).

## M4 — Endgame only (do not start)

Scripts build scenes (DIRECTION.md long game). Entry point is growing
`gui/scene` (`scene.New/Parse/Doc/Page` already exist) toward a builder
surface, so every scripted page is serializable/designer-editable. Start only
after M1–M3 prove the document model against production data apps for a few
months. The raw widget API stays regardless, as embedding substrate.

## Standing guardrails

1. `fynerisor.scene/1` is the frozen contract: additive props/components only;
   unknown majors refuse with a clear error (already the rule).
2. Hosts' existing script-page distribution contracts never break; script
   pages are supported forever.
3. New JSON features are refused by default in favour of a sidecar script
   (DIRECTION complexity budget).
4. Every M-item lands with docs updated per the CLAUDE.md documentation map
   (API facts → generated API.md; format facts → SCENES.md; only these docs
   add plan/doctrine facts).
