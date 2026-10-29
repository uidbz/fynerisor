package gui

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	_ "embed"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/core"
	iomod "github.com/uidbz/fynerisor/modules/io"
	osmod "github.com/uidbz/fynerisor/modules/os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"
	risorwidget "github.com/uidbz/fynerisor/gui/widget"
)

const WindowType object.Type = "window"

// NewApp is a convenience function that creates a new Fyne application and
// fynerisor Window in one call. It's equivalent to calling app.New(),
// app.NewWindow(), and NewWindow() but more concise.
//
// This is the recommended way to create a fynerisor application for simple
// use cases where you don't need direct access to the Fyne app object.
//
// Parameters:
//   - title: The window title
//   - opts: Optional configuration using functional options (WithGlobals, WithStatusCallback, etc.)
//
// Returns:
//   - *Window: The fynerisor window ready for script execution
//
// Example:
//
//	fw := fynerisor.NewApp("My Application",
//	    fynerisor.WithHTTP(),
//	    guidb.WithSQL(), // import "github.com/uidbz/fynerisor/gui/db"
//	)
//	fw.LoadScript(`
//	    require(["v0.2", "@sql"])
//	    let btn = widget.NewButton("Hello", () => {
//	        window.SetStatus("Clicked!")
//	    })
//	    window.SetContent(btn)
//	`)
//	fw.Execute()
//	fw.ShowAndRun()
//
// For more control over the Fyne app, use NewWindow instead:
//
//	a := app.New()
//	a.Settings().SetTheme(theme.DarkTheme())
//	w := a.NewWindow("My App")
//	fw := fynerisor.NewWindow(w, fynerisor.WithHTTP())
func NewApp(title string, opts ...Option) *Window {
	guithread.Init(false) // NewApp is normally called from main
	a := app.New()
	w := a.NewWindow(title)
	fw := NewWindow(w, opts...)
	fw.FyneApp = a
	return fw
}

// Window wraps a fyne.Window and provides Risor scripting capabilities.
// It exposes global objects (window, widget, container, canvas, chart) to Risor scripts
// and manages script execution in a separate goroutine with callback support.
//
// The Window implements object.Object to be accessible from Risor scripts.
// Scripts can call window.SetContent() to update the UI, window.OnDropped() to handle
// file drops, and window.SetStdout() to route printed output (issue #24).
type Window struct {
	FyneWindow fyne.Window
	// FyneApp is the owning Fyne app (set by NewApp; nil when a host
	// supplies its own fyne.Window via NewWindow). Enables app.Quit().
	FyneApp fyne.App

	content        *fyne.Container
	status         string // guarded by statusMutex; see Status()/SetStatus (issue #15)
	statusMutex    sync.Mutex
	state          RunState              // guarded by statusMutex; see State() (issue #39)
	stateErr       error                 // guarded by statusMutex
	stateCallback  func(RunState, error) // WithOnStateChange
	globals        []risor.Option
	env            map[string]any  // Merged built-in globals (shared with imported module VMs)
	userGlobals    []risor.Option  // Opaque user-supplied globals (via WithGlobals)
	enabledModules map[string]bool // Track which modules are enabled
	appName        string          // Name of the embedding application
	appVersion     string          // require()/app.version version; default: deprecated global (issue #44)
	scriptArgs     []string        // this window's os.args(), carried in the run ctx (issue #44)
	droppedPaths   []string
	cbMu           sync.Mutex // guards onDropped and dropOwner
	dropOwner      *vmexec.Run
	onDropped      func(droppedPaths []string)
	run            atomic.Pointer[vmexec.Run] // current script run: VM goroutine, guard, event queue, context (issue #37)

	// Per-runtime text streams (issue #24): print writes to Stdout(),
	// library diagnostics to Stderr(). No os.Stdout hijack, no pipes:
	// two CanvasObjects capture their own output, and a host's
	// fmt.Println is untouched.
	stdoutMu       sync.Mutex // guards the six fields below
	defaultIn      io.Reader  // WithStdin; nil means os.Stdin
	defaultOut     io.Writer  // WithStdout; nil means os.Stdout
	defaultErr     io.Writer  // WithStderr; nil means os.Stderr
	stdoutOverride io.Writer  // script SetStdout/CaptureStdout redirect
	stdoutOwner    *vmexec.Run
	errorHandler   func(error) // WithErrorHandler (issue #31)

	contentMu      sync.Mutex // serializes SetContent swaps between overlapping runs
	statusCallback func(string)
	resultCallback func(string)
	runner         *core.ScriptRunner

	// Module import system: the module cache and import stack live in the
	// per-run importState (see import.go); only the entry script's base is
	// kept here, captured by the next Execute.
	// Keyboard shortcut system (entries tagged with the run that added them)
	shortcuts     map[string]shortcutEntry
	shortcutMutex sync.Mutex
	menuOwner     atomic.Pointer[vmexec.Run] // run that installed the main menu, if a script did

	startMu sync.Mutex   // serializes start (stop previous run + publish the new one)
	envMu   sync.RWMutex // guards env and enabledModules after construction (RegisterGlobal)
}

