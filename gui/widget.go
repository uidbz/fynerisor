package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"
	"os"
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

func (obj *Widget) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'Widget'")
}

func (obj *Widget) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", WidgetType, opType), nil
}

func (obj *Widget) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Widget) Attrs() []object.AttrSpec {
	return nil
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

func (obj *Widget) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "NewButton":
		return object.NewBuiltin("widget.NewButton", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			name, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewButton", args[1])
			if err != nil {
				return nil, err
			}

			// Event: the tap handler is queued on the script's VM goroutine.
			return risorwidget.NewButton(name, func() { cb.Fire() }), nil
		}), true

	case "NewCheck":
		return object.NewBuiltin("widget.NewCheck", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			label, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewCheck", args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewCheck(label, func(changed bool) {
				cb.Fire(object.NewBool(changed))
			}), nil
		}), true

	case "NewCheckWithData":
		return object.NewBuiltin("widget.NewCheckWithData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			label, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			// Get the binding.Bool object
			bindingBool, ok := args[1].Interface().(interface{ Get() (bool, error) })
			if !ok {
				return object.Errorf("type error: expected binding.Bool"), nil
			}

			return risorwidget.NewCheckWithData(label, bindingBool), nil
		}), true

	case "NewCheckGroup":
		return object.NewBuiltin("widget.NewCheckGroup", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			options, err := object.AsStringSlice(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewCheckGroup", args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewCheckGroup(options, func(checked []string) {
				cb.Fire(object.NewStringList(checked))
			}), nil
		}), true

	case "NewSelect":
		return object.NewBuiltin("widget.NewSelect", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			options, err := object.AsStringSlice(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewSelect", args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewSelect(options, func(selected string) {
				cb.Fire(object.NewString(selected))
			}), nil
		}), true

	case "NewSelectEntry":
		return object.NewBuiltin("widget.NewSelectEntry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			options, err := object.AsStringSlice(args[0])
			if err != nil {
				return nil, err
			}

			selectEntry := risorwidget.NewSelectEntry(obj.w, options, nil)
			return selectEntry, nil
		}), true

	case "NewDateEntry":
		return object.NewBuiltin("widget.NewDateEntry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			dateEntry := risorwidget.NewDateEntry(obj.w)
			return dateEntry, nil
		}), true

	case "NewCalendar":
		return object.NewBuiltin("widget.NewCalendar", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) > 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=0, 1, or 2", len(args)), nil
			}

			var startDate time.Time
			var onChanged func(time.Time)

			// First argument: start date (optional)
			if len(args) >= 1 && args[0] != object.Nil && args[0].Type() != object.NIL {
				// Try to extract time.Time from the TimeObject
				if timeVal, ok := args[0].Interface().(time.Time); ok {
					startDate = timeVal
				} else {
					return object.Errorf("argument error: expected time.Time or nil, got %s", args[0].Type()), nil
				}
			} else {
				startDate = time.Now()
			}

			// Second argument: callback function (optional)
			if len(args) == 2 {
				cb, err := vmexec.NewCallback(ctx, "widget.NewCalendar", args[1])
				if err != nil {
					return nil, err
				}
				onChanged = func(selected time.Time) {
					cb.Fire(timemodule.NewTimeObject(selected))
				}
			}

			calendar := risorwidget.NewCalendar(obj.w, startDate, onChanged)
			return calendar, nil
		}), true

	case "NewEntry":
		return object.NewBuiltin("widget.NewEntry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			return risorwidget.NewEntry(obj.w), nil
		}), true

	case "NewEntryWithData":
		return object.NewBuiltin("widget.NewEntryWithData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			// Get the binding.String object
			bindingStr, ok := args[0].Interface().(interface{ Get() (string, error) })
			if !ok {
				return object.Errorf("type error: expected binding.String"), nil
			}

			return risorwidget.NewEntryWithData(bindingStr, obj.w), nil
		}), true

	case "NewMultiLineEntry":
		return object.NewBuiltin("widget.NewMultiLineEntry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			return risorwidget.NewMultiLineEntry(obj.w), nil
		}), true

	case "NewForm":
		return object.NewBuiltin("widget.NewForm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []*fynewidget.FormItem

			for i, x := range args {
				if x.Type() == object.LIST {
					items, err := object.AsList(x)
					if err != nil {
						return object.Errorf("NewForm: Error: %s", err.Error()), nil
					}
					for j, y := range items.Value() {
						o, ok := y.Interface().(*fynewidget.FormItem)
						if !ok {
							return object.Errorf("NewForm: Wrong type, expected FormItem within list at position: %d", j), nil
						}
						objects = append(objects, o)
					}
				} else {
					o, ok := x.Interface().(*fynewidget.FormItem)
					if !ok {
						return object.Errorf("NewForm: Wrong type, expected FormItem at argument: %d", i), nil
					}
					objects = append(objects, o)
				}
			}

			return risorwidget.NewForm(objects...), nil
		}), true

	case "NewFormItem":
		return object.NewBuiltin("widget.NewFormItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			label, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			o, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewFormItem: Wrong type, expected CanvasObject at argument: %d", 1), nil
			}

			return risorwidget.NewFormItem(label, o), nil
		}), true

	case "NewLabel":
		return object.NewBuiltin("widget.NewLabel", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewLabel(text), nil
		}), true

	case "NewLabelWithData":
		return object.NewBuiltin("widget.NewLabelWithData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			// Get the binding.String object
			bindingStr, ok := args[0].Interface().(interface{ Get() (string, error) })
			if !ok {
				return object.Errorf("type error: expected binding.String"), nil
			}

			return risorwidget.NewLabelWithData(bindingStr), nil
		}), true

	case "NewLog":
		return object.NewBuiltin("widget.NewLog", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			maxItems, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			return risorwidget.NewLog(int(maxItems)), nil
		}), true

	case "NewMarkdown":
		return object.NewBuiltin("widget.NewMarkdown", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			markdown, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewMarkdown(markdown), nil
		}), true

	case "NewProgressBar":
		return object.NewBuiltin("widget.NewProgressBar", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewProgressBar(), nil
		}), true

	case "NewProgressBarWithData":
		return object.NewBuiltin("widget.NewProgressBarWithData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			// Get the binding.Float object
			bindingFloat, ok := args[0].Interface().(interface{ Get() (float64, error) })
			if !ok {
				return object.Errorf("type error: expected binding.Float"), nil
			}

			return risorwidget.NewProgressBarWithData(bindingFloat), nil
		}), true

	case "NewProgressBarInfinite":
		return object.NewBuiltin("widget.NewProgressBarInfinite", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewProgressBarInfinite(), nil
		}), true

	case "NewSeparator":
		return object.NewBuiltin("widget.NewSeparator", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewSeparator(), nil
		}), true

	case "NewActivity":
		return object.NewBuiltin("widget.NewActivity", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewActivity(), nil
		}), true

	case "NewSlider":
		return object.NewBuiltin("widget.NewSlider", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			min, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			max, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			return risorwidget.NewSlider(min, max, obj.w), nil
		}), true

	case "NewSliderWithData":
		return object.NewBuiltin("widget.NewSliderWithData", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			min, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			max, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			// Get the binding.Float object
			bindingFloat, ok := args[2].Interface().(interface{ Get() (float64, error) })
			if !ok {
				return object.Errorf("type error: expected binding.Float"), nil
			}

			return risorwidget.NewSliderWithData(min, max, bindingFloat, obj.w), nil
		}), true

	case "NewIcon":
		return object.NewBuiltin("widget.NewIcon", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			resourceName, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			icon, err := risorwidget.NewIcon(resourceName)
			if err != nil {
				return object.Errorf("error creating icon: %v", err), nil
			}
			return icon, nil
		}), true

	case "NewHyperlink":
		return object.NewBuiltin("widget.NewHyperlink", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			urlStr, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}
			hyperlink, err := risorwidget.NewHyperlink(text, urlStr, obj.w)
			if err != nil {
				return object.Errorf("error creating hyperlink: %v", err), nil
			}
			return hyperlink, nil
		}), true

	case "NewCard":
		return object.NewBuiltin("widget.NewCard", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			subtitle, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}
			var content fyne.CanvasObject
			if args[2] != object.Nil {
				contentObj, ok := args[2].(IsCanvasObject)
				if !ok {
					return object.Errorf("argument error: content must be a CanvasObject or nil, got %s", args[2].Type()), nil
				}
				content = contentObj.CanvasObject()
			}
			return risorwidget.NewCard(title, subtitle, content), nil
		}), true

	case "NewTable":
		return object.NewBuiltin("widget.NewTable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			pageSize, err := object.AsInt(args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewTable(title, int(pageSize), obj.w), nil
		}), true

	case "NewRadioGroup":
		return object.NewBuiltin("widget.NewRadioGroup", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			options, err := object.AsStringSlice(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewRadioGroup", args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewRadioGroup(options, func(selected string) {
				cb.Fire(object.NewString(selected))
			}), nil
		}), true

	case "NewAccordion":
		return object.NewBuiltin("widget.NewAccordion", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			var items []*fynewidget.AccordionItem
			for i, arg := range args {
				item, ok := arg.Interface().(*fynewidget.AccordionItem)
				if !ok {
					return object.Errorf("argument error: expected AccordionItem at position %d, got %s", i, arg.Type()), nil
				}
				items = append(items, item)
			}
			return risorwidget.NewAccordion(items...), nil
		}), true

	case "NewAccordionItem":
		return object.NewBuiltin("widget.NewAccordionItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return object.Errorf("wrong number of arguments. got=%d, want=3", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			detail, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}
			var content fyne.CanvasObject
			if args[2] != object.Nil {
				contentObj, ok := args[2].(IsCanvasObject)
				if !ok {
					return object.Errorf("argument error: content must be a CanvasObject or nil, got %s", args[2].Type()), nil
				}
				content = contentObj.CanvasObject()
			}
			return risorwidget.NewAccordionItem(title, detail, content), nil
		}), true

	case "NewToolbar":
		return object.NewBuiltin("widget.NewToolbar", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			var items []fynewidget.ToolbarItem
			for i, arg := range args {
				item, ok := arg.Interface().(fynewidget.ToolbarItem)
				if !ok {
					return object.Errorf("argument error: expected ToolbarItem at position %d, got %s", i, arg.Type()), nil
				}
				items = append(items, item)
			}
			return risorwidget.NewToolbar(items...), nil
		}), true

	case "NewToolbarAction":
		return object.NewBuiltin("widget.NewToolbarAction", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			iconName, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			icon := risorwidget.GetThemeResource(iconName)
			if icon == nil {
				return object.Errorf("unknown icon resource: %s", iconName), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.NewToolbarAction", args[1])
			if err != nil {
				return nil, err
			}

			return risorwidget.NewToolbarAction(icon, func() { cb.Fire() }), nil
		}), true

	case "NewToolbarSeparator":
		return object.NewBuiltin("widget.NewToolbarSeparator", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewToolbarSeparator(), nil
		}), true

	case "NewToolbarSpacer":
		return object.NewBuiltin("widget.NewToolbarSpacer", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewToolbarSpacer(), nil
		}), true

	case "NewList":
		return object.NewBuiltin("widget.NewList", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewList(obj.w), nil
		}), true

	case "NewTree":
		return object.NewBuiltin("widget.NewTree", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewTree(obj.w), nil
		}), true

	case "NewGridWrap":
		return object.NewBuiltin("widget.NewGridWrap", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return risorwidget.NewGridWrap(obj.w), nil
		}), true

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
// after Stop, or when the queue is full, the callback is dropped (the
// latter two with a message on stderr).
func (w *Window) queueCallback(what string, fn func()) {
	r := w.run.Load()
	if r == nil {
		fmt.Fprintf(os.Stderr, "fynerisor: %s called before Execute; callback dropped\n", what)
		return
	}
	_ = r.Enqueue(what, fn)
}

// Do queues fn on the current execution's callback queue so it runs on the
// run's VM goroutine with exclusive access to the Risor VM. It never blocks
// the caller (issue #10); see queueCallback.
func (w *Window) Do(fn func()) {
	w.queueCallback("window.Do", fn)
}

func NewWidget(w *Window) *Widget {
	return &Widget{w: w}
}
