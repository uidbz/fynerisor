// Package designer provides the in-place scene editor: a WYSIWYG canvas that
// shows the real Fyne widgets under a single input-shielding overlay, a
// schema-driven inspector, an outline, a palette, snapping, undo/redo and
// save-with-conflict-detection.
//
// The editor works on a scene.Document. Every mutation goes through
// scene.Tx/scene.Op, so undo/redo and the renderer's incremental Apply are
// exact. The page script is not running while the editor is open.
package designer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// Options configures an Editor.
type Options struct {
	Registry *render.Registry
	Assets   render.AssetResolver
	Store    scene.Store
	Ref      string // document reference for Store.Save/asset resolution
	Version  string // base version for optimistic concurrency
	Window   fyne.Window

	Navigator render.Navigator // preview-mode navigation
	Opener    render.Opener

	Status    func(string)
	OnClose   func(reload bool)
	OnDirty   func(bool)
	AssetPick func(current string) (string, bool) // optional asset browser
}

// Editor is a scene designer instance. Object returns its UI.
type Editor struct {
	doc   *scene.Document
	opts  Options
	view  *render.View
	hist  *scene.History
	issue []scene.Issue

	sel     map[string]bool
	primary string
	tool    string

	snapGrid float64
	preview  bool

	overlay   *overlay
	inspector *inspector
	outline   *outlineView
	palette   *paletteView
	problems  *problemsView

	stage   *fyne.Container
	content fyne.CanvasObject
	status  *widget.Label

	version  string
	saving   bool
	closing  bool
	onDirty  func(bool)
	lastDirt bool
}

// New builds an editor for doc.
func New(doc *scene.Document, opts Options) *Editor {
	if opts.Registry == nil {
		opts.Registry = renderbuiltin.NewRegistry()
	}
	if opts.Assets == nil {
		opts.Assets = render.NewLocalAssets()
	}
	e := &Editor{
		doc:      doc,
		opts:     opts,
		sel:      make(map[string]bool),
		snapGrid: 8,
		version:  opts.Version,
		onDirty:  opts.OnDirty,
	}
	e.issue = scene.Validate(doc, opts.Registry.Registry)
	e.view = render.NewView(doc, render.Options{
		Registry:  opts.Registry,
		Assets:    opts.Assets,
		Mode:      render.ModeEdit,
		Base:      opts.Ref,
		Navigator: opts.Navigator,
		Opener:    opts.Opener,
		Status:    e.setStatus,
	})
	e.hist = scene.NewHistory(200)
	e.overlay = newOverlay(e)
	e.buildUI()
	e.refreshUI()
	return e
}

// Object returns the editor UI as a canvas object.
func (e *Editor) Object() fyne.CanvasObject { return e.content }

// Doc returns the edited document.
func (e *Editor) Doc() *scene.Document { return e.doc }

// View returns the live renderer (preview, geometry).
func (e *Editor) View() *render.View { return e.view }

// Overlay returns the input-shielding overlay (tests, host integration).
func (e *Editor) Overlay() fyne.CanvasObject { return e.overlay }

// Dirty reports whether there are unsaved changes.
func (e *Editor) Dirty() bool { return e.hist.Dirty() }

// Ref returns the document reference.
func (e *Editor) Ref() string { return e.opts.Ref }

// Closed reports whether the editor has already handed control back to its
// host. Hosts use it to ignore stale providers after a navigation.
func (e *Editor) Closed() bool { return e.closing }

// Version returns the last saved/loaded version.
func (e *Editor) Version() string { return e.version }

// Issues returns the last validation result.
func (e *Editor) Issues() []scene.Issue { return e.issue }

// Selection returns the selected node IDs in document order.
func (e *Editor) Selection() []string { return e.selectedIDs() }

// ArmTool selects a palette component for the next canvas gesture.
func (e *Editor) ArmTool(typ string) {
	if _, ok := e.opts.Registry.Spec(typ); !ok {
		return
	}
	e.tool = typ
	e.setStatus(fmt.Sprintf("Place %s: click or drag on the canvas", typ))
	e.overlay.Refresh()
}

// DisarmTool clears the armed palette component.
func (e *Editor) DisarmTool() {
	e.tool = ""
	e.setStatus("")
	e.overlay.Refresh()
}

// SetGrid changes the snapping grid (0 disables grid snapping).
func (e *Editor) SetGrid(v float64) { e.snapGrid = v }

