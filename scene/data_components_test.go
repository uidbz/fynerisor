package scene_test

import (
	"testing"

	"github.com/uidbz/fynerisor/scene"
	scenebuiltin "github.com/uidbz/fynerisor/scene/builtin"
)

// TestValidateDataComponents checks that the form/field/table/log specs are
// registered and that documents using them validate; a field with two
// children must be an error (MaxChildren).
func TestValidateDataComponents(t *testing.T) {
	reg := scenebuiltin.Registry()
	for _, typ := range []string{"form", "field", "table", "log"} {
		if _, ok := reg.Spec(typ); !ok {
			t.Fatalf("spec %q not registered", typ)
		}
	}

	doc := &scene.Document{Title: "data", Root: &scene.Node{
		ID: "root", Type: "free",
		Children: []*scene.Node{
			{ID: "f", Type: "form", Children: []*scene.Node{
				{ID: "f1", Type: "field", Props: map[string]any{"label": "Name"},
					Children: []*scene.Node{{ID: "e1", Type: "entry"}}},
				{ID: "f2", Type: "field", Children: []*scene.Node{
					{ID: "s1", Type: "select", Props: map[string]any{"options": []any{"a", "b"}}}}},
			}},
			{ID: "tbl", Type: "table", Props: map[string]any{"columns": []any{"A", "B"}}},
			{ID: "lg", Type: "log", Props: map[string]any{"text": "one\ntwo"}},
		},
	}}
	for _, iss := range scene.Validate(doc, reg) {
		if iss.Severity == scene.Error {
			t.Errorf("unexpected error: %+v", iss)
		}
	}

	bad := &scene.Document{Title: "bad", Root: &scene.Node{
		ID: "root", Type: "free",
		Children: []*scene.Node{
			{ID: "f", Type: "form", Children: []*scene.Node{
				{ID: "f1", Type: "field", Children: []*scene.Node{
					{ID: "a", Type: "entry"}, {ID: "b", Type: "entry"}}},
			}},
		},
	}}
	var found bool
	for _, iss := range scene.Validate(bad, reg) {
		if iss.Severity == scene.Error && iss.NodeID == "f1" {
			found = true
		}
	}
	if !found {
		t.Fatal("field with two children must be a validation error")
	}
}
