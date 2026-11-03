package builtin

import (
	"net/url"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// -------------------------------------------------------------------- label

func newLabel(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := widget.NewLabel("")
	apply := func(n *scene.Node) {
		l.Text = bc.Str(n, "text", "")
		l.TextStyle = fyne.TextStyle{
			Bold:      bc.Bool(n, "bold", false),
			Italic:    bc.Bool(n, "italic", false),
			Monospace: bc.Bool(n, "mono", false),
		}
		l.Alignment = render.TextAlign(bc.Str(n, "align", "left"))
		l.Wrapping = render.TextWrap(bc.Str(n, "wrap", "word"))
		l.Truncation = render.TextTruncate(bc.Str(n, "truncation", "off"))
		l.Refresh()
	}
	apply(n)
	return newLeaf(l, func(n *scene.Node, _ scene.Fields) error { apply(n); return nil }), nil
}

// --------------------------------------------------------------- canvas text

type scaledText struct {
	widget.BaseWidget
	obj   *canvas.Text
	base  float32
	scale float64
}

func newScaledText() *scaledText {
	t := &scaledText{obj: canvas.NewText("", render.Color("#ffffff", nil)), scale: 1}
	t.ExtendBaseWidget(t)
	return t
}

// CreateRenderer exposes the canvas text as the widget's only object. Like
// hintObject, this must be a real widget: the Fyne walkers and painters do not
// descend into plain CanvasObject wrappers.
func (t *scaledText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.obj)
}

func (t *scaledText) SetScale(sx, sy float64) {
	s := sx
	if sy < s {
		s = sy
	}
	if s <= 0 {
		s = 1
	}
	if s == t.scale {
		return // a layout pass must not re-shape unchanged text
	}
	t.scale = s
	t.obj.TextSize = t.base * float32(s)
	t.obj.Refresh()
}

func newText(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	t := newScaledText()
	apply := func(n *scene.Node) {
		t.obj.Text = bc.Str(n, "text", "")
		t.base = float32(bc.Num(n, "size", 18))
		t.obj.TextSize = t.base * float32(t.scale)
		t.obj.Color = bc.ColorProp(n, "color", render.Color("#ffffff", nil))
		t.obj.TextStyle = fyne.TextStyle{
			Bold:      bc.Bool(n, "bold", false),
			Italic:    bc.Bool(n, "italic", false),
			Monospace: bc.Bool(n, "mono", false),
		}
		t.obj.Alignment = render.TextAlign(bc.Str(n, "align", "left"))
		t.obj.Refresh()
	}
	apply(n)
	return newLeaf(t, func(n *scene.Node, _ scene.Fields) error { apply(n); return nil }), nil
}

// ----------------------------------------------------------------- markdown

func newMarkdown(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	rt := widget.NewRichTextFromMarkdown(bc.Str(n, "text", ""))
	return newLeaf(rt, func(n *scene.Node, _ scene.Fields) error {
		rt.ParseMarkdown(bc.Str(n, "text", ""))
		return nil
	}), nil
}

// -------------------------------------------------------------------- image

type imageLive struct {
	img   *canvas.Image
	stack *fyne.Container
	bc    *render.BuildContext
	gen   atomic.Int64
	ref   string
}

func newImage(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	img := canvas.NewImageFromResource(nil)
	img.FillMode = render.FillMode(bc.Str(n, "fill", "cover"))
	l := &imageLive{img: img, bc: bc}
	l.stack = container.NewStack(img, newTapSurface(bc, n.ID))
	l.apply(n)
	return l, nil
}

func (l *imageLive) apply(n *scene.Node) {
	l.img.FillMode = render.FillMode(l.bc.Str(n, "fill", "cover"))
	if src := l.bc.Str(n, "src", ""); src != l.ref {
		l.ref = src
		if src == "" {
			l.gen.Add(1)
			l.img.Resource = nil
			l.img.Refresh()
			return
		}
		loadAsset(l.bc, src, &l.gen, func(res fyne.Resource, err error) {
			if err != nil {
				l.bc.Statusf("image %s: %v", src, err)
				return
			}
			l.img.Resource = res
			l.img.Refresh()
		})
	}
}

func (l *imageLive) Object() fyne.CanvasObject { return l.stack }

