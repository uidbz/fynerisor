package gui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"github.com/uidbz/fynerisor/gui/vmexec"
)

// Per-runtime text streams (issue #24).
//
// print, and the library's diagnostics, write through the window's
// Stdout()/Stderr() writers instead of the process-global os.Stdout.
// The old CaptureStdout hijacked os.Stdout with a pipe: it leaked into
// host output, leaked across CanvasObjects, and spawned goroutines per
// capture. A writer swap under a mutex does the same job with none of
// that: each Window (and each embedded CanvasObject) captures exactly
// its own script's output.

// Stdout returns the writer the script's print sends its output to.
func (w *Window) Stdout() io.Writer {
	w.stdoutMu.Lock()
	defer w.stdoutMu.Unlock()
	if w.stdoutOverride != nil {
		return w.stdoutOverride
	}
	if w.defaultOut != nil {
		return w.defaultOut
	}
	return os.Stdout
}

// Stderr returns the writer for script-execution errors and library
// diagnostics belonging to this window.
func (w *Window) Stderr() io.Writer {
	w.stdoutMu.Lock()
	defer w.stdoutMu.Unlock()
	if w.defaultErr != nil {
		return w.defaultErr
	}
	return os.Stderr
}

// SetStdout redirects this window's print output from Go code; nil
// restores the WithStdout default. Unlike the script-visible builtin,
// this is host configuration: it survives script reloads.
func (w *Window) SetStdout(sink io.Writer) {
	w.setStdout(sink, nil)
}

// SetStderr redirects this window's diagnostics from Go code; nil
// restores the WithStderr default.
func (w *Window) SetStderr(sink io.Writer) {
	w.stdoutMu.Lock()
	defer w.stdoutMu.Unlock()
	w.defaultErr = sink
}

// setStdout installs a redirect tagged with the run that asked for it,
// so resetRunState drops it when that run is superseded.
func (w *Window) setStdout(sink io.Writer, owner *vmexec.Run) {
	w.stdoutMu.Lock()
	defer w.stdoutMu.Unlock()
	w.stdoutOverride, w.stdoutOwner = sink, owner
}

// Report is the window's single error sink (issue #31): every runtime
// diagnostic — callback raises, panics, dropped callbacks, event
// backlog — goes through here, never through a bare fmt.Print. The
// WithErrorHandler handler, if set, receives the wrapped error; the
// default writes to this window's Stderr(). Library code must never
// write diagnostics to Stdout: with print and diagnostics sharing a
// stream, a log widget or a `fynerisor x.risor | jq` pipeline would
// mix script data with internal noise.
func (w *Window) Report(what string, err error) {
	if err == nil {
		return
	}
	w.stdoutMu.Lock()
	h := w.errorHandler
	w.stdoutMu.Unlock()
	if h != nil {
		h(fmt.Errorf("%s: %w", what, err))
		return
	}
	fmt.Fprintf(w.Stderr(), "fynerisor: %s: %v\n", what, err)
}

// callbackWriter bridges the runtime stdout stream to a script closure:
// each complete line is queued onto the closure's run as a discrete
// event (never coalesced or dropped, issue #51). print itself stays a
// plain Go write; the closure re-enters the VM only via the normal
// event queue, which is safe.
type callbackWriter struct {
	cb      *vmexec.Callback
	mu      sync.Mutex
	pending []byte
}

var _ io.Writer = (*callbackWriter)(nil)

func (c *callbackWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	c.pending = append(c.pending, p...)
	var lines []string
	for {
		i := bytes.IndexByte(c.pending, '\n')
		if i < 0 {
			break
		}
		lines = append(lines, string(c.pending[:i]))
		c.pending = c.pending[i+1:]
	}
	c.mu.Unlock()
	for _, line := range lines {
		c.cb.Fire(object.NewString(line))
	}
	return len(p), nil
}