// shortcutEntry is a keyboard shortcut installed by a script run.
type shortcutEntry struct {
	shortcut fyne.Shortcut
	owner    *vmexec.Run
}

// NewWindow creates a new fynerisor Window that wraps a Fyne window.
//
// Parameters:
//   - window: The fyne.Window to wrap
//   - opts: Optional configuration using functional options (WithGlobals, WithStatusCallback, etc.)
//
// The window automatically provides these global objects to Risor scripts:
//   - window: This Window instance
//   - widget: Widget factory for creating UI components
//   - container: Container factory for layouts (VBox, HBox, Border, etc.)
//   - canvas: Canvas object factory (lines, images)
//   - chart: Chart factory (bar charts)
//
// Example:
//
//	a := app.New()
//	w := a.NewWindow("My App")
//	fyneWindow := fynerisor.NewWindow(w,
//	    fynerisor.WithStatusCallback(func(status string) {
//	        log.Println("Status:", status)
//	    }),
//	)
//	fyneWindow.LoadScript("window.SetContent(widget.NewLabel('Hello'))")
//	fyneWindow.Execute()
//	fyneWindow.ShowAndRun()
func NewWindow(window fyne.Window, opts ...Option) *Window {
	guithread.Init(window != nil)
	w := newWindow(opts...)
	w.FyneWindow = window
	if window != nil {
		window.SetContent(w.content)
		w.installDropHandler()
	}
	return w
}

// newWindow builds a Window runtime without attaching its content to a Fyne
// window. NewWindow uses it for standalone windows; NewCanvasObject uses it to
// host a script inside a single canvas object in a host application.
// windowCreated, when set (tests only), is called for every new Window so a
// test can stop and drain the script runs it created before the next test
// starts.
var windowCreated func(*Window)

// Quiesce stops the current run and waits up to timeout for its VM
// goroutine to exit, i.e. until no more script code or UI work of that run
// can happen. It reports whether the run exited in time.
func (w *Window) Quiesce(timeout time.Duration) bool {
	r := w.run.Load()
	if r == nil {
		return true
	}
	w.Stop()
	select {
	case <-r.Done():
		return true
	case <-time.After(timeout):
		return false
	}
}

