package designer

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

func testDoc() *scene.Document {
	return &scene.Document{
		Title: "designer test",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 1000.0, "designH": 800.0, "scale": "none"},
			Children: []*scene.Node{
				{ID: "door", Type: "hotspot", Name: "Door",
					Place: scene.Place{X: 100, Y: 100, W: 200, H: 150}},
				{ID: "title", Type: "label",
					Place: scene.Place{X: 400, Y: 60, W: 300, H: 40},
					Props: map[string]any{"text": "Hello"}},
				{ID: "wall", Type: "vbox",
					Place: scene.Place{X: 100, Y: 500, W: 300, H: 200},
					Children: []*scene.Node{
						{ID: "wall-a", Type: "label", Props: map[string]any{"text": "A"}},
						{ID: "wall-b", Type: "label", Props: map[string]any{"text": "B"}},
					}},
			},
		},
	}
}

func newTestEditor(t *testing.T) (*Editor, *scene.MemoryStore) {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	store := scene.NewMemoryStore()
	doc := testDoc()
	e := New(doc, Options{
		Registry: renderbuiltin.NewRegistry(),
		Store:    store,
		Ref:      "mem://scene",
	})
	w := test.NewWindow(e.Object())
	w.Resize(fyne.NewSize(1200, 800))
	e.view.RefreshRects()
	e.overlay.Resize(e.stage.Size())
	e.overlay.Refresh()
	return e, store
}

func viewRect(t *testing.T, e *Editor, id string) render.Rect {
	t.Helper()
	e.view.RefreshRects()
	r, ok := e.view.RectOf(id)
	if !ok {
		t.Fatalf("no rect for %q", id)
	}
	return r
}

func TestDragMoveAndUndo(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SetGrid(0)
	r := viewRect(t, e, "door")
	o := e.overlay
	o.begin(fyne.NewPos(r.X+10, r.Y+10))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: 50, DY: 20},
	})
	o.DragEnd()
	n, _ := e.doc.Find("door")
	if n.Place.X != 150 || n.Place.Y != 120 {
		t.Fatalf("place after drag = %+v", n.Place)
	}
	if !e.Dirty() {
		t.Fatal("drag did not dirty the document")
	}
	e.undo()
	n, _ = e.doc.Find("door")
	if n.Place.X != 100 || n.Place.Y != 100 {
		t.Fatalf("undo place = %+v", n.Place)
	}
}

// TestDragChromeFollowsPreview: the selection box and handles are drawn from
// the drag preview, not from the committed document position.
func TestDragChromeFollowsPreview(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SetGrid(0)
	e.Select("door")
	r := viewRect(t, e, "door")
	o := e.overlay
	o.begin(fyne.NewPos(r.X+10, r.Y+10))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: 50, DY: 20},
	})
	cr, ok := o.chromeRect("door")
	if !ok {
		t.Fatal("no chrome rect during drag")
	}
	if cr.X != r.X+50 || cr.Y != r.Y+20 {
		t.Fatalf("chrome rect = %+v, want %+v", cr, render.NewRect(r.X+50, r.Y+20, r.W, r.H))
	}
	o.DragEnd()
	if _, ok := o.previewRects["door"]; ok {
		t.Fatal("preview rect survived DragEnd")
	}
}

func TestResizeHandle(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SetGrid(0)
	e.Select("door")
	r := viewRect(t, e, "door")
	o := e.overlay
	o.begin(fyne.NewPos(r.X+r.W, r.Y+r.H))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: 40, DY: 30},
	})
	o.DragEnd()
	n, _ := e.doc.Find("door")
	if n.Place.W != 240 || n.Place.H != 180 {
		t.Fatalf("size after resize = %+v", n.Place)
	}
	if n.Place.X != 100 || n.Place.Y != 100 {
		t.Fatalf("position moved during se resize = %+v", n.Place)
	}
}

func TestCreateFromPalette(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SetGrid(0)
	e.ArmTool("hotspot")
	o := e.overlay
	o.begin(fyne.NewPos(600, 300))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: 100, DY: 80},
	})
	o.DragEnd()
	root, _ := e.doc.Find("root")
	if len(root.Children) != 4 {
		t.Fatalf("children = %d, want 4", len(root.Children))
	}
	last := root.Children[len(root.Children)-1]
	if last.Type != "hotspot" {
		t.Fatalf("last child type = %q", last.Type)
	}
	if last.Place.W != 100 || last.Place.H != 80 {
		t.Fatalf("new place = %+v", last.Place)
	}
	if e.tool != "" {
		t.Fatal("tool was not disarmed")
	}
	if !e.sel[last.ID] {
		t.Fatal("new node is not selected")
	}
}

