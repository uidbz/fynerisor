package core

import (
	"fmt"
	"strings"
	"testing"
)

// This file covers issue #40: one option mechanism (shared Config,
// single ordered pass) plus capability presets.

func TestOptionSinglePassOrder(t *testing.T) {
	// Later options see/override earlier ones in application order.
	cb := NewContext(
		WithGlobal("x", 1),
		WithGlobal("x", 2),
		WithAppName("first"),
		WithAppName("second"),
	)
	got, err := cb.Eval(`[x, app.name]`)
	if err != nil {
		t.Fatal(err)
	}
	if s := fmt.Sprint(got); !strings.Contains(s, "2") || !strings.Contains(s, "second") {
		t.Errorf("eval = %s, want x=2 and appName=second", s)
	}
}

func TestPresetsCapabilities(t *testing.T) {
	// Safe: data modules only — no os, no exec, no file access.
	cb := NewContext(Safe())
	if _, err := cb.Eval(`os.goos()`); err == nil || !strings.Contains(err.Error(), "os") {
		t.Errorf("Safe: os.goos() err = %v, want undefined os", err)
	}
	if r, err := cb.Eval(`csv.parse("a,b\n1,2")[0].a`); err != nil {
		t.Errorf("Safe: csv failed: %v", err)
	} else if fmt.Sprint(r) != "1" {
		t.Errorf("Safe: csv value = %v, want 1", r)
	}

	// Files(true): the mutating operations do not exist at all.
	ro := NewContext(Files(true))
	dir := t.TempDir()
	p := dir + "/x.txt"
	if _, err := ro.Eval(fmt.Sprintf("io.write(%q, \"data\")", p)); err == nil ||
		!strings.Contains(err.Error(), "write") {
		t.Errorf("Files(true): io.write err = %v, want attribute error", err)
	}
	if _, err := ro.Eval(fmt.Sprintf("os.mkdir_all(%q)", dir+"/deep")); err == nil ||
		!strings.Contains(err.Error(), "mkdir_all") {
		t.Errorf("Files(true): os.mkdir_all err = %v, want attribute error", err)
	}
	// Reads still work.
	if _, err := ro.Eval(fmt.Sprintf("io.read(%q)", p)); err == nil ||
		!strings.Contains(err.Error(), "no such file") {
		t.Errorf("Files(true): io.read of missing file err = %v, want I/O raise", err)
	}

	// Files(false) + Exec(): the writable/exec capabilities come back.
	rw := NewContext(Files(false), Exec())
	if _, err := rw.Eval(fmt.Sprintf("io.write(%q, \"data\")", p)); err != nil {
		t.Errorf("Files(false): io.write: %v", err)
	}
	r, err := rw.Eval(`exec(["echo", "hi"]).stdout`)
	if err != nil {
		t.Errorf("Exec(): %v", err)
	} else if strings.TrimSpace(fmt.Sprint(r)) != "hi" {
		t.Errorf("Exec(): got %q", fmt.Sprint(r))
	}
}

func TestFullPresetRequire(t *testing.T) {
	cb := NewContext(Full())
	if _, err := cb.Eval(`require(["v0.1", "@http", "@io", "@os", "@exec", "@csv", "@json", "@strings", "@filepath", "@time", "@tie"])`); err != nil {
		t.Errorf("Full: require: %v", err)
	}
	// SQL is deliberately not part of Full: it links drivers (opt-in subpackage).
	if _, err := cb.Eval(`require(["@sql"])`); err == nil {
		t.Error("Full: @sql should not be enabled")
	}
}