func newWindow(opts ...Option) *Window {
	// Record the GUI goroutine without overwriting it: a host may create an
	// embedded view on a background goroutine (issue #14).
	guithread.Init(false)

	// One ordered config pass (issue #40): options assemble a shared
	// core.Config, the Window is then built from it. No concrete-type
	// switches, so no option is ever silently ignored.
	cfg := core.Apply(opts...)
	w := &Window{
		enabledModules: cfg.EnabledModules,
		appName:        cfg.AppName,
		appVersion:     cfg.AppVersion,
		scriptArgs:     cfg.ScriptArgs,
		shortcuts:      make(map[string]shortcutEntry),
		defaultIn:      cfg.Stdin, // nil accessors fall back to process streams
		defaultOut:     cfg.Stdout,
		defaultErr:     cfg.Stderr,
		statusCallback: cfg.StatusCallback,
		resultCallback: cfg.ResultCallback,
		errorHandler:   cfg.ErrorHandler,
	}
	if w.appVersion == "" {
		w.appVersion = core.GetAppVersion()
	}
	if hw, ok := cfg.HostWindow.(fyne.Window); ok && hw != nil {
		w.FyneWindow = hw // WithHostWindow (embedded CanvasObject)
	}
	// gui-only option state (callbacks core.Config cannot express).
	for _, opt := range opts {
		if h, ok := opt.(windowHook); ok {
			h.applyToWindow(w)
		}
	}

	// Set base globals
	fynewidget := NewWidget(w)
	fynecanvas := &Canvas{}
	fynecontainer := &Container{}
	chart := &Chart{}
	app := newAppObject(w)
	constants := newConstantsObject()
	fyneobj := &Fyne{w: w}
	bindingobj := NewBinding()
	dialogobj := NewDialog(w)

	globals := map[string]any{
		"window":    w,
		"widget":    fynewidget,
		"canvas":    fynecanvas,
		"container": fynecontainer,
		"chart":     chart,
		"dialog":    dialogobj,
		"app":       app,
		"clipboard": NewFyneClipboard(),
		"constants": constants,
		"fyne":      fyneobj,
		"binding":   bindingobj,
		"print":     w.newPrintBuiltin(),
		"eprint":    w.newEprintBuiltin(),
		"require":   newRequireBuiltin(w),
		"import":    w.newImportBuiltin(),
		"go":        newGoBuiltin(),
	}

	// Build a single merged environment containing the standard library and all
	// built-in globals. This same map is passed to imported module VMs so that
	// functions defined in modules can access globals like widget, http, etc.
	w.env = risor.Builtins()
	for k, v := range globals {
		w.env[k] = v
	}

	// Globals contributed by options (modules, WithGlobal).
	for k, v := range cfg.Globals {
		w.env[k] = v
	}
	w.userGlobals = cfg.UserVMOptions

	// Compose the final risor options: the merged env first, then any opaque
	// user-supplied globals (from WithRisorOptions).
	w.globals = append([]risor.Option{risor.WithEnv(w.env)}, w.userGlobals...)

	// Create script runner
	w.runner = core.NewScriptRunner(w.globals)

	// Initialize content container. A standalone NewWindow attaches it to the
	// Fyne window; an embedded CanvasObject renders it as a widget.
	w.content = container.NewStack()

	if hook := windowCreated; hook != nil {
		hook(w)
	}
	return w
}

// installDropHandler connects Fyne's window-level drop events to the
// window.OnDropped script binding. Only meaningful when a host window exists.
func (w *Window) installDropHandler() {
	if w.FyneWindow == nil {
		return
	}
	w.FyneWindow.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		paths := make([]string, len(uris))
		for i, uri := range uris {
			paths[i] = uri.Path()
		}
		w.droppedPaths = paths
		w.cbMu.Lock()
		onDropped := w.onDropped
		w.cbMu.Unlock()
		if onDropped != nil {
			onDropped(paths) // queues the script callback on its run
		}
	})
}

// newPrintBuiltin implements print(): output goes to the window's stdout
// writer (issue #24), never to a process-global os.Stdout. Each runtime
// captures its own print output; hosts keep theirs.
func (w *Window) newPrintBuiltin() *object.Builtin {
	return object.NewBuiltin("print", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		values := make([]any, len(args))
		for i, arg := range args {
			values[i] = object.PrintableValue(arg)
		}
		fmt.Fprintln(w.Stdout(), values...)
		return object.Nil, nil
	})
}

// newEprintBuiltin implements eprint(): the same text stream contract
// as print, but on the window's stderr writer — a script can be a
// clean pipeline filter, diagnostics on the side channel (issue #42).
func (w *Window) newEprintBuiltin() *object.Builtin {
	return object.NewBuiltin("eprint", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		values := make([]any, len(args))
		for i, arg := range args {
			values[i] = object.PrintableValue(arg)
		}
		fmt.Fprintln(w.Stderr(), values...)
		return object.Nil, nil
	})
}

