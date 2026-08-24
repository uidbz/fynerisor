package gui

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/core"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	risorwidget "github.com/uidbz/fynerisor/gui/widget"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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
	guithread.SetMain()
	a := app.New()
	w := a.NewWindow(title)
	return NewWindow(w, opts...)
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
	Status     string

	content        *fyne.Container
	globals        []risor.Option
	env            map[string]any  // Merged built-in globals (shared with imported module VMs)
	userGlobals    []risor.Option  // Opaque user-supplied globals (via WithGlobals)
	enabledModules map[string]bool // Track which modules are enabled
	appName        string          // Name of the embedding application
	droppedPaths   []string
	onDropped      func(droppedPaths []string)
	onNewStdout    func(text string)
	stdoutSink     func(text string) // Go sink for captured stdout; bypasses the Risor VM
	functionCalls  chan func()
	cancelExec     context.CancelFunc // Cancel function for current execution
	statusCallback func(string)
	resultCallback func(string)
	runner         *ScriptRunner

	// Module import system
	moduleCache map[string]*core.ImportedModule // Cache of imported modules
	importStack []string                        // Track currently importing modules for circular detection
	moduleMutex sync.Mutex                      // Protect cache from concurrent access

	// Keyboard shortcut system
	shortcuts      map[string]fyne.Shortcut // Map shortcut string to Shortcut object
	shortcutMutex  sync.Mutex               // Protect concurrent access
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
	guithread.SetMain()
	w := &Window{
		FyneWindow:     window,
		enabledModules: make(map[string]bool),
		appName:        "fynerisor", // default
		moduleCache:    make(map[string]*core.ImportedModule),
		importStack:    []string{},
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

	// Initialize content container and set it on the window
	w.content = container.NewStack()
	window.SetContent(w.content)

	// Setup drag-and-drop handler
	window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		paths := make([]string, len(uris))
		for i, uri := range uris {
			paths[i] = uri.Path()
		}
		w.droppedPaths = paths
		if w.onDropped != nil {
			w.functionCalls <- func() { guithread.Do(func() { w.onDropped(paths) }) }
		}
	})

	return w
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

func newGoBuiltin() *object.Builtin {
	return object.NewBuiltin("go", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("go: expected 1 argument (function), got %d", len(args))
		}

		fn, ok := args[0].(*object.Closure)
		if !ok {
			return nil, fmt.Errorf("go: expected function, got %s", args[0].Type())
		}

		callFunc, ok := object.GetCallFunc(ctx)
		if !ok {
			return nil, fmt.Errorf("go: unable to get call function")
		}

		// Spawn goroutine to execute the closure
		go func() {
			_, err := safeCall(callFunc, ctx, fn, []object.Object{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "go routine error: %v\n", err)
			}
		}()

		return object.Nil, nil
	})
}

// SetStatus updates the window status and calls the status callback if provided.
func (w *Window) SetStatus(status string) {
	w.Status = status
	if w.statusCallback != nil {
		w.statusCallback(status)
	}
}

