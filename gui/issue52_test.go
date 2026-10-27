package gui

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestStaleRunCannotInstallShortcut covers issue #52: a run that has been
// replaced but whose body is still unwinding must not install window state
// (shortcuts, menus, hooks) after the new run started.
func TestStaleRunCannotInstallShortcut(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	rec := &recorder{}
	var fw *Window
	nav := object.NewBuiltin("nav", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		fw.LoadScript(`rec("second run")`)
		fw.Execute()
		return object.Nil, nil
	})
	fw = NewWindow(w, WithGlobal("nav", nav), WithGlobal("rec", rec.builtin()))
	defer fw.Stop()

	fw.LoadScript(`
nav()
window.AddShortcut("Ctrl+S", () => { rec("stale shortcut") })
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return rec.has("second run") })
	time.Sleep(100 * time.Millisecond) // let the stale body unwind

	cur := fw.run.Load()
	fw.shortcutMutex.Lock()
	defer fw.shortcutMutex.Unlock()
	for key, e := range fw.shortcuts {
		if e.owner != cur {
			t.Fatalf("shortcut %q installed by a replaced run survived", key)
		}
	}
}

// TestConcurrentExecuteLeavesOneLiveRun covers issue #52: concurrent
// Execute calls (a host and a script's browser.Open, say) must leave exactly
// one live run; before, two calls could both replace the same predecessor
// and the loser kept running.
func TestConcurrentExecuteLeavesOneLiveRun(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")

	var live atomic.Int32
	block := object.NewBuiltin("block", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		live.Add(1)
		defer live.Add(-1)
		<-ctx.Done()
		return object.Nil, nil
	})
	fw := NewWindow(w, WithGlobal("block", block))
	fw.LoadScript(`block()`)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fw.Execute()
		}()
	}
	wg.Wait()
	waitFor(t, 3*time.Second, func() bool { return live.Load() == 1 })
	if n := live.Load(); n != 1 {
		t.Fatalf("%d runs alive after concurrent Execute calls, want 1", n)
	}
	fw.Stop()
	waitFor(t, 3*time.Second, func() bool { return live.Load() == 0 })
	if n := live.Load(); n != 0 {
		t.Fatalf("%d runs alive after Stop, want 0", n)
	}
}

// TestRegisterGlobalWhileRunning covers issue #52: RegisterGlobal wrote the
// environment map that running VMs read. Runs now snapshot the environment.
func TestRegisterGlobalWhileRunning(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	rec := &recorder{}
	fw := NewWindow(w, WithGlobal("rec", rec.builtin()))
	defer fw.Stop()

	idx := make([]string, 200)
	for i := range idx {
		idx[i] = fmt.Sprint(i)
	}
	fw.LoadScript(`
require("@rec")
[` + joinComma(idx) + `].each((i) => { rec("tick") })
rec("done")
`)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			fw.RegisterGlobal(fmt.Sprintf("g%d", i), object.NewInt(int64(i)))
		}
	}()
	for i := 0; i < 5; i++ {
		fw.Execute()
		time.Sleep(5 * time.Millisecond)
	}
	<-done
	waitFor(t, 3*time.Second, func() bool { return rec.has("done") })
	if !rec.has("done") {
		t.Fatalf("script did not finish: %q", fw.Status())
	}

	// A global registered before Execute is visible to that run.
	fw.RegisterGlobal("late", object.NewString("late value"))
	fw.LoadScript(`require("@late"); rec(late)`)
	if err := fw.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !rec.has("late value") {
		t.Fatal("global registered before Execute not visible to the run")
	}
}

func joinComma(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ","
		}
		out += x
	}
	return out
}