// Grid returns the snapping grid.
func (e *Editor) Grid() float64 { return e.snapGrid }

// Select replaces the selection with id ("" clears).
func (e *Editor) Select(id string) {
	if id == "" {
		e.clearSelection()
		return
	}
	if _, ok := e.doc.Find(id); !ok {
		return
	}
	e.sel = map[string]bool{id: true}
	e.primary = id
	e.refreshSelection()
}

// SelectMany replaces the selection.
func (e *Editor) SelectMany(ids []string) {
	e.sel = make(map[string]bool, len(ids))
	e.primary = ""
	for _, id := range ids {
		if _, ok := e.doc.Find(id); ok {
			e.sel[id] = true
			e.primary = id
		}
	}
	e.refreshSelection()
}

// ToggleSelect adds or removes id from the selection.
func (e *Editor) ToggleSelect(id string) {
	if _, ok := e.doc.Find(id); !ok {
		return
	}
	if e.sel[id] {
		delete(e.sel, id)
		if e.primary == id {
			e.primary = ""
			for x := range e.sel {
				e.primary = x
				break
			}
		}
	} else {
		e.sel[id] = true
		e.primary = id
	}
	e.refreshSelection()
}

// ClearSelection deselects everything.
func (e *Editor) ClearSelection() { e.clearSelection() }

func (e *Editor) clearSelection() {
	e.sel = make(map[string]bool)
	e.primary = ""
	e.refreshSelection()
}

// Primary returns the primary selected node, if any.
func (e *Editor) Primary() (*scene.Node, bool) {
	if e.primary == "" {
		return nil, false
	}
	return e.doc.Find(e.primary)
}

func (e *Editor) selectedIDs() []string {
	var out []string
	e.doc.Walk(func(n *scene.Node) bool {
		if e.sel[n.ID] {
			out = append(out, n.ID)
		}
		return true
	})
	return out
}

// beginEdit focuses the overlay so keyboard shortcuts work.
func (e *Editor) focusOverlay() {
	if c := e.canvas(); c != nil {
		c.Focus(e.overlay)
	}
}

func (e *Editor) canvas() fyne.Canvas {
	if e.opts.Window == nil {
		return nil
	}
	return e.opts.Window.Canvas()
}

// setStatus updates the editor status line and the host callback.
func (e *Editor) setStatus(msg string) {
	if e.status != nil {
		e.status.SetText(msg)
	}
	if e.opts.Status != nil && msg != "" {
		e.opts.Status(msg)
	}
}

func (e *Editor) buildUI() {
	e.status = widget.NewLabel("")
	e.status.Truncation = fyne.TextTruncateEllipsis

	e.palette = newPalette(e)
	e.outline = newOutline(e)
	e.inspector = newInspector(e)
	e.problems = newProblems(e)

	left := container.NewAppTabs(
		container.NewTabItemWithIcon("Palette", theme.GridIcon(), e.palette.object()),
		container.NewTabItemWithIcon("Outline", theme.ListIcon(), e.outline.object()),
	)
	right := container.NewAppTabs(
		container.NewTabItemWithIcon("Inspector", theme.SettingsIcon(), e.inspector.object()),
		container.NewTabItemWithIcon("Problems", theme.WarningIcon(), e.problems.object()),
	)

	e.stage = container.NewStack(e.view, e.overlay)
	done := widget.NewButton("Done", func() { e.RequestClose() })
	top := container.NewBorder(nil, nil, e.toolbar(), done, nil)
	center := container.NewBorder(top, e.status, nil, nil, e.stage)

	rightSplit := container.NewHSplit(center, right)
	rightSplit.SetOffset(0.72)
	main := container.NewHSplit(left, rightSplit)
	main.SetOffset(0.18)
	e.content = main
}

func (e *Editor) toolbar() fyne.CanvasObject {
	tb := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentUndoIcon(), e.undo),
		widget.NewToolbarAction(theme.ContentRedoIcon(), e.redo),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() { e.Save() }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCopyIcon(), e.duplicateSelection),
		widget.NewToolbarAction(theme.DeleteIcon(), e.deleteSelection),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.MoveUpIcon(), func() { e.reorderSelection(-1) }),
		widget.NewToolbarAction(theme.MoveDownIcon(), func() { e.reorderSelection(1) }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() { e.SetPreview(!e.preview) }),
		widget.NewToolbarAction(theme.GridIcon(), func() {
			if e.snapGrid > 0 {
				e.snapGrid = 0
				e.setStatus("Grid snapping off")
			} else {
				e.snapGrid = 8
				e.setStatus("Grid snapping 8")
			}
		}),
	)
	return tb
}

