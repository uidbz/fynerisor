package risorchart

import (
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

// newTestBarChart builds a bar chart mounted in a test window so it has a
// realistic size, then lays out its renderer (which builds the scene).
func newTestBarChart(t *testing.T, labels []string, values []float64) (*BarChart, func()) {
	t.Helper()
	app := test.NewApp()
	bc, err := NewBarChart("Sales", "Revenue", labels, values)
	if err != nil {
		t.Fatalf("NewBarChart: %v", err)
	}
	win := app.NewWindow("chart")
	win.SetContent(bc)
	win.Resize(fyne.NewSize(500, 360))
	win.Show()

	r := bc.CreateRenderer()
	r.Layout(bc.Size())

	return bc, func() { win.Close(); app.Quit() }
}

func TestNewBarChartMismatch(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	if _, err := NewBarChart("t", "y", []string{"a", "b"}, []float64{1}); err == nil {
		t.Fatal("expected error for mismatched labels/values lengths")
	}
}

func TestBarChartSceneBuilt(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"Q1", "Q2", "Q3"}, []float64{10, 20, 30})
	defer done()

	if bc.Size().Width < 1 {
		t.Fatalf("chart not sized: %v", bc.Size())
	}
	var rects, texts int
	for _, it := range bc.sc.items {
		switch it.kind {
		case kindRect:
			rects++
		case kindText:
			texts++
		}
	}
	// background + 3 bars (at least), plus title/ylabel/axis labels + ticks.
	if rects < 4 {
		t.Errorf("expected >= 4 rects in scene, got %d", rects)
	}
	if texts < 4 {
		t.Errorf("expected >= 4 texts in scene, got %d", texts)
	}
	if len(bc.view.objs) != len(bc.sc.items) {
		t.Errorf("view objects %d != scene items %d", len(bc.view.objs), len(bc.sc.items))
	}
}

func TestBarChartHover(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"A", "B", "C"}, []float64{10, 20, 30})
	defer done()

	// Hover the middle of the middle category column.
	midX := (bc.xScale.toPixel(1.5))
	y := (bc.yScale.lo + bc.yScale.hi) / 2

	bc.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(midX, y)}})
	if bc.hoverIdx != 1 {
		t.Fatalf("hover index = %d, want 1", bc.hoverIdx)
	}
	if bc.tipText != "B: 20" {
		t.Errorf("tooltip = %q, want %q", bc.tipText, "B: 20")
	}

	// Leaving clears hover state.
	bc.MouseOut()
	if bc.hoverIdx != -1 || bc.tipText != "" {
		t.Errorf("MouseOut did not clear hover: idx=%d tip=%q", bc.hoverIdx, bc.tipText)
	}
}

func TestBarChartTapCallbacks(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"A", "B", "C"}, []float64{10, 20, 30})
	defer done()

	var tapped hitResult
	bc.onTapFn = func(h hitResult) { tapped = h }
	var hovered, secondary int
	bc.onHoverFn = func(h hitResult) { hovered = h.Index }
	bc.onSecondaryFn = func(h hitResult) { secondary++ }

	x := bc.xScale.toPixel(2.5)
	y := (bc.yScale.lo + bc.yScale.hi) / 2
	pos := fyne.PointEvent{Position: fyne.NewPos(x, y)}

	bc.MouseMoved(&desktop.MouseEvent{PointEvent: pos})
	if hovered != 2 {
		t.Errorf("hover callback index = %d, want 2", hovered)
	}
	bc.Tapped(&pos)
	if tapped.Index != 2 || tapped.Label != "C" || tapped.Value != 30 {
		t.Errorf("tap hit = %+v, want index 2/C/30", tapped)
	}
	bc.TappedSecondary(&pos)
	bc.TappedSecondary(&pos)
	if secondary != 2 {
		t.Errorf("secondary taps = %d, want 2", secondary)
	}
}

