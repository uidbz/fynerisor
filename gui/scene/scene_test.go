package scene_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

const sampleScene = `{
  "format": "fynerisor.scene/1",
  "title": "Bound",
  "root": {
    "id": "root", "type": "free",
    "place": {"x": 0, "y": 0, "w": 800, "h": 600},
    "props": {"designW": 800.0, "designH": 600.0, "scale": "none"},
    "children": [
      {"id": "title", "type": "label", "place": {"x": 10, "y": 20, "w": 200, "h": 40},
       "props": {"text": "Before"}},
      {"id": "door", "type": "hotspot", "place": {"x": 100, "y": 100, "w": 80, "h": 60}},
      {"id": "btn", "type": "button", "place": {"x": 300, "y": 100, "w": 140, "h": 40},
       "props": {"text": "Go"},
       "on": {"tapped": [{"do": "call", "handler": "hit"}]}},
      {"id": "name", "type": "entry", "place": {"x": 10, "y": 200, "w": 200, "h": 40},
       "props": {"text": "Alice"}}
    ]
  }
}`

func newApp(t *testing.T) fyne.App {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	return a
}

func waitReady(t *testing.T, fw *gui.Window) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		st, err := fw.State()
		switch st {
		case gui.StateReady:
			return
		case gui.StateFailed:
			t.Fatalf("script failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	st, err := fw.State()
	t.Fatalf("script did not become ready: %s %v\nstatus: %s", st, err, fw.Status())
}

func writeScene(t *testing.T) (dir, path string) {
	t.Helper()
	dir = t.TempDir()
	path = filepath.Join(dir, "page.scene.json")
	if err := os.WriteFile(path, []byte(sampleScene), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

func TestSceneLoadPageAndNode(t *testing.T) {
	a := newApp(t)
	_, path := writeScene(t)
	var out bytes.Buffer
	fw := gui.NewWindow(a.NewWindow("scene"), gui.WithStdout(&out),
		guiscene.WithScene(guiscene.Config{}))
	fw.LoadScript(`
		require(["@scene"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		if (page.Title != "Bound") { throw("bad title") }
		page.Node("title").SetText("After")
		page.Node("door").SetPlace({x: 5, y: 6})
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)

	content := fw.GetContentContainer()
	if len(content.Objects) != 1 {
		t.Fatalf("content objects = %d", len(content.Objects))
	}
	view, ok := content.Objects[0].(*render.View)
	if !ok {
		t.Fatalf("content is %T, want *render.View", content.Objects[0])
	}
	label, ok := view.LiveObject("title")
	if !ok {
		t.Fatal("no label live object")
	}
	if l, ok := label.(*widget.Label); !ok || l.Text != "After" {
		t.Fatalf("label = %#v", label)
	}
	view.RefreshRects()
	r, _ := view.RectOf("door")
	if r.X != 5 || r.Y != 6 {
		t.Fatalf("runtime place = %+v", r)
	}
}

func TestBindAndDeclarativeCall(t *testing.T) {
	a := newApp(t)
	_, path := writeScene(t)
	var out syncBuffer
	fw := gui.NewWindow(a.NewWindow("scene"), gui.WithStdout(&out),
		guiscene.WithScene(guiscene.Config{}))
	fw.LoadScript(`
		require(["@scene"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		page.Bind({hit: (evt) => { print("HIT:" + evt.id) }})
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)

	view := fw.GetContentContainer().Objects[0].(*render.View)
	view.EmitEvent("btn", "tapped", nil)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "HIT:btn") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("handler output = %q", out.String())
}

func TestValuesAndListeners(t *testing.T) {
	a := newApp(t)
	_, path := writeScene(t)
	var out syncBuffer
	fw := gui.NewWindow(a.NewWindow("scene"), gui.WithStdout(&out),
		guiscene.WithScene(guiscene.Config{}))
	fw.LoadScript(`
		require(["@scene"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		page.Node("btn").On("tapped", (evt) => {
			let v = page.Values()
			print("VAL:" + v.name)
		})
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)

	view := fw.GetContentContainer().Objects[0].(*render.View)
	entry, _ := view.LiveObject("name")
	entry.(*widget.Entry).SetText("Bob")
	view.EmitEvent("btn", "tapped", nil)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "VAL:Bob") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("values output = %q", out.String())
}

func TestSceneParseAndDoc(t *testing.T) {
	a := newApp(t)
	fw := gui.NewWindow(a.NewWindow("scene"), guiscene.WithScene(guiscene.Config{}))
	fw.LoadScript(`
		require(["@scene"])
		let doc = scene.Parse(` + "`" + sampleScene + "`" + `)
		let issues = doc.Validate()
		if (len(issues) != 0) { throw("unexpected validation issues: " + issues[0]) }
		if (len(doc.Encode()) < 10) { throw("encode failed") }
		let page = doc.Page()
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func nilCtx(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}
