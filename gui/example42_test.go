package gui

import (
	"os"
	"testing"
)

// TestExampleInteractiveCharts executes examples/42-interactive-charts
// headlessly so the script stays valid as the chart API evolves.
func TestExampleInteractiveCharts(t *testing.T) {
	script, err := os.ReadFile("../examples/42-interactive-charts/main.risor")
	if err != nil {
		t.Fatal(err)
	}
	runScriptCapturingErrors(t, string(script))
}
