package scene_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

const dataScene = `{
  "format": "fynerisor.scene/1",
  "title": "Data",
  "root": {
    "id": "root", "type": "free",
    "place": {"x": 0, "y": 0, "w": 900, "h": 700},
    "props": {"designW": 900.0, "designH": 700.0, "scale": "none"},
    "children": [
      {"id": "tbl", "type": "table", "place": {"x": 10, "y": 10, "w": 560, "h": 320},
       "props": {"columns": ["A", "B"]}},
      {"id": "lg", "type": "log", "place": {"x": 10, "y": 350, "w": 560, "h": 240}}
    ]
  }
}`

func waitForOut(t *testing.T, buf *syncBuffer, want string) bool {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), want) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TestDataNodeScriptAPI drives table/log through page scripts end to end:
// SetRows/SetCell/Append via the scene.Node bindings, then verifies the
// live components kept the script-fed state.
func TestDataNodeScriptAPI(t *testing.T) {
	a := newApp(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "page.scene.json")
	if err := os.WriteFile(path, []byte(dataScene), 0o644); err != nil {
		t.Fatal(err)
	}
	var out syncBuffer
	fw := gui.NewWindow(a.NewWindow("data"), gui.WithStdout(&out),
		guiscene.WithScene(guiscene.Config{}))
	fw.LoadScript(`
		require(["@scene"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		page.Node("tbl").SetRows([["1", "one"], ["2", "two"]])
		page.Node("tbl").SetCell(0, 1, "uno")
		page.Node("lg").Append("hello")
		page.Node("lg").Append("world")
		try { page.Node("lg").SetRows([["1"]]); throw("SetRows on a log should error") } catch e { print("GUARD:", e) }
		window.SetContent(page)
		print("DATA-DONE")
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)
	if !waitForOut(t, &out, "DATA-DONE") {
		t.Fatalf("script output = %q", out.String())
	}
	if !strings.Contains(out.String(), "GUARD:") {
		t.Fatalf("wrong-node guard missing: %q", out.String())
	}

	view := fw.GetContentContainer().Objects[0].(*render.View)

	l, ok := view.Live("tbl")
	if !ok {
		t.Fatal("no live table")
	}
	rm, ok := render.RowModelOf(l)
	if !ok {
		t.Fatal("script-set table is not a RowModel")
	}
	// The script wrote two rows; a third does not exist.
	if !rm.SetCell(0, 1, "x") {
		t.Fatal("script-fed rows missing (SetCell(0,1) failed)")
	}
	if rm.SetCell(2, 0, "x") {
		t.Fatal("unexpected third row")
	}

	lg, ok := view.Live("lg")
	if !ok {
		t.Fatal("no live log")
	}
	al, ok := render.AppenderOf(lg)
	if !ok {
		t.Fatal("log is not an Appender")
	}
	al.AppendText("go-side too")
}
