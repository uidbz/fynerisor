package risorchart

import (
	"context"
	"fmt"
	"image/color"
	"sort"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

var _ object.Object = &BoxPlot{}
var _ chartDelegate = &BoxPlot{}
var _ fyne.CanvasObject = &BoxPlot{}

const BoxPlotType object.Type = "chart.boxplot"

// boxColors cycles across datasets (same palette as the legacy renderer).
var boxColors = []color.NRGBA{
	{R: 70, G: 130, B: 180, A: 255}, // steel blue
	{R: 220, G: 20, B: 60, A: 255},  // crimson
	{R: 50, G: 205, B: 50, A: 255},  // lime green
	{R: 255, G: 140, B: 0, A: 255},  // dark orange
	{R: 138, G: 43, B: 226, A: 255}, // blue violet
}

// boxStats holds the pre-computed five-number summary per dataset.
type boxStats struct {
	n           int
	lo, hi      float64 // whisker ends (within 1.5*IQR)
	q1, med, q3 float64
	mean        float64
	outliers    []float64
	valid       bool
}

// BoxPlot is an interactive box plot built from Fyne canvas primitives. Each
// category shows the Tukey box (IQR), median, whiskers (1.5*IQR), mean marker
// and outlier points. Hovering a category summarizes the dataset; wheel zoom
// and drag pan inspect value ranges (double-click resets).
type BoxPlot struct {
	*chartCore

	labels   []string
	datasets [][]float64
	stats    []boxStats
}

// NewBoxPlot creates an interactive box plot with the given title, y-axis
// label, category labels and datasets. Returns an error if labels and
// datasets lengths do not match.
func NewBoxPlot(title, ylabel string, labels []string, datasets [][]float64) (*BoxPlot, error) {
	if len(labels) != len(datasets) {
		return nil, fmt.Errorf("chart error: labels and datasets must have same length (got %d labels, %d datasets)", len(labels), len(datasets))
	}

	b := &BoxPlot{labels: labels, datasets: datasets}
	b.computeStats()
	b.chartCore = newChartCore(string(BoxPlotType), b)
	b.title = title
	b.yLabel = ylabel
	b.ExtendBaseWidget(b)
	return b, nil
}

// --- object.Object ----------------------------------------------------------

func (obj *BoxPlot) CanvasObject() fyne.CanvasObject {
	return obj
}

func (r *BoxPlot) Type() object.Type {
	return BoxPlotType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (r *BoxPlot) Inspect() string {
	return guithread.Inspect(r.inspectUI)
}

func (r *BoxPlot) inspectUI() string {
	return "chart.boxplot"
}

func (r *BoxPlot) Interface() interface{} {
	return r
}

func (r *BoxPlot) IsTruthy() bool {
	return true
}

func (r *BoxPlot) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", BoxPlotType, opType)
}

func (r *BoxPlot) Equals(other object.Object) bool {
	return r == other
}

func (*BoxPlot) Attrs() []object.AttrSpec {
	return append(boxPlotAttrs.Specs(), saveImageSpec)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (r *BoxPlot) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, r.setAttrUI)
}

func (r *BoxPlot) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BoxPlotType, name)
}

// --- statistics ---------------------------------------------------------------

func (b *BoxPlot) computeStats() {
	b.stats = make([]boxStats, len(b.datasets))
	for i, data := range b.datasets {
		st := &b.stats[i]
		if len(data) == 0 {
			continue
		}
		sorted := append([]float64(nil), data...)
		sort.Float64s(sorted)
		st.n = len(sorted)
		st.valid = true
		st.lo = sorted[0]
		st.hi = sorted[len(sorted)-1]
		st.q1 = quantile(sorted, 0.25)
		st.med = quantile(sorted, 0.5)
		st.q3 = quantile(sorted, 0.75)
		st.mean = meanOf(sorted)

		// Tukey fences for whiskers and outliers.
		iqr := st.q3 - st.q1
		fenceLo, fenceHi := st.q1-1.5*iqr, st.q3+1.5*iqr
		loW, hiW := st.hi, st.lo
		for _, v := range sorted {
			if v < fenceLo || v > fenceHi {
				st.outliers = append(st.outliers, v)
				continue
			}
			if v < loW {
				loW = v
			}
			if v > hiW {
				hiW = v
			}
		}
		st.lo, st.hi = loW, hiW
	}
}

func meanOf(sorted []float64) float64 {
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	return sum / float64(len(sorted))
}

// --- chartDelegate ------------------------------------------------------------

func (b *BoxPlot) dataExtent() (float64, float64, float64, float64) {
	n := float64(len(b.datasets))
	if n == 0 {
		return 0, 1, 0, 1
	}
	first := true
	var ymin, ymax float64
	for _, data := range b.datasets {
		for _, v := range data {
			if first || v < ymin {
				ymin = v
			}
			if first || v > ymax {
				ymax = v
			}
			first = false
		}
	}
	if first {
		return 0, n, 0, 1
	}
	pad := (ymax - ymin) * 0.08
	return 0, n, ymin - pad, ymax + pad
}

