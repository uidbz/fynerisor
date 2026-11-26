package strings

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestTrimArities(t *testing.T) {
	// One argument is whitespace trim (like trim_space). The strict
	// 2-argument version used to abort whole GUI event handlers over a
	// bare trim(s), with the error invisible on Windows builds.
	got, err := trim(context.Background(), object.NewString("  hi  "))
	if err != nil {
		t.Fatal(err)
	}
	if got.(*object.String).Value() != "hi" {
		t.Fatalf("trim(s) = %q", got.(*object.String).Value())
	}

	got, err = trim(context.Background(), object.NewString("xxhixx"), object.NewString("x"))
	if err != nil {
		t.Fatal(err)
	}
	if got.(*object.String).Value() != "hi" {
		t.Fatalf("trim(s, cutset) = %q", got.(*object.String).Value())
	}

	if _, err := trim(context.Background()); err == nil {
		t.Fatal("trim() should error")
	}
	if _, err := trim(context.Background(), object.NewString("a"), object.NewString("b"), object.NewString("c")); err == nil {
		t.Fatal("trim(a,b,c) should error")
	}
}

func TestTrimSpace(t *testing.T) {
	got, err := trimSpace(context.Background(), object.NewString(" \t\n hi there \r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.(*object.String).Value() != "hi there" {
		t.Fatalf("trim_space(s) = %q", got.(*object.String).Value())
	}
	if _, err := trimSpace(context.Background()); err == nil {
		t.Fatal("trim_space() should error")
	}
	if _, err := trimSpace(context.Background(), object.NewString("a"), object.NewString("b")); err == nil {
		t.Fatal("trim_space(a, b) should error")
	}
	if _, err := trimSpace(context.Background(), object.NewInt(1)); err == nil {
		t.Fatal("trim_space(1) should error")
	}

	// Reachable through the module under its documented name.
	fn, ok := Module().GetAttr("trim_space")
	if !ok {
		t.Fatal("strings.trim_space not exported by Module()")
	}
	got, err = fn.(*object.Builtin).Call(context.Background(), object.NewString("  x  "))
	if err != nil || got.(*object.String).Value() != "x" {
		t.Fatalf("strings.trim_space via module = %v, %v", got, err)
	}
}
