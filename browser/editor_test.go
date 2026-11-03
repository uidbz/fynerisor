package browser

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/gui"
)

// fakeEditor records which close path the browser took.
type fakeEditor struct {
	requested int
	closed    int
}

func (f *fakeEditor) CanEdit(*Page) bool                     { return true }
func (f *fakeEditor) Open(*Page, *Browser) fyne.CanvasObject { return nil }
func (f *fakeEditor) BeforeLeave(done func(bool))            { done(true) }
func (f *fakeEditor) RequestClose()                          { f.requested++ }
func (f *fakeEditor) EditorClosed()                          { f.closed++ }

// ToggleEdit must ask the editor to close (save prompt) rather than discarding
// its work through CloseEditor.
func TestToggleEditUsesRequestClose(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	defer w.Stop()

	b := New(w, Config{LoadPage: func(context.Context, string, string, string) (*Page, error) {
		return &Page{URL: "tie:/scenes/home", Kind: KindScene, Writable: true}, nil
	}})
	ed := &fakeEditor{}
	b.EnableEditor(ed)
	b.editing = true
	b.ToggleEdit()
	if ed.requested != 1 {
		t.Fatalf("RequestClose called %d times, want 1", ed.requested)
	}
}

// CloseEditor tells the provider the editor is gone, so it can stop autosave.
func TestCloseEditorNotifiesProvider(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	defer w.Stop()

	b := New(w, Config{LoadPage: func(context.Context, string, string, string) (*Page, error) {
		return &Page{URL: "tie:/scenes/home", Kind: KindScene, Writable: true}, nil
	}})
	ed := &fakeEditor{}
	b.EnableEditor(ed)
	b.editing = true
	b.CloseEditor(true)
	if ed.closed != 1 {
		t.Fatalf("EditorClosed called %d times, want 1", ed.closed)
	}
}

// tie: links resolve like file paths (sibling scenes), not like directory URLs.
func TestTieRelativeRefs(t *testing.T) {
	cases := []struct{ current, target, want string }{
		{"tie:/scenes/home", "kitchen", "tie:/scenes/kitchen"},
		{"tie:/scenes/home", "/scenes/other", "tie:/scenes/other"},
		{"tie:/scenes/home", "tie:/absolute", "tie:/absolute"},
		{"tie:/scenes/home", "tie:abcdef", "tie:abcdef"},
	}
	for _, c := range cases {
		if got := resolveRef(c.current, c.target); got != c.want {
			t.Errorf("resolveRef(%q, %q) = %q, want %q", c.current, c.target, got, c.want)
		}
	}
}