// Save encodes, validates and saves the document asynchronously.
func (e *Editor) Save() { e.save(nil) }

func (e *Editor) save(after func(bool)) {
	if e.saving {
		e.setStatus("Save already in progress")
		if after != nil {
			after(false)
		}
		return
	}
	if e.opts.Store == nil || e.opts.Ref == "" {
		e.setStatus("No store configured for this document")
		if after != nil {
			after(false)
		}
		return
	}
	e.issue = scene.Validate(e.doc, e.opts.Registry.Registry)
	e.problems.set(e.issue)
	if scene.Errors(e.issue) > 0 {
		e.setStatus(fmt.Sprintf("Cannot save: %d validation errors", scene.Errors(e.issue)))
		if after != nil {
			after(false)
		}
		return
	}
	data, err := scene.Encode(e.doc)
	if err != nil {
		e.setStatus("Encode: " + err.Error())
		if after != nil {
			after(false)
		}
		return
	}
	ref, version := e.opts.Ref, e.version
	e.saving = true
	e.setStatus("Saving…")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		snap, err := e.opts.Store.Save(ctx, ref, data, version)
		guithread.Do(func() {
			e.saving = false
			if err != nil {
				e.setStatus("Save failed: " + err.Error())
				if after != nil {
					after(false)
				}
				return
			}
			e.version = snap.Version
			e.hist.MarkSaved()
			e.refreshDirty()
			e.setStatus(fmt.Sprintf("Saved %s", ref))
			if after != nil {
				after(true)
			}
		})
	}()
}

// RequestClose closes the editor, asking about unsaved changes.
func (e *Editor) RequestClose() {
	if e.closing {
		return
	}
	if !e.Dirty() {
		e.close(false)
		return
	}
	w := e.opts.Window
	if w == nil {
		e.close(false)
		return
	}
	d := dialog.NewConfirm("Unsaved changes", "Save before closing?", func(save bool) {
		if save {
			e.save(func(ok bool) {
				if ok {
					e.close(false)
				}
			})
			return
		}
		e.close(false)
	}, w)
	d.SetConfirmText("Save")
	d.SetDismissText("Discard")
	d.Show()
}

// BeforeLeave asks for confirmation before a navigation discards unsaved work.
// done is called exactly once.
func (e *Editor) BeforeLeave(done func(ok bool)) {
	if !e.Dirty() {
		done(true)
		return
	}
	w := e.opts.Window
	if w == nil {
		done(true)
		return
	}
	d := dialog.NewConfirm("Unsaved changes", "Discard them and leave this page?", func(ok bool) {
		done(ok)
	}, w)
	d.SetConfirmText("Discard")
	d.SetDismissText("Stay")
	d.Show()
}

func (e *Editor) close(reload bool) {
	if e.closing {
		return
	}
	e.closing = true
	e.view.Destroy()
	if e.opts.OnClose != nil {
		e.opts.OnClose(reload)
	}
}

// SetPreview toggles between editing and playing the page.
func (e *Editor) SetPreview(on bool) {
	e.preview = on
	if on {
		e.view.SetMode(render.ModePlay)
		e.overlay.Hide()
		e.setStatus("Preview — actions run; press the preview button to return")
		return
	}
	e.view.SetMode(render.ModeEdit)
	e.overlay.Show()
	e.view.RefreshRects()
	e.overlay.Refresh()
	e.setStatus("Editing")
}

func (e *Editor) refreshUI() {
	e.pruneSelection()
	e.issue = scene.Validate(e.doc, e.opts.Registry.Registry)
	e.problems.set(e.issue)
	e.inspector.refresh()
	e.outline.refresh()
	e.overlay.Refresh()
	e.refreshDirty()
}

func (e *Editor) refreshSelection() {
	e.inspector.refresh()
	e.outline.syncSelection()
	e.overlay.Refresh()
	e.refreshStatus()
}

func (e *Editor) refreshStatus() {
	if len(e.sel) == 0 {
		e.setStatus("Select a node, or click a palette item to add one")
		return
	}
	if len(e.sel) == 1 {
		if n, ok := e.Primary(); ok {
			r, _ := e.view.RectOf(n.ID)
			e.setStatus(fmt.Sprintf("%s · %s · x %.0f y %.0f  %.0f×%.0f",
				n.Type, n.ID, r.X, r.Y, r.W, r.H))
			return
		}
	}
	e.setStatus(fmt.Sprintf("%d nodes selected", len(e.sel)))
}