func (l *imageLive) Update(n *scene.Node, _ scene.Fields) error {
	l.apply(n)
	return nil
}

func (l *imageLive) Destroy() { l.gen.Add(1) }

// --------------------------------------------------------------------- icon

func newIcon(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	icon := widget.NewIcon(render.IconResource(bc.Str(n, "name", "info")))
	stack := container.NewStack(icon, newTapSurface(bc, n.ID))
	return newLeaf(stack, func(n *scene.Node, _ scene.Fields) error {
		icon.SetResource(render.IconResource(bc.Str(n, "name", "info")))
		return nil
	}), nil
}

// ---------------------------------------------------------------- separator

func newSeparator(_ *render.BuildContext, _ *scene.Node) (render.Live, error) {
	return newLeaf(widget.NewSeparator(), nil), nil
}

// ----------------------------------------------------------------- progress

func newProgress(_ *render.BuildContext, _ *scene.Node) (render.Live, error) {
	p := widget.NewProgressBar()
	p.Min, p.Max = 0, 1
	return newLeaf(p, func(n *scene.Node, _ scene.Fields) error {
		v := n.FloatProp("value", 0.5)
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		p.SetValue(v)
		return nil
	}), nil
}

// ----------------------------------------------------------------- activity

func newActivity(_ *render.BuildContext, _ *scene.Node) (render.Live, error) {
	a := widget.NewActivity()
	a.Start()
	return newLeaf(a, nil), nil
}

// ------------------------------------------------------------------- shapes

func newRect(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	r := canvas.NewRectangle(render.Color("#3b82f6", nil))
	apply := func(n *scene.Node) {
		r.FillColor = bc.ColorProp(n, "fill", render.Color("#3b82f6", nil))
		r.StrokeColor = bc.ColorProp(n, "stroke", render.Color("#00000000", nil))
		r.StrokeWidth = float32(bc.Num(n, "strokeWidth", 0))
		r.CornerRadius = float32(bc.Num(n, "cornerRadius", 0))
		r.Refresh()
	}
	apply(n)
	stack := container.NewStack(r, newTapSurface(bc, n.ID))
	return newLeaf(stack, func(n *scene.Node, _ scene.Fields) error { apply(n); return nil }), nil
}

func newCircle(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	c := canvas.NewCircle(render.Color("#22c55e", nil))
	apply := func(n *scene.Node) {
		c.FillColor = bc.ColorProp(n, "fill", render.Color("#22c55e", nil))
		c.StrokeColor = bc.ColorProp(n, "stroke", render.Color("#00000000", nil))
		c.StrokeWidth = float32(bc.Num(n, "strokeWidth", 0))
		c.Refresh()
	}
	apply(n)
	stack := container.NewStack(c, newTapSurface(bc, n.ID))
	return newLeaf(stack, func(n *scene.Node, _ scene.Fields) error { apply(n); return nil }), nil
}

func newLine(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := canvas.NewLine(render.Color("#ffffff", nil))
	apply := func(n *scene.Node) {
		l.StrokeColor = bc.ColorProp(n, "color", render.Color("#ffffff", nil))
		l.StrokeWidth = float32(bc.Num(n, "width", 2))
		l.Refresh()
	}
	apply(n)
	return newLeaf(l, func(n *scene.Node, _ scene.Fields) error { apply(n); return nil }), nil
}

// ------------------------------------------------------------------ inputs

func newButton(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	btn := widget.NewButton(bc.Str(n, "text", "Button"), func() {
		bc.Emit(n.ID, "tapped", nil)
	})
	var icon fyne.Resource
	if name := bc.Str(n, "icon", ""); name != "" {
		icon = render.IconResource(name)
	}
	btn.Icon = icon
	btn.Importance = render.Importance(bc.Str(n, "importance", "medium"))
	if bc.Bool(n, "disabled", false) {
		btn.Disable()
	}
	return newLeaf(btn, func(n *scene.Node, _ scene.Fields) error {
		btn.SetText(bc.Str(n, "text", "Button"))
		btn.Importance = render.Importance(bc.Str(n, "importance", "medium"))
		if name := bc.Str(n, "icon", ""); name != "" {
			btn.Icon = render.IconResource(name)
		} else {
			btn.Icon = nil
		}
		if bc.Bool(n, "disabled", false) {
			btn.Disable()
		} else {
			btn.Enable()
		}
		btn.Refresh()
		return nil
	}), nil
}

