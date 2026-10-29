package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Toolbar) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Toolbar) inspectUI() string {
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

func (obj *Toolbar) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for Toolbar: %v", opType)
}

func (obj *Toolbar) Equals(other object.Object) bool {
	if other, ok := other.(*Toolbar); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *Toolbar) Attrs() []object.AttrSpec {
	return toolbarAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Toolbar) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Toolbar) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: Toolbar has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Toolbar) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var toolbarAttrs = objattrs.Register(string(ToolbarType), object.NewAttrRegistry[*Toolbar](string(ToolbarType)))

func (obj *Toolbar) getAttrUI(name string) (object.Object, bool) {
	return toolbarAttrs.GetAttr(obj, name)

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ToolbarAction) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ToolbarAction) inspectUI() string {
	return "ToolbarAction"
}

func (obj *ToolbarAction) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarAction) IsTruthy() bool {
	return true
}

func (obj *ToolbarAction) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for ToolbarAction: %v", opType)
}

func (obj *ToolbarAction) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarAction); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarAction) Attrs() []object.AttrSpec {
	return toolbarActionAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ToolbarAction) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ToolbarAction) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarAction has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *ToolbarAction) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var toolbarActionAttrs = objattrs.Register(string(ToolbarActionType), object.NewAttrRegistry[*ToolbarAction](string(ToolbarActionType)))

func (obj *ToolbarAction) getAttrUI(name string) (object.Object, bool) {
	return toolbarActionAttrs.GetAttr(obj, name)
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ToolbarSeparator) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ToolbarSeparator) inspectUI() string {
	return "ToolbarSeparator"
}

func (obj *ToolbarSeparator) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarSeparator) IsTruthy() bool {
	return true
}

func (obj *ToolbarSeparator) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for ToolbarSeparator: %v", opType)
}

func (obj *ToolbarSeparator) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarSeparator); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarSeparator) Attrs() []object.AttrSpec {
	return toolbarSeparatorAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ToolbarSeparator) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ToolbarSeparator) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarSeparator has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *ToolbarSeparator) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var toolbarSeparatorAttrs = objattrs.Register(string(ToolbarSeparatorType), object.NewAttrRegistry[*ToolbarSeparator](string(ToolbarSeparatorType)))

func (obj *ToolbarSeparator) getAttrUI(name string) (object.Object, bool) {
	return toolbarSeparatorAttrs.GetAttr(obj, name)
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ToolbarSpacer) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ToolbarSpacer) inspectUI() string {
	return "ToolbarSpacer"
}

func (obj *ToolbarSpacer) Interface() interface{} {
	return obj.instance
}

func (obj *ToolbarSpacer) IsTruthy() bool {
	return true
}

func (obj *ToolbarSpacer) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for ToolbarSpacer: %v", opType)
}

func (obj *ToolbarSpacer) Equals(other object.Object) bool {
	if other, ok := other.(*ToolbarSpacer); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *ToolbarSpacer) Attrs() []object.AttrSpec {
	return toolbarSpacerAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ToolbarSpacer) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ToolbarSpacer) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: ToolbarSpacer has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *ToolbarSpacer) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var toolbarSpacerAttrs = objattrs.Register(string(ToolbarSpacerType), object.NewAttrRegistry[*ToolbarSpacer](string(ToolbarSpacerType)))

func (obj *ToolbarSpacer) getAttrUI(name string) (object.Object, bool) {
	return toolbarSpacerAttrs.GetAttr(obj, name)
}
func init() {
	objattrs.Add(toolbarAttrs)
	toolbarAttrs.Define("Append").Args("arg1").
		Returns("nil").Impl(
		func(obj *Toolbar, ctx context.
			Context, args ...object.Object) (object.
			Object, error) {
			item, ok := args[0].Interface().(widget.ToolbarItem)
			if !ok {
				return nil,
					object.
						Errorf("argument error: expected ToolbarItem, got %s",

							args[0].Type())
			}
			guithread.Do(func() {

				obj.instance.Append(item)
			})
			return object.Nil, nil
		})
	toolbarAttrs.Define("Prepend").Args(
		"arg1").
		Returns("nil").Impl(func(obj *Toolbar, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		item, ok := args[0].Interface().(widget.ToolbarItem)
		if !ok {
			return nil,

				object.
					Errorf("argument error: expected ToolbarItem, got %s",

						args[0].Type())
		}
		guithread.Do(func() {
			obj.instance.Prepend(item)
		})
		return object.Nil, nil
	})
}
