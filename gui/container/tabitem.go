package container

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/container"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &TabItem{}

const TabItemType object.Type = "container.TabItem"

type TabItem struct {
	instance *container.TabItem
}

func (obj *TabItem) Type() object.Type {
	return TabItemType
}

func (obj *TabItem) Inspect() string {
	return fmt.Sprintf("container.TabItem(%q)", obj.instance.Text)
}

func (obj *TabItem) Interface() interface{} {
	return obj.instance
}

func (obj *TabItem) IsTruthy() bool {
	return true
}

func (obj *TabItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.TabItem'")
}

func (obj *TabItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(TabItemType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", TabItemType, opType)
	return errObj, err
}

func (obj *TabItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *TabItem) Attrs() []object.AttrSpec {
	return nil
}

func (obj *TabItem) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Text = text
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", TabItemType, name)
}

func (obj *TabItem) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true
	case "Disabled":
		return object.NewBool(obj.instance.Disabled()), true
	}
	return nil, false
}

func NewTabItem(text string, content fyne.CanvasObject) *TabItem {
	return &TabItem{
		instance: container.NewTabItem(text, content),
	}
}

func NewTabItemWithIcon(text string, icon fyne.Resource, content fyne.CanvasObject) *TabItem {
	return &TabItem{
		instance: container.NewTabItemWithIcon(text, icon, content),
	}
}
