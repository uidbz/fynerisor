package designer

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/scene"
)

func TestInlineTextEditing(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("title")
	o := e.overlay
	o.startInline("title", "Hello")
	if o.inline == nil {
		t.Fatal("inline entry did not start")
	}
	o.inline.SetText("Edited inline")
	o.commitInline()
	if o.inline != nil {
		t.Fatal("inline entry did not clear")
	}
	n, _ := e.doc.Find("title")
	if n.StringProp("text", "") != "Edited inline" {
		t.Fatalf("text = %q", n.StringProp("text", ""))
	}
	e.undo()
	n, _ = e.doc.Find("title")
	if n.StringProp("text", "") != "Hello" {
		t.Fatalf("undo text = %q", n.StringProp("text", ""))
	}
}

func TestWrapAndUnwrapSelection(t *testing.T) {
	e, _ := newTestEditor(t)
	e.SelectMany([]string{"door", "title"})
	e.wrapSelection("vbox")
	root, _ := e.doc.Find("root")
	if len(root.Children) != 2 {
		t.Fatalf("root children = %d, want 2", len(root.Children))
	}
	wrapped, ok := e.doc.Find(e.primary)
	if !ok || wrapped.Type != "vbox" || len(wrapped.Children) != 2 {
		t.Fatalf("wrapped = %+v", wrapped)
	}
	e.unwrapSelection()
	root, _ = e.doc.Find("root")
	if len(root.Children) != 3 {
		t.Fatalf("root children after unwrap = %d, want 3", len(root.Children))
	}
	if _, ok := e.doc.Find(wrapped.ID); ok {
		t.Fatal("container still exists after unwrap")
	}
}

func TestConvertContainer(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("wall")
	if !e.parentFree("wall") {
		t.Fatal("wall should start in a free parent")
	}
	e.convertSelection("free")
	n, _ := e.doc.Find("wall")
	if n.Type != "free" {
		t.Fatalf("type = %q", n.Type)
	}
	if n.FloatProp("designW", 0) <= 0 {
		t.Fatalf("designW = %v", n.FloatProp("designW", 0))
	}
	for _, c := range n.Children {
		if c.Place.W <= 0 || c.Place.H <= 0 {
			t.Fatalf("child %s was not converted to x/y/w/h: %+v", c.ID, c.Place)
		}
	}
	e.convertSelection("vbox")
	n, _ = e.doc.Find("wall")
	if n.Type != "vbox" {
		t.Fatalf("type after convert back = %q", n.Type)
	}
}

func TestPolygonPoints(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Select("door")
	e.setProp("door", "shape", "polygon", "")
	e.setProp("door", "points", []any{
		map[string]any{"x": 0.0, "y": 0.0},
		map[string]any{"x": 1.0, "y": 0.0},
		map[string]any{"x": 0.5, "y": 1.0},
	}, "")
	n, _ := e.doc.Find("door")
	pts, _ := n.Prop("points")
	if len(pts.([]any)) != 3 {
		t.Fatalf("points = %#v", pts)
	}
	r, _ := e.view.RectOf("door")
	e.addPolygonPoint("door", fyne.NewPos(r.X+r.W/2, r.Y+r.H/4))
	n, _ = e.doc.Find("door")
	pts, _ = n.Prop("points")
	if len(pts.([]any)) != 4 {
		t.Fatalf("points after add = %d", len(pts.([]any)))
	}
	e.removePolygonPoint("door", 3)
	n, _ = e.doc.Find("door")
	pts, _ = n.Prop("points")
	if len(pts.([]any)) != 3 {
		t.Fatalf("points after remove = %d", len(pts.([]any)))
	}
}

func TestAutosaveDraft(t *testing.T) {
	e, _ := newTestEditor(t)
	e.opts.Ref = fmt.Sprintf("draft-%d.scene.json", time.Now().UnixNano())
	t.Cleanup(func() { removeDraft(e.opts.Ref) })
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)
	e.autosaveDraft()
	d, err := readDraft(e.opts.Ref)
	if err != nil {
		t.Fatalf("read draft: %v", err)
	}
	doc, err := scene.Decode(d.Data)
	if err != nil {
		t.Fatalf("draft decode: %v", err)
	}
	n, _ := doc.Find("door")
	if n.Place.X != 101 {
		t.Fatalf("draft place = %+v", n.Place)
	}
}

func TestStoryEdges(t *testing.T) {
	d := testDoc()
	d.Root.Children[0].On = map[string][]scene.Action{
		"tapped": {{Do: "navigate", To: "kitchen.scene.json"}},
	}
	d.Root.Children[1].On = map[string][]scene.Action{
		"tapped": {{Do: "navigate", To: "help.scene.json"}},
	}
	edges := storyEdges(d)
	if len(edges) != 2 || edges[0].To != "kitchen.scene.json" {
		t.Fatalf("edges = %#v", edges)
	}
}

func TestDeviceAndZoomPreview(t *testing.T) {
	e, _ := newTestEditor(t)
	e.device = "phone"
	e.zoom = 1.5
	e.rebuildStage()
	if len(e.center.Objects) != 1 {
		t.Fatalf("center objects = %d", len(e.center.Objects))
	}
	e.device = ""
	e.zoom = 1
	e.rebuildStage()
	if e.stage.Objects[0] != fyne.CanvasObject(e.view) {
		t.Fatalf("stage view object = %T", e.stage.Objects[0])
	}
}
