package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const ToolbarType object.Type = "widget.Toolbar"

type Toolbar struct {
	instance *widget.Toolbar
}

func NewToolbar(items ...widget.ToolbarItem) *Toolbar {
	return &Toolbar{
		instance: widget.NewToolbar(items...),
	}
}

func (obj *Toolbar) Type() object.Type {
	return ToolbarType
}

func (obj *Toolbar) Inspect() string {
	return fmt.Sprintf("Toolbar(items=%d)", len(obj.instance.Items))
}

func (obj *Toolbar) Interface() interface{} {
	return obj.instance
}

func (obj *Toolbar) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Toolbar) IsTruthy() bool {
	return true
}

func (obj *Toolbar) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal Toolbar")
}

func (obj *Toolbar) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for Toolbar: %v", opType), nil
}

func (obj *Toolbar) Equals(other object.Object) bool {
	if other, ok := other.(*Toolbar); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *Toolbar) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Toolbar) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: Toolbar has no writable attributes")
}

func (obj *Toolbar) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Append":
		return object.NewBuiltin("Toolbar.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			item, ok := args[0].Interface().(widget.ToolbarItem)
			if !ok {
				return object.Errorf("argument error: expected ToolbarItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.Append(item)
			})
			return object.Nil, nil
		}), true

	case "Prepend":
		return object.NewBuiltin("Toolbar.Prepend", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			item, ok := args[0].Interface().(widget.ToolbarItem)
			if !ok {
				return object.Errorf("argument error: expected ToolbarItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.Prepend(item)
			})
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("Toolbar.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Toolbar.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Toolbar.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

// ToolbarAction wraps a toolbar action button
const ToolbarActionType object.Type = "widget.ToolbarAction"

type ToolbarAction struct {
	instance *widget.ToolbarAction
}

func NewToolbarAction(icon fyne.Resource, onActivated func()) *ToolbarAction {
	return &ToolbarAction{
		instance: widget.NewToolbarAction(icon, onActivated),
	}
}

func (obj *ToolbarAction) Type() object.Type {
	return ToolbarActionType
}

func (obj *ToolbarAction) Inspect() string {
	return "ToolbarAction"
}

func (obj *ToolbarAction) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarAction) IsTruthy() bool {
	return true
}

func (obj *ToolbarAction) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal ToolbarAction")
}

func (obj *ToolbarAction) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for ToolbarAction: %v", opType), nil
}

func (obj *ToolbarAction) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarAction); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarAction) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ToolbarAction) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarAction has no writable attributes")
}

func (obj *ToolbarAction) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

// ToolbarSeparator wraps a toolbar separator
const ToolbarSeparatorType object.Type = "widget.ToolbarSeparator"

type ToolbarSeparator struct {
	instance *widget.ToolbarSeparator
}

func NewToolbarSeparator() *ToolbarSeparator {
	return &ToolbarSeparator{
		instance: widget.NewToolbarSeparator(),
	}
}

func (obj *ToolbarSeparator) Type() object.Type {
	return ToolbarSeparatorType
}

func (obj *ToolbarSeparator) Inspect() string {
	return "ToolbarSeparator"
}

func (obj *ToolbarSeparator) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarSeparator) IsTruthy() bool {
	return true
}

func (obj *ToolbarSeparator) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal ToolbarSeparator")
}

func (obj *ToolbarSeparator) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for ToolbarSeparator: %v", opType), nil
}

func (obj *ToolbarSeparator) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarSeparator); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarSeparator) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ToolbarSeparator) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarSeparator has no writable attributes")
}

func (obj *ToolbarSeparator) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

// ToolbarSpacer wraps a toolbar spacer
const ToolbarSpacerType object.Type = "widget.ToolbarSpacer"

type ToolbarSpacer struct {
	instance *widget.ToolbarSpacer
}

func NewToolbarSpacer() *ToolbarSpacer {
	return &ToolbarSpacer{
		instance: widget.NewToolbarSpacer(),
	}
}

func (obj *ToolbarSpacer) Type() object.Type {
	return ToolbarSpacerType
}

func (obj *ToolbarSpacer) Inspect() string {
	return "ToolbarSpacer"
}

func (obj *ToolbarSpacer) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarSpacer) IsTruthy() bool {
	return true
}

func (obj *ToolbarSpacer) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal ToolbarSpacer")
}

func (obj *ToolbarSpacer) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for ToolbarSpacer: %v", opType), nil
}

func (obj *ToolbarSpacer) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarSpacer); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarSpacer) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ToolbarSpacer) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarSpacer has no writable attributes")
}

func (obj *ToolbarSpacer) GetAttr(name string) (object.Object, bool) {
	return nil, false
}
