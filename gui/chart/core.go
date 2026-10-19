package risorchart

import (
	"image/color"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// Layout metrics shared by all interactive charts (device-independent pixels).
const (
	plotPadTop    = float32(34)
	plotPadBottom = float32(34)
	plotPadLeft   = float32(48)
	plotPadRight  = float32(14)

	titleFontSize = float32(14)
	labelFontSize = float32(11)
	tickFontSize  = float32(10)
	tipFontSize   = float32(11)

	minChartWidth  = float32(420)
	minChartHeight = float32(300)
)

// Fixed palette so on-screen charts and SVG exports render identically
// regardless of the active Fyne theme.
var (
	colBackground = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	colFrame      = color.NRGBA{0x66, 0x66, 0x66, 0xff}
	colGrid       = color.NRGBA{0xd9, 0xd9, 0xd9, 0xff}
	colText       = color.NRGBA{0x33, 0x33, 0x33, 0xff}
	colAxisLabel  = color.NRGBA{0x66, 0x66, 0x66, 0xff}
	colTipBg      = color.NRGBA{0xff, 0xff, 0xef, 0xff}
	colTipBorder  = color.NRGBA{0x99, 0x99, 0x99, 0xff}
)

// hitResult identifies the data element under the pointer.
type hitResult struct {
	Index    int
	Label    string
	Value    float64
	ValueStr string
}

// seriesOpts carries transient interaction state into buildSeries.
type seriesOpts struct {
	hoverIdx   int
	showValues bool
}

// chartDelegate is implemented by each chart type (bar, line, scatter, ...).
// The engine owns axes, grid, chrome, tooltips and interaction; the delegate
// owns data geometry and hit-testing.
type chartDelegate interface {
	// dataExtent returns the full data domain in data units.
	dataExtent() (x0, x1, y0, y1 float64)
	// xTicks returns positions and labels for the *current* x domain.
	xTicks() ([]float64, []string)
	// buildSeries draws the series primitives into the clipped plot painter.
	buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts)
	// hitTest finds the element under widget-local pixels (px, py), which the
	// engine has already confirmed to be inside the plot area.
	hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool)
}

// chartCore is the shared interactive chart engine. It is a Fyne widget that
// renders a scene of canvas primitives and handles hover, tap, right-click,
// wheel zoom, drag pan and double-click reset. Chart types embed it.
type chartCore struct {
	widget.BaseWidget

	typeName string // Risor type name, used in error messages
	delegate chartDelegate

	title, xLabel, yLabel string

	xScale, yScale *linearScale

	showGrid    bool
	showValues  bool
	zoomEnabled bool

	hoverIdx int
	hoverPt  fyne.Position
	tipText  string

	sc     scene
	view   sceneView
	render *chartRenderer

	// Go-level hooks; the Risor layer assigns closures through these.
	onTapFn       func(hitResult)
	onHoverFn     func(hitResult)
	onSecondaryFn func(hitResult)
}

func newChartCore(typeName string, d chartDelegate) *chartCore {
	c := &chartCore{
		typeName:    typeName,
		delegate:    d,
		showGrid:    true,
		zoomEnabled: true,
		hoverIdx:    -1,
	}
	c.recomputeDomains()
	return c
}

// recomputeDomains rebuilds both scales from the delegate's current extent.
func (c *chartCore) recomputeDomains() {
	x0, x1, y0, y1 := c.delegate.dataExtent()
	c.xScale = newLinearScale(x0, x1)
	c.yScale = newLinearScale(y0, y1)
}

// --- Fyne widget plumbing ---------------------------------------------------

func (c *chartCore) MinSize() fyne.Size {
	return fyne.NewSize(minChartWidth, minChartHeight)
}

func (c *chartCore) CreateRenderer() fyne.WidgetRenderer {
	c.render = &chartRenderer{core: c}
	return c.render
}

func (c *chartCore) refresh() {
	c.BaseWidget.Refresh()
}

// chartRenderer keeps the scene and the pooled canvas objects in sync.
type chartRenderer struct {
	core *chartCore
	objs []fyne.CanvasObject
}

func (r *chartRenderer) Layout(size fyne.Size) {
	r.core.build(size)
	r.core.view.sync(&r.core.sc)
	r.objs = r.core.view.objects()
}

