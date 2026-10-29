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

// ScatterChart is an interactive scatter plot built from Fyne canvas
// primitives. Hovering highlights the nearest point and shows its
// coordinates; the mouse wheel zooms around the cursor and dragging pans
// (double-click resets). OnTap/OnHover/OnSecondaryTap expose the hit point
// to scripts, and SaveImage exports the chart as SVG.
type ScatterChart struct {
	*chartCore

	xvals, yvals []float64
	color        color.Color
	markerSize   float32
}

// NewScatterChart creates an interactive scatter plot with the given title,
// axis labels, and x/y values. Returns an error if x and y lengths mismatch.
func NewScatterChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*ScatterChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	s := &ScatterChart{
		xvals:      xvalues,
		yvals:      yvalues,
		color:      color.NRGBA{R: 70, G: 130, B: 180, A: 255}, // steel blue
		markerSize: 6,
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
	if len(s.xvals) == 0 {
		return 0, 1, 0, 1
	}
	xmin, xmax := s.xvals[0], s.xvals[0]
	ymin, ymax := s.yvals[0], s.yvals[0]
	for i := range s.xvals {
		if s.xvals[i] < xmin {
			xmin = s.xvals[i]
		}
		if s.xvals[i] > xmax {
			xmax = s.xvals[i]
		}
		if s.yvals[i] < ymin {
			ymin = s.yvals[i]
		}
		if s.yvals[i] > ymax {
			ymax = s.yvals[i]
		}
	}
	xpad := (xmax - xmin) * 0.05
	ypad := (ymax - ymin) * 0.05
	return xmin - xpad, xmax + xpad, ymin - ypad, ymax + ypad
}

func (s *ScatterChart) xTicks() ([]float64, []string) {
	return autoTicks(s.xScale)
}

func (s *ScatterChart) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	for i := range s.xvals {
		px := xs.toPixel(s.xvals[i])
		py := ys.toPixel(s.yvals[i])
		size := s.markerSize
		col := s.color
		if i == opts.hoverIdx {
			size = s.markerSize * 1.6
			col = darken(col, 0.72)
		}
		p.rect(px-size/2, py-size/2, size, size, col)

		if opts.showValues {
			lbl := "(" + formatTick(s.xvals[i]) + ", " + formatTick(s.yvals[i]) + ")"
			p.text(px+size, py-14, 80, 12, lbl, colText, 10, fyne.TextAlignLeading)
		}
	}
}

func (s *ScatterChart) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	// Nearest point within a small radius; linear scan is a handful of
	// microseconds even for tens of thousands of points.
	best, bestD := -1, float32(14)
	for i := range s.xvals {
		dx := xs.toPixel(s.xvals[i]) - px
		dy := ys.toPixel(s.yvals[i]) - py
		d := float32(math.Hypot(float64(dx), float64(dy)))
		if d < bestD {
			bestD, best = d, i
		}
	}
	if best < 0 {
		return hitResult{}, false
	}
	return hitResult{
		Index:    best,
		Value:    s.yvals[best],
		ValueStr: "(" + formatTick(s.xvals[best]) + ", " + formatTick(s.yvals[best]) + ")",
	}, true
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
func init() {
	objattrs.Add(scatterChartAttrs)
	addInteractive(scatterChartAttrs)
	scatterChartAttrs.Define("SetPoints").Args("arg1",
		"arg2").Returns("nil").Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
		xs, err := floatSliceArg(args[0])
		if err != nil {
			return nil,
				err
		}
		ys, err := floatSliceArg(args[1])
		if err != nil {
			return nil, err
		}
		if len(xs) != len(ys) {
			return nil,

				fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)",

					len(xs), len(ys))
		}
		guithread.Do(
			func() {
				g.xvals, g.yvals = xs, ys
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			})
		return object.Nil, nil
	})
	scatterChartAttrs.Define("SetColor").Args("arg1").Returns("nil").Impl(
		func(g *ScatterChart,

			ctx context.Context, args ...object.Object) (object.
			Object, error) {
			col, err := colorArg(args[0])
			if err != nil {
				return nil, err
			}
			guithread.
				Do(func() {
					g.color = col
					g.
						refresh()
				})
			return object.Nil, nil
		},
	)
	scatterChartAttrs.Define("SetMarkerSize").Args(
		"arg1").Returns("nil").
		Impl(func(g *ScatterChart, ctx context.Context, args ...object.Object) (object.Object, error) {
			f, err := object.
				AsFloat(args[0])
			if err != nil {
				return nil,

					err
			}
			if f < 1 || f > 30 {
				return nil,
					fmt.Errorf("chart error: marker size must be in [1, 30] (got %v)",
						f)
			}
			guithread.
				Do(func() { g.markerSize = float32(f); g.refresh() })
			return object.Nil, nil
		})
}
