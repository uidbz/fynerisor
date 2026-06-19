package gui

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestSafeCallRecoversPanic(t *testing.T) {
	panicking := func(ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
		var s *object.String
		return s.RunOperation(0, nil) // nil deref, mirrors the real crash
	}

	result, err := safeCall(panicking, context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error from recovered panic, got nil")
	}
	if result != object.Nil {
		t.Fatalf("expected object.Nil result, got %v", result)
	}
}

func TestSafeCallPassesThrough(t *testing.T) {
	want := object.NewString("ok")
	good := func(ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
		return want, nil
	}

	result, err := safeCall(good, context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != want {
		t.Fatalf("expected pass-through result, got %v", result)
	}
}
