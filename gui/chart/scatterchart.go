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

var _ object.Object = &ScatterChart{}
var _ chartDelegate = &ScatterChart{}
var _ fyne.CanvasObject = &ScatterChart{}

const ScatterChartType object.Type = "chart.scatterchart"

// scatterSeries is one plotted series: a point cloud, an optional
// connecting polyline, or both. The primary series (index 0) holds the
// values given to NewScatterChart; AddSeries appends more.
//
// The per-point vectors (markerSizes, pointCols, pointLabels) are optional
// annotations: when non-empty they must match the point count and override
// the series defaults per point (a nil color or empty label falls back to
// the series style). They drive the chemometrics plot types: outlier
// emphasis (sizes/colors), sample names (labels).
type scatterSeries struct {
	xvals, yvals []float64
	col          color.Color
	markerSize   float32 // <= 0 hides markers
	lineWidth    float32 // > 0 connects the points in supply order

	markerSizes []float32     // optional per-point marker sizes
	pointCols   []color.Color // optional per-point colors
	pointLabels []string      // optional per-point text labels
}

// ScatterChart is an interactive multi-series scatter plot built from Fyne
// canvas primitives. Hovering highlights the nearest point and shows its
// coordinates; the mouse wheel zooms around the cursor and dragging pans
// (double-click resets). OnTap/OnHover/OnSecondaryTap expose the hit point
// to scripts, and SaveImage exports the chart as SVG.
type ScatterChart struct {
	*chartCore

	series []*scatterSeries
}

// NewScatterChart creates an interactive scatter plot with the given title,
// axis labels, and x/y values. Returns an error if x and y lengths mismatch.
func NewScatterChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*ScatterChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	s := &ScatterChart{
		series: []*scatterSeries{{
			xvals:      xvalues,
			yvals:      yvalues,
			col:        paletteColor(0), // steel blue
			markerSize: 6,
		}},
	}
	s.chartCore = newChartCore(string(ScatterChartType), s)
	s.title = title
	s.xLabel = xlabel
	s.yLabel = ylabel
	s.ExtendBaseWidget(s)
	return s, nil
}

// --- object.Object ----------------------------------------------------------

func (obj *ScatterChart) CanvasObject() fyne.CanvasObject {
	return obj
}

func (r *ScatterChart) Type() object.Type {
	return ScatterChartType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (r *ScatterChart) Inspect() string {
	return guithread.Inspect(r.inspectUI)
}

func (r *ScatterChart) inspectUI() string {
	return "chart.scatterchart"
}

func (r *ScatterChart) Interface() interface{} {
	return r
}

func (r *ScatterChart) IsTruthy() bool {
	return true
}

func (r *ScatterChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ScatterChartType, opType)
}

func (r *ScatterChart) Equals(other object.Object) bool {
	return r == other
}

func (*ScatterChart) Attrs() []object.AttrSpec {
	return append(scatterChartAttrs.Specs(), saveImageSpec)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (r *ScatterChart) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, r.setAttrUI)
}

func (r *ScatterChart) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ScatterChartType, name)
}

// --- chartDelegate ----------------------------------------------------------

