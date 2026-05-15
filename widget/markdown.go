package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Markdown{}

const MarkdownType object.Type = "widget.Markdown"

type Markdown struct {
	instance *widget.RichText
	callback func(text, url string)
}

func (obj *Markdown) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Markdown) onTapped(text, url string) {
	obj.callback(text, url)
}

func (obj *Markdown) Type() object.Type {
	return MarkdownType
}

func (obj *Markdown) Inspect() string {
	return "widget.Markdown"
}

func (obj *Markdown) Interface() interface{} {
	return obj.instance
}

func (obj *Markdown) IsTruthy() bool {
	return true
}

func (obj *Markdown) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'Markdown'")
}

func (obj *Markdown) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(MarkdownType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", MarkdownType, opType)
	return errObj, err
}

func (obj *Markdown) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Markdown) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Markdown) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", MarkdownType, name)
}

func (obj *Markdown) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetOnTapped":
		return object.NewBuiltin("widget.Markdown.SetOnTapped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			obj.callback = func(text, url string) {
				callFunc(ctx, fn, []object.Object{object.NewString(text), object.NewString(url)})
			}

			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewMarkdown(markdown string) *Markdown {
	richtext := &Markdown{
		instance: widget.NewRichTextFromMarkdown(markdown),
	}
	for _, s := range richtext.instance.Segments {
		if link, ok := s.(*widget.HyperlinkSegment); ok {
			link.OnTapped = func() {
				richtext.onTapped(link.Text, link.URL.String())
			}
		}
	}
	return richtext
}
