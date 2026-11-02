package builtin

import (
	"image/color"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// ------------------------------------------------------------------ hotspot

// hotspot is the point-and-click primitive: a clickable region over a
// background. In edit mode it shows a faint outline; in play mode it is
// invisible unless showInPlay is set or the pointer hovers it.
type hotspot struct {
	widget.BaseWidget
	bc          *render.BuildContext
	id          string
	edit        bool
	shape       string
	label       string
	fill        color.Color
	hoverFill   color.Color
	stroke      color.Color
	strokeWidth float32
	showInPlay  bool
	hovered     bool
	points      []fyne.Position

	rect    *canvas.Rectangle
	polygon *canvas.ArbitraryPolygon
	text    *canvas.Text
}

func newHotspot(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	h := &hotspot{bc: bc, id: n.ID, edit: bc.Mode == render.ModeEdit}
	h.apply(n)
	h.ExtendBaseWidget(h)
	return h, nil
}

func (h *hotspot) apply(n *scene.Node) {
	h.shape = h.bc.Str(n, "shape", "rect")
	h.label = h.bc.Str(n, "label", "")
	h.fill = h.bc.ColorProp(n, "fill", render.Color("#ffffff00", nil))
	h.hoverFill = h.bc.ColorProp(n, "hoverFill", render.Color("#ffffff33", nil))
	h.stroke = h.bc.ColorProp(n, "stroke", render.Color("#ffffffaa", nil))
	h.strokeWidth = float32(h.bc.Num(n, "strokeWidth", 1))
	h.showInPlay = h.bc.Bool(n, "showInPlay", false)
	h.points = polygonPoints(n)
	if h.rect == nil {
		h.rect = canvas.NewRectangle(h.fill)
		h.polygon = &canvas.ArbitraryPolygon{NormalizedPoints: true}
		h.text = canvas.NewText(h.label, render.Color("theme:foreground", nil))
		h.text.Alignment = fyne.TextAlignCenter
	}
	h.refreshPaint()
}

func (h *hotspot) refreshPaint() {
	if h.rect == nil {
		return
	}
	fill, stroke, strokeWidth := render.Color("#00000000", nil), render.Color("#00000000", nil), float32(0)
	switch {
	case h.hovered:
		fill, stroke, strokeWidth = h.hoverFill, h.stroke, h.strokeWidth
	case h.edit:
		fill, stroke, strokeWidth = render.Color("#ffffff18", nil), h.stroke, h.strokeWidth
	case h.showInPlay:
		fill, stroke, strokeWidth = h.fill, h.stroke, h.strokeWidth
	}
	h.rect.FillColor = fill
	h.rect.StrokeColor = stroke
	h.rect.StrokeWidth = strokeWidth
	h.rect.Hidden = h.shape == "polygon"
	if h.polygon != nil {
		h.polygon.Points = h.points
		h.polygon.FillColor = fill
		h.polygon.StrokeColor = stroke
		h.polygon.StrokeWidth = strokeWidth
		h.polygon.Hidden = h.shape != "polygon"
		h.polygon.Refresh()
	}
	h.text.Text = h.label
	if h.edit && h.label != "" {
		h.text.Show()
	} else {
		h.text.Hide()
	}
	h.rect.Refresh()
	h.text.Refresh()
}

func (h *hotspot) Object() fyne.CanvasObject { return h }

func (h *hotspot) Update(n *scene.Node, _ scene.Fields) error {
	h.apply(n)
	return nil
}

func (h *hotspot) Destroy() {}

func (h *hotspot) hit(pos fyne.Position) bool {
	if h.shape != "polygon" || len(h.points) < 3 {
		return true
	}
	size := h.Size()
	if size.Width <= 0 || size.Height <= 0 {
		return true
	}
	px := float64(pos.X) / float64(size.Width)
	py := float64(pos.Y) / float64(size.Height)
	inside := false
	j := len(h.points) - 1
	for i := 0; i < len(h.points); i++ {
		xi, yi := float64(h.points[i].X), float64(h.points[i].Y)
		xj, yj := float64(h.points[j].X), float64(h.points[j].Y)
		if (yi > py) != (yj > py) && px < (xj-xi)*(py-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}

func (h *hotspot) Tapped(e *fyne.PointEvent) {
	if !h.hit(e.Position) {
		return
	}
	h.bc.Emit(h.id, "tapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) DoubleTapped(e *fyne.PointEvent) {
	if !h.hit(e.Position) {
		return
	}
	h.bc.Emit(h.id, "doubleTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) TappedSecondary(e *fyne.PointEvent) {
	if !h.hit(e.Position) {
		return
	}
	h.bc.Emit(h.id, "secondaryTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) MouseIn(e *desktop.MouseEvent) {
	if !h.hit(e.Position) {
		return
	}
	if !h.edit {
		h.hovered = true
		h.refreshPaint()
	}
	h.bc.Emit(h.id, "hoverIn", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) MouseOut() {
	if h.hovered {
		h.hovered = false
		h.refreshPaint()
	}
	h.bc.Emit(h.id, "hoverOut", nil)
}

func (h *hotspot) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (h *hotspot) CreateRenderer() fyne.WidgetRenderer {
	h.ExtendBaseWidget(h)
	if h.rect == nil {
		h.apply(&scene.Node{ID: h.id, Type: "hotspot"})
	}
	return &hotspotRenderer{h: h}
}

type hotspotRenderer struct {
	h    *hotspot
	objs []fyne.CanvasObject
}

func (r *hotspotRenderer) Layout(size fyne.Size) {
	r.h.rect.Move(fyne.NewPos(0, 0))
	r.h.rect.Resize(size)
	if r.h.polygon != nil {
		r.h.polygon.Move(fyne.NewPos(0, 0))
		r.h.polygon.Resize(size)
	}
	if r.h.shape == "ellipse" {
		rad := size.Width
		if size.Height < rad {
			rad = size.Height
		}
		r.h.rect.CornerRadius = rad / 2
	} else {
		r.h.rect.CornerRadius = 0
	}
	ts := r.h.text.MinSize()
	r.h.text.Move(fyne.NewPos((size.Width-ts.Width)/2, (size.Height-ts.Height)/2))
	r.h.text.Resize(ts)
}

func (r *hotspotRenderer) MinSize() fyne.Size { return fyne.NewSize(1, 1) }

func (r *hotspotRenderer) Refresh() {
	r.h.refreshPaint()
	r.h.rect.Refresh()
}

func (r *hotspotRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.h.rect, r.h.polygon, r.h.text}
}

func (r *hotspotRenderer) Destroy() {}

// polygonPoints reads the normalized polygon points property.
func polygonPoints(n *scene.Node) []fyne.Position {
	v, ok := n.Prop("points")
	if !ok {
		return nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]fyne.Position, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		x, _ := m["x"].(float64)
		y, _ := m["y"].(float64)
		out = append(out, fyne.NewPos(float32(x), float32(y)))
	}
	return out
}

// --------------------------------------------------------------------- link

func newLink(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	btn := widget.NewButton(bc.Str(n, "text", "Open"), func() {
		bc.Emit(n.ID, "tapped", nil)
	})
	btn.Importance = widget.LowImportance
	return newLeaf(btn, func(n *scene.Node, _ scene.Fields) error {
		btn.SetText(bc.Str(n, "text", "Open"))
		btn.Refresh()
		return nil
	}), nil
}

// -------------------------------------------------------------------- embed

type embedLive struct {
	stack *fyne.Container
	label *widget.Label
	view  *render.View
	bc    *render.BuildContext
	ref   string
	gen   atomic.Int64
}

func newEmbed(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := widget.NewLabel("Embedded scene")
	l.Alignment = fyne.TextAlignCenter
	stack := container.NewStack(l)
	l2 := &embedLive{stack: stack, label: l, bc: bc}
	l2.apply(n)
	return l2, nil
}

func (l *embedLive) apply(n *scene.Node) {
	src := l.bc.Str(n, "src", "")
	if src == l.ref && (src == "" || l.view != nil) {
		return
	}
	l.ref = src
	l.gen.Add(1)
	if l.view != nil {
		l.view.Destroy()
		l.view = nil
	}
	if src == "" {
		l.label.SetText("Embedded scene\n(no source)")
		l.label.Show()
		l.stack.Objects = []fyne.CanvasObject{l.label}
		l.stack.Refresh()
		return
	}
	l.label.SetText("Loading embedded scene…")
	l.label.Show()
	l.stack.Objects = []fyne.CanvasObject{l.label}
	l.stack.Refresh()
	if l.bc.Embed == nil {
		l.label.SetText("Embedded scene\n(no resolver)")
		return
	}
	gen := l.gen.Load()
	go func() {
		doc, base, err := l.bc.Embed.LoadEmbed(l.bc.Context, l.bc.Base, src)
		guithread.Do(func() {
			if l.gen.Load() != gen {
				return
			}
			if err != nil {
				l.label.SetText("Embedded scene\n" + err.Error())
				l.label.Show()
				l.stack.Objects = []fyne.CanvasObject{l.label}
				l.stack.Refresh()
				return
			}
			l.view = l.bc.View.Embedded(doc, base)
			l.stack.Objects = []fyne.CanvasObject{l.view}
			l.stack.Refresh()
		})
	}()
}

func (l *embedLive) Object() fyne.CanvasObject { return l.stack }

func (l *embedLive) Update(n *scene.Node, _ scene.Fields) error {
	l.apply(n)
	return nil
}

func (l *embedLive) Destroy() {
	l.gen.Add(1)
	if l.view != nil {
		l.view.Destroy()
	}
}

// --------------------------------------------------------------------- slot

func newSlot(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	label := widget.NewLabel("")
	label.Alignment = fyne.TextAlignCenter
	apply := func(n *scene.Node) {
		label.SetText(bc.Str(n, "hint", "Script fills this area"))
	}
	apply(n)
	rect := canvas.NewRectangle(render.Color("#00000000", nil))
	rect.StrokeColor = render.Color("#22c55e80", nil)
	rect.StrokeWidth = 1
	box := container.NewStack(rect, container.NewCenter(label))
	box.Hidden = bc.Mode != render.ModeEdit
	l := &slotLive{box: box, rect: rect, label: label, bc: bc}
	return l, nil
}

type slotLive struct {
	box     *fyne.Container
	rect    *canvas.Rectangle
	label   *widget.Label
	content fyne.CanvasObject
	bc      *render.BuildContext
}

func (s *slotLive) Object() fyne.CanvasObject { return s.box }

func (s *slotLive) Update(n *scene.Node, _ scene.Fields) error {
	s.label.SetText(s.bc.Str(n, "hint", "Script fills this area"))
	if s.bc.Mode == render.ModeEdit {
		s.box.Show()
	} else {
		s.box.Hide()
	}
	return nil
}

// SetContent fills the slot at runtime (node.SetContent in a page script).
func (s *slotLive) SetContent(obj fyne.CanvasObject) {
	s.content = obj
	objs := []fyne.CanvasObject{}
	if obj != nil {
		objs = append(objs, obj)
	} else if s.bc.Mode == render.ModeEdit {
		objs = append(objs, s.rect, container.NewCenter(s.label))
	}
	s.box.Objects = objs
	s.box.Show()
	s.box.Refresh()
}

func (s *slotLive) Destroy() {}
