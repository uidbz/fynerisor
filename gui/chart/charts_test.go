package risorchart

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

// mountChart puts a chart in a test window, lays out its renderer and returns
// a teardown func.
func mountChart(t *testing.T, obj interface{ fyne.CanvasObject }, w, h float32) func() {
	t.Helper()
	app := test.NewApp()
	win := app.NewWindow("chart")
	win.SetContent(obj)
	win.Resize(fyne.NewSize(w, h))
	win.Show()
	if r := obj.(interface{ CreateRenderer() fyne.WidgetRenderer }).CreateRenderer(); r != nil {
		r.Layout(obj.(fyne.CanvasObject).Size())
	}
	return func() { win.Close(); app.Quit() }
}

func TestScatterBasics(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5}
	ys := []float64{2, 4, 1, 8, 7}
	sc, err := NewScatterChart("T", "X", "Y", xs, ys)
	if err != nil {
		t.Fatalf("NewScatterChart: %v", err)
	}
	defer mountChart(t, sc, 460, 340)()

	if _, err := NewScatterChart("T", "X", "Y", []float64{1}, []float64{}); err == nil {
		t.Error("expected mismatch error")
	}

	// Hover near point 3 (x=4, y=8).
	px, py := sc.xScale.toPixel(4), sc.yScale.toPixel(8)
	h, ok := sc.hitTest(px+2, py-3, sc.xScale, sc.yScale)
	if !ok || h.Index != 3 {
		t.Errorf("hit near (4,8) = %+v ok=%v, want index 3", h, ok)
	}
	if h.ValueStr != "(4, 8)" {
		t.Errorf("value string = %q", h.ValueStr)
	}
	// Far away from any point yields no hit.
	if _, ok := sc.hitTest(sc.xScale.lo+1, sc.yScale.hi+1, sc.xScale, sc.yScale); ok {
		t.Error("expected miss far from points")
	}

	svg := sc.writeSVG(400, 300)
	if !strings.Contains(svg, "<text") || strings.Contains(svg, "NaN") {
		t.Error("scatter SVG malformed")
	}

	// Wheel zoom narrows both domains (2D zoom).
	xw, yw := sc.xScale.max-sc.xScale.min, sc.yScale.max-sc.yScale.min
	sc.Scrolled(&fyne.ScrollEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(200, 150)}, Scrolled: fyne.NewDelta(0, 1)})
	if sc.xScale.max-sc.xScale.min >= xw || sc.yScale.max-sc.yScale.min >= yw {
		t.Error("2D zoom should shrink both domains")
	}
}

func TestScatterSetPoints(t *testing.T) {
	sc, err := NewScatterChart("T", "X", "Y", []float64{1, 2}, []float64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, sc, 460, 340)()

	sc.xvals, sc.yvals = []float64{5, 6, 7}, []float64{25, 26, 27}
	sc.clearHover()
	sc.recomputeDomains()
	if _, _, _, y1 := sc.dataExtent(); y1 < 27 {
		t.Errorf("domain did not follow new points: y1=%v", y1)
	}
}

func TestLineBasics(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	ys := []float64{10, 20, 15, 30}
	li, err := NewLineChart("T", "X", "Y", xs, ys)
	if err != nil {
		t.Fatalf("NewLineChart: %v", err)
	}
	defer mountChart(t, li, 460, 340)()

	if _, err := NewLineChart("T", "X", "Y", []float64{1, 2}, []float64{1}); err == nil {
		t.Error("expected mismatch error")
	}

	// Hit snaps to nearest x sample.
	h, ok := li.hitTest(li.xScale.toPixel(2.2), li.yScale.toPixel(19), li.xScale, li.yScale)
	if !ok || h.Index != 1 {
		t.Errorf("hit = %+v ok=%v, want index 1", h, ok)
	}
	// Polyline segments are in the scene.
	var lines int
	for _, it := range li.sc.items {
		if it.kind == kindLine && it.lineW == 2 {
			lines++
		}
	}
	if lines < 3 {
		t.Errorf("expected >= 3 polyline segments, got %d", lines)
	}
	if strings.Contains(li.writeSVG(400, 300), "NaN") {
		t.Error("line SVG contains NaN")
	}
}

