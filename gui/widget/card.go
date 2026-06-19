package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

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

func (obj *Card) Inspect() string {
	return fmt.Sprintf("widget.Card(title=%q, subtitle=%q)", obj.instance.Title, obj.instance.Subtitle)
}

func (obj *Card) Interface() interface{} {
	return obj.instance
}

func (obj *Card) IsTruthy() bool {
	return true
}

func (obj *Card) Cost() int {
	return 0
}

func (obj *Card) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Card) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Card'")
}

func (obj *Card) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(CardType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CardType, opType)
	return errObj, err
}

func (obj *Card) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Card) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Card) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Title":
		return object.NewString(obj.instance.Title), true
	case "Subtitle":
		return object.NewString(obj.instance.Subtitle), true
	case "SetTitle":
		return object.NewBuiltin("widget.Card.SetTitle", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				obj.instance.SetTitle(title)
			})
			return object.Nil, nil
		}), true
	case "SetSubTitle":
		return object.NewBuiltin("widget.Card.SetSubTitle", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			subtitle, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() {
				obj.instance.SetSubTitle(subtitle)
			})
			return object.Nil, nil
		}), true
	case "SetContent":
		return object.NewBuiltin("widget.Card.SetContent", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			contentObj, ok := args[0].(IsCanvasObject)
			if !ok {
				return object.Errorf("argument error: expected CanvasObject, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.SetContent(contentObj.CanvasObject())
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Card.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Card.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func (obj *Card) SetAttr(name string, value object.Object) error {
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
