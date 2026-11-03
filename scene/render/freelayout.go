package render

import (
	"math"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/scene"
)

// ScaleMode values understood by FreeLayout.
const (
	ScaleNone    = "none"
	ScaleFit     = "fit"
	ScaleFill    = "fill"
	ScaleStretch = "stretch"
)

// FreeLayout places children at their design-unit x/y/w/h. It is a real Fyne
// layout, so positions are recomputed from the document on every layout pass:
// unlike absolute Move calls, they cannot drift when a parent resizes.
type FreeLayout struct {
	Design fyne.Size
	Mode   string

	places  map[fyne.CanvasObject]scene.Place
	preview map[fyne.CanvasObject]scene.Place
}

// NewFreeLayout returns a layout for a design size and scale mode.
func NewFreeLayout(design fyne.Size, mode string) *FreeLayout {
	return &FreeLayout{
		Design:  design,
		Mode:    mode,
		places:  make(map[fyne.CanvasObject]scene.Place),
		preview: make(map[fyne.CanvasObject]scene.Place),
	}
}

// SetDesign updates the design size and scale mode.
func (l *FreeLayout) SetDesign(design fyne.Size, mode string) {
	l.Design, l.Mode = design, mode
}

// SetPlace records a child's placement.
func (l *FreeLayout) SetPlace(o fyne.CanvasObject, p scene.Place) {
	if l.places == nil {
		l.places = make(map[fyne.CanvasObject]scene.Place)
	}
	l.places[o] = p
}

// Place returns a child's placement.
func (l *FreeLayout) Place(o fyne.CanvasObject) (scene.Place, bool) {
	p, ok := l.places[o]
	return p, ok
}

// DeletePlace forgets a child's placement.
func (l *FreeLayout) DeletePlace(o fyne.CanvasObject) {
	delete(l.places, o)
	delete(l.preview, o)
}

// SetPreview shows a temporary placement during a drag.
func (l *FreeLayout) SetPreview(o fyne.CanvasObject, p scene.Place) {
	if l.preview == nil {
		l.preview = make(map[fyne.CanvasObject]scene.Place)
	}
	l.preview[o] = p
}

// ClearPreview removes a drag preview.
func (l *FreeLayout) ClearPreview(o fyne.CanvasObject) {
	delete(l.preview, o)
}

// ClearPreviews removes every drag preview.
func (l *FreeLayout) ClearPreviews() { l.preview = nil }

func (l *FreeLayout) placeOf(o fyne.CanvasObject) scene.Place {
	if p, ok := l.preview[o]; ok {
		return p
	}
	return l.places[o]
}

// Transform returns the design→local transform for a container of size.
func (l *FreeLayout) Transform(size fyne.Size) Transform {
	mode := l.Mode
	if mode == "" {
		mode = ScaleFit
	}
	dw, dh := float64(l.Design.Width), float64(l.Design.Height)
	w, h := float64(size.Width), float64(size.Height)
	if dw <= 0 || dh <= 0 {
		return Transform{SX: 1, SY: 1}
	}
	switch mode {
	case ScaleNone:
		return Transform{SX: 1, SY: 1}
	case ScaleStretch:
		return Transform{SX: w / dw, SY: h / dh}
	case ScaleFill:
		s := math.Max(w/dw, h/dh)
		return Transform{SX: s, SY: s, OX: (w - dw*s) / 2, OY: (h - dh*s) / 2}
	default: // fit
		s := math.Min(w/dw, h/dh)
		return Transform{SX: s, SY: s, OX: (w - dw*s) / 2, OY: (h - dh*s) / 2}
	}
}

// ChildRect returns the local rectangle of a child placed in a container of
// size, applying anchors in "none" mode.
func (l *FreeLayout) ChildRect(size fyne.Size, p scene.Place) Rect {
	if l.Mode == ScaleNone || l.Mode == "" {
		p = applyAnchor(p, l.Design, size)
	}
	t := l.Transform(size)
	return NewRect(
		float32(t.OX+p.X*t.SX),
		float32(t.OY+p.Y*t.SY),
		float32(p.W*t.SX),
		float32(p.H*t.SY),
	)
}

// Layout implements fyne.Layout.
func (l *FreeLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objs {
		l.LayoutObject(o, size)
	}
}

// LayoutObject places one child exactly as Layout would, without touching the
// others. Drag previews use it: re-running the whole layout (or refreshing the
// container) on every pointer event also refreshes every sibling, and a
// container refresh re-decodes every image in the tree.
func (l *FreeLayout) LayoutObject(o fyne.CanvasObject, size fyne.Size) {
	if o == nil {
		return
	}
	t := l.Transform(size)
	if s, ok := o.(interface{ SetScale(sx, sy float64) }); ok {
		s.SetScale(t.SX, t.SY)
	}
	p := l.placeOf(o)
	if l.Mode == ScaleNone || l.Mode == "" {
		p = applyAnchor(p, l.Design, size)
	}
	o.Move(fyne.NewPos(
		float32(t.OX+p.X*t.SX),
		float32(t.OY+p.Y*t.SY),
	))
	o.Resize(fyne.NewSize(float32(p.W*t.SX), float32(p.H*t.SY)))
}

// MinSize implements fyne.Layout. Scaled layouts can shrink to nothing; an
// unscaled design keeps its size so a Scroll parent has something to scroll.
func (l *FreeLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	if l.Mode == ScaleNone && l.Design.Width > 0 && l.Design.Height > 0 {
		return l.Design
	}
	return fyne.NewSize(1, 1)
}

// AnchorHorizontal and AnchorVertical values.
const (
	AnchorLeft    = "left"
	AnchorRight   = "right"
	AnchorCenter  = "center"
	AnchorStretch = "stretch"
	AnchorTop     = "top"
	AnchorBottom  = "bottom"
)

func applyAnchor(p scene.Place, design, size fyne.Size) scene.Place {
	if p.Anchor == nil {
		return p
	}
	dw, dh := design.Width, design.Height
	if dw <= 0 {
		dw = size.Width
	}
	if dh <= 0 {
		dh = size.Height
	}
	switch p.Anchor.H {
	case AnchorRight:
		margin := float64(dw) - (p.X + p.W)
		p.X = float64(size.Width) - margin - p.W
	case AnchorCenter:
		p.X += (float64(size.Width) - float64(dw)) / 2
	case AnchorStretch:
		right := p.X + p.W
		p.W = float64(size.Width) - p.X - (float64(dw) - right)
		if p.W < 1 {
			p.W = 1
		}
	}
	switch p.Anchor.V {
	case AnchorBottom:
		margin := float64(dh) - (p.Y + p.H)
		p.Y = float64(size.Height) - margin - p.H
	case AnchorCenter:
		p.Y += (float64(size.Height) - float64(dh)) / 2
	case AnchorStretch:
		bottom := p.Y + p.H
		p.H = float64(size.Height) - p.Y - (float64(dh) - bottom)
		if p.H < 1 {
			p.H = 1
		}
	}
	return p
}
