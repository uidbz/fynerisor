package scene_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// TestHandlerErrorVisibleInStatus pins the rule that a failing page.Bind
// handler surfaces in the app's own status UI. GUI builds have no stderr;
// without this, "button does nothing" failures were silently swallowed by
// the VM callback layer (the platepilot reader-PC bug).
func TestHandlerErrorVisibleInStatus(t *testing.T) {
	a := newApp(t)
	_, path := writeScene(t)
	var status bytes.Buffer
	fw := gui.NewWindow(a.NewWindow("scene"), gui.WithStrings(),
		guiscene.WithScene(guiscene.Config{Status: func(s string) { status.WriteString(s + "\n") }}))
	fw.LoadScript(`
		require(["@scene", "@strings"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		page.Bind({hit: (evt) => { strings.trim() }})
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)

	view := fw.GetContentContainer().Objects[0].(*render.View)
	view.EmitEvent("btn", "tapped", nil)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(status.String(), "handler hit:") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("status = %q, want handler hit: <error>", status.String())
}
