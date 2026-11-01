package designer

import (
	"image/color"

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

	start       map[string]scene.Place
	startRects  map[string]render.Rect
	lastPlaces  map[string]scene.Place
	guideParent string
	guides      []guide

	marquee    render.Rect
	marqueeIDs []string
	dropRect   render.Rect
	dropParent string
	createRect render.Rect

	hover       string
	suppressTap bool
	focused     bool
	objs        []fyne.CanvasObject
}

func newOverlay(e *Editor) *overlay {
	o := &overlay{ed: e, start: map[string]scene.Place{}, startRects: map[string]render.Rect{}, lastPlaces: map[string]scene.Place{}}
	o.ExtendBaseWidget(o)
	return o
}

func (o *overlay) CreateRenderer() fyne.WidgetRenderer {
	o.refreshChrome()
	return &overlayRenderer{o: o}
}

// ---------------------------------------------------------------- gestures

func (o *overlay) begin(pos fyne.Position) {
	o.press, o.mouse = pos, pos
	o.acc = fyne.Delta{}
	o.start = map[string]scene.Place{}
	o.startRects = map[string]render.Rect{}
	o.lastPlaces = map[string]scene.Place{}
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
	mods := currentMods()
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
	o.begin(e.Position)
	if e.Button == desktop.MouseButtonSecondary {
		o.showContext(e.Position)
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
}

func (o *overlay) Tapped(e *fyne.PointEvent) {
	o.ed.focusOverlay()
	if o.suppressTap {
		o.suppressTap = false
		return
	}
	if o.ed.tool != "" {
		o.createClick(e.Position)
		return
	}
	id := o.ed.hitTest(e.Position)
	if id == "" {
		o.ed.ClearSelection()
		return
	}
	if currentMods()&(fyne.KeyModifierShift|fyne.KeyModifierControl|fyne.KeyModifierSuper) != 0 {
		o.ed.ToggleSelect(id)
		return
	}
	o.ed.Select(id)
}

func (o *overlay) DoubleTapped(e *fyne.PointEvent) {
	if o.ed.tool != "" {
		return
	}
	id := o.ed.hitTest(e.Position)
	if id != "" {
		o.ed.Select(id)
		if o.ed.inspector != nil {
			o.ed.inspector.focusFirst()
		}
	}
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
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Duplicate", o.ed.duplicateSelection),
		fyne.NewMenuItem("Delete", o.ed.deleteSelection),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Bring to front", o.ed.bringToFront),
		fyne.NewMenuItem("Send to back", o.ed.sendToBack),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Select parent", func() { o.ed.selectParent() }),
	}
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
		o.dragResize()
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
	o.ed.view.RefreshRects()
	o.Refresh()
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
	parent, spot, ok := o.ed.view.DropSpotAt(o.mouse)
	if !ok || parent == "" || parent == current {
		return "", render.DropSpot{}, false
	}
	if o.ed.doc.IsAncestor(o.ed.primary, parent) {
		return "", render.DropSpot{}, false
	}
	return parent, spot, true
}

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
	p := start
	switch o.handle {
	case "nw":
		p.X += dx
		p.Y += dy
		p.W -= dx
		p.H -= dy
	case "n":
		p.Y += dy
		p.H -= dy
	case "ne":
		p.Y += dy
		p.W += dx
		p.H -= dy
	case "e":
		p.W += dx
	case "se":
		p.W += dx
		p.H += dy
	case "s":
		p.H += dy
	case "sw":
		p.X += dx
		p.W -= dx
		p.H += dy
	case "w":
		p.X += dx
		p.W -= dx
	}
	if p.W < 8 {
		p.W = 8
	}
	if p.H < 8 {
		p.H = 8
	}
	if currentMods()&fyne.KeyModifierShift != 0 && start.W > 0 && start.H > 0 && (o.handle == "nw" || o.handle == "ne" || o.handle == "se" || o.handle == "sw") {
		ratio := start.H / start.W
		p.H = p.W * ratio
		if o.handle == "nw" || o.handle == "ne" {
			p.Y = start.Y + start.H - p.H
		}
	}
	if o.ed.snapGrid > 0 {
		p.X = roundTo(p.X, o.ed.snapGrid)
		p.Y = roundTo(p.Y, o.ed.snapGrid)
		p.W = roundTo(p.W, o.ed.snapGrid)
		p.H = roundTo(p.H, o.ed.snapGrid)
	}
	o.lastPlaces[id] = p
	o.ed.view.PreviewPlace(id, p)
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
	if len(ids) != 1 || !o.ed.parentFree(ids[0]) {
		return ""
	}
	r, ok := o.ed.view.RectOf(ids[0])
	if !ok {
		return ""
	}
	for _, name := range handleNames {
		if handleRects(r)[name].Inset(-3).Contains(pos) {
			return name
		}
	}
	return ""
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
		if currentMods()&fyne.KeyModifierShift != 0 {
			step *= 8
		}
		p.X += dx * step
		p.Y += dy * step
		ops = append(ops, scene.SetPlaceOp(id, n.Place, p))
	}
	if len(ops) > 0 {
		o.ed.applyEdit("nudge", "", ops...)
		o.ed.inspector.refresh()
	}
}

