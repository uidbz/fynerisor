package vmexec

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Callback is a script function registered by a builtin (a button's tap
// handler, a list's Length getter, a dialog's result handler, ...). It is
// the only way widget code should invoke script code (issue #12):
//
//	cb, err := vmexec.NewCallback(ctx, "widget.Button.OnTapped", args[0])
//	if err != nil {
//	    return nil, err
//	}
//	button.OnTapped = func() { cb.Fire() }            // event: async, ordered
//	list.Length = func() int { v, err := cb.Call() }  // getter: sync, may be ErrBusy
//
// The callback remembers the run whose VM registered it (via the builtin's
// context), so it always enters the right VM under that VM's guard, and it
// becomes a no-op once that run is stopped.
type Callback struct {
	run     *Run // nil when registered outside a Run (e.g. a bare risor.Eval)
	ctx     context.Context
	call    object.CallFunc
	fn      *object.Closure
	name    string
	onError func(error)
	refresh func()
}

// NewCallback validates fn and captures what is needed to call it later
// from any goroutine. name prefixes diagnostics ("widget.List.Length").
func NewCallback(ctx context.Context, name string, fn object.Object) (*Callback, error) {
	closure, ok := fn.(*object.Closure)
	if !ok {
		got := "nil"
		if fn != nil {
			got = string(fn.Type())
		}
		return nil, fmt.Errorf("%s: expected function, got %s", name, got)
	}
	call, ok := object.GetCallFunc(ctx)
	if !ok {
		return nil, fmt.Errorf("%s: unable to get call function", name)
	}
	return &Callback{run: FromContext(ctx), ctx: ctx, call: call, fn: closure, name: name}, nil
}

// Name returns the diagnostic name of the callback.
func (c *Callback) Name() string { return c.name }

// OnError replaces the default error reporting (stderr via the run's
// Report sink) for errors raised by the script function. It returns c.
func (c *Callback) OnError(fn func(error)) *Callback {
	c.onError = fn
	return c
}

// RefreshWhenBusy registers fn to run (on the VM goroutine, once the VM is
// idle) whenever Call hits ErrBusy, so a widget that rendered last-known
// values during contention is refreshed with fresh ones. fn should marshal
// to the UI thread itself (guithread.Do). It returns c.
func (c *Callback) RefreshWhenBusy(fn func()) *Callback {
	c.refresh = fn
	return c
}

// Fire invokes the script function asynchronously on the owning run's VM
// goroutine, after the main body and any earlier events. It never blocks
// and is safe from any goroutine, including Fyne's UI thread. Events of a
// stopped run are dropped.
func (c *Callback) Fire(args ...object.Object) {
	if c.run == nil {
		if c.ctx.Err() == nil {
			c.invoke(args)
		}
		return
	}
	// Enqueue reports ErrQueueFull itself; ErrStopped is silent.
	_ = c.run.Enqueue(c.name, func() { c.invoke(args) })
}

// Call invokes the script function synchronously and returns its result.
// Use it for getters that Fyne calls during layout/render. It returns
// ErrBusy when the VM is busy on another goroutine and ErrStopped when the
// owning run has been stopped; both are quiet (see Quiet) and the caller
// should fall back to its last-known value.
func (c *Callback) Call(args ...object.Object) (object.Object, error) {
	if c.run == nil {
		if err := c.ctx.Err(); err != nil {
			return object.Nil, ErrStopped
		}
		return safeCall(c.call, c.ctx, c.fn, args)
	}
	res, err := c.run.TryCall(func() (object.Object, error) {
		return c.call(c.ctx, c.fn, args)
	})
	if err == ErrBusy && c.refresh != nil {
		c.run.WhenIdle(c, c.refresh)
	}
	return res, err
}

// Report logs err with the callback's name, unless it is quiet (contention
// or a stopped run). Getters use it for non-quiet errors and bad results.
func (c *Callback) Report(err error) {
	if err == nil || Quiet(err) {
		return
	}
	if c.run != nil {
		c.run.Report(c.name, err)
		return
	}
	if c.ctx.Err() == nil {
		defaultReport(c.name, err)
	}
}

func (c *Callback) invoke(args []object.Object) {
	_, err := safeCall(c.call, c.ctx, c.fn, args)
	if err == nil || Quiet(err) || c.ctx.Err() != nil {
		return
	}
	if c.onError != nil {
		c.onError(err)
		return
	}
	c.Report(err)
}

// safeCall calls a closure, converting a panic into an error so a faulty
// script can never crash the GUI process.
func safeCall(call object.CallFunc, ctx context.Context, fn *object.Closure, args []object.Object) (result object.Object, err error) {
	defer func() {
		if p := recover(); p != nil {
			result = object.Nil
			err = fmt.Errorf("panic in script callback: %v\n%s", p, debug.Stack())
		}
	}()
	return call(ctx, fn, args)
}
