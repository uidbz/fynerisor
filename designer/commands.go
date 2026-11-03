package designer

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

const fragmentFormat = "fynerisor.scene.fragment/1"

type fragment struct {
	Format string        `json:"format"`
	Nodes  []*scene.Node `json:"nodes"`
}

// defaultPlaceAt returns a default-sized placement centered on a canvas point.
func (e *Editor) defaultPlaceAt(pos fyne.Position, parent string) scene.Place {
	spec, ok := e.opts.Registry.Spec(e.tool)
	if !ok {
		return scene.Place{W: 100, H: 100}
	}
	p := spec.DefaultPlace
	if p.W == 0 {
		p.W = 100
	}
	if p.H == 0 {
		p.H = 100
	}
	if e.view.ContainerKind(parent) == scene.Free {
		if t, _, ok := e.freeTransform(parent); ok {
			x, y := t.ToDesign(pos)
			p.X = roundTo(x-float64(p.W)/2, e.snapGrid)
			p.Y = roundTo(y-float64(p.H)/2, e.snapGrid)
		}
		return p
	}
	return scene.Place{MinW: p.W, MinH: p.H}
}

// selectParent moves the selection to the primary node's parent.
func (e *Editor) selectParent() {
	if e.primary == "" {
		return
	}
	if parent := e.view.ParentID(e.primary); parent != "" {
		e.Select(parent)
		return
	}
	e.ClearSelection()
}

// selectAll selects the root's direct children.
func (e *Editor) selectAll() {
	if e.doc.Root == nil {
		return
	}
	var ids []string
	for _, c := range e.doc.Root.Children {
		ids = append(ids, c.ID)
	}
	e.SelectMany(ids)
}

// bringToFront moves the primary node to the end of its parent's child list.
func (e *Editor) bringToFront() {
	if e.primary == "" {
		return
	}
	parent, _, ok := e.doc.ParentOf(e.primary)
	if !ok || parent == nil {
		return
	}
	op, err := scene.MoveOp(e.doc, e.primary, parent.ID, len(parent.Children)-1)
	if err != nil {
		e.setStatus(err.Error())
		return
	}
	e.applyEdit("bring to front", "", op)
}

// sendToBack moves the primary node to the start of its parent's child list.
func (e *Editor) sendToBack() {
	if e.primary == "" {
		return
	}
	parent, _, ok := e.doc.ParentOf(e.primary)
	if !ok || parent == nil {
		return
	}
	op, err := scene.MoveOp(e.doc, e.primary, parent.ID, 0)
	if err != nil {
		e.setStatus(err.Error())
		return
	}
	e.applyEdit("send to back", "", op)
}

// copySelection writes the selected subtrees to the clipboard as canonical
// scene fragment JSON.
func (e *Editor) copySelection() bool {
	ids := e.selectedIDs()
	if len(ids) == 0 {
		return false
	}
	var nodes []*scene.Node
	for _, id := range ids {
		if n, ok := e.doc.Find(id); ok {
			nodes = append(nodes, n.Clone())
		}
	}
	if len(nodes) == 0 {
		return false
	}
	data, err := json.MarshalIndent(fragment{Format: fragmentFormat, Nodes: nodes}, "", "  ")
	if err != nil {
		e.setStatus(err.Error())
		return false
	}
	app := fyne.CurrentApp()
	if app == nil {
		e.setStatus("No application clipboard")
		return false
	}
	app.Clipboard().SetContent(string(data))
	e.setStatus(fmt.Sprintf("Copied %d node(s)", len(nodes)))
	return true
}

// cutSelection copies and deletes the selection.
func (e *Editor) cutSelection() {
	if !e.copySelection() {
		return
	}
	e.deleteSelection()
}

