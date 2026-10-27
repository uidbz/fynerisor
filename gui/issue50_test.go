package gui

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// This file covers issue #50: a List/GridWrap/Tree template request that
// hits a busy VM must not answer with a placeholder that Fyne then pools
// forever. Fyne holds the object a template getter returns for the
// lifetime of the row, so a placeholder would be what every later
// UpdateItem hands the script — a Label where the script expects its own
// widget. The Table form of the regression lives in the tablewidget
// package (flextable_slot_test.go).

// cellTracker records, per row (List/GridWrap id or Tree uid), the type of
// the widget the script's update callback was handed. After a render that
// ran against a busy VM, every entry must still be the script's widget
// type; a "widget.Label" entry means a busy-VM placeholder was pooled.
type cellTracker struct {
	mu   sync.Mutex
	seen map[string]string
}

func newCellTracker() *cellTracker {
	return &cellTracker{seen: map[string]string{}}
}

// builtin returns a script builtin cell(key, item) for update callbacks.
func (c *cellTracker) builtin(name string) *object.Builtin {
	return object.NewBuiltin(name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return object.Errorf("%s: expected 2 arguments, got %d", name, len(args)), nil
		}
		key := strings.Trim(args[0].Inspect(), `"`)
		c.mu.Lock()
		defer c.mu.Unlock()
		c.seen[key] = string(args[1].Type())
		return object.Nil, nil
	})
}

func (c *cellTracker) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.seen)
}

func (c *cellTracker) ofType(t string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, typ := range c.seen {
		if typ == t {
			n++
		}
	}
	return n
}

// findGridWrap walks a content tree for fyne's GridWrap widget.
func findGridWrap(obj fyne.CanvasObject) *fynewidget.GridWrap {
	switch o := obj.(type) {
	case *fynewidget.GridWrap:
		return o
	case *fyne.Container:
		for _, child := range o.Objects {
			if g := findGridWrap(child); g != nil {
				return g
			}
		}
	}
	return nil
}

