package objattrs

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

type verbTestObj struct{}

func (v *verbTestObj) CanvasObject() fyne.CanvasObject { return nil }

// issue #58: a type that defines a verb (e.g. Table's bespoke Refresh) before
// calling Add must keep it, and Add must not panic on the duplicate.
func TestAddSkipsPredefinedVerbs(t *testing.T) {
	r := object.NewAttrRegistry[*verbTestObj]("verbTestObj")
	r.Define("Refresh").Doc("custom refresh").Args().Returns("nil").
		Impl(func(o *verbTestObj, _ context.Context, _ ...object.Object) (object.Object, error) {
			return object.Nil, nil
		})

	Add(r) // must not panic

	count, doc := 0, ""
	for _, spec := range r.Specs() {
		if spec.Name == "Refresh" {
			count++
			doc = spec.Doc
		}
	}
	if count != 1 {
		t.Fatalf("Refresh registered %d times, want 1", count)
	}
	if doc != "custom refresh" {
		t.Fatalf("custom Refresh was overwritten by the generic verb: %q", doc)
	}
	// the remaining verbs must still be registered
	for _, want := range []string{"Show", "Hide", "Resize", "Move", "MinSize", "Visible"} {
		found := false
		for _, spec := range r.Specs() {
			if spec.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Add skipped %q but should have registered it", want)
		}
	}
}
