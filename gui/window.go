package gui

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/core"

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
//	    fynerisor.WithSQL(),
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
// file drops, and window.CaptureStdout() to capture printed output.
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
	droppedPaths   []string
	cbMu           sync.Mutex // guards onDropped, onNewStdout, stdoutSink: set by the VM goroutine, read by the UI thread / capture goroutine
	onDropped      func(droppedPaths []string)
	onNewStdout    func(text string)
	stdoutSink     func(text string)          // Go sink for captured stdout; bypasses the Risor VM
	run            atomic.Pointer[vmexec.Run] // current script run: VM goroutine, guard, event queue, context (issue #37)
	contentMu      sync.Mutex                 // serializes SetContent swaps between overlapping runs
	statusCallback func(string)
	resultCallback func(string)
	runner         *ScriptRunner

	// Module import system: the module cache and import stack live in the
	// per-run importState (see import.go); only the entry script's base is
	// kept here, captured by the next Execute.
	scriptBase string // Base of the entry script (dir or URL); "" = process CWD

	// Keyboard shortcut system
	shortcuts     map[string]fyne.Shortcut // Map shortcut string to Shortcut object
	shortcutMutex sync.Mutex               // Protect concurrent access
	scriptMenu    atomic.Bool              // the running script installed the main menu
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
func newWindow(opts ...Option) *Window {
	// Record the GUI goroutine without overwriting it: a host may create an
	// embedded view on a background goroutine (issue #14).
	guithread.Init(false)
	w := &Window{
		enabledModules: make(map[string]bool),
		appName:        "fynerisor", // default
		shortcuts:      make(map[string]fyne.Shortcut),
	}

	// Apply appName and callback options first (before globals)
	for _, opt := range opts {
		if _, ok := opt.(appNameOption); ok {
			opt.applyToWindow(w)
		}
		if _, ok := opt.(windowOption); ok {
			opt.applyToWindow(w)
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
		"print":     newPrintBuiltin(),
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

	// Apply module options (which populate w.env, w.userGlobals and enabledModules)
	for _, opt := range opts {
		if _, ok := opt.(moduleOption); ok {
			opt.applyToWindow(w)
		}
	}

	// Compose the final risor options: the merged env first, then any opaque
	// user-supplied globals (from WithGlobals).
	w.globals = append([]risor.Option{risor.WithEnv(w.env)}, w.userGlobals...)

	// Create script runner
	w.runner = NewScriptRunner(w.globals)

	// Initialize content container. A standalone NewWindow attaches it to the
	// Fyne window; an embedded CanvasObject renders it as a widget.
	w.content = container.NewStack()

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

func newPrintBuiltin() *object.Builtin {
	return object.NewBuiltin("print", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		values := make([]any, len(args))
		for i, arg := range args {
			values[i] = object.PrintableValue(arg)
		}
		fmt.Println(values...)
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
	w.scriptBase = ""
	w.runner.LoadScript(script)
}

// LoadScriptFrom loads a script together with its source (file path or
// URL). Relative import() calls in the script resolve against that source:
// the containing directory for local files, the source URL (same origin
// only) for http(s) sources. See issue #8.
func (w *Window) LoadScriptFrom(script, source string) {
	w.scriptBase = core.ImportBaseFor(source)
	w.runner.LoadScript(script)
}

// CompileScript parses and compiles the currently loaded script without
// executing it, returning any compile error. Use it to validate scripts
// before Execute (issue #6).
func (w *Window) CompileScript() error {
	return w.runner.Compile()
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

// ImportScript is deprecated and has been removed in favor of runtime import().
// Use the import() function directly in scripts for module-scoped imports:
//
//	let utils = import("utils.risor")
//	utils.myFunction()
//
// This provides proper namespacing and prevents global scope pollution.
// The old concatenation-based import system is no longer supported.
func (w *Window) ImportScript(source string) error {
	return fmt.Errorf("ImportScript() is deprecated: use import() function in scripts instead")
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
	if prev := w.run.Load(); prev != nil {
		prev.Stop()
	}
	w.resetRunState()

	// Each run gets its own module system state, carried in its context.
	parent := context.WithValue(ctx, importStateKey{}, newImportState(w.scriptBase))
	r := vmexec.New(parent, vmexec.Config{QueueSize: callbackQueueSize})
	w.run.Store(r)
	w.setState(r, StateRunning, nil)
	// However the run ends — Stop, its context, a newer Execute — report it
	// (setState ignores runs that are no longer current).
	go func() {
		<-r.Context().Done()
		w.setState(r, StateStopped, nil)
	}()

	done := make(chan error, 1)
	r.Start(func(ctx context.Context) {
		result, err := w.evalBody(ctx)
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
			// Log error to stderr so it's visible in terminal
			fmt.Fprintf(os.Stderr, "Script execution error: %v\n", err)
			w.SetStatus("ERROR: " + err.Error())
			w.setState(r, StateFailed, err)
			done <- err
			return
		}
		if result != nil {
			if objResult, ok := result.(object.Object); ok && objResult != object.Nil && w.resultCallback != nil {
				w.resultCallback(fmt.Sprintln(objResult))
			}
		}

		w.SetStatus("Ready!")
		w.setState(r, StateReady, nil)
		done <- nil
	})
	return done
}

// evalBody runs the main script body, converting a panic into an error so
// the lifecycle still reports StateFailed.
func (w *Window) evalBody(ctx context.Context) (result any, err error) {
	defer func() {
		if p := recover(); p != nil {
			result, err = nil, fmt.Errorf("panic in script: %v", p)
		}
	}()
	return w.runner.EvalContext(ctx)
}

// resetRunState removes what the previous run installed on the window:
// its keyboard shortcuts, the main menu (only if a script set it), and its
// drop and stdout-capture handlers. Their callbacks belong to a stopped run
// and would only be dropped; removing them frees the keys and menu for the
// next run.
func (w *Window) resetRunState() {
	w.cbMu.Lock()
	w.onDropped = nil
	w.onNewStdout = nil
	w.stdoutSink = nil
	w.cbMu.Unlock()

	w.shortcutMutex.Lock()
	old := w.shortcuts
	w.shortcuts = make(map[string]fyne.Shortcut)
	w.shortcutMutex.Unlock()

	resetMenu := w.scriptMenu.Swap(false)
	if w.FyneWindow == nil || (len(old) == 0 && !resetMenu) {
		return
	}
	host := w.FyneWindow
	guithread.Do(func() {
		for _, sc := range old {
			host.Canvas().RemoveShortcut(sc)
		}
		if resetMenu {
			host.SetMainMenu(nil)
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

func (w *Window) Inspect() string {
	return "window"
}

func (w *Window) Interface() interface{} {
	return nil
}

func (w *Window) IsTruthy() bool {
	return true
}

func (w *Window) Cost() int {
	return 0
}

func (w *Window) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'w'")
}

func (w *Window) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(WindowType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", WindowType, opType)
	return errObj, err
}
func (w *Window) Equals(other object.Object) bool {
	return w == other
}

func (w *Window) Attrs() []object.AttrSpec {
	return nil
}

func (w *Window) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", WindowType, name)
}

type IsCanvasObject interface {
	CanvasObject() fyne.CanvasObject
}

func (w *Window) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "DroppedPaths":
		return object.NewStringList(w.droppedPaths), true

	case "Close":
		return object.NewBuiltin("w.Close", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			if w.FyneWindow == nil {
				return object.Errorf("window.Close: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)"), nil
			}
			guithread.Do(func() {
				w.FyneWindow.Close()
			})
			return object.Nil, nil
		}), true

	case "SetContent":
		return object.NewBuiltin("w.SetContent", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			obj, ok := args[0].(IsCanvasObject)
			if !ok {
				return object.Errorf("argument error: expected IsCanvasObject, got %s", args[0].Type()), nil
			}

			guithread.Do(func() {
				// A run that has been stopped or replaced must not swap the
				// content of the run that replaced it.
				if ctx.Err() != nil {
					return
				}
				w.contentMu.Lock()
				defer w.contentMu.Unlock()
				w.content.Objects = []fyne.CanvasObject{obj.CanvasObject()}
				w.content.Refresh()
			})

			return object.Nil, nil
		}), true

	case "CaptureStdout":
		return object.NewBuiltin("w.CaptureStdout", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			// Preferred form: pass a Log widget directly. Captured lines are
			// appended via pure Go (AppendGo), never re-entering the Risor VM,
			// so stdout capture is safe while a go() goroutine holds the single
			// non-threadsafe VM.
			if logWidget, ok := args[0].(*risorwidget.Log); ok {
				w.cbMu.Lock()
				w.stdoutSink = logWidget.AppendGo
				w.cbMu.Unlock()
				w.captureStdout()
				return object.Nil, nil
			}

			// Legacy form: a closure. This re-enters the VM on each line and is
			// NOT safe together with go() that produces output — prefer passing
			// the Log widget instead.
			if _, ok := args[0].(*object.Closure); !ok {
				return object.Errorf("argument error: expected widget.Log or function, got %s", args[0].Type()), nil
			}
			cb, err := vmexec.NewCallback(ctx, "window.CaptureStdout", args[0])
			if err != nil {
				return nil, err
			}

			w.cbMu.Lock()
			w.onNewStdout = func(text string) {
				cb.Fire(object.NewString(text))
			}
			w.cbMu.Unlock()

			w.captureStdout()

			return object.Nil, nil
		}), true

	case "OnDropped":
		return object.NewBuiltin("w.OnDropped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "window.OnDropped", args[0])
			if err != nil {
				return nil, err
			}

			w.cbMu.Lock()
			w.onDropped = func(droppedPaths []string) {
				cb.Fire(object.NewStringList(droppedPaths))
			}
			w.cbMu.Unlock()

			return object.Nil, nil
		}), true

	case "SetStatus":
		return object.NewBuiltin("w.SetStatus", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			status, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			// A stopped or replaced run must not overwrite the status of
			// the run that replaced it (issue #16).
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			w.SetStatus(status)

			return object.Nil, nil
		}), true

	case "Do":
		return object.NewBuiltin("w.Do", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "window.Do", args[0])
			if err != nil {
				return nil, err
			}

			// Runs on the run's VM goroutine after the current body/event;
			// widget methods marshal their own UI work (guithread.Do).
			cb.OnError(func(err error) { w.SetStatus("ERROR: " + err.Error()) }).Fire()

			return object.Nil, nil
		}), true

	case "Canvas":
		return w, true // Window itself implements Canvas() method

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
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			width, err := object.AsFloat(args[0])
			if err != nil {
				return object.Errorf("argument error: width must be a number, got %s", args[0].Type()), nil
			}

			height, err := object.AsFloat(args[1])
			if err != nil {
				return object.Errorf("argument error: height must be a number, got %s", args[1].Type()), nil
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
				return object.Errorf("window.AddShortcut: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)"), nil
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
			cb.OnError(func(err error) { w.SetStatus("ERROR: " + err.Error()) })

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

			// Track shortcut for removal
			w.shortcutMutex.Lock()
			w.shortcuts[shortcutStr.Value()] = shortcut
			w.shortcutMutex.Unlock()

			return object.Nil, nil
		}), true

	case "RemoveShortcut":
		return object.NewBuiltin("window.RemoveShortcut", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("RemoveShortcut requires 1 argument (shortcut string)")
			}
			if w.FyneWindow == nil {
				return object.Errorf("window.RemoveShortcut: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)"), nil
			}

			shortcutStr, ok := args[0].(*object.String)
			if !ok {
				return nil, fmt.Errorf("argument must be string")
			}

			// Find and remove shortcut
			w.shortcutMutex.Lock()
			shortcut, found := w.shortcuts[shortcutStr.Value()]
			if found {
				delete(w.shortcuts, shortcutStr.Value())
			}
			w.shortcutMutex.Unlock()

			if found {
				w.FyneWindow.Canvas().RemoveShortcut(shortcut)
			}

			return object.Nil, nil
		}), true

	case "SetMainMenu":
		return object.NewBuiltin("window.SetMainMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("SetMainMenu requires 1 argument (MainMenu)")
			}
			if w.FyneWindow == nil {
				return object.Errorf("window.SetMainMenu: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)"), nil
			}

			mainMenuObj, ok := args[0].(*MainMenu)
			if !ok {
				return nil, fmt.Errorf("argument must be MainMenu object")
			}
			if err := ctx.Err(); err != nil {
				return nil, err // stopped run (issue #16)
			}

			// Remember that the script owns the menu, so the next run can
			// remove it together with its (dead) item callbacks.
			w.scriptMenu.Store(true)
			guithread.Do(func() {
				if ctx.Err() == nil {
					w.FyneWindow.SetMainMenu(mainMenuObj.instance)
				}
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
func (w *Window) RegisterGlobal(name string, value any) {
	w.env[name] = value
	w.enabledModules[name] = true // Register for require() validation
}

func (w *Window) Canvas() fyne.Canvas {
	if w.FyneWindow == nil {
		return nil
	}
	return w.FyneWindow.Canvas()
}

func (w *Window) captureStdout() {
	// Save the original os.Stdout
	// originalStdout := os.Stdout

	// Create a pipe
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// Redirect os.Stdout to the pipe writer
	os.Stdout = writer

	// Unbounded, mutex-guarded buffer between the pipe reader and the GUI
	// dispatcher. The pipe MUST always be drained: if the reader ever blocks
	// (e.g. the GUI thread is busy and a bounded channel fills), the OS pipe
	// buffer fills and any fmt.Println/print on the redirected stdout blocks
	// the calling goroutine, deadlocking the script. Buffering here decouples
	// pipe draining from GUI processing so print() never blocks.
	var (
		mu      sync.Mutex
		pending []string
		notify  = make(chan struct{}, 1)
	)

	// Goroutine to read from the pipe; never blocks on the GUI.
	//
	// Use bufio.Reader.ReadString rather than bufio.Scanner: Scanner has a
	// 64KB max line length and returns false (killing this loop) on a longer
	// line. If the reader dies, the pipe stops draining, the OS pipe buffer
	// fills, and the next fmt.Println/print blocks the calling goroutine
	// forever - freezing the script. ReadString has no length limit.
	go func() {
		br := bufio.NewReader(reader)
		for {
			line, err := br.ReadString('\n')
			if len(line) > 0 {
				mu.Lock()
				pending = append(pending, strings.TrimRight(line, "\n"))
				mu.Unlock()
				select {
				case notify <- struct{}{}:
				default:
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Dispatcher: forwards buffered lines to the GUI as fast as it can drain.
	go func() {
		for range notify {
			for {
				mu.Lock()
				if len(pending) == 0 {
					mu.Unlock()
					break
				}
				msg := pending[0]
				pending = pending[1:]
				mu.Unlock()
				// Prefer the Go sink: it appends to the log widget without
				// entering the Risor VM at all. Otherwise the script closure
				// is queued on its run like any other event.
				w.cbMu.Lock()
				sink, onLine := w.stdoutSink, w.onNewStdout
				w.cbMu.Unlock()
				switch {
				case sink != nil:
					guithread.Do(func() { sink(msg) })
				case onLine != nil:
					onLine(msg)
				}
			}
		}
	}()
}

// loadScriptFromPath loads script code from a file path or HTTP(S) URL
func loadScriptFromPath(path string) (string, error) {
	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return loadScriptFromURL(path)
	}

	// Otherwise treat as file path
	return loadScriptFromFile(path)
}

// loadScriptFromFile loads script code from a local file
func loadScriptFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

// loadScriptFromURL loads script code from an HTTP(S) URL
func loadScriptFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return string(data), nil
}