func (b *BoxPlot) xTicks() ([]float64, []string) {
	pos := make([]float64, len(b.labels))
	for i := range b.labels {
		pos[i] = float64(i) + 0.5
	}
	return pos, b.labels
}

func (b *BoxPlot) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	unit := xs.toPixel(1) - xs.toPixel(0)
	boxW := unit * 0.5
	capW := unit * 0.22

	for i := range b.datasets {
		st := &b.stats[i]
		if !st.valid {
			continue
		}
		col := boxColors[i%len(boxColors)]
		hover := i == opts.hoverIdx
		cx := xs.toPixel(float64(i) + 0.5)

		// IQR box.
		fill := color.NRGBA{R: col.R, G: col.G, B: col.B, A: 110}
		if hover {
			fill = color.NRGBA{R: col.R / 2, G: col.G / 2, B: col.B / 2, A: 160}
		}
		boxTop := ys.toPixel(st.q3)
		boxBot := ys.toPixel(st.q1)
		p.rect(cx-boxW/2, boxTop, boxW, boxBot-boxTop, fill)

		// Box outline + median.
		p.line(cx-boxW/2, boxTop, cx+boxW/2, boxTop, col, 1)
		p.line(cx-boxW/2, boxBot, cx+boxW/2, boxBot, col, 1)
		p.line(cx-boxW/2, boxTop, cx-boxW/2, boxBot, col, 1)
		p.line(cx+boxW/2, boxTop, cx+boxW/2, boxBot, col, 1)
		medY := ys.toPixel(st.med)
		p.line(cx-boxW/2, medY, cx+boxW/2, medY, darken(col, 0.6), 2)

		// Whiskers with caps.
		loY, hiY := ys.toPixel(st.lo), ys.toPixel(st.hi)
		p.line(cx, loY, cx, boxBot, col, 1)
		p.line(cx, boxTop, cx, hiY, col, 1)
		p.line(cx-capW/2, loY, cx+capW/2, loY, col, 1)
		p.line(cx-capW/2, hiY, cx+capW/2, hiY, col, 1)

		// Mean marker.
		mY := ys.toPixel(st.mean)
		p.rect(cx-3, mY-3, 6, 6, darken(col, 0.5))

		// Outliers.
		for _, o := range st.outliers {
			oY := ys.toPixel(o)
			p.rect(cx-2.5, oY-2.5, 5, 5, col)
		}

		if opts.showValues {
			p.text(cx-40, hiY-16, 80, 12, "med "+formatTick(st.med), colText, 10, fyne.TextAlignCenter)
		}
	}
}

func (b *BoxPlot) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	if len(b.datasets) == 0 {
		return hitResult{}, false
	}
	i := int(xs.fromPixel(px))
	if i < 0 {
		i = 0
	}
	if i >= len(b.datasets) {
		i = len(b.datasets) - 1
	}
	label := ""
	if i < len(b.labels) {
		label = b.labels[i]
	}
	st := &b.stats[i]
	summary := "n/a"
	med := 0.0
	if st.valid {
		med = st.med
		summary = fmt.Sprintf("median %s (n=%d)", formatTick(st.med), st.n)
	}
	return hitResult{Index: i, Label: label, Value: med, ValueStr: summary}, true
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *BoxPlot) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

var boxPlotAttrs = object.NewAttrRegistry[*BoxPlot](string(BoxPlotType))

func (g *BoxPlot) getAttrUI(name string) (object.Object, bool) {
	if o, ok := boxPlotAttrs.GetAttr(g, name); ok {
		return o, true
	}
	return g.chartCore.getAttr(name)
}
func init() {
	objattrs.Add(boxPlotAttrs)
	addInteractive(boxPlotAttrs)
	boxPlotAttrs.Define(
		"SetData",
	).Args("arg1",
		"arg2",
	).Returns("nil").Impl(func(g *BoxPlot, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		labels, err := stringSliceArg(args[0])
		if err != nil {
			return nil,
				err

		}
		list, ok := args[1].(*object.
			List)
		if !ok {
			return nil, fmt.
				Errorf("type error: expected a list of lists for datasets, got %s",

					args[1].Type())
		}

		datasets := make([][]float64, 0, len(list.Value()))
		for _, item := range list.
			Value() {
			ds, err := floatSliceArg(item)
			if err != nil {
				return nil, err
			}
			datasets = append(datasets, ds)
		}
		if len(labels) != len(datasets) {
			return nil, fmt.Errorf("chart error: labels and datasets must have same length (got %d labels, %d datasets)",
				len(labels), len(datasets))
		}
		guithread.Do(func() {
			g.labels, g.datasets =
				labels, datasets
			g.computeStats()
			g.clearHover()
			g.recomputeDomains()
			g.refresh()
		})
		return object.Nil, nil
	},
	)
}
