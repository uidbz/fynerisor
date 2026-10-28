package gui

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// runScriptCapturingErrors loads a script into a test window, executes it,
// and fails the test if the runner logged "Script execution error" (the
// harness logs errors to stderr and sets the window status).
func runScriptCapturingErrors(t *testing.T, src string) {
	t.Helper()
	a := newTestApp(t)
	defer a.Quit()

	old := os.Stderr
	r, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = wr

	w := a.NewWindow("Script test")
	fw := NewWindow(w)
	fw.LoadScript(src)
	fw.Execute()
	time.Sleep(150 * time.Millisecond)

	wr.Close()
	os.Stderr = old
	out, _ := io.ReadAll(r)
	if strings.Contains(string(out), "Script execution error") {
		t.Errorf("script errors:\n%s", string(out))
	}
}

// TestExampleCharts executes examples/35-charts headlessly so the legacy
// chart scripts stay valid as the chart API evolves.
func TestExampleCharts(t *testing.T) {
	script, err := os.ReadFile("../examples/35-charts/main.risor")
	if err != nil {
		t.Fatal(err)
	}
	runScriptCapturingErrors(t, string(script))
}
