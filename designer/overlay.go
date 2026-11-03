package designer

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

type gestureKind int

const (
	gIdle gestureKind = iota
	gPressed
	gMoving
	gResizing
	gMarquee
	gCreating
)

type guide struct {
	vertical bool
	at       float64 // design units in guideParent
}

// overlay is the single widget that receives every edit-mode gesture. It sits
// above the live View in a Stack (Fyne hit-tests the last child first), so the
// real widgets stay inert without any per-node shielding.
type overlay struct {
	widget.BaseWidget
	ed *Editor

	kind   gestureKind
	press  fyne.Position
	mouse  fyne.Position
	acc    fyne.Delta
	handle string
	hit    string

	start        map[string]scene.Place
	startRects   map[string]render.Rect
	lastPlaces   map[string]scene.Place
	previewRects map[string]render.Rect
	guideParent  string
	guides       []guide

	// lastTapAt/lastTapPos drive our own double-tap detection. The overlay
	// deliberately does not implement fyne.DoubleTappable: the driver then
	// delays every single Tapped by the 300 ms double-tap window, which made
	// click-to-place and selection feel laggy.
	lastTapAt  time.Time
	lastTapPos fyne.Position

	marquee    render.Rect
	marqueeIDs []string
	dropRect   render.Rect
	dropParent string
	createRect render.Rect

	hover       string
	suppressTap bool
	focused     bool
	objs        []fyne.CanvasObject
	chrome      []*canvas.Rectangle

	inline     *editorEntry
	inlineID   string
	inlineRect render.Rect

	lastPoints any
}

func newOverlay(e *Editor) *overlay {
	o := &overlay{
		ed: e, start: map[string]scene.Place{}, startRects: map[string]render.Rect{},
		lastPlaces: map[string]scene.Place{}, previewRects: map[string]render.Rect{},
	}
	o.ExtendBaseWidget(o)
	return o
}

func (o *overlay) CreateRenderer() fyne.WidgetRenderer {
	o.refreshChrome()
	return &overlayRenderer{o: o}
}

// Palette drag support: the palette item is outside the overlay, so it maps
// absolute pointer positions into overlay-local ones and drives the same
// creation gesture the canvas uses.

func (o *overlay) beginPalette(pos fyne.Position) {
	o.kind = gCreating
	o.press, o.mouse = pos, pos
	o.createRect = o.paletteRect(pos)
	o.suppressTap = true
	o.Refresh()
}

func (o *overlay) movePalette(pos fyne.Position) {
	o.mouse = pos
	o.createRect = o.paletteRect(pos)
	o.Refresh()
}

func (o *overlay) dropPalette(pos fyne.Position) {
	o.mouse = pos
	o.createRect = o.paletteRect(pos)
	// Drop the component at its default size under the pointer, not as a
	// rubber-band rectangle from wherever the drag entered the canvas.
	parent := o.ed.view.ContainerAt(pos)
	if parent == "" && o.ed.doc.Root != nil {
		parent = o.ed.doc.Root.ID
	}
	if parent != "" {
		o.createAt(parent, o.ed.defaultPlaceAt(pos, parent), pos)
	}
	o.kind = gIdle
	o.createRect = render.Rect{}
	o.suppressTap = false
	o.ed.view.RefreshRects()
	o.Refresh()
}

// paletteRect is the ghost shown while dragging a palette item: the component's
// default size, centred under the pointer.
func (o *overlay) paletteRect(pos fyne.Position) render.Rect {
	parent := o.ed.view.ContainerAt(pos)
	if parent == "" && o.ed.doc.Root != nil {
		parent = o.ed.doc.Root.ID
	}
	if parent == "" || o.ed.view.ContainerKind(parent) != scene.Free {
		return render.NewRect(pos.X, pos.Y, 8, 8)
	}
	p := o.ed.defaultPlaceAt(pos, parent)
	t, _, ok := o.ed.freeTransform(parent)
	if !ok {
		return render.NewRect(pos.X, pos.Y, 8, 8)
	}
	local := t.ToLocal(p.X, p.Y)
	return render.NewRect(local.X, local.Y, float32(p.W*t.SX), float32(p.H*t.SY))
}

// ---------------------------------------------------------------- gestures

func (o *overlay) begin(pos fyne.Position) {
	o.press, o.mouse = pos, pos
	o.acc = fyne.Delta{}
	o.start = map[string]scene.Place{}
	o.startRects = map[string]render.Rect{}
	o.lastPlaces = map[string]scene.Place{}
	o.previewRects = map[string]render.Rect{}
	o.guides = nil
	o.kind = gPressed
	o.suppressTap = false
	o.ed.focusOverlay()

	if h := o.handleAt(pos); h != "" {
		o.handle = h
		o.hit = o.ed.primary
		o.captureStarts()
		return
	}
	o.handle = ""
	id := o.ed.hitTest(pos)
	if o.ed.tool != "" {
		o.hit = id
		return
	}
	if id == "" {
		o.hit = ""
		o.ed.ClearSelection()
		o.suppressTap = true
		return
	}
	mods := o.mods()
	if mods&(fyne.KeyModifierShift|fyne.KeyModifierControl|fyne.KeyModifierSuper) != 0 {
		o.ed.ToggleSelect(id)
		o.suppressTap = true
		o.hit = ""
		return
	}
	if !o.ed.sel[id] {
		o.ed.Select(id)
		o.suppressTap = true
	}
	o.hit = id
	o.captureStarts()
}