func (r *chartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(minChartWidth, minChartHeight)
}

func (r *chartRenderer) Objects() []fyne.CanvasObject {
	return r.objs
}

func (r *chartRenderer) Refresh() {
	r.Layout(r.core.Size())
}

func (r *chartRenderer) Destroy() {}

// --- scene construction -----------------------------------------------------

// build reconstructs the full scene for the given size: background, title,
// axes, grid, tick labels, series (via the delegate) and hover tooltip.
func (c *chartCore) build(size fyne.Size) {
	if size.Width < 1 || size.Height < 1 {
		return
	}

	s := &c.sc
	s.clear()
	s.rect(0, 0, size.Width, size.Height, colBackground)

	left, top := plotPadLeft, plotPadTop
	right, bottom := size.Width-plotPadRight, size.Height-plotPadBottom
	if right <= left || bottom <= top {
		return
	}

	// Title centered at top; axis labels: x bottom-center, y top-left.
	if c.title != "" {
		s.text(0, 6, size.Width, titleFontSize+4, c.title, colText, titleFontSize, fyne.TextAlignCenter)
	}
	if c.xLabel != "" {
		s.text(0, size.Height-labelFontSize-4, size.Width, labelFontSize+2, c.xLabel, colAxisLabel, labelFontSize, fyne.TextAlignCenter)
	}
	if c.yLabel != "" {
		s.text(6, 6+titleFontSize+4, left*2, labelFontSize+2, c.yLabel, colAxisLabel, labelFontSize, fyne.TextAlignLeading)
	}

	// Scales: y pixels are inverted (lo = bottom, hi = top).
	c.xScale.setRange(left, right)
	c.yScale.setRange(bottom, top)

	// Grid + tick labels.
	yTicks := niceTicks(c.yScale.min, c.yScale.max, 7)
	for _, tv := range yTicks {
		py := c.yScale.toPixel(tv)
		if c.showGrid && py > top && py < bottom {
			s.line(left+1, py, right, py, colGrid, 1)
		}
		s.text(4, py-tickFontSize/2-1, left-8, tickFontSize+2, formatTick(tv), colAxisLabel, tickFontSize, fyne.TextAlignTrailing)
	}
	xPos, xLab := c.delegate.xTicks()
	for i, tv := range xPos {
		px := c.xScale.toPixel(tv)
		if px < left-0.5 || px > right+0.5 {
			continue
		}
		if c.showGrid && px > left && px < right && (i > 0 || len(xPos) > 2) {
			s.line(px, top, px, bottom-1, colGrid, 1)
		}
		if i < len(xLab) && xLab[i] != "" {
			s.text(px-30, bottom+4, 60, tickFontSize+2, xLab[i], colAxisLabel, tickFontSize, fyne.TextAlignCenter)
		}
	}

	// Frame drawn over grid so the plot area stays crisp.
	s.line(left, top, left, bottom, colFrame, 1)
	s.line(left, bottom, right, bottom, colFrame, 1)

	// Series.
	p := newPlotPainter(s, left, top, right, bottom)
	c.delegate.buildSeries(p, c.xScale, c.yScale, &seriesOpts{hoverIdx: c.hoverIdx, showValues: c.showValues})

	// Hover tooltip on top.
	if c.hoverIdx >= 0 && c.tipText != "" {
		c.drawTooltip(s, size)
	}
}

func (c *chartCore) drawTooltip(s *scene, size fyne.Size) {
	// Deterministic measurement so exports match on-screen rendering.
	w := 12 + float32(utf8.RuneCountInString(c.tipText))*tipFontSize*0.58
	h := tipFontSize + 10

	x := c.hoverPt.X + 14
	y := c.hoverPt.Y + 14
	if x+w > size.Width-2 {
		x = c.hoverPt.X - w - 10
	}
	if x < 2 {
		x = 2
	}
	if y+h > size.Height-2 {
		y = size.Height - h - 2
	}
	if y < 2 {
		y = 2
	}

	s.roundRect(x-1, y-1, w+2, h+2, 3, colTipBorder)
	s.roundRect(x, y, w, h, 3, colTipBg)
	s.text(x+6, y+4, w-12, h-8, c.tipText, colText, tipFontSize, fyne.TextAlignLeading)
}

