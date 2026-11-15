package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHandlersWritesBothFiles(t *testing.T) {
	dir := t.TempDir()
	scenePath := filepath.Join(dir, "page.scene.json")
	scriptPath := filepath.Join(dir, "page.risor")

	sceneSrc := `{
  "format": "fynerisor.scene/1",
  "title": "Demo",
  "script": "page.risor",
  "root": {"id": "root", "type": "vbox", "children": [
    {"id": "btn", "type": "button", "props": {"text": "Go"},
     "on": {"tapped": [{"do": "call"}]}}
  ]}
}`
	scriptSrc := "let count = 0\n\npage.Bind({\n    old: (evt) => { count = count + 1 }\n})\n"
	if err := os.WriteFile(scenePath, []byte(sceneSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte(scriptSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := runHandlers([]string{"-write", scenePath}); code != 0 {
		t.Fatalf("runHandlers = %d", code)
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"let on_btn_tapped = (evt) => {",
		"let old = (evt) => {",
		"page.Bind({\n    old,\n    on_btn_tapped,\n})",
	} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("missing %q in:\n%s", want, script)
		}
	}

	doc, err := os.ReadFile(scenePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), `"handler": "on_btn_tapped"`) {
		t.Fatalf("document not updated:\n%s", doc)
	}
}
