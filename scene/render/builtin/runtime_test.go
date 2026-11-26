package builtin

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

func runtimeDoc() *scene.Document {
	return &scene.Document{
		Title: "runtime",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 800.0, "designH": 600.0, "scale": "none"},
			Children: []*scene.Node{
				{ID: "e", Type: "entry", Place: scene.Place{X: 10, Y: 10, W: 200, H: 40},
					Props: map[string]any{"text": "${a}", "placeholder": "${b}"}},
				{ID: "lbl", Type: "label", Place: scene.Place{X: 10, Y: 60, W: 200, H: 40},
					Props: map[string]any{"text": "doc"}},
			},
		},
	}
}

func newRuntimeView(t *testing.T) *render.View {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	v := render.NewView(runtimeDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(800, 600))
	return v
}

func liveEntry(t *testing.T, v *render.View, id string) *widget.Entry {
	t.Helper()
	obj, ok := v.LiveObject(id)
	if !ok {
		t.Fatalf("no live %s", id)
	}
	e, ok := obj.(*widget.Entry)
	if !ok {
		t.Fatalf("%s is %T", id, obj)
	}
	return e
}

func liveLabel(t *testing.T, v *render.View, id string) *widget.Label {
	t.Helper()
	obj, _ := v.LiveObject(id)
	l, ok := obj.(*widget.Label)
	if !ok {
		t.Fatalf("%s is %T", id, obj)
	}
	return l
}

// Two runtime props on one node must not clobber each other: the second
// RuntimeSet used to re-apply the document's raw template for the first
// (#63.1).
func TestRuntimeSetPropsDoNotClobber(t *testing.T) {
	for _, order := range [][2]string{{"text", "placeholder"}, {"placeholder", "text"}} {
		v := newRuntimeView(t)
		vals := map[string]string{"text": "A", "placeholder": "B"}
		for _, p := range order {
			if !v.RuntimeSet("e", p, vals[p]) {
				t.Fatalf("RuntimeSet(%s) failed", p)
			}
		}
		e := liveEntry(t, v, "e")
		if e.Text != "A" || e.PlaceHolder != "B" {
			t.Fatalf("order %v: text=%q placeholder=%q, want A/B", order, e.Text, e.PlaceHolder)
		}
	}
}

// A runtime prop set earlier (node.Set) survives later runtime updates of
// other props on the same node (#63.1).
func TestRuntimeSetKeepsEarlierRuntimeProps(t *testing.T) {
	v := newRuntimeView(t)
	v.RuntimeSet("lbl", "bold", true)
	v.RuntimeSet("lbl", "text", "bound")
	l := liveLabel(t, v, "lbl")
	if !l.TextStyle.Bold || l.Text != "bound" {
		t.Fatalf("label bold=%v text=%q, want true/bound", l.TextStyle.Bold, l.Text)
	}
}

// ErrRebuild props rebuild from the runtime values, not the document: the
// new value takes effect and the node's other runtime props stay (#63.2).
func TestRuntimeRebuildKeepsRuntimeValues(t *testing.T) {
	v := newRuntimeView(t)
	v.RuntimeSet("e", "text", "hello")
	if !v.RuntimeSet("e", "password", true) {
		t.Fatal("RuntimeSet(password) failed")
	}
	e := liveEntry(t, v, "e")
	if !e.Password {
		t.Fatal("password ignored after rebuild")
	}
	if e.Text != "hello" {
		t.Fatalf("rebuilt entry text = %q, want hello", e.Text)
	}
	// A full rebuild (mode switch) keeps runtime values too.
	v.SetMode(render.ModeEdit)
	if e := liveEntry(t, v, "e"); !e.Password || e.Text != "hello" {
		t.Fatalf("after SetMode: password=%v text=%q", e.Password, e.Text)
	}
}

// Updating another prop must not replace what the user typed (a bound
// placeholder refreshes after every keystroke handler).
func TestRuntimeSetKeepsTypedText(t *testing.T) {
	v := newRuntimeView(t)
	e := liveEntry(t, v, "e")
	e.SetText("typed")
	v.RuntimeSet("e", "placeholder", "hint")
	if e.Text != "typed" {
		t.Fatalf("typed text reset to %q", e.Text)
	}
	if e.PlaceHolder != "hint" {
		t.Fatalf("placeholder = %q", e.PlaceHolder)
	}
}

// A document edit of a prop replaces its runtime value; edits of other props
// keep it.
func TestApplyDocumentEditVersusRuntimeProps(t *testing.T) {
	v := newRuntimeView(t)
	v.RuntimeSet("lbl", "text", "runtime")
	apply := func(key string, val any) {
		t.Helper()
		changes, err := v.Doc().Apply(scene.Tx{Ops: []scene.Op{scene.SetPropOp("lbl", key, nil, val)}})
		if err != nil {
			t.Fatal(err)
		}
		if err := v.Apply(changes); err != nil {
			t.Fatal(err)
		}
	}
	apply("bold", true)
	if l := liveLabel(t, v, "lbl"); l.Text != "runtime" || !l.TextStyle.Bold {
		t.Fatalf("after bold edit: text=%q bold=%v", l.Text, l.TextStyle.Bold)
	}
	apply("text", "edited")
	if l := liveLabel(t, v, "lbl"); l.Text != "edited" {
		t.Fatalf("after text edit: text=%q, want edited", l.Text)
	}
}
