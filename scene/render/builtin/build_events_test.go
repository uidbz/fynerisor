package builtin

import (
	"context"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

type spyCaller struct{ got []string }

func (s *spyCaller) Call(h string, _ map[string]any) bool { s.got = append(s.got, h); return true }

func TestFormSubmitTextAtBuild(t *testing.T) {
	test.NewApp()
	doc := &scene.Document{Format: "fynerisor.scene/1", Title: "t", Root: &scene.Node{
		ID: "root", Type: "form",
		Props: map[string]any{"submit-text": "Save settings"},
		Children: []*scene.Node{{ID: "f1", Type: "field", Props: map[string]any{"label": "L"},
			Children: []*scene.Node{{ID: "e1", Type: "entry"}}}}}}
	v := render.NewView(doc, render.Options{Registry: NewRegistry(), Context: context.Background()})
	lo, ok := v.Live("root")
	if !ok {
		t.Fatal("no live for root")
	}
	f, ok := render.UnwrapLive(lo).Object().(*widget.Form)
	if !ok {
		t.Fatalf("root not a Form: %T", render.UnwrapLive(lo).Object())
	}
	if f.SubmitText != "Save settings" {
		t.Fatalf("BUG: SubmitText=%q at initial build (props never applied)", f.SubmitText)
	}
	if f.OnSubmit == nil {
		t.Fatal("BUG: OnSubmit not wired at initial build")
	}
}

func TestButtonTapRoutesToCaller(t *testing.T) {
	test.NewApp()
	doc := &scene.Document{Format: "fynerisor.scene/1", Title: "t", Root: &scene.Node{
		ID: "root", Type: "vbox",
		Children: []*scene.Node{{ID: "b1", Type: "button",
			Props: map[string]any{"text": "..."},
			On:    map[string][]scene.Action{"tapped": {{Do: "call", Handler: "doBrowseBase"}}}}}}}
	sc := &spyCaller{}
	v := render.NewView(doc, render.Options{Registry: NewRegistry(), Caller: sc, Context: context.Background()})
	lo, ok := v.Live("b1")
	if !ok {
		t.Fatal("no live for b1")
	}
	btn, ok := render.UnwrapLive(lo).Object().(*widget.Button)
	if !ok {
		t.Fatalf("b1 not a Button: %T", render.UnwrapLive(lo).Object())
	}
	if btn.OnTapped == nil {
		t.Fatal("BUG: OnTapped nil on scene button")
	}
	btn.OnTapped()
	if len(sc.got) != 1 || sc.got[0] != "doBrowseBase" {
		t.Fatalf("BUG: tap did not route to Caller, calls=%v", sc.got)
	}
}
