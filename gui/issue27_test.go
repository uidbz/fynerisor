package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"

	"github.com/deepnoodle-ai/risor/v2"
)

// This file covers issue #27 (gui half): imported modules are compiled
// with exactly the globals their VM receives. WithGlobal values are
// visible to modules; names injected via opaque WithRisorOptions fail
// the import with a clear error instead of failing at runtime.

func TestImportGlobalsMatchModuleVM(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	dir := t.TempDir()
	mod := filepath.Join(dir, "uses.risor")
	if err := os.WriteFile(mod, []byte("let hi = () => cfg[\"name\"] + \"!\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(dir, "opaque.risor")
	if err := os.WriteFile(bad, []byte("let hi = () => smuggled\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf safeBuffer
	viewA := NewCanvasObject(`
let m = import("`+mod+`")
print(m.hi())
`, WithStdout(&buf), WithGlobal("cfg", map[string]any{"name": "widget"}))
	viewB := NewCanvasObject(`
let m = import("`+bad+`")
print(m.hi())
`, WithRisorOptions(risor.WithEnv(map[string]any{"smuggled": "x"})))

	host.SetContent(container.NewVBox(viewA, viewB))
	defer viewA.Stop()
	defer viewB.Stop()

	viewA.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewA.Status() == "Ready!" })
	if viewA.Status() != "Ready!" {
		t.Fatalf("view A failed: %s", viewA.Status())
	}
	if got := buf.String(); got != "widget!\n" {
		t.Errorf("module output = %q, want %q", got, "widget!\n")
	}

	viewB.Execute()
	waitFor(t, 2*time.Second, func() bool { return strings.HasPrefix(viewB.Status(), "ERROR") })
	if !strings.Contains(viewB.Status(), "smuggled") {
		t.Errorf("view B status = %q, want an error naming the undefined global", viewB.Status())
	}
}
