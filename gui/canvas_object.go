package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// CanvasObject hosts a fynerisor script inside a normal Fyne application.
//
// It is a regular fyne.CanvasObject, so it can be placed anywhere a widget can:
// in a container, a border slot, a tab, a dialog, etc. The script's
// window.SetContent(...) call renders inside this object rather than replacing
// the host window's content.
//
// Use WithHostWindow to give the script access to window-level features
// (dialogs, shortcuts, menus, window.Canvas and window.Close). Without it the
// object is fully self-contained but those features return script errors.
//
// Example:
//
//	a := app.New()
//	host := a.NewWindow("My App")
//
//	view := gui.NewCanvasObject(`
//	    let count = 0
//	    let label = widget.NewLabel("Not clicked")
//	    let button = widget.NewButton("Click me", () => {
//	        count = count + 1
//	        label.SetText(sprintf("Clicked %d times", count))
//	    })
//	    window.SetContent(container.NewVBox(label, button))
//	`, gui.WithHostWindow(host))
//	view.Execute()
//
//	host.SetContent(container.NewBorder(
//	    widget.NewLabel("Native Fyne header"),
//	    nil, nil, nil,
//	    view,
//	))
//	host.ShowAndRun()
//
// The script runs asynchronously on its own goroutine, just like a top-level
// fynerisor window. Call Stop when the object is no longer needed to cancel
// the script and its callback loop.
type CanvasObject struct {
	widget.BaseWidget
	window *Window
}

// Compile-time checks: an embedded fynerisor view is both a Fyne widget and a
// canvas object.
var (
	_ fyne.Widget       = (*CanvasObject)(nil)
	_ fyne.CanvasObject = (*CanvasObject)(nil)
)

// NewCanvasObject creates an embedded fynerisor view and loads script into it.
//
// The script is not executed until Execute is called. An empty script is
// allowed, so callers can create the object first and call LoadScript later.
//
// options are the same ones accepted by NewWindow (WithHTTP, WithGlobal,
// WithStatusCallback, ...). Use WithHostWindow to enable dialogs and other
// window-level APIs.
func NewCanvasObject(script string, opts ...Option) *CanvasObject {
	w := newWindow(opts...)
	if w.FyneWindow != nil {
		w.installDropHandler()
	}
	c := &CanvasObject{window: w}
	c.ExtendBaseWidget(c)
	if script != "" {
		w.LoadScript(script)
	}
	return c
}

// CreateRenderer implements fyne.Widget. It renders the script's content
// container, which window.SetContent updates.
func (c *CanvasObject) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.window.content)
}

// Window returns the fynerisor runtime backing this object. Use it for
// features not exposed directly on CanvasObject, such as RegisterGlobal.
func (c *CanvasObject) Window() *Window {
	return c.window
}

// LoadScript loads a script that will be executed by the next Execute call.
func (c *CanvasObject) LoadScript(script string) {
	c.window.LoadScript(script)
}

// CompileScript parses and compiles the loaded script without executing it.
func (c *CanvasObject) CompileScript() error {
	return c.window.CompileScript()
}

// Execute runs the loaded script asynchronously. Previous execution is
// cancelled, matching Window.Execute.
func (c *CanvasObject) Execute() {
	c.window.Execute()
}

// Stop cancels the running script and its callback loop.
func (c *CanvasObject) Stop() {
	c.window.Stop()
}

// Clear removes the loaded script from the object without stopping a script
// that is already executing.
func (c *CanvasObject) Clear() {
	c.window.Clear()
}

// Status returns the current execution status ("Ready!", "ERROR: ...", or ""
// before the first execution).
func (c *CanvasObject) Status() string {
	return c.window.Status
}
