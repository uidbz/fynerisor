package builtin

import (
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

type baseLive struct {
	obj    fyne.CanvasObject
	update func(n *scene.Node, changed scene.Fields) error
}

func (l *baseLive) Object() fyne.CanvasObject { return l.obj }

func (l *baseLive) Update(n *scene.Node, changed scene.Fields) error {
	if l.update == nil {
		return nil
	}
	return l.update(n, changed)
}

func (l *baseLive) Destroy() {}

func newLeaf(obj fyne.CanvasObject, update func(*scene.Node, scene.Fields) error) render.Live {
	return &baseLive{obj: obj, update: update}
}

// resolveAsset turns a document-relative reference into a resolver reference.
func resolveAsset(bc *render.BuildContext, ref string) string {
	if bc.Assets == nil || ref == "" {
		return ref
	}
	if resolved, err := bc.Assets.Resolve(bc.Base, ref); err == nil && resolved != "" {
		return resolved
	}
	return ref
}

// loadAsset loads a resource off the UI thread and applies it only while gen is
// still current. Components bump gen on every source change and on Destroy.
func loadAsset(bc *render.BuildContext, ref string, gen *atomic.Int64, apply func(fyne.Resource, error)) {
	n := gen.Add(1)
	if ref == "" {
		apply(nil, nil)
		return
	}
	if bc.Assets == nil {
		bc.Statusf("no asset resolver for %q", ref)
		return
	}
	ref = resolveAsset(bc, ref)
	go func() {
		res, err := bc.Assets.Open(bc.Context, ref)
		guithread.Do(func() {
			if gen.Load() != n {
				return
			}
			apply(res, err)
		})
	}()
}

// tapSurface is a transparent widget that turns taps and hover into scene
// events. It sits above a visual object in a Stack so non-interactive visuals
// (images, shapes) can still have declarative actions.
type tapSurface struct {
	widget.BaseWidget
	bc *render.BuildContext
	id string
}

func newTapSurface(bc *render.BuildContext, id string) *tapSurface {
	t := &tapSurface{bc: bc, id: id}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapSurface) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(render.Color("#00000000", nil)))
}

func (t *tapSurface) Tapped(e *fyne.PointEvent) {
	t.bc.Emit(t.id, "tapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (t *tapSurface) DoubleTapped(e *fyne.PointEvent) {
	t.bc.Emit(t.id, "doubleTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (t *tapSurface) TappedSecondary(e *fyne.PointEvent) {
	t.bc.Emit(t.id, "secondaryTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (t *tapSurface) MouseIn(e *desktop.MouseEvent) {
	t.bc.Emit(t.id, "hoverIn", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (t *tapSurface) MouseOut() {
	t.bc.Emit(t.id, "hoverOut", nil)
}
