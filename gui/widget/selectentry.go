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

var _ object.Object = &SelectEntry{}

const SelectEntryType object.Type = "widget.SelectEntry"

// SelectEntry wraps fyne's SelectEntry widget - a searchable dropdown entry
type SelectEntry struct {
	instance *widget.SelectEntry
	w        WindowInterface
}

func NewSelectEntry(w WindowInterface, options []string, onChanged func(string)) *SelectEntry {
	var instance *widget.SelectEntry
	if onChanged != nil {
		instance = widget.NewSelectEntry(options)
		instance.OnChanged = onChanged
	} else {
		instance = widget.NewSelectEntry(options)
	}

	return &SelectEntry{
		instance: instance,
		w:        w,
	}
}

func (obj *SelectEntry) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *SelectEntry) Type() object.Type {
	return SelectEntryType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *SelectEntry) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *SelectEntry) inspectUI() string {
	return "widget.SelectEntry"
}

func (obj *SelectEntry) Interface() interface{} {
	return obj.instance
}

func (obj *SelectEntry) IsTruthy() bool {
	return true
}

func (obj *SelectEntry) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", SelectEntryType, opType)
}

func (obj *SelectEntry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *SelectEntry) Attrs() []object.AttrSpec {
	return selectEntryAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *SelectEntry) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *SelectEntry) setAttrUI(name string, value object.Object) error {
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

	case "PlaceHolder":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.PlaceHolder = text
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", SelectEntryType, name)
}

var selectEntryAttrs = object.NewAttrRegistry[*SelectEntry](string(SelectEntryType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *SelectEntry) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *SelectEntry) getAttrUI(name string) (object.Object, bool) {
	return selectEntryAttrs.GetAttr(obj, name)

	// Change event: coalesced, latest text wins (issue #51).

}
func init() {
	objattrs.Add(selectEntryAttrs)
	selectEntryAttrs.
		Define("Text").Doc(
		"The text.",
	).Getter(func(obj *SelectEntry) object.Object {
		return object.NewString(obj.instance.Text)
	})
	selectEntryAttrs.
		Define("PlaceHolder").Getter(func(obj *SelectEntry) object.Object {
		return object.NewString(obj.instance.PlaceHolder)
	})
	selectEntryAttrs.
		Define("SetText").Doc("Set the text.").Args("arg1").Returns("nil").Impl(
		func(obj *SelectEntry, ctx context.Context,
			args ...object.Object) (object.Object, error) {
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			guithread.Do(
				func() {
					obj.instance.SetText(text)
				},
			)
			return object.Nil, nil
		})
	selectEntryAttrs.
		Define("SetOptions").Args(
		"arg1",
	).Returns("nil").Impl(func(obj *SelectEntry, ctx context.Context, args ...object.Object) (object.Object, error) {
		list, err :=

			object.AsList(args[0])
		if err != nil {
			return nil, err
		}
		options := make([]string, len(list.
			Value()))
		for i, item := range list.
			Value() {
			opt, err := object.AsString(item)
			if err != nil {
				return nil, err
			}
			options[i] = opt
		}
		guithread.Do(func() {
			obj.instance.
				SetOptions(options)
		},
		)
		return object.Nil, nil
	})
	selectEntryAttrs.
		Define("OnChanged").
		Doc("Set the change handler; fires on the script VM goroutine.").Args("arg1").Returns("nil").Impl(func(obj *SelectEntry, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "widget.SelectEntry.OnChanged",

			args[0])
		if err !=
			nil {
			return nil, err
		}
		obj.instance.OnChanged =
			func(text string) {
				cb.FireCoalesced(object.NewString(text))
			}
		return object.
			Nil, nil
	})
}
