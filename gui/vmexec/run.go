// Package vmexec serializes every entry into a script's Risor VM.
//
// A Risor VM is single-threaded. A fynerisor script is entered from its
// main body, from widget event handlers (taps, changes, dialogs, menus,
// shortcuts), from Fyne's render-time getters (List.Length, Table.Data,
// Tree.ChildUIDs, ...) and from go(). Before this package each entry point
// picked its own threading discipline; vmexec gives them one (issue #37):
//
//   - A Run owns one VM goroutine. It runs the main body and then, one at a
//     time, every queued event job (Enqueue / Callback.Fire). Event handlers
//     never run on Fyne's UI thread and never block it.
//   - Render getters must answer synchronously on the UI thread. They use
//     TryCall / Callback.Call: if the VM is free they run inline, if the VM
//     goroutine is busy they get ErrBusy immediately and fall back to their
//     last-known value. Callback.RefreshWhenBusy schedules a refresh once
//     the VM is idle again, so the fallback is never permanent.
//   - Mutual exclusion comes from a reentrant guard owned by the Run, i.e.
//     one per VM — never one per process (issue #13).
//   - Every Run has a context. It is cancelled by Stop, which halts the VM;
//     callbacks registered by a stopped run are dropped instead of entering
//     a dead VM (issue #16). Callbacks find the run that registered them
//     through the context Risor passes to builtins (FromContext).
//
// The UI thread only ever enqueues (non-blocking) or try-locks, and the VM
// goroutine may wait for the UI thread (guithread.DoAndWaitContext), so the
// two can never deadlock.
package vmexec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

var (
	// ErrBusy is returned by TryCall / Callback.Call when another goroutine
	// is inside the VM. It is expected, transient contention: callers keep
	// their last-known value and do not log it.
	ErrBusy = errors.New("script VM busy")

	// ErrStopped is returned when the run that owns a callback has been
	// stopped (Stop, or a new Execute replaced it). Callers drop the work
	// silently.
	ErrStopped = errors.New("script run stopped")

	// ErrQueueFull is returned by Enqueue when the event queue is full.
	ErrQueueFull = errors.New("callback queue full")
)

// DefaultQueueSize bounds the queue of pending event jobs. When it is full,
// events are dropped with a diagnostic instead of blocking the UI thread.
const DefaultQueueSize = 100

// Config configures a Run.
type Config struct {
	// QueueSize bounds pending event jobs; <= 0 means DefaultQueueSize.
	QueueSize int
	// Report receives callback errors and dropped events. nil writes a
	// one-line diagnostic to stderr.
	Report func(what string, err error)
}

type job struct {
	what string
	fn   func()
}

type runKey struct{}

// Run is one execution of a script: one VM, one VM goroutine, one guard,
// one event queue, one context.
type Run struct {
	ctx    context.Context
	cancel context.CancelFunc
	queue  chan job
	guard  *guard
	done   chan struct{}
	report func(what string, err error)

	idleMu sync.Mutex
	idle   map[any]func()
}

// warmRisor initializes Risor's lazily-created, unsynchronized package
// globals (the default type registry) exactly once, so that VMs of
// different runs — two windows or embedded views starting at the same
// time — only ever read them. core.EvalBatch does the same before fanning
// out.
var warmRisor sync.Once

// New creates a run whose context derives from parent. The run is attached
// to its context, so builtins invoked by the run's VM can find it with
// FromContext.
func New(parent context.Context, cfg Config) *Run {
	warmRisor.Do(func() { object.DefaultRegistry() })
	if parent == nil {
		parent = context.Background()
	}
	size := cfg.QueueSize
	if size <= 0 {
		size = DefaultQueueSize
	}
	r := &Run{
		queue:  make(chan job, size),
		guard:  newGuard(),
		done:   make(chan struct{}),
		report: cfg.Report,
		idle:   map[any]func(){},
	}
	if r.report == nil {
		r.report = defaultReport
	}
	r.ctx, r.cancel = context.WithCancel(context.WithValue(parent, runKey{}, r))
	return r
}

func defaultReport(what string, err error) {
	fmt.Fprintf(os.Stderr, "fynerisor: %s: %v\n", what, err)
}