func (e *Editor) pasteClipboard() {
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	text := app.Clipboard().Content()
	if text == "" {
		return
	}
	var frag fragment
	if err := json.Unmarshal([]byte(text), &frag); err != nil || frag.Format != fragmentFormat {
		e.setStatus("Clipboard does not contain scene nodes")
		return
	}
	if len(frag.Nodes) == 0 {
		return
	}
	parent := e.pasteParent()
	if parent == "" {
		e.setStatus("No container for pasting")
		return
	}
	pn, ok := e.doc.Find(parent)
	if !ok {
		return
	}
	pspec, _ := e.opts.Registry.Spec(pn.Type)
	taken := e.takenIDs()
	var ops []scene.Op
	var newIDs []string
	for _, n := range frag.Nodes {
		if n == nil {
			continue
		}
		clone := n.Clone()
		clone.ReID(taken)
		if pspec.Kind == scene.Free {
			clone.Place.X += 10
			clone.Place.Y += 10
		} else {
			clone.Place.X, clone.Place.Y, clone.Place.W, clone.Place.H = 0, 0, 0, 0
			if spec, ok := e.opts.Registry.Spec(clone.Type); ok {
				if clone.Place.MinW == 0 {
					clone.Place.MinW = float64(spec.DefaultPlace.W)
				}
				if clone.Place.MinH == 0 {
					clone.Place.MinH = float64(spec.DefaultPlace.H)
				}
			}
			switch pspec.Kind {
			case scene.Slots:
				used := map[string]bool{}
				for _, c := range pn.Children {
					used[c.Place.Slot] = true
				}
				slot := ""
				for _, s := range pspec.Slots {
					if !used[s] {
						slot = s
						break
					}
				}
				if slot == "" && len(pspec.Slots) > 0 {
					slot = pspec.Slots[0]
				}
				clone.Place.Slot = slot
			case scene.Tabs:
				clone.Place.Title = clone.Type
			}
		}
		ops = append(ops, scene.InsertOp(parent, e.endIndex(parent), clone))
		newIDs = append(newIDs, clone.ID)
	}
	if e.apply("paste", "", ops...) {
		e.SelectMany(newIDs)
	}
}

func (e *Editor) pasteParent() string {
	if e.primary != "" {
		if e.view.ContainerKind(e.primary) != scene.Leaf {
			return e.primary
		}
		if p := e.view.ParentID(e.primary); p != "" {
			return p
		}
	}
	if e.doc.Root != nil {
		return e.doc.Root.ID
	}
	return ""
}

// alignSelection aligns selected free siblings on an edge (left, right, top,
// bottom, centerX, centerY).
func (e *Editor) alignSelection(edge string) {
	ids := e.selectedIDs()
	if len(ids) < 2 {
		return
	}
	parent := e.view.ParentID(ids[0])
	if parent == "" || e.view.ContainerKind(parent) != scene.Free {
		e.setStatus("Align needs at least two nodes in a free container")
		return
	}
	for _, id := range ids {
		if e.view.ParentID(id) != parent {
			e.setStatus("Align needs siblings in the same container")
			return
		}
	}
	type box struct {
		left, top, right, bottom float64
	}
	var b box
	first := true
	for _, id := range ids {
		n, _ := e.doc.Find(id)
		p := n.Place
		if first {
			b = box{p.X, p.Y, p.X + p.W, p.Y + p.H}
			first = false
			continue
		}
		if p.X < b.left {
			b.left = p.X
		}
		if p.Y < b.top {
			b.top = p.Y
		}
		if p.X+p.W > b.right {
			b.right = p.X + p.W
		}
		if p.Y+p.H > b.bottom {
			b.bottom = p.Y + p.H
		}
	}
	var ops []scene.Op
	for _, id := range ids {
		n, _ := e.doc.Find(id)
		p := n.Place
		switch edge {
		case "left":
			p.X = b.left
		case "right":
			p.X = b.right - p.W
		case "top":
			p.Y = b.top
		case "bottom":
			p.Y = b.bottom - p.H
		case "centerX":
			p.X = (b.left + b.right - p.W) / 2
		case "centerY":
			p.Y = (b.top + b.bottom - p.H) / 2
		default:
			return
		}
		ops = append(ops, scene.SetPlaceOp(id, n.Place, p))
	}
	e.applyEdit("align "+edge, "", ops...)
}

