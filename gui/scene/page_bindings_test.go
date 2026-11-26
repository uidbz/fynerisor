package scene_test

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/widget"

	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

func labelText(view *render.View, id string) string {
	obj, _ := view.LiveObject(id)
	if l, ok := obj.(*widget.Label); ok {
		return l.Text
	}
	return ""
}

// Binding expressions can call builtins and contain braces (#66).
func TestBindingExpressionsBuiltinsAndBraces(t *testing.T) {
	var stderr syncBuffer
	doc := strings.Replace(sampleScene, `"text": "Before"`, `"text": "${len(items)} / ${ ({v: items[0]}).v }"`, 1)
	view := runSceneDoc(t, doc, guiscene.Config{}, &stderr,
		`page.BindValues(() => ({items: ["a", "b", "c"]}))`)
	if got := labelText(view, "title"); got != "3 / a" {
		t.Fatalf("label = %q, want %q (stderr %q)", got, "3 / a", stderr.String())
	}
}

// Two ${} props on one node both resolve (#63: each RuntimeSet re-applied the
// other's raw template from the document).
func TestTwoBindingsOnOneNode(t *testing.T) {
	var stderr syncBuffer
	doc := strings.Replace(sampleScene, `"props": {"text": "Alice"}`, `"props": {"text": "${a}", "placeholder": "${b}"}`, 1)
	view := runSceneDoc(t, doc, guiscene.Config{}, &stderr,
		`page.BindValues(() => ({a: "A", b: "B"}))`)
	obj, _ := view.LiveObject("name")
	e := obj.(*widget.Entry)
	if e.Text != "A" || e.PlaceHolder != "B" {
		t.Fatalf("entry text=%q placeholder=%q, want A/B (stderr %q)", e.Text, e.PlaceHolder, stderr.String())
	}
}
