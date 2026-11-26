package strings

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestTrimArities(t *testing.T) {
	// One argument is whitespace trim (Risor stdlib parity). The strict
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
