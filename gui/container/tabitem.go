package container

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *TabItem) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *TabItem) inspectUI() string {
	return fmt.Sprintf("container.TabItem(%q)", obj.instance.Text)
}

func (obj *TabItem) Interface() interface{} {
	return obj.instance
}

func (obj *TabItem) IsTruthy() bool {
	return true
}

func (obj *TabItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", TabItemType, opType)
}

func (obj *TabItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *TabItem) Attrs() []object.AttrSpec {
	return tabItemAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *TabItem) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *TabItem) setAttrUI(name string, value object.Object) error {
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

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *TabItem) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var tabItemAttrs = object.NewAttrRegistry[*TabItem](string(TabItemType))

func (obj *TabItem) getAttrUI(name string) (object.Object, bool) {
	return tabItemAttrs.GetAttr(obj, name)

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
func init() {
	tabItemAttrs.Define("Text").
		Doc("The text.").Getter(func(obj *TabItem) object.Object {
		return object.NewString(obj.instance.Text)
	})
	tabItemAttrs.Define("Disabled").Doc(
		"Whether the object is disabled.",
	).Getter(func(obj *TabItem) object.Object {
		return object.NewBool(obj.instance.Disabled())
	})
}