func TestBarChartZoomPanReset(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"A", "B", "C"}, []float64{10, 20, 30})
	defer done()

	baseW := bc.xScale.max - bc.xScale.min
	cx := (bc.xScale.lo + bc.xScale.hi) / 2
	cy := (bc.yScale.lo + bc.yScale.hi) / 2

	// Wheel up zooms in.
	bc.Scrolled(&fyne.ScrollEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(cx, cy)}, Scrolled: fyne.NewDelta(0, 1)})
	if w := bc.xScale.max - bc.xScale.min; w >= baseW {
		t.Errorf("after wheel-up width %v, want < %v", w, baseW)
	}

	// Drag pans while zoomed, staying clamped inside the base domain.
	beforeMin := bc.xScale.min
	bc.Dragged(&fyne.DragEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(cx, cy)}, Dragged: fyne.NewDelta(60, 0)})
	if bc.xScale.min >= beforeMin {
		t.Errorf("drag right should decrease domain min: %v -> %v", beforeMin, bc.xScale.min)
	}
	if bc.xScale.min < bc.xScale.baseMin || bc.xScale.max > bc.xScale.baseMax {
		t.Errorf("pan escaped base domain: %v..%v", bc.xScale.min, bc.xScale.max)
	}

	// Double-tap resets.
	bc.DoubleTapped(&fyne.PointEvent{Position: fyne.NewPos(cx, cy)})
	if bc.xScale.min != 0 || bc.xScale.max != baseW {
		t.Errorf("reset failed: %v..%v", bc.xScale.min, bc.xScale.max)
	}

	// Zoom can be disabled.
	bc.zoomEnabled = false
	w := bc.xScale.max - bc.xScale.min
	bc.Scrolled(&fyne.ScrollEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(cx, cy)}, Scrolled: fyne.NewDelta(0, 1)})
	if bc.xScale.max-bc.xScale.min != w {
		t.Error("Scrolled should be ignored when zoom is disabled")
	}
}

func TestBarChartDataUpdates(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"A", "B"}, []float64{1, 2})
	defer done()

	bc.updateData([]string{"A", "B", "C"}, []float64{1, 2, 100})
	if _, _, _, y1 := bc.dataExtent(); y1 < 100 {
		t.Errorf("domain did not grow with new value: y1=%v", y1)
	}
	if bc.hoverIdx != -1 {
		t.Error("data update should clear hover state")
	}

	// Hover still resolves to a valid index with more bars.
	x := bc.xScale.toPixel(2.5)
	y := (bc.yScale.lo + bc.yScale.hi) / 2
	bc.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(x, y)}})
	if bc.hoverIdx != 2 {
		t.Errorf("hover index = %d, want 2", bc.hoverIdx)
	}

	bc.values[0] = -50
	bc.recomputeDomains()
	x0, _, y0, _ := bc.dataExtent()
	if x0 != 0 || y0 > -50 {
		t.Errorf("negative baseline not in domain: %v %v", x0, y0)
	}
}

func TestBarChartBarColors(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"A", "B"}, []float64{5, 5})
	defer done()

	bc.colors = append(bc.colors, nil, nil)
	bc.colors[1] = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	r := bc.CreateRenderer()
	r.Layout(bc.Size())

	red := 0
	for _, it := range bc.sc.items {
		if it.kind == kindRect && it.col != nil {
			c, _ := it.col.(color.NRGBA)
			if c.R == 255 && c.G == 0 && c.B == 0 {
				red++
			}
		}
	}
	if red != 1 {
		t.Errorf("expected one red bar rect, got %d", red)
	}
}

func TestBarChartSVGExport(t *testing.T) {
	bc, done := newTestBarChart(t, []string{"Q1", "Q2"}, []float64{10, 20})
	defer done()

	svg := bc.writeSVG(400, 300)
	if !strings.HasPrefix(svg, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Errorf("missing SVG header: %.60s", svg)
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("SVG not terminated")
	}
	for _, want := range []string{`>Sales<`, `>Q1<`, `>Q2<`, `<rect`, `<text`} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG missing %q", want)
		}
	}
	// Export must not disturb the live scene.
	bc.sc = scene{}
	r := bc.CreateRenderer()
	r.Layout(bc.Size())
	if len(bc.sc.items) == 0 {
		t.Error("live scene empty after export")
	}

	// Values in SVG are finite (no NaN from zero-size charts).
	if strings.Contains(svg, "NaN") {
		t.Error("SVG contains NaN coordinates")
	}
}

func TestBarChartExportEmptyData(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	bc, err := NewBarChart("Empty", "Y", []string{}, []float64{})
	if err != nil {
		t.Fatalf("NewBarChart: %v", err)
	}
	svg := bc.writeSVG(300, 200)
	if strings.Contains(svg, "NaN") {
		t.Error("empty chart export contains NaN")
	}
}
