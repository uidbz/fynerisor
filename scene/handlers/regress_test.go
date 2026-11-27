package handlers

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/parser"

	"github.com/uidbz/fynerisor/scene"
)

// mustCompile fails the test when src is not a valid Risor program.
func mustCompile(t *testing.T, src string) {
	t.Helper()
	env := map[string]any{"page": map[string]any{}, "print": map[string]any{}, "require": map[string]any{}, "m": map[string]any{}, "someFn": map[string]any{}, "k": ""}
	if _, err := risor.Compile(context.Background(), src, risor.WithEnv(env)); err != nil {
		t.Fatalf("result does not compile: %v\n%s", err, src)
	}
}

func callDoc(t *testing.T, actions ...string) *scene.Document {
	t.Helper()
	// actions are "node/event/handler" triples; handler may be empty.
	var children []string
	for _, a := range actions {
		p := strings.SplitN(a, "/", 3)
		children = append(children, `{"id": "`+p[0]+`", "type": "button", "on": {"`+p[1]+`": [{"do": "call", "handler": "`+p[2]+`"}]}}`)
	}
	return mustDoc(t, `{"format": "fynerisor.scene/1", "root": {"id": "root", "type": "vbox", "children": [`+strings.Join(children, ",")+`]}}`)
}

// #60: the old example 44 (inline handler referenced by a call action) must
// keep its handler body.
func TestApplyKeepsInlineBodyExample44(t *testing.T) {
	src, err := os.ReadFile("testdata/page44.risor")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("testdata/page44.scene.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := scene.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(doc, string(src))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"let hello = (evt) => {\n    page.Node(\"title\").SetText(\"Hello from page.risor\")\n    page.Node(\"light\").Set(\"fill\", \"#22c55e\")\n    print(\"page.risor handler ran for \" + evt.id)\n}",
		"page.Bind({\n    hello\n})",
	} {
		if !strings.Contains(res.Source, want) {
			t.Fatalf("missing %q in:\n%s", want, res.Source)
		}
	}
	if strings.Contains(res.Source, "TODO") || len(res.Report.Created) != 0 {
		t.Fatalf("stub created for an inline handler: %+v\n%s", res.Report, res.Source)
	}
	if !slices.Equal(res.Report.Converted, []string{"hello"}) {
		t.Fatalf("converted = %v", res.Report.Converted)
	}
	mustCompile(t, res.Source)
}

// #60: an inline function whose name is already declared is kept, not dropped.
func TestApplyKeepsInlineWhenNameDeclared(t *testing.T) {
	src := "let save = (evt) => { print(\"old\") }\n\npage.Bind({\n    save: (evt) => { print(\"new code\") },\n})\n"
	doc := callDoc(t, "btn/tapped/save")
	res, err := Apply(doc, src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != src {
		t.Fatalf("source changed:\n%s", res.Source)
	}
	if len(res.Report.Warnings) != 1 || !strings.Contains(res.Report.Warnings[0], "save") {
		t.Fatalf("warnings = %v", res.Report.Warnings)
	}
}

// #60: two Bind calls that both define inline a keep both bodies.
func TestApplyTwoInlineSameName(t *testing.T) {
	src := "page.Bind({a: (evt) => { print(\"first\") }})\npage.Bind({a: (evt) => { print(\"second\") }})\n"
	res, err := Apply(callDoc(t, "btn/tapped/a"), src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"first", "second"} {
		if !strings.Contains(res.Source, want) {
			t.Fatalf("lost %q:\n%s", want, res.Source)
		}
	}
	if len(res.Report.Warnings) != 1 {
		t.Fatalf("warnings = %v", res.Report.Warnings)
	}
	mustCompile(t, res.Source)
}

// #61: keys that are not identifiers are left exactly as written.
func TestConvertNonIdentifierKeys(t *testing.T) {
	src := "page.Bind({\n    \"a-b\": (evt) => { print(1) },\n    `${k}`: someFn,\n    ...m,\n})\n"
	out, rep, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != src {
		t.Fatalf("source changed:\n%s", out)
	}
	if len(rep.Converted) != 0 {
		t.Fatalf("report = %+v", rep)
	}
}

// #61: a quoted identifier key is converted like a bare one.
func TestConvertQuotedIdentifierKey(t *testing.T) {
	src := "page.Bind({\n    \"hello\": (evt) => { print(1) },\n})\n"
	out, _, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "let hello = (evt) => { print(1) }") || !strings.Contains(out, "page.Bind({\n    hello,\n})") {
		t.Fatalf("not converted:\n%s", out)
	}
	mustCompile(t, out)
}

// #61: a sidecar that does not parse still gets handler names in the
// document; the script is left alone and a warning explains why.
func TestApplyParseErrorNamesHandlers(t *testing.T) {
	doc := callDoc(t, "btn/tapped/")
	src := "let on_btn_tapped = (evt) => {\n"
	res, err := Apply(doc, src)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Source != src {
		t.Fatalf("source changed:\n%s", res.Source)
	}
	if got := doc.Root.Children[0].On["tapped"][0].Handler; got != "on_btn_tapped" {
		t.Fatalf("handler = %q", got)
	}
	if len(res.Report.Warnings) != 1 {
		t.Fatalf("warnings = %v", res.Report.Warnings)
	}
}

// #61: a handler name that is not an identifier gets no stub and no binding.
func TestApplyInvalidHandlerName(t *testing.T) {
	src := "page.Bind({})\n"
	res, err := Apply(callDoc(t, "btn/tapped/my-handler"), src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != src {
		t.Fatalf("source changed:\n%s", res.Source)
	}
	if len(res.Report.Warnings) != 1 || !strings.Contains(res.Report.Warnings[0], "my-handler") {
		t.Fatalf("warnings = %v", res.Report.Warnings)
	}
	// A quoted key binds such a name; then there is nothing to warn about.
	src = "page.Bind({\"my-handler\": someFn})\n"
	res, err = Apply(callDoc(t, "btn/tapped/my-handler"), src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != src || len(res.Report.Warnings) != 0 {
		t.Fatalf("got %+v\n%s", res.Report, res.Source)
	}
}

// #62: new names are bound even when the first page.Bind is not a literal,
// and a comment mentioning page.Bind( does not count as a bindings section.
func TestApplyBindsAfterNonLiteralBind(t *testing.T) {
	src := "page.Bind(m)\npage.Bind({a: someFn})\n"
	res, err := Apply(callDoc(t, "btn/tapped/"), src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Source, "page.Bind({a: someFn, on_btn_tapped})") {
		t.Fatalf("not bound:\n%s", res.Source)
	}
	mustCompile(t, res.Source)

	src = "// call page.Bind(...) below\npage.Bind(m)\n"
	res, err = Apply(callDoc(t, "btn/tapped/"), src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Source, "page.Bind({\n    on_btn_tapped,\n})") {
		t.Fatalf("not bound:\n%s", res.Source)
	}
	mustCompile(t, res.Source)
}

// #62: comments inside the bindings map survive, and a script that needs no
// change is returned byte for byte.
func TestApplyPreservesComments(t *testing.T) {
	src := "let on_btn_tapped = (evt) => {}\n\npage.Bind({\n    // buttons\n    on_btn_tapped, // the button\n})\n"
	res, err := Apply(callDoc(t, "btn/tapped/"), src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != src {
		t.Fatalf("unchanged script rewritten:\n%s", res.Source)
	}
	res, err = Apply(callDoc(t, "btn/tapped/", "go/tapped/"), src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// buttons", "on_btn_tapped, // the button\n    on_go_tapped,\n})"} {
		if !strings.Contains(res.Source, want) {
			t.Fatalf("missing %q:\n%s", want, res.Source)
		}
	}
	mustCompile(t, res.Source)
}

// #62: function, const and multi-name let declarations count as declared.
func TestApplyDeclarationForms(t *testing.T) {
	for _, decl := range []string{
		"function on_btn_tapped(evt) { print(1) }",
		"const on_btn_tapped = (evt) => { print(1) }",
		"let a, on_btn_tapped = [1, (evt) => { print(1) }]",
	} {
		src := decl + "\n\npage.Bind({\n    on_btn_tapped,\n})\n"
		res, err := Apply(callDoc(t, "btn/tapped/"), src)
		if err != nil {
			t.Fatal(err)
		}
		if res.Source != src || len(res.Report.Created) != 0 {
			t.Fatalf("%s: duplicated declaration:\n%s", decl, res.Source)
		}
	}
}

// #62: generated names that collide get a suffix and a warning.
func TestApplyHandlerNameCollision(t *testing.T) {
	doc := callDoc(t, "table-cell/tapped/", "table/cellTapped/")
	res, err := Apply(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	a := doc.Root.Children[0].On["tapped"][0].Handler
	b := doc.Root.Children[1].On["cellTapped"][0].Handler
	if a != "on_table_cell_tapped" || b != "on_table_cell_tapped_2" {
		t.Fatalf("handlers = %q, %q", a, b)
	}
	if len(res.Report.Warnings) != 1 || len(res.Report.Created) != 2 {
		t.Fatalf("report = %+v", res.Report)
	}
	mustCompile(t, res.Source)
}

func TestHandlerNameAcronym(t *testing.T) {
	if got := HandlerName("req", "HTTPDone"); got != "on_req_http_done" {
		t.Fatalf("got %q", got)
	}
	if got := HandlerName("x", "a/b"); !IsIdent(got) {
		t.Fatalf("not an identifier: %q", got)
	}
}

// #62: lines inside a multi-line raw string keep their indentation.
func TestConvertKeepsRawString(t *testing.T) {
	src := "page.Bind({\n    hello: (evt) => {\n        let s = `line1\n    line2`\n        print(s)\n    },\n})\n"
	out, _, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "let hello = (evt) => {\n    let s = `line1\n    line2`\n    print(s)\n}"
	if !strings.Contains(out, want) {
		t.Fatalf("missing %q in:\n%s", want, out)
	}
}

// #62: CRLF files stay CRLF.
func TestApplyCRLF(t *testing.T) {
	src := "let x = 1\r\n\r\npage.Bind({\r\n    hello: (evt) => {\r\n        print(x)\r\n    },\r\n})\r\n"
	res, err := Apply(callDoc(t, "btn/tapped/", "go/tapped/hello"), src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(res.Source, "\n") != strings.Count(res.Source, "\r\n") {
		t.Fatalf("mixed line endings: %q", res.Source)
	}
	mustCompile(t, res.Source)
}

// #62: an unchanged parse is reused; the result of Apply parses.
func TestApplyResultParses(t *testing.T) {
	res, err := Apply(callDoc(t, "btn/tapped/"), "page.Bind({})\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse(context.Background(), res.Source, nil); err != nil {
		t.Fatalf("%v\n%s", err, res.Source)
	}
	if !strings.Contains(res.Source, "page.Bind({\n    on_btn_tapped,\n})") {
		t.Fatalf("not bound:\n%s", res.Source)
	}
}

// #61: a parse error in the sidecar does not block saving.
func TestApplySidecarParseError(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "page.risor")
	broken := "let on_btn_tapped = (evt) => {\n"
	if err := os.WriteFile(script, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := callDoc(t, "btn/tapped/")
	doc.Script = "page.risor"
	rep, changed, err := ApplySidecar(doc, filepath.Join(dir, "page.scene.json"))
	if err != nil {
		t.Fatalf("ApplySidecar: %v", err)
	}
	if !changed || len(rep.Warnings) != 1 {
		t.Fatalf("changed=%v report=%+v", changed, rep)
	}
	got, _ := os.ReadFile(script)
	if string(got) != broken {
		t.Fatalf("script rewritten:\n%s", got)
	}
}

// #62: the sidecar is replaced atomically and keeps its mode.
func TestApplySidecarKeepsMode(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "page.risor")
	if err := os.WriteFile(script, []byte("page.Bind({})\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := callDoc(t, "btn/tapped/")
	doc.Script = "page.risor"
	if _, _, err := ApplySidecar(doc, filepath.Join(dir, "page.scene.json")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(script)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v", fi.Mode().Perm())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary files left: %v", entries)
	}
}

// #62: without a sidecar the document still gets handler names, so it can
// pass validation.
func TestApplySidecarMissingNamesHandlers(t *testing.T) {
	dir := t.TempDir()
	doc := callDoc(t, "btn/tapped/")
	doc.Script = "page.risor"
	_, changed, err := ApplySidecar(doc, filepath.Join(dir, "page.scene.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
	if !changed || doc.Root.Children[0].On["tapped"][0].Handler != "on_btn_tapped" {
		t.Fatalf("changed=%v handler=%q", changed, doc.Root.Children[0].On["tapped"][0].Handler)
	}
	doc = callDoc(t, "btn/tapped/")
	if _, changed, err := ApplySidecar(doc, filepath.Join(dir, "page.scene.json")); err != nil || !changed {
		t.Fatalf("no script: changed=%v err=%v", changed, err)
	}
}

// #60: the exported two-step pass (EnsureHandlers then ConvertInlineBindings)
// must not stub, and then drop, an inline handler either.
func TestEnsureThenConvertKeepsInlineBody(t *testing.T) {
	src := "page.Bind({\n    hello: (evt) => {\n        print(\"body\")\n    },\n})\n"
	out, names, rep, err := EnsureHandlers(callDoc(t, "btn/tapped/hello", "go/tapped/"), src)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(rep.Created, []string{"on_go_tapped"}) {
		t.Fatalf("created = %v", rep.Created)
	}
	out, _, err = ConvertInlineBindings(out, names)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"let hello = (evt) => {\n    print(\"body\")\n}", "page.Bind({\n    hello,\n    on_go_tapped,\n})"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	mustCompile(t, out)
}

// Inline functions in a page.Bind nested in a function may capture locals,
// so they stay where they are; a page.Bind inside a let is lifted above it.
func TestConvertScopes(t *testing.T) {
	src := "let setup = () => {\n    let n = 1\n    page.Bind({a: (evt) => { print(n) }})\n}\n"
	out, rep, err := ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != src || len(rep.Converted) != 0 {
		t.Fatalf("nested bind rewritten:\n%s", out)
	}
	src = "let r = page.Bind({\n    a: (evt) => { print(1) },\n})\n"
	out, _, err = ConvertInlineBindings(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "let a = (evt) => { print(1) }\n\nlet r = page.Bind({\n    a,\n})") {
		t.Fatalf("not lifted above the statement:\n%s", out)
	}
	mustCompile(t, out)
}

// An empty bindings map with the brace on its own line gets one name per line.
func TestApplyEmptyMultilineBind(t *testing.T) {
	res, err := Apply(callDoc(t, "btn/tapped/"), "page.Bind({\n})\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(res.Source, "page.Bind({\n    on_btn_tapped,\n})\n") {
		t.Fatalf("got:\n%s", res.Source)
	}
	mustCompile(t, res.Source)
}
