package gui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynewidget "fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestPropertyReadsDoNotRaceTheUIThread covers issue #49: with a real UI
// goroutine (uiLoop), a script that keeps reading entry.Text while the user
// types must not race. Before the fix the read happened on the VM goroutine
// (`go test -race` reports it); now every attribute access runs on the UI
// thread.
func TestPropertyReadsDoNotRaceTheUIThread(t *testing.T) {
	u, a := startUILoop(t)

	started := make(chan struct{})
	ready := object.NewBuiltin("ready", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		close(started)
		return object.Nil, nil
	})
	rec := &recorder{}

	var w fyne.Window
	var fw *Window
	u.run(func() {
		w = a.NewWindow("Test")
		fw = NewWindow(w, WithGlobal("ready", ready), WithGlobal("rec", rec.builtin()))
	})
	defer fw.Stop()

	idx := make([]string, 300)
	for i := range idx {
		idx[i] = fmt.Sprint(i)
	}
	fw.LoadScript(`
let e = widget.NewEntry()
window.SetContent(e)
ready()
[` + strings.Join(idx, ",") + `].each((i) => {
    let t = e.Text
    e.PlaceHolder = t
})
rec("done")
`)
	fw.Execute()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatalf("script did not start: %q", fw.Status())
	}

	var entry *fynewidget.Entry
	u.run(func() { entry = findEntryInContent(w.Content()) })
	if entry == nil {
		t.Fatal("entry not found")
	}
	for i := 0; i < 300 && !rec.has("done"); i++ {
		u.run(func() { entry.SetText(fmt.Sprintf("typed %d", i)) })
	}
	waitFor(t, 5*time.Second, func() bool { return rec.has("done") })
	if !rec.has("done") {
		t.Fatalf("script did not finish: %q", fw.Status())
	}
}
