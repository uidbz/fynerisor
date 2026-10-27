package guithread

import (
	"context"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// The script API that fynerisor exposes for Fyne objects (widgets,
// containers, canvas objects, dialogs, menus, the window, ...) runs on
// Fyne's UI thread (issue #49). Script code itself runs on the run's VM
// goroutine, so without this, reading entry.Text while the user types — or
// any other attribute read/write — raced with the UI thread.
//
// Each wrapper type routes its object.Object methods through these helpers:
//
//	func (obj *Entry) GetAttr(name string) (object.Object, bool) { return guithread.GetAttr(name, obj.getAttrUI) }
//	func (obj *Entry) SetAttr(name string, v object.Object) error { return guithread.SetAttr(name, v, obj.setAttrUI) }
//	func (obj *Entry) Inspect() string                             { return guithread.Inspect(obj.inspectUI) }
//
// GetAttr also wraps returned builtins (methods) so that calling them runs
// on the UI thread. Builtins that do slow I/O (network, files) and must not
// block the UI are marked with OffUI and handle their own marshalling.
//
// Deadlock freedom: the UI thread never waits for a VM goroutine (it only
// enqueues events or try-locks the VM, see gui/vmexec), so a VM goroutine
// waiting for the UI thread always makes progress. On the UI thread these
// helpers run inline.

// offUI marks a builtin that must run on the caller's goroutine.
type offUI struct{ *object.Builtin }

// OffUI marks b to run on the script's VM goroutine instead of the UI
// thread. Use it for builtins that block on I/O; they must touch Fyne
// objects only through Do / DoAndWaitContext.
func OffUI(b *object.Builtin) object.Object { return offUI{b} }

// GetAttr runs get on the UI thread and wraps a returned builtin so that
// calling it also runs on the UI thread.
func GetAttr(name string, get func(string) (object.Object, bool)) (object.Object, bool) {
	var (
		o  object.Object
		ok bool
	)
	DoAndWait(func() { o, ok = get(name) })
	return OnUI(o), ok
}

// SetAttr runs set on the UI thread.
func SetAttr(name string, value object.Object, set func(string, object.Object) error) error {
	var err error
	DoAndWait(func() { err = set(name, value) })
	return err
}

// Inspect runs inspect on the UI thread.
func Inspect(inspect func() string) string {
	var s string
	DoAndWait(func() { s = inspect() })
	return s
}

// OnUI returns o with builtins re-targeted to the UI thread: calling the
// returned builtin runs the original on the UI thread (inline when already
// there) and gives up if the calling script run is stopped meanwhile.
// Builtins marked with OffUI are returned unwrapped; other objects are
// returned as they are.
func OnUI(o object.Object) object.Object {
	switch b := o.(type) {
	case offUI:
		return b.Builtin
	case *object.Builtin:
		return object.NewBuiltin(b.Name(), func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if IsMain() {
				return b.Call(ctx, args...)
			}
			var (
				res object.Object
				err error
			)
			if waitErr := DoAndWaitContext(ctx, func() { res, err = b.Call(ctx, args...) }); waitErr != nil {
				return nil, waitErr
			}
			return res, err
		})
	}
	return o
}
