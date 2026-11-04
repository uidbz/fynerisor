package risorchart

import (
	"context"
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

var _ object.Object = &LineChart{}
var _ chartDelegate = &LineChart{}
var _ fyne.CanvasObject = &LineChart{}

const LineChartType object.Type = "chart.linechart"

// lineSeries is one plotted polyline. The primary series (index 0) holds the
// values given to NewLineChart; AddSeries appends more.
type lineSeries struct {
	xvals, yvals []float64
	col          color.Color
	width        float32 // <= 0 hides the series stroke
}

// LineChart is an interactive multi-series line chart built from Fyne canvas
// primitives. Hovering snaps to the nearest x sample and shows its value;
// wheel zoom and drag pan work on both axes (double-click resets). Points
// are connected in the order supplied, so unsorted x values draw a path,
// not a scatter join.
type LineChart struct {
	*chartCore

	series []*lineSeries
}

// NewLineChart creates an interactive line chart with the given title, axis
// labels, and x/y values. Returns an error if x and y lengths mismatch.
func NewLineChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*LineChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	l := &LineChart{
		series: []*lineSeries{{
			xvals: xvalues,
			yvals: yvalues,
			col:   paletteColor(0), // steel blue
			width: 2,
		}},
	}
	l.chartCore = newChartCore(string(LineChartType), l)
	l.title = title
	l.xLabel = xlabel
	l.yLabel = ylabel
	l.ExtendBaseWidget(l)
	return l, nil
}

// --- object.Object ----------------------------------------------------------

func (obj *LineChart) CanvasObject() fyne.CanvasObject {
	return obj
}

func (r *LineChart) Type() object.Type {
	return LineChartType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (r *LineChart) Inspect() string {
	return guithread.Inspect(r.inspectUI)
}

func (r *LineChart) inspectUI() string {
	return "chart.linechart"
}

func (r *LineChart) Interface() interface{} {
	return r
}

func (r *LineChart) IsTruthy() bool {
	return true
}

func (r *LineChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", LineChartType, opType)
}

func (r *LineChart) Equals(other object.Object) bool {
	return r == other
}

func (*LineChart) Attrs() []object.AttrSpec {
	return append(lineChartAttrs.Specs(), saveImageSpec)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (r *LineChart) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, r.setAttrUI)
}

func (r *LineChart) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LineChartType, name)
}

// --- chartDelegate ----------------------------------------------------------

func (l *LineChart) dataExtent() (float64, float64, float64, float64) {
	seeded := false
	var xmin, xmax, ymin, ymax float64
	for _, sr := range l.series {
		for i := range sr.xvals {
			x, y := sr.xvals[i], sr.yvals[i]
			if !seeded {
				xmin, xmax, ymin, ymax = x, x, y, y
				seeded = true
				continue
			}
			if x < xmin {
				xmin = x
			}
			if x > xmax {
				xmax = x
			}
			if y < ymin {
				ymin = y
			}
			if y > ymax {
				ymax = y
			}
		}
	}
	if !seeded {
		return 0, 1, 0, 1
	}
	xpad := (xmax - xmin) * 0.05
	ypad := (ymax - ymin) * 0.05
	return xmin - xpad, xmax + xpad, ymin - ypad, ymax + ypad
}

func (l *LineChart) xTicks() ([]float64, []string) {
	return autoTicks(l.xScale)
}

func (l *LineChart) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	for si, sr := range l.series {
		if len(sr.xvals) == 0 {
			continue
		}
		pxs := make([]float32, len(sr.xvals))
		pys := make([]float32, len(sr.xvals))
		for i := range sr.xvals {
			pxs[i] = xs.toPixel(sr.xvals[i])
			pys[i] = ys.toPixel(sr.yvals[i])
		}
		if sr.width > 0 {
			p.polyline(pxs, pys, sr.col, sr.width)
		}

		if si == opts.hoverSeries && opts.hoverIdx >= 0 && opts.hoverIdx < len(pxs) {
			// Small highlight marker on the hovered sample.
			p.rect(pxs[opts.hoverIdx]-4, pys[opts.hoverIdx]-4, 8, 8, darken(sr.col, 0.72))
		}
		if opts.showValues {
			for i := range sr.xvals {
				p.text(pxs[i]-30, pys[i]-16, 60, 12, formatTick(sr.yvals[i]), colText, 10, fyne.TextAlignCenter)
			}
		}
	}
}

func (l *LineChart) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	// For each series snap to the nearest x sample, then keep the series
	// whose snapped point is vertically closest to the pointer. Far-away
	// vertical distance disqualifies a series.
	x := float64(xs.fromPixel(px))
	bestS, bestI := -1, -1
	bestNear := math.Inf(1)
	for si, sr := range l.series {
		bi, bd := -1, math.Inf(1)
		for i := range sr.xvals {
			d := math.Abs(sr.xvals[i] - x)
			if d < bd {
				bd, bi = d, i
			}
		}
		if bi < 0 {
			continue
		}
		near := math.Abs(float64(ys.toPixel(sr.yvals[bi]) - py))
		if near > 60 {
			continue
		}
		if near < bestNear {
			bestNear, bestS, bestI = near, si, bi
		}
	}
	if bestS < 0 {
		return hitResult{}, false
	}
	sr := l.series[bestS]
	return hitResult{
		Series:   bestS,
		Index:    bestI,
		Value:    sr.yvals[bestI],
		ValueStr: "(" + formatTick(sr.xvals[bestI]) + ", " + formatTick(sr.yvals[bestI]) + ")",
	}, true
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *LineChart) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

