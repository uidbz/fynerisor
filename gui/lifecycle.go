package gui

import (
	"context"
	"fmt"

	"github.com/uidbz/fynerisor/gui/vmexec"
)

// RunState is the lifecycle state of a window's script run (issue #39).
//
// Hosts should use State / Run / WithOnStateChange instead of parsing the
// "Ready!" / "ERROR: ..." status strings: those are human-readable text that
// scripts can also write with window.SetStatus.
type RunState int

const (
	// StateIdle: no script has been executed yet.
	StateIdle RunState = iota
	// StateRunning: the main body is executing.
	StateRunning
	// StateReady: the main body finished; widget events are being handled.
	StateReady
	// StateFailed: the main body raised an error (see State's error).
	StateFailed
	// StateStopped: the run was stopped with Stop or by its context.
	StateStopped
)

func (s RunState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateReady:
		return "ready"
	case StateFailed:
		return "failed"
	case StateStopped:
		return "stopped"
	}
	return fmt.Sprintf("RunState(%d)", int(s))
}

// WithOnStateChange registers a callback for lifecycle transitions of the
// current run: StateRunning when Execute/Start begins, then StateReady or
// StateFailed (with the error) when the main body ends, and StateStopped
// when the run is stopped. A run replaced by a newer Execute reports
// nothing further. The callback runs on the script's VM goroutine or the
// caller of Stop; it must not block.
func WithOnStateChange(fn func(state RunState, err error)) Option {
	return hookOption{fn: func(w *Window) { w.stateCallback = fn }}
}

// State returns the lifecycle state of the current run and, for
// StateFailed, the error that ended the main body.
func (w *Window) State() (RunState, error) {
	w.statusMutex.Lock()
	defer w.statusMutex.Unlock()
	return w.state, w.stateErr
}

// setState records a transition of run r; transitions of a run that is no
// longer current are ignored.
func (w *Window) setState(r *vmexec.Run, s RunState, err error) {
	w.statusMutex.Lock()
	if r != nil && w.run.Load() != r {
		w.statusMutex.Unlock()
		return
	}
	if s == StateStopped && w.state == StateStopped {
		w.statusMutex.Unlock()
		return // reported once, by Stop or the run's context
	}
	w.state, w.stateErr = s, err
	cb := w.stateCallback
	w.statusMutex.Unlock()
	if cb != nil {
		cb(s, err)
	}
}

// Start executes the loaded script like Execute, under ctx: cancelling ctx
// stops the run. The returned channel receives exactly one value when the
// main body has ended — nil on success, the script error on failure, or a
// cancellation error if the run was stopped or replaced first. Widget
// events keep being handled after that until the run is stopped.
func (w *Window) Start(ctx context.Context) <-chan error {
	return w.start(ctx)
}

// Run executes the loaded script and blocks until its main body has ended
// (see Start). If ctx ends first the run is stopped and ctx's error is
// returned. Headless hosts use it instead of polling the status string:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	if err := fw.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
func (w *Window) Run(ctx context.Context) error {
	done := w.Start(ctx)
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		w.Stop()
		return ctx.Err()
	}
}
