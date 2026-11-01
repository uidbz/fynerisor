package designer

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/scene"
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
