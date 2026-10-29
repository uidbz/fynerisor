package gui

import (
	"context"
	"errors"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"time"

	"fyne.io/fyne/v2"
	fynewidget "fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/vmexec"
	risorwidget "github.com/uidbz/fynerisor/gui/widget"
	timemodule "github.com/uidbz/fynerisor/modules/time"
)

const WidgetType object.Type = "widget"

// Widget is a factory object that provides widget creation functions to Risor scripts.
// It implements object.Object and exposes widget creation via GetAttr() for creating
// Fyne widgets like buttons, labels, entries, forms, tables, and more.
//
// Available in Risor scripts as the global 'widget' object.
//
// Example usage in Risor:
//
//	let btn = widget.NewButton("Click Me", () => { print("Clicked!") })
//	let lbl = widget.NewLabel("Hello World")
//	let entry = widget.NewEntry()
//	let table = widget.NewTable("Data", 20)
//
// Thread Safety: All widget creation and property access is thread-safe.
// Callbacks are automatically marshalled to the GUI thread.
type Widget struct {
	w *Window
}

func (obj *Widget) Type() object.Type {
	return WidgetType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Widget) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Widget) inspectUI() string {
	return "widget"
}

func (obj *Widget) Interface() interface{} {
	return nil
}

func (obj *Widget) IsTruthy() bool {
	return true
}

func (obj *Widget) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", WidgetType, opType)
}

func (obj *Widget) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Widget) Attrs() []object.AttrSpec {
	return append(widgetAttrs.Specs(),
		object.AttrSpec{Name: "NewTextGrid", Doc: "Create a TextGrid widget", Returns: "widget.TextGrid"},
		object.AttrSpec{Name: "NewRichText", Doc: "Create a RichText widget from text or segments", Args: []string{"text"}, Returns: "widget.RichText"},
		object.AttrSpec{Name: "NewPopUp", Doc: "Create a PopUp wrapping a canvas object", Args: []string{"object"}, Returns: "widget.PopUp"},
		object.AttrSpec{Name: "NewModalPopUp", Doc: "Create a modal PopUp wrapping a canvas object", Args: []string{"object"}, Returns: "widget.PopUp"},
		object.AttrSpec{Name: "NewFileIcon", Doc: "Create a FileIcon widget", Returns: "widget.FileIcon"},
		object.AttrSpec{Name: "NewPopUpMenu", Doc: "Create a PopUpMenu for a menu", Args: []string{"menu"}, Returns: "widget.PopUpMenu"},
	)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Widget) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Widget) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", WidgetType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Widget) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var widgetAttrs = objattrs.Register(string(WidgetType), object.NewAttrRegistry[*Widget](string(WidgetType)))

func (obj *Widget) getAttrUI(name string) (object.Object, bool) {
	if o, ok := widgetAttrs.GetAttr(obj, name); ok {
		return o, true
	}
	switch name {

	// Event: the tap handler is queued on the script's VM goroutine.

	// Get the binding.Bool object

	// First argument: start date (optional)

	// Try to extract time.Time from the TimeObject

	// Second argument: callback function (optional)

	// Get the binding.String object

	// Get the binding.String object

	// Get the binding.Float object

	// Get the binding.Float object

	case "NewTextGrid":
		return object.NewBuiltin("widget.NewTextGrid", risorwidget.NewTextGrid), true

	case "NewRichText":
		return object.NewBuiltin("widget.NewRichText", risorwidget.NewRichText), true

	case "NewPopUp":
		return object.NewBuiltin("widget.NewPopUp", risorwidget.NewPopUp), true

	case "NewModalPopUp":
		return object.NewBuiltin("widget.NewModalPopUp", risorwidget.NewModalPopUp), true

	case "NewFileIcon":
		return object.NewBuiltin("widget.NewFileIcon", risorwidget.NewFileIcon), true

	case "NewPopUpMenu":
		return object.NewBuiltin("widget.NewPopUpMenu", risorwidget.NewPopUpMenu), true
	}
	return nil, false

}

// callbackQueueSize bounds the queue of pending event callbacks. When it is
// full, events are dropped with a diagnostic instead of blocking Fyne's UI
// thread (issue #10).
const callbackQueueSize = vmexec.DefaultQueueSize

// queueCallback runs fn on the current run's VM goroutine, after the main
// body and earlier events. It never blocks the sender: before Execute,
// after Stop, or when the queue is full, the callback is dropped and the
// window's error sink reports it (issue #31: never the process stderr
// directly, so hosts can route or suppress).
func (w *Window) queueCallback(what string, fn func()) {
	r := w.run.Load()
	if r == nil {
		w.Report(what, errors.New("called before Execute; callback dropped"))
		return
	}
	_ = r.Enqueue(what, fn)
}

