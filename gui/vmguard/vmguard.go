// Package vmguard provides a non-blocking, reentrant guard against concurrent
// Risor VM access from the GUI layer.
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
// Reentrancy: a single VM call legitimately nests on ONE goroutine — e.g. a
// button/window.Do handler runs through Call, calls Table.Refresh(), and that
// synchronously invokes the guarded Data/RowCount callbacks (also through Call)
// on the same goroutine. That is NOT a race. The guard tracks the owning
// goroutine and lets it re-enter; only a DIFFERENT goroutine is rejected.
//
// This guard does NOT make genuinely concurrent VM access safe — that is
// impossible without a real fix (keep background work in Go; see
// docs/CONCURRENCY.md). It converts silent corruption into an honest error at
// the moment a second goroutine races the owner, and recovers from any panic so
// a faulty script can never crash the GUI process.
package vmguard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// ErrConcurrentAccess is returned by Call when a DIFFERENT goroutine is already
// inside the VM. It is expected, transient contention (Fyne's render goroutine
// hitting a widget getter while the VM goroutine runs a handler), not a script
// bug — callers should recover by returning a last-known value rather than
// logging. Compare with errors.Is.
var ErrConcurrentAccess = errors.New("concurrent VM access detected: a go() worker is racing the GUI thread. " +
	"The Risor VM is single-threaded — move work out of go() or keep it in Go (see docs/CONCURRENCY.md)")

var (
	mu    sync.Mutex // guards owner/depth; held only for brief bookkeeping, never across the VM call
	owner uint64     // goroutine id currently inside the VM (0 = none)
	depth int        // reentrancy depth for owner
)

// Call runs a VM callback under the reentrant, non-blocking guard.
//
//   - If no goroutine is inside the VM, the caller becomes the owner and runs.
//   - If the SAME goroutine is already inside (legitimate nesting, e.g. a
//     handler calling Table.Refresh which calls back into Data), it re-enters.
//   - If a DIFFERENT goroutine is inside, this returns a diagnostic error
//     instead of entering and corrupting the shared stack.
//
// Any panic raised inside the VM is recovered and returned as an error: several
// entry points (notably the table render-thread getters in
// gui/widget/table.go) call this directly rather than through gui.safeCall, and
// a panic on Fyne's GL thread is otherwise fatal to the whole process.
func Call(callFunc object.CallFunc, ctx context.Context, fn *object.Closure, args []object.Object) (result object.Object, err error) {
	gid := goroutineID()

	mu.Lock()
	if owner == gid {
		depth++ // reentrant on the owning goroutine — allowed
	} else if owner == 0 {
		owner = gid
		depth = 1
	} else {
		mu.Unlock()
		return object.Nil, ErrConcurrentAccess
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		depth--
		if depth == 0 {
			owner = 0
		}
		mu.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "recovered panic in VM callback: %v\n%s\n", r, debug.Stack())
			result = object.Nil
			err = fmt.Errorf("panic in VM callback: %v", r)
		}
	}()

	return callFunc(ctx, fn, args)
}

// goroutineID parses the current goroutine's ID from its stack header — the same
// technique Fyne and gui/guithread use (the async helper is internal and cannot
// be imported here).
func goroutineID() (id uint64) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Header looks like: "goroutine 123 [running]:"
	const prefix = "goroutine "
	s := buf[len(prefix):n]
	for i := 0; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		id = id*10 + uint64(s[i]-'0')
	}
	return id
}
