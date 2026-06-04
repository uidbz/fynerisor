package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &FileDialog{}

const FileDialogType object.Type = "dialog.FileDialog"

// FileDialog wraps dialog.FileDialog for Risor scripting
type FileDialog struct {
	instance *dialog.FileDialog
	window   fyne.Window
}

func (obj *FileDialog) Type() object.Type {
	return FileDialogType
}

func (obj *FileDialog) Inspect() string {
	return "dialog.FileDialog"
}

func (obj *FileDialog) Interface() interface{} {
	return obj.instance
}

func (obj *FileDialog) IsTruthy() bool {
	return true
}

func (obj *FileDialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'dialog.FileDialog'")
}

func (obj *FileDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", FileDialogType, opType), nil
}

func (obj *FileDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FileDialog) Attrs() []object.AttrSpec {
	return nil
}

func (obj *FileDialog) SetAttr(name string, value object.Object) error {
	switch name {
	case "FileName":
		filename, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetFileName(filename)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", FileDialogType, name)
}

func (obj *FileDialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("dialog.FileDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.FileDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.FileDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetFileName":
		return object.NewBuiltin("dialog.FileDialog.SetFileName", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			filename, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetFileName(filename)
			})
			return object.Nil, nil
		}), true
	case "SetFilter":
		return object.NewBuiltin("dialog.FileDialog.SetFilter", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			filter, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			// Create a simple extension filter
			fyne.Do(func() {
				obj.instance.SetFilter(storage.NewExtensionFileFilter([]string{filter}))
			})
			return object.Nil, nil
		}), true
	case "SetLocation":
		return object.NewBuiltin("dialog.FileDialog.SetLocation", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				uri := storage.NewFileURI(path)
				listableURI, _ := storage.ListerForURI(uri)
				if listableURI != nil {
					obj.instance.SetLocation(listableURI)
				}
			})
			return object.Nil, nil
		}), true
	case "SetConfirmText":
		return object.NewBuiltin("dialog.FileDialog.SetConfirmText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetConfirmText(text)
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.FileDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetDismissText(text)
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewFileOpen(callback func(fyne.URIReadCloser, error), parent fyne.Window) *FileDialog {
	return &FileDialog{
		instance: dialog.NewFileOpen(callback, parent),
		window:   parent,
	}
}

func NewFileSave(callback func(fyne.URIWriteCloser, error), parent fyne.Window) *FileDialog {
	return &FileDialog{
		instance: dialog.NewFileSave(callback, parent),
		window:   parent,
	}
}

func NewFolderOpen(callback func(fyne.ListableURI, error), parent fyne.Window) *FileDialog {
	return &FileDialog{
		instance: dialog.NewFolderOpen(callback, parent),
		window:   parent,
	}
}
