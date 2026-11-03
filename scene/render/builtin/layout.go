package builtin

import (
	"image/color"
	"math"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// scaleTheme scales text and spacing with the free canvas transform, so a
// scene laid out at 60% (the editor stage) looks like the same scene at 100%
// (the play view) instead of keeping full-size text in shrunken boxes.
type scaleTheme struct {
	base  fyne.Theme
	scale float32
}

func newScaleTheme() *scaleTheme {
	base := theme.DefaultTheme()
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil && app.Settings().Theme() != nil {
		base = app.Settings().Theme()
	}
	return &scaleTheme{base: base, scale: 1}
}

func (t *scaleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}
func (t *scaleTheme) Font(style fyne.TextStyle) fyne.Resource { return t.base.Font(style) }
func (t *scaleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}
func (t *scaleTheme) Size(name fyne.ThemeSizeName) float32 {
	size := t.base.Size(name)
	if t.scale <= 0 || t.scale == 1 {
		return size
	}
	switch name {
	case theme.SizeNameText, theme.SizeNameHeadingText, theme.SizeNameSubHeadingText,
		theme.SizeNameCaptionText, theme.SizeNamePadding, theme.SizeNameInnerPadding,
		theme.SizeNameInlineIcon, theme.SizeNameScrollBar, theme.SizeNameScrollBarSmall,
		theme.SizeNameSeparatorThickness, theme.SizeNameInputBorder:
		return size * t.scale
	}
	return size
}

// --------------------------------------------------------------------- free

type freeLive struct {
	stack, layer *fyne.Container
	over         *container.ThemeOverride
	theme        *scaleTheme
	outer        fyne.CanvasObject
	clip         *container.Clip
	clipped      bool
	letterbox    *canvas.Rectangle
	bgFill       *canvas.Rectangle
	bg           *canvas.Image
	fl           *render.FreeLayout
	children     []render.Live
	bc           *render.BuildContext
	bgGen        atomic.Int64
	bgRef        string
	scaleText    bool
}

func newFree(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	fl := render.NewFreeLayout(fyne.NewSize(1600, 900), "fit")
	layer := &fyne.Container{Layout: fl}
	bgFill := canvas.NewRectangle(color.Transparent)
	bgFill.Hide()
	bg := canvas.NewImageFromResource(nil)
	bg.FillMode = canvas.ImageFillStretch
	bg.Hide()
	layer.Objects = []fyne.CanvasObject{bgFill, bg}
	letterbox := canvas.NewRectangle(render.Color("#000000", nil))
	stack := container.NewStack(letterbox, layer)
	l := &freeLive{
		stack: stack, layer: layer, letterbox: letterbox, bgFill: bgFill, bg: bg, fl: fl, bc: bc,
		theme: newScaleTheme(), scaleText: true,
	}
	// Widget text and spacing scale with the canvas transform, so the editor
	// (which lays the stage out at ~60%) matches play mode.
	l.over = container.NewThemeOverride(stack, l.theme)
	fl.OnLayout = func(size fyne.Size) { l.applyScale(size) }
	if bc.Bool(n, "clip", false) {
		l.clip = container.NewClip(l.over)
		l.clipped = true
		l.outer = l.clip
	} else {
		l.outer = l.over
	}
	if err := l.apply(n); err != nil {
		return nil, err
	}
	return l, nil
}

// applyScale updates the widget theme from the current free-layout transform.
// The scale is quantized so a window resize does not refresh the whole subtree
// on every pixel; the refresh is queued (fyne.Do) so it cannot re-enter the
// layout that is running.
func (l *freeLive) applyScale(size fyne.Size) {
	if !l.scaleText {
		return
	}
	s := float32(l.fl.Transform(size).SX)
	if s <= 0 {
		s = 1
	}
	s = float32(math.Round(float64(s)*10)) / 10
	if s == l.theme.scale {
		return
	}
	l.theme.scale = s
	over := l.over
	fyne.Do(func() { over.Refresh() })
}

func (l *freeLive) Object() fyne.CanvasObject { return l.outer }

func (l *freeLive) FreeLayout() *render.FreeLayout { return l.fl }

func (l *freeLive) apply(n *scene.Node) error {
	designW := float32(l.bc.Num(n, "designW", 1600))
	designH := float32(l.bc.Num(n, "designH", 900))
	if designW <= 0 {
		designW = 1
	}
	if designH <= 0 {
		designH = 1
	}
	design := fyne.NewSize(designW, designH)
	l.fl.SetDesign(design, l.bc.Str(n, "scale", render.ScaleFit))
	l.scaleText = l.bc.Bool(n, "scaleText", true)
	if !l.scaleText && l.theme.scale != 1 {
		l.theme.scale = 1
		over := l.over
		fyne.Do(func() { over.Refresh() })
	}
	l.letterbox.FillColor = l.bc.ColorProp(n, "letterbox", render.Color("#000000", nil))
	// The design-area background is separate from the letterbox. An empty
	// value keeps the historic look (the letterbox shows through).
	l.fl.SetPlace(l.bgFill, scene.Place{W: float64(designW), H: float64(designH)})
	if bg := l.bc.Str(n, "backgroundColor", ""); bg == "" {
		l.bgFill.Hide()
	} else {
		l.bgFill.FillColor = l.bc.ColorProp(n, "backgroundColor", color.Transparent)
		l.bgFill.Show()
	}
	l.bg.FillMode = render.FillMode(l.bc.Str(n, "backgroundFill", "stretch"))
	l.fl.SetPlace(l.bg, scene.Place{W: float64(designW), H: float64(designH)})

	clip := l.bc.Bool(n, "clip", false)
	if clip != l.clipped {
		return render.ErrRebuild
	}
	if src := l.bc.Str(n, "background", ""); src != l.bgRef {
		l.bgRef = src
		if src == "" {
			l.bgGen.Add(1)
			l.bg.Hide()
			l.bg.Resource = nil
			l.bg.Refresh()
		} else {
			loadAsset(l.bc, src, &l.bgGen, func(res fyne.Resource, err error) {
				if err != nil {
					l.bc.Statusf("background %s: %v", src, err)
					return
				}
				l.bg.Resource = res
				l.bg.Show()
				l.bg.Refresh()
			})
		}
	}
	// Design/scale changes must re-run the free layout even when the container
	// size itself did not change. Repaint only: a Container.Refresh would
	// recursively refresh (and re-decode) every child on every property edit.
	l.fl.Layout(l.layer.Objects, l.layer.Size())
	canvas.Refresh(l.letterbox)
	canvas.Refresh(l.bgFill)
	canvas.Refresh(l.bg)
	canvas.Refresh(l.layer)
	return nil
}

func (l *freeLive) Update(n *scene.Node, _ scene.Fields) error {
	if err := l.apply(n); err != nil {
		return err
	}
	return nil
}

func (l *freeLive) Insert(index int, child render.Live, p scene.Place) error {
	index = clamp(index, 0, len(l.children))
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = child
	l.fl.SetPlace(child.Object(), p)
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sync()
	return nil
}

func (l *freeLive) Remove(child render.Live) error {
	for i, c := range l.children {
		if c == child {
			l.fl.DeletePlace(c.Object())
			l.children = append(l.children[:i], l.children[i+1:]...)
			l.sync()
			return nil
		}
	}
	return nil
}

func (l *freeLive) Reorder(child render.Live, index int) error {
	from := -1
	for i, c := range l.children {
		if c == child {
			from = i
			break
		}
	}
	if from < 0 {
		return nil
	}
	index = clamp(index, 0, len(l.children)-1)
	if from == index {
		return nil
	}
	c := l.children[from]
	l.children = append(l.children[:from], l.children[from+1:]...)
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = c
	l.sync()
	return nil
}

func (l *freeLive) SetPlace(child render.Live, p scene.Place) error {
	l.fl.SetPlace(child.Object(), p)
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.fl.LayoutObject(child.Object(), l.layer.Size())
	canvas.Refresh(child.Object())
	return nil
}

func (l *freeLive) Preview(child render.Live, p scene.Place) {
	obj := child.Object()
	l.fl.SetPreview(obj, p)
	l.fl.LayoutObject(obj, l.layer.Size())
	canvas.Refresh(obj)
}

func (l *freeLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	return render.DropSpot{Index: len(l.children), Rect: render.NewRect(local.X, local.Y, 1, 1)}, true
}

// CancelPreview removes a drag preview.
func (l *freeLive) CancelPreview(render.Live) {
	l.fl.ClearPreviews()
	canvas.Refresh(l.layer)
}

func (l *freeLive) Transform(size fyne.Size) render.Transform {
	return l.fl.Transform(size)
}

func (l *freeLive) sync() {
	objs := make([]fyne.CanvasObject, 2+len(l.children))
	objs[0] = l.bgFill
	objs[1] = l.bg
	for i, c := range l.children {
		objs[i+2] = c.Object()
	}
	l.layer.Objects = objs
	l.fl.Layout(objs, l.layer.Size())
	canvas.Refresh(l.layer)
}

func (l *freeLive) Destroy() { l.bgGen.Add(1) }

// ------------------------------------------------------------------ ordered

type orderedLive struct {
	box      *fyne.Container
	children []render.Live
	axis     byte
	bc       *render.BuildContext
	kind     string
}

func newOrdered(kind string) func(*render.BuildContext, *scene.Node) (render.Live, error) {
	return func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		var box *fyne.Container
		var axis byte
		switch kind {
		case "vbox":
			box, axis = container.NewVBox(), 'y'
		case "hbox":
			box, axis = container.NewHBox(), 'x'
		case "grid":
			box = container.NewGridWithColumns(int(bc.Num(n, "cols", 2)))
		case "gridwrap":
			box = container.NewGridWrap(fyne.NewSize(120, 120))
		case "stack":
			box = container.NewStack()
		case "center":
			box = container.NewCenter()
		case "padded":
			box = container.NewPadded()
		default:
			box = container.NewVBox()
		}
		return &orderedLive{box: box, axis: axis, bc: bc, kind: kind}, nil
	}
}

