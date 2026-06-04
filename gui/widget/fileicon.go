package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

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
		return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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

func (obj *FileIcon) Inspect() string {
	return "widget.FileIcon"
}

func (obj *FileIcon) Interface() interface{} {
	return obj.instance
}

func (obj *FileIcon) IsTruthy() bool {
	return true
}

func (obj *FileIcon) Cost() int {
	return 0
}

func (obj *FileIcon) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.FileIcon'")
}

func (obj *FileIcon) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(FileIconType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", FileIconType, opType)
	return errObj, err
}

func (obj *FileIcon) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FileIcon) Attrs() []object.AttrSpec {
	return nil
}

func (obj *FileIcon) SetAttr(name string, value object.Object) error {
	switch name {
	case "Selected":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetSelected(b)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", FileIconType, name)
}

func (obj *FileIcon) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "URI":
		if obj.instance.URI == nil {
			return object.Nil, true
		}
		return object.NewString(obj.instance.URI.String()), true

	case "Selected":
		return object.NewBool(obj.instance.Selected), true

	case "SetURI":
		return object.NewBuiltin("widget.FileIcon.SetURI", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			uriStr, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			uri, err := storage.ParseURI(uriStr)
			if err != nil {
				return nil, fmt.Errorf("invalid URI: %w", err)
			}

			fyne.Do(func() {
				obj.instance.SetURI(uri)
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.FileIcon.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.FileIcon.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}

	return nil, false
}
