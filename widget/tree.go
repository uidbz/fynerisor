package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const TreeType object.Type = "widget.Tree"

type Tree struct {
	instance *widget.Tree
	w        WindowInterface
}

func NewTree(w WindowInterface) *Tree {
	tree := &Tree{w: w}
	tree.instance = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return []widget.TreeNodeID{}
		},
		func(uid widget.TreeNodeID) bool {
			return false
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
		},
	)
	return tree
}

func (t *Tree) CanvasObject() fyne.CanvasObject {
	return t.instance
}

func (t *Tree) Type() object.Type {
	return TreeType
}

func (t *Tree) Inspect() string {
	return fmt.Sprintf("Tree()")
}

func (t *Tree) Interface() interface{} {
	return t.instance
}

func (t *Tree) Equals(other object.Object) bool {
	return t == other
}

func (t *Tree) IsTruthy() bool {
	return true
}

func (t *Tree) Cost() int {
	return 0
}

func (t *Tree) MarshalJSON() ([]byte, error) {
	return nil, errors.New("tree cannot be marshaled to JSON")
}

func (t *Tree) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for tree: %v", opType), nil
}

func (t *Tree) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "ChildUIDs":
		return object.NewBuiltin("tree.ChildUIDs", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.ChildUIDs: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.ChildUIDs: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.ChildUIDs: unable to get call function")
			}

			t.instance.ChildUIDs = func(uid widget.TreeNodeID) []widget.TreeNodeID {
				result, err := callFunc(ctx, fn, []object.Object{object.NewString(string(uid))})
				if err != nil {
					return []widget.TreeNodeID{}
				}
				list, ok := result.(*object.List)
				if !ok {
					return []widget.TreeNodeID{}
				}
				items := list.Value()
				if len(items) == 0 {
					return []widget.TreeNodeID{}
				}
				uids := make([]widget.TreeNodeID, 0, len(items))
				for _, item := range items {
					if strObj, ok := item.(*object.String); ok {
						uids = append(uids, widget.TreeNodeID(strObj.Value()))
					}
				}
				return uids
			}
			return object.Nil, nil
		}), true

	case "IsBranch":
		return object.NewBuiltin("tree.IsBranch", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.IsBranch: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.IsBranch: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.IsBranch: unable to get call function")
			}

			t.instance.IsBranch = func(uid widget.TreeNodeID) bool {
				result, err := callFunc(ctx, fn, []object.Object{object.NewString(string(uid))})
				if err != nil {
					return false
				}
				return result.IsTruthy()
			}
			return object.Nil, nil
		}), true

	case "CreateNode":
		return object.NewBuiltin("tree.CreateNode", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.CreateNode: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.CreateNode: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.CreateNode: unable to get call function")
			}

			t.instance.CreateNode = func(branch bool) fyne.CanvasObject {
				result, err := callFunc(ctx, fn, []object.Object{object.NewBool(branch)})
				if err != nil {
					return widget.NewLabel("")
				}
				if obj, ok := result.(IsCanvasObject); ok {
					return obj.CanvasObject()
				}
				return widget.NewLabel("")
			}
			return object.Nil, nil
		}), true

	case "UpdateNode":
		return object.NewBuiltin("tree.UpdateNode", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.UpdateNode: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.UpdateNode: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.UpdateNode: unable to get call function")
			}

			t.instance.UpdateNode = func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
				// Wrap the CanvasObject
				var wrappedObj object.Object
				if label, ok := obj.(*widget.Label); ok {
					wrappedObj = &Label{instance: label}
				} else {
					wrappedObj = object.Nil
				}

				// Call synchronously to avoid race condition with IsBranch/ChildUIDs
				callFunc(ctx, fn, []object.Object{
					object.NewString(string(uid)),
					object.NewBool(branch),
					wrappedObj,
				})
			}
			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("tree.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.OnSelected: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.OnSelected: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.OnSelected: unable to get call function")
			}

			t.instance.OnSelected = func(uid widget.TreeNodeID) {
				// Call synchronously to avoid race condition
				callFunc(ctx, fn, []object.Object{object.NewString(string(uid))})
			}
			return object.Nil, nil
		}), true

	case "OnUnselected":
		return object.NewBuiltin("tree.OnUnselected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.OnUnselected: expected 1 argument (callback), got %d", len(args))
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return nil, fmt.Errorf("tree.OnUnselected: expected function, got %s", args[0].Type())
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return nil, fmt.Errorf("tree.OnUnselected: unable to get call function")
			}

			t.instance.OnUnselected = func(uid widget.TreeNodeID) {
				// Call synchronously to avoid race condition
				callFunc(ctx, fn, []object.Object{object.NewString(string(uid))})
			}
			return object.Nil, nil
		}), true

	case "OpenBranch":
		return object.NewBuiltin("tree.OpenBranch", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.OpenBranch: expected 1 argument (uid), got %d", len(args))
			}
			uid, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			t.instance.OpenBranch(widget.TreeNodeID(uid))
			return object.Nil, nil
		}), true

	case "CloseBranch":
		return object.NewBuiltin("tree.CloseBranch", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.CloseBranch: expected 1 argument (uid), got %d", len(args))
			}
			uid, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			t.instance.CloseBranch(widget.TreeNodeID(uid))
			return object.Nil, nil
		}), true

	case "IsBranchOpen":
		return object.NewBuiltin("tree.IsBranchOpen", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.IsBranchOpen: expected 1 argument (uid), got %d", len(args))
			}
			uid, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			return object.NewBool(t.instance.IsBranchOpen(widget.TreeNodeID(uid))), nil
		}), true

	case "Select":
		return object.NewBuiltin("tree.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.Select: expected 1 argument (uid), got %d", len(args))
			}
			uid, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			t.instance.Select(widget.TreeNodeID(uid))
			return object.Nil, nil
		}), true

	case "Unselect":
		return object.NewBuiltin("tree.Unselect", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.Unselect: expected 1 argument (uid), got %d", len(args))
			}
			uid, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			t.instance.Unselect(widget.TreeNodeID(uid))
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("tree.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("tree.Refresh: expected 0 arguments, got %d", len(args))
			}
			t.instance.Refresh()
			return object.Nil, nil
		}), true
	}

	return nil, false
}

func (t *Tree) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: tree object has no settable attribute %q", name)
}

func (t *Tree) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "ChildUIDs"},
		{Name: "IsBranch"},
		{Name: "CreateNode"},
		{Name: "UpdateNode"},
		{Name: "OnSelected"},
		{Name: "OnUnselected"},
		{Name: "OpenBranch"},
		{Name: "CloseBranch"},
		{Name: "IsBranchOpen"},
		{Name: "Select"},
		{Name: "Unselect"},
		{Name: "Refresh"},
	}
}