// TestBusyVMListTemplateNotPooled renders a List whose template callback
// must run while the VM is busy, and requires every visible row to end up
// holding the script's Card — never a pooled placeholder.
//
// Choreography: the script lists twenty Card rows in a short window, so
// only some rows exist when it goes idle. A button tap then parks the
// run's VM goroutine inside block(); while it is parked the test grows the
// window, forcing Fyne to call CreateItem for the newly visible rows —
// against a busy VM. Before the fix those rows were pooled as empty Labels
// forever (UpdateItem then saw widget.Label and SetTitle failed). Now
// CreateItem hands Fyne an empty slot that is filled with the script's
// Card by the first UpdateItem that can run, once the VM is free.
func TestBusyVMListTemplateNotPooled(t *testing.T) {
	u, a := startUILoop(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseAll()
	cells := newCellTracker()

	var fw *Window
	u.run(func() {
		w := a.NewWindow("Test")
		w.Resize(fyne.NewSize(400, 200))
		fw = NewWindow(w,
			WithGlobal("block", blockingBuiltin(entered, release)),
			WithGlobal("cell", cells.builtin("cell")))
	})
	defer fw.Stop()

	fw.LoadScript(`
let l = widget.NewList()
l.Length(() => 20)
l.CreateItem(() => widget.NewCard("c", "", widget.NewLabel("x")))
l.UpdateItem((id, item) => {
    cell(id, item)
    item.SetTitle("row")
})
let b = widget.NewButton("busy", () => { block() })
window.SetContent(container.NewBorder(b, nil, nil, nil, l))
`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	var list *fynewidget.List
	var btn *fynewidget.Button
	u.run(func() {
		content := fw.FyneWindow.Content()
		list = findList(content)
		btn = findButton(content)
	})
	if list == nil || btn == nil {
		t.Fatal("list/button not found")
	}

	// The idle render must have produced real Cards, and fewer than the
	// twenty rows: the rest are what the busy resize will force to create.
	waitFor(t, 5*time.Second, func() bool { return cells.count() > 0 })
	if n := cells.count(); n >= 20 {
		t.Fatalf("initial render already showed all 20 rows (%d distinct); window too tall for the test", n)
	}
	var rowH float32
	u.run(func() { rowH = fynewidget.NewCard("c", "", fynewidget.NewLabel("x")).MinSize().Height })

	// Park the VM in a handler, then grow the window while it is parked.
	u.run(func() { test.Tap(btn) })
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler never started")
	}
	u.run(func() { fw.FyneWindow.Resize(fyne.NewSize(400, 400+20*rowH)) })

	// Free the VM; the pending slots must become the script's Cards.
	releaseAll()
	waitFor(t, 10*time.Second, func() bool { return cells.count() == 20 })
	if n := cells.count(); n != 20 {
		t.Fatalf("only %d of 20 rows ever reached UpdateItem; the resized list never rendered fully", n)
	}
	if n := cells.ofType("widget.Label"); n != 0 {
		t.Fatalf("%d rows hand UpdateItem a pooled placeholder Label; the busy-VM template was pooled (issue #50)", n)
	}
}

// TestBusyVMTreeTemplateNotPooled scrolls a Tree while the script's VM is
// parked: nodes that scroll into view need fresh node widgets, so Fyne
// calls CreateNode against a busy VM (ChildUIDs falls back to its cache).
// Those nodes must end up holding the script's Card, not a pooled Label
// (issue #50).
func TestBusyVMTreeTemplateNotPooled(t *testing.T) {
	u, a := startUILoop(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseAll()
	cells := newCellTracker()

	var fw *Window
	u.run(func() {
		w := a.NewWindow("Test")
		w.Resize(fyne.NewSize(400, 200))
		fw = NewWindow(w,
			WithGlobal("block", blockingBuiltin(entered, release)),
			WithGlobal("cell", cells.builtin("cell")))
	})
	defer fw.Stop()

	fw.LoadScript(`
let t = widget.NewTree()
t.ChildUIDs((uid) => match uid {
    "" => ["m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", "m14", "m15", "m16", "m17", "m18", "m19", "m20"]
    _ => []
})
t.IsBranch((uid) => uid == "")
t.CreateNode((branch) => widget.NewCard("n", "", widget.NewLabel("x")))
t.UpdateNode((uid, branch, item) => {
    cell(uid, item)
    item.SetTitle("node")
})
let b = widget.NewButton("busy", () => { block() })
window.SetContent(container.NewBorder(b, nil, nil, nil, t))
t.OpenBranch("")
`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	var tree *fynewidget.Tree
	var btn *fynewidget.Button
	u.run(func() {
		content := fw.FyneWindow.Content()
		tree = findTree(content)
		btn = findButton(content)
	})
	if tree == nil || btn == nil {
		t.Fatal("tree/button not found")
	}

	// Idle render: the nodes in view are real Cards, and fewer than the
	// 21 open nodes, so scrolling must create fresh ones.
	waitFor(t, 5*time.Second, func() bool { return cells.count() > 0 })
	before := cells.count()
	if before >= 20 {
		t.Fatalf("initial render already showed all 20 nodes (%d distinct); window too tall for the test", before)
	}

	// Park the VM, then scroll while the getters run against it.
	u.run(func() { test.Tap(btn) })
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler never started")
	}
	u.run(func() { tree.ScrollToBottom() })

	// Free the VM; the pending slots must become the script's Cards.
	releaseAll()
	waitFor(t, 10*time.Second, func() bool { return cells.count() >= before+2 })
	if n := cells.count(); n < before+2 {
		t.Fatalf("only %d nodes reached UpdateItem after scrolling (%d before); the scroll never rendered new nodes", n, before)
	}
	if n := cells.ofType("widget.Label"); n != 0 {
		t.Fatalf("%d nodes hand UpdateNode a pooled placeholder Label; the busy-VM template was pooled (issue #50)", n)
	}
}

// TestBusyVMGridWrapTemplateNotPooled is the GridWrap form of the List
// test: growing the grid while the VM is parked forces fresh template
// cells, and every cell must end up holding the script's Card.
func TestBusyVMGridWrapTemplateNotPooled(t *testing.T) {
	u, a := startUILoop(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseAll()
	cells := newCellTracker()

	var fw *Window
	u.run(func() {
		w := a.NewWindow("Test")
		w.Resize(fyne.NewSize(400, 110))
		fw = NewWindow(w,
			WithGlobal("block", blockingBuiltin(entered, release)),
			WithGlobal("cell", cells.builtin("cell")))
	})
	defer fw.Stop()

	// A hundred small cells: the tiny viewport shows only the first rows,
	// so scrolling to the bottom must create the rest of the templates.
	fw.LoadScript(`
let g = widget.NewGridWrap()
g.Length(() => 100)
g.CreateItem(() => widget.NewCard("c", "", widget.NewLabel("x")))
g.UpdateItem((id, item) => {
    cell(id, item)
    item.SetTitle("cell")
})
let b = widget.NewButton("busy", () => { block() })
window.SetContent(container.NewBorder(b, nil, nil, nil, g))
`)
	fw.Execute()
	waitFor(t, 5*time.Second, func() bool { return fw.Status() == "Ready!" })

	var grid *fynewidget.GridWrap
	var btn *fynewidget.Button
	u.run(func() {
		content := fw.FyneWindow.Content()
		grid = findGridWrap(content)
		btn = findButton(content)
	})
	if grid == nil || btn == nil {
		t.Fatal("gridwrap/button not found")
	}

	waitFor(t, 5*time.Second, func() bool { return cells.count() >= 1 })

	// Park the VM, then scroll to the bottom: rows never seen before must
	// get their template created while the VM is busy.
	u.run(func() { test.Tap(btn) })
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler never started")
	}
	u.run(func() { grid.ScrollToBottom() })

	// Free the VM. Every cell must hold the script's Card by now, and no
	// cell may have been handed a pooled placeholder along the way.
	releaseAll()
	waitFor(t, 10*time.Second, func() bool { return cells.count() == 100 })
	if n := cells.count(); n != 100 {
		t.Fatalf("only %d of 100 cells ever reached UpdateItem", n)
	}
	if n := cells.ofType("widget.Label"); n != 0 {
		t.Fatalf("%d cells hand UpdateItem a pooled placeholder Label; the busy-VM template was pooled (issue #50)", n)
	}
}
