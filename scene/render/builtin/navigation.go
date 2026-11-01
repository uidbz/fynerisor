package builtin

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

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

	rect *canvas.Rectangle
	text *canvas.Text
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
	if h.rect == nil {
		h.rect = canvas.NewRectangle(h.fill)
		h.text = canvas.NewText(h.label, render.Color("theme:foreground", nil))
		h.text.Alignment = fyne.TextAlignCenter
	}
	h.refreshPaint()
}

func (h *hotspot) refreshPaint() {
	if h.rect == nil {
		return
	}
	switch {
	case h.hovered:
		h.rect.FillColor = h.hoverFill
		h.rect.StrokeColor = h.stroke
		h.rect.StrokeWidth = h.strokeWidth
	case h.edit:
		h.rect.FillColor = render.Color("#ffffff18", nil)
		h.rect.StrokeColor = h.stroke
		h.rect.StrokeWidth = h.strokeWidth
	case h.showInPlay:
		h.rect.FillColor = h.fill
		h.rect.StrokeColor = h.stroke
		h.rect.StrokeWidth = h.strokeWidth
	default:
		h.rect.FillColor = render.Color("#00000000", nil)
		h.rect.StrokeColor = render.Color("#00000000", nil)
		h.rect.StrokeWidth = 0
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

func (h *hotspot) Tapped(e *fyne.PointEvent) {
	h.bc.Emit(h.id, "tapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) DoubleTapped(e *fyne.PointEvent) {
	h.bc.Emit(h.id, "doubleTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) TappedSecondary(e *fyne.PointEvent) {
	h.bc.Emit(h.id, "secondaryTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}

func (h *hotspot) MouseIn(e *desktop.MouseEvent) {
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
	return []fyne.CanvasObject{r.h.rect, r.h.text}
}

func (r *hotspotRenderer) Destroy() {}

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

func newEmbed(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := widget.NewLabel("")
	apply := func(n *scene.Node) {
		src := bc.Str(n, "src", "")
		if src == "" {
			l.SetText("Embedded scene\n(no source)")
		} else {
			l.SetText("Embedded scene\n" + src)
		}
	}
	apply(n)
	rect := canvas.NewRectangle(render.Color("#00000000", nil))
	rect.StrokeColor = render.Color("#a05adc80", nil)
	rect.StrokeWidth = 1
	stack := container.NewStack(rect, container.NewCenter(l))
	return newLeaf(stack, func(n *scene.Node, _ scene.Fields) error {
		apply(n)
		l.Refresh()
		return nil
	}), nil
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
