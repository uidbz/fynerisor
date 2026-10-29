package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"

	risorwidget "github.com/uidbz/fynerisor/gui/widget"
)

// This file covers issue #38 (first batch): attribute registries and
// the uniform CanvasObject verb set.

func TestLabelRegistrySurface(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var out safeBuffer
	view := NewCanvasObject(`
let l = widget.NewLabel("hi")
window.SetContent(l)
l.SetText("bye")
print(l.Text + "|" + string(l.Visible()) + "|" + string(l.MinSize().width >= 0))
l.Resize(50, 50)
l.Move(5, 5)
l.Hide()
print(string(l.Visible()))
l.Refresh()
l.Show()
print(string(l.Visible()))
`, WithStdout(&out))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 3*time.Second, func() bool { return view.Status() == "Ready!" })
	if view.Status() != "Ready!" {
		t.Fatalf("script failed: %s", view.Status())
	}
	want := "bye|true|true\nfalse\ntrue\n"
	if out.String() != want {
		t.Errorf("output = %q, want %q", out.String(), want)
	}
}

// Arity checks now come from the registry with a clearer message.
func TestLabelRegistryArityRaises(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	view := NewCanvasObject(`let l = widget.NewLabel("x")
l.SetText()`)
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return strings.HasPrefix(view.Status(), "ERROR") })
	if !strings.Contains(view.Status(), "SetText: expected 1 argument, got 0") {
		t.Errorf("status = %q, want the registry arity error", view.Status())
	}
}

// Attrs() is real introspection data now (drives docs/tooling).
func TestLabelAttrsIntrospection(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	l := risorwidget.NewLabel("x")
	specs := l.Attrs()
	if len(specs) == 0 {
		t.Fatal("Label.Attrs() is still nil")
	}
	names := map[string]bool{}
	for _, s := range specs {
		names[s.Name] = true
	}
	for _, want := range []string{"Text", "SetText", "Show", "Hide", "Refresh", "Resize", "Move", "MinSize", "Visible", "TextStyle"} {
		if !names[want] {
			t.Errorf("Attrs() missing %q", want)
		}
	}
	if specs[0].Doc == "" {
		t.Error("AttrSpec.Doc empty: registry docs not flowing through")
	}
}
