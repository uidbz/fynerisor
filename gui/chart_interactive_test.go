package gui

import (
	"os"
	"strings"
	"testing"
	"time"

)

// TestInteractiveChartsScript verifies the new interactive chart method
// surface end-to-end through the Risor VM: registration of event callbacks,
// feature toggles, live data updates, zoom controls, and SVG export.
func TestInteractiveChartsScript(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	svgPath := os.TempDir() + "/fynerisor-chart-test.svg"
	_ = os.Remove(svgPath)
	defer os.Remove(svgPath)

	w := a.NewWindow("Chart Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		// --- bar chart: constructors unchanged, methods additive ---
		let bar = chart.NewBarChart("Sales", "Revenue", ["Q1", "Q2", "Q3"], [100, 150, 200])
		if (bar == nil) { throw("NewBarChart returned nil") }

		bar.OnTap((i, label, value) => print(sprintf("tap %s=%v", label, value)))
		bar.OnHover((i, label, value) => 1)
		bar.OnSecondaryTap((i, label, value) => 1)

		bar.ShowGrid(false)
		bar.ShowValues(true)
		bar.SetZoomEnabled(true)
		bar.SetColor("orange")
		bar.SetBarColor(1, "#ff0000")
		bar.SetValue(0, 120)
		bar.SetLabel(2, "Q3b")
		bar.SetData(["A", "B"], [1, 2])
		bar.ResetZoom()

		// error cases surface as Risor errors, not crashes
		let failed = false
		try { bar.SetValue(99, 1) } catch (err) { failed = true }
		if (!failed) { throw("SetValue out of range did not error") }
		failed = false
		try { chart.NewBarChart("t", "y", ["a", "b"], [1]) } catch (err) { failed = true }
		if (!failed) { throw("mismatched constructor args did not error") }
		failed = false
		try { bar.SaveImage("out.png") } catch (err) { failed = true }
		if (!failed) { throw("non-svg SaveImage did not error") }

		let svg = bar.WriteSVG(400, 300)
		if (!svg.contains("<svg")) { throw("WriteSVG returned no svg") }
		bar.SaveImage("` + svgPath + `")

		// --- the other chart types keep their constructors, gain methods ---
		let line = chart.NewLineChart("L", "x", "y", [1, 2, 3], [4, 5, 6])
		line.SetPoints([1, 2, 3, 4], [2, 4, 8, 7])
		line.SetColor("red")
		line.SetLineWidth(3.0)

		let sc = chart.NewScatterChart("S", "x", "y", [1, 2, 3], [3, 2, 1])
		sc.SetPoints([1, 2], [2, 1])
		sc.SetMarkerSize(8.0)
		sc.OnTap((i, label, value) => 1)

		let hist = chart.NewHistogram("H", "x", "y", [1.0, 2.0, 2.0, 3.0], 4)
		hist.SetBins(6)
		hist.ShowNormal(false)
		hist.SetValues([1.0, 1.5, 2.0])

		let box = chart.NewBoxPlot("B", "y", ["A", "B"], [[1, 2, 3], [4, 5, 6, 20]])
		box.SetData(["A"], [[1, 2]])

		// layout the charts like a script would
		let tabs = container.NewBorder(widget.NewLabel("charts"), nil, nil, nil,
			container.NewGridWithColumns(2, bar, line, sc, hist, box))
		window.SetContent(tabs)
	`)
	fw.Execute()
	time.Sleep(50 * time.Millisecond)

	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("SaveImage did not write file: %v", err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Errorf("saved file is not SVG: %.60s", string(data))
	}
}
