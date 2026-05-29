package gui

import (
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const AppType object.Type = "app"

// App provides information about the embedding application.
// Available in Risor scripts as the global 'app' object.
type App struct {
	appName string
}

func newAppObject(w *Window) *App {
	return &App{appName: w.appName}
}

// newAppObjectForContext removed - not needed in GUI package

func (a *App) Type() object.Type {
	return AppType
}

func (a *App) Inspect() string {
	return "app"
}

func (a *App) String() string {
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

func (a *App) Cost() int {
	return 0
}

func (a *App) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "name":
		return object.NewString(a.appName), true
	case "version":
		// Return app version if set, otherwise empty string
		if appVersion != "" {
			return object.NewString(appVersion), true
		}
		return object.NewString(""), true
	}
	return nil, false
}

func (a *App) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: app object attributes are read-only")
}

func (a *App) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "name", Doc: "Name of the embedding application"},
		{Name: "version", Doc: "Version of the embedding application (if set via SetAppVersion)"},
	}
}
