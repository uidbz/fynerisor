package scene_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// statusLog collects Config.Status messages; Status is called from the VM
// goroutine, so it must be goroutine-safe.
type statusLog struct {
	mu   sync.Mutex
	msgs []string
}

func (s *statusLog) add(m string) {
	s.mu.Lock()
	s.msgs = append(s.msgs, m)
	s.mu.Unlock()
}

func (s *statusLog) all() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.msgs...)
}

func waitFor(t *testing.T, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func runScene(t *testing.T, cfg guiscene.Config, stderr *syncBuffer, body string) *render.View {
	t.Helper()
	return runSceneDoc(t, sampleScene, cfg, stderr, body)
}

func runSceneDoc(t *testing.T, doc string, cfg guiscene.Config, stderr *syncBuffer, body string) *render.View {
	t.Helper()
	a := newApp(t)
	path := filepath.Join(t.TempDir(), "page.scene.json")
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	fw := gui.NewWindow(a.NewWindow("scene"), gui.WithStderr(stderr), guiscene.WithScene(cfg))
	fw.LoadScript(`
		require(["@scene"])
		let page = scene.Load("` + filepath.ToSlash(path) + `")
		` + body + `
		window.SetContent(page)
	`)
	fw.Run(nilCtx(t))
	waitReady(t, fw)
	return fw.GetContentContainer().Objects[0].(*render.View)
}

// With a status sink installed, a handler failure is still logged in full to
// the run's error sink (console builds keep their stderr log), and the
// one-line status gets only the first line of a multi-line error (#66).
func TestHandlerErrorLoggedAndFirstLineToStatus(t *testing.T) {
	var stderr syncBuffer
	var status statusLog
	view := runScene(t, guiscene.Config{Status: status.add}, &stderr,
		`page.Bind({hit: (evt) => { throw("first line\nsecond line") }})`)
	view.EmitEvent("btn", "tapped", nil)
	if !waitFor(t, func() bool { return strings.Contains(stderr.String(), "second line") }) {
		t.Fatalf("stderr = %q, want the full handler error", stderr.String())
	}
	if !strings.Contains(stderr.String(), "page.Bind.hit") {
		t.Fatalf("stderr = %q, want the callback name", stderr.String())
	}
	if !waitFor(t, func() bool { return len(status.all()) > 0 }) {
		t.Fatal("no status message")
	}
	for _, m := range status.all() {
		if strings.ContainsAny(m, "\r\n") {
			t.Fatalf("status message is multi-line: %q", m)
		}
	}
	if got := status.all()[0]; !strings.HasPrefix(got, "handler hit: ") || !strings.Contains(got, "first line") {
		t.Fatalf("status = %q", got)
	}
}

// Without Config.Status (every in-repo host but the browser), binding
// failures must still be visible: they go to the run's error sink (#66).
func TestBindingErrorWithoutStatusReachesStderr(t *testing.T) {
	var stderr syncBuffer
	doc := strings.Replace(sampleScene, `"text": "Before"`, `"text": "${x}"`, 1)
	runSceneDoc(t, doc, guiscene.Config{}, &stderr, `page.BindValues(() => 42)`)
	if !waitFor(t, func() bool { return strings.Contains(stderr.String(), "must return a map") }) {
		t.Fatalf("stderr = %q, want the BindValues error", stderr.String())
	}
}

// node.On listener failures reach Config.Status like page.Bind handlers (#66).
func TestListenerErrorVisibleInStatus(t *testing.T) {
	var stderr syncBuffer
	var status statusLog
	view := runScene(t, guiscene.Config{Status: status.add}, &stderr,
		`page.Node("btn").On("tapped", (evt) => { throw("boom") })`)
	view.EmitEvent("btn", "tapped", nil)
	ok := waitFor(t, func() bool {
		for _, m := range status.all() {
			if strings.HasPrefix(m, "listener btn.tapped: ") && strings.Contains(m, "boom") {
				return true
			}
		}
		return false
	})
	if !ok {
		t.Fatalf("status = %q, want listener btn.tapped: boom", status.all())
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Fatalf("stderr = %q, want the listener error logged too", stderr.String())
	}
}