// wrapSelection puts the selected siblings into a new container.
func (e *Editor) wrapSelection(kind string) {
	ids := e.selectedIDs()
	if len(ids) == 0 {
		return
	}
	parent := e.view.ParentID(ids[0])
	if parent == "" {
		e.setStatus("Cannot wrap the root node")
		return
	}
	for _, id := range ids {
		if e.view.ParentID(id) != parent {
			e.setStatus("Wrap needs siblings in the same container")
			return
		}
	}
	pn, ok := e.doc.Find(parent)
	if !ok {
		return
	}
	pspec, _ := e.opts.Registry.Spec(pn.Type)
	if pspec.Kind != scene.Free && pspec.Kind != scene.Ordered && pspec.Kind != scene.Tabs {
		e.setStatus("Wrap is not supported in this container")
		return
	}
	var b render.Rect
	first := true
	for _, id := range ids {
		r, ok := e.view.RectOf(id)
		if !ok {
			continue
		}
		if first {
			b, first = r, false
		} else {
			b = b.Union(r)
		}
	}
	if first {
		return
	}
	// Rect is in view space; convert to the parent's design/local space.
	pr, _ := e.view.RectOf(parent)
	local := render.NewRect(b.X-pr.X, b.Y-pr.Y, b.W, b.H)
	newID := e.uniqueID(kind)
	spec, _ := e.opts.Registry.Spec(kind)
	if spec.Type == "" {
		e.setStatus("Unknown container " + kind)
		return
	}
	container := &scene.Node{ID: newID, Type: kind, Name: "Group"}
	for _, ps := range spec.Props {
		if ps.Default == nil {
			continue
		}
		if s, ok := ps.Default.(string); ok && s == "" {
			continue
		}
		container.SetProp(ps.Name, ps.Default)
	}
	var ops []scene.Op
	var childPlaces []scene.Place
	if pspec.Kind == scene.Free {
		container.Place = scene.Place{X: float64(local.X), Y: float64(local.Y), W: float64(local.W), H: float64(local.H)}
		if kind == "free" {
			container.SetProp("designW", float64(local.W))
			container.SetProp("designH", float64(local.H))
			container.SetProp("scale", "none")
		}
		for _, id := range ids {
			n, _ := e.doc.Find(id)
			cp := n.Place
			cp.X -= container.Place.X
			cp.Y -= container.Place.Y
			childPlaces = append(childPlaces, cp)
		}
	} else {
		container.Place = scene.Place{MinW: float64(local.W), MinH: float64(local.H)}
		for range ids {
			childPlaces = append(childPlaces, scene.Place{})
		}
	}
	ops = append(ops, scene.InsertOp(parent, e.endIndex(parent), container))
	for i, id := range ids {
		mv, err := scene.MoveOp(e.doc, id, newID, i)
		if err != nil {
			e.setStatus(err.Error())
			return
		}
		ops = append(ops, mv)
		if pspec.Kind == scene.Free {
			n, _ := e.doc.Find(id)
			ops = append(ops, scene.SetPlaceOp(id, n.Place, childPlaces[i]))
		}
	}
	if e.apply("wrap "+kind, "", ops...) {
		e.Select(newID)
	}
}

// unwrapSelection moves a container's children to its parent and removes it.
func (e *Editor) unwrapSelection() {
	n, ok := e.Primary()
	if !ok || len(n.Children) == 0 {
		e.setStatus("Select a non-empty container to unwrap")
		return
	}
	parent := e.view.ParentID(n.ID)
	if parent == "" {
		e.setStatus("Cannot unwrap the root node")
		return
	}
	pn, _ := e.doc.Find(parent)
	pspec, _ := e.opts.Registry.Spec(pn.Type)
	var ops []scene.Op
	var moved []string
	for _, c := range n.Children {
		lr, ok := e.view.RectOf(c.ID)
		if !ok {
			continue
		}
		pr, _ := e.view.RectOf(n.ID)
		mv, err := scene.MoveOp(e.doc, c.ID, parent, e.endIndex(parent))
		if err != nil {
			e.setStatus(err.Error())
			return
		}
		ops = append(ops, mv)
		if pspec.Kind == scene.Free {
			p := scene.Place{
				X: float64(lr.X-pr.X) + n.Place.X,
				Y: float64(lr.Y-pr.Y) + n.Place.Y,
				W: float64(lr.W), H: float64(lr.H),
			}
			ops = append(ops, scene.SetPlaceOp(c.ID, c.Place, p))
		} else {
			p := scene.Place{MinW: float64(lr.W), MinH: float64(lr.H)}
			ops = append(ops, scene.SetPlaceOp(c.ID, c.Place, p))
		}
		moved = append(moved, c.ID)
	}
	rm, err := scene.RemoveOp(e.doc, n.ID)
	if err != nil {
		e.setStatus(err.Error())
		return
	}
	ops = append(ops, rm)
	if e.apply("unwrap", "", ops...) {
		e.SelectMany(moved)
	}
}

