# Direction: documents and scripts, two layers of one model

Status: doctrine (settled 2026-11-12). This document records *why* fynerisor
has both a scene document layer and the classic Risor widget API, why neither
one supersedes the other, and what the complexity budget is. It exists because
the project felt like it was drifting between "two ways to make a page" — the
resolution is that there is **one way with two layers**. The follow-up work is
in [CONVERGENCE.md](CONVERGENCE.md).

## The doctrine

> **A page is a scene document. Behaviour is a sidecar Risor script. The
> widget API is the escape hatch and the embedding substrate — never a
> second page kind you must choose between.**

- The **scene document** (`fynerisor.scene/1`, see [SCENES.md](SCENES.md)) is
  the *artifact*: what the designer edits, what tie version-stores
  content-addressed, what diffs and links are computed over, and what can be
  rendered with zero code execution.
- The **sidecar script** is the *behaviour*: event handlers, live data,
  anything clever. It runs on the same `vmexec`/`guithread` machinery as any
  other fynerisor run and reaches the page through the `scene` Risor bindings
  (`gui/scene`).
- The **raw widget API** (`gui/widget`, `gui/container`, …) is the substrate:
  hosts embed it directly (app chrome, embedding runtimes, scripted tooling),
  sidecars use it for the parts of a page the document cannot express, and
  scripted pages (`KindRisor`) are built entirely on it.

## Why a document format and not "the editor edits .risor"

The alternative design — a WYSIWYG editor that reads and writes Risor source —
was seriously considered and is rejected on purpose:

1. **Code is an expression; a document is data.** Dragging a node to a new
   position means writing that change back to its source. If the source is
   code, the editor must parse it, find which expression *produced* the widget
   (it may come from a loop, a helper function, or a computed layout), rewrite
   it without breaking formatting, comments, or surrounding logic, and keep
   the script's meaning identical to the canvas. That is program synthesis and
   unsolvable in general.
2. **Every serious visual GUI editor converged on a document format** —
   storyboards/XIBs, Android XML, Qt `.ui`, HTML/CSS. Apple's bet that "the
   code is the canvas" (SwiftUI + Xcode canvas) has been fighting this
   round-trip problem for years.
3. **The complexity was never avoidable.** "What does editing this page
   mean" requires a defined model of valid placement, identity and
   properties. The document format pays that cost *once, explicitly*. A
   code-editing editor pays it *forever, implicitly*, in the editor.

## Why JSON is not a human-authoring problem

`scene.json` is not written by hand by humans; that was never the authoring
surface:

- **Humans** edit through the designer; they read the document as *diffs*.
  Canonical encoding (`scene.Encode`: sorted keys, fixed field order, numbers
  rounded to 0.01) makes those diffs stable and meaningful, and makes a
  document's hash its version.
- **AI** writes and repairs JSON natively — and the schema is a *closed
  feedback loop*: validate → error → fix, which free-form code does not
  offer. The claim that "AI can generate code anyway, so JSON is redundant"
  is backwards: JSON's wins are editability at both ends, validation before
  execution, diff/version/linkability — not generation ease.
- **Humans who want to touch behaviour** use the sidecar `page.risor`, which
  is already the human/script surface.

## Why the widget API stays

It is tempting to argue that the document layer "supersedes" the bindings and
the classic API can be retired. It cannot, for two structural reasons.

First, the workloads differ. The document layer's showcase
([tie-browser](https://github.com/uidbz/tie-gui)) is the point-and-click /
media page. But the bulk of production Risor GUIs in the wider fynerisor
ecosystem are interactive *data tools* — forms, tables, logs, progress
readouts, live entries — built as scripts: those constructors (Button,
FormItem/Form, Entry, Label, Check, Select, Markdown, Log, Table, progress
bars) dominate real usage by a wide margin. The scene component set (~50
types) covered the point-and-click page well but lacked `form`, `table` and
`log`. That was a **component gap**, not a model failure; it was closed by
[CONVERGENCE.md](CONVERGENCE.md) M1 (`form`, `field`, `table`, `log`).

Second, removing the bindings would save little and break much:

- The expensive, hard-won parts of fynerisor — `vmexec`, `guithread`,
  `vmguard`, `objattrs`, the threading model — are **shared substrate**.
  Scene sidecar scripts run on exactly the same machinery. Deprecating the
  bindings does not remove that maintenance; it only removes the thin
  wrapper layer (39 widgets, mostly mechanical).
- It would break every existing script app, every embedding host (e.g.
  [anneal](https://github.com/uidbz/anneal)'s scripted UIs), and every
  host-side distribution contract built on script pages — while gaining
  nothing. Script pages (`KindRisor`) are a supported page kind forever.

## The complexity budget (how to keep the format small)

Drift risk is contained by rules, not by re-litigating the model:

1. **The sidecar script is the pressure valve.** Any "I need the page to do
   something clever" goes into Risor, not into new JSON features.
2. **No new `ContainerKind`s.** The four placement semantics (free,
   ordered, slots, tabs) are frozen; new components reuse them.
3. **Components are additive.** New component types are additive spec+factory
   pairs; unknown types/props in a document are warnings or clear errors, so
   old renderers degrade legibly.
4. **Script-visible node API grows narrowly.** Component-specific node methods
   (e.g. `Log.Append`, `Table.SetRows`) are added case by case, routed through
   narrow component capabilities (`View.Live` + `render.RowModelOf` /
   `render.AppenderOf`), not as a general widget-API passthrough — that would
   reintroduce the two page kinds inside the script layer.

## The long game (not committed)

Scripts *building* scenes rather than raw widgets (the `gui/scene`
`New`/`Parse`/`Page` bindings already point this way) would make every
scripted page serializable, designer-editable and diffable — full convergence
with one runtime. This is the endgame, deliberately postponed: M1–M3 in
CONVERGENCE.md must prove the document model against production data-tool
workloads first.
