package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Label{}

const LabelType object.Type = "widget.Label"

type Label struct {
	instance *widget.Label
}

func (obj *Label) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Label) Type() object.Type {
	return LabelType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Label) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Label) inspectUI() string {
	return "widget.Label"
}

func (obj *Label) Interface() interface{} {
	return obj.instance
}

func (obj *Label) IsTruthy() bool {
	return true
}

func (obj *Label) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", LabelType, opType)
}

func (obj *Label) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Label) Attrs() []object.AttrSpec {
	return labelAttrs.Specs()
}

// labelAttrs declares the widget.Label attribute surface (issue #38):
// introspectable via Attrs(), with arity handled by the registry.
var labelAttrs = object.NewAttrRegistry[*Label](string(LabelType))

func init() {
	objattrs.Add(labelAttrs)

	labelAttrs.Define("Text").
		Doc("The label's text.").
		Returns("string").
		Getter(func(o *Label) object.Object { return object.NewString(o.instance.Text) })
	labelAttrs.Define("Wrapping").
		Doc("Text wrapping mode (a constants.TextWrap* value).").
		Returns("int").
		Getter(func(o *Label) object.Object { return object.NewInt(int64(o.instance.Wrapping)) })
	labelAttrs.Define("Truncation").
		Doc("Text truncation mode.").
		Returns("int").
		Getter(func(o *Label) object.Object { return object.NewInt(int64(o.instance.Truncation)) })
	labelAttrs.Define("Alignment").
		Doc("Text alignment.").
		Returns("int").
		Getter(func(o *Label) object.Object { return object.NewInt(int64(o.instance.Alignment)) })
	labelAttrs.Define("Importance").
		Doc("Visual importance level.").
		Returns("int").
		Getter(func(o *Label) object.Object { return object.NewInt(int64(o.instance.Importance)) })
	labelAttrs.Define("Selectable").
		Doc("Whether the text can be selected.").
		Returns("bool").
		Getter(func(o *Label) object.Object { return object.NewBool(o.instance.Selectable) })
	labelAttrs.Define("TextStyle").
		Doc("Text style object; set Bold/Italic/Monospace on it.").
		Returns("TextStyle").
		Getter(func(o *Label) object.Object { return NewTextStyle(&o.instance.TextStyle, o.instance) })

	labelAttrs.Define("SetText").
		Doc("Set the label text.").
		Arg("text").
		Returns("nil").
		Impl(func(o *Label, _ context.Context, args ...object.Object) (object.Object, error) {
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			guithread.Do(func() { o.instance.SetText(text) })
			return object.Nil, nil
		})
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Label) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Label) setAttrUI(name string, value object.Object) error {
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
	case "Alignment":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Alignment = fyne.TextAlign(i)
			obj.instance.Refresh()
		})
		return nil
	case "Importance":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Importance = widget.Importance(i)
			obj.instance.Refresh()
		})
		return nil
	case "Selectable":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Selectable = b
			obj.instance.Refresh()
		})
		return nil
	case "TextStyle":
		return fmt.Errorf("attribute error: TextStyle is read-only, set Bold/Italic/Monospace properties instead")
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", LabelType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Label) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Label) getAttrUI(name string) (object.Object, bool) {
	return labelAttrs.GetAttr(obj, name)
}

func NewLabel(text string) *Label {
	return &Label{instance: widget.NewLabel(text)}
}

func NewLabelWithData(data interface{ Get() (string, error) }) *Label {
	// Cast to binding.String
	bindingStr, ok := data.(binding.String)
	if !ok {
		// Fallback: create a label with empty text
		return &Label{instance: widget.NewLabel("")}
	}

	return &Label{instance: widget.NewLabelWithData(bindingStr)}
}
