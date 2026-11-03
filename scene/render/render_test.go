package render_test

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
	tieassets "github.com/uidbz/fynerisor/scene/render/tie"
)

func freeDoc() *scene.Document {
	return &scene.Document{
		Title: "test",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 1000.0, "designH": 500.0, "scale": "fit"},
			Children: []*scene.Node{
				{ID: "door", Type: "hotspot", Place: scene.Place{X: 100, Y: 50, W: 200, H: 100}},
				{ID: "title", Type: "label", Place: scene.Place{X: 20, Y: 20, W: 300, H: 40},
					Props: map[string]any{"text": "Hello"}},
				{ID: "box", Type: "vbox", Place: scene.Place{X: 400, Y: 100, W: 300, H: 200},
					Children: []*scene.Node{
						{ID: "box-a", Type: "label", Props: map[string]any{"text": "A"}},
						{ID: "box-b", Type: "label", Props: map[string]any{"text": "B"}},
					}},
			},
		},
	}
}

func newView(t *testing.T, d *scene.Document) *render.View {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry()})
	w := test.NewWindow(v)
	w.Resize(fyne.NewSize(500, 250))
	v.Resize(fyne.NewSize(500, 250))
	return v
}

func TestViewRectsScale(t *testing.T) {
	d := freeDoc()
	v := newView(t, d)
	v.RefreshRects()
	r, ok := v.RectOf("door")
	if !ok {
		t.Fatal("door has no rect")
	}
	// fit: 500/1000 = 0.5, 250/500 = 0.5.
	if r.X != 50 || r.Y != 25 || r.W != 100 || r.H != 50 {
		t.Fatalf("door rect = %+v, want {50 25 100 50}", r)
	}
	r, _ = v.RectOf("title")
	if r.X != 10 || r.Y != 10 || r.W != 150 || r.H != 20 {
		t.Fatalf("title rect = %+v", r)
	}
	// Ordered child geometry comes from Fyne's absolute positions.
	r, ok = v.RectOf("box-a")
	if !ok || r.W <= 0 || r.H <= 0 {
		t.Fatalf("box-a rect = %+v ok=%v", r, ok)
	}
}

