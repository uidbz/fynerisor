package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &FileIcon{}

const FileIconType object.Type = "widget.FileIcon"

// FileIcon wraps fyne's FileIcon widget - displays an icon for files and folders.
//
// FileIcon shows a visual representation of a file or folder with an appropriate icon
// based on the file type or path.
//
// Example usage in Risor:
//
//	let icon = widget.NewFileIcon("file:///home/user/document.pdf")
//	window.SetContent(icon)
//
//	// Update to different file
//	icon.SetURI("file:///home/user/image.png")
//
// The icon automatically selects the appropriate visual based on file extension.
type FileIcon struct {
	instance *widget.FileIcon
}

func NewFileIcon(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	uriStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	uri, err := storage.ParseURI(uriStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	instance := widget.NewFileIcon(uri)

	return &FileIcon{
		instance: instance,
	}, nil
}

func (obj *FileIcon) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *FileIcon) Type() object.Type {
	return FileIconType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *FileIcon) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *FileIcon) inspectUI() string {
	return "widget.FileIcon"
}

func (obj *FileIcon) Interface() interface{} {
	return obj.instance
}

func (obj *FileIcon) IsTruthy() bool {
	return true
}

func (obj *FileIcon) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FileIconType, opType)
}

func (obj *FileIcon) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FileIcon) Attrs() []object.AttrSpec {
	return fileIconAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *FileIcon) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *FileIcon) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Selected":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetSelected(b)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", FileIconType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *FileIcon) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var fileIconAttrs = objattrs.Register(string(FileIconType), object.NewAttrRegistry[*FileIcon](string(FileIconType)))

func init() {
	objattrs.Add(fileIconAttrs)
	fileIconAttrs.Define("URI").
		Doc("The file URI shown by the icon, or nil.").
		Getter(func(obj *FileIcon) object.Object {
			if obj.instance.URI == nil {
				return object.Nil
			}
			return object.NewString(obj.instance.URI.String())
		})
	fileIconAttrs.Define("Selected").
		Doc("Whether the file is selected.").
		Getter(func(obj *FileIcon) object.Object {
			return object.NewBool(obj.instance.Selected)
		})
	fileIconAttrs.Define("SetURI").
		Doc("Set the file URI from a string.").
		Arg("uri").
		Returns("nil").
		Impl(func(obj *FileIcon, _ context.Context, args ...object.Object) (object.Object, error) {
			uriStr, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			uri, err := storage.ParseURI(uriStr)
			if err != nil {
				return nil, fmt.Errorf("invalid URI: %w", err)
			}
			guithread.Do(func() {
				obj.instance.SetURI(uri)
			})
			return object.Nil, nil
		})
}

func (obj *FileIcon) getAttrUI(name string) (object.Object, bool) {
	return fileIconAttrs.GetAttr(obj, name)
}