func (e *Editor) refreshDirty() {
	d := e.hist.Dirty()
	if e.lastDirt != d {
		e.lastDirt = d
		if e.onDirty != nil {
			e.onDirty(d)
		}
	}
}

func (e *Editor) pruneSelection() {
	for id := range e.sel {
		if _, ok := e.doc.Find(id); !ok {
			delete(e.sel, id)
		}
	}
	if _, ok := e.doc.Find(e.primary); !ok {
		e.primary = ""
		for id := range e.sel {
			e.primary = id
			break
		}
	}
}

// apply runs one undoable transaction and applies its changes to the view.
func (e *Editor) apply(label, merge string, ops ...scene.Op) bool {
	if len(ops) == 0 {
		return false
	}
	tx := scene.Tx{Label: label, MergeKey: merge, Ops: ops}
	changes, err := e.doc.Apply(tx)
	if err != nil {
		e.setStatus(err.Error())
		return false
	}
	if err := e.view.Apply(changes); err != nil {
		e.setStatus("render: " + err.Error())
	}
	e.hist.Push(tx)
	e.view.RefreshRects()
	e.refreshUI()
	if e.primary != "" {
		e.outline.syncSelection()
	}
	e.refreshStatus()
	return true
}

// applyEdit applies a transaction without rebuilding the inspector, so typing
// in a field keeps focus. Structural changes use apply.
func (e *Editor) applyEdit(label, merge string, ops ...scene.Op) bool {
	if len(ops) == 0 {
		return false
	}
	tx := scene.Tx{Label: label, MergeKey: merge, Ops: ops}
	changes, err := e.doc.Apply(tx)
	if err != nil {
		e.setStatus(err.Error())
		return false
	}
	if err := e.view.Apply(changes); err != nil {
		e.setStatus("render: " + err.Error())
	}
	e.hist.Push(tx)
	e.view.RefreshRects()
	e.issue = scene.Validate(e.doc, e.opts.Registry.Registry)
	e.problems.set(e.issue)
	e.overlay.Refresh()
	e.refreshDirty()
	e.refreshStatus()
	return true
}

func (e *Editor) undo() {
	changes, ok, err := e.hist.Undo(e.doc)
	if err != nil {
		e.setStatus("Undo: " + err.Error())
		return
	}
	if !ok {
		return
	}
	if err := e.view.Apply(changes); err != nil {
		e.setStatus("render: " + err.Error())
	}
	e.view.RefreshRects()
	e.refreshUI()
	e.setStatus("Undo: " + e.hist.RedoLabel())
}

func (e *Editor) redo() {
	changes, ok, err := e.hist.Redo(e.doc)
	if err != nil {
		e.setStatus("Redo: " + err.Error())
		return
	}
	if !ok {
		return
	}
	if err := e.view.Apply(changes); err != nil {
		e.setStatus("render: " + err.Error())
	}
	e.view.RefreshRects()
	e.refreshUI()
	e.setStatus("Redo")
}

// ------------------------------------------------------------------ editing

func (e *Editor) deleteSelection() {
	ids := e.selectedIDs()
	if len(ids) == 0 {
		return
	}
	sort.SliceStable(ids, func(i, j int) bool {
		return e.depth(ids[i]) > e.depth(ids[j])
	})
	var ops []scene.Op
	for _, id := range ids {
		if e.hasSelectedAncestor(id) {
			continue
		}
		op, err := scene.RemoveOp(e.doc, id)
		if err != nil {
			e.setStatus(err.Error())
			continue
		}
		ops = append(ops, op)
	}
	if e.apply("delete", "", ops...) {
		e.clearSelection()
	}
}

func (e *Editor) hasSelectedAncestor(id string) bool {
	for {
		p := e.view.ParentID(id)
		if p == "" {
			return false
		}
		if e.sel[p] {
			return true
		}
		id = p
	}
}

func (e *Editor) depth(id string) int {
	d := 0
	for {
		p := e.view.ParentID(id)
		if p == "" {
			return d
		}
		d++
		id = p
	}
}