func (o *overlay) captureStarts() {
	for _, id := range o.ed.selectedIDs() {
		if n, ok := o.ed.doc.Find(id); ok {
			o.start[id] = n.Place
		}
		if r, ok := o.ed.view.RectOf(id); ok {
			o.startRects[id] = r
		}
	}
}

func (o *overlay) MouseDown(e *desktop.MouseEvent) {
	o.ed.beginGesture()
	o.begin(e.Position)
	if e.Button == desktop.MouseButtonSecondary {
		o.showContext(e.Position)
		o.ed.endGesture()
		return
	}
	o.Refresh()
}

func (o *overlay) MouseUp(*desktop.MouseEvent) {
	if o.kind == gPressed {
		o.kind = gIdle
		o.hit = ""
		o.suppressTap = false
		o.Refresh()
	}
	o.ed.endGesture()
}

func (o *overlay) Tapped(e *fyne.PointEvent) {
	if o.inline != nil {
		if !o.inlineRect.Contains(e.Position) {
			o.commitInline()
		}
		return
	}
	o.ed.focusOverlay()
	if o.suppressTap {
		o.suppressTap = false
		return
	}
	// Our own double-tap detection: the overlay must not implement
	// fyne.DoubleTappable, or every single tap waits out the driver's
	// double-tap window.
	now := time.Now()
	double := now.Sub(o.lastTapAt) < 400*time.Millisecond &&
		abs(float64(e.Position.X-o.lastTapPos.X)) < 6 &&
		abs(float64(e.Position.Y-o.lastTapPos.Y)) < 6
	o.lastTapAt, o.lastTapPos = now, e.Position
	if o.ed.tool != "" {
		o.createClick(e.Position)
		return
	}
	id := o.ed.hitTest(e.Position)
	if id == "" {
		o.ed.ClearSelection()
		return
	}
	if double {
		o.doubleTap(id)
		return
	}
	if o.mods()&(fyne.KeyModifierShift|fyne.KeyModifierControl|fyne.KeyModifierSuper) != 0 {
		o.ed.ToggleSelect(id)
		return
	}
	o.ed.Select(id)
}

// doubleTap opens inline text editing (or focuses the first inspector field).
func (o *overlay) doubleTap(id string) {
	o.ed.Select(id)
	if n, ok := o.ed.doc.Find(id); ok {
		if spec, ok := o.ed.opts.Registry.Spec(n.Type); ok {
			if p, ok := spec.Prop("text"); ok && p.Kind != scene.PropPoints {
				if text, ok := n.Prop("text"); ok {
					o.startInline(id, fmt.Sprint(text))
					return
				}
			}
		}
	}
	if o.ed.inspector != nil {
		o.ed.inspector.focusFirst()
	}
}

// startInline places a real entry over a text node.
func (o *overlay) startInline(id, text string) {
	if o.inline != nil {
		o.commitInline()
	}
	r, ok := o.ed.view.RectOf(id)
	if !ok {
		return
	}
	entry := newEditorEntry(o.ed)
	entry.SetText(text)
	entry.OnSubmitted = func(s string) { o.commitInline() }
	entry.OnChanged = func(string) {}
	entry.SetPlaceHolder("Text")
	entry.onCommit = func() { o.commitInline() }
	entry.onCancel = func() { o.cancelInline() }
	o.inline = entry
	o.inlineID = id
	o.inlineRect = r.Inset(-2)
	o.Refresh()
	if c := o.ed.canvas(); c != nil {
		c.Focus(entry)
	}
}

func (o *overlay) commitInline() {
	if o.inline == nil {
		return
	}
	id, text := o.inlineID, o.inline.Text
	o.inline, o.inlineID = nil, ""
	o.refreshChrome()
	n, ok := o.ed.doc.Find(id)
	if !ok {
		return
	}
	old, _ := n.Prop("text")
	if fmt.Sprint(old) == text {
		return
	}
	o.ed.applyEdit("edit text", "text:"+id, scene.SetPropOp(id, "text", old, text))
	o.ed.inspector.refresh()
}

func (o *overlay) cancelInline() {
	o.inline, o.inlineID = nil, ""
	o.refreshChrome()
}

func (o *overlay) TappedSecondary(e *fyne.PointEvent) {
	o.showContext(e.Position)
}

