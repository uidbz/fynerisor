package widget

import (
	"context"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Hyperlink{}

const HyperlinkType object.Type = "widget.Hyperlink"

// Hyperlink wraps widget.Hyperlink for Risor scripting
type Hyperlink struct {
	instance *widget.Hyperlink
	w        WindowInterface
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Hyperlink) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Hyperlink) inspectUI() string {
	return fmt.Sprintf("widget.Hyperlink(text=%q, url=%q)", obj.instance.Text, obj.instance.URL.String())
}

func (obj *Hyperlink) Interface() interface{} {
	return obj.instance
}

func (obj *Hyperlink) IsTruthy() bool {
	return true
}

func (obj *Hyperlink) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Hyperlink) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", HyperlinkType, opType)
}

func (obj *Hyperlink) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Hyperlink) Attrs() []object.AttrSpec {
	return hyperlinkAttrs.Specs()
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
var hyperlinkAttrs = objattrs.Register(string(HyperlinkType), object.NewAttrRegistry[*Hyperlink](string(HyperlinkType)))

func (obj *Hyperlink) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Hyperlink) getAttrUI(name string) (object.Object, bool) {
	// Event: queued on the script's VM goroutine.
	return hyperlinkAttrs.GetAttr(obj, name)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Hyperlink) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Hyperlink) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
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
		guithread.Do(func() {
			obj.instance.SetURL(parsedURL)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", HyperlinkType, name)
}
func init() {
	objattrs.Add(hyperlinkAttrs)
	hyperlinkAttrs.
		Define("Text").Doc("The text.").Getter(
		func(obj *Hyperlink) object.Object {
			return object.
				NewString(obj.instance.Text)
		})
	hyperlinkAttrs.
		Define("URL").Getter(
		func(obj *Hyperlink) object.Object {
			return object.NewString(obj.
				instance.URL.String())
		})
	hyperlinkAttrs.
		Define("SetText").Doc("Set the text.").
		Args("arg1").Returns("nil").Impl(func(obj *Hyperlink, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		guithread.Do(func() {
			obj.
				instance.
				SetText(text)
		})
		return object.
			Nil, nil
	})
	hyperlinkAttrs.
		Define("SetURL").Args("arg1").Returns("nil").Impl(func(obj *Hyperlink, ctx context.Context, args ...object.Object) (object.Object, error) {
		urlStr, err := object.AsString(
			args[0])
		if err != nil {
			return nil, err
		}
		parsedURL,
			err := url.Parse(urlStr)
		if err !=
			nil {
			return nil, object.Errorf("invalid URL: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.
				SetURL(parsedURL)
		})
		return object.Nil, nil
	})
	hyperlinkAttrs.
		Define("OnTapped").Doc("Set the tap handler; fires on the script VM goroutine.").Args("arg1").Returns("nil").Impl(func(obj *Hyperlink, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		cb, err := vmexec.
			NewCallback(ctx, "widget.Hyperlink.OnTapped",

				args[0])
		if err != nil {
			return nil, err
		}
		obj.instance.OnTapped = func() {
			cb.Fire()
		}
		return object.Nil, nil
	})
}
