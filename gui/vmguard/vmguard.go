// Package vmguard provides a non-blocking guard against concurrent Risor VM
// access from the GUI layer.
//
// The Risor VM is single-threaded and not thread-safe: a single window owns one
// VM with no internal lock. Two goroutines running VM bytecode at once corrupt
// the shared stack/frame pointer, which surfaces as misleading errors such as
// "function takes 1 argument (2 given)" or panics like "unknown opcode: 0".
//
// The usual trigger is a go(() => {...}) worker that touches the VM (e.g. an
// .each()/.append() or a widget setter) while Fyne synchronously calls a widget
// getter (Table.Data, RowCount, List.Length, ...) on the GUI thread.
//
// This guard does NOT make concurrent VM access safe — that is impossible
// without a real fix (keep background work in Go; see docs/CONCURRENCY.md). It
// uses a non-blocking TryLock so it never freezes the GUI behind an in-flight
// go() worker; instead it converts the silent corruption into an honest,
// actionable error at the moment of the race.
package vmguard

import (
	"context"
	"fmt"
	"sync"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

var mu sync.Mutex

// Call runs a VM callback under the non-blocking guard. If another goroutine is
// already inside the VM, it returns a diagnostic error instead of entering and
// corrupting the shared stack.
func Call(callFunc object.CallFunc, ctx context.Context, fn *object.Closure, args []object.Object) (object.Object, error) {
	if !mu.TryLock() {
		return object.Nil, fmt.Errorf("concurrent VM access detected: a go() worker is racing the GUI thread. " +
			"The Risor VM is single-threaded — move work out of go() or keep it in Go (see docs/CONCURRENCY.md)")
	}
	defer mu.Unlock()
	return callFunc(ctx, fn, args)
}
