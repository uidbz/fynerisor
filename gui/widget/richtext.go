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

var _ object.Object = &RichText{}

const RichTextType object.Type = "widget.RichText"

// RichText wraps fyne's RichText widget - formatted text with markdown support.
//
// RichText displays formatted text with support for bold, italic, headers, links, lists,
// and more through markdown parsing. Direct segment manipulation is not exposed in
// Risor bindings - use ParseMarkdown() instead.
//
// Example usage in Risor:
//
//	let text = widget.NewRichText(`# Welcome
//
//	This supports **bold** and *italic* text, as well as [links](https://example.com).
//	`)
//
//	// Update content dynamically
//	text.ParseMarkdown("## New Content\n\nWith different **formatting**.")
//
//	// Control wrapping
//	text.Wrapping = 2  // 0=off, 1=break, 2=word
//
// Properties:
//   - Wrapping: Text wrap mode (0=off, 1=break, 2=word)
//   - Truncation: Truncation mode (0=off, 1=clip, 2=ellipsis)
//   - Scroll: Scroll direction (0=none, 1=horizontal, 2=vertical, 3=both)
//
// Note: Complex segment types (Image, List, Paragraph) not fully supported.
// Use ParseMarkdown() for rich formatting.
type RichText struct {
	instance *widget.RichText
}

func NewRichText(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) > 1 {
		return object.Errorf("wrong number of arguments. got=%d, want=0 or 1", len(args)), nil
	}

	var instance *widget.RichText

	if len(args) == 1 {
		// Create with text
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		instance = widget.NewRichTextFromMarkdown(text)
	} else {
		instance = widget.NewRichText()
	}

	return &RichText{
		instance: instance,
	}, nil
}

func (obj *RichText) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *RichText) Type() object.Type {
	return RichTextType
}

func (obj *RichText) Inspect() string {
	return "widget.RichText"
}

func (obj *RichText) Interface() interface{} {
	return obj.instance
}

func (obj *RichText) IsTruthy() bool {
	return true
}

func (obj *RichText) Cost() int {
	return 0
}

func (obj *RichText) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.RichText'")
}

func (obj *RichText) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(RichTextType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", RichTextType, opType)
	return errObj, err
}

func (obj *RichText) Equals(other object.Object) bool {
	return obj == other
}

func (obj *RichText) Attrs() []object.AttrSpec {
	return nil
}

func (obj *RichText) SetAttr(name string, value object.Object) error {
	switch name {
	case "Wrapping":
		// fyne.TextWrap values: WrapOff, WrapBreak, WrapWord
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Wrapping = fyne.TextWrap(i)
			obj.instance.Refresh()
		})
		return nil
	case "Truncation":
		// fyne.TextTruncation values: TruncateOff, TruncateClip, TruncateEllipsis
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Truncation = fyne.TextTruncation(i)
			obj.instance.Refresh()
		})
		return nil
	case "Scroll":
		// fyne.ScrollDirection values: ScrollNone, ScrollHorizontalOnly, ScrollVerticalOnly, ScrollBoth
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Scroll = fyne.ScrollDirection(i)
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", RichTextType, name)
}

func (obj *RichText) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Wrapping":
		return object.NewInt(int64(obj.instance.Wrapping)), true
	case "Truncation":
		return object.NewInt(int64(obj.instance.Truncation)), true
	case "Scroll":
		return object.NewInt(int64(obj.instance.Scroll)), true

	case "ParseMarkdown":
		return object.NewBuiltin("widget.RichText.ParseMarkdown", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			fyne.Do(func() {
				obj.instance.ParseMarkdown(text)
			})

			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.RichText.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			fyne.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true

	// Note: Direct segment manipulation not implemented
	// Use ParseMarkdown() for rich text content instead
	case "Segments":
		return object.Errorf("attribute error: Segments manipulation not supported in Risor bindings. Use ParseMarkdown() instead."), false
	}

	return nil, false
}
