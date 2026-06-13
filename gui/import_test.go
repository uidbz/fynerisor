package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"
)

// findWidgets walks a CanvasObject tree collecting the first Button and Label.
func findWidgets(obj fyne.CanvasObject, btn **fynewidget.Button, lbl **fynewidget.Label) {
	switch o := obj.(type) {
	case *fynewidget.Button:
		if *btn == nil {
			*btn = o
		}
	case *fynewidget.Label:
		if *lbl == nil {
			*lbl = o
		}
	case *fyne.Container:
		for _, child := range o.Objects {
			findWidgets(child, btn, lbl)
		}
	}
}

func writeModuleFile(t *testing.T, dir, name, src string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatalf("write module %s: %v", name, err)
	}
	return p
}

// TestImportModuleFunctionReferencesModuleVariable verifies that an imported
// module function can reference module-level variables and functions, and also
// use host globals such as widget.
func TestImportModuleFunctionReferencesModuleVariable(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	dir := t.TempDir()
	mod := writeModuleFile(t, dir, "ui_utils.risor", `
let PREFIX = "Area: "
let square = (x) => x * x
let areaLabel = (r) => widget.NewLabel(PREFIX + string(3 * square(r)))
`)

	fw := NewWindow(w)
	fw.LoadScript(`
let u = import("` + mod + `")
window.SetContent(u.areaLabel(2))
`)
	fw.Execute()
	time.Sleep(150 * time.Millisecond)

	if fw.Status != "Ready!" {
		t.Fatalf("expected Ready!, got: %s", fw.Status)
	}
}

// TestImportModuleVariableReferenceNoCrash specifically guards against the
// previous "index out of range" failure when calling a module function that
// references a module-level variable.
func TestImportModuleVariableReferenceNoCrash(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	dir := t.TempDir()
	mod := writeModuleFile(t, dir, "math_utils.risor", `
let PI = 3.14159
let square = (x) => x * x
let circleArea = (r) => PI * square(r)
`)

	fw := NewWindow(w)
	fw.LoadScript(`
let m = import("` + mod + `")
let area = m.circleArea(2)
window.SetContent(widget.NewLabel("area=" + string(area)))
`)
	fw.Execute()
	time.Sleep(150 * time.Millisecond)

	if strings.Contains(fw.Status, "ERROR") {
		t.Fatalf("module variable reference failed: %s", fw.Status)
	}
	if fw.Status != "Ready!" {
		t.Fatalf("expected Ready!, got: %s", fw.Status)
	}
}

// TestImportModuleFunctionFromButtonCallback exercises the realistic GUI path:
// a button callback (invoked after the script finishes) calls a module function
// that references a module-level variable and function.
func TestImportModuleFunctionFromButtonCallback(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	dir := t.TempDir()
	mod := writeModuleFile(t, dir, "fmt_utils.risor", `
let PREFIX = ">> "
let emphasize = (s) => s.to_upper()
let tag = (s) => PREFIX + emphasize(s)
`)

	fw := NewWindow(w)
	fw.LoadScript(`
let u = import("` + mod + `")
let out = widget.NewLabel("init")
let b = widget.NewButton("go", () => { out.SetText(u.tag("hi")) })
window.SetContent(container.NewVBox([b, out]))
`)
	fw.Execute()

	// Wait for the script to finish executing.
	waitFor(t, 2*time.Second, func() bool { return fw.Status == "Ready!" })

	var btn *fynewidget.Button
	var lbl *fynewidget.Label
	findWidgets(fw.GetContentContainer(), &btn, &lbl)
	if btn == nil || lbl == nil {
		t.Fatalf("could not find button (%v) and label (%v) in content", btn, lbl)
	}

	// Tap the button; the callback runs on the Execute goroutine and calls the
	// module function. Poll until the label reflects the transformed text.
	test.Tap(btn)
	waitFor(t, 2*time.Second, func() bool { return lbl.Text == ">> HI" })

	if lbl.Text != ">> HI" {
		t.Fatalf("label text = %q, want %q (status: %s)", lbl.Text, ">> HI", fw.Status)
	}
}

// waitFor polls cond until it returns true or the timeout elapses.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
