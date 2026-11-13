package risorchart

import (
	"context"
	"math"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/fynerisor/gui/canvas"
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

	sc.series[0].xvals, sc.series[0].yvals = []float64{5, 6, 7}, []float64{25, 26, 27}
	sc.clearHover()
	sc.recomputeDomains()
	if _, _, _, y1 := sc.dataExtent(); y1 < 27 {
		t.Errorf("domain did not follow new points: y1=%v", y1)
	}
}

func TestScatterMultiSeries(t *testing.T) {
	sc, err := NewScatterChart("T", "X", "Y", []float64{1, 2, 3}, []float64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	// Series 1: markers hidden, drawn as a 2px y=x line spanning 0..10.
	sc.series = append(sc.series, &scatterSeries{
		xvals: []float64{0, 10}, yvals: []float64{0, 10},
		col: paletteColor(1), markerSize: 0, lineWidth: 2,
	})
	sc.recomputeDomains()
	defer mountChart(t, sc, 460, 340)()

	if _, _, _, y1 := sc.dataExtent(); y1 < 10 {
		t.Errorf("extent must cover every series, y1=%v", y1)
	}
	if x0, _, _, _ := sc.dataExtent(); x0 > 0 {
		t.Errorf("extent must cover every series, x0=%v", x0)
	}

	// hitTest reaches series 1 and reports its index.
	px, py := sc.xScale.toPixel(10), sc.yScale.toPixel(10)
	h, ok := sc.hitTest(px, py, sc.xScale, sc.yScale)
	if !ok || h.Series != 1 || h.Index != 1 {
		t.Errorf("hit = %+v ok=%v, want series 1 index 1", h, ok)
	}
	if h.ValueStr != "(10, 10)" {
		t.Errorf("value string = %q", h.ValueStr)
	}

	// The line series contributes 2px polyline segments to the scene.
	var segs int
	for _, it := range sc.sc.items {
		if it.kind == kindLine && it.lineW == 2 {
			segs++
		}
	}
	if segs < 1 {
		t.Errorf("expected polyline segments for the line series, got %d", segs)
	}

	// Hover identity distinguishes series: hovering the line series point
	// must not collide with the same index in series 0.
	sc.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(px, py)}})
	if sc.hoverSeries != 1 || sc.hoverIdx != 1 {
		t.Errorf("hover = series %d index %d, want series 1 index 1", sc.hoverSeries, sc.hoverIdx)
	}
}

func TestLineMultiSeries(t *testing.T) {
	li, err := NewLineChart("T", "X", "Y", []float64{1, 2, 3}, []float64{10, 20, 30})
	if err != nil {
		t.Fatal(err)
	}
	li.series = append(li.series, &lineSeries{
		xvals: []float64{1, 2, 3}, yvals: []float64{1, 2, 100},
		col: paletteColor(2), width: 2,
	})
	li.recomputeDomains()
	defer mountChart(t, li, 460, 340)()

	if _, _, _, y1 := li.dataExtent(); y1 < 100 {
		t.Errorf("y domain %v must cover the second series", y1)
	}

	// x-snap is per series: near the second line's endpoint, not the first.
	px, py := li.xScale.toPixel(3), li.yScale.toPixel(100)
	h, ok := li.hitTest(px, py, li.xScale, li.yScale)
	if !ok || h.Series != 1 || h.Index != 2 {
		t.Errorf("hit = %+v ok=%v, want series 1 index 2", h, ok)
	}
	// And near the primary line, the primary wins.
	px, py = li.xScale.toPixel(3), li.yScale.toPixel(30)
	h, ok = li.hitTest(px, py, li.xScale, li.yScale)
	if !ok || h.Series != 0 || h.Index != 2 {
		t.Errorf("hit = %+v ok=%v, want series 0 index 2", h, ok)
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

// callChart invokes a chart's Risor method by name (builtins run inline on
// the test's main goroutine via guithread) and fails the test on error.
func callChart(t *testing.T, obj interface {
	GetAttr(string) (object.Object, bool)
}, name string, args ...object.Object) object.Object {
	t.Helper()
	a, ok := obj.GetAttr(name)
	if !ok {
		t.Fatalf("%s: attribute not bound", name)
	}
	b, ok := a.(*object.Builtin)
	if !ok {
		t.Fatalf("%s: not a builtin (%s)", name, a.Type())
	}
	res, err := b.Call(context.Background(), args...)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.Type() == object.ERROR {
		t.Fatalf("%s: %s", name, res.Inspect())
	}
	return res
}

func floatList(vals ...float64) *object.List {
	items := make([]object.Object, len(vals))
	for i, v := range vals {
		items[i] = object.NewFloat(v)
	}
	return object.NewList(items)
}

func stringList(vals ...string) *object.List {
	items := make([]object.Object, len(vals))
	for i, v := range vals {
		items[i] = object.NewString(v)
	}
	return object.NewList(items)
}

func countScene(sc *ScatterChart, kind sceneKind, pred func(sceneItem) bool) int {
	var n int
	for _, it := range sc.sc.items {
		if it.kind == kind && pred(it) {
			n++
		}
	}
	return n
}

func TestScatterAnnotations(t *testing.T) {
	sc, err := NewScatterChart("scores", "t1", "t2", []float64{1, 2, 3}, []float64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, sc, 460, 340)()

	callChart(t, sc, "SetSeriesMarkerSizes", object.NewInt(0), floatList(2, 10, 0))
	callChart(t, sc, "SetSeriesPointColors", object.NewInt(0), stringList("", "red", ""))
	callChart(t, sc, "SetSeriesPointLabels", object.NewInt(0), stringList("A", "", "C"))

	if n := countScene(sc, kindRect, func(it sceneItem) bool { return it.w == 10 }); n != 1 {
		t.Errorf("expected 1 marker of size 10, got %d", n)
	}
	if n := countScene(sc, kindRect, func(it sceneItem) bool { return it.w == 2 }); n != 1 {
		t.Errorf("expected 1 marker of size 2, got %d", n)
	}
	// Per-point color took effect on point 1 (the one parsed from "red").
	wantRed, _ := risorcanvas.ParseColor("red")
	if n := countScene(sc, kindRect, func(it sceneItem) bool { return it.w == 10 && it.col == wantRed }); n != 1 {
		t.Errorf("expected the size-10 marker to use the per-point red, got %d", n)
	}
	// Labels A and C drawn, B skipped; point 3 (size 0) has no marker but
	// still shows its label.
	for _, want := range []string{"A", "C"} {
		if n := countScene(sc, kindText, func(it sceneItem) bool { return it.text == want }); n != 1 {
			t.Errorf("expected label %q once, got %d", want, n)
		}
	}

	// Validation: length mismatch and bad colors are errors, not panics.
	a, _ := sc.GetAttr("SetSeriesMarkerSizes")
	res, err := a.(*object.Builtin).Call(context.Background(), object.NewInt(0), floatList(1, 2))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("expected error for mismatched sizes length")
	}
	a, _ = sc.GetAttr("SetSeriesPointColors")
	res, err = a.(*object.Builtin).Call(context.Background(), object.NewInt(0), stringList("notacolor", "", ""))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("expected error for invalid color name")
	}
	// Labels must survive SVG export too (same scene).
	svg := sc.writeSVG(400, 300)
	if !strings.Contains(svg, "A</text>") || !strings.Contains(svg, "C</text>") {
		t.Error("SVG export must include point labels")
	}
}

