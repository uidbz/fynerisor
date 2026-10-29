package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
)

// This file covers issue #44 (gui half): script args and the app
// version are per-runtime configuration. Two embedded views in one
// process must not see each other's os.args() or app.version, and a
// require() version check must consult its own window's version.

func TestArgsAndVersionArePerRuntime(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var bufA, bufC safeBuffer
	viewA := NewCanvasObject(`
require(["v9.9"])
print(os.args()[0], app.version)
`, WithOS(), WithStdout(&bufA), WithScriptArgs("alpha", "--force"), WithAppVersion("9.9"))
	// Same require() requirement as A, but this window has no app
	// version: the check runs against fynerisor (v0.9.x) and must fail.
	viewB := NewCanvasObject(`require(["v9.9"])`, WithOS())
	// No per-runtime args: the process-global fallback is unset here, so
	// the script must see an empty list — never A's args.
	viewC := NewCanvasObject(`print(os.args())`, WithOS(), WithStdout(&bufC))

	host.SetContent(container.NewVBox(viewA, viewB, viewC))
	defer viewA.Stop()
	defer viewB.Stop()
	defer viewC.Stop()

	viewA.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewA.Status() == "Ready!" })
	if viewA.Status() != "Ready!" {
		t.Fatalf("view A failed despite its own app version: %s", viewA.Status())
	}
	if got := bufA.String(); got != "alpha 9.9\n" {
		t.Errorf("view A printed %q, want %q", got, "alpha 9.9\n")
	}

	viewB.Execute()
	waitFor(t, 2*time.Second, func() bool { return strings.HasPrefix(viewB.Status(), "ERROR") })
	if viewB.Status() == "Ready!" || !strings.Contains(viewB.Status(), "v9.9") {
		t.Fatalf("view B status = %q, want a v9.9 version failure (version leaked from A)", viewB.Status())
	}

	viewC.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewC.Status() == "Ready!" })
	if got := bufC.String(); got != "[]\n" {
		t.Errorf("view C printed %q, want %q (view A's args leaked)", got, "[]\n")
	}
}
