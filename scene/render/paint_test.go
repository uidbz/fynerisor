package render_test

import (
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// paintDoc is a small free canvas with a solid red rectangle at a known place.
func paintDoc() *scene.Document {
	return &scene.Document{
		Title: "paint",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 100.0, "designH": 100.0, "scale": "none"},
			Children: []*scene.Node{
				{ID: "box", Type: "rect", Place: scene.Place{X: 10, Y: 10, W: 40, H: 40},
					Props: map[string]any{"fill": "#ff0000"}},
			},
		},
	}
}

func showView(t *testing.T, v *render.View, size fyne.Size) fyne.Canvas {
	t.Helper()
	w := test.NewWindow(v)
	w.Resize(size)
	v.Resize(size)
	return w.Canvas()
}

func capture(t *testing.T, c fyne.Canvas) image.Image {
	t.Helper()
	capable, ok := c.(interface{ Capture() image.Image })
	if !ok {
		t.Skip("test canvas has no Capture")
	}
	return capable.Capture()
}

// TestViewPaintsNodes is the regression test for the hint wrapper hiding the
// whole scene tree from Fyne's render walk (nothing was painted at all).
func TestViewPaintsNodes(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := render.NewView(paintDoc(), render.Options{Registry: renderbuiltin.NewRegistry()})
	img := capture(t, showView(t, v, fyne.NewSize(100, 100)))
	r, g, b, _ := img.At(30, 30).RGBA()
	if r < 0xc000 || g > 0x4000 || b > 0x4000 {
		t.Fatalf("pixel at node centre = #%04x%04x%04x, want red", r, g, b)
	}
}

// TestDesignBackground: an explicit backgroundColor fills the design area
// (empty keeps the historic letterbox-through look).
func TestDesignBackground(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	d := paintDoc()
	d.Root.Props["backgroundColor"] = "#00ff00"
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry()})
	img := capture(t, showView(t, v, fyne.NewSize(100, 100)))
	r, g, b, _ := img.At(5, 5).RGBA()
	if g < 0xc000 || r > 0x4000 || b > 0x4000 {
		t.Fatalf("design background = #%04x%04x%04x, want green", r, g, b)
	}
}

// TestViewTextPaints covers the scaledText component, which was also invisible.
func TestViewTextPaints(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	d := paintDoc()
	d.Root.Children = []*scene.Node{{
		ID: "title", Type: "text", Place: scene.Place{X: 10, Y: 10, W: 80, H: 40},
		Props: map[string]any{"text": "Hello", "size": 30.0, "color": "#ffffff"},
	}}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry()})
	img := capture(t, showView(t, v, fyne.NewSize(100, 100)))
	for y := 10; y < 50; y++ {
		for x := 10; x < 90; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r > 0x8000 && g > 0x8000 && b > 0x8000 {
				return
			}
		}
	}
	t.Fatal("canvas text component painted no pixels")
}

// TestViewHitTestsNodes proves nodes are reachable through the Fyne tree walk
// (taps, hover, popups and absolute positioning all depend on it).
func TestViewHitTestsNodes(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	d := paintDoc()
	d.Root.Children[0].On = map[string][]scene.Action{
		"tapped": {{Do: "navigate", To: "next"}},
	}
	n := &nav{}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry(), Navigator: n})
	canvas := showView(t, v, fyne.NewSize(100, 100))
	test.TapCanvas(canvas, fyne.NewPos(30, 30))
	if len(n.refs) != 1 || n.refs[0] != "next" {
		t.Fatalf("tap did not reach the node: refs=%v", n.refs)
	}
}

// TestOrderedChildPosition checks that a child of a layout container gets its
// absolute position from the canvas (it used to be 0,0 because the walk stopped
// at the hint wrapper), which the designer uses for selection boxes.
func TestOrderedChildPosition(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	d := &scene.Document{Root: &scene.Node{
		ID: "root", Type: "free",
		Props: map[string]any{"designW": 200.0, "designH": 200.0, "scale": "none"},
		Children: []*scene.Node{{
			ID: "box", Type: "vbox", Place: scene.Place{X: 20, Y: 30, W: 120, H: 80},
			Children: []*scene.Node{
				{ID: "a", Type: "label", Props: map[string]any{"text": "A"}},
				{ID: "b", Type: "label", Props: map[string]any{"text": "B"}},
			},
		}},
	}}
	v := render.NewView(d, render.Options{Registry: renderbuiltin.NewRegistry()})
	showView(t, v, fyne.NewSize(200, 200))
	v.RefreshRects()
	rb, ok := v.RectOf("b")
	if !ok {
		t.Fatal("b has no rect")
	}
	if rb.X < 20 || rb.Y <= 30 {
		t.Fatalf("ordered child rect = %+v, want inside the vbox at x>=20 y>30", rb)
	}
}
