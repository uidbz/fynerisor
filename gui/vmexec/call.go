package vmexec

import (
	"context"
	"errors"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Call invokes a script closure from a host dispatch path with the VM's
// call function installed in its context.
//
// A host that receives a script function (an event bridge's on(handler), a
// command palette callback, ...) captures the VM call function with
// object.GetCallFunc when the registration builtin runs. Passing that call
// function and the closure to Call later — from a window.Do job or a Go
// goroutine — makes closures created by the handler work: without the call
// function the VM cannot invoke them, and widget constructors that take a
// callback (widget.NewButton, widget.OnSubmitted, ...) fail with "unable to
// get call function".
//
// Call installs call into ctx, so the caller never needs
// object.WithCallFunc. When ctx carries a run (a builtin context, or the
// Run's Context), Call enters that run's reentrant guard: it runs inline
// when the VM is free or already owned by the calling goroutine and returns
// ErrBusy otherwise. Call recovers panics into errors, so a faulty handler
// cannot crash the GUI process.
//
// Call and Run.Call run the closure on the calling goroutine (like TryCall,
// for render-time getters). Schedule them with Window.Do, or use
// Callback.Fire, so script code never runs on Fyne's UI thread.
//
// Call is the low-level host dispatch helper. Capture the Run at registration
// time and use Run.Call when possible: it rejects closures from stopped runs
// and always runs with that run's context. Host code that only needs a queued
// handler can pass the closure to NewCallback instead, which captures the
// same call function and provides Fire/Call.
func Call(call object.CallFunc, ctx context.Context, fn *object.Closure, args ...object.Object) (object.Object, error) {
	if fn == nil {
		return nil, errors.New("vmexec.Call: nil function")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if call == nil {
		existing, ok := object.GetCallFunc(ctx)
		if !ok {
			return nil, errors.New("vmexec.Call: no VM call function")
		}
		call = existing
	}
	if r := FromContext(ctx); r != nil {
		return r.Call(call, fn, args...)
	}
	return safeCall(call, object.WithCallFunc(ctx, call), fn, args)
}

// Call invokes fn with args on the run's VM context, installing the VM call
// function and entering the run's reentrant guard. It runs on the calling
// goroutine; call it from a Window.Do job (never Fyne's UI thread).
//
// Hosts should prefer it over the package-level Call when they have the Run
// captured at registration: it binds the invocation to the run, so a closure
// from a stopped or replaced run returns ErrStopped instead of entering a
// different VM, and UI writes from the handler are correctly ignored for a
// stale run. It returns ErrBusy when another goroutine is inside the VM.
func (r *Run) Call(call object.CallFunc, fn *object.Closure, args ...object.Object) (object.Object, error) {
	if fn == nil {
		return nil, errors.New("vmexec.Run.Call: nil function")
	}
	if call == nil {
		return nil, errors.New("vmexec.Run.Call: no VM call function")
	}
	if r.Stopped() {
		return object.Nil, ErrStopped
	}
	if !r.guard.tryEnter() {
		return object.Nil, ErrBusy
	}
	defer r.guard.exit()
	return safeCall(call, object.WithCallFunc(r.ctx, call), fn, args)
}