func TestHistogramBasics(t *testing.T) {
	values := []float64{1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5}
	hc, err := NewHistogram("T", "X", "Y", values, 5)
	if err != nil {
		t.Fatalf("NewHistogram: %v", err)
	}
	defer mountChart(t, hc, 460, 340)()

	if _, err := NewHistogram("T", "X", "Y", values, 0); err == nil {
		t.Error("expected bins <= 0 error")
	}

	total := 0
	for _, c := range hc.counts {
		total += c
	}
	if total != len(values) {
		t.Errorf("bins hold %d values, want %d", total, len(values))
	}

	// Hover over the leftmost bin.
	px := hc.xScale.toPixel(hc.edgeMin + hc.binWidth*0.5)
	py := (hc.yScale.lo + hc.yScale.hi) / 2
	hc.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(px, py)}})
	if hc.hoverIdx != 0 {
		t.Errorf("hover index = %d, want 0", hc.hoverIdx)
	}

	// Normal curve adds polyline segments and legend items.
	withCurve := len(hc.sc.items)
	hc.showNormal = false
	r := hc.CreateRenderer()
	r.Layout(hc.Size())
	withoutCurve := len(hc.sc.items)
	if withCurve <= withoutCurve {
		t.Error("normal curve/legend should add scene items")
	}
	if strings.Contains(hc.writeSVG(400, 300), "NaN") {
		t.Error("histogram SVG contains NaN")
	}
}

func TestHistogramEqualValues(t *testing.T) {
	hc, err := NewHistogram("T", "X", "Y", []float64{7, 7, 7}, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, hc, 460, 340)()
	if hc.counts[1] != 3 {
		t.Errorf("equal values should bin together: %+v", hc.counts)
	}
}

func TestBoxPlotBasics(t *testing.T) {
	labels := []string{"A", "B"}
	datasets := [][]float64{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		{2, 3, 3, 3, 4, 5, 5, 6, 30}, // 30 is an outlier
	}
	bp, err := NewBoxPlot("T", "Y", labels, datasets)
	if err != nil {
		t.Fatalf("NewBoxPlot: %v", err)
	}
	defer mountChart(t, bp, 460, 340)()

	if _, err := NewBoxPlot("T", "Y", labels, datasets[:1]); err == nil {
		t.Error("expected mismatch error")
	}

	st := bp.stats[1]
	if !st.valid || st.med != 4 {
		t.Errorf("median of dataset B = %v, want 4", st.med)
	}
	if len(st.outliers) != 1 || st.outliers[0] != 30 {
		t.Errorf("outliers = %v, want [30]", st.outliers)
	}
	if st.hi != 6 {
		t.Errorf("upper whisker = %v, want 6", st.hi)
	}

	// Hover middle of category B column.
	px := bp.xScale.toPixel(1.5)
	py := (bp.yScale.lo + bp.yScale.hi) / 2
	bp.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(px, py)}})
	if bp.hoverIdx != 1 {
		t.Errorf("hover index = %d, want 1", bp.hoverIdx)
	}
	if !strings.Contains(bp.tipText, "median 4") {
		t.Errorf("tooltip = %q", bp.tipText)
	}
	if strings.Contains(bp.writeSVG(400, 300), "NaN") {
		t.Error("boxplot SVG contains NaN")
	}
}

func TestQuantile(t *testing.T) {
	sorted := []float64{1, 2, 3, 4, 5}
	if got := quantile(sorted, 0.5); got != 3 {
		t.Errorf("median = %v, want 3", got)
	}
	if got := quantile(sorted, 0.25); got != 2 {
		t.Errorf("q1 = %v, want 2", got)
	}
	if got := quantile([]float64{4}, 0.9); got != 4 {
		t.Errorf("single-value quantile = %v, want 4", got)
	}
}
