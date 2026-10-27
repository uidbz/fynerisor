package risorchart

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
)

var _ object.Object = &LineChart{}
var _ chartDelegate = &LineChart{}
var _ fyne.CanvasObject = &LineChart{}

const LineChartType object.Type = "chart.linechart"

// LineChart is an interactive line chart built from Fyne canvas primitives.
// Hovering snaps to the nearest x sample and shows its value; wheel zoom and
// drag pan work on both axes (double-click resets). Points are connected in
// the order supplied, so unsorted x values draw a path, not a scatter join.
type LineChart struct {
	*chartCore

	xvals, yvals []float64
	color        color.Color
	width        float32
}

// NewLineChart creates an interactive line chart with the given title, axis
// labels, and x/y values. Returns an error if x and y lengths mismatch.
func NewLineChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*LineChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	l := &LineChart{
		xvals: xvalues,
		yvals: yvalues,
		color: color.NRGBA{R: 70, G: 130, B: 180, A: 255}, // steel blue
		width: 2,
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

func (r *LineChart) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(LineChartType) + "'")
}

func (r *LineChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(LineChartType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", LineChartType, opType)

	return obj, err
}

func (r *LineChart) Equals(other object.Object) bool {
	return r == other
}

func (r *LineChart) Attrs() []object.AttrSpec {
	return nil
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
	if len(l.xvals) == 0 {
		return 0, 1, 0, 1
	}
	xmin, xmax := l.xvals[0], l.xvals[0]
	ymin, ymax := l.yvals[0], l.yvals[0]
	for i := range l.xvals {
		if l.xvals[i] < xmin {
			xmin = l.xvals[i]
		}
		if l.xvals[i] > xmax {
			xmax = l.xvals[i]
		}
		if l.yvals[i] < ymin {
			ymin = l.yvals[i]
		}
		if l.yvals[i] > ymax {
			ymax = l.yvals[i]
		}
	}
	xpad := (xmax - xmin) * 0.05
	ypad := (ymax - ymin) * 0.05
	return xmin - xpad, xmax + xpad, ymin - ypad, ymax + ypad
}

func (l *LineChart) xTicks() ([]float64, []string) {
	return autoTicks(l.xScale)
}

func (l *LineChart) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	if len(l.xvals) == 0 {
		return
	}
	pxs := make([]float32, len(l.xvals))
	pys := make([]float32, len(l.xvals))
	for i := range l.xvals {
		pxs[i] = xs.toPixel(l.xvals[i])
		pys[i] = ys.toPixel(l.yvals[i])
	}
	p.polyline(pxs, pys, l.color, l.width)

	if opts.hoverIdx >= 0 && opts.hoverIdx < len(pxs) {
		// Small highlight marker on the hovered sample.
		p.rect(pxs[opts.hoverIdx]-4, pys[opts.hoverIdx]-4, 8, 8, darken(l.color, 0.72))
	}
	if opts.showValues {
		for i := range l.xvals {
			p.text(pxs[i]-30, pys[i]-16, 60, 12, formatTick(l.yvals[i]), colText, 10, fyne.TextAlignCenter)
		}
	}
}

func (l *LineChart) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	if len(l.xvals) == 0 {
		return hitResult{}, false
	}
	// Snap to nearest sample by x, but ignore far-away vertical distance.
	best, bestD := -1, math.MaxFloat32
	x := float64(xs.fromPixel(px))
	for i := range l.xvals {
		d := math.Abs(l.xvals[i] - x)
		if d < bestD {
			bestD, best = d, i
		}
	}
	if best < 0 {
		return hitResult{}, false
	}
	pyNear := math.Abs(float64(ys.toPixel(l.yvals[best]) - py))
	if pyNear > 60 {
		return hitResult{}, false
	}
	return hitResult{
		Index:    best,
		Value:    l.yvals[best],
		ValueStr: "(" + formatTick(l.xvals[best]) + ", " + formatTick(l.yvals[best]) + ")",
	}, true
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *LineChart) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

func (g *LineChart) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "SetPoints":
		return object.NewBuiltin("chart.linechart.SetPoints", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
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
				g.xvals, g.yvals = xs, ys
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			})
			return object.Nil, nil
		}), true
	case "SetColor":
		return object.NewBuiltin("chart.linechart.SetColor", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			col, err := colorArg(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				g.color = col
				g.refresh()
			})
			return object.Nil, nil
		}), true
	case "SetLineWidth":
		return object.NewBuiltin("chart.linechart.SetLineWidth", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			f, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			if f < 0.5 || f > 20 {
				return nil, fmt.Errorf("chart error: line width must be in [0.5, 20] (got %v)", f)
			}
			guithread.Do(func() {
				g.width = float32(f)
				g.refresh()
			})
			return object.Nil, nil
		}), true
	}
	return g.chartCore.getAttr(name)
}