func (l *orderedLive) Object() fyne.CanvasObject { return l.box }

func (l *orderedLive) Update(n *scene.Node, _ scene.Fields) error {
	if l.kind == "grid" {
		l.box.Layout = layout.NewGridLayoutWithColumns(int(l.bc.Num(n, "cols", 2)))
		l.box.Refresh()
	}
	return nil
}

func (l *orderedLive) Insert(index int, child render.Live, p scene.Place) error {
	index = clamp(index, 0, len(l.children))
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = child
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sync()
	return nil
}

func (l *orderedLive) Remove(child render.Live) error {
	for i, c := range l.children {
		if c == child {
			l.children = append(l.children[:i], l.children[i+1:]...)
			l.sync()
			return nil
		}
	}
	return nil
}

func (l *orderedLive) Reorder(child render.Live, index int) error {
	from := -1
	for i, c := range l.children {
		if c == child {
			from = i
			break
		}
	}
	if from < 0 {
		return nil
	}
	index = clamp(index, 0, len(l.children)-1)
	if from == index {
		return nil
	}
	c := l.children[from]
	l.children = append(l.children[:from], l.children[from+1:]...)
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = c
	l.sync()
	return nil
}

func (l *orderedLive) SetPlace(child render.Live, p scene.Place) error {
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.box.Refresh()
	return nil
}

