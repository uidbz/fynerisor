package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	fynewidget "fyne.io/fyne/v2/widget"

	risorwidget "github.com/uidbz/fynerisor/gui/widget"
	timemodule "github.com/uidbz/fynerisor/modules/time"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const WidgetType object.Type = "widget"

// Widget is a factory object that provides widget creation functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating various
// Fyne widgets like buttons, labels, entries, forms, tables, and more.
//
// Available in Risor scripts as the global 'widget' object.
//
// Example usage in Risor:
//
//	let btn = widget.NewButton("Click", () => { print("clicked") })
//	let lbl = widget.NewLabel("Hello")
//	let entry = widget.NewEntry()

// Widget is a factory object that provides widget creation functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating
// UI components like buttons, labels, forms, tables, and more.
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

func (obj *Widget) Inspect() string {
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

func (obj *Widget) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", WidgetType, name)
}

func (obj *Widget) GetAttr(name string) (object.Object, bool) {
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			// Create button with closure that captures fn by value
			button := func(f *object.Closure) object.Object {
				return risorwidget.NewButton(name, func() {
					obj.w.Do(func() {
						_, err := safeCall(callFunc, ctx, f, []object.Object{})
						if err != nil {
							fmt.Println(err)
						}
					})
				})
			}(fn)

			return button, nil
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			// Create check with closure that captures fn by value
			chk := func(f *object.Closure) object.Object {
				return risorwidget.NewCheck(label, func(changed bool) {
					obj.w.Do(func() {
						val := object.NewBool(changed)
						safeCall(callFunc, ctx, f, []object.Object{val})
					})
				})
			}(fn)

			return chk, nil
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			// Create check group with closure that captures fn by value
			chkGroup := func(f *object.Closure) object.Object {
				return risorwidget.NewCheckGroup(options, func(checked []string) {
					obj.w.Do(func() {
						list := object.NewStringList(checked)
						safeCall(callFunc, ctx, f, []object.Object{list})
					})
				})
			}(fn)

			return chkGroup, nil
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			// Create select with closure that captures fn by value
			chkGroup := func(f *object.Closure) object.Object {
				return risorwidget.NewSelect(options, func(selected string) {
					obj.w.Do(func() {
						selectedObj := object.NewString(selected)
						safeCall(callFunc, ctx, f, []object.Object{selectedObj})
					})
				})
			}(fn)

			return chkGroup, nil
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
			if len(args) >= 1 {
				// Try to extract time.Time from the TimeObject
				if timeVal, ok := args[0].Interface().(time.Time); ok {
					startDate = timeVal
				} else {
					startDate = time.Now()
				}
			} else {
				startDate = time.Now()
			}

			// Second argument: callback function (optional)
			if len(args) == 2 {
				fn, ok := args[1].(*object.Closure)
				if !ok {
					return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
				}
				callFunc, ok := object.GetCallFunc(ctx)
				if !ok {
					return object.Errorf("calendar: unable to get call function"), nil
				}

				// Create callback with closure that captures fn by value
				onChanged = func(f *object.Closure) func(time.Time) {
					return func(selected time.Time) {
						obj.w.Do(func() {
							timeObj := timemodule.NewTimeObject(selected)
							safeCall(callFunc, ctx, f, []object.Object{timeObj})
						})
					}
				}(fn)
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
		return object.NewBuiltin("widget.NewMuliLineEntry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			return risorwidget.NewMultiLineEntry(), nil
		}), true

	case "NewForm":
		return object.NewBuiltin("widget.NewForm", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []*fynewidget.FormItem

			for i, x := range args {
				if x.Type() == object.LIST {
					items, err := object.AsList(args[0])
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("radiogroup: unable to get call function"), nil
			}

			// Create radio group with closure that captures fn by value
			radioGroup := func(f *object.Closure) object.Object {
				return risorwidget.NewRadioGroup(options, func(selected string) {
					obj.w.Do(func() {
						selectedObj := object.NewString(selected)
						safeCall(callFunc, ctx, f, []object.Object{selectedObj})
					})
				})
			}(fn)

			return radioGroup, nil
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
			// detail is currently unused in NewAccordionItem, but we accept it for API consistency
			_, err = object.AsString(args[1])
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
			return risorwidget.NewAccordionItem(title, "", content), nil
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
			fn, ok := args[1].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("toolbar action: unable to get call function"), nil
			}

			action := risorwidget.NewToolbarAction(icon, func() {
				obj.w.Do(func() {
					_, err := safeCall(callFunc, ctx, fn, []object.Object{})
					if err != nil {
						fmt.Println(err)
					}
				})
			})

			return action, nil
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

func (w *Window) Do(fn func()) {
	w.functionCalls <- fn
}

func NewWidget(w *Window) *Widget {
	return &Widget{w: w}
}