var lineChartAttrs = objattrs.Register(string(LineChartType), object.NewAttrRegistry[*LineChart](string(LineChartType)))

func (g *LineChart) getAttrUI(name string) (object.Object, bool) {
	if o, ok := lineChartAttrs.GetAttr(g, name); ok {
		return o, true
	}
	return g.chartCore.getAttr(name)
}

func init() {
	objattrs.Add(lineChartAttrs)
	addInteractive(lineChartAttrs)
	lineChartAttrs.Define("SetPoints").
		Doc("Replace the primary series' x and y values.").
		Args("xvalues", "yvalues").Returns("nil").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			xs, err := floatSliceArg(args[0])
			if err != nil {
				return nil, err
			}
			ys, err := floatSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			if len(xs) != len(ys) {
				return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xs), len(ys))
			}
			guithread.Do(func() {
				g.series[0].xvals, g.series[0].yvals = xs, ys
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			})
			return object.Nil, nil
		})
	lineChartAttrs.Define("AddSeries").
		Doc("Add a new series with the given x and y values; optional color and line width.").
		Args("xvalues", "yvalues").OptionalArg("color").OptionalArg("lineWidth").
		Returns("int, the new series index").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			xs, err := floatSliceArg(args[0])
			if err != nil {
				return nil, err
			}
			ys, err := floatSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			if len(xs) != len(ys) {
				return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xs), len(ys))
			}
			var idx int
			if err := guithread.DoAndWaitContext(ctx, func() {
				sr := &lineSeries{col: paletteColor(len(g.series)), width: 2}
				if len(args) > 2 {
					col, cerr := colorArg(args[2])
					if cerr != nil {
						err = fmt.Errorf("AddSeries: %w", cerr)
						return
					}
					sr.col = col
				}
				if len(args) > 3 {
					f, ferr := object.AsFloat(args[3])
					if ferr != nil {
						err = fmt.Errorf("AddSeries: line width must be a number (got %s)", args[3].Type())
						return
					}
					if f < 0 || f > 20 {
						err = fmt.Errorf("chart error: AddSeries: line width must be in [0, 20] (got %v)", f)
						return
					}
					sr.width = float32(f)
				}
				sr.xvals, sr.yvals = xs, ys
				g.series = append(g.series, sr)
				idx = len(g.series) - 1
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			}); err != nil {
				return nil, err
			}
			return object.NewInt(int64(idx)), nil
		})
	lineChartAttrs.Define("ClearSeries").
		Doc("Remove all series except the primary one.").
		Returns("nil").
		Impl(func(g *LineChart, _ context.Context, _ ...object.Object) (object.Object, error) {
			guithread.Do(func() {
				if len(g.series) > 1 {
					g.series = g.series[:1]
					g.clearHover()
					g.recomputeDomains()
					g.refresh()
				}
			})
			return object.Nil, nil
		})
	lineChartAttrs.Define("SeriesCount").
		Doc("Number of plotted series (>= 1).").
		Returns("int").
		Impl(func(g *LineChart, ctx context.Context, _ ...object.Object) (object.Object, error) {
			var n int
			if err := guithread.DoAndWaitContext(ctx, func() { n = len(g.series) }); err != nil {
				return nil, err
			}
			return object.NewInt(int64(n)), nil
		})
	lineChartAttrs.Define("SetSeriesColor").
		Doc("Set one series' color by name or #rrggbb.").
		Args("index", "color").Returns("nil").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			col, err := colorArg(args[1])
			if err != nil {
				return nil, err
			}
			if err := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesColor", args[0], len(g.series)); err != nil {
					return
				}
				g.series[i].col = col
				g.refresh()
			}); err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	lineChartAttrs.Define("SetSeriesLineWidth").
		Doc("Set one series' stroke width; 0 hides the line.").
		Args("index", "width").Returns("nil").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			if f < 0 || f > 20 {
				return nil, fmt.Errorf("chart error: line width must be in [0, 20] (got %v)", f)
			}
			if err := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesLineWidth", args[0], len(g.series)); err != nil {
					return
				}
				g.series[i].width = float32(f)
				g.refresh()
			}); err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	// Primary-series shorthands (series 0), kept for compatibility.
	lineChartAttrs.Define("SetColor").
		Doc("Set the primary series' color (name or #rrggbb).").
		Arg("color").Returns("nil").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			col, err := colorArg(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				g.series[0].col = col
				g.refresh()
			})
			return object.Nil, nil
		})
	lineChartAttrs.Define("SetLineWidth").
		Doc("Set the primary series' stroke width.").
		Arg("width").Returns("nil").
		Impl(func(g *LineChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			if f < 0.5 || f > 20 {
				return nil, fmt.Errorf("chart error: line width must be in [0.5, 20] (got %v)", f)
			}
			guithread.Do(func() {
				g.series[0].width = float32(f)
				g.refresh()
			})
			return object.Nil, nil
		})
}
