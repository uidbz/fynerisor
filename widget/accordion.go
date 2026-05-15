package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const AccordionType object.Type = "widget.Accordion"
const AccordionItemType object.Type = "widget.AccordionItem"

type Accordion struct {
	instance *widget.Accordion
}

func NewAccordion(items ...*widget.AccordionItem) *Accordion {
	return &Accordion{
		instance: widget.NewAccordion(items...),
	}
}

func (obj *Accordion) Type() object.Type {
	return AccordionType
}

func (obj *Accordion) Inspect() string {
	return fmt.Sprintf("Accordion(items=%d)", len(obj.instance.Items))
}

func (obj *Accordion) Interface() interface{} {
	return obj.instance
}

func (obj *Accordion) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Accordion) IsTruthy() bool {
	return true
}

func (obj *Accordion) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal Accordion")
}

func (obj *Accordion) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for Accordion: %v", opType), nil
}

func (obj *Accordion) Equals(other object.Object) bool {
	if other, ok := other.(*Accordion); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *Accordion) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Accordion) SetAttr(name string, value object.Object) error {
	switch name {
	case "MultiOpen":
		multiOpen, err := object.AsBool(value)
		if err != nil {
			return err
		}
		fyne.Do(func() {
			obj.instance.MultiOpen = multiOpen
			obj.instance.Refresh()
		})
		return nil
	default:
		return fmt.Errorf("attribute error: Accordion has no attribute %q", name)
	}
}

func (obj *Accordion) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "MultiOpen":
		return object.NewBool(obj.instance.MultiOpen), true

	case "Append":
		return object.NewBuiltin("Accordion.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			item, ok := args[0].Interface().(*widget.AccordionItem)
			if !ok {
				return object.Errorf("argument error: expected AccordionItem, got %s", args[0].Type()), nil
			}
			fyne.Do(func() {
				obj.instance.Append(item)
			})
			return object.Nil, nil
		}), true

	case "Remove":
		return object.NewBuiltin("Accordion.Remove", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			item, ok := args[0].Interface().(*widget.AccordionItem)
			if !ok {
				return object.Errorf("argument error: expected AccordionItem, got %s", args[0].Type()), nil
			}
			fyne.Do(func() {
				obj.instance.Remove(item)
			})
			return object.Nil, nil
		}), true

	case "RemoveIndex":
		return object.NewBuiltin("Accordion.RemoveIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.RemoveIndex(int(index))
			})
			return object.Nil, nil
		}), true

	case "Open":
		return object.NewBuiltin("Accordion.Open", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.Open(int(index))
			})
			return object.Nil, nil
		}), true

	case "Close":
		return object.NewBuiltin("Accordion.Close", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.Close(int(index))
			})
			return object.Nil, nil
		}), true

	case "CloseAll":
		return object.NewBuiltin("Accordion.CloseAll", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.CloseAll()
			})
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("Accordion.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

// AccordionItem wraps a Fyne AccordionItem
type AccordionItem struct {
	instance *widget.AccordionItem
}

func NewAccordionItem(title string, detail string, content fyne.CanvasObject) *AccordionItem {
	return &AccordionItem{
		instance: widget.NewAccordionItem(title, content),
	}
}

func (obj *AccordionItem) Type() object.Type {
	return AccordionItemType
}

func (obj *AccordionItem) Inspect() string {
	return fmt.Sprintf("AccordionItem(%q)", obj.instance.Title)
}

func (obj *AccordionItem) Interface() interface{} {
	return obj.instance
}

func (obj *AccordionItem) IsTruthy() bool {
	return true
}

func (obj *AccordionItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal AccordionItem")
}

func (obj *AccordionItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for AccordionItem: %v", opType), nil
}

func (obj *AccordionItem) Equals(other object.Object) bool {
	if other, ok := other.(*AccordionItem); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *AccordionItem) Attrs() []object.AttrSpec {
	return nil
}

func (obj *AccordionItem) SetAttr(name string, value object.Object) error {
	switch name {
	case "Title":
		title, err := object.AsString(value)
		if err != nil {
			return err
		}
		obj.instance.Title = title
		return nil
	case "Open":
		open, err := object.AsBool(value)
		if err != nil {
			return err
		}
		obj.instance.Open = open
		return nil
	default:
		return fmt.Errorf("attribute error: AccordionItem has no attribute %q", name)
	}
}

func (obj *AccordionItem) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Title":
		return object.NewString(obj.instance.Title), true
	case "Open":
		return object.NewBool(obj.instance.Open), true
	}
	return nil, false
}
