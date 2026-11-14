package v3scene

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
)

func newTestApp(t testing.TB, mode Mode, doc, script string) *App {
	t.Helper()
	test.NewApp()
	guithread.SetMain()

	document, err := LoadDocument(doc)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	src, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	app, err := New(Config{
		Document: document,
		Script:   string(src),
		Mode:     mode,
		Base:     filepath.Dir(doc),
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(app.Stop)
	return app
}

func step(t *testing.T, app *App, event map[string]any) {
	t.Helper()
	if err := app.Step(context.Background(), event); err != nil {
		t.Fatalf("step %v: %v", event["type"], err)
	}
}

func labelText(t *testing.T, app *App, id string) string {
	t.Helper()
	obj, ok := app.View().LiveObject(id)
	if !ok {
		t.Fatalf("no live object %q", id)
	}
	label, ok := obj.(*widget.Label)
	if !ok {
		t.Fatalf("%q is %T, not *widget.Label", id, obj)
	}
	return label.Text
}

func TestCounter(t *testing.T) {
	app := newTestApp(t, ModeActions,
		"testdata/counter.scene.json",
		"testdata/counter/update.risor")

	step(t, app, map[string]any{"type": "init"})
	if got := app.State().Count; got != 0 {
		t.Fatalf("init count = %d, want 0", got)
	}
	if got := labelText(t, app, "count"); got != "0" {
		t.Fatalf("init label = %q, want 0", got)
	}

	step(t, app, map[string]any{"type": "tapped", "node": "inc"})
	if got := app.State().Count; got != 1 {
		t.Fatalf("after inc count = %d, want 1", got)
	}
	if got := labelText(t, app, "count"); got != "1" {
		t.Fatalf("after inc label = %q, want 1", got)
	}

	step(t, app, map[string]any{"type": "tapped", "node": "reset"})
	if got := app.State().Count; got != 0 {
		t.Fatalf("after reset count = %d, want 0", got)
	}
}

func TestDataToolActions(t *testing.T) {
	app := newTestApp(t, ModeActions,
		"testdata/customers/customers.scene.json",
		"testdata/customers/actions.risor")

	// init -> fetch action -> loaded event -> rows.
	step(t, app, map[string]any{"type": "init"})
	if got := len(app.State().Customers); got != 3 {
		t.Fatalf("after init customers = %d, want 3", got)
	}

	// add -> save action, state and table grow.
	step(t, app, map[string]any{
		"type": "tapped", "node": "add",
		"values": map[string]any{"name": "Katherine Johnson", "city": "Hampton"},
	})
	if got := len(app.State().Customers); got != 4 {
		t.Fatalf("after add customers = %d, want 4", got)
	}

	// The save action persisted in the host store, so refresh sees four.
	step(t, app, map[string]any{"type": "tapped", "node": "refresh"})
	if got := len(app.State().Customers); got != 4 {
		t.Fatalf("after refresh customers = %d, want 4", got)
	}

	// Invalid input: no save, no row, but a log line.
	step(t, app, map[string]any{
		"type": "tapped", "node": "add",
		"values": map[string]any{"name": "", "city": ""},
	})
	if got := len(app.State().Customers); got != 4 {
		t.Fatalf("after invalid add customers = %d, want 4", got)
	}
}

func TestDataToolHostFunctions(t *testing.T) {
	app := newTestApp(t, ModeHostFunctions,
		"testdata/customers/customers.scene.json",
		"testdata/customers/host-functions.risor")

	step(t, app, map[string]any{"type": "init"})
	if got := len(app.State().Customers); got != 3 {
		t.Fatalf("after init customers = %d, want 3", got)
	}

	step(t, app, map[string]any{
		"type": "tapped", "node": "add",
		"values": map[string]any{"name": "Katherine Johnson", "city": "Hampton"},
	})
	if got := len(app.State().Customers); got != 4 {
		t.Fatalf("after add customers = %d, want 4", got)
	}

	// A new program run sees the row the host function persisted.
	step(t, app, map[string]any{"type": "tapped", "node": "refresh"})
	if got := len(app.State().Customers); got != 4 {
		t.Fatalf("after refresh customers = %d, want 4", got)
	}
}

func TestCheckErrorsAtLoad(t *testing.T) {
	doc, err := LoadDocument("testdata/counter.scene.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("unknown state field", func(t *testing.T) {
		_, err := New(Config{
			Document: doc,
			Script:   `{state: state.nope, actions: []}`,
			Mode:     ModeActions,
		})
		if err == nil || !strings.Contains(err.Error(), "no field") {
			t.Fatalf("got %v, want a no-field check error", err)
		}
	})

	t.Run("incomplete action arm", func(t *testing.T) {
		_, err := New(Config{
			Document: doc,
			Script:   `{state: state, actions: [{do: "rows", node: "table"}]}`,
			Mode:     ModeActions,
		})
		if err == nil || !strings.Contains(err.Error(), "check") {
			t.Fatalf("got %v, want a check error", err)
		}
	})

	t.Run("host function in actions mode", func(t *testing.T) {
		_, err := New(Config{
			Document: doc,
			Script:   `{state: state, actions: [{do: "status", text: string(store.customers())}]}`,
			Mode:     ModeActions,
		})
		if err == nil || !strings.Contains(err.Error(), "check") {
			t.Fatalf("got %v, want store to be undefined in actions mode", err)
		}
	})
}

func TestStartEventLoopAndValues(t *testing.T) {
	app := newTestApp(t, ModeHostFunctions,
		"testdata/customers/customers.scene.json",
		"testdata/customers/host-functions.risor")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Set the inputs before the event loop starts; the test driver runs
	// fyne.Do inline, so touching widgets while the loop applies table rows
	// would race inside Fyne's test cache.
	if obj, ok := app.View().LiveObject("name"); ok {
		obj.(*widget.Entry).SetText("Katherine Johnson")
	}
	if obj, ok := app.View().LiveObject("city"); ok {
		obj.(*widget.Entry).SetText("Hampton")
	}

	app.Start(ctx)

	waitFor(t, func() bool { return len(app.State().Customers) == 3 })

	// The listener path gathers input values. Emit the same event a button
	// tap would while the loop is idle.
	app.View().EmitEvent("add", "tapped", nil)
	waitFor(t, func() bool { return len(app.State().Customers) == 4 })
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not reached")
}