func TestScatterReferenceGeometry(t *testing.T) {
	sc, err := NewScatterChart("scores", "t1", "t2", []float64{-1, 1}, []float64{-1, 1})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, sc, 460, 340)()

	if got := callChart(t, sc, "AddHLine", object.NewFloat(2)).(*object.Int).Value(); got != 1 {
		t.Errorf("AddHLine series index = %d, want 1", got)
	}
	if got := callChart(t, sc, "AddEllipse", object.NewFloat(0), object.NewFloat(0),
		object.NewFloat(2), object.NewFloat(1)).(*object.Int).Value(); got != 2 {
		t.Errorf("AddEllipse series index = %d, want 2", got)
	}
	x0, x1, y0, y1 := sc.dataExtent()
	if x0 > -2 || x1 < 2 {
		t.Errorf("extent must cover the ellipse x: [%v, %v]", x0, x1)
	}
	if y0 > -1 || y1 < 2 {
		t.Errorf("extent must cover ellipse + hline y: [%v, %v]", y0, y1)
	}
	if n := countScene(sc, kindLine, func(it sceneItem) bool { return it.lineW == 1.5 }); n < 64 {
		t.Errorf("ellipse must render ~64 polyline segments, got %d", n)
	}
	a, _ := sc.GetAttr("AddEllipse")
	res, err := a.(*object.Builtin).Call(context.Background(),
		object.NewFloat(0), object.NewFloat(0), object.NewFloat(0), object.NewFloat(1))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("expected error for zero radius")
	}
}

func TestEllipsePoints(t *testing.T) {
	xs, ys := ellipsePoints(3, -2, 4, 1, 64)
	if len(xs) != 65 || len(ys) != 65 {
		t.Fatalf("expected 65 points, got %d/%d", len(xs), len(ys))
	}
	if xs[0] != xs[64] || ys[0] != ys[64] {
		t.Error("ellipse polyline must close")
	}
	minX, maxX, minY, maxY := xs[0], xs[0], ys[0], ys[0]
	for i := range xs {
		if xs[i] < minX {
			minX = xs[i]
		}
		if xs[i] > maxX {
			maxX = xs[i]
		}
		if ys[i] < minY {
			minY = ys[i]
		}
		if ys[i] > maxY {
			maxY = ys[i]
		}
	}
	if math.Abs(maxX-(3+4)) > 1e-9 || math.Abs(minX-(3-4)) > 1e-9 {
		t.Errorf("x span [%v, %v], want [-1, 7]", minX, maxX)
	}
	if math.Abs(maxY-(-1)) > 1e-9 || math.Abs(minY-(-3)) > 1e-9 {
		t.Errorf("y span [%v, %v], want [-3, -1]", minY, maxY)
	}
}

