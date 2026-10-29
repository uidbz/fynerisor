package container

import (
	"context"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/gui/guithread"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// canvasArg unwraps a script argument into a Fyne canvas object.
func canvasArg(a object.Object) (fyne.CanvasObject, bool) {
	co, ok := a.Interface().(fyne.CanvasObject)
	return co, ok
}

// containerOps provides the shared Append/Remove/Refresh builtins for layout
// containers whose Fyne instance is a *fyne.Container (VBox, HBox, Stack,
// Center, Max, Padded, Grid). Children are mutated via Objects + Refresh,
// the same mechanism window.SetContent uses, always on the GUI thread.
//
// Call from a container's GetAttr:
//
//	if o, ok := containerOps("HBox", obj.instance, name); ok {
//		return o, true
//	}
func containerOps(owner string, inst fyne.CanvasObject, name string) (object.Object, bool) {
	switch name {
	case "Append", "Remove", "Refresh":
	default:
		return nil, false
	}
	c, ok := inst.(*fyne.Container)
	if !ok {
		return nil, false
	}

	switch name {
	case "Append":
		return object.NewBuiltin(owner+".Append", func(_ context.Context, args ...object.Object) (object.Object, error) {
			objs := make([]fyne.CanvasObject, 0, len(args))
			for i, a := range args {
				co, ok := canvasArg(a)
				if !ok {
					return nil, object.Errorf("Append: expected canvas object at argument %d, got %s", i, a.Type())
				}
				objs = append(objs, co)
			}
			guithread.Do(func() {
				c.Objects = append(c.Objects, objs...)
				c.Refresh()
			})
			return object.Nil, nil
		}), true

	case "Remove":
		return object.NewBuiltin(owner+".Remove", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			co, ok := canvasArg(args[0])
			if !ok {
				return nil, object.Errorf("Remove: expected canvas object, got %s", args[0].Type())
			}
			guithread.Do(func() {
				for i, o := range c.Objects {
					if o == co {
						c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
						break
					}
				}
				c.Refresh()
			})
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin(owner+".Refresh", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() { c.Refresh() })
			return object.Nil, nil
		}), true
	}
	return nil, false
}
