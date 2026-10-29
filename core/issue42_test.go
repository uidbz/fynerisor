package core

import (
	"bytes"
	"strings"
	"testing"
)

// This file covers issue #42 (core half): a headless Context's print,
// eprint and io.* streams are its own, which is what makes
// `cat data.csv | fynerisor --headless -e '...' | jq` a clean filter.

func TestContextTextStreams(t *testing.T) {
	var out, errs bytes.Buffer
	cb := NewContext(WithIO(),
		WithStdin(strings.NewReader("x\ny\n")),
		WithStdout(&out), WithStderr(&errs))

	if _, err := cb.Eval(`io.stdout.write(io.stdin.read())`); err != nil {
		t.Fatalf("eval: %v", err)
	}
	if out.String() != "x\ny\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "x\ny\n")
	}

	if _, err := cb.Eval(`eprint("diag"); print("data")`); err != nil {
		t.Fatalf("eval: %v", err)
	}
	if errs.String() != "diag\n" {
		t.Errorf("stderr = %q, want %q", errs.String(), "diag\n")
	}
	if out.String() != "x\ny\ndata\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "x\ny\ndata\n")
	}

	// A second Context streams elsewhere; nothing leaks or is shared.
	var out2 bytes.Buffer
	cb2 := NewContext(WithIO(), WithStdin(strings.NewReader("other")), WithStdout(&out2))
	if _, err := cb2.Eval(`print(io.stdin.read())`); err != nil {
		t.Fatalf("eval: %v", err)
	}
	if out2.String() != "other\n" {
		t.Errorf("context B stdout = %q, want %q", out2.String(), "other\n")
	}
	if out.String() != "x\ny\ndata\n" {
		t.Errorf("context A stdout grew from context B's run: %q", out.String())
	}
}
