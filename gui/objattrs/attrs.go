// Package objattrs registers the CanvasObject verbs that every script-
// visible Fyne object shares (issue #38). Before this, Show/Hide/
// Refresh were copy-pasted per widget and Resize/Move/MinSize/Visible
// existed on only some types; scripts got an inconsistent canvas API.
// One generic registration fixes the inconsistency and feeds the
// Attrs() introspection used for docs tooling.
package objattrs

import (
	"context"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"github.com/uidbz/fynerisor/gui/guithread"
)

// HasCanvas is implemented by every wrapper around a fyne.CanvasObject
// (widgets, containers, canvas objects).
type HasCanvas interface {
	CanvasObject() fyne.CanvasObject
}

// Add registers Show, Hide, Refresh, Resize, Move, MinSize and Visible
// on the type's attribute registry. Types whose script API defines its
// own geometry verbs (PopUp's Move/Resize) register the base trio via
// Show/Hide/Refresh instead. Bodies marshal to the UI thread exactly
// like the hand-written versions they replace (issue #49).
func Add[T HasCanvas](r *object.AttrRegistry[T]) {
	Show(r)
	Hide(r)
	Refresh(r)
	Resize(r)
	Move(r)
	MinSize(r)
	Visible(r)
}

// Show registers the Show verb alone (see Add).
func Show[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Show").
		Doc("Make the object visible.").
		Returns("nil").
		Impl(func(o T, _ context.Context, _ ...object.Object) (object.Object, error) {
			guithread.Do(func() { o.CanvasObject().Show() })
			return object.Nil, nil
		})
}

// Hide registers the Hide verb alone (see Add).
func Hide[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Hide").
		Doc("Hide the object.").
		Returns("nil").
		Impl(func(o T, _ context.Context, _ ...object.Object) (object.Object, error) {
			guithread.Do(func() { o.CanvasObject().Hide() })
			return object.Nil, nil
		})
}

// Refresh registers the Refresh verb alone (see Add).
func Refresh[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Refresh").
		Doc("Redraw the object.").
		Returns("nil").
		Impl(func(o T, _ context.Context, _ ...object.Object) (object.Object, error) {
			guithread.Do(func() { o.CanvasObject().Refresh() })
			return object.Nil, nil
		})
}

// Resize registers the Resize verb alone (see Add).
func Resize[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Resize").
		Doc("Resize the object to width, height.").
		Args("width", "height").
		Returns("nil").
		Impl(func(o T, _ context.Context, args ...object.Object) (object.Object, error) {
			w, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			h, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() { o.CanvasObject().Resize(fyne.NewSize(float32(w), float32(h))) })
			return object.Nil, nil
		})
}

// Move registers the Move verb alone (see Add).
func Move[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Move").
		Doc("Move the object to x, y (relative to its parent canvas; layout containers reposition on the next refresh).").
		Args("x", "y").
		Returns("nil").
		Impl(func(o T, _ context.Context, args ...object.Object) (object.Object, error) {
			x, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			y, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() { o.CanvasObject().Move(fyne.NewPos(float32(x), float32(y))) })
			return object.Nil, nil
		})
}

// MinSize registers the MinSize verb alone (see Add).
func MinSize[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("MinSize").
		Doc("Minimum size of the object as {width, height}.").
		Returns("map").
		Impl(func(o T, _ context.Context, _ ...object.Object) (object.Object, error) {
			var sz fyne.Size
			guithread.Do(func() { sz = o.CanvasObject().MinSize() })
			return sizeObject(sz), nil
		})
}

// Visible registers the Visible verb alone (see Add).
func Visible[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Visible").
		Doc("Whether the object is currently visible.").
		Returns("bool").
		Impl(func(o T, _ context.Context, _ ...object.Object) (object.Object, error) {
			var vis bool
			guithread.Do(func() { vis = o.CanvasObject().Visible() })
			return object.NewBool(vis), nil
		})
}

// Append registers the child-list Append verb for layout containers whose
// Fyne instance is a *fyne.Container (VBox, HBox, Stack, Border, ...):
// Append(child, ...) adds to the container and refreshes, on the UI
// thread. Replaces the hand-written containerOps helper (issue #38).
func Append[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Append").
		Doc("Append child objects to the container.").
		Variadic("children").
		Returns("nil").
		Impl(func(o T, _ context.Context, args ...object.Object) (object.Object, error) {
			objs := make([]fyne.CanvasObject, 0, len(args))
			for i, a := range args {
				co, ok := a.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("Append: expected canvas object at argument %d, got %s", i, a.Type())
				}
				objs = append(objs, co)
			}
			guithread.Do(func() {
				c := o.CanvasObject().(*fyne.Container)
				c.Objects = append(c.Objects, objs...)
				c.Refresh()
			})
			return object.Nil, nil
		})
}

// Remove registers the child-list Remove verb (see Append).
func Remove[T HasCanvas](r *object.AttrRegistry[T]) {
	r.Define("Remove").
		Doc("Remove a child object from the container.").
		Arg("child").
		Returns("nil").
		Impl(func(o T, _ context.Context, args ...object.Object) (object.Object, error) {
			co, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("Remove: expected canvas object, got %s", args[0].Type())
			}
			guithread.Do(func() {
				c := o.CanvasObject().(*fyne.Container)
				for i, x := range c.Objects {
					if x == co {
						c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
						break
					}
				}
				c.Refresh()
			})
			return object.Nil, nil
		})
}

// ContainerOps registers the full layout-container verb set: the canvas
// verbs plus Append/Remove. Refresh comes from the canvas set.
func ContainerOps[T HasCanvas](r *object.AttrRegistry[T]) {
	Add(r)
	Append(r)
	Remove(r)
}

func sizeObject(sz fyne.Size) object.Object {
	return object.NewMap(map[string]object.Object{
		"width":  object.NewFloat(float64(sz.Width)),
		"height": object.NewFloat(float64(sz.Height)),
	})
}