func TestViewApplyIncremental(t *testing.T) {
	d := freeDoc()
	v := newView(t, d)
	v.RefreshRects()

	door, _ := d.Find("door")
	tx := scene.Tx{Label: "move", Ops: []scene.Op{
		scene.SetPlaceOp("door", door.Place, scene.Place{X: 300, Y: 100, W: 200, H: 100}),
	}}
	changes, err := d.Apply(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	v.RefreshRects()
	r, _ := v.RectOf("door")
	if r.X != 150 || r.Y != 50 {
		t.Fatalf("door rect after move = %+v", r)
	}

	// Insert a child into the free canvas.
	child := &scene.Node{ID: "new", Type: "label", Place: scene.Place{X: 500, Y: 400, W: 100, H: 50}}
	changes, err = d.Apply(scene.Tx{Ops: []scene.Op{scene.InsertOp("root", 0, child)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	v.RefreshRects()
	if r, ok := v.RectOf("new"); !ok || r.X != 250 || r.Y != 200 {
		t.Fatalf("inserted rect = %+v ok=%v", r, ok)
	}

	// Insert into an ordered container.
	item := &scene.Node{ID: "box-c", Type: "label", Props: map[string]any{"text": "C"}}
	changes, err = d.Apply(scene.Tx{Ops: []scene.Op{scene.InsertOp("box", 1, item)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	if _, ok := v.RectOf("box-c"); !ok {
		t.Fatal("box-c has no rect")
	}

	// Remove and undo.
	rm, err := scene.RemoveOp(d, "new")
	if err != nil {
		t.Fatal(err)
	}
	changes, err = d.Apply(scene.Tx{Ops: []scene.Op{rm}})
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	if _, ok := v.RectOf("new"); ok {
		t.Fatal("removed node still has a rect")
	}
	changes, err = d.Apply(scene.Tx{Ops: []scene.Op{rm.Invert()}})
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	if _, ok := v.RectOf("new"); !ok {
		t.Fatal("undo did not restore the node")
	}
}

type nav struct {
	refs []string
	back int
}

func (n *nav) Navigate(ref string) { n.refs = append(n.refs, ref) }
func (n *nav) Back()               { n.back++ }

type caller struct {
	handler string
	evt     map[string]any
}

func (c *caller) Call(handler string, evt map[string]any) bool {
	c.handler = handler
	c.evt = evt
	return true
}

func TestActions(t *testing.T) {
	d := freeDoc()
	d.Root.Children[0].On = map[string][]scene.Action{
		"tapped": {
			{Do: "navigate", To: "kitchen"},
			{Do: "call", Handler: "hello"},
			{Do: "hide", Node: "title"},
			{Do: "set", Node: "title", Prop: "text", Value: "Changed"},
		},
	}
	n := &nav{}
	c := &caller{}
	v := render.NewView(d, render.Options{
		Registry: renderbuiltin.NewRegistry(), Navigator: n, Caller: c, Base: "file:///srv/scenes/house.json",
	})
	v.Resize(fyne.NewSize(500, 250))
	if !v.IsVisible("title") {
		t.Fatal("title should start visible")
	}
	v.EmitEvent("door", "tapped", nil)
	if len(n.refs) != 1 || n.refs[0] != "file:///srv/scenes/kitchen" {
		t.Fatalf("navigate refs = %#v", n.refs)
	}
	if c.handler != "hello" {
		t.Fatalf("caller handler = %q", c.handler)
	}
	for _, key := range []string{"id", "name", "type", "event", "x", "y", "item", "value"} {
		if _, ok := c.evt[key]; !ok {
			t.Errorf("event map missing %q: %#v", key, c.evt)
		}
	}
	if c.evt["id"] != "door" || c.evt["event"] != "tapped" {
		t.Fatalf("event map = %#v", c.evt)
	}
	if v.IsVisible("title") {
		t.Fatal("hide action did not hide the title")
	}
	if !v.RuntimeSet("title", "text", "Changed") {
		t.Fatal("runtime set failed")
	}
}

func TestListenersRunAfterActions(t *testing.T) {
	d := freeDoc()
	d.Root.Children[0].On = map[string][]scene.Action{
		"tapped": {{Do: "hide", Node: "title"}},
	}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry()})
	var order []string
	v.On("door", "tapped", func(evt map[string]any) {
		order = append(order, "listener")
		if v.IsVisible("title") {
			order = append(order, "visible")
		}
	})
	v.On("door", "tapped", func(map[string]any) { order = append(order, "listener2") })
	v.EmitEvent("door", "tapped", nil)
	if len(order) != 2 || order[0] != "listener" || order[1] != "listener2" {
		t.Fatalf("listener order = %#v", order)
	}
}

type opener struct {
	played []string
	opened []string
}

func (o *opener) Open(ref string) { o.opened = append(o.opened, ref) }
func (o *opener) Play(ref string) { o.played = append(o.played, ref) }
func (o *opener) Pause()          {}

// tie.audio and tie.video work without hand-wired actions.
func TestTieMediaDefaultActions(t *testing.T) {
	reg := renderbuiltin.NewRegistry()
	tieassets.Register(reg, nil)
	d := &scene.Document{Root: &scene.Node{ID: "root", Type: "free", Children: []*scene.Node{
		{ID: "track", Type: "tie.audio", Props: map[string]any{"subject": "abc"}},
		{ID: "clip", Type: "tie.video", Props: map[string]any{"subject": "def"}},
	}}}
	o := &opener{}
	v := render.NewView(d, render.Options{Registry: reg, Opener: o})
	v.EmitEvent("track", "tapped", map[string]any{"item": "abc"})
	v.EmitEvent("clip", "tapped", map[string]any{"item": "def"})
	if len(o.played) != 1 || o.played[0] != "abc" {
		t.Fatalf("played = %v, want [abc]", o.played)
	}
	if len(o.opened) != 1 || o.opened[0] != "def" {
		t.Fatalf("opened = %v, want [def]", o.opened)
	}
}

func TestLinkFallbackAndUnknownType(t *testing.T) {
	d := &scene.Document{Root: &scene.Node{ID: "root", Type: "free", Children: []*scene.Node{
		{ID: "go", Type: "link", Props: map[string]any{"text": "Go", "to": "next"}},
		{ID: "future", Type: "future.widget", Place: scene.Place{W: 100, H: 50}},
	}}}
	n := &nav{}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry(), Navigator: n, Base: "https://example.com/a.json"})
	if _, ok := v.LiveObject("future"); !ok {
		t.Fatal("unknown component has no live placeholder")
	}
	v.EmitEvent("go", "tapped", nil)
	if len(n.refs) != 1 || n.refs[0] != "https://example.com/next" {
		t.Fatalf("link fallback refs = %#v", n.refs)
	}
}

func TestModeRebuildShaderSlot(t *testing.T) {
	d := &scene.Document{Root: &scene.Node{ID: "root", Type: "free", Children: []*scene.Node{
		{ID: "slot", Type: "slot", Place: scene.Place{W: 100, H: 100}},
	}}}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry(), Mode: render.ModeEdit})
	if obj, ok := v.LiveObject("slot"); !ok || !obj.Visible() {
		t.Fatal("slot placeholder should be visible in edit mode")
	}
	v.SetMode(render.ModePlay)
	if obj, ok := v.LiveObject("slot"); !ok || obj.Visible() {
		t.Fatal("slot placeholder should be hidden in play mode")
	}
}