func (o *overlay) showContext(pos fyne.Position) {
	id := o.ed.hitTest(pos)
	if id != "" && !o.ed.sel[id] {
		o.ed.Select(id)
	}
	c := o.ed.canvas()
	if c == nil {
		return
	}
	items := []*fyne.MenuItem{}
	if o.polygonSelected() {
		if h := o.handleAt(pos); strings.HasPrefix(h, "pt") {
			var idx int
			fmt.Sscanf(h, "pt%d", &idx)
			items = append(items, fyne.NewMenuItem("Remove point", func() { o.ed.removePolygonPoint(o.ed.primary, idx) }))
		}
		items = append(items, fyne.NewMenuItem("Add point here", func() { o.ed.addPolygonPoint(o.ed.primary, pos) }))
		items = append(items, fyne.NewMenuItemSeparator())
	}
	items = append(items,
		fyne.NewMenuItem("Duplicate", o.ed.duplicateSelection),
		fyne.NewMenuItem("Delete", o.ed.deleteSelection),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Bring to front", o.ed.bringToFront),
		fyne.NewMenuItem("Send to back", o.ed.sendToBack),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Wrap in vbox", func() { o.ed.wrapSelection("vbox") }),
		fyne.NewMenuItem("Wrap in hbox", func() { o.ed.wrapSelection("hbox") }),
		fyne.NewMenuItem("Wrap in free", func() { o.ed.wrapSelection("free") }),
		fyne.NewMenuItem("Unwrap", o.ed.unwrapSelection),
		fyne.NewMenuItem("Convert to free", func() { o.ed.convertSelection("free") }),
		fyne.NewMenuItem("Convert to vbox", func() { o.ed.convertSelection("vbox") }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Select parent", func() { o.ed.selectParent() }),
	)
	if o.ed.tool != "" {
		items = append(items, fyne.NewMenuItem("Cancel placement", o.ed.DisarmTool))
	}
	abs := fyne.CurrentApp().Driver().AbsolutePositionForObject(o)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", items...), c, fyne.NewPos(abs.X+pos.X, abs.Y+pos.Y))
}

func (o *overlay) Dragged(e *fyne.DragEvent) {
	if o.kind == gIdle || o.kind == gPressed {
		if o.kind == gIdle {
			o.begin(e.Position)
		}
		switch {
		case o.handle != "":
			o.kind = gResizing
			o.captureStarts()
		case o.ed.tool != "":
			o.kind = gCreating
		case o.hit != "":
			o.kind = gMoving
		default:
			o.kind = gMarquee
		}
		o.suppressTap = false
	}
	o.acc.DX += e.Dragged.DX
	o.acc.DY += e.Dragged.DY
	o.mouse = fyne.NewPos(o.press.X+o.acc.DX, o.press.Y+o.acc.DY)
	switch o.kind {
	case gMoving:
		o.dragMove()
	case gResizing:
		if strings.HasPrefix(o.handle, "pt") {
			o.dragPoint()
		} else {
			o.dragResize()
		}
	case gMarquee:
		o.marquee = rectBetween(o.press, o.mouse)
		o.marqueeIDs = o.nodesIn(o.marquee)
	case gCreating:
		o.createRect = rectBetween(o.press, o.mouse)
	}
	o.Refresh()
}

func (o *overlay) DragEnd() {
	switch o.kind {
	case gMoving:
		o.commitMove()
	case gResizing:
		o.commitResize()
	case gMarquee:
		o.ed.SelectMany(o.marqueeIDs)
	case gCreating:
		o.commitCreate()
	}
	for id := range o.lastPlaces {
		o.ed.view.ClearPreview(id)
	}
	o.kind = gIdle
	o.hit = ""
	o.handle = ""
	o.marquee = render.Rect{}
	o.createRect = render.Rect{}
	o.dropRect = render.Rect{}
	o.guides = nil
	o.lastPoints = nil
	o.previewRects = map[string]render.Rect{}
	o.ed.view.RefreshRects()
	o.Refresh()
	o.ed.endGesture()
}

func (o *overlay) dragMove() {
	if len(o.start) == 0 {
		return
	}
	primary := o.ed.primary
	start, ok := o.start[primary]
	if !ok || !o.ed.parentFree(primary) {
		// Ordered parents: show the insertion point; commitMove handles it.
		parent, spot, ok := o.ed.view.DropSpotAt(o.mouse)
		if ok {
			o.dropParent = parent
			o.dropRect = spot.Rect
		}
		return
	}
	parent := o.ed.view.ParentID(primary)
	t, _, _ := o.ed.freeTransform(parent)
	dx := float64(o.acc.DX)
	dy := float64(o.acc.DY)
	if t.SX != 0 {
		dx /= t.SX
	}
	if t.SY != 0 {
		dy /= t.SY
	}
	p := start
	p.X += dx
	p.Y += dy
	p, guides := o.ed.snapPlace(primary, p)
	o.guides = guides
	o.guideParent = parent
	// Apply the snapped correction to the whole selection.
	corrX, corrY := p.X-start.X-dx, p.Y-start.Y-dy
	for id, sp := range o.start {
		if !o.ed.parentFree(id) {
			continue
		}
		q := sp
		q.X += dx + corrX
		q.Y += dy + corrY
		o.lastPlaces[id] = q
		o.ed.view.PreviewPlace(id, q)
		// Track the outline/handle rectangle from the preview, so the chrome
		// follows the pointer instead of sitting at the committed position.
		if sr, ok := o.startRects[id]; ok {
			o.previewRects[id] = render.NewRect(
				sr.X+float32((q.X-sp.X)*t.SX),
				sr.Y+float32((q.Y-sp.Y)*t.SY),
				sr.W, sr.H,
			)
		}
	}
	if target, spot, ok := o.dropTarget(parent); ok {
		o.dropParent, o.dropRect = target, spot.Rect
	} else {
		o.dropParent, o.dropRect = "", render.Rect{}
	}
}

// dropTarget reports a different container under the pointer that can accept
// the dragged node.
func (o *overlay) dropTarget(current string) (string, render.DropSpot, bool) {
	// Reparenting is explicit: hold Alt. Without it a drag that happens to
	// pass over another container silently moves the node into it.
	if o.mods()&fyne.KeyModifierAlt == 0 {
		return "", render.DropSpot{}, false
	}
	parent, spot, ok := o.ed.view.DropSpotAt(o.mouse)
	if !ok || parent == "" || parent == current {
		return "", render.DropSpot{}, false
	}
	if o.ed.doc.IsAncestor(o.ed.primary, parent) {
		return "", render.DropSpot{}, false
	}
	return parent, spot, true
}

func (o *overlay) dragPoint() {
	id := o.ed.primary
	if !o.polygonSelected() {
		return
	}
	r, ok := o.startRects[id]
	if !ok || r.W <= 0 || r.H <= 0 {
		return
	}
	idx := 0
	if _, err := fmt.Sscanf(o.handle, "pt%d", &idx); err != nil {
		return
	}
	n, ok := o.ed.doc.Find(id)
	if !ok {
		return
	}
	pts := polygonPointsOf(n)
	if idx < 0 || idx >= len(pts) {
		return
	}
	x := (o.mouse.X - r.X) / r.W
	y := (o.mouse.Y - r.Y) / r.H
	x = float32(maxf(float64(x), 0))
	y = float32(maxf(float64(y), 0))
	x = float32(minf(float64(x), 1))
	y = float32(minf(float64(y), 1))
	pts[idx] = map[string]any{"x": scene.Round(float64(x)), "y": scene.Round(float64(y))}
	o.lastPoints = pts
	o.ed.view.RuntimeSet(id, "points", pts)
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// dragResize resizes by moving edges, keeping the opposite edge fixed. Snapping
// applies to the moving edges only (snapping x and width independently made the
// anchored edge jitter), and the minimum size stops the moving edge instead of
// sliding the whole box.
func (o *overlay) dragResize() {
	id := o.ed.primary
	start, ok := o.start[id]
	if !ok || !o.ed.parentFree(id) || o.handle == "" {
		return
	}
	parent := o.ed.view.ParentID(id)
	t, _, _ := o.ed.freeTransform(parent)
	dx, dy := float64(o.acc.DX), float64(o.acc.DY)
	if t.SX != 0 {
		dx /= t.SX
	}
	if t.SY != 0 {
		dy /= t.SY
	}
	mods := o.mods()
	// Alt temporarily disables snapping (same convention as move).
	snap := o.ed.snapGrid > 0 && mods&fyne.KeyModifierAlt == 0

	left, top := start.X, start.Y
	right, bottom := start.X+start.W, start.Y+start.H
	moveW := strings.Contains(o.handle, "w")
	moveE := strings.Contains(o.handle, "e")
	moveN := strings.Contains(o.handle, "n")
	moveS := strings.Contains(o.handle, "s")
	if moveW {
		left += dx
	}
	if moveE {
		right += dx
	}
	if moveN {
		top += dy
	}
	if moveS {
		bottom += dy
	}

	const minSize = 8.0
	if mods&fyne.KeyModifierShift != 0 && start.W > 0 && start.H > 0 && moveW != moveE && moveN != moveS {
		// Aspect lock around the opposite corner.
		ratio := start.H / start.W
		w := right - left
		h := w * ratio
		if moveN {
			top = bottom - h
		} else {
			bottom = top + h
		}
	}

	if snap {
		if moveW {
			left = roundTo(left, o.ed.snapGrid)
		}
		if moveE {
			right = roundTo(right, o.ed.snapGrid)
		}
		if moveN {
			top = roundTo(top, o.ed.snapGrid)
		}
		if moveS {
			bottom = roundTo(bottom, o.ed.snapGrid)
		}
	}

	if right-left < minSize {
		if moveW {
			left = right - minSize
		} else {
			right = left + minSize
		}
	}
	if bottom-top < minSize {
		if moveN {
			top = bottom - minSize
		} else {
			bottom = top + minSize
		}
	}

	p := scene.Place{X: left, Y: top, W: right - left, H: bottom - top}
	o.lastPlaces[id] = p
	o.ed.view.PreviewPlace(id, p)
	if sr, ok := o.startRects[id]; ok {
		o.previewRects[id] = render.NewRect(
			sr.X+float32((p.X-start.X)*t.SX),
			sr.Y+float32((p.Y-start.Y)*t.SY),
			float32(p.W*t.SX),
			float32(p.H*t.SY),
		)
	}
}

func (o *overlay) commitMove() {
	if o.dropParent != "" && o.dropParent != o.ed.view.ParentID(o.ed.primary) {
		o.commitReparent()
		return
	}
	if len(o.lastPlaces) == 0 {
		// Ordered parent: move the node to the drop index.
		id := o.ed.primary
		if id == "" {
			return
		}
		parent, spot, ok := o.ed.view.DropSpotAt(o.mouse)
		if !ok || parent == "" {
			return
		}
		curParent := o.ed.view.ParentID(id)
		if parent == curParent {
			if n, ok := o.ed.doc.Find(parent); ok {
				_ = n
			}
			op, err := scene.MoveOp(o.ed.doc, id, parent, spot.Index)
			if err == nil {
				o.ed.applyEdit("reorder", "", op)
			}
			return
		}
		if o.ed.doc.IsAncestor(id, parent) {
			o.ed.setStatus("Cannot move a node into itself")
			return
		}
		op, err := scene.MoveOp(o.ed.doc, id, parent, spot.Index)
		if err != nil {
			o.ed.setStatus(err.Error())
			return
		}
		o.ed.applyEdit("move", "", op)
		o.ed.Select(id)
		return
	}
	var ops []scene.Op
	for id, p := range o.lastPlaces {
		old, ok := o.start[id]
		if !ok {
			continue
		}
		if old == p {
			continue
		}
		ops = append(ops, scene.SetPlaceOp(id, old, p))
	}
	if len(ops) > 0 {
		o.ed.applyEdit("move", "", ops...)
		o.ed.inspector.refresh()
	}
}

// commitReparent moves the primary node into the container under the pointer.
func (o *overlay) commitReparent() {
	id := o.ed.primary
	target := o.dropParent
	_, spot, ok := o.ed.view.DropSpotAt(o.mouse)
	if !ok {
		return
	}
	old, has := o.start[id]
	if !has {
		if n, ok := o.ed.doc.Find(id); ok {
			old = n.Place
		}
	}
	p := scene.Place{}
	switch o.ed.view.ContainerKind(target) {
	case scene.Free:
		r := o.startRects[id]
		r.X += o.acc.DX
		r.Y += o.acc.DY
		if t, _, ok := o.ed.freeTransform(target); ok {
			x, y := t.ToDesign(render.NewRect(r.X, r.Y, 0, 0).Position())
			p = scene.Place{X: x, Y: y, W: float64(r.W) / t.SX, H: float64(r.H) / t.SY}
		} else {
			p = old
		}
	default:
		p = scene.Place{MinW: old.W, MinH: old.H, Slot: spot.Slot}
		if spec, ok := o.ed.opts.Registry.Spec(target); ok && spec.Kind == scene.Tabs {
			p.Title = old.Title
			if p.Title == "" {
				p.Title = "Tab"
			}
		}
	}
	mv, err := scene.MoveOp(o.ed.doc, id, target, spot.Index)
	if err != nil {
		o.ed.setStatus(err.Error())
		return
	}
	o.ed.applyEdit("move", "", mv, scene.SetPlaceOp(id, old, p))
	o.ed.Select(id)
	o.ed.inspector.refresh()
}

func (o *overlay) commitResize() {
	if strings.HasPrefix(o.handle, "pt") {
		o.commitPolygon()
		return
	}
	id := o.ed.primary
	p, ok := o.lastPlaces[id]
	if !ok {
		return
	}
	old, ok := o.start[id]
	if !ok || old == p {
		return
	}
	o.ed.applyEdit("resize", "", scene.SetPlaceOp(id, old, p))
	o.ed.inspector.refresh()
}

func (o *overlay) commitPolygon() {
	id := o.ed.primary
	if id == "" || o.lastPoints == nil {
		return
	}
	n, ok := o.ed.doc.Find(id)
	if !ok {
		return
	}
	old, _ := n.Prop("points")
	o.ed.applyEdit("polygon", "", scene.SetPropOp(id, "points", old, o.lastPoints))
	o.ed.inspector.refresh()
}

func (o *overlay) commitCreate() {
	parent := o.ed.view.ContainerAt(o.press)
	if parent == "" {
		parent = o.ed.doc.Root.ID
	}
	var p scene.Place
	if o.ed.view.ContainerKind(parent) == scene.Free {
		t, _, ok := o.ed.freeTransform(parent)
		if ok {
			x1, y1 := t.ToDesign(fyne.NewPos(fyne.Min(o.press.X, o.mouse.X), fyne.Min(o.press.Y, o.mouse.Y)))
			x2, y2 := t.ToDesign(fyne.NewPos(fyne.Max(o.press.X, o.mouse.X), fyne.Max(o.press.Y, o.mouse.Y)))
			p = scene.Place{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
		}
		if p.W < 4 || p.H < 4 {
			p = o.ed.defaultPlaceAt(o.press, parent)
		}
	}
	o.createAt(parent, p, o.mouse)
}

func (o *overlay) createClick(pos fyne.Position) {
	parent := o.ed.view.ContainerAt(pos)
	if parent == "" {
		parent = o.ed.doc.Root.ID
	}
	var p scene.Place
	if o.ed.view.ContainerKind(parent) == scene.Free {
		p = o.ed.defaultPlaceAt(pos, parent)
	}
	o.createAt(parent, p, pos)
}

func (o *overlay) createAt(parent string, p scene.Place, pos fyne.Position) {
	typ := o.ed.tool
	if typ == "" {
		return
	}
	index := -1
	if o.ed.view.ContainerKind(parent) != scene.Free {
		_, spot, ok := o.ed.view.DropSpotAt(pos)
		if ok {
			index = spot.Index
			p = scene.Place{}
		}
	}
	if index < 0 {
		index = o.ed.endIndex(parent)
	}
	o.ed.addNode(typ, parent, index, p)
	o.ed.DisarmTool()
}

func (o *overlay) nodesIn(r render.Rect) []string {
	var out []string
	root := ""
	if o.ed.doc.Root != nil {
		root = o.ed.doc.Root.ID
	}
	o.ed.view.EachRect(func(id string, rr render.Rect) {
		if id == root {
			return
		}
		n, ok := o.ed.doc.Find(id)
		if !ok || n.Locked || n.Hidden || !o.ed.view.IsVisible(id) {
			return
		}
		if r.Intersects(rr) {
			out = append(out, id)
		}
	})
	return out
}

// ---------------------------------------------------------------- handles

var handleNames = []string{"nw", "n", "ne", "e", "se", "s", "sw", "w"}

func handleRects(r render.Rect) map[string]render.Rect {
	const s float32 = 8
	out := make(map[string]render.Rect, 8)
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	out["nw"] = render.NewRect(r.X-s/2, r.Y-s/2, s, s)
	out["n"] = render.NewRect(cx-s/2, r.Y-s/2, s, s)
	out["ne"] = render.NewRect(r.X+r.W-s/2, r.Y-s/2, s, s)
	out["e"] = render.NewRect(r.X+r.W-s/2, cy-s/2, s, s)
	out["se"] = render.NewRect(r.X+r.W-s/2, r.Y+r.H-s/2, s, s)
	out["s"] = render.NewRect(cx-s/2, r.Y+r.H-s/2, s, s)
	out["sw"] = render.NewRect(r.X-s/2, r.Y+r.H-s/2, s, s)
	out["w"] = render.NewRect(r.X-s/2, cy-s/2, s, s)
	return out
}

func (o *overlay) handleAt(pos fyne.Position) string {
	ids := o.ed.selectedIDs()
	if len(ids) != 1 {
		return ""
	}
	if o.polygonSelected() {
		r, _ := o.ed.view.RectOf(ids[0])
		for name, hr := range o.pointRects(r) {
			if hr.Inset(-(o.ed.handlePad())).Contains(pos) {
				return name
			}
		}
		return ""
	}
	if !o.ed.parentFree(ids[0]) {
		return ""
	}
	r, ok := o.ed.view.RectOf(ids[0])
	if !ok {
		return ""
	}
	for _, name := range handleNames {
		if handleRects(r)[name].Inset(-o.ed.handlePad()).Contains(pos) {
			return name
		}
	}
	return ""
}

// polygonSelected reports whether the primary node is a polygon hotspot.
func (o *overlay) polygonSelected() bool {
	n, ok := o.ed.Primary()
	if !ok || n.Type != "hotspot" || n.StringProp("shape", "") != "polygon" {
		return false
	}
	pts, ok := n.Prop("points")
	return ok && pts != nil
}

func (o *overlay) pointRects(r render.Rect) map[string]render.Rect {
	n, ok := o.ed.Primary()
	if !ok {
		return nil
	}
	list, _ := n.Prop("points")
	items, _ := list.([]any)
	out := make(map[string]render.Rect, len(items))
	size := float32(o.ed.handleSize())
	for i, item := range items {
		m, _ := item.(map[string]any)
		x, _ := m["x"].(float64)
		y, _ := m["y"].(float64)
		cx := r.X + float32(x)*r.W
		cy := r.Y + float32(y)*r.H
		out[fmt.Sprintf("pt%d", i)] = render.NewRect(cx-size/2, cy-size/2, size, size)
	}
	return out
}

// polygonPoints returns the node's point list (copyable).
func polygonPointsOf(n *scene.Node) []any {
	v, _ := n.Prop("points")
	list, _ := v.([]any)
	out := make([]any, len(list))
	copy(out, list)
	return out
}

// ------------------------------------------------------------------ hover

func (o *overlay) MouseMoved(e *desktop.MouseEvent) {
	o.mouse = e.Position
	id := o.ed.hitTest(e.Position)
	if id != o.hover {
		o.hover = id
		o.Refresh()
	}
}

func (o *overlay) MouseIn(*desktop.MouseEvent) {}
func (o *overlay) MouseOut()                   { o.hover = ""; o.Refresh() }

func (o *overlay) Cursor() desktop.Cursor {
	if o.ed.tool != "" {
		return desktop.CrosshairCursor
	}
	if o.kind == gResizing && o.handle != "" {
		return cursorForHandle(o.handle)
	}
	switch o.handleAt(o.mouse) {
	case "nw", "se":
		return desktop.NWSEResizeCursor
	case "ne", "sw":
		return desktop.NESWResizeCursor
	case "n", "s":
		return desktop.VResizeCursor
	case "e", "w":
		return desktop.HResizeCursor
	}
	return desktop.DefaultCursor
}

func cursorForHandle(h string) desktop.Cursor {
	switch h {
	case "nw", "se":
		return desktop.NWSEResizeCursor
	case "ne", "sw":
		return desktop.NESWResizeCursor
	case "n", "s":
		return desktop.VResizeCursor
	case "e", "w":
		return desktop.HResizeCursor
	}
	return desktop.DefaultCursor
}

// --------------------------------------------------------------- keyboard

func (o *overlay) TypedRune(rune) {}

func (o *overlay) FocusGained() { o.focused = true; o.Refresh() }
func (o *overlay) FocusLost()   { o.focused = false; o.Refresh() }

func (o *overlay) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyF4:
		o.ed.RequestClose()
		return
	case fyne.KeyEscape:
		if o.ed.tool != "" {
			o.ed.DisarmTool()
			return
		}
		if len(o.ed.sel) > 0 {
			o.ed.selectParent()
			return
		}
		o.ed.ClearSelection()
	case fyne.KeyDelete, fyne.KeyBackspace:
		if len(o.ed.sel) > 0 {
			o.ed.deleteSelection()
		}
	case fyne.KeyLeft:
		o.nudge(-1, 0)
	case fyne.KeyRight:
		o.nudge(1, 0)
	case fyne.KeyUp:
		o.nudge(0, -1)
	case fyne.KeyDown:
		o.nudge(0, 1)
	}
}

func (o *overlay) TypedShortcut(s fyne.Shortcut) {
	switch sh := s.(type) {
	case *fyne.ShortcutUndo:
		o.ed.undo()
	case *fyne.ShortcutRedo:
		o.ed.redo()
	case *fyne.ShortcutCopy:
		o.ed.copySelection()
	case *fyne.ShortcutCut:
		o.ed.cutSelection()
	case *fyne.ShortcutPaste:
		o.ed.pasteClipboard()
	case *fyne.ShortcutSelectAll:
		o.ed.selectAll()
	case *desktop.CustomShortcut:
		ctrl := sh.Modifier&(fyne.KeyModifierControl|fyne.KeyModifierSuper) != 0
		switch {
		case ctrl && sh.KeyName == fyne.KeyS:
			o.ed.Save()
		case ctrl && sh.KeyName == fyne.KeyD:
			o.ed.duplicateSelection()
		}
	}
}

func (o *overlay) nudge(dx, dy float64) {
	var ops []scene.Op
	for _, id := range o.ed.selectedIDs() {
		if !o.ed.parentFree(id) {
			continue
		}
		n, ok := o.ed.doc.Find(id)
		if !ok {
			continue
		}
		p := n.Place
		step := float64(1)
		if o.ed.snapGrid > 0 {
			step = o.ed.snapGrid
		}
		if o.mods()&fyne.KeyModifierShift != 0 {
			step *= 8
		}
		p.X += dx * step
		p.Y += dy * step
		ops = append(ops, scene.SetPlaceOp(id, n.Place, p))
	}
	if len(ops) > 0 {
		// Key repeat must not create one undo step per pixel.
		o.ed.applyEdit("nudge", "nudge:"+strings.Join(o.ed.selectedIDs(), ","), ops...)
		o.ed.inspector.refresh()
	}
}

// --------------------------------------------------------------- rendering

type overlayRenderer struct{ o *overlay }

func (r *overlayRenderer) Layout(fyne.Size)   { r.o.refreshChrome() }
func (r *overlayRenderer) MinSize() fyne.Size { return fyne.NewSize(1, 1) }
func (r *overlayRenderer) Refresh()           { r.o.refreshChrome() }
func (r *overlayRenderer) Destroy()           {}
func (r *overlayRenderer) Objects() []fyne.CanvasObject {
	return r.o.objs
}

// chromeRect returns the rectangle to draw chrome for id: the live preview
// during a gesture, a translated rect for a selected descendant of a dragged
// node, or the view's committed rect.
func (o *overlay) chromeRect(id string) (render.Rect, bool) {
	if r, ok := o.previewRects[id]; ok {
		return r, true
	}
	for p := o.ed.view.ParentID(id); p != ""; p = o.ed.view.ParentID(p) {
		pr, ok := o.previewRects[p]
		if !ok {
			continue
		}
		sr, ok := o.startRects[p]
		if !ok {
			continue
		}
		if r, ok := o.startRects[id]; ok {
			dx, dy := pr.X-sr.X, pr.Y-sr.Y
			return render.NewRect(r.X+dx, r.Y+dy, r.W, r.H), true
		}
	}
	return o.ed.view.RectOf(id)
}

func (o *overlay) refreshChrome() {
	o.objs = o.objs[:0]
	used := 0
	addRect := func(r render.Rect, stroke color.Color, width float32, fill color.Color) {
		// Reuse chrome rectangles: a refresh runs on every hover/drag event,
		// and allocating a fresh object per outline churned Fyne's render
		// cache for no visual benefit.
		var rect *canvas.Rectangle
		if used < len(o.chrome) {
			rect = o.chrome[used]
		} else {
			rect = canvas.NewRectangle(fill)
			o.chrome = append(o.chrome, rect)
		}
		rect.FillColor = fill
		rect.StrokeColor = stroke
		rect.StrokeWidth = width
		rect.Move(r.Position())
		rect.Resize(r.Size())
		rect.Show()
		o.objs = append(o.objs, rect)
		used++
	}
	blue := color.NRGBA{R: 59, G: 130, B: 246, A: 255}
	green := color.NRGBA{R: 34, G: 197, B: 94, A: 255}
	gray := color.NRGBA{R: 160, G: 160, B: 160, A: 200}

	for _, id := range o.ed.selectedIDs() {
		if r, ok := o.chromeRect(id); ok {
			addRect(r, blue, 1.5, color.NRGBA{})
		}
	}
	if len(o.ed.sel) == 1 && o.ed.primary != "" && o.polygonSelected() {
		if r, ok := o.chromeRect(o.ed.primary); ok {
			orange := color.NRGBA{R: 249, G: 115, B: 22, A: 255}
			for _, hr := range o.pointRects(r) {
				addRect(hr, orange, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	} else if len(o.ed.sel) == 1 && o.ed.primary != "" && o.ed.parentFree(o.ed.primary) {
		if r, ok := o.chromeRect(o.ed.primary); ok && o.ed.doc.Root != nil && o.ed.primary != o.ed.doc.Root.ID {
			for _, name := range handleNames {
				hr := handleRects(r)[name]
				addRect(hr, blue, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	if o.hover != "" && !o.ed.sel[o.hover] {
		if r, ok := o.chromeRect(o.hover); ok {
			addRect(r, gray, 1, color.NRGBA{})
		}
	}
	// Edit-mode outlines for invisible interactive regions. A hotspot's own
	// edit fill is white-ish and vanishes on a light page, so the overlay
	// draws its own outline independent of the page theme.
	o.ed.doc.Walk(func(n *scene.Node) bool {
		if n.Type != "hotspot" && n.Type != "slot" {
			return true
		}
		if o.ed.sel[n.ID] || o.hover == n.ID || n.Hidden || !o.ed.view.IsVisible(n.ID) {
			return true
		}
		if r, ok := o.ed.view.RectOf(n.ID); ok && r.W > 1 && r.H > 1 {
			addRect(r, gray, 1, color.NRGBA{})
		}
		return true
	})
	if o.dropRect.W > 0 || o.dropRect.H > 0 {
		addRect(o.dropRect, green, 1, color.NRGBA{R: 34, G: 197, B: 94, A: 64})
	}
	if o.marquee.W > 0 || o.marquee.H > 0 {
		addRect(o.marquee, blue, 1, color.NRGBA{R: 59, G: 130, B: 246, A: 40})
	}
	if o.createRect.W > 0 || o.createRect.H > 0 {
		addRect(o.createRect, green, 1, color.NRGBA{R: 34, G: 197, B: 94, A: 40})
	}
	if o.inline != nil {
		o.inline.Move(o.inlineRect.Position())
		o.inline.Resize(o.inlineRect.Size())
		o.objs = append(o.objs, o.inline)
	}
	if len(o.guides) > 0 {
		t, _, ok := o.ed.freeTransform(o.guideParent)
		if ok {
			red := color.NRGBA{R: 239, G: 68, B: 68, A: 220}
			for _, g := range o.guides {
				if g.vertical {
					x := float32(t.OX + g.at*t.SX)
					addRect(render.NewRect(x, 0, 1, o.Size().Height), red, 0, red)
				} else {
					y := float32(t.OY + g.at*t.SY)
					addRect(render.NewRect(0, y, o.Size().Width, 1), red, 0, red)
				}
			}
		}
	}
	// Hide pooled rectangles that this frame did not use.
	for i := used; i < len(o.chrome); i++ {
		o.chrome[i].Hide()
	}
}

// ------------------------------------------------------------------ helpers

func rectBetween(a, b fyne.Position) render.Rect {
	x1, x2 := fyne.Min(a.X, b.X), fyne.Max(a.X, b.X)
	y1, y2 := fyne.Min(a.Y, b.Y), fyne.Max(a.Y, b.Y)
	return render.NewRect(x1, y1, x2-x1, y2-y1)
}

// mods returns the modifier state, honouring the test override.
func (o *overlay) mods() fyne.KeyModifier {
	if o.ed.modsOverride != 0 {
		return o.ed.modsOverride
	}
	return currentMods()
}

func currentMods() fyne.KeyModifier {
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return 0
	}
	if d, ok := app.Driver().(desktop.Driver); ok {
		return d.CurrentKeyModifiers()
	}
	return 0
}

var _ fyne.Tappable = (*overlay)(nil)
var _ fyne.SecondaryTappable = (*overlay)(nil)
var _ fyne.Draggable = (*overlay)(nil)
var _ desktop.Mouseable = (*overlay)(nil)
var _ desktop.Hoverable = (*overlay)(nil)
var _ desktop.Cursorable = (*overlay)(nil)
var _ fyne.Focusable = (*overlay)(nil)
var _ fyne.Shortcutable = (*overlay)(nil)
