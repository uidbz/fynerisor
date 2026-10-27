package gui

import (
	"context"
	"testing"


	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Issue #2: window.Close() and app.Quit() must be bound — the README and
// docs/EXAMPLES.md already use them.
func TestCloseAndQuitBound(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	w := a.NewWindow("issue2")
	fw := NewWindow(w)

	closeAttr, ok := fw.GetAttr("Close")
	if !ok {
		t.Fatal("window.Close not bound")
	}
	if closeAttr.Type() != object.BUILTIN {
		t.Fatalf("window.Close: expected builtin, got %s", closeAttr.Type())
	}

	appObj := newAppObject(fw)
	quitAttr, ok := appObj.GetAttr("Quit")
	if !ok {
		t.Fatal("app.Quit not bound")
	}
	if quitAttr.Type() != object.BUILTIN {
		t.Fatalf("app.Quit: expected builtin, got %s", quitAttr.Type())
	}

	// NewWindow hosts keep their own fyne.App: Quit reports unavailable
	// rather than panicking.
	b := quitAttr.(*object.Builtin)
	res, err := b.Call(context.Background())
	if err == nil && (res == nil || res.Type() != object.ERROR) {
		t.Fatal("app.Quit with nil FyneApp: expected error result")
	}

	// With NewApp the owning app is recorded (FyneApp), so Quit resolves.
	fw.FyneApp = a
	res, err = b.Call(context.Background())
	if err != nil {
		t.Fatalf("app.Quit with FyneApp: %v", err)
	}
	if res.Type() == object.ERROR {
		t.Fatalf("app.Quit with FyneApp failed: %v", res)
	}
}
