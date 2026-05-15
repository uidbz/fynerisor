package widget

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Hyperlink{}

const HyperlinkType object.Type = "widget.Hyperlink"

// Hyperlink wraps widget.Hyperlink for Risor scripting
type Hyperlink struct {
	instance *widget.Hyperlink
	w WindowInterface
}

// NewHyperlink creates a new hyperlink widget
func NewHyperlink(text, urlStr string, w WindowInterface) (*Hyperlink, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}
	return &Hyperlink{
		instance: widget.NewHyperlink(text, parsedURL),
		w:        w,
	}, nil
}

func (obj *Hyperlink) Type() object.Type {
	return HyperlinkType
}

func (obj *Hyperlink) Inspect() string {
	return fmt.Sprintf("widget.Hyperlink(text=%q, url=%q)", obj.instance.Text, obj.instance.URL.String())
}

func (obj *Hyperlink) Interface() interface{} {
	return obj.instance
}

func (obj *Hyperlink) IsTruthy() bool {
	return true
}

func (obj *Hyperlink) Cost() int {
	return 0
}

func (obj *Hyperlink) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Hyperlink) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Hyperlink'")
}

func (obj *Hyperlink) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(HyperlinkType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", HyperlinkType, opType)
	return errObj, err
}

func (obj *Hyperlink) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Hyperlink) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Hyperlink) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true
	case "URL":
		return object.NewString(obj.instance.URL.String()), true
	case "SetText":
		return object.NewBuiltin("widget.Hyperlink.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.SetText(text)
			})
			return object.Nil, nil
		}), true
	case "SetURL":
		return object.NewBuiltin("widget.Hyperlink.SetURL", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			urlStr, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			parsedURL, err := url.Parse(urlStr)
			if err != nil {
				return object.Errorf("invalid URL: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetURL(parsedURL)
			})
			return object.Nil, nil
		}), true
	case "OnTapped":
		return object.NewBuiltin("widget.Hyperlink.OnTapped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("hyperlink: unable to get call function"), nil
			}

			obj.instance.OnTapped = func() {
				obj.w.Do(func() {
					callFunc(ctx, fn, []object.Object{})
				})
			}
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *Hyperlink) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetText(text)
		})
		return nil
	case "URL":
		urlStr, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetURL(parsedURL)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", HyperlinkType, name)
}
