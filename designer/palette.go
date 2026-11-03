package designer

import (
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

type paletteView struct {
	ed     *Editor
	box    *fyne.Container
	scroll *container.Scroll
}

func newPalette(e *Editor) *paletteView {
	p := &paletteView{ed: e, box: container.NewVBox()}
	p.rebuild()
	p.scroll = container.NewVScroll(p.box)
	return p
}

func (p *paletteView) object() fyne.CanvasObject { return p.scroll }

func (p *paletteView) rebuild() {
	p.box.Objects = nil
	hint := widget.NewLabel("Click a component, then click or drag on the canvas.")
	hint.Wrapping = fyne.TextWrapWord
	p.box.Add(hint)
	p.box.Add(widget.NewSeparator())

	groups := make(map[string][]scene.ComponentSpec)
	var order []string
	for _, typ := range p.ed.opts.Registry.Types() {
		spec, ok := p.ed.opts.Registry.Spec(typ)
		if !ok {
			continue
		}
		cat := spec.Category
		if cat == "" {
			cat = "Other"
		}
		if _, ok := groups[cat]; !ok {
			order = append(order, cat)
		}
		groups[cat] = append(groups[cat], spec)
	}
	rank := map[string]int{"Layout": 0, "Display": 1, "Shapes": 2, "Input": 3, "Navigation": 4}
	sort.SliceStable(order, func(a, b int) bool {
		ra, oka := rank[order[a]]
		rb, okb := rank[order[b]]
		if !oka {
			ra = 99
		}
		if !okb {
			rb = 99
		}
		if ra == rb {
			return order[a] < order[b]
		}
		return ra < rb
	})
	for _, cat := range order {
		label := widget.NewLabel(cat)
		label.TextStyle.Bold = true
		p.box.Add(label)
		var buttons []fyne.CanvasObject
		for _, spec := range groups[cat] {
			typ := spec.Type
			title := spec.Title
			if title == "" {
				title = typ
			}
			buttons = append(buttons, newPaletteItem(p.ed, typ, title))
		}
		p.box.Add(container.NewGridWithColumns(2, buttons...))
	}
	p.box.Refresh()
}

// paletteItem is a draggable component chip. Clicking arms the tool; dragging
// starts a creation gesture on the canvas (Fyne has no cross-widget DnD).
type paletteItem struct {
	widget.BaseWidget
	ed    *Editor
	typ   string
	label string

	rect *canvas.Rectangle
	text *canvas.Text

	dragging bool
	last     fyne.Position
}

func newPaletteItem(e *Editor, typ, label string) *paletteItem {
	p := &paletteItem{ed: e, typ: typ, label: label}
	p.rect = canvas.NewRectangle(color.NRGBA{R: 59, G: 130, B: 246, A: 30})
	p.rect.StrokeColor = color.NRGBA{R: 59, G: 130, B: 246, A: 160}
	p.rect.StrokeWidth = 1
	p.rect.CornerRadius = 4
	p.text = canvas.NewText(label, theme.ForegroundColor())
	p.text.TextSize = theme.TextSize()
	p.ExtendBaseWidget(p)
	return p
}

func (p *paletteItem) CreateRenderer() fyne.WidgetRenderer {
	return &paletteRenderer{p: p, objs: []fyne.CanvasObject{p.rect, p.text}}
}

func (p *paletteItem) Tapped(*fyne.PointEvent) { p.ed.ArmTool(p.typ) }

func (p *paletteItem) Dragged(e *fyne.DragEvent) {
	if !p.dragging {
		p.dragging = true
	}
	// AbsolutePosition is the pointer: the item's own top-left plus deltas is
	// offset by however far into the chip the drag started.
	p.last = e.AbsolutePosition
	p.ed.paletteDrag(p.typ, p.last)
}

func (p *paletteItem) DragEnd() {
	if !p.dragging {
		return
	}
	p.dragging = false
	p.ed.paletteDrop(p.typ, p.last)
}

type paletteRenderer struct {
	p    *paletteItem
	objs []fyne.CanvasObject
}

func (r *paletteRenderer) Layout(size fyne.Size) {
	r.p.rect.Move(fyne.NewPos(0, 0))
	r.p.rect.Resize(size)
	ts := r.p.text.MinSize()
	r.p.text.Move(fyne.NewPos(theme.Padding(), (size.Height-ts.Height)/2))
	r.p.text.Resize(ts)
}
func (r *paletteRenderer) MinSize() fyne.Size {
	ts := r.p.text.MinSize()
	return fyne.NewSize(ts.Width+2*theme.Padding(), ts.Height+theme.Padding())
}
func (r *paletteRenderer) Refresh()                     { r.p.rect.Refresh(); r.p.text.Refresh() }
func (r *paletteRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *paletteRenderer) Destroy()                     {}