// newGoBuiltin implements go(fn): fn runs on the script's VM goroutine after
// the current main body or event, like any other queued event (issue #17).
//
// It is deliberately not a goroutine. A Risor VM is single-threaded, so a
// closure running in parallel with the body or another callback corrupts
// the VM (docs/CONCURRENCY.md). Deferring fn still keeps the GUI responsive:
// script code never runs on Fyne's UI thread, so while fn works the window
// keeps repainting and accepting input (events wait their turn, render
// getters use last-known values). Parallel work belongs in Go — a host
// global or module that does the work and reports back through a callback.
func newGoBuiltin() *object.Builtin {
	return object.NewBuiltin("go", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("go: expected 1 argument (function), got %d", len(args))
		}
		cb, err := vmexec.NewCallback(ctx, "go()", args[0])
		if err != nil {
			return nil, err
		}
		cb.Fire()
		return object.Nil, nil
	})
}

// SetStatus updates the window status and calls the status callback if provided.
func (w *Window) SetStatus(status string) {
	w.statusMutex.Lock()
	w.status = status
	w.statusMutex.Unlock()
	if w.statusCallback != nil {
		w.statusCallback(status)
	}
}

// Status returns the most recent status string. The status is written by
// the Execute goroutine (SetStatus) and read from host goroutines, so both
// sides go through statusMutex (issue #15).
func (w *Window) Status() string {
	w.statusMutex.Lock()
	defer w.statusMutex.Unlock()
	return w.status
}

// LoadScript loads a Risor script that will be executed when Execute() is
// called. Relative import() paths in the script resolve against the process
// working directory; use LoadScriptFrom when the script's own location is
// known (issue #8).
func (w *Window) LoadScript(script string) {
	w.runner.LoadScriptFrom(script, "")
}

// LoadScriptFrom loads a script together with its source (file path or
// URL). Relative import() calls in the script resolve against that source:
// the containing directory for local files, the source URL (same origin
// only) for http(s) sources. See issue #8.
func (w *Window) LoadScriptFrom(script, source string) {
	w.runner.LoadScriptFrom(script, core.ImportBaseFor(source))
}

// CompileScript parses and compiles the currently loaded script without
// executing it, returning any compile error. Use it to validate scripts
// before Execute (issue #6).
func (w *Window) CompileScript() error {
	script, _ := w.runner.Snapshot()
	_, opts := w.runOptions()
	_, err := risor.Compile(context.Background(), script, opts...)
	return err
}

// Clear clears all loaded scripts (imports and main script).
// Call this before loading a new script to avoid accumulation.
func (w *Window) Clear() {
	w.runner.Clear()
}

// GetContentContainer returns the fynerisor Window's content container.
// This allows external code to use the same container that window.SetContent() updates.
func (w *Window) GetContentContainer() *fyne.Container {
	return w.content
}

// Execute executes the loaded Risor script asynchronously and returns
// immediately. The script runs on its own VM goroutine (a vmexec.Run): the
// main body first, then every widget event, one at a time. Sets status to
// "ERROR: ..." on failure or "Ready!" on success.
//
// Calling Execute again stops the previous run first: its VM halts, events
// still arriving for its widgets are dropped, its keyboard shortcuts, main
// menu and drop/stdout handlers are removed, and its imported modules are
// discarded (issue #16). Execute never waits for the previous run, so it is
// safe to call from inside a running script (e.g. browser.Open navigating
// to a new page).
//
// Execute is Start(context.Background()) without waiting for the result;
// see Start, Run and State for the typed lifecycle.
func (w *Window) Execute() {
	w.start(context.Background())
}

