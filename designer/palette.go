package designer

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
			btn := widget.NewButton(title, func() { p.ed.ArmTool(typ) })
			buttons = append(buttons, btn)
		}
		p.box.Add(container.NewGridWithColumns(2, buttons...))
	}
	p.box.Refresh()
}
