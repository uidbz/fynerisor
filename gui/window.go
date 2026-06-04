package gui

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"

	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"fyne.io/fyne/v2"
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
	enabledModules map[string]bool // Track which modules are enabled
	appName        string          // Name of the embedding application
	droppedPaths   []string
	onDropped      func(droppedPaths []string)
	onNewStdout    func(text string)
	functionCalls  chan func()
	cancelExec     context.CancelFunc // Cancel function for current execution
	statusCallback func(string)
	resultCallback func(string)
	runner         *ScriptRunner
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
	w := &Window{
		FyneWindow:     window,
		enabledModules: make(map[string]bool),
		appName:        "fynerisor", // default
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
		"import":    newImportBuiltin(),
		"go":        newGoBuiltin(),
	}

	w.globals = []risor.Option{risor.WithEnv(risor.Builtins()), risor.WithEnv(globals)}

	// Apply module options (which add to w.globals)
	for _, opt := range opts {
		if _, ok := opt.(moduleOption); ok {
			opt.applyToWindow(w)
		}
	}

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
			w.functionCalls <- func() { fyne.Do(func() { w.onDropped(paths) }) }
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
			_, err := callFunc(ctx, fn, []object.Object{})
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

// ImportScript loads an additional script to be prepended to the main script.
// The imported script's code will be executed before the main script.
// Multiple imports are executed in the order they are added.
// Can load from local files or HTTP(S) URLs.
func (w *Window) ImportScript(source string) error {
	return w.runner.ImportScript(source)
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

			fyne.Do(func() {
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
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			w.onNewStdout = func(text string) {
				callFunc(ctx, fn, []object.Object{object.NewString(text)})
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
				callFunc(ctx, fn, []object.Object{object.NewStringList(droppedPaths)})
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
				fyne.Do(func() {
					_, err := callFunc(ctx, fn, []object.Object{})
					if err != nil {
						w.SetStatus("ERROR: " + err.Error())
					}
				})
			}

			return object.Nil, nil
		}), true

	case "Canvas":
		return w, true  // Window itself implements Canvas() method

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
			fyne.Do(func() {
				w.Resize(float32(width), float32(height))
			})

			return object.Nil, nil
		}), true
	}
	return nil, false
}

// Canvas returns the Fyne canvas for this window.
// This allows scripts to create PopUp widgets.
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

	// Channel to receive output
	outputChan := make(chan string)

	// Goroutine to read from the pipe
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			// Send to channel
			outputChan <- line
			// Also write to original stdout
			// fmt.Fprintln(originalStdout, line)
		}
		fmt.Println("Channel closing")
		close(outputChan)
	}()

	// Another goroutine to consume from the channel
	go func() {
		for msg := range outputChan {
			w.functionCalls <- func() { fyne.Do(func() { w.onNewStdout(msg) }) }
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

