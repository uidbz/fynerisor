package gui

import (
	"context"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// statusSignal returns an option that reports status changes over a channel.
// Tests use it instead of polling Window.Status so the status write and the
// test's read are ordered by the channel operation.
func statusSignal() (chan string, Option) {
	ch := make(chan string, 8)
	return ch, WithStatusCallback(func(status string) {
		select {
		case ch <- status:
		default:
		}
	})
}

func waitForStatus(t *testing.T, ch <-chan string, want string) {
	t.Helper()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case got := <-ch:
			if got == want {
				return
			}
			if strings.HasPrefix(got, "ERROR") {
				t.Fatalf("script failed: %s", got)
			}
		case <-timeout:
			t.Fatalf("timed out waiting for status %q", want)
		}
	}
}

// TestCanvasObjectEmbedsScript verifies that a fynerisor script can be hosted
// as a CanvasObject inside content owned by a normal Fyne application.
func TestCanvasObjectEmbedsScript(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("Host")

	status, withStatus := statusSignal()
	view := NewCanvasObject(`
		let label = widget.NewLabel("from script")
		window.SetContent(container.NewVBox(label))
	`, withStatus)
	if view.Window().FyneWindow != nil {
		t.Fatal("embedded view without WithHostWindow should not have a Fyne window")
	}
	view.Execute()
	waitForStatus(t, status, "Ready!")

	// The host owns the window content; the script gets only the view's area.
	// SetContent happens after the script is ready here to keep the headless
	// test driver single-threaded.
	host.SetContent(container.NewBorder(fynewidget.NewLabel("native header"), nil, nil, nil, view))

	content := view.Window().GetContentContainer()
	if len(content.Objects) != 1 {
		t.Fatalf("content has %d objects, want 1", len(content.Objects))
	}
	vbox, ok := content.Objects[0].(*fyne.Container)
	if !ok {
		t.Fatalf("script content is %T, want *fyne.Container", content.Objects[0])
	}
	if len(vbox.Objects) != 1 {
		t.Fatalf("vbox has %d children, want 1", len(vbox.Objects))
	}
	label, ok := vbox.Objects[0].(*fynewidget.Label)
	if !ok {
		t.Fatalf("script content child is %T, want *widget.Label", vbox.Objects[0])
	}
	if label.Text != "from script" {
		t.Fatalf("label text = %q, want %q", label.Text, "from script")
	}

	// The view must be renderable as an ordinary Fyne widget.
	renderer := test.WidgetRenderer(view)
	if renderer.MinSize().Width <= 1 {
		t.Fatalf("renderer MinSize = %v, expected script content to contribute size", renderer.MinSize())
	}
	if len(renderer.Objects()) != 1 || renderer.Objects()[0] != content {
		t.Fatalf("renderer does not render the script content container: %#v", renderer.Objects())
	}
}

// TestCanvasObjectHostWindow verifies that WithHostWindow enables window-level
// features without taking over the host window's content.
func TestCanvasObjectHostWindow(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("Host")

	status, withStatus := statusSignal()
	view := NewCanvasObject(`window.SetContent(widget.NewLabel("embedded"))`, WithHostWindow(host), withStatus)
	if view.Window().FyneWindow != host {
		t.Fatal("WithHostWindow did not set the host window")
	}
	if view.Window().Canvas() != host.Canvas() {
		t.Fatal("window.Canvas should be the host canvas")
	}

	view.Execute()
	waitForStatus(t, status, "Ready!")

	// Creating an embedded view must never replace the host content.
	host.SetContent(container.NewVBox(fynewidget.NewLabel("native"), view))
	hostContent, ok := host.Content().(*fyne.Container)
	if !ok || len(hostContent.Objects) != 2 {
		t.Fatalf("host content was replaced: %#v", host.Content())
	}
	if len(view.Window().GetContentContainer().Objects) != 1 {
		t.Fatal("script content was not rendered in the embedded view")
	}
}

// TestCanvasObjectWithoutHostGuards checks that window features return script
// errors instead of panicking when no host window was supplied.
func TestCanvasObjectWithoutHostGuards(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	view := NewCanvasObject("")
	w := view.Window()

	if w.Canvas() != nil {
		t.Fatal("Canvas() should be nil without a host window")
	}

	closeBuiltin, ok := w.GetAttr("Close")
	if !ok {
		t.Fatal("Close attr missing")
	}
	result, err := closeBuiltin.(*object.Builtin).Call(context.Background())
	if err != nil {
		t.Fatalf("Close call returned Go error: %v", err)
	}
	if _, ok := result.(*object.Error); !ok {
		t.Fatalf("Close without host returned %T, want *object.Error", result)
	}

	dialogBuiltin, ok := NewDialog(w).GetAttr("ShowInformation")
	if !ok {
		t.Fatal("dialog.ShowInformation attr missing")
	}
	result, err = dialogBuiltin.(*object.Builtin).Call(context.Background())
	if err != nil {
		t.Fatalf("dialog call returned Go error: %v", err)
	}
	if _, ok := result.(*object.Error); !ok {
		t.Fatalf("dialog without host returned %T, want *object.Error", result)
	}
}

// TestCanvasObjectCompileAndStop covers the convenience API around the
// embedded runtime.
func TestCanvasObjectCompileAndStop(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	view := NewCanvasObject("let x = ")
	if err := view.CompileScript(); err == nil {
		t.Fatal("CompileScript should report the syntax error")
	}

	status, withStatus := statusSignal()
	view = NewCanvasObject(`window.SetContent(widget.NewLabel("ok"))`, withStatus)
	if err := view.CompileScript(); err != nil {
		t.Fatalf("valid script failed to compile: %v", err)
	}
	view.Execute()
	waitForStatus(t, status, "Ready!")
	if view.Status() != "Ready!" {
		t.Fatalf("Status() = %q", view.Status())
	}
	view.Stop()
}