func TestDuplicateDeleteAndSelect(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("door")
	e.duplicateSelection()
	root, _ := e.doc.Find("root")
	if len(root.Children) != 4 {
		t.Fatalf("children after duplicate = %d", len(root.Children))
	}
	if len(e.sel) != 1 {
		t.Fatalf("selection after duplicate = %v", e.sel)
	}
	e.deleteSelection()
	root, _ = e.doc.Find("root")
	if len(root.Children) != 3 {
		t.Fatalf("children after delete = %d", len(root.Children))
	}
	e.undo()
	root, _ = e.doc.Find("root")
	if len(root.Children) != 4 {
		t.Fatalf("children after undo = %d", len(root.Children))
	}
}

func TestInspectorPropAndName(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("title")
	e.inspector.refresh()
	e.inspector.setProp("text", "Changed")
	n, _ := e.doc.Find("title")
	if n.StringProp("text", "") != "Changed" {
		t.Fatalf("text = %q", n.StringProp("text", ""))
	}
	e.undo()
	n, _ = e.doc.Find("title")
	if n.StringProp("text", "") != "Hello" {
		t.Fatalf("text after undo = %q", n.StringProp("text", ""))
	}

	e.Select("door")
	e.inspector.refresh()
	name := "Front door"
	var ops []scene.Op
	cur, _ := e.doc.Find("door")
	ops = append(ops, scene.SetFieldOp("door", "name", cur.Name, name))
	e.applyEdit("rename", "", ops...)
	cur, _ = e.doc.Find("door")
	if cur.Name != name {
		t.Fatalf("name = %q", cur.Name)
	}
}

func TestNudgeAndSnap(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)
	n, _ := e.doc.Find("door")
	if n.Place.X != 101 {
		t.Fatalf("nudge x = %v", n.Place.X)
	}
	p, guides := e.snapPlace("door", scene.Place{X: 497, Y: 62, W: 200, H: 150})
	if p.X != 400 && p.X != 500 {
		t.Fatalf("unexpected snap x: %+v", p)
	}
	if len(guides) == 0 {
		t.Fatal("expected snapping guides")
	}
}

func TestSaveToMemoryStore(t *testing.T) {
	e, store := newTestEditor(t)
	sig := newSaveSignal(e)
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)
	e.Save()
	sig.wait(t, 1)
	snap, err := store.Load(context.Background(), "mem://scene")
	if err != nil {
		t.Fatalf("save did not reach the store: %v", err)
	}
	if _, _, err := scene.DecodeValidated(snap.Data, e.opts.Registry.Registry); err != nil {
		t.Fatalf("saved document does not decode: %v", err)
	}
	if e.Dirty() {
		t.Fatal("document is still dirty after save")
	}
}

func TestKeyboardDeleteAndEscape(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("door")
	e.overlay.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDelete})
	if _, ok := e.doc.Find("door"); ok {
		t.Fatal("delete key did not remove the node")
	}
	e.undo()
	if _, ok := e.doc.Find("door"); !ok {
		t.Fatal("undo did not restore the node")
	}
	e.Select("wall-a")
	e.overlay.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if e.primary != "wall" {
		t.Fatalf("escape selected %q, want wall", e.primary)
	}
}

func TestHitTestAndOrderedDrop(t *testing.T) {
	e, _ := newTestEditor(t)
	r := viewRect(t, e, "door")
	if id := e.hitTest(fyne.NewPos(r.X+5, r.Y+5)); id != "door" {
		t.Fatalf("hit test = %q", id)
	}
	// A plain drag over another container must NOT reparent it.
	e.SetGrid(0)
	e.Select("title")
	tr := viewRect(t, e, "title")
	o := e.overlay
	o.begin(fyne.NewPos(tr.X+tr.W/2, tr.Y+tr.H/2))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: -200, DY: 500},
	})
	o.DragEnd()
	wall, _ := e.doc.Find("wall")
	if len(wall.Children) != 2 {
		t.Fatalf("plain drag reparented: wall children = %d, want 2", len(wall.Children))
	}
	e.undo() // restore the title's position for the second drag

	// Holding Alt makes the same drag reparent into the ordered box.
	e.modsOverride = fyne.KeyModifierAlt
	e.Select("title")
	tr = viewRect(t, e, "title")
	o.begin(fyne.NewPos(tr.X+tr.W/2, tr.Y+tr.H/2))
	o.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: o.press},
		Dragged:    fyne.Delta{DX: -200, DY: 500},
	})
	o.DragEnd()
	e.modsOverride = 0
	wall, _ = e.doc.Find("wall")
	if len(wall.Children) != 3 {
		t.Fatalf("alt-drag wall children = %d, want 3", len(wall.Children))
	}
}
