package gui

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func findList(obj fyne.CanvasObject) *fynewidget.List {
	switch o := obj.(type) {
	case *fynewidget.List:
		return o
	case *fyne.Container:
		for _, child := range o.Objects {
			if l := findList(child); l != nil {
				return l
			}
		}
	}
	return nil
}

// blockingBuiltin returns a builtin that signals entered and then blocks
// until release is closed (or the run is stopped).
func blockingBuiltin(entered chan<- struct{}, release <-chan struct{}) *object.Builtin {
	return object.NewBuiltin("block", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return object.Nil, nil
	})
}

// TestTwoViewsDoNotBlockEachOther covers issue #13: the VM guard used to be
// one process-wide variable, so while view A ran a handler, view B's render
// getters were rejected as "concurrent VM access" although B has its own VM.
func TestTwoViewsDoNotBlockEachOther(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)

	viewA := NewCanvasObject(`
let btn = widget.NewButton("busy", () => { block() })
window.SetContent(btn)
`, WithGlobal("block", blockingBuiltin(entered, release)))
	var bCalls atomic.Int32
	probe := object.NewBuiltin("probe", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		bCalls.Add(1)
		return object.NewInt(3), nil
	})
	viewB := NewCanvasObject(`
let l = widget.NewList()
l.Length(() => probe())
l.CreateItem(() => widget.NewLabel("x"))
l.UpdateItem((id, item) => {})
window.SetContent(l)
`, WithGlobal("probe", probe))
	host.SetContent(container.NewVBox(viewA, viewB))
	defer viewA.Stop()
	defer viewB.Stop()

	// Start the views one after the other: Fyne's test driver runs fyne.Do
	// inline on the calling goroutine, so two bodies rendering at once would
	// race inside Fyne itself (the real driver serializes on the UI thread).
	viewA.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewA.Status() == "Ready!" })
	viewB.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewB.Status() == "Ready!" })
	if viewA.Status() != "Ready!" || viewB.Status() != "Ready!" {
		t.Fatalf("scripts did not finish: A=%q B=%q", viewA.Status(), viewB.Status())
	}

	btn := findButton(viewA.Window().GetContentContainer())
	if btn == nil {
		t.Fatal("button not found in view A")
	}
	test.Tap(btn)
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("view A's handler never started")
	}

	// View A is now inside its VM. View B's getter must still answer.
	list := findList(viewB.Window().GetContentContainer())
	if list == nil {
		t.Fatal("list not found in view B")
	}
	before := bCalls.Load()
	if n := list.Length(); n != 3 {
		t.Fatalf("view B's List.Length returned %d while view A was busy, want 3", n)
	}
	if bCalls.Load() == before {
		t.Fatal("view B's Length getter was rejected (stale value returned) while view A was busy")
	}
}