func (s *ScatterChart) dataExtent() (float64, float64, float64, float64) {
	seeded := false
	var xmin, xmax, ymin, ymax float64
	for _, sr := range s.series {
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

func (s *ScatterChart) xTicks() ([]float64, []string) {
	return autoTicks(s.xScale)
}

func (s *ScatterChart) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	for si, sr := range s.series {
		if len(sr.xvals) == 0 {
			continue
		}
		pxs := make([]float32, len(sr.xvals))
		pys := make([]float32, len(sr.xvals))
		for i := range sr.xvals {
			pxs[i] = xs.toPixel(sr.xvals[i])
			pys[i] = ys.toPixel(sr.yvals[i])
		}
		if sr.lineWidth > 0 {
			p.polyline(pxs, pys, sr.col, sr.lineWidth)
		}
		if sr.markerSize > 0 || len(sr.markerSizes) > 0 {
			for i := range pxs {
				size, col := sr.markerSize, sr.col
				if i < len(sr.markerSizes) {
					size = sr.markerSizes[i]
				}
				if i < len(sr.pointCols) && sr.pointCols[i] != nil {
					col = sr.pointCols[i]
				}
				if si == opts.hoverSeries && i == opts.hoverIdx {
					size = size * 1.6
					col = darken(col, 0.72)
				}
				if size > 0 {
					p.rect(pxs[i]-size/2, pys[i]-size/2, size, size, col)
				}

				lbl := ""
				if i < len(sr.pointLabels) {
					lbl = sr.pointLabels[i]
				}
				if lbl != "" || opts.showValues {
					if lbl == "" {
						lbl = "(" + formatTick(sr.xvals[i]) + ", " + formatTick(sr.yvals[i]) + ")"
					}
					// Near the plot top the above-point label anchor would be
					// clipped by the painter, so flip the label below the point.
					ly := pys[i] - 14
					if pys[i]-14 < p.top+2 {
						ly = pys[i] + 10
					}
					p.text(pxs[i]+size, ly, 80, 12, lbl, colText, 10, fyne.TextAlignLeading)
				}
			}
		}
	}
}

func (s *ScatterChart) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	// Nearest point within a small radius across all series; linear scan is
	// a handful of microseconds even for tens of thousands of points.
	best, bestS, bestD := -1, -1, float32(14)
	for si, sr := range s.series {
		for i := range sr.xvals {
			dx := xs.toPixel(sr.xvals[i]) - px
			dy := ys.toPixel(sr.yvals[i]) - py
			d := float32(math.Hypot(float64(dx), float64(dy)))
			if d < bestD {
				bestD, best, bestS = d, i, si
			}
		}
	}
	if bestS < 0 {
		return hitResult{}, false
	}
	sr := s.series[bestS]
	return hitResult{
		Series:   bestS,
		Index:    best,
		Value:    sr.yvals[best],
		ValueStr: "(" + formatTick(sr.xvals[best]) + ", " + formatTick(sr.yvals[best]) + ")",
	}, true
}

// ellipsePoints samples a closed ellipse polyline (steps+1 points, first
// equals last) centered at (cx,cy) with the given data-unit radii.
func ellipsePoints(cx, cy, rx, ry float64, steps int) ([]float64, []float64) {
	xs := make([]float64, steps+1)
	ys := make([]float64, steps+1)
	for k := 0; k <= steps; k++ {
		t := 2 * math.Pi * float64(k) / float64(steps)
		xs[k] = cx + rx*math.Cos(t)
		ys[k] = cy + ry*math.Sin(t)
	}
	// Close exactly: sin/cos(2*pi) are not bit-exact roundings of 0/1.
	xs[steps], ys[steps] = xs[0], ys[0]
	return xs, ys
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *ScatterChart) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

var scatterChartAttrs = objattrs.Register(string(ScatterChartType), object.NewAttrRegistry[*ScatterChart](string(ScatterChartType)))

func (g *ScatterChart) getAttrUI(name string) (object.Object, bool) {
	if o, ok := scatterChartAttrs.GetAttr(g, name); ok {
		return o, true
	}
	return g.chartCore.getAttr(name)
}

// scatterSeriesStyle parses the optional trailing (color, lineWidth) args of
// AddSeries. defCol is the palette default for the new series index.
func scatterSeriesStyle(name string, args []object.Object, defCol color.Color) (*scatterSeries, error) {
	sr := &scatterSeries{col: defCol, markerSize: 6}
	if len(args) > 0 {
		col, err := colorArg(args[0])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		sr.col = col
	}
	if len(args) > 1 {
		f, err := object.AsFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("%s: line width must be a number (got %s)", name, args[1].Type())
		}
		if f < 0 || f > 20 {
			return nil, fmt.Errorf("chart error: %s: line width must be in [0, 20] (got %v)", name, f)
		}
		sr.lineWidth = float32(f)
	}
	return sr, nil
}