func (w *Window) start(ctx context.Context) <-chan error {
	if ctx == nil {
		ctx = context.Background()
	}
	script, base := w.runner.Snapshot()
	env, opts := w.runOptions()

	// startMu serializes concurrent Execute calls (a host and a script's
	// browser.Open, say): each stops the run it replaces, so there is never
	// more than one live run. Nothing here waits on another goroutine.
	w.startMu.Lock()
	if prev := w.run.Load(); prev != nil {
		prev.Stop()
	}
	// Each run gets its own module system state and environment snapshot,
	// carried in its context.
	parent := context.WithValue(ctx, importStateKey{}, newImportState(base, env))
	if w.scriptArgs != nil {
		// os.args() answers from the run's context, not a process global
		// (issue #44).
		parent = osmod.ContextWithArgs(parent, w.scriptArgs)
	}
	// io.stdin/stdout/stderr resolve to this window's streams (issue #42).
	parent = iomod.ContextWithStreams(parent, w)
	r := vmexec.New(parent, vmexec.Config{
		QueueSize: callbackQueueSize,
		// Every error the runtime reports (callback raises, panics,
		// event backlog) flows to this window's sink, not the process
		// stderr (issue #31).
		Report: w.Report,
	})
	// Publishing the run and its first state under statusMutex makes
	// "is this run current?" checks (setStatusFor, setState) exact.
	w.statusMutex.Lock()
	w.run.Store(r)
	w.statusMutex.Unlock()
	w.startMu.Unlock()

	w.setState(r, StateRunning, nil)
	w.resetRunState()
	// However the run ends — Stop, its context, a newer Execute — report it
	// (setState ignores runs that are no longer current).
	go func() {
		<-r.Context().Done()
		w.setState(r, StateStopped, nil)
	}()

	done := make(chan error, 1)
	r.Start(func(ctx context.Context) {
		result, err := w.evalBody(ctx, script, opts)
		// A run that was stopped or replaced reports nothing: its status
		// would overwrite the status of the run that replaced it.
		if ctx.Err() != nil || w.run.Load() != r {
			if err == nil {
				err = vmexec.ErrStopped
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				err = ctxErr
			}
			done <- err
			return
		}
		if err != nil {
			// Report on the window's stderr writer (issue #24); the process's
			// own stderr is never hijacked.
			fmt.Fprintf(w.Stderr(), "Script execution error: %v\n", err)
			w.setStatusFor(r, "ERROR: "+err.Error())
			w.setState(r, StateFailed, err)
			done <- err
			return
		}
		if result != nil {
			if objResult, ok := result.(object.Object); ok && objResult != object.Nil && w.resultCallback != nil {
				w.resultCallback(fmt.Sprintln(objResult))
			}
		}

		w.setStatusFor(r, "Ready!")
		w.setState(r, StateReady, nil)
		done <- nil
	})
	return done
}

// evalBody runs the main script body, converting a panic into an error so
// the lifecycle still reports StateFailed.
func (w *Window) evalBody(ctx context.Context, script string, opts []risor.Option) (result any, err error) {
	defer func() {
		if p := recover(); p != nil {
			result, err = nil, fmt.Errorf("panic in script: %v", p)
		}
	}()
	return core.EvalScript(ctx, script, opts)
}

// runOptions snapshots the script environment for one run: a copy of the
// env map (so RegisterGlobal never mutates a map a running VM reads) plus
// the opaque user options.
func (w *Window) runOptions() (map[string]any, []risor.Option) {
	w.envMu.RLock()
	defer w.envMu.RUnlock()
	env := make(map[string]any, len(w.env))
	for k, v := range w.env {
		env[k] = v
	}
	opts := append([]risor.Option{risor.WithEnv(env)}, w.userGlobals...)
	return env, opts
}

// moduleEnabled reports whether a module (or registered global) is enabled.
func (w *Window) moduleEnabled(name string) bool {
	w.envMu.RLock()
	defer w.envMu.RUnlock()
	return w.enabledModules[name]
}

// isCurrent reports whether ctx belongs to the window's current run.
func (w *Window) isCurrent(ctx context.Context) bool {
	r := vmexec.FromContext(ctx)
	return r != nil && r == w.run.Load()
}

// stale reports whether ctx belongs to a run that is no longer the window's
// current run (stopped or replaced). Builtins called outside any run (a host
// invoking them directly) are never stale. Window builtins run on the UI
// thread, like resetRunState's cleanup, so "check stale, then install" can
// never interleave with the cleanup (issue #52).
func (w *Window) stale(ctx context.Context) bool {
	r := vmexec.FromContext(ctx)
	return r != nil && (r != w.run.Load() || r.Stopped())
}

// setStatusFor sets the status on behalf of run r: ignored unless r is the
// current run. The check and the write happen under statusMutex, the lock
// under which start publishes a new run, so a replaced run can never
// overwrite the status of the run that replaced it.
func (w *Window) setStatusFor(r *vmexec.Run, status string) {
	w.statusMutex.Lock()
	if r == nil || w.run.Load() != r {
		w.statusMutex.Unlock()
		return
	}
	w.status = status
	w.statusMutex.Unlock()
	if w.statusCallback != nil {
		w.statusCallback(status)
	}
}

