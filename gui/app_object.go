package gui

import (
	"context"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

const AppType object.Type = "app"

// App provides information about the embedding application.
// Available in Risor scripts as the global 'app' object.
type App struct {
	appName string
	win     *Window
}

func newAppObject(w *Window) *App {
	return &App{appName: w.appName, win: w}
}

func (a *App) Type() object.Type {
	return AppType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (a *App) Inspect() string {
	return guithread.Inspect(a.inspectUI)
}

func (a *App) inspectUI() string {
	return "app"
}

func (a *App) Interface() interface{} {
	return a
}

func (a *App) Equals(other object.Object) bool {
	_, ok := other.(*App)
	return ok
}

func (a *App) IsTruthy() bool {
	return true
}

func (a *App) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for app: %v", opType)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (a *App) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, a.getAttrUI)
}

var appAttrs = objattrs.Register(string(AppType), object.NewAttrRegistry[*App](string(AppType)))

func (a *App) getAttrUI(name string) (object.Object, bool) {
	return appAttrs.GetAttr(a, name)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (a *App) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, a.setAttrUI)
}

func (a *App) setAttrUI(name string, value object.Object) error {
	return object.Errorf("type error: app object attributes are read-only")
}

func (a *App) Attrs() []object.AttrSpec {
	return appAttrs.Specs()
}
func init() {
	appAttrs.Define("name").Doc("Name of the embedding application").Getter(func(a *App) object.Object {
		return object.NewString(a.appName)
	})
	// This window's app version (issue #44), empty if unset.
	appAttrs.Define("version").Doc("Version of the embedding application (if set via SetAppVersion)").Getter(func(a *App) object.Object {
		return object.NewString(a.win.appVersion)
	})
	appAttrs.Define("Quit").Doc("Quit the application (app.Quit())").Args().Returns("nil").Impl(func(a *App, ctx context.Context, args ...object.Object) (object.Object, error) {
		if a.win == nil || a.win.FyneApp == nil {
			return nil, object.Errorf("runtime error: app.Quit unavailable (no Fyne app; host owns the window)")
		}
		guithread.Do(func() { a.win.FyneApp.Quit() })
		return object.Nil,
			nil
	})
}
