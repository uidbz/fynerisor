package vmguard

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestCallSerializesAndDetectsContention verifies that while one goroutine is
// inside the guard, a second concurrent Call returns the diagnostic error
// rather than entering the VM.
func TestCallSerializesAndDetectsContention(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})

	// First call: blocks inside the guarded callFunc until released.
	slow := object.CallFunc(func(ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
		close(entered)
		<-release
		return object.Nil, nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = Call(slow, context.Background(), nil, nil)
	}()

	<-entered // guard is now held by the slow call

	// Second call while the guard is held must be rejected, not block.
	fast := object.CallFunc(func(ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
		t.Fatal("second Call entered the VM during contention — guard failed")
		return object.Nil, nil
	})
	_, err := Call(fast, context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "concurrent VM access") {
		t.Fatalf("expected concurrent-access diagnostic, got: %v", err)
	}

	close(release)
	wg.Wait()

	// After release, the guard is free again and Call works normally.
	ok := object.CallFunc(func(ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
		return object.NewString("ok"), nil
	})
	res, err := Call(ok, context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("post-release Call errored: %v", err)
	}
	if s, _ := res.(*object.String); s == nil || s.Value() != "ok" {
		t.Fatalf("post-release Call returned unexpected result: %v", res)
	}
}
