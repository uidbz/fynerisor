package gui

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestStopCancelsRunningScript covers issue #9: the Execute context must be
// passed to the Risor VM so Stop() interrupts a running script instead of
// leaving it spinning.
func TestStopCancelsRunningScript(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	// slow() fails as soon as its context is cancelled, and otherwise
	// returns after 1 s. With the fix, Stop() cancels the VM context, the
	// builtin raises, and the script is abandoned before it can set a
	// status; without the fix the builtin returns normally and the status
	// line below runs.
	slow := object.NewBuiltin("slow", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		select {
		case <-ctx.Done():
			return nil, object.Errorf("slow: context cancelled")
		case <-time.After(1 * time.Second):
			return object.NewString("timeout"), nil
		}
	})

	fw := NewWindow(w, WithGlobal("slow", slow))
	fw.LoadScript(`
let r = slow()
window.SetStatus("finished: " + r)
`)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)

	if fw.Status != "" {
		t.Fatalf("script finished before Stop(): %q", fw.Status)
	}

	fw.Stop()

	// After Stop the VM must observe the cancelled context; no further
	// statement of the script body may run.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(25 * time.Millisecond)
		if fw.Status != "" {
			t.Fatalf("script kept running after Stop: status=%q", fw.Status)
		}
	}
}