func TestLineAddHLine(t *testing.T) {
	lc, err := NewLineChart("monitor", "sample", "T2", []float64{1, 2, 3, 4}, []float64{5, 2, 4, 3})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, lc, 460, 340)()

	callChart(t, lc, "AddHLine", object.NewFloat(4.5))
	if len(lc.series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(lc.series))
	}
	lim := lc.series[1]
	if lim.xvals[0] != 1 || lim.xvals[1] != 4 {
		t.Errorf("limit line must span the data x-range, got %v", lim.xvals)
	}
	if lim.yvals[0] != 4.5 || lim.yvals[1] != 4.5 {
		t.Errorf("limit line y = %v, want constant 4.5", lim.yvals)
	}
	if lim.col != colorRed || lim.width != 1.5 {
		t.Errorf("default limit style must be red 1.5px, got col=%v width=%v", lim.col, lim.width)
	}
	callChart(t, lc, "AddHLine", object.NewFloat(2), object.NewString("green"), object.NewFloat(3))
	if lc.series[2].width != 3 {
		t.Errorf("explicit line width not applied: %v", lc.series[2].width)
	}
}

func TestBarErrorBars(t *testing.T) {
	bc, err := NewBarChart("coeffs", "beta", []string{"v1", "v2"}, []float64{2, 4})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, bc, 460, 340)()

	callChart(t, bc, "SetErrorBars", floatList(1, 0))

	var whiskers int
	for _, it := range bc.sc.items {
		if it.kind == kindLine && it.lineW == 1.2 {
			whiskers++
		}
	}
	if whiskers != 3 { // stem + 2 caps for bar 0 only
		t.Errorf("expected 3 whisker segments, got %d", whiskers)
	}
	if _, _, _, y1 := bc.dataExtent(); y1 < 3 {
		t.Errorf("extent must cover value+err (3), got %v", y1)
	}
	a, _ := bc.GetAttr("SetErrorBars")
	res, err := a.(*object.Builtin).Call(context.Background(), floatList(1))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("expected error for wrong error-bars length")
	}
	res, err = a.(*object.Builtin).Call(context.Background(), floatList(-1, 0))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("expected error for negative error bar")
	}
	// Clearing restores the plain extent.
	callChart(t, bc, "SetErrorBars", floatList())
	if _, _, _, y1 := bc.dataExtent(); y1 > 4+4*0.06+1e-9 {
		t.Errorf("cleared error bars must shrink extent back, got %v", y1)
	}
}

// TestChartErrorsPropagate guards the DoAndWaitContext shadowing bug:
// validation failures set inside the UI-thread closure must reach the
// caller, and a later successful call must not inherit the stale error.
func TestChartErrorsPropagate(t *testing.T) {
	sc, err := NewScatterChart("t", "x", "y", []float64{1}, []float64{1})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, sc, 300, 200)()

	a, _ := sc.GetAttr("AddSeries")
	res, err := a.(*object.Builtin).Call(context.Background(),
		floatList(1, 2), floatList(1, 2), object.NewString("bogus"))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("AddSeries must surface invalid color errors")
	}
	res, err = a.(*object.Builtin).Call(context.Background(), floatList(1, 2), floatList(3, 4))
	if err != nil {
		t.Fatalf("AddSeries after failed call: %v", err)
	}
	if res.Type() != object.INT || res.(*object.Int).Value() != 1 {
		t.Errorf("AddSeries index = %s, want int 1", res.Inspect())
	}

	lc, err := NewLineChart("t", "x", "y", []float64{1, 2}, []float64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	a, _ = lc.GetAttr("AddSeries")
	res, err = a.(*object.Builtin).Call(context.Background(),
		floatList(1, 2), floatList(1, 2), object.NewString("bogus"))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("line AddSeries must surface invalid color errors")
	}
}

func TestBarAddHLine(t *testing.T) {
	bc, err := NewBarChart("monitor", "T2", []string{"s0", "s1"}, []float64{2, 4})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, bc, 460, 340)()

	callChart(t, bc, "AddHLine", object.NewFloat(6))
	if _, _, _, y1 := bc.dataExtent(); y1 < 6 {
		t.Errorf("extent must cover the limit line, got %v", y1)
	}
	var n int
	for _, it := range bc.sc.items {
		if it.kind == kindLine && it.col == colorRed && it.lineW == 1.5 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("expected 1 red limit line, got %d", n)
	}
	a, _ := bc.GetAttr("AddHLine")
	res, err := a.(*object.Builtin).Call(context.Background(), object.NewFloat(1), object.NewString("bogus"))
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Error("AddHLine must surface invalid color errors")
	}
	callChart(t, bc, "RemoveHLines")
	if _, _, _, y1 := bc.dataExtent(); y1 > 4+4*0.06+1e-9 {
		t.Errorf("removed lines must shrink extent, got %v", y1)
	}
}
