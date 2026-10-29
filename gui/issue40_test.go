package gui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"

	"github.com/uidbz/fynerisor/core"
)

// This file covers issue #40 (gui half): gui.Option is core.Option, so
// options cross package boundaries in both directions.

func TestCoreOptionsWorkInGui(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var out safeBuffer
	// A core option (not a gui one) enables json for this embedded view.
	view := NewCanvasObject(`print(json.format([1, 2]))`, core.WithJSON(), WithStdout(&out))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	if view.Status() != "Ready!" {
		t.Fatalf("script failed: %s", view.Status())
	}
	if !strings.Contains(out.String(), "[1,2]") {
		t.Errorf("output = %q, want json.format result [1,2]", out.String())
	}

	// Without the option json is genuinely absent: the positive case
	// above really came from core.WithJSON().
	view2 := NewCanvasObject(`json.format([1])`)
	host.SetContent(container.NewVBox(view2))
	defer view2.Stop()
	view2.Execute()
	waitFor(t, 2*time.Second, func() bool { return strings.HasPrefix(view2.Status(), "ERROR") })
	if !strings.Contains(view2.Status(), "json") {
		t.Errorf("status = %q, want undefined json", view2.Status())
	}
}

func TestGuiOptionsAcceptedByCore(t *testing.T) {
	// gui-defined options are ordinary core.Options; a headless Context
	// simply has no status concept, and building one must not break.
	cb := core.NewContext(
		WithStatusCallback(func(string) {}),
		WithErrorHandler(func(error) {}),
		WithAppName("from-gui-option"),
	)
	got, err := cb.Eval(`app.name`)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if fmt.Sprint(got) != "from-gui-option" {
		t.Errorf("app.name = %v, want from-gui-option", got)
	}
}