func newHyperlink(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	parse := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			return &url.URL{}
		}
		return u
	}
	hl := widget.NewHyperlink(bc.Str(n, "text", "Link"), parse(bc.Str(n, "url", "")))
	hl.OnTapped = func() { bc.Emit(n.ID, "tapped", nil) }
	return newLeaf(hl, func(n *scene.Node, _ scene.Fields) error {
		hl.Text = bc.Str(n, "text", "Link")
		hl.URL = parse(bc.Str(n, "url", ""))
		hl.Refresh()
		return nil
	}), nil
}

func newEntry(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	entry := widget.NewEntry()
	if bc.Bool(n, "multiline", false) {
		entry = widget.NewMultiLineEntry()
	}
	if bc.Bool(n, "password", false) {
		entry.Password = true
	}
	entry.SetText(bc.Str(n, "text", ""))
	entry.SetPlaceHolder(bc.Str(n, "placeholder", ""))
	entry.OnChanged = func(s string) {
		bc.Emit(n.ID, "changed", map[string]any{"value": s})
	}
	entry.OnSubmitted = func(s string) {
		bc.Emit(n.ID, "submitted", map[string]any{"value": s})
	}
	multiline, password := bc.Bool(n, "multiline", false), bc.Bool(n, "password", false)
	return newLeaf(entry, func(n *scene.Node, _ scene.Fields) error {
		if bc.Bool(n, "multiline", false) != multiline || bc.Bool(n, "password", false) != password {
			return render.ErrRebuild
		}
		if entry.Text != bc.Str(n, "text", "") {
			entry.SetText(bc.Str(n, "text", ""))
		}
		entry.SetPlaceHolder(bc.Str(n, "placeholder", ""))
		return nil
	}), nil
}

func newCheck(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	c := widget.NewCheck(bc.Str(n, "text", "Check"), func(on bool) {
		bc.Emit(n.ID, "changed", map[string]any{"value": on})
	})
	c.SetChecked(bc.Bool(n, "checked", false))
	return newLeaf(c, func(n *scene.Node, _ scene.Fields) error {
		c.Text = bc.Str(n, "text", "Check")
		c.SetChecked(bc.Bool(n, "checked", false))
		c.Refresh()
		return nil
	}), nil
}

func newSelect(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	opts := bc.List(n, "options")
	sel := widget.NewSelect(opts, func(s string) {
		bc.Emit(n.ID, "changed", map[string]any{"value": s})
	})
	sel.SetSelected(bc.Str(n, "selected", ""))
	return newLeaf(sel, func(n *scene.Node, _ scene.Fields) error {
		sel.Options = bc.List(n, "options")
		sel.SetSelected(bc.Str(n, "selected", ""))
		sel.Refresh()
		return nil
	}), nil
}

func newRadio(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	rg := widget.NewRadioGroup(bc.List(n, "options"), func(s string) {
		bc.Emit(n.ID, "changed", map[string]any{"value": s})
	})
	rg.SetSelected(bc.Str(n, "selected", ""))
	return newLeaf(rg, func(n *scene.Node, _ scene.Fields) error {
		rg.Options = bc.List(n, "options")
		rg.SetSelected(bc.Str(n, "selected", ""))
		rg.Refresh()
		return nil
	}), nil
}

func newSlider(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	min := bc.Num(n, "min", 0)
	max := bc.Num(n, "max", 100)
	s := widget.NewSlider(min, max)
	s.Step = bc.Num(n, "step", 1)
	s.SetValue(bc.Num(n, "value", 0))
	s.OnChanged = func(v float64) {
		bc.Emit(n.ID, "changed", map[string]any{"value": v})
	}
	return newLeaf(s, func(n *scene.Node, _ scene.Fields) error {
		s.Min = bc.Num(n, "min", 0)
		s.Max = bc.Num(n, "max", 100)
		s.Step = bc.Num(n, "step", 1)
		s.SetValue(bc.Num(n, "value", 0))
		s.Refresh()
		return nil
	}), nil
}
