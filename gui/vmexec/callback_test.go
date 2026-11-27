package vmexec

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// A callback registered outside a Run (run == nil) whose handler stops the
// context must not run the follow-up: FireThen re-checks ctx.Err() after the
// handler, exactly like the queued path (issue #66).
func TestFireThenNoRunSkipsAfterWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cb := &Callback{
		ctx:  ctx,
		name: "test",
		call: func(context.Context, *object.Closure, []object.Object) (object.Object, error) {
			cancel()
			return object.Nil, nil
		},
	}
	ran := false
	cb.FireThen(func(context.Context) { ran = true })
	if ran {
		t.Fatal("after ran although the handler cancelled the context")
	}

	// Sanity: with a live context the follow-up runs.
	cb2 := &Callback{
		ctx:  context.Background(),
		name: "test",
		call: func(context.Context, *object.Closure, []object.Object) (object.Object, error) {
			return object.Nil, nil
		},
	}
	cb2.FireThen(func(context.Context) { ran = true })
	if !ran {
		t.Fatal("after did not run with a live context")
	}
}
