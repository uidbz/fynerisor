package scene_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/builtin"
)

// The shipped example must stay valid and use only documented components.
func TestExampleSceneValidates(t *testing.T) {
	path := filepath.Join("..", "examples", "44-scene-page", "page.scene.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, issues, err := scene.DecodeValidated(data, builtin.Registry())
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range issues {
		t.Errorf("%s", issue)
	}
	if scene.Errors(issues) != 0 {
		t.Fatal("example has validation errors")
	}
	if doc.Title == "" || doc.Count() < 5 {
		t.Fatalf("example looks empty: %+v", doc)
	}
	// Canonical encoding round-trips.
	enc, err := scene.Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scene.Decode(enc); err != nil {
		t.Fatalf("re-decode: %v", err)
	}
}

// The generated API reference must list every built-in component. Removing a
// spec without regenerating docs fails here.
func TestGeneratedDocsListComponents(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "API.md"))
	if err != nil {
		t.Skip("docs/API.md not available")
	}
	docs := string(data)
	for _, typ := range builtin.Registry().Types() {
		if !strings.Contains(docs, "| `"+typ+"` |") {
			t.Errorf("docs/API.md does not document scene component %q; run go generate", typ)
		}
	}
}