// resetRunState removes window state owned by runs that are no longer
// current: keyboard shortcuts, the main menu (only if a script set it) and
// the drop and stdout-capture hooks. State is tagged with its owning run,
// so this never removes what the current run has already installed, and it
// never needs to wait: the removal of Fyne state is scheduled on the UI
// thread, where all script access to the window runs (issue #52).
func (w *Window) resetRunState() {
	current := w.run.Load()

	w.cbMu.Lock()
	if w.dropOwner != current {
		w.onDropped, w.dropOwner = nil, nil
	}
	w.cbMu.Unlock()

	// Drop stdout redirects installed by a superseded run (script
	// SetStdout/CaptureStdout); a host-configured writer has no owner and
	// persists across runs (issue #24).
	w.stdoutMu.Lock()
	if w.stdoutOwner != nil && w.stdoutOwner != current {
		w.stdoutOverride, w.stdoutOwner = nil, nil
	}
	w.stdoutMu.Unlock()

	if w.FyneWindow == nil {
		return
	}
	host := w.FyneWindow
	guithread.Do(func() {
		cur := w.run.Load()
		w.shortcutMutex.Lock()
		var stale []fyne.Shortcut
		for key, e := range w.shortcuts {
			if e.owner != cur {
				stale = append(stale, e.shortcut)
				delete(w.shortcuts, key)
			}
		}
		w.shortcutMutex.Unlock()
		for _, sc := range stale {
			host.Canvas().RemoveShortcut(sc)
		}
		if owner := w.menuOwner.Load(); owner != nil && owner != cur {
			if w.menuOwner.CompareAndSwap(owner, nil) {
				host.SetMainMenu(nil)
			}
		}
	})
}

// Stop cancels the currently executing script and stops processing queued
// callbacks. It is safe to call when no script is running and is used by
// CanvasObject to release a script when the embedded object is discarded.
func (w *Window) Stop() {
	if r := w.run.Load(); r != nil {
		r.Stop()
		w.setState(r, StateStopped, nil)
	}
}

// ShowAndRun displays the window and runs the Fyne application main loop.
// This is a blocking call that runs until the window is closed.
//
// For a CanvasObject there is no window to show, so ShowAndRun is a no-op;
// the host application owns the main loop.
func (w *Window) ShowAndRun() {
	if w.FyneWindow == nil {
		return
	}
	w.FyneWindow.ShowAndRun()
}

// Resize changes the size of the window.
// This is useful when creating windows with NewApp and needing custom sizes.
//
// Example:
//
//	fw := fynerisor.NewApp("My App")
//	fw.Resize(800, 600)
//	fw.LoadScript(script)
//	fw.Execute()
//	fw.ShowAndRun()
func (w *Window) Resize(width, height float32) {
	if w.FyneWindow != nil {
		w.FyneWindow.Resize(fyne.NewSize(width, height))
		return
	}
	// Embedded CanvasObject: there is no window to resize. Resizing the content
	// at least gives scripts a way to influence their own layout.
	guithread.Do(func() {
		w.content.Resize(fyne.NewSize(width, height))
	})
}

func (w *Window) Type() object.Type {
	return WindowType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (w *Window) Inspect() string {
	return guithread.Inspect(w.inspectUI)
}

func (w *Window) inspectUI() string {
	return "window"
}

func (w *Window) Interface() interface{} {
	return nil
}

func (w *Window) IsTruthy() bool {
	return true
}

func (w *Window) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", WindowType, opType)
}
func (w *Window) Equals(other object.Object) bool {
	return w == other
}

func (w *Window) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (w *Window) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, w.setAttrUI)
}

func (w *Window) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", WindowType, name)
}

// IsCanvasObject is the shared interface for objects convertible to a
// fyne.CanvasObject (deduplicated with gui/widget, issue #33).
type IsCanvasObject = risorwidget.IsCanvasObject

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (w *Window) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, w.getAttrUI)
}

