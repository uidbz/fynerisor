package gui

import (
	"io"

	"github.com/deepnoodle-ai/risor/v2"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/core"
)

// Option is core.Option verbatim (issue #40): the very same option
// values configure a Window, a CanvasObject view, or a headless
// core.Context — module options from either package interoperate, and
// options apply in one ordered pass. gui-defined options (callbacks,
// host window) fill the same shared core.Config; Fyne-typed values are
// boxed in Config.HostWindow so core stays Fyne-free.
//
// Core options are re-exported below for discoverability; passing
// core.WithHTTP() here works identically.
type Option = core.Option

// windowHook lets gui-only options set Window state that has no place
// in the neutral core.Config (issue #40); core runtimes never see it.
type windowHook interface {
	applyToWindow(*Window)
}

type hookOption struct {
	fn func(*Window)
}

func (h hookOption) ApplyTo(*core.Config)    {}
func (h hookOption) applyToWindow(w *Window) { h.fn(w) }

// WithRisorOptions adds advanced Risor VM configuration options; see
// core.WithRisorOptions. Not forwarded to imported modules (issue #27).
func WithRisorOptions(globals ...risor.Option) Option {
	return core.WithRisorOptions(globals...)
}

// WithStatusCallback sets a callback that is called when script
// execution status changes ("Ready!", "ERROR: ...", ...).
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithStatusCallback(func(status string) {
//	        log.Println("Status:", status)
//	    }),
//	)
func WithStatusCallback(callback func(string)) Option {
	return core.NewOption(func(c *core.Config) { c.StatusCallback = callback })
}

// WithResultCallback sets a callback that is called when script
// execution completes with a non-nil result value.
func WithResultCallback(callback func(string)) Option {
	return core.NewOption(func(c *core.Config) { c.ResultCallback = callback })
}

// WithAppName sets the application name exposed to Risor scripts via
// app.name; see core.WithAppName.
func WithAppName(name string) Option {
	return core.WithAppName(name)
}

// WithHostWindow tells an embedded CanvasObject which Fyne window it
// lives in.
//
// Fyne has no way to discover the window containing a CanvasObject, so
// window-level script features (dialogs, shortcuts, main menu,
// window.Close, window.OnDropped) need the host window explicitly. The
// object is still rendered only inside its own CanvasObject;
// NewCanvasObject never replaces the host window's content.
//
// Example:
//
//	view := gui.NewCanvasObject(script, gui.WithHostWindow(host))
//
// WithHostWindow is ignored by NewWindow, which already receives its window.
func WithHostWindow(window fyne.Window) Option {
	return core.NewOption(func(c *core.Config) { c.HostWindow = window })
}

// WithStdout sets where the script's print output goes for this runtime
// (issue #24); see core.WithStdout. A *widget.Log is an io.Writer, so a
// log widget is a ready-made sink. Scripts change the same setting at
// runtime with window.SetStdout().
func WithStdout(out io.Writer) Option {
	return core.WithStdout(out)
}

// WithStdin sets where this window's io.stdin reads from (issue #42);
// see core.WithStdin.
func WithStdin(in io.Reader) Option {
	return core.WithStdin(in)
}

// WithStderr sets where this window reports script-execution errors and
// library diagnostics (issue #24); see core.WithStderr.
func WithStderr(err io.Writer) Option {
	return core.WithStderr(err)
}

// WithErrorHandler routes every runtime diagnostic of this window —
// callback errors, panics, dropped callbacks, event-backlog warnings —
// to fn instead of the WithStderr writer (issue #31); see
// core.Config.ErrorHandler.
func WithErrorHandler(fn func(error)) Option {
	return core.NewOption(func(c *core.Config) { c.ErrorHandler = fn })
}

// Module and capability options delegate to core — the same option
// objects work for gui and core runtimes (issue #40). SQL is opt-in via
// the gui/db subpackage so the database drivers are only linked when
// actually used.

func WithHTTP() Option                         { return core.WithHTTP() }
func WithHTTPImport() Option                   { return core.WithHTTPImport() }
func WithOS() Option                           { return core.WithOS() }
func WithScriptArgs(args ...string) Option     { return core.WithScriptArgs(args...) }
func WithAppVersion(version string) Option     { return core.WithAppVersion(version) }
func WithJSON() Option                         { return core.WithJSON() }
func WithCSV() Option                          { return core.WithCSV() }
func WithStrings() Option                      { return core.WithStrings() }
func WithFilepath() Option                     { return core.WithFilepath() }
func WithTime() Option                         { return core.WithTime() }
func WithTie() Option                          { return core.WithTie() }
func WithIO() Option                           { return core.WithIO() }
func WithExec() Option                         { return core.WithExec() }
func WithGlobal(name string, value any) Option { return core.WithGlobal(name, value) }

// Capability presets (issue #40): Safe, Files(readOnly), Exec, HTTP and
// Full express what a script may touch instead of enumerating modules.
//
// Example:
//
//	view := gui.NewCanvasObject(script, gui.Safe(), gui.Files(true))

func Safe() Option               { return core.Safe() }
func Files(readOnly bool) Option { return core.Files(readOnly) }

// WithSQL moved to the gui/db subpackage (issue #40): it forced the
// MySQL/PostgreSQL/SQLite(cgo)/SQL Server drivers into every binary
// importing gui. Migrate:
//
//	import "github.com/uidbz/fynerisor/gui/db"
//	window := gui.NewWindow(w, db.WithSQL())
