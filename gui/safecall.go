package gui

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/fynerisor/gui/vmguard"
)

// safeCall invokes a script callback, recovering from any panic so that a
// faulty script (e.g. a nil operation or type error in a callback) surfaces as
// an error instead of crashing the entire GUI process. Callbacks run on the
// Fyne GUI thread or in goroutines, where an unrecovered panic is fatal.
func safeCall(callFunc object.CallFunc, ctx context.Context, fn *object.Closure, args []object.Object) (result object.Object, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "recovered panic in script callback: %v\n%s\n", r, debug.Stack())
			result = object.Nil
			err = fmt.Errorf("panic in script callback: %v", r)
		}
	}()
	return vmguard.Call(callFunc, ctx, fn, args)
}
