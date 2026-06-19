// Package guithread centralizes dispatching work onto the Fyne GUI thread.
//
// Fyne requires GUI mutations to happen on the main goroutine. fyne.Do always
// enqueues onto the GUI event queue and never runs inline — even when called
// from the main goroutine — so wrapping an already-on-main call (e.g. a widget
// method invoked from inside another fyne.Do callback or a button handler) just
// re-queues it for a later frame, where it can be dropped or run out of order.
//
// Do solves this: it runs the function inline when already on the GUI thread,
// and dispatches via fyne.Do otherwise. This makes widget methods safe to call
// from background goroutines AND from the main thread, so scripts no longer
// need to wrap GUI updates in window.Do.
package guithread

import (
	"runtime"
	"sync/atomic"

	"fyne.io/fyne/v2"
)

var mainGoroutineID atomic.Uint64

// SetMain records the current goroutine as the GUI/main goroutine. Call this
// once during app setup, from the main goroutine, before any Do calls.
func SetMain() {
	mainGoroutineID.Store(goroutineID())
}

// IsMain reports whether the caller is running on the GUI/main goroutine.
func IsMain() bool {
	return goroutineID() == mainGoroutineID.Load()
}

// Do runs fn on the GUI thread. If already on the GUI thread it runs inline
// (avoiding a re-queue that fyne.Do would cause); otherwise it dispatches via
// fyne.Do. If SetMain was never called, mainGoroutineID is 0 and IsMain is
// false, so Do falls back to fyne.Do — matching the previous behavior.
func Do(fn func()) {
	if IsMain() {
		fn()
		return
	}
	fyne.Do(fn)
}

// goroutineID parses the current goroutine's ID from its stack header. This is
// the same technique Fyne uses internally (its async package is internal and
// cannot be imported here).
func goroutineID() (id uint64) {
	var buf [30]byte
	runtime.Stack(buf[:], false)
	for i := 10; buf[i] != ' '; i++ {
		id = id*10 + uint64(buf[i]&15)
	}
	return id
}