func (e *Editor) duplicateSelection() {
	ids := e.selectedIDs()
	if len(ids) == 0 {
		return
	}
	taken := e.takenIDs()
	var ops []scene.Op
	var newIDs []string
	for _, id := range ids {
		n, ok := e.doc.Find(id)
		if !ok {
			continue
		}
		clone := n.Clone()
		clone.ReID(taken)
		if parent := e.view.ParentID(id); parent != "" {
			if e.view.ContainerKind(parent) == scene.Free {
				clone.Place.X += 10
				clone.Place.Y += 10
			}
			ops = append(ops, scene.InsertOp(parent, e.endIndex(parent), clone))
		} else {
			// Duplicating the root is not supported; duplicate its children.
			for _, c := range clone.Children {
				ops = append(ops, scene.InsertOp(id, e.endIndex(id), c))
				newIDs = append(newIDs, c.ID)
			}
			continue
		}
		newIDs = append(newIDs, clone.ID)
	}
	if e.apply("duplicate", "", ops...) {
		e.SelectMany(newIDs)
	}
}

// takenIDs is the set of node IDs used for ReID collision checks.
func (e *Editor) takenIDs() map[string]bool {
	out := make(map[string]bool)
	e.doc.Walk(func(n *scene.Node) bool { out[n.ID] = true; return true })
	return out
}

func (e *Editor) endIndex(parent string) int {
	n, ok := e.doc.Find(parent)
	if !ok {
		return 0
	}
	return len(n.Children)
}

func (e *Editor) reorderSelection(delta int) {
	if e.primary == "" {
		return
	}
	parent, index, ok := e.doc.ParentOf(e.primary)
	if !ok || parent == nil {
		return
	}
	target := index + delta
	if target < 0 || target >= len(parent.Children) {
		return
	}
	op, err := scene.MoveOp(e.doc, e.primary, parent.ID, target)
	if err != nil {
		e.setStatus(err.Error())
		return
	}
	e.apply("reorder", "", op)
}

// addNode inserts a component and returns its ID.
func (e *Editor) addNode(typ, parent string, index int, p scene.Place) string {
	spec, ok := e.opts.Registry.Spec(typ)
	if !ok {
		e.setStatus("Unknown component " + typ)
		return ""
	}
	pn, ok := e.doc.Find(parent)
	if !ok {
		return ""
	}
	pspec, _ := e.opts.Registry.Spec(pn.Type)
	if pspec.Kind != scene.Free {
		p = scene.Place{MinW: spec.DefaultPlace.W, MinH: spec.DefaultPlace.H}
		if pspec.Kind == scene.Slots {
			slot := ""
			used := make(map[string]bool)
			for _, c := range pn.Children {
				used[c.Place.Slot] = true
			}
			for _, s := range pspec.Slots {
				if !used[s] {
					slot = s
					break
				}
			}
			if slot == "" {
				slot = pspec.Slots[0]
			}
			p.Slot = slot
		}
		if pspec.Kind == scene.Tabs {
			p.Title = spec.Title
		}
		if pspec.Kind == scene.Tabs && p.Title == "" {
			p.Title = spec.Title
		}
	} else if p.W == 0 || p.H == 0 {
		p.W = float64(spec.DefaultPlace.W)
		p.H = float64(spec.DefaultPlace.H)
	}
	if p.W == 0 {
		p.W = float64(spec.DefaultPlace.W)
	}
	if p.H == 0 {
		p.H = float64(spec.DefaultPlace.H)
	}
	n := &scene.Node{ID: e.uniqueID(scene.Slug(spec.Title)), Type: typ, Place: p}
	for _, ps := range spec.Props {
		if ps.Default == nil {
			continue
		}
		if s, ok := ps.Default.(string); ok && s == "" {
			continue
		}
		n.SetProp(ps.Name, ps.Default)
	}
	if e.apply("add "+typ, "", scene.InsertOp(parent, index, n)) {
		e.Select(n.ID)
		return n.ID
	}
	return ""
}

