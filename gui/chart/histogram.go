package risorchart

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

var _ object.Object = &Histogram{}
var _ chartDelegate = &Histogram{}
var _ fyne.CanvasObject = &Histogram{}

const HistogramType object.Type = "chart.histogram"

var colNormal = color.NRGBA{R: 220, G: 20, B: 60, A: 255} // crimson

// Histogram is an interactive histogram built from Fyne canvas primitives,
// optionally overlaid with a fitted normal distribution curve. Hovering a bin
// shows its range and count; wheel zoom and drag pan inspect regions of the
// distribution (double-click resets).
type Histogram struct {
	*chartCore

	values []float64
	bins   int

	// computed binning
	edgeMin, edgeMax, binWidth float64
	counts                     []int

	showNormal bool
}

// NewHistogram creates an interactive histogram with the given title, axis
// labels, values, and number of bins. Returns an error if bins <= 0.
func NewHistogram(title, xlabel, ylabel string, values []float64, bins int) (*Histogram, error) {
	if bins <= 0 {
		return nil, fmt.Errorf("chart error: bins must be greater than 0 (got %d)", bins)
	}

	h := &Histogram{
		values:     values,
		bins:       bins,
		showNormal: true,
	}
	h.computeBins()
	h.chartCore = newChartCore(string(HistogramType), h)
	h.title = title
	h.xLabel = xlabel
	h.yLabel = ylabel
	h.ExtendBaseWidget(h)
	return h, nil
}

// --- object.Object ----------------------------------------------------------

func (obj *Histogram) CanvasObject() fyne.CanvasObject {
	return obj
}

func (r *Histogram) Type() object.Type {
	return HistogramType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (r *Histogram) Inspect() string {
	return guithread.Inspect(r.inspectUI)
}

func (r *Histogram) inspectUI() string {
	return "chart.histogram"
}

func (r *Histogram) Interface() interface{} {
	return r
}

func (r *Histogram) IsTruthy() bool {
	return true
}

func (r *Histogram) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", HistogramType, opType)
}

func (r *Histogram) Equals(other object.Object) bool {
	return r == other
}

func (*Histogram) Attrs() []object.AttrSpec {
	return append(histogramAttrs.Specs(), saveImageSpec)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (r *Histogram) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, r.setAttrUI)
}

func (r *Histogram) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", HistogramType, name)
}

// --- binning -----------------------------------------------------------------

func (h *Histogram) computeBins() {
	h.counts = make([]int, h.bins)
	if len(h.values) == 0 {
		h.edgeMin, h.edgeMax, h.binWidth = 0, 1, 1.0/float64(h.bins)
		return
	}
	minV, maxV := h.values[0], h.values[0]
	for _, v := range h.values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if maxV == minV {
		minV, maxV = minV-0.5, maxV+0.5
	}
	h.edgeMin, h.edgeMax = minV, maxV
	h.binWidth = (maxV - minV) / float64(h.bins)
	for _, v := range h.values {
		i := int((v - minV) / h.binWidth)
		if i >= h.bins {
			i = h.bins - 1
		}
		if i < 0 {
			i = 0
		}
		h.counts[i]++
	}
}

// --- chartDelegate ----------------------------------------------------------

