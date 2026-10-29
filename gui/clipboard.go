package gui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

var _ object.Object = &Clipboard{}

const ClipboardType object.Type = "clipboard"

// Clipboard exposes the system clipboard to scripts.
// Available in Risor scripts as the global 'clipboard' object.
//
// Example:
//
//	clipboard.Copy("some text")   // copy to clipboard
//	let s = clipboard.Read()      // read current clipboard text
type Clipboard struct{}

func (obj *Clipboard) Type() object.Type {
	return ClipboardType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Clipboard) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Clipboard) inspectUI() string {
	return "clipboard"
}

func (obj *Clipboard) Interface() interface{} {
	return nil
}

func (obj *Clipboard) IsTruthy() bool {
	return true
}

func (obj *Clipboard) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ClipboardType, opType)
}

func (obj *Clipboard) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Clipboard) Attrs() []object.AttrSpec {
	return clipboardAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Clipboard) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Clipboard) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ClipboardType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Clipboard) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var clipboardAttrs = objattrs.Register(string(ClipboardType), object.NewAttrRegistry[*Clipboard](string(ClipboardType)))

func (obj *Clipboard) getAttrUI(name string) (object.Object, bool) {
	return clipboardAttrs.GetAttr(obj, name)
}

func NewFyneClipboard() *Clipboard {
	return &Clipboard{}
}
func init() {
	clipboardAttrs.
		Define(
			"Copy").Doc("Copy text to the clipboard.").Args("arg1").Returns("nil").Impl(func(obj *Clipboard, ctx context.Context, args ...object.Object) (object.Object, error) {
		s, err := object.AsString(args[0])
		if err !=

			nil {
			return nil, err
		}
		guithread.
			DoAndWait(func() {
				fyne.CurrentApp().Clipboard().SetContent(s)
			})
		return object.Nil, nil
	})
	clipboardAttrs.
		Define(
			"Read").Doc("Read the clipboard contents. DoAndWait: the script body runs off the GUI thread and needs the value back (issue #18).").Args().Returns("string").Impl(func(obj *Clipboard, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		var content string
		guithread.DoAndWait(func() {
			content = fyne.CurrentApp().Clipboard().
				Content()
		})
		return object.NewString(content),
			nil
	})
}
