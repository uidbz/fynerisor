package gui

import (
	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"github.com/uidbz/fynerisor/core"
)

// newRequireBuiltin creates the require() function for the Window. The
// engine lives in core and is shared with headless Contexts (issue #32);
// only the binding differs: @gui is available here, module enablement is
// queried live (globals registered after construction count), and the
// fix-it hints name gui's options.
func newRequireBuiltin(w *Window) *object.Builtin {
	return core.NewRequireBuiltin(core.RequireEnv{
		PkgName:       "gui",
		GUIAvailable:  true,
		AppVersion:    func() string { return w.appVersion },
		ModuleEnabled: func(name string) bool { return w.moduleEnabled(name) },
	})
}