func (w *Window) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "DroppedPaths":
		return object.NewStringList(w.droppedPaths), true

	case "Close":
		return object.NewBuiltin("w.Close", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			if w.FyneWindow == nil {
				return nil, object.Errorf("window.Close: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
			}
			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}
			guithread.Do(func() {
				w.FyneWindow.Close()
			})
			return object.Nil, nil
		}), true

	case "SetContent":
		return object.NewBuiltin("w.SetContent", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			obj, ok := args[0].(IsCanvasObject)
			if !ok {
				return nil, object.Errorf("argument error: expected IsCanvasObject, got %s", args[0].Type())
			}

			guithread.Do(func() {
				// A run that has been stopped or replaced must not swap the
				// content of the run that replaced it.
				if w.stale(ctx) {
					return
				}
				w.contentMu.Lock()
				defer w.contentMu.Unlock()
				w.content.Objects = []fyne.CanvasObject{obj.CanvasObject()}
				w.content.Refresh()
			})

			return object.Nil, nil
		}), true

	case "SetStdout", "CaptureStdout":
		return object.NewBuiltin("w.SetStdout", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}
			owner := vmexec.FromContext(ctx)

			// A Log widget is the preferred sink: it is an io.Writer fed by
			// print on the VM goroutine, appended via pure Go, never
			// re-entering the VM (issue #24). nil restores the default.
			// A closure is bridged as an io.Writer that queues each complete
			// line onto its run like any other event.
			switch arg := args[0].(type) {
			case *object.NilType:
				w.setStdout(nil, owner)
			case *risorwidget.Log:
				w.setStdout(arg, owner)
			case *object.Closure:
				cb, err := vmexec.NewCallback(ctx, "window.SetStdout", args[0])
				if err != nil {
					return nil, err
				}
				w.setStdout(&callbackWriter{cb: cb}, owner)
			default:
				return nil, object.Errorf("argument error: expected widget.Log, function, or nil, got %s", args[0].Type())
			}
			return object.Nil, nil
		}), true

	case "OnDropped":
		return object.NewBuiltin("w.OnDropped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "window.OnDropped", args[0])
			if err != nil {
				return nil, err
			}
			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}

			w.cbMu.Lock()
			w.onDropped = func(droppedPaths []string) {
				cb.Fire(object.NewStringList(droppedPaths))
			}
			w.dropOwner = vmexec.FromContext(ctx)
			w.cbMu.Unlock()

			return object.Nil, nil
		}), true

	case "SetStatus":
		return object.NewBuiltin("w.SetStatus", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			status, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			// A stopped or replaced run must not overwrite the status of
			// the run that replaced it (issues #16, #52).
			if r := vmexec.FromContext(ctx); r != nil {
				if w.stale(ctx) {
					return nil, vmexec.ErrStopped
				}
				w.setStatusFor(r, status)
			} else {
				w.SetStatus(status)
			}

			return object.Nil, nil
		}), true

	case "Do":
		return object.NewBuiltin("w.Do", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "window.Do", args[0])
			if err != nil {
				return nil, err
			}

			// Runs on the run's VM goroutine after the current body/event;
			// widget methods marshal their own UI work (guithread.Do).
			r := vmexec.FromContext(ctx)
			cb.OnError(func(err error) { w.setStatusFor(r, "ERROR: "+err.Error()) }).Fire()

			return object.Nil, nil
		}), true

	case "Size":
		if w.FyneWindow != nil {
			size := w.FyneWindow.Canvas().Size()
			return object.NewMap(map[string]object.Object{
				"width":  object.NewFloat(float64(size.Width)),
				"height": object.NewFloat(float64(size.Height)),
			}), true
		}
		size := w.content.Size()
		if size.Width == 0 && size.Height == 0 {
			size = w.content.MinSize()
		}
		return object.NewMap(map[string]object.Object{
			"width":  object.NewFloat(float64(size.Width)),
			"height": object.NewFloat(float64(size.Height)),
		}), true

	case "Resize":
		return object.NewBuiltin("w.Resize", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			width, err := object.AsFloat(args[0])
			if err != nil {
				return nil, object.Errorf("argument error: width must be a number, got %s", args[0].Type())
			}

			height, err := object.AsFloat(args[1])
			if err != nil {
				return nil, object.Errorf("argument error: height must be a number, got %s", args[1].Type())
			}

			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}
			// Must call Resize on the UI thread
			guithread.Do(func() {
				w.Resize(float32(width), float32(height))
			})

			return object.Nil, nil
		}), true

	case "AddShortcut":
		return object.NewBuiltin("window.AddShortcut", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("AddShortcut requires 2 arguments (shortcut, callback)")
			}
			if w.FyneWindow == nil {
				return nil, object.Errorf("window.AddShortcut: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
			}

			// Parse shortcut string
			shortcutStr, ok := args[0].(*object.String)
			if !ok {
				return nil, fmt.Errorf("first argument must be string (e.g., 'Ctrl+S')")
			}

			cb, err := vmexec.NewCallback(ctx, "window.AddShortcut", args[1])
			if err != nil {
				return nil, err
			}
			run := vmexec.FromContext(ctx)
			cb.OnError(func(err error) { w.setStatusFor(run, "ERROR: "+err.Error()) })
			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}

			// Parse the shortcut string
			shortcut, err := ParseShortcutString(shortcutStr.Value())
			if err != nil {
				return nil, fmt.Errorf("invalid shortcut: %w", err)
			}

			// Register shortcut on canvas. Event: the handler is queued on
			// the script's VM goroutine.
			w.FyneWindow.Canvas().AddShortcut(shortcut, func(s fyne.Shortcut) {
				cb.Fire()
			})

			// Track shortcut for removal (and for cleanup when the run ends)
			w.shortcutMutex.Lock()
			w.shortcuts[shortcutStr.Value()] = shortcutEntry{shortcut: shortcut, owner: run}
			w.shortcutMutex.Unlock()

			return object.Nil, nil
		}), true

	case "RemoveShortcut":
		return object.NewBuiltin("window.RemoveShortcut", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("RemoveShortcut requires 1 argument (shortcut string)")
			}
			if w.FyneWindow == nil {
				return nil, object.Errorf("window.RemoveShortcut: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
			}

			shortcutStr, ok := args[0].(*object.String)
			if !ok {
				return nil, fmt.Errorf("argument must be string")
			}

			if w.stale(ctx) {
				return nil, vmexec.ErrStopped
			}
			// Find and remove shortcut
			w.shortcutMutex.Lock()
			entry, found := w.shortcuts[shortcutStr.Value()]
			if found {
				delete(w.shortcuts, shortcutStr.Value())
			}
			w.shortcutMutex.Unlock()

			if found {
				w.FyneWindow.Canvas().RemoveShortcut(entry.shortcut)
			}

			return object.Nil, nil
		}), true

	case "SetMainMenu":
		return object.NewBuiltin("window.SetMainMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("SetMainMenu requires 1 argument (MainMenu)")
			}
			if w.FyneWindow == nil {
				return nil, object.Errorf("window.SetMainMenu: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
			}

			mainMenuObj, ok := args[0].(*MainMenu)
			if !ok {
				return nil, fmt.Errorf("argument must be MainMenu object")
			}
			if w.stale(ctx) {
				return nil, vmexec.ErrStopped // stopped run (issues #16, #52)
			}

			// Remember which run owns the menu, so the next run can remove
			// it together with its (dead) item callbacks.
			w.menuOwner.Store(vmexec.FromContext(ctx))
			guithread.Do(func() {
				w.FyneWindow.SetMainMenu(mainMenuObj.instance)
			})

			return object.Nil, nil
		}), true
	}
	return nil, false
}

// RegisterGlobal adds a global object to the script environment after window
// creation. This is useful for objects that need to reference the window or
// other components that are created after the window initialization
// (e.g., browser, custom plugins).
//
// It is safe to call while a script runs: each run uses a snapshot of the
// environment taken when it starts, so the new global is visible from the
// next Execute.
func (w *Window) RegisterGlobal(name string, value any) {
	w.envMu.Lock()
	defer w.envMu.Unlock()
	w.env[name] = value
	w.enabledModules[name] = true // Register for require() validation
}

func (w *Window) Canvas() fyne.Canvas {
	if w.FyneWindow == nil {
		return nil
	}
	return w.FyneWindow.Canvas()
}
