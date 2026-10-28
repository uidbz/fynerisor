package gui

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// stateLog records lifecycle transitions reported by WithOnStateChange.
type stateLog struct {
	mu     sync.Mutex
	states []RunState
	errs   []error
}

func (l *stateLog) option() Option {
	return WithOnStateChange(func(s RunState, err error) {
		l.mu.Lock()
		l.states = append(l.states, s)
		l.errs = append(l.errs, err)
		l.mu.Unlock()
	})
}

func (l *stateLog) snapshot() []RunState {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]RunState(nil), l.states...)
}

// TestRunReportsSuccess covers issue #39: Run blocks until the main body
// ends and reports success through the return value and State.
func TestRunReportsSuccess(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	log := &stateLog{}
	fw := NewWindow(a.NewWindow("Test"), log.option())
	defer fw.Stop()

	if s, _ := fw.State(); s != StateIdle {
		t.Fatalf("initial state %v, want idle", s)
	}
	fw.LoadScript(`window.SetContent(widget.NewLabel("hi"))`)
	if err := fw.Run(context.Background()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	if s, err := fw.State(); s != StateReady || err != nil {
		t.Fatalf("State = %v, %v; want ready, nil", s, err)
	}
	got := log.snapshot()
	if len(got) != 2 || got[0] != StateRunning || got[1] != StateReady {
		t.Fatalf("transitions %v, want [running ready]", got)
	}
}

// TestRunReportsFailureAndIgnoresSpoofedStatus covers issue #39: a failed
// body is reported as StateFailed with its error, while a script that merely
// writes an "ERROR: ..." status text stays StateReady.
func TestRunReportsFailureAndIgnoresSpoofedStatus(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	log := &stateLog{}
	fw := NewWindow(a.NewWindow("Test"), log.option())
	defer fw.Stop()

	fw.LoadScript(`window.SetStatus("ERROR: not really")`)
	if err := fw.Run(context.Background()); err != nil {
		t.Fatalf("spoofing script: Run returned %v", err)
	}
	if s, _ := fw.State(); s != StateReady {
		t.Fatalf("a script-set ERROR status changed the state to %v", s)
	}

	fw.LoadScript(`throw("real failure")`)
	err := fw.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "real failure") {
		t.Fatalf("Run returned %v, want the script error", err)
	}
	s, stateErr := fw.State()
	if s != StateFailed || stateErr == nil || !strings.Contains(stateErr.Error(), "real failure") {
		t.Fatalf("State = %v, %v; want failed with the script error", s, stateErr)
	}
	got := log.snapshot()
	if got[len(got)-1] != StateFailed {
		t.Fatalf("last transition %v, want failed (all: %v)", got[len(got)-1], got)
	}
}

// TestRunContextTimeoutStops covers issue #39: a Run whose context ends
// stops the script and returns the context error.
func TestRunContextTimeoutStops(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	fw := NewWindow(a.NewWindow("Test"), WithGlobal("block", blockingBuiltin(entered, release)))
	defer fw.Stop()

	fw.LoadScript(`block()`)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := fw.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run returned %v, want context.DeadlineExceeded", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("Run did not return promptly after its context ended")
	}
	if s, _ := fw.State(); s != StateStopped {
		t.Fatalf("State after timeout = %v, want stopped", s)
	}
}

// TestStopReportsStopped covers issue #39.
func TestStopReportsStopped(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	log := &stateLog{}
	fw := NewWindow(a.NewWindow("Test"), log.option())
	fw.LoadScript(`window.SetContent(widget.NewLabel("x"))`)
	if err := fw.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	fw.Stop()
	if s, _ := fw.State(); s != StateStopped {
		t.Fatalf("State after Stop = %v, want stopped", s)
	}
	fw.Stop() // idempotent: no second transition
	got := log.snapshot()
	if len(got) != 3 || got[2] != StateStopped {
		t.Fatalf("transitions %v, want [running ready stopped]", got)
	}
}
