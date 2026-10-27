package gui

import (
	"context"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestGoRunsAfterBodyInOrder pins the semantics decided in issue #17: go(fn)
// never runs in parallel with the script — fn is queued on the script's VM
// goroutine and runs after the current body, in call order — and a panic in
// fn is recovered without stopping later work.
func TestGoRunsAfterBodyInOrder(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	rec := &recorder{}
	boom := object.NewBuiltin("boom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		panic("boom inside go()")
	})
	fw := NewWindow(w, WithGlobal("rec", rec.builtin()), WithGlobal("boom", boom))
	defer fw.Stop()

	fw.LoadScript(`
go(() => { rec("go 1") })
go(() => { boom() })
go(() => { rec("go 3") })
rec("body end")
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return rec.has("go 3") })

	rec.mu.Lock()
	defer rec.mu.Unlock()
	want := []string{"body end", "go 1", "go 3"}
	if len(rec.seen) != len(want) {
		t.Fatalf("got %v, want %v", rec.seen, want)
	}
	for i := range want {
		if rec.seen[i] != want[i] {
			t.Fatalf("got %v, want %v (go() must run after the body, in order)", rec.seen, want)
		}
	}
}
