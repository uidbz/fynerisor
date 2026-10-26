package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeModule writes a Risor module file into dir and returns its path.
func writeModule(t *testing.T, dir, name, src string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatalf("write module %s: %v", name, err)
	}
	return p
}

// TestModuleFunctionReferencesModuleVariable verifies that an imported function
// can reference a module-level constant (the limitation this change lifts).
func TestModuleFunctionReferencesModuleVariable(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "math_utils.risor", `
let PI = 3.14159
let square = (x) => x * x
let circleArea = (r) => PI * square(r)
`)

	ctx := NewContext()
	res, err := ctx.Eval(`
let m = import("` + mod + `")
m.circleArea(2)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := res.(float64)
	if !ok {
		t.Fatalf("expected float64 result, got %T (%v)", res, res)
	}
	want := 3.14159 * 4
	if got != want {
		t.Fatalf("circleArea(2) = %v, want %v", got, want)
	}
}

// TestModuleFunctionReferencesModuleFunction verifies one exported function can
// call another module-level function.
func TestModuleFunctionReferencesModuleFunction(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "chain.risor", `
let double = (x) => x * 2
let quad = (x) => double(double(x))
`)

	ctx := NewContext()
	res, err := ctx.Eval(`import("` + mod + `").quad(3)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := res.(int64); !ok || got != 12 {
		t.Fatalf("quad(3) = %v (%T), want 12", res, res)
	}
}

// TestModulePlainValueExport verifies non-function exports are returned directly.
func TestModulePlainValueExport(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "consts.risor", `
let PI = 3.14159
let NAME = "fynerisor"
`)

	ctx := NewContext()
	res, err := ctx.Eval(`
let c = import("` + mod + `")
[c.PI, c.NAME]
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list, ok := res.([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("expected 2-element list, got %T (%v)", res, res)
	}
	if list[0].(float64) != 3.14159 || list[1].(string) != "fynerisor" {
		t.Fatalf("unexpected values: %v", list)
	}
}

// TestModuleFunctionUsesHostGlobal verifies that a module function can use a
// host-provided global module (here: strings), i.e. the module VM is seeded
// with the same environment as the main script.
func TestModuleFunctionUsesHostGlobal(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "shout.risor", `
let shout = (s) => strings.to_upper(s) + "!"
`)

	ctx := NewContext(WithStrings())
	res, err := ctx.Eval(`import("` + mod + `").shout("hi")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := res.(string); !ok || got != "HI!" {
		t.Fatalf("shout(\"hi\") = %v (%T), want HI!", res, res)
	}
}

// TestModuleCrossModuleComposition verifies a module can import another module
// and call its functions, with each module's globals resolving correctly.
func TestModuleCrossModuleComposition(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "base.risor", `
let BASE = 10
let bump = (x) => x + BASE
`)
	top := writeModule(t, dir, "top.risor", `
let base = import("`+filepath.Join(dir, "base.risor")+`")
let FACTOR = 3
let compute = (x) => base.bump(x) * FACTOR
`)

	ctx := NewContext()
	res, err := ctx.Eval(`import("` + top + `").compute(5)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (5 + 10) * 3 = 45
	if got, ok := res.(int64); !ok || got != 45 {
		t.Fatalf("compute(5) = %v (%T), want 45", res, res)
	}
}

// TestModuleCachingReturnsSameInstance verifies repeated imports of the same
// path share state.
func TestModuleCachingReturnsSameInstance(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "id.risor", `
let value = 42
let get = () => value
`)

	ctx := NewContext()
	res, err := ctx.Eval(`
let a = import("` + mod + `")
let b = import("` + mod + `")
a.get() + b.get()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := res.(int64); !ok || got != 84 {
		t.Fatalf("got %v (%T), want 84", res, res)
	}
}

// TestImportHTTPNon200Fails covers issue #19: importing a URL that answers
// with a non-200 status must fail loudly instead of importing an empty
// module.
func TestImportHTTPNon200Fails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := NewContext(WithHTTPImport())
	_, err := ctx.Eval(`import("` + srv.URL + `/missing.risor")`)
	if err == nil {
		t.Fatal("expected import of a 404 URL to fail")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected the error to mention HTTP 404, got: %v", err)
	}
}

// TestImportRelativeResolvesAgainstScriptDir covers issue #8: a relative
// import in a script loaded via LoadScriptFrom resolves against the
// script's own directory, not the process CWD.
func TestImportRelativeResolvesAgainstScriptDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "apps")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeModule(t, dir, "lib.risor", "let VAL = 2\n") // CWD copy — must NOT be picked
	writeModule(t, sub, "lib.risor", "let VAL = 1\n") // next to the script — must be picked
	top := filepath.Join(sub, "main.risor")

	cb := NewContext()
	cb.LoadScriptFrom(`import("lib.risor").VAL`, top)
	res, err := cb.Eval("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := res.(int64); !ok || got != 1 {
		t.Fatalf("VAL = %v (%T), want 1 (import resolved against wrong base?)", res, res)
	}
}

// TestImportCacheKeyCanonicalised covers issue #8: "./a.risor" and
// "a.risor" are the same module instance.
func TestImportCacheKeyCanonicalised(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "lib.risor", "let VAL = 5\n")

	cb := NewContext()
	cb.LoadScriptFrom(`let x = import("lib.risor")
let y = import("./lib.risor")
let z = import("`+filepath.ToSlash(dir)+`/lib.risor")
[x.VAL, y.VAL, z.VAL]`, filepath.Join(dir, "main.risor"))
	if _, err := cb.Eval(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cb.moduleCache) != 1 {
		t.Fatalf("expected 1 cached module, got %d (%v)", len(cb.moduleCache), keysOf(cb.moduleCache))
	}
}

func keysOf(m map[string]*ImportedModule) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestRemoteRelativeImportUsesRemoteBase covers issue #8: a script whose
// base is a URL resolves relative imports against that URL (same origin)
// and fetches them over HTTP — without needing the httpimport module.
func TestRemoteRelativeImportUsesRemoteBase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/lib.risor" {
			fmt.Fprint(w, "let VAL = 42\n")
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cb := NewContext()
	cb.LoadScriptFrom(`import("lib.risor").VAL`, srv.URL+"/app/index.risor")
	res, err := cb.Eval("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := res.(int64); !ok || got != 42 {
		t.Fatalf("VAL = %v (%T), want 42", res, res)
	}
}

// TestRemoteImportCrossOriginBlocked covers issue #8: a remote script may
// not import from another origin.
func TestRemoteImportCrossOriginBlocked(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	cb := NewContext(WithHTTPImport())
	cb.LoadScriptFrom(`import("http://cross-origin.invalid/x.risor")`, srv.URL+"/app/index.risor")
	_, err := cb.Eval("")
	if err == nil {
		t.Fatal("expected cross-origin import to fail")
	}
	if !strings.Contains(err.Error(), "cross-origin") {
		t.Fatalf("expected a cross-origin error, got: %v", err)
	}
}

// TestRemoteImportCannotReadLocalFiles covers issue #8: an absolute local
// path in a remote script never touches the local disk.
func TestRemoteImportCannotReadLocalFiles(t *testing.T) {
	dir := t.TempDir()
	secret := writeModule(t, dir, "secret.risor", "let SECRET = 999\n")

	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	cb := NewContext()
	cb.LoadScriptFrom(`import("`+filepath.ToSlash(secret)+`").SECRET`, srv.URL+"/app/index.risor")
	res, err := cb.Eval("")
	if err == nil {
		t.Fatalf("remote script read a local file: %v", res)
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected the import to fail as a remote 404, got: %v", err)
	}
}