// convertSelection changes a container's layout kind, keeping the visual
// arrangement as close as possible. Positions come from the live rectangles.
func (e *Editor) convertSelection(kind string) {
	n, ok := e.Primary()
	if !ok {
		return
	}
	spec, ok := e.opts.Registry.Spec(n.Type)
	if !ok || (spec.Kind != scene.Free && spec.Kind != scene.Ordered) {
		e.setStatus("Select a free or ordered container")
		return
	}
	target, ok := e.opts.Registry.Spec(kind)
	if !ok || (target.Kind != scene.Free && target.Kind != scene.Ordered) {
		e.setStatus("Convert target must be free, vbox or hbox")
		return
	}
	if n.Type == kind {
		return
	}
	nr, _ := e.view.RectOf(n.ID)
	var ops []scene.Op
	ops = append(ops, scene.SetFieldOp(n.ID, "type", n.Type, kind))
	// Clear props that only make sense on the old kind.
	if target.Kind == scene.Ordered {
		for _, key := range []string{"designW", "designH", "scale", "background", "backgroundFill", "letterbox", "clip", "scaleText"} {
			if v, ok := n.Prop(key); ok {
				ops = append(ops, scene.SetPropOp(n.ID, key, v, nil))
			}
		}
		ops = append(ops, scene.SetPropOp(n.ID, "cols", nil, 1))
	} else {
		p := n.Place
		if p.W == 0 {
			p.W = float64(nr.W)
		}
		if p.H == 0 {
			p.H = float64(nr.H)
		}
		ops = append(ops, scene.SetPropOp(n.ID, "designW", nil, float64(nr.W)))
		ops = append(ops, scene.SetPropOp(n.ID, "designH", nil, float64(nr.H)))
		ops = append(ops, scene.SetPropOp(n.ID, "scale", nil, "none"))
		ops = append(ops, scene.SetPlaceOp(n.ID, n.Place, p))
	}
	for _, c := range n.Children {
		cr, ok := e.view.RectOf(c.ID)
		if !ok {
			continue
		}
		if target.Kind == scene.Free {
			// Local coordinates inside the newly free container.
			local := scene.Place{
				X: float64(cr.X - nr.X), Y: float64(cr.Y - nr.Y),
				W: float64(cr.W), H: float64(cr.H),
			}
			ops = append(ops, scene.SetPlaceOp(c.ID, c.Place, local))
		} else {
			ops = append(ops, scene.SetPlaceOp(c.ID, c.Place, scene.Place{MinW: float64(cr.W), MinH: float64(cr.H)}))
		}
	}
	if e.apply("convert to "+kind, "", ops...) {
		e.Select(n.ID)
	}
}

// setProp is a convenience used by conversion and inline editing.
func (e *Editor) setProp(id, key string, value any, merge string) {
	n, ok := e.doc.Find(id)
	if !ok {
		return
	}
	old, _ := n.Prop(key)
	e.applyEdit("set "+key, merge, scene.SetPropOp(id, key, old, value))
}

// addPolygonPoint appends a normalized point at a canvas position.
func (e *Editor) addPolygonPoint(id string, pos fyne.Position) {
	n, ok := e.doc.Find(id)
	if !ok {
		return
	}
	r, ok := e.view.RectOf(id)
	if !ok || r.W <= 0 || r.H <= 0 {
		return
	}
	x := float64((pos.X - r.X) / r.W)
	y := float64((pos.Y - r.Y) / r.H)
	if x < 0 || x > 1 || y < 0 || y > 1 {
		return
	}
	pts := polygonPointsOf(n)
	pts = append(pts, map[string]any{"x": scene.Round(x), "y": scene.Round(y)})
	old, _ := n.Prop("points")
	e.applyEdit("add point", "", scene.SetPropOp(id, "points", old, pts))
	e.inspector.refresh()
}

// removePolygonPoint deletes point idx.
func (e *Editor) removePolygonPoint(id string, idx int) {
	n, ok := e.doc.Find(id)
	if !ok {
		return
	}
	pts := polygonPointsOf(n)
	if idx < 0 || idx >= len(pts) || len(pts) <= 3 {
		e.setStatus("A polygon needs at least three points")
		return
	}
	next := append(append([]any(nil), pts[:idx]...), pts[idx+1:]...)
	old, _ := n.Prop("points")
	e.applyEdit("remove point", "", scene.SetPropOp(id, "points", old, next))
	e.inspector.refresh()
}