func (h *Histogram) dataExtent() (float64, float64, float64, float64) {
	maxCount := 0
	for _, c := range h.counts {
		if c > maxCount {
			maxCount = c
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}
	return h.edgeMin, h.edgeMax, 0, float64(maxCount) * 1.15
}

func (h *Histogram) xTicks() ([]float64, []string) {
	return autoTicks(h.xScale)
}

func (h *Histogram) buildSeries(p *plotPainter, xs, ys *linearScale, opts *seriesOpts) {
	base := color.NRGBA{R: 70, G: 130, B: 180, A: 180} // translucent steel blue
	baseline := ys.toPixel(0)
	for i, c := range h.counts {
		x0 := xs.toPixel(h.edgeMin + float64(i)*h.binWidth)
		x1 := xs.toPixel(h.edgeMin + float64(i+1)*h.binWidth)
		top := ys.toPixel(float64(c))
		var col color.Color = base
		if i == opts.hoverIdx {
			col = darken(base, 0.7)
		}
		p.rect(x0, top, x1-x0, baseline-top, col)

		if opts.showValues && c > 0 {
			p.text((x0+x1)/2-20, top-14, 40, 12, fmt.Sprintf("%d", c), colText, 10, fyne.TextAlignCenter)
		}
	}

	if h.showNormal {
		mean, stddev := calculateStats(h.values)
		if stddev > 0 && len(h.values) > 0 {
			totalArea := float64(len(h.values)) * h.binWidth
			const samples = 60
			pxs := make([]float32, samples)
			pys := make([]float32, samples)
			for i := 0; i < samples; i++ {
				x := h.edgeMin + (h.edgeMax-h.edgeMin)*float64(i)/float64(samples-1)
				pxs[i] = xs.toPixel(x)
				pys[i] = ys.toPixel(normalPDF(x, mean, stddev) * totalArea)
			}
			p.polyline(pxs, pys, colNormal, 2.5)

			// Legend for the two overlays, top-left of the plot area.
			lx, ly := p.left+10, p.top+8
			p.rect(lx, ly, 12, 12, base)
			p.text(lx+16, ly, 110, 12, "Data", colText, 10, fyne.TextAlignLeading)
			p.rect(lx, ly+18, 12, 12, colNormal)
			p.text(lx+16, ly+18, 130, 12, "Normal Distribution", colText, 10, fyne.TextAlignLeading)
		}
	}
}

func (h *Histogram) hitTest(px, py float32, xs, ys *linearScale) (hitResult, bool) {
	if len(h.counts) == 0 {
		return hitResult{}, false
	}
	x := xs.fromPixel(px)
	i := int((x - h.edgeMin) / h.binWidth)
	if i < 0 || i >= len(h.counts) {
		return hitResult{}, false
	}
	a := h.edgeMin + float64(i)*h.binWidth
	b := a + h.binWidth
	return hitResult{
		Index:    i,
		Label:    formatTick(a) + " - " + formatTick(b),
		Value:    float64(h.counts[i]),
		ValueStr: fmt.Sprintf("%d", h.counts[i]),
	}, true
}

// --- Risor method surface ----------------------------------------------------

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *Histogram) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

var histogramAttrs = object.NewAttrRegistry[*Histogram](string(HistogramType))

func (g *Histogram) getAttrUI(name string) (object.Object, bool) {
	if o, ok := histogramAttrs.GetAttr(g, name); ok {
		return o, true
	}
	return g.chartCore.getAttr(name)
}
func init() {
	objattrs.Add(histogramAttrs)
	addInteractive(histogramAttrs)
	histogramAttrs.Define("SetValues").Args(
		"arg1").
		Returns("nil").
		Impl(func(g *Histogram, ctx context.Context, args ...object.Object) (object.Object, error) {
			values, err := floatSliceArg(args[0])
			if err != nil {
				return nil,

					err
			}
			guithread.Do(func() {
				g.values = values
				g.computeBins()
				g.
					clearHover()
				g.recomputeDomains()
				g.refresh()
			})
			return object.Nil, nil
		})
	histogramAttrs.Define("SetBins").Args("arg1").Returns("nil").Impl(func(g *Histogram, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		n, err := object.AsInt(args[0])
		if err != nil {
			return nil, err
		}

		if n < 1 || n > 200 {
			return nil,
				fmt.Errorf("chart error: bins must be in [1, 200] (got %d)",
					n)
		}
		guithread.
			Do(func() {
				g.bins = int(n)
				g.computeBins()
				g.
					clearHover()
				g.recomputeDomains()
				g.refresh()
			})
		return object.Nil,

			nil
	})
	histogramAttrs.Define("ShowNormal").Args("arg1").Returns("nil").
		Impl(func(g *Histogram, ctx context.Context, args ...object.Object) (object.Object, error) {
			on := args[0].IsTruthy()
			guithread.Do(func() {
				g.showNormal =
					on
				g.refresh()
			})
			return object.Nil,
				nil
		})
}
