package render_test

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

type probeObject struct {
	widget.BaseWidget
	refreshes *int
}

func (p *probeObject) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (p *probeObject) Refresh() { *p.refreshes++ }

type probeLive struct{ obj *probeObject }

func (p *probeLive) Object() fyne.CanvasObject              { return p.obj }
func (p *probeLive) Update(*scene.Node, scene.Fields) error { return nil }
func (p *probeLive) Destroy()                               {}

func probeRegistry(refreshes *int) *render.Registry {
	r := renderbuiltin.NewRegistry()
	r.Register(scene.ComponentSpec{
		Type: "probe", Title: "probe", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 10, H: 10},
	}, func(*render.BuildContext, *scene.Node) (render.Live, error) {
		obj := &probeObject{refreshes: refreshes}
		obj.ExtendBaseWidget(obj)
		return &probeLive{obj: obj}, nil
	})
	return r
}

// TestPreviewDoesNotRefreshSiblings pins the drag hot path: moving one node
// must not refresh its siblings (a container refresh re-decodes every image in
// the tree, which is what made dragging stutter).
func TestPreviewDoesNotRefreshSiblings(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	refreshes := 0
	d := paintDoc()
	d.Root.Children = append(d.Root.Children,
		&scene.Node{ID: "probe", Type: "probe", Place: scene.Place{X: 60, Y: 60, W: 10, H: 10}})
	v := render.NewView(d, render.Options{Registry: probeRegistry(&refreshes)})
	showView(t, v, fyne.NewSize(100, 100))
	refreshes = 0

	for i := 0; i < 20; i++ {
		v.PreviewPlace("box", scene.Place{X: float64(10 + i), Y: 10, W: 40, H: 40})
	}
	if refreshes != 0 {
		t.Fatalf("sibling refreshed %d times during a drag preview", refreshes)
	}

	changes, err := d.Apply(scene.Tx{Ops: []scene.Op{
		scene.SetPlaceOp("box", d.Root.Children[0].Place, scene.Place{X: 5, Y: 5, W: 40, H: 40}),
	}})
	if err != nil {
		t.Fatal(err)
	}
	refreshes = 0
	if err := v.Apply(changes); err != nil {
		t.Fatal(err)
	}
	if refreshes != 0 {
		t.Fatalf("sibling refreshed %d times when one node moved", refreshes)
	}
}