func (l *orderedLive) Preview(render.Live, scene.Place) {}

func (l *orderedLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	index := len(l.children)
	switch l.axis {
	case 'y':
		for i, c := range l.children {
			o := c.Object()
			if local.Y < o.Position().Y+o.Size().Height/2 {
				index = i
				break
			}
		}
		y := float32(0)
		if index > 0 && index <= len(l.children) {
			o := l.children[index-1].Object()
			y = o.Position().Y + o.Size().Height
		}
		return render.DropSpot{Index: index, Rect: render.NewRect(0, y, l.box.Size().Width, 2)}, true
	case 'x':
		for i, c := range l.children {
			o := c.Object()
			if local.X < o.Position().X+o.Size().Width/2 {
				index = i
				break
			}
		}
		x := float32(0)
		if index > 0 && index <= len(l.children) {
			o := l.children[index-1].Object()
			x = o.Position().X + o.Size().Width
		}
		return render.DropSpot{Index: index, Rect: render.NewRect(x, 0, 2, l.box.Size().Height)}, true
	}
	return render.DropSpot{Index: index, Rect: render.NewRect(0, l.box.Size().Height, l.box.Size().Width, 2)}, true
}

func (l *orderedLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

// DirectLayout: ordered children are direct objects of the box, so their
// positions are relative to the container's rect (no canvas walk needed).
func (l *orderedLive) DirectLayout() bool { return true }

func (l *orderedLive) Destroy() {}

func (l *orderedLive) sync() {
	objs := make([]fyne.CanvasObject, len(l.children))
	for i, c := range l.children {
		objs[i] = c.Object()
	}
	l.box.Objects = objs
	if l.box.Layout != nil {
		l.box.Layout.Layout(objs, l.box.Size())
	}
	canvas.Refresh(l.box)
}

// ------------------------------------------------------------------- border

var borderSlots = []string{"top", "bottom", "left", "right", "center"}

type borderLive struct {
	box  *fyne.Container
	slot map[string]render.Live
	bc   *render.BuildContext
}

func newBorder(bc *render.BuildContext, _ *scene.Node) (render.Live, error) {
	box := container.NewBorder(nil, nil, nil, nil, nil)
	return &borderLive{box: box, slot: make(map[string]render.Live), bc: bc}, nil
}

func (l *borderLive) Object() fyne.CanvasObject { return l.box }

func (l *borderLive) Update(*scene.Node, scene.Fields) error { return nil }

func (l *borderLive) Insert(_ int, child render.Live, p scene.Place) error {
	slot := p.Slot
	if slot == "" {
		slot = "center"
	}
	l.removeChild(child)
	l.slot[slot] = child
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sync()
	return nil
}

func (l *borderLive) removeChild(child render.Live) {
	for name, c := range l.slot {
		if c == child {
			delete(l.slot, name)
			return
		}
	}
	for _, name := range borderSlots {
		if l.slot[name] == child {
			delete(l.slot, name)
		}
	}
}

func (l *borderLive) Remove(child render.Live) error {
	l.removeChild(child)
	l.sync()
	return nil
}

func (l *borderLive) Reorder(render.Live, int) error { return nil }

func (l *borderLive) SetPlace(child render.Live, p scene.Place) error {
	cur := ""
	for name, c := range l.slot {
		if c == child {
			cur = name
			break
		}
	}
	slot := p.Slot
	if slot == "" {
		slot = "center"
	}
	if cur != slot {
		delete(l.slot, cur)
		l.slot[slot] = child
	}
	l.sync()
	return nil
}

func (l *borderLive) Preview(render.Live, scene.Place) {}

func (l *borderLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	w, h := l.box.Size().Width, l.box.Size().Height
	slot := "center"
	rect := render.NewRect(0, 0, w, h)
	switch {
	case local.Y < h*0.2:
		slot, rect = "top", render.NewRect(0, 0, w, h*0.2)
	case local.Y > h*0.8:
		slot, rect = "bottom", render.NewRect(0, h*0.8, w, h*0.2)
	case local.X < w*0.15:
		slot, rect = "left", render.NewRect(0, 0, w*0.15, h)
	case local.X > w*0.85:
		slot, rect = "right", render.NewRect(w*0.85, 0, w*0.15, h)
	}
	return render.DropSpot{Slot: slot, Rect: rect}, true
}

func (l *borderLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *borderLive) DirectLayout() bool { return true }

func (l *borderLive) Destroy() {}

func (l *borderLive) slotObject(name string) fyne.CanvasObject {
	if c, ok := l.slot[name]; ok && c != nil {
		return c.Object()
	}
	return nil
}

func (l *borderLive) sync() {
	// Only occupied slots go into Objects: Fyne's border layout dereferences
	// every entry, and an unoccupied slot used to be a nil in a 5-element
	// slice (a panic on the first refresh).
	objs := make([]fyne.CanvasObject, 0, len(l.slot))
	for _, name := range borderSlots {
		if o := l.slotObject(name); o != nil {
			objs = append(objs, o)
		}
	}
	l.box.Objects = objs
	l.box.Layout = layout.NewBorderLayout(
		l.slotObject("top"), l.slotObject("bottom"),
		l.slotObject("left"), l.slotObject("right"))
	l.box.Refresh()
}

// -------------------------------------------------------------------- split

type splitLive struct {
	split    *container.Split
	leading  render.Live
	trailing render.Live
	emptyL   fyne.CanvasObject
	emptyT   fyne.CanvasObject
}

func newSplit(horizontal bool) func(*render.BuildContext, *scene.Node) (render.Live, error) {
	return func(*render.BuildContext, *scene.Node) (render.Live, error) {
		emptyL := canvas.NewRectangle(render.Color("#00000000", nil))
		emptyT := canvas.NewRectangle(render.Color("#00000000", nil))
		var sp *container.Split
		if horizontal {
			sp = container.NewHSplit(emptyL, emptyT)
		} else {
			sp = container.NewVSplit(emptyL, emptyT)
		}
		return &splitLive{split: sp, emptyL: emptyL, emptyT: emptyT}, nil
	}
}

func (l *splitLive) Object() fyne.CanvasObject { return l.split }

func (l *splitLive) Update(n *scene.Node, _ scene.Fields) error {
	off := 0.5
	if v, ok := n.Prop("offset"); ok {
		if f, ok := v.(float64); ok {
			off = f
		}
	}
	if off < 0.02 {
		off = 0.02
	}
	if off > 0.98 {
		off = 0.98
	}
	l.split.SetOffset(off)
	return nil
}

func (l *splitLive) Insert(_ int, child render.Live, p scene.Place) error {
	switch p.Slot {
	case "trailing":
		l.trailing = child
		l.split.Trailing = child.Object()
	default:
		l.leading = child
		l.split.Leading = child.Object()
	}
	l.split.Refresh()
	return nil
}

func (l *splitLive) Remove(child render.Live) error {
	if l.leading == child {
		l.leading = nil
		l.split.Leading = l.emptyL
	}
	if l.trailing == child {
		l.trailing = nil
		l.split.Trailing = l.emptyT
	}
	l.split.Refresh()
	return nil
}

func (l *splitLive) Reorder(render.Live, int) error { return nil }

func (l *splitLive) SetPlace(child render.Live, p scene.Place) error {
	slot := "leading"
	if p.Slot == "trailing" {
		slot = "trailing"
	}
	if slot == "trailing" && l.trailing == child {
		return nil
	}
	if slot == "leading" && l.leading == child {
		return nil
	}
	// Move between slots.
	if l.leading == child {
		l.leading = nil
		l.split.Leading = l.emptyL
	} else if l.trailing == child {
		l.trailing = nil
		l.split.Trailing = l.emptyT
	}
	return l.Insert(0, child, p)
}

func (l *splitLive) Preview(render.Live, scene.Place) {}

func (l *splitLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	if l.split.Horizontal {
		half := l.split.Size().Width / 2
		if local.X < half {
			return render.DropSpot{Slot: "leading", Rect: render.NewRect(0, 0, half, l.split.Size().Height)}, true
		}
		return render.DropSpot{Slot: "trailing", Rect: render.NewRect(half, 0, half, l.split.Size().Height)}, true
	}
	half := l.split.Size().Height / 2
	if local.Y < half {
		return render.DropSpot{Slot: "leading", Rect: render.NewRect(0, 0, l.split.Size().Width, half)}, true
	}
	return render.DropSpot{Slot: "trailing", Rect: render.NewRect(0, half, l.split.Size().Width, half)}, true
}

func (l *splitLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *splitLive) Destroy() {}

// ------------------------------------------------------------------- scroll

type scrollLive struct {
	sc      *container.Scroll
	content render.Live
	empty   fyne.CanvasObject
}

func newScroll(_ *render.BuildContext, _ *scene.Node) (render.Live, error) {
	empty := canvas.NewRectangle(render.Color("#00000000", nil))
	return &scrollLive{sc: container.NewScroll(empty), empty: empty}, nil
}

func (l *scrollLive) Object() fyne.CanvasObject { return l.sc }

func (l *scrollLive) Update(*scene.Node, scene.Fields) error { return nil }

func (l *scrollLive) Insert(_ int, child render.Live, p scene.Place) error {
	l.content = child
	l.sc.Content = child.Object()
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sc.Refresh()
	return nil
}

func (l *scrollLive) Remove(child render.Live) error {
	if l.content == child {
		l.content = nil
		l.sc.Content = l.empty
		l.sc.Refresh()
	}
	return nil
}

func (l *scrollLive) Reorder(render.Live, int) error { return nil }

func (l *scrollLive) SetPlace(child render.Live, p scene.Place) error {
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sc.Refresh()
	return nil
}

func (l *scrollLive) Preview(render.Live, scene.Place) {}

func (l *scrollLive) DropSpot(fyne.Position) (render.DropSpot, bool) {
	return render.DropSpot{Slot: "content", Rect: render.NewRect(0, 0, l.sc.Size().Width, l.sc.Size().Height)}, true
}

func (l *scrollLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *scrollLive) Destroy() {}

// --------------------------------------------------------------------- card

type cardLive struct {
	card    *widget.Card
	content render.Live
	empty   fyne.CanvasObject
	bc      *render.BuildContext
}

func newCard(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	empty := canvas.NewRectangle(render.Color("#00000000", nil))
	card := widget.NewCard(bc.Str(n, "title", ""), bc.Str(n, "subtitle", ""), empty)
	return &cardLive{card: card, empty: empty, bc: bc}, nil
}

func (l *cardLive) Object() fyne.CanvasObject { return l.card }

func (l *cardLive) Update(n *scene.Node, _ scene.Fields) error {
	l.card.Title = l.bc.Str(n, "title", "")
	l.card.Subtitle = l.bc.Str(n, "subtitle", "")
	l.card.Refresh()
	return nil
}

func (l *cardLive) Insert(_ int, child render.Live, p scene.Place) error {
	l.content = child
	l.card.Content = child.Object()
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.card.Refresh()
	return nil
}

func (l *cardLive) Remove(child render.Live) error {
	if l.content == child {
		l.content = nil
		l.card.Content = l.empty
		l.card.Refresh()
	}
	return nil
}

func (l *cardLive) Reorder(render.Live, int) error { return nil }

func (l *cardLive) SetPlace(child render.Live, p scene.Place) error {
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.card.Refresh()
	return nil
}

func (l *cardLive) Preview(render.Live, scene.Place) {}

func (l *cardLive) DropSpot(fyne.Position) (render.DropSpot, bool) {
	return render.DropSpot{Slot: "content", Rect: render.NewRect(0, 0, l.card.Size().Width, l.card.Size().Height)}, true
}

func (l *cardLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *cardLive) Destroy() {}

// --------------------------------------------------------------------- tabs

type tabsLive struct {
	tabs     *container.AppTabs
	children []render.Live
	items    map[render.Live]*container.TabItem
	bc       *render.BuildContext
}

func newTabs(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	tabs := container.NewAppTabs()
	l := &tabsLive{tabs: tabs, items: make(map[render.Live]*container.TabItem), bc: bc}
	l.applyLocation(n)
	return l, nil
}

func (l *tabsLive) applyLocation(n *scene.Node) {
	switch l.bc.Str(n, "location", "top") {
	case "bottom":
		l.tabs.SetTabLocation(container.TabLocationBottom)
	case "left":
		l.tabs.SetTabLocation(container.TabLocationLeading)
	case "right":
		l.tabs.SetTabLocation(container.TabLocationTrailing)
	default:
		l.tabs.SetTabLocation(container.TabLocationTop)
	}
}

func (l *tabsLive) Object() fyne.CanvasObject { return l.tabs }

func (l *tabsLive) Update(n *scene.Node, _ scene.Fields) error {
	l.applyLocation(n)
	return nil
}

func tabTitle(p scene.Place, index int) string {
	if p.Title != "" {
		return p.Title
	}
	return "Tab " + itoa(index+1)
}

func tabIcon(p scene.Place) fyne.Resource {
	if p.Icon == "" {
		return nil
	}
	return render.IconResource(p.Icon)
}

func (l *tabsLive) Insert(index int, child render.Live, p scene.Place) error {
	index = clamp(index, 0, len(l.children))
	item := container.NewTabItem(tabTitle(p, index), child.Object())
	if icon := tabIcon(p); icon != nil {
		item.Icon = icon
	}
	l.items[child] = item
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = child
	l.sync(index)
	return nil
}

func (l *tabsLive) Remove(child render.Live) error {
	for i, c := range l.children {
		if c == child {
			l.children = append(l.children[:i], l.children[i+1:]...)
			delete(l.items, child)
			l.sync(0)
			return nil
		}
	}
	return nil
}

func (l *tabsLive) Reorder(child render.Live, index int) error {
	from := -1
	for i, c := range l.children {
		if c == child {
			from = i
			break
		}
	}
	if from < 0 {
		return nil
	}
	index = clamp(index, 0, len(l.children)-1)
	c := l.children[from]
	l.children = append(l.children[:from], l.children[from+1:]...)
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = c
	l.sync(index)
	return nil
}

func (l *tabsLive) SetPlace(child render.Live, p scene.Place) error {
	if item, ok := l.items[child]; ok {
		item.Text = tabTitle(p, 0)
		item.Icon = tabIcon(p)
		l.tabs.SetItems(l.tabItems())
		l.tabs.Refresh()
	}
	return nil
}

func (l *tabsLive) Preview(render.Live, scene.Place) {}

func (l *tabsLive) DropSpot(fyne.Position) (render.DropSpot, bool) {
	return render.DropSpot{Index: len(l.children), Rect: render.NewRect(0, 0, l.tabs.Size().Width, l.tabs.Size().Height)}, true
}

func (l *tabsLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *tabsLive) tabItems() []*container.TabItem {
	out := make([]*container.TabItem, len(l.children))
	for i, c := range l.children {
		out[i] = l.items[c]
	}
	return out
}

func (l *tabsLive) Destroy() {}

func (l *tabsLive) sync(selectIndex int) {
	l.tabs.SetItems(l.tabItems())
	if selectIndex >= 0 && selectIndex < len(l.children) {
		l.tabs.SelectIndex(selectIndex)
	}
	l.tabs.Refresh()
}

// ---------------------------------------------------------------- accordion

type accordionLive struct {
	acc      *widget.Accordion
	children []render.Live
	items    map[render.Live]*widget.AccordionItem
}

func newAccordion(_ *render.BuildContext, _ *scene.Node) (render.Live, error) {
	return &accordionLive{acc: widget.NewAccordion(), items: make(map[render.Live]*widget.AccordionItem)}, nil
}

func (l *accordionLive) Object() fyne.CanvasObject { return l.acc }

func (l *accordionLive) Update(*scene.Node, scene.Fields) error { return nil }

func (l *accordionLive) Insert(index int, child render.Live, p scene.Place) error {
	index = clamp(index, 0, len(l.children))
	item := widget.NewAccordionItem(tabTitle(p, index), child.Object())
	l.items[child] = item
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = child
	l.sync()
	return nil
}

func (l *accordionLive) Remove(child render.Live) error {
	for i, c := range l.children {
		if c == child {
			l.children = append(l.children[:i], l.children[i+1:]...)
			delete(l.items, child)
			l.sync()
			return nil
		}
	}
	return nil
}

func (l *accordionLive) Reorder(child render.Live, index int) error {
	from := -1
	for i, c := range l.children {
		if c == child {
			from = i
			break
		}
	}
	if from < 0 {
		return nil
	}
	index = clamp(index, 0, len(l.children)-1)
	c := l.children[from]
	l.children = append(l.children[:from], l.children[from+1:]...)
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = c
	l.sync()
	return nil
}

func (l *accordionLive) SetPlace(child render.Live, p scene.Place) error {
	if item, ok := l.items[child]; ok {
		item.Title = tabTitle(p, 0)
		l.acc.Refresh()
	}
	return nil
}

func (l *accordionLive) Preview(render.Live, scene.Place) {}

func (l *accordionLive) DropSpot(fyne.Position) (render.DropSpot, bool) {
	return render.DropSpot{Index: len(l.children), Rect: render.NewRect(0, 0, l.acc.Size().Width, l.acc.Size().Height)}, true
}

func (l *accordionLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *accordionLive) Destroy() {}

func (l *accordionLive) sync() {
	items := make([]*widget.AccordionItem, len(l.children))
	for i, c := range l.children {
		items[i] = l.items[c]
	}
	l.acc.Items = items
	l.acc.Refresh()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
