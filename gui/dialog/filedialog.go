package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *FileDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *FileDialog) inspectUI() string {
	return "dialog.FileDialog"
}

func (obj *FileDialog) Interface() interface{} {
	return obj.instance
}

func (obj *FileDialog) IsTruthy() bool {
	return true
}

func (obj *FileDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FileDialogType, opType)
}

func (obj *FileDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FileDialog) Attrs() []object.AttrSpec {
	return fileDialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *FileDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *FileDialog) setAttrUI(name string, value object.Object) error {
	switch name {
	case "FileName":
		filename, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetFileName(filename)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", FileDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *FileDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var fileDialogAttrs = objattrs.Register(string(FileDialogType), object.NewAttrRegistry[*FileDialog](string(FileDialogType)))

func (obj *FileDialog) getAttrUI(name string) (object.Object, bool) {
	// Create a simple extension filter
	return fileDialogAttrs.GetAttr(obj, name)
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
func init() {
	fileDialogAttrs.
		Define("Show").Args().Returns("nil").Impl(func(
		obj *FileDialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
		guithread.Do(func() { obj.instance.Show() })
		return object.
			Nil, nil
	})
	fileDialogAttrs.
		Define("Hide").Args().Returns("nil").Impl(func(
		obj *FileDialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
		guithread.Do(func() { obj.instance.Hide() })
		return object.
			Nil, nil
	})
	fileDialogAttrs.
		Define("Refresh").Args().Returns("nil").Impl(func(obj *FileDialog, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Refresh() })
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetFileName").Args(
		"arg1",
	).Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		filename, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.SetFileName(filename)
		})
		return object.
			Nil, nil
	})
	fileDialogAttrs.
		Define("SetFilter").
		Args("arg1").Returns("nil").
		Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
			filter,
				err := object.AsString(args[0])
			if err !=
				nil {
				return nil, object.Errorf("type error: %v",

					err)
			}
			guithread.Do(func() {

				obj.instance.SetFilter(storage.NewExtensionFileFilter([]string{filter}))
			})
			return object.Nil, nil
		})
	fileDialogAttrs.
		Define("SetLocation").Args(
		"arg1",
	).Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		path, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {

			uri := storage.NewFileURI(path)
			listableURI, _ := storage.
				ListerForURI(uri)
			if listableURI != nil {
				obj.instance.
					SetLocation(listableURI)
			}
		})
		return object.Nil,
			nil
	})
	fileDialogAttrs.
		Define("SetConfirmText").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() { obj.instance.SetConfirmText(text) })
		return object.
				Nil,
			nil
	})
	fileDialogAttrs.
		Define("SetDismissText").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() { obj.instance.SetDismissText(text) })
		return object.
				Nil,
			nil
	})
}