func init() {
	objattrs.Add(scatterChartAttrs)
	addInteractive(scatterChartAttrs)
	scatterChartAttrs.Define("SetPoints").
		Doc("Replace the primary series' x and y values.").
		Args("xvalues", "yvalues").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
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
	scatterChartAttrs.Define("AddSeries").
		Doc("Add a new series with the given x and y values; optional color and line width.").
		Args("xvalues", "yvalues").OptionalArg("color").OptionalArg("lineWidth").
		Returns("int, the new series index").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
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
			if derr := guithread.DoAndWaitContext(ctx, func() {
				sr, serr := scatterSeriesStyle("AddSeries", args[2:], paletteColor(len(g.series)))
				if serr != nil {
					err = serr
					return
				}
				sr.xvals, sr.yvals = xs, ys
				g.series = append(g.series, sr)
				idx = len(g.series) - 1
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.NewInt(int64(idx)), nil
		})
	scatterChartAttrs.Define("ClearSeries").
		Doc("Remove all series except the primary one.").
		Returns("nil").
		Impl(func(g *ScatterChart, _ context.Context, _ ...object.Object) (object.Object, error) {
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
	scatterChartAttrs.Define("SeriesCount").
		Doc("Number of plotted series (>= 1).").
		Returns("int").
		Impl(func(g *ScatterChart, ctx context.Context, _ ...object.Object) (object.Object, error) {
			var n int
			if err := guithread.DoAndWaitContext(ctx, func() { n = len(g.series) }); err != nil {
				return nil, err
			}
			return object.NewInt(int64(n)), nil
		})
	scatterChartAttrs.Define("SetSeriesColor").
		Doc("Set one series' color by name or #rrggbb.").
		Args("index", "color").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			col, err := colorArg(args[1])
			if err != nil {
				return nil, err
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesColor", args[0], len(g.series)); err != nil {
					return
				}
				g.series[i].col = col
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	scatterChartAttrs.Define("SetSeriesMarkerSize").
		Doc("Set one series' marker size; 0 hides its markers.").
		Args("index", "size").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			if f < 0 || f > 30 {
				return nil, fmt.Errorf("chart error: marker size must be in [0, 30] (got %v)", f)
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesMarkerSize", args[0], len(g.series)); err != nil {
					return
				}
				g.series[i].markerSize = float32(f)
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	scatterChartAttrs.Define("SetSeriesLineWidth").
		Doc("Connect one series' points with a line of the given width; 0 disables.").
		Args("index", "width").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			if f < 0 || f > 20 {
				return nil, fmt.Errorf("chart error: line width must be in [0, 20] (got %v)", f)
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesLineWidth", args[0], len(g.series)); err != nil {
					return
				}
				g.series[i].lineWidth = float32(f)
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	scatterChartAttrs.Define("SetSeriesMarkerSizes").
		Doc("Per-point marker sizes for one series (overrides SetSeriesMarkerSize per point; 0 hides a point).").
		Args("index", "sizes").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			sizes, err := floatSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesMarkerSizes", args[0], len(g.series)); err != nil {
					return
				}
				if len(sizes) != len(g.series[i].xvals) {
					err = fmt.Errorf("chart error: SetSeriesMarkerSizes: %d sizes for %d points", len(sizes), len(g.series[i].xvals))
					return
				}
				for _, s := range sizes {
					if s < 0 || s > 30 {
						err = fmt.Errorf("chart error: SetSeriesMarkerSizes: sizes must be in [0, 30] (got %v)", s)
						return
					}
				}
				g.series[i].markerSizes = make([]float32, len(sizes))
				for k, s := range sizes {
					g.series[i].markerSizes[k] = float32(s)
				}
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	scatterChartAttrs.Define("SetSeriesPointColors").
		Doc("Per-point colors for one series (name or #rrggbb; empty string uses the series color).").
		Args("index", "colors").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			strs, err := stringSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			cols := make([]color.Color, len(strs))
			for k, s := range strs {
				if s == "" {
					continue
				}
				c, cerr := colorArg(object.NewString(s))
				if cerr != nil {
					return nil, fmt.Errorf("SetSeriesPointColors: %w", cerr)
				}
				cols[k] = c
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesPointColors", args[0], len(g.series)); err != nil {
					return
				}
				if len(cols) != len(g.series[i].xvals) {
					err = fmt.Errorf("chart error: SetSeriesPointColors: %d colors for %d points", len(cols), len(g.series[i].xvals))
					return
				}
				g.series[i].pointCols = cols
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	scatterChartAttrs.Define("SetSeriesPointLabels").
		Doc("Per-point text labels for one series (empty strings are skipped).").
		Args("index", "labels").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			labels, err := stringSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			if derr := guithread.DoAndWaitContext(ctx, func() {
				var i int
				if i, err = seriesIndexArg("SetSeriesPointLabels", args[0], len(g.series)); err != nil {
					return
				}
				if len(labels) != len(g.series[i].xvals) {
					err = fmt.Errorf("chart error: SetSeriesPointLabels: %d labels for %d points", len(labels), len(g.series[i].xvals))
					return
				}
				g.series[i].pointLabels = labels
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	// Reference geometries: horizontal/vertical limits and ellipses land as
	// pure polyline series (markers off), so extent, hover and SVG export
	// all treat them like any other series.
	scatterChartAttrs.Define("AddHLine").
		Doc("Horizontal reference line at y across the data range; returns the series index.").
		Args("y").OptionalArg("color").OptionalArg("lineWidth").
		Returns("int").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			yv, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			var idx int
			if derr := guithread.DoAndWaitContext(ctx, func() {
				x0, x1, _, _ := g.dataExtent()
				sr, serr := scatterSeriesStyle("AddHLine", args[1:], colorRed)
				if serr != nil {
					err = serr
					return
				}
				sr.markerSize = 0
				if sr.lineWidth == 0 {
					sr.lineWidth = 1.5
				}
				sr.xvals, sr.yvals = []float64{x0, x1}, []float64{yv, yv}
				g.series = append(g.series, sr)
				idx = len(g.series) - 1
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.NewInt(int64(idx)), nil
		})
	scatterChartAttrs.Define("AddVLine").
		Doc("Vertical reference line at x across the data range; returns the series index.").
		Args("x").OptionalArg("color").OptionalArg("lineWidth").
		Returns("int").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			xv, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			var idx int
			if derr := guithread.DoAndWaitContext(ctx, func() {
				_, _, y0, y1 := g.dataExtent()
				sr, serr := scatterSeriesStyle("AddVLine", args[1:], colorRed)
				if serr != nil {
					err = serr
					return
				}
				sr.markerSize = 0
				if sr.lineWidth == 0 {
					sr.lineWidth = 1.5
				}
				sr.xvals, sr.yvals = []float64{xv, xv}, []float64{y0, y1}
				g.series = append(g.series, sr)
				idx = len(g.series) - 1
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.NewInt(int64(idx)), nil
		})
	scatterChartAttrs.Define("AddEllipse").
		Doc("Ellipse centered at (cx,cy) with data-space radii rx/ry (e.g. a Hotelling T2 confidence ellipse); returns the series index.").
		Args("cx", "cy", "rx", "ry").OptionalArg("color").OptionalArg("lineWidth").
		Returns("int").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			cx, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			cy, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			rx, err := object.AsFloat(args[2])
			if err != nil {
				return nil, err
			}
			ry, err := object.AsFloat(args[3])
			if err != nil {
				return nil, err
			}
			if rx <= 0 || ry <= 0 {
				return nil, fmt.Errorf("chart error: AddEllipse: radii must be > 0 (got %v, %v)", rx, ry)
			}
			var idx int
			if derr := guithread.DoAndWaitContext(ctx, func() {
				sr, serr := scatterSeriesStyle("AddEllipse", args[4:], paletteColor(len(g.series)))
				if serr != nil {
					err = serr
					return
				}
				sr.markerSize = 0
				if sr.lineWidth == 0 {
					sr.lineWidth = 1.5
				}
				sr.xvals, sr.yvals = ellipsePoints(cx, cy, rx, ry, 64)
				g.series = append(g.series, sr)
				idx = len(g.series) - 1
				g.recomputeDomains()
				g.refresh()
			}); derr != nil {
				return nil, derr
			}
			if err != nil {
				return nil, err
			}
			return object.NewInt(int64(idx)), nil
		})
	// Primary-series shorthands (series 0), kept for compatibility.
	scatterChartAttrs.Define("SetColor").
		Doc("Set the primary series' color (name or #rrggbb).").
		Arg("color").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
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
	scatterChartAttrs.Define("SetMarkerSize").
		Doc("Set the primary series' marker size; 0 hides markers.").
		Arg("size").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			if f < 0 || f > 30 {
				return nil, fmt.Errorf("chart error: marker size must be in [0, 30] (got %v)", f)
			}
			guithread.Do(func() {
				g.series[0].markerSize = float32(f)
				g.refresh()
			})
			return object.Nil, nil
		})
}
