package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uidbz/fynerisor/gui/objattrs"
)

// Issue #47: the docs command renders the compiled registries; the CI
// drift step compares docs/API.md against `fynerisor docs`.

func TestDocsMarkdown(t *testing.T) {
	out := filepath.Join(t.TempDir(), "api.md")
	if code := runDocs([]string{"--markdown", "-o", out}); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)
	for _, want := range []string{
		"## widget.Label",
		"`SetText(text)`",
		"`Text` —", // property: bare name, no parentheses
		"## window",
		"`Resize(width, height)`",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
	if !strings.HasPrefix(md, "# fynerisor API reference") {
		t.Error("missing title")
	}
}

func TestDocsJSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "api.json")
	if code := runDocs([]string{"--json", "-o", out}); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var specs []objattrs.TypeSpecs
	if err := json.Unmarshal(b, &specs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(specs) < 70 {
		t.Errorf("got %d types, want the full wrapper surface", len(specs))
	}
	for i := 1; i < len(specs); i++ {
		if specs[i-1].TypeName > specs[i].TypeName {
			t.Fatalf("not sorted at %d: %s > %s", i, specs[i-1].TypeName, specs[i].TypeName)
		}
	}
}

func TestDocsCatalogCoversWithin(t *testing.T) {
	// Every canvas wrapper must carry at least the shared canvas trio;
	// Resize/Move/MinSize/Visible may be type-specific (PopUp defines its
	// own Move/Resize) and non-canvas wrappers get none (issue #38).
	notCanvas := map[string]bool{
		"widget.AccordionItem": true, "widget.FormItem": true,
		"widget.ToolbarAction": true, "widget.ToolbarSeparator": true,
		"widget.ToolbarSpacer": true, "widget.GenericCanvasObjectTable": true,
		"widget.ImageCellWrapper": true, "widget.TextStyle": true,
	}
	for _, ts := range objattrs.All() {
		if !strings.HasPrefix(ts.TypeName, "widget.") || ts.TypeName == "widget" || notCanvas[ts.TypeName] {
			continue
		}
		have := map[string]bool{}
		for _, a := range ts.Specs {
			have[a.Name] = true
		}
		for _, v := range []string{"Show", "Hide", "Refresh"} {
			if !have[v] {
				t.Errorf("%s missing shared verb %s", ts.TypeName, v)
			}
		}
	}
}