// Do queues fn on the current execution's callback queue so it runs on the
// run's VM goroutine with exclusive access to the Risor VM. It never blocks
// the caller (issue #10); see queueCallback. This is the HOST API — a host
// goroutine's way onto the VM goroutine; guithread.Do targets the UI thread
// instead. (The script-visible window.Do attribute is a separate, deprecated
// alias of go(): scripts already run on the VM goroutine.)
func (w *Window) Do(fn func()) {
	w.queueCallback("window.Do", fn)
}

func NewWidget(w *Window) *Widget {
	return &Widget{w: w}
}
func init() {
	widgetAttrs.
		Define("NewButton").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(
					args))
		}
		name, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		cb, err :=
			vmexec.NewCallback(ctx, "widget.NewButton", args[1])

		if err != nil {
			return nil, err
		}
		return risorwidget.NewButton(name, func() { cb.Fire() }), nil
	})
	widgetAttrs.
		Define("NewCheck").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		cb, err :=
			vmexec.NewCallback(ctx, "widget.NewCheck", args[1])
		if err != nil {
			return nil, err
		}
		return risorwidget.NewCheck(label, func(changed bool) { cb.Fire(object.NewBool(changed)) }), nil
	})
	widgetAttrs.
		Define("NewCheckWithData").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		bindingBool,
			ok := args[1].Interface().(interface {
			Get() (
				bool, error)
		})
		if !ok {
			return nil, object.
				Errorf("type error: expected binding.Bool")
		}
		return risorwidget.NewCheckWithData(
			label, bindingBool), nil
	})
	widgetAttrs.
		Define("NewCheckGroup").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {

		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",

				len(args))
		}
		options, err := object.AsStringSlice(args[0])
		if err != nil {
			return nil, err
		}
		cb, err := vmexec.NewCallback(ctx, "widget.NewCheckGroup",

			args[1])
		if err != nil {
			return nil,

				err
		}
		return risorwidget.NewCheckGroup(options, func(checked []string) {
			cb.Fire(object.
				NewStringList(checked))
		}), nil
	})
	widgetAttrs.
		Define("NewSelect").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(
					args))
		}
		options, err := object.AsStringSlice(args[0])
		if err != nil {
			return nil, err
		}
		cb,
			err := vmexec.NewCallback(ctx, "widget.NewSelect",
			args[1])
		if err != nil {
			return nil, err
		}
		return risorwidget.NewSelect(options, func(selected string) {
			cb.Fire(object.NewString(selected))
		}), nil
	})
	widgetAttrs.
		Define("NewSelectEntry").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf(
				"wrong number of arguments. got=%d, want=1",
				len(args))
		}
		options, err := object.AsStringSlice(args[0])
		if err != nil {
			return nil, err
		}
		selectEntry := risorwidget.NewSelectEntry(obj.w, options,

			nil)
		return selectEntry, nil
	})
	widgetAttrs.
		Define("NewDateEntry").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		dateEntry := risorwidget.NewDateEntry(obj.w)
		return dateEntry, nil
	})
	widgetAttrs.
		Define("NewCalendar").
		Variadic(
			"args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) > 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0, 1, or 2",

				len(args))
		}
		var startDate time.Time
		var onChanged func(time.Time)
		if len(args) >= 1 &&
			args[0] != object.Nil && args[0].Type() != object.NIL {
			if timeVal, ok := args[0].Interface().(time.Time); ok {
				startDate = timeVal
			} else {
				return nil, object.Errorf("argument error: expected time.Time or nil, got %s",
					args[0].Type())
			}
		} else {
			startDate = time.
				Now()
		}
		if len(args) == 2 {
			cb, err := vmexec.NewCallback(ctx, "widget.NewCalendar",

				args[1])
			if err != nil {
				return nil, err
			}
			onChanged = func(selected time.
				Time) {
				cb.Fire(timemodule.NewTimeObject(selected))
			}
		}
		calendar :=
			risorwidget.NewCalendar(obj.w, startDate, onChanged)
		return calendar, nil
	},
	)
	widgetAttrs.
		Define("NewEntry").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewEntry(obj.
			w), nil
	})
	widgetAttrs.
		Define("NewEntryWithData").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		bindingStr, ok := args[0].Interface().(interface{ Get() (string, error) })
		if !ok {
			return nil, object.Errorf("type error: expected binding.String")
		}
		return risorwidget.NewEntryWithData(bindingStr, obj.w), nil
	})
	widgetAttrs.
		Define("NewMultiLineEntry").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewMultiLineEntry(obj.w), nil
	})
	widgetAttrs.
		Define("NewForm").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) == 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1",

				len(args))
		}
		var objects []*fynewidget.
			FormItem
		for i, x := range args {
			if x.Type() ==
				object.LIST {
				items, err := object.AsList(x)
				if err !=
					nil {
					return nil, object.Errorf("NewForm: Error: %s",

						err.Error())
				}
				for j, y := range items.Value() {
					o, ok := y.Interface().(*fynewidget.FormItem)
					if !ok {
						return nil, object.Errorf("NewForm: Wrong type, expected FormItem within list at position: %d",
							j)
					}
					objects = append(objects, o)
				}
			} else {
				o, ok := x.Interface().(*fynewidget.FormItem)
				if !ok {
					return nil, object.Errorf("NewForm: Wrong type, expected FormItem at argument: %d",
						i)
				}
				objects = append(
					objects, o)
			}
		}
		return risorwidget.
			NewForm(objects...), nil
	})
	widgetAttrs.
		Define("NewFormItem").
		Variadic(
			"args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		o, ok := args[1].Interface().(fyne.CanvasObject)
		if !ok {
			return nil, object.Errorf("NewFormItem: Wrong type, expected CanvasObject at argument: %d",

				1)
		}
		return risorwidget.NewFormItem(label,
			o), nil
	})
	widgetAttrs.
		Define("NewLabel").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		return risorwidget.
			NewLabel(text), nil
	})
	widgetAttrs.
		Define("NewLabelWithData").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		bindingStr, ok := args[0].Interface().(interface{ Get() (string, error) })
		if !ok {
			return nil, object.Errorf("type error: expected binding.String")
		}
		return risorwidget.NewLabelWithData(bindingStr), nil
	})
	widgetAttrs.
		Define("NewLog").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(
			args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		maxItems, err := object.AsInt(args[0])
		if err != nil {
			return nil, err
		}
		return risorwidget.
			NewLog(int(maxItems)), nil
	})
	widgetAttrs.
		Define("NewMarkdown").
		Variadic(
			"args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		markdown, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		return risorwidget.NewMarkdown(markdown), nil
	})
	widgetAttrs.
		Define("NewProgressBar").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf(
				"wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewProgressBar(), nil
	})
	widgetAttrs.
		Define("NewProgressBarWithData").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.Context, args ...object.Object) (object.Object,

		error) {
		if len(args) != 1 {
			return nil, object.
				Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
		}
		bindingFloat, ok :=
			args[0].Interface().(interface {
				Get() (float64, error)
			})
		if !ok {
			return nil, object.Errorf("type error: expected binding.Float")
		}
		return risorwidget.
			NewProgressBarWithData(bindingFloat), nil
	})
	widgetAttrs.
		Define("NewProgressBarInfinite").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.Context, args ...object.Object) (object.Object,

		error) {
		if len(args) != 0 {
			return nil, object.
				Errorf("wrong number of arguments. got=%d, want=0",
					len(args))
		}
		return risorwidget.
			NewProgressBarInfinite(), nil
	})
	widgetAttrs.
		Define("NewSeparator").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewSeparator(), nil
	})
	widgetAttrs.
		Define("NewActivity").
		Variadic(
			"args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewActivity(), nil
	})
	widgetAttrs.
		Define("NewSlider").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(
					args))
		}
		min, err := object.AsFloat(args[0])
		if err != nil {
			return nil, err
		}
		max, err :=
			object.AsFloat(args[1])
		if err != nil {
			return nil,
				err

		}
		return risorwidget.NewSlider(min, max,

			obj.w), nil
	})
	widgetAttrs.
		Define("NewSliderWithData").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		min, err := object.AsFloat(args[0])
		if err != nil {
			return nil, err
		}
		max,
			err := object.AsFloat(args[1])
		if err != nil {
			return nil, err
		}
		bindingFloat, ok := args[2].
			Interface().(interface{ Get() (float64, error) })
		if !ok {
			return nil, object.Errorf("type error: expected binding.Float")
		}
		return risorwidget.NewSliderWithData(min, max,
			bindingFloat, obj.w), nil
	})
	widgetAttrs.
		Define("NewIcon").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		resourceName, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		icon,
			err := risorwidget.NewIcon(resourceName)
		if err !=
			nil {
			return nil, object.Errorf("error creating icon: %v",

				err)
		}
		return icon, nil
	})
	widgetAttrs.
		Define("NewHyperlink").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		urlStr,
			err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		hyperlink, err := risorwidget.
			NewHyperlink(text, urlStr, obj.w)
		if err != nil {
			return nil, object.Errorf("error creating hyperlink: %v",
				err)
		}
		return hyperlink, nil
	})
	widgetAttrs.
		Define("NewCard").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) != 3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		subtitle,
			err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		var content fyne.CanvasObject

		if args[2] != object.Nil {
			contentObj, ok := args[2].(IsCanvasObject)
			if !ok {
				return nil, object.Errorf("argument error: content must be a CanvasObject or nil, got %s",
					args[2].Type())
			}
			content = contentObj.CanvasObject()
		}
		return risorwidget.
			NewCard(title, subtitle, content), nil
	})
	widgetAttrs.
		Define("NewTable").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		pageSize,
			err := object.AsInt(args[1])
		if err != nil {
			return nil,

				err
		}
		return risorwidget.NewTable(title,

			int(pageSize), obj.w), nil
	})
	widgetAttrs.
		Define("NewRadioGroup").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {

		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",

				len(args))
		}
		options, err := object.AsStringSlice(args[0])
		if err != nil {
			return nil, err
		}
		cb, err := vmexec.NewCallback(ctx, "widget.NewRadioGroup",

			args[1])
		if err != nil {
			return nil,

				err
		}
		return risorwidget.NewRadioGroup(options, func(selected string) {
			cb.Fire(object.
				NewString(selected))
		}), nil
	})
	widgetAttrs.
		Define("NewAccordion").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		var items []*fynewidget.AccordionItem
		for i, arg := range args {
			item, ok := arg.Interface().(*fynewidget.AccordionItem)
			if !ok {
				return nil, object.Errorf("argument error: expected AccordionItem at position %d, got %s",
					i, arg.Type())
			}

			items = append(items, item)
		}
		return risorwidget.
			NewAccordion(items...), nil
	})
	widgetAttrs.
		Define("NewAccordionItem").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 3 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=3",
				len(args))
		}
		title, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		detail,
			err := object.AsString(args[1])
		if err != nil {

			return nil, err
		}
		var content fyne.CanvasObject

		if args[2] != object.Nil {
			contentObj, ok := args[2].(IsCanvasObject)
			if !ok {
				return nil, object.Errorf("argument error: content must be a CanvasObject or nil, got %s",
					args[2].Type())
			}
			content = contentObj.CanvasObject()
		}
		return risorwidget.NewAccordionItem(title, detail, content), nil
	})
	widgetAttrs.
		Define("NewToolbar").
		Variadic("args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		var items []fynewidget.ToolbarItem
		for i, arg := range args {
			item, ok := arg.Interface().(fynewidget.
				ToolbarItem)
			if !ok {
				return nil, object.Errorf("argument error: expected ToolbarItem at position %d, got %s",
					i, arg.Type())
			}
			items =
				append(items, item)
		}
		return risorwidget.
			NewToolbar(items...), nil
	})
	widgetAttrs.
		Define("NewToolbarAction").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		iconName, err := object.
			AsString(args[0])
		if err != nil {
			return nil, err
		}
		icon := risorwidget.GetThemeResource(iconName)
		if icon ==
			nil {
			return nil, object.Errorf("unknown icon resource: %s",

				iconName)
		}
		cb, err := vmexec.NewCallback(ctx, "widget.NewToolbarAction",
			args[1])
		if err != nil {
			return nil, err
		}
		return risorwidget.NewToolbarAction(icon, func() { cb.Fire() }), nil
	})
	widgetAttrs.
		Define("NewToolbarSeparator").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewToolbarSeparator(), nil
	})
	widgetAttrs.
		Define("NewToolbarSpacer").Variadic("args").Returns("nil").Impl(func(obj *Widget,
		ctx context.
			Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewToolbarSpacer(), nil
	})
	widgetAttrs.
		Define("NewList").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewList(obj.w),
			nil
	})
	widgetAttrs.
		Define("NewTree").Variadic("args").Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewTree(obj.w),
			nil
	})
	widgetAttrs.
		Define("NewGridWrap").
		Variadic(
			"args").
		Returns("nil").Impl(func(obj *Widget, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0",
				len(args))
		}
		return risorwidget.NewGridWrap(obj.w), nil
	})
}