// FromContext returns the run attached to ctx, or nil.
func FromContext(ctx context.Context) *Run {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(runKey{}).(*Run)
	return r
}

// Context returns the run's context. Pass it to risor.Eval so the VM, and
// every builtin it calls, carries the run.
func (r *Run) Context() context.Context { return r.ctx }

// Start runs body on the run's VM goroutine while holding the VM, then
// processes queued event jobs until the run is stopped. Events queued while
// the body runs wait for it to finish (they are not lost and do not block
// their sender).
func (r *Run) Start(body func(ctx context.Context)) {
	go func() {
		defer close(r.done)

		r.guard.enter()
		func() {
			defer r.guard.exit()
			defer r.recoverInto("script")
			body(r.ctx)
		}()
		r.flushIdle()

		for {
			select {
			case <-r.ctx.Done():
				return
			case j := <-r.queue:
				r.runJob(j)
			}
		}
	}()
}

func (r *Run) runJob(j job) {
	if r.ctx.Err() != nil {
		return
	}
	r.guard.enter()
	func() {
		defer r.guard.exit()
		defer r.recoverInto(j.what)
		j.fn()
	}()
	r.flushIdle()
}

func (r *Run) recoverInto(what string) {
	if p := recover(); p != nil {
		r.report(what, fmt.Errorf("panic: %v\n%s", p, debug.Stack()))
	}
}

// Stop cancels the run: the VM halts at its next instruction, queued and
// future events of this run are dropped.
func (r *Run) Stop() { r.cancel() }

// Stopped reports whether the run has been stopped.
func (r *Run) Stopped() bool { return r.ctx.Err() != nil }

// Done is closed when the VM goroutine has exited (after Stop).
func (r *Run) Done() <-chan struct{} { return r.done }

// InVM reports whether the calling goroutine is currently inside this run's
// VM (the main body, an event job, or a getter it triggered).
func (r *Run) InVM() bool { return r.guard.ownedByCaller() }

// Enqueue schedules fn to run on the VM goroutine after the current body or
// job. It never blocks: it returns ErrStopped for a stopped run and
// ErrQueueFull (after reporting it) when the queue is full.
func (r *Run) Enqueue(what string, fn func()) error {
	if r.ctx.Err() != nil {
		return ErrStopped
	}
	select {
	case r.queue <- job{what: what, fn: fn}:
		return nil
	default:
		r.report(what, fmt.Errorf("%w (%d); event dropped", ErrQueueFull, cap(r.queue)))
		return ErrQueueFull
	}
}

// TryCall runs fn synchronously inside the VM if the VM is free or already
// owned by the caller (reentrant). Otherwise it returns ErrBusy without
// waiting. Panics are recovered into errors.
func (r *Run) TryCall(fn func() (object.Object, error)) (result object.Object, err error) {
	if r.ctx.Err() != nil {
		return object.Nil, ErrStopped
	}
	if !r.guard.tryEnter() {
		return object.Nil, ErrBusy
	}
	defer r.guard.exit()
	defer func() {
		if p := recover(); p != nil {
			result = object.Nil
			err = fmt.Errorf("panic in script callback: %v\n%s", p, debug.Stack())
		}
	}()
	return fn()
}

// WhenIdle runs fn on the VM goroutine once the current body or job has
// finished. Calls with the same key are coalesced until then. Used to
// refresh widgets whose getters hit ErrBusy.
func (r *Run) WhenIdle(key any, fn func()) {
	r.idleMu.Lock()
	r.idle[key] = fn
	r.idleMu.Unlock()
}

func (r *Run) flushIdle() {
	r.idleMu.Lock()
	pending := r.idle
	r.idle = map[any]func(){}
	r.idleMu.Unlock()
	for _, fn := range pending {
		fn()
	}
}

// Report sends err to the run's error sink unless it is an expected,
// silent condition (ErrBusy, ErrStopped, context cancellation).
func (r *Run) Report(what string, err error) {
	if err == nil || Quiet(err) || r.ctx.Err() != nil {
		return
	}
	r.report(what, err)
}

// Quiet reports whether err is an expected condition that should not be
// logged: VM contention or a stopped run.
func Quiet(err error) bool {
	return errors.Is(err, ErrBusy) || errors.Is(err, ErrStopped) ||
		errors.Is(err, context.Canceled)
}
