package gui

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynebinding "fyne.io/fyne/v2/data/binding"
	fynewidget "fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	risorbinding "github.com/uidbz/fynerisor/gui/binding"
)

// This file covers issue #45: regression tests for the lifecycle and
// thread-safety properties the October 2026 review asked to lock down.
// (The widget-level regressions from the same review live in
// issue10/11/12/13/16/17/21/49/50_test.go; -race now runs in CI.)

// TestReplacedRunIsReclaimed proves the per-run model leaks nothing on the
// fynerisor side: after a run is replaced, everything the window kept on
// its behalf — the run itself, its callbacks, and its VM environment —
// becomes garbage-collectable, without any manual Destroy or finalizer on
// script handles.
//
// The script stores a probe object in a global and references it from an
// OnDropped handler and a shortcut handler, the window-owned registrations
// that resetRunState must drop. If any of them (or the replaced Run, or
// the window's run pointer) survived, the probe's finalizer never fires.
//
// Deliberately no content widgets here: Fyne keeps painted widgets in its
// own expiring widget cache (docs/CONCURRENCY.md, ValidDuration, one
// minute by default), so tree teardown is bounded by Fyne's cache sweep,
// not by anything fynerisor holds.
func TestReplacedRunIsReclaimed(t *testing.T) {
	u, a := startUILoop(t)

	reclaimed := make(chan struct{})
	sentinel := object.NewBuiltin("sentinel", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		s := object.NewString("gc-probe")
		runtime.SetFinalizer(s, func(*object.String) { close(reclaimed) })
		return s, nil
	})

	var fw *Window
	u.run(func() {
		fw = NewWindow(a.NewWindow("Reclaim"), WithGlobal("sentinel", sentinel))
	})
	defer fw.Stop()

	fw.LoadScript(`
let p = sentinel()
window.OnDropped(() => p.Inspect())
window.AddShortcut("Ctrl+K", () => p.Inspect())
`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	// Replace the run: resetRunState drops the shortcuts, the drop handler,
	// and the stale Run itself.
	fw.LoadScript(`window.SetStatus("second")`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "second" })
	u.run(func() {}) // drain the UI queue: resetRunState's cleanup runs there

	for i := 0; i < 50; i++ {
		runtime.GC()
		select {
		case <-reclaimed:
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
	t.Fatal("the replaced run's objects were never reclaimed: a stale reference keeps the old VM environment alive")
}

// TestBindingConcurrentWritesKeepLatestValue locks down the binding-layer
// thread-safety contract (issue #45, phase 4):
//
//   - A Go host may Set a binding from any goroutine, even many at once:
//     fyne's binding storage is internally locked and the wrapper adds no
//     unsynchronized state of its own.
//   - The script's AddListener handler never runs on the writer's goroutine:
//     each notification becomes a coalesced queue event (issue #51), and
//     the value written last is always observed.
func TestBindingConcurrentWritesKeepLatestValue(t *testing.T) {
	u, a := startUILoop(t)

	var mu sync.Mutex
	observations := 0
	var lastSeen int64
	record := object.NewBuiltin("record", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return object.Errorf("record: expected 1 argument, got %d", len(args)), nil
		}
		v, err := object.AsInt(args[0])
		if err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		observations++
		lastSeen = v
		return object.Nil, nil
	})

	b := risorbinding.NewInt()

	var fw *Window
	u.run(func() {
		fw = NewWindow(a.NewWindow("BindingRace"), WithGlobal("b", b), WithGlobal("record", record))
	})
	defer fw.Stop()

	fw.LoadScript(`b.AddListener(() => record(b.Get()))`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	fb := b.Interface().(fynebinding.Int)

	// Hammer the binding from the host side. Under -race this detects any
	// unsynchronized access in the binding layer; correctness below proves
	// the events are never dropped.
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(base int64) {
			defer wg.Done()
			for i := int64(0); i < 100; i++ {
				_ = fb.Set(int(base + i))
			}
		}(int64(g) * 1000)
	}
	wg.Wait()

	// The last write must survive the storm (coalescing keeps the newest).
	const final = 4242
	_ = fb.Set(final)
	waitFor(t, 5*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return lastSeen == final
	})
	mu.Lock()
	defer mu.Unlock()
	if observations == 0 {
		t.Fatal("the script listener never ran")
	}
}

// TestBindingBoundWidgetSyncFromHostGoroutines is the widget half of the
// same contract: a host goroutine writing a binding that a script's widget
// is bound to must neither race nor lose the final value. The widget is
// touched only on the UI thread, the write happens on the VM goroutine.
func TestBindingBoundWidgetSyncFromHostGoroutines(t *testing.T) {
	u, a := startUILoop(t)

	s := risorbinding.NewString()

	var fw *Window
	u.run(func() {
		fw = NewWindow(a.NewWindow("BindSync"))
	})
	defer fw.Stop()

	fw.LoadScript(`
let l = widget.NewLabel("waiting")
window.SetContent(l)
`)
	// Bind via Go instead of the script binding global so the host owns one
	// side of the contract: this is the embedding scenario (a data source
	// feeding widgets from its own goroutines).
	fb := s.Interface().(fynebinding.String)

	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	// Find the label in the content and bind it to the host-side binding.
	var label *fynewidget.Label
	u.run(func() {
		label = findLabel(fw.GetContentContainer().Objects[0])
		if label != nil {
			label.Bind(fb)
		}
	})
	if label == nil {
		t.Fatal("label not found in content")
	}

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(base string) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = fb.Set(fmt.Sprintf("%s-%d", base, i))
			}
		}(fmt.Sprintf("host%d", g))
	}
	wg.Wait()
	_ = fb.Set("final-value")

	waitFor(t, 5*time.Second, func() bool {
		var text string
		u.run(func() { text = label.Text })
		return text == "final-value"
	})
}

func findLabel(obj fyne.CanvasObject) *fynewidget.Label {
	switch o := obj.(type) {
	case *fynewidget.Label:
		return o
	case *fyne.Container:
		for _, child := range o.Objects {
			if l := findLabel(child); l != nil {
				return l
			}
		}
	}
	return nil
}