// LoadScript loads a Risor script that will be executed when Execute() is called.
// If ImportScript() was called before LoadScript(), the imports will be prepended.
func (w *Window) LoadScript(script string) {
	w.runner.LoadScript(script)
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

// Execute executes the loaded Risor script in a goroutine.
// It evaluates the script with all configured globals and processes any
// queued function calls from callbacks. Sets status to "ERROR: ..." on failure
// or "Ready!" on success.
//
// This method returns immediately; script execution happens asynchronously.
// Call LoadScript() before calling this method.
//
// Multiple calls to Execute() are safe. Previous execution will be cancelled
// to prevent resource leaks from abandoned goroutines.
func (w *Window) Execute() {
	// Cancel any previous execution
	if w.cancelExec != nil {
		w.cancelExec()
	}

	// Create new context for this execution
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelExec = cancel

	// Create new channel for this execution
	// Don't close old channel - let it be garbage collected
	// Old goroutine will exit via context cancellation
	callChan := make(chan func(), 10)
	w.functionCalls = callChan

	go func() {
		result, err := w.runner.Eval()
		if err != nil {
			// Check if it was cancelled
			if ctx.Err() == context.Canceled {
				return // Silently exit if cancelled
			}
			// Log error to stderr so it's visible in terminal
			fmt.Fprintf(os.Stderr, "Script execution error: %v\n", err)
			w.SetStatus("ERROR: " + err.Error())
			return
		}
		if result != nil {
			if objResult, ok := result.(object.Object); ok && objResult != object.Nil && w.resultCallback != nil {
				w.resultCallback(fmt.Sprintln(objResult))
			}
		}

		w.SetStatus("Ready!")

		// Process callbacks until cancelled
		// Don't close the channel - callbacks might still arrive
		for {
			select {
			case <-ctx.Done():
				return // Exit if cancelled
			case x := <-callChan:
				x()
			}
		}
	}()
}

// ShowAndRun displays the window and runs the Fyne application main loop.
// This is a blocking call that runs until the window is closed.
func (w *Window) ShowAndRun() {
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
	w.FyneWindow.Resize(fyne.NewSize(width, height))
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

	case "SetContent":
		return object.NewBuiltin("w.SetContent", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			obj, ok := args[0].(IsCanvasObject)
			if !ok {
				return object.Errorf("argument error: expected IsCanvasObject, got %s", args[1].Type()), nil
			}

			guithread.Do(func() {
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
				w.stdoutSink = logWidget.AppendGo
				w.captureStdout()
				return object.Nil, nil
			}

			// Legacy form: a closure. This re-enters the VM on each line and is
			// NOT safe together with go() that produces output — prefer passing
			// the Log widget instead.
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected widget.Log or function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			w.onNewStdout = func(text string) {
				safeCall(callFunc, ctx, fn, []object.Object{object.NewString(text)})
			}

			w.captureStdout()

			return object.Nil, nil
		}), true

	case "OnDropped":
		return object.NewBuiltin("w.OnDropped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			w.onDropped = func(droppedPaths []string) {
				safeCall(callFunc, ctx, fn, []object.Object{object.NewStringList(droppedPaths)})
			}

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

			w.SetStatus(status)

			return object.Nil, nil
		}), true

	case "Do":
		return object.NewBuiltin("w.Do", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("Do: unable to get call function"), nil
			}

			// Queue the function to run on the GUI thread
			w.functionCalls <- func() {
				guithread.Do(func() {
					_, err := safeCall(callFunc, ctx, fn, []object.Object{})
					if err != nil {
						w.SetStatus("ERROR: " + err.Error())
					}
				})
			}

			return object.Nil, nil
		}), true

	case "Canvas":
		return w, true  // Window itself implements Canvas() method

	case "Size":
		size := w.FyneWindow.Canvas().Size()
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

			// Parse shortcut string
			shortcutStr, ok := args[0].(*object.String)
			if !ok {
				return nil, fmt.Errorf("first argument must be string (e.g., 'Ctrl+S')")
			}

			// Get callback function
			callback, ok := args[1].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("second argument must be function")
			}

			// Parse the shortcut string
			shortcut, err := ParseShortcutString(shortcutStr.Value())
			if err != nil {
				return nil, fmt.Errorf("invalid shortcut: %w", err)
			}

			// Get call function from context
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("unable to get call function")
			}

			// Register shortcut on canvas
			w.FyneWindow.Canvas().AddShortcut(shortcut, func(s fyne.Shortcut) {
				// Queue callback on function channel (for thread safety)
				w.functionCalls <- func() {
					guithread.Do(func() {
						_, err := safeCall(callFunc, ctx, callback, []object.Object{})
						if err != nil {
							w.SetStatus("ERROR: " + err.Error())
						}
					})
				}
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

			mainMenuObj, ok := args[0].(*MainMenu)
			if !ok {
				return nil, fmt.Errorf("argument must be MainMenu object")
			}

			w.FyneWindow.SetMainMenu(mainMenuObj.instance)

			return object.Nil, nil
		}), true
	}
	return nil, false
}

// Canvas returns the Fyne canvas for this window.
// This allows scripts to create PopUp widgets.
// RegisterGlobal adds a global object to the script environment after window creation
// This is useful for objects that need to reference the window or other components
// that are created after the window initialization (e.g., browser, custom plugins)
func (w *Window) RegisterGlobal(name string, value any) {
	w.env[name] = value
	w.enabledModules[name] = true // Register for require() validation
}

func (w *Window) Canvas() fyne.Canvas {
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
				// entering the non-threadsafe Risor VM, so capture can run
				// concurrently with a go() script goroutine that holds the VM.
				// Only fall back to the script closure (which re-enters the VM)
				// when no Go sink is registered.
				if w.stdoutSink != nil {
					sink := w.stdoutSink
					guithread.Do(func() { sink(msg) })
				} else {
					w.functionCalls <- func() { guithread.Do(func() { w.onNewStdout(msg) }) }
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

