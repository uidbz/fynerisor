package gui

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
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

func (obj *Clipboard) Inspect() string {
	return "clipboard"
}

func (obj *Clipboard) Interface() interface{} {
	return nil
}

func (obj *Clipboard) IsTruthy() bool {
	return true
}

func (obj *Clipboard) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'clipboard'")
}

func (obj *Clipboard) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ClipboardType))
	return object.Errorf("eval error: unsupported operation for %s: %v", ClipboardType, opType), err
}

func (obj *Clipboard) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Clipboard) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Clipboard) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ClipboardType, name)
}

func (obj *Clipboard) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Copy":
		return object.NewBuiltin("clipboard.Copy", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			s, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				fyne.CurrentApp().Clipboard().SetContent(s)
			})
			return object.Nil, nil
		}), true
	case "Read":
		return object.NewBuiltin("clipboard.Read", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			var content string
			guithread.Do(func() {
				content = fyne.CurrentApp().Clipboard().Content()
			})
			return object.NewString(content), nil
		}), true
	}
	return nil, false
}

func NewFyneClipboard() *Clipboard {
	return &Clipboard{}
}
