package handlers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2"

	"github.com/uidbz/fynerisor/scene"
)

func TestHandlerName(t *testing.T) {
	tests := []struct{ node, event, want string }{
		{"submit-btn", "tapped", "on_submit_btn_tapped"},
		{"table", "cellTapped", "on_table_cell_tapped"},
		{"name-entry", "changed", "on_name_entry_changed"},
		{"a.b c", "hoverIn", "on_a_b_c_hover_in"},
	}
	for _, tc := range tests {
		if got := HandlerName(tc.node, tc.event); got != tc.want {
			t.Errorf("HandlerName(%q,%q) = %q, want %q", tc.node, tc.event, got, tc.want)
		}
	}
}

func TestConvertInlineBindings(t *testing.T) {
	src := `require(["v0.10", "@gui", "@scene"])

let count = 0

page.Bind({
    hello: (evt) => {
        count = count + 1
    },
    other: someFn
})

page.Node("title").SetText("hi")
`
	out, rep, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "let hello = (evt) => {") {
		t.Fatalf("standalone handler missing:\n%s", out)
	}
	if strings.Contains(out, "hello: (evt)") {
		t.Fatalf("inline handler still present:\n%s", out)
	}
	// The inline item becomes a shorthand in place; other items stay.
	if !strings.Contains(out, "page.Bind({\n    hello,\n    other: someFn\n})") {
		t.Fatalf("bindings section not rewritten:\n%s", out)
	}
	if len(rep.Converted) != 1 || rep.Converted[0] != "hello" {
		t.Fatalf("report = %+v", rep)
	}
}

func TestConvertAddsNames(t *testing.T) {
	src := `page.Bind({
    existing: someFn
})
`
	out, rep, err := ConvertInlineBindings(src, []string{"on_btn_tapped"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "page.Bind({\n    existing: someFn,\n    on_btn_tapped,\n})") {
		t.Fatalf("new binding missing:\n%s", out)
	}
	if len(rep.Bound) != 1 || rep.Bound[0] != "on_btn_tapped" {
		t.Fatalf("report = %+v", rep)
	}
}

func TestApplyCreatesHandlersAndBindings(t *testing.T) {
	doc := mustDoc(t, `{
  "format": "fynerisor.scene/1",
  "title": "Test",
  "root": {"id": "root", "type": "vbox", "children": [
    {"id": "btn", "type": "button", "props": {"text": "Go"},
     "on": {"tapped": [{"do": "call"}]}}
  ]}
}`)
	src := "page.Node(\"btn\").SetText(\"ready\")\n"

	res, err := Apply(doc, src)
	if err != nil {
		t.Fatal(err)
	}
	action := doc.Root.Children[0].On["tapped"][0]
	if action.Handler != "on_btn_tapped" {
		t.Fatalf("handler = %q, want on_btn_tapped", action.Handler)
	}
	for _, want := range []string{
		"let on_btn_tapped = (evt) => {",
		"page.Bind({\n    on_btn_tapped,\n})",
	} {
		if !strings.Contains(res.Source, want) {
			t.Fatalf("missing %q in:\n%s", want, res.Source)
		}
	}
	if len(res.Report.Created) != 1 {
		t.Fatalf("created = %v", res.Report.Created)
	}
}

func TestApplyKeepsExistingHandler(t *testing.T) {
	doc := mustDoc(t, `{
  "format": "fynerisor.scene/1",
  "root": {"id": "root", "type": "vbox", "children": [
    {"id": "btn", "type": "button",
     "on": {"tapped": [{"do": "call", "handler": "hello"}]}}
  ]}
}`)
	src := `let hello = (evt) => { print("hi") }` + "\n"
	res, err := Apply(doc, src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(res.Source, "let hello") != 1 {
		t.Fatalf("duplicated handler:\n%s", res.Source)
	}
	if !strings.Contains(res.Source, "page.Bind({\n    hello,\n})") {
		t.Fatalf("missing binding:\n%s", res.Source)
	}
}

func mustDoc(t *testing.T, src string) *scene.Document {
	t.Helper()
	doc, err := scene.Decode([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestUnicodeBeforeBind(t *testing.T) {
	src := "let title = \"café — ready\"\n\npage.Bind({\n    on_x_tapped,\n})\n"
	out, _, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "page.Bind({\n    on_x_tapped,\n})") {
		t.Fatalf("bindings corrupted by rune offset:\n%s", out)
	}
	if strings.Count(out, "on_x_tapped") != 1 {
		t.Fatalf("unexpected duplicates:\n%s", out)
	}
}

func TestUnicodeInlineHandler(t *testing.T) {
	src := "let title = \"café — ready\"\n\npage.Bind({\n    on_x_tapped: (evt) => {\n        print(\"x\")\n    },\n})\n"
	out, _, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "let on_x_tapped = (evt) => {\n    print(\"x\")\n}") {
		t.Fatalf("handler corrupted by rune offset:\n%s", out)
	}
}

func TestConventionCompiles(t *testing.T) {
	src := `let count = 0

let on_inc_tapped = (evt) => {
    count = count + 1
}

page.Bind({
    on_inc_tapped,
})
`
	if _, err := risor.Compile(context.Background(), src, risor.WithEnv(map[string]any{"page": map[string]any{}})); err != nil {
		t.Fatalf("convention does not compile: %v", err)
	}
}

func TestApplySidecar(t *testing.T) {
	dir := t.TempDir()
	doc := mustDoc(t, `{
  "format": "fynerisor.scene/1",
  "title": "Test",
  "script": "page.risor",
  "root": {"id": "root", "type": "vbox", "children": [
    {"id": "btn", "type": "button", "on": {"tapped": [{"do": "call"}]}}
  ]}
}`)
	script := filepath.Join(dir, "page.risor")
	if err := os.WriteFile(script, []byte("page.Bind({old: (evt) => {}})\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, changed, err := ApplySidecar(doc, filepath.Join(dir, "page.scene.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("document should have changed (handler name)")
	}
	if len(rep.Created) != 1 || rep.Created[0] != "on_btn_tapped" {
		t.Fatalf("report = %+v", rep)
	}
	got, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"let on_btn_tapped = (evt) => {",
		"let old = (evt) => {}",
		"page.Bind({old, on_btn_tapped})",
	} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestApplySidecarRemote(t *testing.T) {
	doc := &scene.Document{Script: "page.risor"}
	_, _, err := ApplySidecar(doc, "https://example.com/page.scene.json")
	if !errors.Is(err, ErrRemoteSidecar) {
		t.Fatalf("got %v, want ErrRemoteSidecar", err)
	}
}
