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
	"context"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/gui/internal/goid"
)

var (
	mainGoroutineID atomic.Uint64
	probeOnce       sync.Once
)

// SetMain records the current goroutine as the GUI/main goroutine,
// overwriting any earlier value. Hosts may call it once from main during
// setup; tests use it to make Do run inline. fynerisor itself never calls
// SetMain — it uses Init, which cannot be fooled by a window created on a
// background goroutine (issue #14).
func SetMain() {
	mainGoroutineID.Store(goid.ID())
}

// Init records the GUI goroutine without overwriting it.
//
// The first caller wins (normally main, via NewApp/NewWindow). When an app
// is running (appRunning), Init also schedules a one-time probe through
// fyne.Do: the probe runs on Fyne's real event-loop goroutine and records
// that goroutine, correcting the guess if the first caller was a background
// goroutine. Init is safe to call any number of times.
func Init(appRunning bool) {
	mainGoroutineID.CompareAndSwap(0, goid.ID())
	if !appRunning {
		return
	}
	probeOnce.Do(func() {
		fyne.Do(func() { mainGoroutineID.Store(goid.ID()) })
	})
}

// IsMain reports whether the caller is running on the GUI/main goroutine.
func IsMain() bool {
	return goid.ID() == mainGoroutineID.Load()
}

// Do runs fn on the GUI thread. If already on the GUI thread it runs inline
// (avoiding a re-queue that fyne.Do would cause); otherwise it dispatches via
// fyne.Do. If no GUI goroutine is known, IsMain is false and Do falls back
// to fyne.Do.
func Do(fn func()) {
	if IsMain() {
		fn()
		return
	}
	fyne.Do(fn)
}

// DoAndWait runs fn on the GUI thread and blocks until it has completed.
// Inline when already on the GUI thread, fyne.DoAndWait otherwise. Use this
// for reads that must return a value from Fyne state; Do is asynchronous
// when called off-main and the read would observe a stale value (issue #18).
//
// Script builtins should prefer DoAndWaitContext so a stopped script cannot
// stay blocked on a busy GUI thread.
func DoAndWait(fn func()) {
	if IsMain() {
		fn()
		return
	}
	fyne.DoAndWait(fn)
}

// DoAndWaitContext is DoAndWait that gives up when ctx is cancelled. It
// returns ctx.Err() in that case; fn may still run later on the GUI thread,
// so callers must not read values fn writes unless DoAndWaitContext
// returned nil.
func DoAndWaitContext(ctx context.Context, fn func()) error {
	if IsMain() {
		fn()
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	done := make(chan struct{})
	fyne.Do(func() {
		fn()
		close(done)
	})
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
