package gui

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorDialog "github.com/uidbz/fynerisor/gui/dialog"
	"github.com/uidbz/fynerisor/gui/vmexec"
)

const DialogType object.Type = "dialog"

// Dialog is a factory object that provides dialog functions to Risor scripts.
// It exposes methods for showing information, errors, confirmations, file pickers,
// color pickers, forms, and custom dialogs.
//
// Available in Risor scripts as the global 'dialog' object.
//
// Example usage in Risor:
//
//	dialog.ShowInformation("Success", "Operation completed")
//	dialog.ShowConfirm("Delete?", "Are you sure?", (confirmed) => {
//	    if (confirmed) { print("Deleted") }
//	})
//	dialog.ShowFileOpen((path, err) => {
//	    if (path != nil) { print("Selected:", path) }
//	})
type Dialog struct {
	w *Window
}

// fileDialogResult converts a file open/save dialog result into the
// (path, error) arguments passed to the script callback. The file handle is
// closed here: scripts only receive the path.
func fileDialogResult(f interface {
	URI() fyne.URI
	Close() error
}, err error) (object.Object, object.Object) {
	var path, errObj object.Object = object.Nil, object.Nil
	if f != nil {
		path = object.NewString(f.URI().Path())
		f.Close()
	}
	if err != nil {
		errObj = object.NewString(err.Error())
	}
	return path, errObj
}

// folderDialogResult converts a folder dialog result into the (path, error)
// arguments passed to the script callback.
func folderDialogResult(uri fyne.ListableURI, err error) (object.Object, object.Object) {
	var path, errObj object.Object = object.Nil, object.Nil
	if uri != nil {
		path = object.NewString(uri.Path())
	}
	if err != nil {
		errObj = object.NewString(err.Error())
	}
	return path, errObj
}

func NewDialog(w *Window) *Dialog {
	return &Dialog{w: w}
}

func (d *Dialog) Type() object.Type {
	return DialogType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (d *Dialog) Inspect() string {
	return guithread.Inspect(d.inspectUI)
}

func (d *Dialog) inspectUI() string {
	return "dialog"
}

func (d *Dialog) Interface() interface{} {
	return nil
}

func (d *Dialog) IsTruthy() bool {
	return true
}

func (d *Dialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", DialogType, opType)
}

func (d *Dialog) Equals(other object.Object) bool {
	return d == other
}

func (d *Dialog) Attrs() []object.AttrSpec {
	return dialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (d *Dialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, d.setAttrUI)
}

func (d *Dialog) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", DialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (d *Dialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, d.getAttrUI)
}

var dialogAttrs = objattrs.Register(string(DialogType), object.NewAttrRegistry[*Dialog](string(DialogType)))