// --------------------------------------------------------------- rendering

type overlayRenderer struct{ o *overlay }

func (r *overlayRenderer) Layout(fyne.Size)   {}
func (r *overlayRenderer) MinSize() fyne.Size { return fyne.NewSize(1, 1) }
func (r *overlayRenderer) Refresh()           { r.o.refreshChrome() }
func (r *overlayRenderer) Destroy()           {}
func (r *overlayRenderer) Objects() []fyne.CanvasObject {
	return r.o.objs
}

func (o *overlay) refreshChrome() {
	o.objs = nil
	addRect := func(r render.Rect, stroke color.Color, width float32, fill color.Color) {
		rect := canvas.NewRectangle(fill)
		rect.StrokeColor = stroke
		rect.StrokeWidth = width
		rect.Move(r.Position())
		rect.Resize(r.Size())
		o.objs = append(o.objs, rect)
	}
	blue := color.NRGBA{R: 59, G: 130, B: 246, A: 255}
	green := color.NRGBA{R: 34, G: 197, B: 94, A: 255}
	gray := color.NRGBA{R: 160, G: 160, B: 160, A: 200}

	for _, id := range o.ed.selectedIDs() {
		if r, ok := o.ed.view.RectOf(id); ok {
			addRect(r, blue, 1.5, color.NRGBA{})
		}
	}
	if len(o.ed.sel) == 1 && o.ed.primary != "" && o.ed.parentFree(o.ed.primary) {
		if r, ok := o.ed.view.RectOf(o.ed.primary); ok && o.ed.doc.Root != nil && o.ed.primary != o.ed.doc.Root.ID {
			for _, name := range handleNames {
				hr := handleRects(r)[name]
				addRect(hr, blue, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	if o.hover != "" && !o.ed.sel[o.hover] {
		if r, ok := o.ed.view.RectOf(o.hover); ok {
			addRect(r, gray, 1, color.NRGBA{})
		}
	}
	if o.dropRect.W > 0 || o.dropRect.H > 0 {
		addRect(o.dropRect, green, 1, color.NRGBA{R: 34, G: 197, B: 94, A: 64})
	}
	if o.marquee.W > 0 || o.marquee.H > 0 {
		addRect(o.marquee, blue, 1, color.NRGBA{R: 59, G: 130, B: 246, A: 40})
	}
	if o.createRect.W > 0 || o.createRect.H > 0 {
		addRect(o.createRect, green, 1, color.NRGBA{R: 34, G: 197, B: 94, A: 40})
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
}

// ------------------------------------------------------------------ helpers

func rectBetween(a, b fyne.Position) render.Rect {
	x1, x2 := fyne.Min(a.X, b.X), fyne.Max(a.X, b.X)
	y1, y2 := fyne.Min(a.Y, b.Y), fyne.Max(a.Y, b.Y)
	return render.NewRect(x1, y1, x2-x1, y2-y1)
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
var _ fyne.DoubleTappable = (*overlay)(nil)
var _ fyne.SecondaryTappable = (*overlay)(nil)
var _ fyne.Draggable = (*overlay)(nil)
var _ desktop.Mouseable = (*overlay)(nil)
var _ desktop.Hoverable = (*overlay)(nil)
var _ desktop.Cursorable = (*overlay)(nil)
var _ fyne.Focusable = (*overlay)(nil)
var _ fyne.Shortcutable = (*overlay)(nil)
