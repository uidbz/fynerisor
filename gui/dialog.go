package gui

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorDialog "github.com/uidbz/fynerisor/gui/dialog"
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

func NewDialog(w *Window) *Dialog {
	return &Dialog{w: w}
}

func (d *Dialog) Type() object.Type {
	return DialogType
}

func (d *Dialog) Inspect() string {
	return "dialog"
}

func (d *Dialog) Interface() interface{} {
	return nil
}

func (d *Dialog) IsTruthy() bool {
	return true
}

func (d *Dialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'Dialog'")
}

func (d *Dialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", DialogType, opType), nil
}

func (d *Dialog) Equals(other object.Object) bool {
	return d == other
}

func (d *Dialog) Attrs() []object.AttrSpec {
	return nil
}

func (d *Dialog) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", DialogType, name)
}

func (d *Dialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "ShowInformation":
		return object.NewBuiltin("dialog.ShowInformation", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[1].Type()), nil
			}
			fyne.Do(func() {
				dialog.ShowInformation(title, message, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowError":
		return object.NewBuiltin("dialog.ShowError", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			message, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[0].Type()), nil
			}
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("%s", message), d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowConfirm":
		return object.NewBuiltin("dialog.ShowConfirm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[1].Type()), nil
			}
			fn, ok := args[2].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[2].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			fyne.Do(func() {
				dialog.ShowConfirm(title, message, callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowFileOpen":
		return object.NewBuiltin("dialog.ShowFileOpen", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(reader fyne.URIReadCloser, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if reader != nil {
						path = object.NewString(reader.URI().Path())
						defer reader.Close()
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fyne.Do(func() {
				dialog.ShowFileOpen(callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowFileSave":
		return object.NewBuiltin("dialog.ShowFileSave", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(writer fyne.URIWriteCloser, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if writer != nil {
						path = object.NewString(writer.URI().Path())
						defer writer.Close()
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fyne.Do(func() {
				dialog.ShowFileSave(callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowFolderOpen":
		return object.NewBuiltin("dialog.ShowFolderOpen", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(uri fyne.ListableURI, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if uri != nil {
						path = object.NewString(uri.Path())
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fyne.Do(func() {
				dialog.ShowFolderOpen(callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowCustom":
		return object.NewBuiltin("dialog.ShowCustom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			dismiss, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[1].Type()), nil
			}

			// Content must be a CanvasObject
			content, ok := args[2].(interface{ CanvasObject() fyne.CanvasObject })
			if !ok {
				return object.Errorf("type error: content must be a widget or canvas object"), nil
			}

			fyne.Do(func() {
				dialog.ShowCustom(title, dismiss, content.CanvasObject(), d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowCustomConfirm":
		return object.NewBuiltin("dialog.ShowCustomConfirm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 5 {
				return object.Errorf("wrong number of arguments. got=%d, want=5", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			confirm, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: confirm must be string, got %s", args[1].Type()), nil
			}
			dismiss, err := object.AsString(args[2])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[2].Type()), nil
			}

			// Content must be a CanvasObject
			content, ok := args[3].(interface{ CanvasObject() fyne.CanvasObject })
			if !ok {
				return object.Errorf("type error: content must be a widget or canvas object"), nil
			}

			fn, ok := args[4].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[4].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			fyne.Do(func() {
				dialog.ShowCustomConfirm(title, confirm, dismiss, content.CanvasObject(), callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowColorPicker":
		return object.NewBuiltin("dialog.ShowColorPicker", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[1].Type()), nil
			}
			fn, ok := args[2].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[2].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(c color.Color) {
				d.w.Do(func() {
					// Convert color to hex string for easier use in scripts
					r, g, b, a := c.RGBA()
					// RGBA returns values in range 0-65535, convert to 0-255
					colorMap := object.NewMap(map[string]object.Object{
						"R": object.NewInt(int64(r >> 8)),
						"G": object.NewInt(int64(g >> 8)),
						"B": object.NewInt(int64(b >> 8)),
						"A": object.NewInt(int64(a >> 8)),
					})
					callFunc(ctx, fn, []object.Object{colorMap})
				})
			}

			fyne.Do(func() {
				dialog.ShowColorPicker(title, message, callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "ShowForm":
		return object.NewBuiltin("dialog.ShowForm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 5 {
				return object.Errorf("wrong number of arguments. got=%d, want=5", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			confirm, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: confirm must be string, got %s", args[1].Type()), nil
			}
			dismiss, err := object.AsString(args[2])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[2].Type()), nil
			}

			// Items must be a list of FormItems
			itemsList, ok := args[3].(*object.List)
			if !ok {
				return object.Errorf("type error: items must be a list, got %s", args[3].Type()), nil
			}

			var formItems []*widget.FormItem
			for i, item := range itemsList.Value() {
				formItemObj, ok := item.(interface{ Interface() interface{} })
				if !ok {
					return object.Errorf("type error: item %d is not a valid form item", i), nil
				}
				formItem, ok := formItemObj.Interface().(*widget.FormItem)
				if !ok {
					return object.Errorf("type error: item %d is not a form item", i), nil
				}
				formItems = append(formItems, formItem)
			}

			fn, ok := args[4].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[4].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			fyne.Do(func() {
				dialog.ShowForm(title, confirm, dismiss, formItems, callback, d.w.FyneWindow)
			})
			return object.Nil, nil
		}), true

	case "NewFileOpen":
		return object.NewBuiltin("dialog.NewFileOpen", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(reader fyne.URIReadCloser, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if reader != nil {
						path = object.NewString(reader.URI().Path())
						defer reader.Close()
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fileDialog := risorDialog.NewFileOpen(callback, d.w.FyneWindow)
			return fileDialog, nil
		}), true

	case "NewFileSave":
		return object.NewBuiltin("dialog.NewFileSave", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(writer fyne.URIWriteCloser, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if writer != nil {
						path = object.NewString(writer.URI().Path())
						defer writer.Close()
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fileDialog := risorDialog.NewFileSave(callback, d.w.FyneWindow)
			return fileDialog, nil
		}), true

	case "NewFolderOpen":
		return object.NewBuiltin("dialog.NewFolderOpen", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(uri fyne.ListableURI, err error) {
				d.w.Do(func() {
					var path object.Object = object.Nil
					var errObj object.Object = object.Nil

					if uri != nil {
						path = object.NewString(uri.Path())
					}
					if err != nil {
						errObj = object.NewString(err.Error())
					}

					callFunc(ctx, fn, []object.Object{path, errObj})
				})
			}

			fileDialog := risorDialog.NewFolderOpen(callback, d.w.FyneWindow)
			return fileDialog, nil
		}), true

	case "NewConfirm":
		return object.NewBuiltin("dialog.NewConfirm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[1].Type()), nil
			}
			fn, ok := args[2].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[2].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			confirmDialog := risorDialog.NewConfirm(title, message, callback, d.w.FyneWindow)
			return confirmDialog, nil
		}), true

	case "NewCustomConfirm":
		return object.NewBuiltin("dialog.NewCustomConfirm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 5 {
				return object.Errorf("wrong number of arguments. got=%d, want=5", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			confirm, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: confirm must be string, got %s", args[1].Type()), nil
			}
			dismiss, err := object.AsString(args[2])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[2].Type()), nil
			}

			content, ok := args[3].(interface{ CanvasObject() fyne.CanvasObject })
			if !ok {
				return object.Errorf("type error: content must be a widget or canvas object"), nil
			}

			fn, ok := args[4].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[4].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			confirmDialog := risorDialog.NewCustomConfirm(title, confirm, dismiss, content.CanvasObject(), callback, d.w.FyneWindow)
			return confirmDialog, nil
		}), true

	case "NewCustom":
		return object.NewBuiltin("dialog.NewCustom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			dismiss, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[1].Type()), nil
			}

			content, ok := args[2].(interface{ CanvasObject() fyne.CanvasObject })
			if !ok {
				return object.Errorf("type error: content must be a widget or canvas object"), nil
			}

			customDialog := risorDialog.NewCustom(title, dismiss, content.CanvasObject(), d.w.FyneWindow)
			return customDialog, nil
		}), true

	case "NewColorPicker":
		return object.NewBuiltin("dialog.NewColorPicker", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			message, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: message must be string, got %s", args[1].Type()), nil
			}
			fn, ok := args[2].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[2].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(c color.Color) {
				d.w.Do(func() {
					r, g, b, a := c.RGBA()
					colorMap := object.NewMap(map[string]object.Object{
						"R": object.NewInt(int64(r >> 8)),
						"G": object.NewInt(int64(g >> 8)),
						"B": object.NewInt(int64(b >> 8)),
						"A": object.NewInt(int64(a >> 8)),
					})
					callFunc(ctx, fn, []object.Object{colorMap})
				})
			}

			colorDialog := risorDialog.NewColorPicker(title, message, callback, d.w.FyneWindow)
			return colorDialog, nil
		}), true

	case "NewForm":
		return object.NewBuiltin("dialog.NewForm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 5 {
				return object.Errorf("wrong number of arguments. got=%d, want=5", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: title must be string, got %s", args[0].Type()), nil
			}
			confirm, err := object.AsString(args[1])
			if err != nil {
				return object.Errorf("type error: confirm must be string, got %s", args[1].Type()), nil
			}
			dismiss, err := object.AsString(args[2])
			if err != nil {
				return object.Errorf("type error: dismiss must be string, got %s", args[2].Type()), nil
			}

			itemsList, ok := args[3].(*object.List)
			if !ok {
				return object.Errorf("type error: items must be a list, got %s", args[3].Type()), nil
			}

			var formItems []*widget.FormItem
			for i, item := range itemsList.Value() {
				formItemObj, ok := item.(interface{ Interface() interface{} })
				if !ok {
					return object.Errorf("type error: item %d is not a valid form item", i), nil
				}
				formItem, ok := formItemObj.Interface().(*widget.FormItem)
				if !ok {
					return object.Errorf("type error: item %d is not a form item", i), nil
				}
				formItems = append(formItems, formItem)
			}

			fn, ok := args[4].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[4].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("internal error: no call function in context"), nil
			}

			callback := func(confirmed bool) {
				d.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewBool(confirmed)})
				})
			}

			formDialog := risorDialog.NewForm(title, confirm, dismiss, formItems, callback, d.w.FyneWindow)
			return formDialog, nil
		}), true
	}

	return nil, false
}
