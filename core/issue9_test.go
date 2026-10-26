package core

import (
	"context"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestEvalContextCancellation covers issue #9: EvalContext must pass the
// caller's context to the Risor VM so cancellation interrupts the run.
func TestEvalContextCancellation(t *testing.T) {
	slow := object.NewBuiltin("slow", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		select {
		case <-ctx.Done():
			return nil, object.Errorf("slow: context cancelled")
		case <-time.After(1 * time.Second):
			return object.NewString("timeout"), nil
		}
	})

	cb := NewContext(WithGlobal("slow", slow))

	ctx, cancel := context.WithCancel(context.Background())
	type evalResult struct {
		res any
		err error
	}
	done := make(chan evalResult, 1)
	go func() {
		res, err := cb.EvalContext(ctx, `slow()`)
		done <- evalResult{res, err}
	}()

	time.Sleep(100 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("script completed before the context was cancelled")
	default:
	}

	cancel()

	select {
	case r := <-done:
		// With the fix the cancelled context aborts the run; the script must
		// not complete normally.
		if r.err == nil {
			t.Fatalf("script completed normally after cancel (result %v)", r.res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("EvalContext still running 2s after context cancellation")
	}
}
