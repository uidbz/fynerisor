package gui

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func findTree(obj fyne.CanvasObject) *fynewidget.Tree {
	switch o := obj.(type) {
	case *fynewidget.Tree:
		return o
	case *fyne.Container:
		for _, child := range o.Objects {
			if tr := findTree(child); tr != nil {
				return tr
			}
		}
	}
	return nil
}

// TestOnTappedPanicIsRecovered covers issue #12: Button.OnTapped called the
// VM directly on the UI thread without panic recovery, so a Go panic inside
// a builtin used by the handler crashed the whole process.
func TestOnTappedPanicIsRecovered(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	boom := object.NewBuiltin("boom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		panic("boom from a builtin")
	})
	fw := NewWindow(w, WithGlobal("boom", boom))
	defer fw.Stop()

	fw.LoadScript(`
let btn = widget.NewButton("tap", () => {})
btn.OnTapped(() => {
    boom()
})
let ok = widget.NewButton("ok", () => { window.SetStatus("still alive") })
window.SetContent(container.NewVBox(btn, ok))
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "Ready!" })

	var buttons []*fynewidget.Button
	var collect func(fyne.CanvasObject)
	collect = func(o fyne.CanvasObject) {
		switch v := o.(type) {
		case *fynewidget.Button:
			buttons = append(buttons, v)
		case *fyne.Container:
			for _, c := range v.Objects {
				collect(c)
			}
		}
	}
	collect(w.Content())
	if len(buttons) != 2 {
		t.Fatalf("expected 2 buttons, found %d", len(buttons))
	}

	test.Tap(buttons[0]) // panics inside the handler: must be recovered
	test.Tap(buttons[1]) // the run keeps processing events afterwards
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "still alive" })
	if fw.Status() != "still alive" {
		t.Fatalf("events stopped after a panicking handler, status=%q", fw.Status())
	}
}

// TestTreeGetterDoesNotEnterBusyVM covers issue #12: Tree getters called the
// VM directly on the UI thread, concurrently with a handler running on the
// VM goroutine. They must now be turned away (and fall back to last-known
// values) while the VM is busy.
func TestTreeGetterDoesNotEnterBusyVM(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	entered := make(chan struct{})
	release := make(chan struct{})
	var getterCalls atomic.Int32
	probe := object.NewBuiltin("probe", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		getterCalls.Add(1)
		return object.NewList([]object.Object{object.NewString("a"), object.NewString("b")}), nil
	})
	fw := NewWindow(w,
		WithGlobal("block", blockingBuiltin(entered, release)),
		WithGlobal("probe", probe))
	defer fw.Stop()

	fw.LoadScript(`
let tree = widget.NewTree()
tree.ChildUIDs((uid) => probe())
tree.IsBranch((uid) => uid == "")
tree.CreateNode((branch) => widget.NewLabel(""))
tree.UpdateNode((uid, branch, obj) => {})
let btn = widget.NewButton("busy", () => { block() })
window.SetContent(container.NewVBox(btn, tree))
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "Ready!" })

	tree := findTree(w.Content())
	if tree == nil {
		t.Fatal("tree not found")
	}
	// Prime the last-known cache while the VM is idle.
	if got := tree.ChildUIDs(""); len(got) != 2 {
		t.Fatalf("idle ChildUIDs returned %v, want 2 children", got)
	}

	test.Tap(findButton(w.Content()))
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never started")
	}

	before := getterCalls.Load()
	got := tree.ChildUIDs("")
	if getterCalls.Load() != before {
		close(release)
		t.Fatal("Tree.ChildUIDs entered the VM while a handler was running in it")
	}
	if len(got) != 2 {
		t.Fatalf("busy ChildUIDs returned %v, want the last-known 2 children", got)
	}
	close(release)
}
