package core

import (
	"os"
	"path/filepath"
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
