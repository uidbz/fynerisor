package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Markdown{}

const MarkdownType object.Type = "widget.Markdown"

type Markdown struct {
	instance *widget.RichText
	src      string // last markdown source, so .Text is readable
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Markdown) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Markdown) inspectUI() string {
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

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Markdown) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Markdown) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Wrapping":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Wrapping = fyne.TextWrap(i)
			obj.instance.Refresh()
		})
		return nil
	case "Truncation":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Truncation = fyne.TextTruncation(i)
			obj.instance.Refresh()
		})
		return nil
	case "Text":
		s, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.src = s
		guithread.Do(func() {
			obj.instance.ParseMarkdown(s)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", MarkdownType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Markdown) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Markdown) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.src), true
	case "SetText":
		return object.NewBuiltin("widget.Markdown.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			s, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			obj.src = s
			guithread.Do(func() {
				obj.instance.ParseMarkdown(s)
			})
			return object.Nil, nil
		}), true
	case "Wrapping":
		return object.NewInt(int64(obj.instance.Wrapping)), true
	case "Truncation":
		return object.NewInt(int64(obj.instance.Truncation)), true
	case "SetOnTapped":
		return object.NewBuiltin("widget.Markdown.SetOnTapped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Markdown.SetOnTapped", args[0])
			if err != nil {
				return nil, err
			}

			// Event: link taps are queued on the script's VM goroutine.
			obj.callback = func(text, url string) {
				cb.Fire(object.NewString(text), object.NewString(url))
			}

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Markdown.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Markdown.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewMarkdown(markdown string) *Markdown {
	richtext := &Markdown{
		instance: widget.NewRichTextFromMarkdown(markdown),
		src:      markdown,
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
