package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
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

func (obj *Markdown) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", MarkdownType, opType)
}

func (obj *Markdown) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Markdown) Attrs() []object.AttrSpec {
	return markdownAttrs.Specs()
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

var markdownAttrs = object.NewAttrRegistry[*Markdown](string(MarkdownType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Markdown) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Markdown) getAttrUI(name string) (object.Object, bool) {
	// Event: link taps are queued on the script's VM goroutine.
	return markdownAttrs.GetAttr(obj, name)
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
func init() {
	objattrs.Add(markdownAttrs)
	markdownAttrs.Define("Text").Doc("The text.").Getter(func(obj *Markdown) object.Object {
		return object.NewString(obj.src)
	})
	markdownAttrs.Define("SetText").Doc(
		"Set the text.",
	).Args("arg1").Returns("nil").Impl(func(obj *Markdown, ctx context.Context, args ...object.Object) (object.Object, error) {

		s, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		obj.src = s
		guithread.Do(func() { obj.instance.ParseMarkdown(s) })
		return object.Nil, nil
	})
	markdownAttrs.Define("Wrapping").Doc("Text wrapping mode (a constants.TextWrap* value).").
		Getter(func(obj *Markdown) object.Object {
			return object.NewInt(int64(obj.instance.Wrapping))
		})
	markdownAttrs.Define("Truncation").Getter(func(obj *Markdown) object.
		Object {
		return object.
			NewInt(int64(obj.instance.Truncation))
	})
	markdownAttrs.Define("SetOnTapped").
		Args("arg1").Returns("nil").
		Impl(func(obj *Markdown,
			ctx context.Context, args ...object.Object) (object.Object, error) {
			cb, err := vmexec.
				NewCallback(ctx, "widget.Markdown.SetOnTapped",
					args[0])
			if err != nil {
				return nil, err
			}

			obj.callback = func(text, url string) {
				cb.Fire(object.NewString(text), object.NewString(url))
			}
			return object.
				Nil, nil
		})
}
