package browser

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/gui"
)

// Scene pages are detected from LoadPage's kind or from the body format, and
// RenderScene is called instead of executing a script.
func TestRenderScenePage(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	defer w.Stop()

	rendered := make(chan *Page, 1)
	var mu sync.Mutex
	b := New(w, Config{
		LoadPage: func(_ context.Context, url, _, _ string) (*Page, error) {
			return &Page{URL: url, Kind: KindScene, Body: []byte(`{"format":"fynerisor.scene/1","root":{"id":"root","type":"free"}}`), Version: "v1", Writable: true}, nil
		},
		RenderScene: func(_ *Browser, p *Page) error {
			mu.Lock()
			rendered <- p
			mu.Unlock()
			return nil
		},
	})
	b.Navigate("https://example.com/page.scene.json")
	select {
	case p := <-rendered:
		if p.Kind != KindScene || !p.Writable {
			t.Fatalf("page = %+v", p)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("RenderScene was not called")
	}
	waitFor(t, time.Second, func() bool { return b.CurrentPage() != nil }, "current page")
	if b.CurrentPage().Version != "v1" {
		t.Fatalf("current page version = %q", b.CurrentPage().Version)
	}
}

// A scene body returned for a URL without a .scene.json suffix is still a
// scene page (format sniffing).
func TestSceneBodySniffing(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	defer w.Stop()

	seen := make(chan PageKind, 1)
	b := New(w, Config{
		LoadContent: func(context.Context, string, string, string) (string, error) {
			return `{"format":"fynerisor.scene/1","root":{"id":"root","type":"free"}}`, nil
		},
		RenderScene: func(_ *Browser, p *Page) error {
			seen <- p.Kind
			return nil
		},
	})
	b.Navigate("https://example.com/page")
	select {
	case kind := <-seen:
		if kind != KindScene {
			t.Fatalf("kind = %v, want scene", kind)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("scene was not rendered")
	}
}

func TestSceneEditorGuard(t *testing.T) {
	if !strings.HasSuffix("a.scene.json", ".scene.json") {
		t.Fatal("test setup")
	}
	if isSceneURL("https://x/a.scene.json?q=1") != true {
		t.Fatal("query string must not hide the scene suffix")
	}
	if isSceneURL("https://x/a.risor") {
		t.Fatal("script URL treated as scene")
	}
	if !looksLikeScene([]byte(`  {"format": "fynerisor.scene/1"}`)) {
		t.Fatal("scene body not detected")
	}
}
