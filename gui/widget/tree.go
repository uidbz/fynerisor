package widget

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"
)

const TreeType object.Type = "widget.Tree"

type Tree struct {
	instance *widget.Tree
	w        WindowInterface

	// createCb is the registered node template callback; UI thread only
	// (see List). UpdateNode uses it to materialize slots left by a
	// busy-VM CreateNode call (issue #50).
	createCb *vmexec.Callback

	// Last-known getter results per node, reused while the script's VM is
	// busy (vmexec.ErrBusy) so the tree does not collapse; the tree is
	// refreshed once the VM is idle (Callback.RefreshWhenBusy).
	cacheMu      sync.Mutex
	lastChildren map[widget.TreeNodeID][]widget.TreeNodeID
	lastBranch   map[widget.TreeNodeID]bool
}

func NewTree(w WindowInterface) *Tree {
	tree := &Tree{
		w:            w,
		lastChildren: map[widget.TreeNodeID][]widget.TreeNodeID{},
		lastBranch:   map[widget.TreeNodeID]bool{},
	}
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

// refresh re-renders the tree with fresh getter results; scheduled by the
// getters' callbacks when they hit VM contention.
func (t *Tree) refresh() {
	guithread.Do(t.instance.Refresh)
}

func (t *Tree) CanvasObject() fyne.CanvasObject {
	return t.instance
}

func (t *Tree) Type() object.Type {
	return TreeType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (t *Tree) Inspect() string {
	return guithread.Inspect(t.inspectUI)
}

func (t *Tree) inspectUI() string {
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
	return nil, object.Errorf("eval error: unsupported operation for tree: %v", opType)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (t *Tree) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, t.getAttrUI)
}

func (t *Tree) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "ChildUIDs":
		return object.NewBuiltin("tree.ChildUIDs", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.ChildUIDs: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.ChildUIDs", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(t.refresh)

			// Getter: called synchronously by Fyne during layout.
			t.instance.ChildUIDs = func(uid widget.TreeNodeID) []widget.TreeNodeID {
				result, err := cb.Call(object.NewString(string(uid)))
				if err != nil {
					cb.Report(err)
					t.cacheMu.Lock()
					defer t.cacheMu.Unlock()
					return t.lastChildren[uid]
				}
				list, ok := result.(*object.List)
				if !ok {
					cb.Report(fmt.Errorf("expected list of node IDs, got %s", result.Type()))
					return []widget.TreeNodeID{}
				}
				items := list.Value()
				uids := make([]widget.TreeNodeID, 0, len(items))
				for _, item := range items {
					if strObj, ok := item.(*object.String); ok {
						uids = append(uids, widget.TreeNodeID(strObj.Value()))
					}
				}
				t.cacheMu.Lock()
				t.lastChildren[uid] = uids
				t.cacheMu.Unlock()
				return uids
			}
			return object.Nil, nil
		}), true

	case "IsBranch":
		return object.NewBuiltin("tree.IsBranch", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.IsBranch: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.IsBranch", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(t.refresh)

			// Getter: called synchronously by Fyne during layout.
			t.instance.IsBranch = func(uid widget.TreeNodeID) bool {
				result, err := cb.Call(object.NewString(string(uid)))
				t.cacheMu.Lock()
				defer t.cacheMu.Unlock()
				if err != nil {
					cb.Report(err)
					return t.lastBranch[uid]
				}
				t.lastBranch[uid] = result.IsTruthy()
				return t.lastBranch[uid]
			}
			return object.Nil, nil
		}), true

	case "CreateNode":
		return object.NewBuiltin("tree.CreateNode", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.CreateNode: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.CreateNode", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(t.refresh)
			t.createCb = cb

			// Getter: called synchronously by Fyne during layout.
			t.instance.CreateNode = func(branch bool) fyne.CanvasObject {
				inner, ok := callTemplate(cb, object.NewBool(branch))
				if !ok {
					// VM busy or run stopped: never pool a fake (issue #50).
					return newTemplateSlot()
				}
				return inner
			}
			return object.Nil, nil
		}), true

	case "UpdateNode":
		return object.NewBuiltin("tree.UpdateNode", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.UpdateNode: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.UpdateNode", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(t.refresh)

			// Getter-like: called synchronously by Fyne during render.
			t.instance.UpdateNode = func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
				// Materialize any slot left by a busy-VM CreateNode call so
				// the script always receives its own node widget, then wrap
				// it so Risor can access it (issue #50).
				node, ok := resolveTemplateItem(node, func() (fyne.CanvasObject, bool) {
					return callTemplate(t.createCb, object.NewBool(branch))
				})
				if !ok {
					return // VM still busy; a later refresh retries
				}
				_, err := cb.Call(object.NewString(string(uid)), object.NewBool(branch), wrapCanvasObject(node))
				cb.Report(err)
			}
			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("tree.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.OnSelected: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.OnSelected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			t.instance.OnSelected = func(uid widget.TreeNodeID) {
				cb.Fire(object.NewString(string(uid)))
			}
			return object.Nil, nil
		}), true

	case "OnUnselected":
		return object.NewBuiltin("tree.OnUnselected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("tree.OnUnselected: expected 1 argument (callback), got %d", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "tree.OnUnselected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			t.instance.OnUnselected = func(uid widget.TreeNodeID) {
				cb.Fire(object.NewString(string(uid)))
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
	case "Hide":
		return object.NewBuiltin("widget.Tree.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				t.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Tree.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				t.instance.Show()
			})
			return object.Nil, nil
		}), true
	}

	return nil, false
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (t *Tree) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, t.setAttrUI)
}

func (t *Tree) setAttrUI(name string, value object.Object) error {
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
