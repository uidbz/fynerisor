package risorchart

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/canvas"
	"github.com/uidbz/fynerisor/gui/guithread"
)

var _ object.Object = &BarChart{}
var _ chartDelegate = &BarChart{}
var _ fyne.CanvasObject = &BarChart{}

const BarChartType object.Type = "chart.barchart"

// BarChart is an interactive bar chart widget built from Fyne canvas
// primitives. Hovering a bar highlights it and shows a label/value tooltip;
// clicking fires the OnTap callback; right-click fires OnSecondaryTap; the
// mouse wheel zooms and dragging pans (both clamped to the data domain), and
// double-click resets the view. Data and appearance can be changed live.
//
// The same engine renders SVG exports, so bar.SaveImage("out.svg") produces
// exactly what is on screen.
type BarChart struct {
	*chartCore

	labels    []string
	values    []float64
	colors    []color.Color // per-bar overrides, nil = baseColor
	baseColor color.Color
}

// NewBarChart creates an interactive bar chart with the given title, y-axis
// label, x-axis labels and values. Returns an error if labels and values
// lengths do not match.
func NewBarChart(title, ylabel string, labels []string, values []float64) (*BarChart, error) {
	if len(labels) != len(values) {
		return nil, fmt.Errorf("chart error: labels and values must have same length (got %d labels, %d values)", len(labels), len(values))
	}

	b := &BarChart{
		labels:    labels,
		values:    values,
		baseColor: color.NRGBA{R: 70, G: 130, B: 180, A: 255}, // steel blue
	}
	b.chartCore = newChartCore(string(BarChartType), b)
	b.title = title
	b.yLabel = ylabel
	b.ExtendBaseWidget(b)
	return b, nil
}

// --- object.Object ----------------------------------------------------------

func (obj *BarChart) CanvasObject() fyne.CanvasObject {
	return obj
}

func (r *BarChart) Type() object.Type {
	return BarChartType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (r *BarChart) Inspect() string {
	return guithread.Inspect(r.inspectUI)
}

func (r *BarChart) inspectUI() string {
	return "chart.barchart"
}

func (r *BarChart) Interface() interface{} {
	return r
}

func (r *BarChart) IsTruthy() bool {
	return true
}

func (r *BarChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", BarChartType, opType)
}

func (r *BarChart) Equals(other object.Object) bool {
	return r == other
}

func (r *BarChart) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (r *BarChart) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, r.setAttrUI)
}

func (r *BarChart) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BarChartType, name)
}

// --- chartDelegate ----------------------------------------------------------

func (b *BarChart) dataExtent() (float64, float64, float64, float64) {
	n := float64(len(b.values))
	x0, x1 := 0.0, n
	if n == 0 {
		return 0, 1, -1, 1
	}
	minV, maxV := 0.0, 0.0
	for i, v := range b.values {
		if i == 0 || v < minV {
			minV = v
		}
		if i == 0 || v > maxV {
			maxV = v
		}
	}
	// Bars start at zero, so the domain always includes 0, with a little headroom.
	if minV > 0 {
		minV = 0
	}
	if maxV < 0 {
		maxV = 0
	}
	if minV == 0 && maxV == 0 {
		maxV = 1
	}
	pad := (maxV - minV) * 0.06
	return x0, x1, minV - pad*ifNeg(minV), maxV + pad
}

// ifNeg returns 1 when v < 0 (padding applies below a negative baseline).
func ifNeg(v float64) float64 {
	if v < 0 {
		return 1
	}
	return 0
}

func (b *BarChart) xTicks() ([]float64, []string) {
	pos := make([]float64, len(b.labels))
	for i := range b.labels {
		pos[i] = float64(i) + 0.5
	}
	return pos, b.labels
}

func (b *BarChart) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	n := len(b.values)
	if n == 0 {
		return
	}
	unit := xs.toPixel(1) - xs.toPixel(0) // pixel width of one category slot
	barW := unit * 0.7
	baseline := ys.toPixel(0)

	for i, v := range b.values {
		col := b.baseColor
		if i < len(b.colors) && b.colors[i] != nil {
			col = b.colors[i]
		}
		if i == opts.hoverIdx {
			col = darken(col, 0.72)
		}
		cx := xs.toPixel(float64(i) + 0.5)
		top := ys.toPixel(v)
		p.rect(cx-barW/2, top, barW, baseline-top, col)

		if opts.showValues {
			lbl := formatTick(v)
			y := top - 14
			if v < 0 {
				y = baseline + 2
			}
			p.text(cx-40, y, 80, 12, lbl, colText, 10, fyne.TextAlignCenter)
		}
	}
}

