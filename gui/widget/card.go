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

var _ object.Object = &Card{}

const CardType object.Type = "widget.Card"

// Card wraps widget.Card for Risor scripting
type Card struct {
	instance *widget.Card
}

// NewCard creates a new card widget
func NewCard(title, subtitle string, content fyne.CanvasObject) *Card {
	return &Card{instance: widget.NewCard(title, subtitle, content)}
}

func (obj *Card) Type() object.Type {
	return CardType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Card) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Card) inspectUI() string {
	return fmt.Sprintf("widget.Card(title=%q, subtitle=%q)", obj.instance.Title, obj.instance.Subtitle)
}

func (obj *Card) Interface() interface{} {
	return obj.instance
}

func (obj *Card) IsTruthy() bool {
	return true
}

func (obj *Card) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Card) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CardType, opType)
}

func (obj *Card) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Card) Attrs() []object.AttrSpec {
	return cardAttrs.Specs()
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
var cardAttrs = object.NewAttrRegistry[*Card](string(CardType))

func (obj *Card) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Card) getAttrUI(name string) (object.Object, bool) {
	return cardAttrs.GetAttr(obj, name)

}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Card) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Card) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Title":
		title, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetTitle(title)
		})
		return nil
	case "Subtitle":
		subtitle, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetSubTitle(subtitle)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", CardType, name)
}
func init() {
	objattrs.Add(cardAttrs)
	cardAttrs.Define("Title").
		Getter(func(obj *Card) object.Object {
			return object.NewString(
				obj.instance.Title)
		})
	cardAttrs.Define("Subtitle").Getter(
		func(obj *Card) object.Object {
			return object.NewString(obj.instance.Subtitle)
		})
	cardAttrs.Define("SetTitle").Args("arg1").Returns("nil").Impl(func(obj *Card, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		title, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		guithread.
			Do(func() {
				obj.instance.
					SetTitle(title)
			})
		return object.Nil, nil
	},
	)
	cardAttrs.Define("SetSubTitle").Args("arg1").
		Returns("nil").Impl(func(obj *Card, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		subtitle, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}

		guithread.Do(func() {
			obj.instance.
				SetSubTitle(subtitle)
		})
		return object.
			Nil, nil
	})
	cardAttrs.Define("SetContent").Args(
		"arg1").
		Returns("nil").Impl(func(obj *Card, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		contentObj, ok := args[0].(IsCanvasObject)
		if !ok {
			return nil, object.Errorf(
				"argument error: expected CanvasObject, got %s",

				args[0].Type())
		}
		guithread.
			Do(func() {
				obj.instance.SetContent(contentObj.
					CanvasObject(),
				)
			})
		return object.Nil,

			nil
	})
}