// inPlot reports whether widget-local pixels are inside the plot area.
func (c *chartCore) inPlot(x, y float32) bool {
	xa, xb := c.xScale.lo, c.xScale.hi
	ya, yb := c.yScale.hi, c.yScale.lo // hi is the top (smaller pixel y)
	if xa > xb {
		xa, xb = xb, xa
	}
	if ya > yb {
		ya, yb = yb, ya
	}
	return x >= xa && x <= xb && y >= ya && y <= yb
}

func (c *chartCore) hit(px, py float32) (hitResult, bool) {
	if !c.inPlot(px, py) {
		return hitResult{}, false
	}
	return c.delegate.hitTest(px, py, c.xScale, c.yScale)
}

// autoTicks is the default continuous x-axis tick provider for delegates.
func autoTicks(s *linearScale) ([]float64, []string) {
	tk := niceTicks(s.min, s.max, 7)
	labels := make([]string, len(tk))
	for i, v := range tk {
		labels[i] = formatTick(v)
	}
	return tk, labels
}

// --- scene view: pooled Fyne canvas objects ---------------------------------

// sceneView keeps one canvas object per scene item, preserving draw order.
// Objects are reused across refreshes; only mismatched kinds are replaced.
type sceneView struct {
	objs  []fyne.CanvasObject
	kinds []sceneKind
}

func (v *sceneView) objects() []fyne.CanvasObject {
	return v.objs
}

func (v *sceneView) sync(s *scene) {
	items := s.items
	for i, it := range items {
		if i >= len(v.objs) || v.kinds[i] != it.kind {
			var obj fyne.CanvasObject
			switch it.kind {
			case kindRect:
				obj = canvas.NewRectangle(it.col)
			case kindLine:
				obj = canvas.NewLine(it.col)
			case kindText:
				obj = canvas.NewText(it.text, it.col)
			}
			v.insertAt(i, obj, it.kind)
		}
		obj := v.objs[i]
		switch it.kind {
		case kindRect:
			r := obj.(*canvas.Rectangle)
			r.FillColor = it.col
			r.CornerRadius = it.radius
			r.Move(fyne.NewPos(it.x, it.y))
			r.Resize(fyne.NewSize(it.w, it.h))
		case kindLine:
			l := obj.(*canvas.Line)
			l.StrokeColor = it.col
			l.StrokeWidth = it.lineW
			// Endpoints are stored in widget-local coordinates; Position()
			// derives the bounding-box origin, so set them directly.
			l.Position1 = fyne.NewPos(it.x, it.y)
			l.Position2 = fyne.NewPos(it.w, it.h)
		case kindText:
			t := obj.(*canvas.Text)
			t.Text = it.text
			t.Color = it.col
			t.Alignment = it.align
			if it.fontSz > 0 {
				t.TextSize = it.fontSz
			}
			t.Move(fyne.NewPos(it.x, it.y))
			t.Resize(fyne.NewSize(it.w, it.h))
			t.Refresh()
		}
	}
	if len(items) < len(v.objs) {
		v.objs = v.objs[:len(items)]
		v.kinds = v.kinds[:len(items)]
	}
}

func (v *sceneView) insertAt(i int, obj fyne.CanvasObject, kind sceneKind) {
	if i >= len(v.objs) {
		v.objs = append(v.objs, obj)
		v.kinds = append(v.kinds, kind)
		return
	}
	v.objs = append(v.objs, nil)
	copy(v.objs[i+1:], v.objs[i:])
	v.objs[i] = obj
	v.kinds = append(v.kinds, 0)
	copy(v.kinds[i+1:], v.kinds[i:])
	v.kinds[i] = kind
}

// --- export -----------------------------------------------------------------

// writeSVG serializes the chart at the current size (or an explicit one) as
// SVG markup. Export renders exactly the same scene the widget shows; the live
// scene is rebuilt on the next refresh, and pixel ranges are recomputed by
// every build, so scratch building is safe on the GUI thread.
func (c *chartCore) writeSVG(width, height float32) string {
	if width <= 0 || height <= 0 {
		width, height = c.Size().Width, c.Size().Height
	}
	if width <= 0 || height <= 0 {
		width, height = 800, 600
	}
	saved := c.sc
	c.sc = scene{}
	c.build(fyne.NewSize(width, height))
	svg := c.sc.toSVG(width, height)
	c.sc = saved
	return svg
}