func (b *BarChart) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	if len(b.values) == 0 {
		return hitResult{}, false
	}
	// Category columns are the hit targets: the full column height is active.
	i := int(float64(xs.fromPixel(px)) / 1.0)
	if i < 0 {
		i = 0
	}
	if i >= len(b.values) {
		i = len(b.values) - 1
	}
	label := ""
	if i < len(b.labels) {
		label = b.labels[i]
	}
	return hitResult{Index: i, Label: label, Value: b.values[i], ValueStr: formatTick(b.values[i])}, true
}

// --- helpers ----------------------------------------------------------------

func darken(c color.Color, f float64) color.Color {
	r, g, b, a := c.RGBA()
	return color.NRGBA{
		R: uint8(float64(r>>8) * f),
		G: uint8(float64(g>>8) * f),
		B: uint8(float64(b>>8) * f),
		A: uint8(a >> 8),
	}
}

// updateData replaces labels/values and refreshes domains and hover state.
func (b *BarChart) updateData(labels []string, values []float64) {
	b.labels = labels
	b.values = values
	if len(b.colors) > len(values) {
		b.colors = b.colors[:len(values)]
	}
	b.clearHover()
	b.recomputeDomains()
	b.refresh()
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *BarChart) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

func (g *BarChart) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "SetValues":
		return object.NewBuiltin("chart.barchart.SetValues", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			values, err := floatSliceArg(args[0])
			if err != nil {
				return nil, err
			}
			if len(values) != len(g.labels) {
				return nil, fmt.Errorf("chart error: values length %d does not match labels length %d", len(values), len(g.labels))
			}
			guithread.Do(func() { g.updateData(g.labels, values) })
			return object.Nil, nil
		}), true
	case "SetData":
		return object.NewBuiltin("chart.barchart.SetData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
			labels, err := stringSliceArg(args[0])
			if err != nil {
				return nil, err
			}
			values, err := floatSliceArg(args[1])
			if err != nil {
				return nil, err
			}
			if len(labels) != len(values) {
				return nil, fmt.Errorf("chart error: labels and values must have same length (got %d labels, %d values)", len(labels), len(values))
			}
			guithread.Do(func() { g.updateData(labels, values) })
			return object.Nil, nil
		}), true
	case "SetValue":
		return object.NewBuiltin("chart.barchart.SetValue", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
			idx, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			v, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= int64(len(g.values)) {
				return nil, fmt.Errorf("chart error: index %d out of range [0, %d)", idx, len(g.values))
			}
			guithread.Do(func() {
				g.values[idx] = v
				g.clearHover()
				g.recomputeDomains()
				g.refresh()
			})
			return object.Nil, nil
		}), true
	case "SetLabel":
		return object.NewBuiltin("chart.barchart.SetLabel", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
			idx, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			s, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= int64(len(g.labels)) {
				return nil, fmt.Errorf("chart error: index %d out of range [0, %d)", idx, len(g.labels))
			}
			guithread.Do(func() {
				g.labels[idx] = s
				g.clearHover()
				g.refresh()
			})
			return object.Nil, nil
		}), true
	case "SetColor":
		return object.NewBuiltin("chart.barchart.SetColor", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			col, err := colorArg(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				g.baseColor = col
				g.refresh()
			})
			return object.Nil, nil
		}), true
	case "SetBarColor":
		return object.NewBuiltin("chart.barchart.SetBarColor", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
			idx, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			col, err := colorArg(args[1])
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= int64(len(g.values)) {
				return nil, fmt.Errorf("chart error: index %d out of range [0, %d)", idx, len(g.values))
			}
			guithread.Do(func() {
				for len(g.colors) <= int(idx) {
					g.colors = append(g.colors, nil)
				}
				g.colors[idx] = col
				g.refresh()
			})
			return object.Nil, nil
		}), true
	}
	return g.chartCore.getAttr(name)
}

// --- shared argument helpers --------------------------------------------------

func colorArg(o object.Object) (color.Color, error) {
	if s, ok := o.(*object.String); ok {
		return risorcanvas.ParseColor(s.Value())
	}
	return nil, fmt.Errorf("type error: expected a color name or hex string, got %s", o.Type())
}

func floatSliceArg(o object.Object) ([]float64, error) {
	list, ok := o.(*object.List)
	if !ok {
		return nil, fmt.Errorf("type error: expected a list (%s given)", o.Type())
	}
	result := make([]float64, 0, len(list.Value()))
	for _, item := range list.Value() {
		f, err := object.AsFloat(item)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func stringSliceArg(o object.Object) ([]string, error) {
	list, ok := o.(*object.List)
	if !ok {
		return nil, fmt.Errorf("type error: expected a list (%s given)", o.Type())
	}
	result := make([]string, 0, len(list.Value()))
	for _, item := range list.Value() {
		s, err := object.AsString(item)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}