func (e *Editor) uniqueID(base string) string {
	taken := e.doc.IDs()
	id := base
	for i := 2; ; i++ {
		if taken[id] == nil {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
}

// hitTest returns the selectable node at a point in view coordinates.
func (e *Editor) hitTest(pos fyne.Position) string { return e.view.NodeAt(pos) }

// parentFree reports whether a node's parent uses free placement.
func (e *Editor) parentFree(id string) bool {
	p := e.view.ParentID(id)
	if p == "" {
		return false
	}
	return e.view.ContainerKind(p) == scene.Free
}

// freeTransform returns the design→view transform of a free parent.
func (e *Editor) freeTransform(parent string) (render.Transform, render.Rect, bool) {
	r, ok := e.view.RectOf(parent)
	if !ok {
		return render.Transform{SX: 1, SY: 1}, r, false
	}
	n, ok := e.doc.Find(parent)
	if !ok || e.view.ContainerKind(parent) != scene.Free {
		return render.Transform{SX: 1, SY: 1, OX: float64(r.X), OY: float64(r.Y)}, r, false
	}
	design := fyne.NewSize(float32(propNum(e.opts.Registry.Registry, n, "designW", 1600)), float32(propNum(e.opts.Registry.Registry, n, "designH", 900)))
	mode := propStr(e.opts.Registry.Registry, n, "scale", "fit")
	fl := render.NewFreeLayout(design, mode)
	t := fl.Transform(r.Size())
	t.OX += float64(r.X)
	t.OY += float64(r.Y)
	return t, r, true
}

func propStr(reg *scene.Registry, n *scene.Node, key, def string) string {
	if v, ok := n.Prop(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if spec, ok := reg.Spec(n.Type); ok {
		if p, ok := spec.Prop(key); ok {
			if s, ok := p.Default.(string); ok {
				return s
			}
		}
	}
	return def
}

func propNum(reg *scene.Registry, n *scene.Node, key string, def float64) float64 {
	if v, ok := n.Prop(key); ok {
		switch x := v.(type) {
		case float64:
			return x
		case int:
			return float64(x)
		}
	}
	if spec, ok := reg.Spec(n.Type); ok {
		if p, ok := spec.Prop(key); ok {
			if v, ok := p.Default.(float64); ok {
				return v
			}
		}
	}
	return def
}

// -------------------------------------------------------------------- snaps

// snapPlace applies grid and sibling-edge snapping to a free placement.
func (e *Editor) snapPlace(id string, p scene.Place) (scene.Place, []guide) {
	parent := e.view.ParentID(id)
	if parent == "" {
		return p, nil
	}
	var guides []guide
	if e.snapGrid > 0 {
		p.X = roundTo(p.X, e.snapGrid)
		p.Y = roundTo(p.Y, e.snapGrid)
	}
	siblings := e.siblingPlaces(parent, id)
	threshold := 6.0
	if t, _, ok := e.freeTransform(parent); ok {
		s := t.SX
		if t.SY < s {
			s = t.SY
		}
		if s > 0 {
			threshold = 6.0 / s
		}
	}
	// Vertical edges: left, center, right against sibling edges and centers.
	bestDX, bestVX, foundX := 0.0, 0.0, false
	for _, s := range siblings {
		for _, target := range []float64{s.X, s.X + s.W/2, s.X + s.W} {
			for _, edge := range []float64{p.X, p.X + p.W/2, p.X + p.W} {
				d := target - edge
				if abs(d) <= threshold && (!foundX || abs(d) < abs(bestDX)) {
					bestDX, bestVX, foundX = d, target, true
				}
			}
		}
	}
	if foundX {
		p.X += bestDX
		guides = append(guides, guide{vertical: true, at: bestVX})
	}
	bestDY, bestVY, foundY := 0.0, 0.0, false
	for _, s := range siblings {
		for _, target := range []float64{s.Y, s.Y + s.H/2, s.Y + s.H} {
			for _, edge := range []float64{p.Y, p.Y + p.H/2, p.Y + p.H} {
				d := target - edge
				if abs(d) <= threshold && (!foundY || abs(d) < abs(bestDY)) {
					bestDY, bestVY, foundY = d, target, true
				}
			}
		}
	}
	if foundY {
		p.Y += bestDY
		guides = append(guides, guide{vertical: false, at: bestVY})
	}
	return p, guides
}

func (e *Editor) siblingPlaces(parent, id string) []scene.Place {
	pn, ok := e.doc.Find(parent)
	if !ok {
		return nil
	}
	out := make([]scene.Place, 0, len(pn.Children))
	for _, c := range pn.Children {
		if c.ID != id {
			out = append(out, c.Place)
		}
	}
	return out
}

func roundTo(v, grid float64) float64 {
	if grid <= 0 {
		return v
	}
	return scene.Round(math.Round(v/grid) * grid)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// ------------------------------------------------------------------ helpers

// validNodeID converts text into an unused node ID.
func (e *Editor) validNodeID(id, except string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	id = scene.Slug(id)
	if id == "" {
		return ""
	}
	if n, ok := e.doc.Find(id); ok && n.ID != except {
		return ""
	}
	return id
}