func (d *Dialog) getAttrUI(name string) (object.Object, bool) {
	// Every dialog function needs a Fyne window to attach to. An embedded
	// CanvasObject only has one when the host passes WithHostWindow. Return a
	// clear script error instead of dereferencing a nil window.
	switch name {
	case "ShowInformation", "ShowError", "ShowConfirm",
		"ShowFileOpen", "ShowFileSave", "ShowFolderOpen",
		"ShowCustom", "ShowCustomConfirm", "ShowColorPicker", "ShowForm",
		"NewFileOpen", "NewFileSave", "NewFolderOpen",
		"NewConfirm", "NewCustomConfirm", "NewCustom",
		"NewColorPicker", "NewForm":
		if d.w.FyneWindow == nil {
			return object.NewBuiltin("dialog."+name, func(context.Context, ...object.Object) (object.Object, error) {
				return nil, object.Errorf("dialog.%s: no host window (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)", name)
			}), true
		}
	}
	return dialogAttrs.GetAttr(d, name)

	// Content must be a CanvasObject

	// Content must be a CanvasObject

	// Convert color to hex string for easier use in scripts

	// RGBA returns values in range 0-65535, convert to 0-255

	// Items must be a list of FormItems

}
func init() {
	dialogAttrs.
		Define("ShowInformation").Variadic("args").
		Returns("nil").Impl(
		func(d *Dialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
			if len(args) !=
				2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: title must be string, got %s",

					args[0].
						Type())
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return nil,
					object.Errorf("type error: message must be string, got %s", args[1].Type())
			}
			guithread.
				Do(func() {
					dialog.ShowInformation(title,

						message, d.w.FyneWindow)
				})
			return object.Nil, nil
		})
	dialogAttrs.
		Define("ShowError").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		message,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: message must be string, got %s",

				args[0].Type())
		}
		guithread.Do(func() {
			dialog.ShowError(fmt.Errorf("%s", message), d.w.
				FyneWindow)
		})
		return object.Nil, nil
	})
	dialogAttrs.
		Define("ShowConfirm").
		Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=

			3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type())
		}
		message, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: message must be string, got %s", args[1].Type())
		}
		cb, err := vmexec.NewCallback(ctx, "dialog.ShowConfirm",

			args[2])
		if err != nil {
			return nil, err
		}
		callback := func(confirmed bool) {
			cb.Fire(object.
				NewBool(confirmed))
		}
		guithread.Do(func() {
			dialog.ShowConfirm(title, message, callback, d.w.FyneWindow)
		})
		return object.
			Nil, nil
	})
	dialogAttrs.
		Define("ShowFileOpen").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=

			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.ShowFileOpen",
			args[0])
		if err != nil {
			return nil,

				err
		}
		callback := func(reader fyne.URIReadCloser, err error) {
			cb.Fire(fileDialogResult(reader, err))
		}
		guithread.Do(func() {
			dialog.ShowFileOpen(callback, d.w.FyneWindow)
		})
		return object.Nil, nil
	})
	dialogAttrs.
		Define("ShowFileSave").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=

			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.ShowFileSave",
			args[0])
		if err != nil {
			return nil,

				err
		}
		callback := func(writer fyne.URIWriteCloser, err error) {
			cb.Fire(fileDialogResult(writer, err))
		}
		guithread.Do(
			func() { dialog.ShowFileSave(callback, d.w.FyneWindow) })
		return object.Nil, nil
	},
	)
	dialogAttrs.
		Define("ShowFolderOpen").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) !=
			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb, err := vmexec.NewCallback(ctx, "dialog.ShowFolderOpen",
			args[0])
		if err != nil {
			return nil,

				err
		}
		callback := func(uri fyne.ListableURI, err error) {
			cb.Fire(folderDialogResult(uri, err))
		}
		guithread.Do(func() {
			dialog.ShowFolderOpen(callback, d.w.FyneWindow)
		})
		return object.Nil, nil
	})
	dialogAttrs.
		Define("ShowCustom").
		Variadic("args").Returns("nil").Impl(func(
		d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=
			3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type(),
			)
		}
		dismiss, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: dismiss must be string, got %s", args[1].Type())
		}
		content,
			ok := args[2].(interface {
			CanvasObject() fyne.
				CanvasObject
		})
		if !ok {
			return nil, object.Errorf("type error: content must be a widget or canvas object")
		}
		guithread.Do(func() {
			dialog.ShowCustom(title,
				dismiss, content.CanvasObject(), d.w.FyneWindow)
		})
		return object.
			Nil, nil
	})
	dialogAttrs.
		Define("ShowCustomConfirm").
		Variadic("args").
		Returns("nil").Impl(func(d *Dialog, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 5 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=5",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].
					Type())
		}
		confirm, err := object.AsString(args[1])
		if err != nil {
			return nil,
				object.Errorf("type error: confirm must be string, got %s", args[1].Type())
		}
		dismiss,
			err := object.AsString(args[2])
		if err !=
			nil {
			return nil, object.Errorf("type error: dismiss must be string, got %s",

				args[2].Type())
		}
		content, ok := args[3].(interface {
			CanvasObject() fyne.
				CanvasObject
		})
		if !ok {
			return nil, object.Errorf("type error: content must be a widget or canvas object")
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.ShowCustomConfirm",
			args[4])
		if err != nil {
			return nil,
				err
		}
		callback :=
			func(confirmed bool) {
				cb.Fire(object.NewBool(confirmed))
			}
		guithread.Do(func() {
			dialog.ShowCustomConfirm(title, confirm, dismiss, content.CanvasObject(), callback,
				d.w.FyneWindow)
		})
		return object.Nil,
			nil
	})
	dialogAttrs.
		Define("ShowColorPicker").Variadic("args").
		Returns("nil").Impl(
		func(d *Dialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
			if len(args) !=
				3 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
					len(args))
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: title must be string, got %s",

					args[0].
						Type())
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return nil,
					object.Errorf("type error: message must be string, got %s", args[1].Type())
			}
			cb,
				err := vmexec.NewCallback(ctx, "dialog.ShowColorPicker",

				args[2])
			if err != nil {
				return nil, err
			}
			callback := func(c color.Color) {
				r,
					g, b, a := c.RGBA()
				colorMap := object.NewMap(map[string]object.
					Object{"R": object.NewInt(int64(r >> 8)), "G": object.NewInt(int64(g >> 8)), "B": object.NewInt(int64(b >>
					8)),
					"A": object.NewInt(int64(a >> 8))})
				cb.
					Fire(colorMap)
			}
			guithread.Do(func() {
				dialog.ShowColorPicker(title, message,
					callback,
					d.
						w.
						FyneWindow)
			})
			return object.Nil, nil
		})
	dialogAttrs.
		Define("ShowForm").Variadic(
		"args").Returns(
		"nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 5 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=5",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type())

		}
		confirm, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: confirm must be string, got %s", args[1].Type())
		}
		dismiss, err := object.AsString(args[2])
		if err != nil {
			return nil, object.Errorf("type error: dismiss must be string, got %s",

				args[2].Type())
		}
		itemsList,
			ok := args[3].(*object.List)
		if !ok {
			return nil,
				object.Errorf("type error: items must be a list, got %s", args[3].
					Type())
		}
		var formItems []*widget.FormItem
		for i, item := range itemsList.Value() {
			formItemObj,
				ok := item.(interface {
				Interface() interface{}
			})
			if !ok {
				return nil, object.
					Errorf("type error: item %d is not a valid form item", i)
			}
			formItem,
				ok := formItemObj.Interface().(*widget.FormItem)
			if !ok {
				return nil, object.
					Errorf("type error: item %d is not a form item",
						i)
			}
			formItems = append(formItems,
				formItem)
		}
		cb, err :=
			vmexec.NewCallback(ctx, "dialog.ShowForm", args[4])
		if err != nil {
			return nil, err
		}
		callback := func(confirmed bool) {
			cb.Fire(object.NewBool(confirmed))
		}
		guithread.
			Do(
				func() {
					dialog.
						ShowForm(title, confirm, dismiss, formItems,

							callback, d.w.FyneWindow,
						)
				})
		return object.Nil, nil
	})
	dialogAttrs.
		Define("NewFileOpen").
		Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=

			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.NewFileOpen",
			args[0])
		if err != nil {
			return nil,
				err

		}
		callback := func(reader fyne.URIReadCloser, err error) {
			cb.Fire(fileDialogResult(reader, err))
		}
		fileDialog := risorDialog.
			NewFileOpen(callback, d.w.FyneWindow)
		return fileDialog, nil
	})
	dialogAttrs.
		Define("NewFileSave").
		Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=

			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.NewFileSave",
			args[0])
		if err != nil {
			return nil,
				err

		}
		callback := func(writer fyne.URIWriteCloser, err error) {
			cb.Fire(fileDialogResult(writer, err))
		}
		fileDialog := risorDialog.
			NewFileSave(callback, d.w.FyneWindow)
		return fileDialog, nil
	})
	dialogAttrs.
		Define("NewFolderOpen").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) !=
			1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		cb, err := vmexec.NewCallback(ctx, "dialog.NewFolderOpen",
			args[0])
		if err != nil {
			return nil,

				err
		}
		callback := func(uri fyne.ListableURI, err error) {
			cb.Fire(folderDialogResult(uri, err))
		}
		fileDialog := risorDialog.
			NewFolderOpen(callback, d.w.FyneWindow)
		return fileDialog, nil
	})
	dialogAttrs.
		Define("NewConfirm").
		Variadic("args").Returns("nil").Impl(func(
		d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) !=
			3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type(),
			)
		}
		message, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: message must be string, got %s", args[1].Type())
		}
		cb, err := vmexec.NewCallback(ctx, "dialog.NewConfirm",

			args[2])
		if err != nil {
			return nil, err
		}
		callback := func(confirmed bool) {
			cb.Fire(object.
				NewBool(confirmed))
		}
		confirmDialog := risorDialog.NewConfirm(title, message, callback, d.w.FyneWindow)
		return confirmDialog,
			nil
	})
	dialogAttrs.
		Define("NewCustomConfirm").Variadic("args").
		Returns("nil").Impl(func(d *Dialog, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 5 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=5",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].
					Type())
		}
		confirm, err := object.AsString(args[1])
		if err != nil {
			return nil,
				object.Errorf("type error: confirm must be string, got %s", args[1].Type())
		}
		dismiss,
			err := object.AsString(args[2])
		if err !=
			nil {
			return nil, object.Errorf("type error: dismiss must be string, got %s",

				args[2].Type())
		}
		content, ok := args[3].(interface {
			CanvasObject() fyne.
				CanvasObject
		})
		if !ok {
			return nil, object.Errorf("type error: content must be a widget or canvas object")
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.NewCustomConfirm",
			args[4])
		if err != nil {
			return nil,
				err
		}
		callback := func(confirmed bool) {
			cb.Fire(object.NewBool(confirmed))
		}
		confirmDialog := risorDialog.
			NewCustomConfirm(title, confirm, dismiss, content.CanvasObject(), callback,
				d.w.FyneWindow)
		return confirmDialog, nil
	})
	dialogAttrs.
		Define("NewCustom").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type(),
			)
		}
		dismiss, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: dismiss must be string, got %s", args[1].Type())
		}
		content,
			ok := args[2].(interface {
			CanvasObject() fyne.
				CanvasObject
		})
		if !ok {
			return nil, object.Errorf("type error: content must be a widget or canvas object")
		}
		customDialog := risorDialog.NewCustom(title,
			dismiss, content.CanvasObject(), d.w.FyneWindow)
		return customDialog,
			nil
	})
	dialogAttrs.
		Define("NewColorPicker").Variadic("args").Returns("nil").Impl(func(d *Dialog, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) !=
			3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type())
		}
		message, err := object.AsString(args[1])
		if err != nil {
			return nil,
				object.Errorf("type error: message must be string, got %s", args[1].Type())
		}
		cb,
			err := vmexec.NewCallback(ctx, "dialog.NewColorPicker",

			args[2])
		if err != nil {
			return nil, err
		}
		callback :=
			func(c color.Color) {
				r,
					g,
					b, a := c.RGBA()
				colorMap := object.NewMap(map[string]object.
					Object{"R": object.NewInt(int64(r >> 8)), "G": object.NewInt(int64(g >> 8)), "B": object.NewInt(int64(b >>
					8)), "A": object.NewInt(int64(a >> 8))})
				cb.Fire(colorMap)
			}
		colorDialog := risorDialog.NewColorPicker(
			title, message, callback,
			d.w.FyneWindow,
		)
		return colorDialog, nil
	})
	dialogAttrs.
		Define("NewForm").Variadic("args").Returns(
		"nil",
	).Impl(func(d *Dialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 5 {

			return nil, object.Errorf("wrong number of arguments. got=%d, want=5",
				len(args))
		}
		title,
			err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: title must be string, got %s",

				args[0].Type())
		}
		confirm, err := object.AsString(args[1])
		if err != nil {
			return nil, object.
				Errorf("type error: confirm must be string, got %s", args[1].Type())
		}
		dismiss, err := object.AsString(args[2])
		if err != nil {
			return nil, object.Errorf("type error: dismiss must be string, got %s",

				args[2].Type())
		}
		itemsList,
			ok := args[3].(*object.List)
		if !ok {
			return nil,
				object.Errorf("type error: items must be a list, got %s", args[3].Type())
		}
		var formItems []*widget.FormItem
		for i, item := range itemsList.Value() {
			formItemObj,
				ok := item.(interface {
				Interface() interface{}
			})
			if !ok {
				return nil,
					object.
						Errorf("type error: item %d is not a valid form item", i)
			}
			formItem,
				ok := formItemObj.Interface().(*widget.FormItem)
			if !ok {
				return nil, object.
					Errorf("type error: item %d is not a form item",
						i)
			}
			formItems = append(formItems,
				formItem)
		}
		cb, err := vmexec.
			NewCallback(ctx, "dialog.NewForm", args[4])
		if err != nil {
			return nil, err
		}
		callback := func(
			confirmed bool) {
			cb.Fire(object.NewBool(confirmed))
		}
		formDialog :=
			risorDialog.
				NewForm(title, confirm, dismiss, formItems, callback,

					d.w.FyneWindow)
		return formDialog, nil
	})
}
