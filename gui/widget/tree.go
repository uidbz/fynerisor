package widget

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
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

func (t *Tree) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for tree: %v", opType)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (t *Tree) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, t.getAttrUI)
}

var treeAttrs = object.NewAttrRegistry[*Tree](string(TreeType))

func (t *Tree) getAttrUI(name string) (object.Object, bool) {
	// Getter: called synchronously by Fyne during layout.

	// Getter: called synchronously by Fyne during layout.

	// Getter: called synchronously by Fyne during layout.

	// VM busy or run stopped: never pool a fake (issue #50).

	// Getter-like: called synchronously by Fyne during render.

	// Materialize any slot left by a busy-VM CreateNode call so
	// the script always receives its own node widget, then wrap
	// it so Risor can access it (issue #50).

	// VM still busy; a later refresh retries

	// Event: queued on the script's VM goroutine.

	// Event: queued on the script's VM goroutine.
	return treeAttrs.GetAttr(t, name)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (t *Tree) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, t.setAttrUI)
}

func (t *Tree) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: tree object has no settable attribute %q", name)
}

func (t *Tree) Attrs() []object.AttrSpec {
	return treeAttrs.Specs()

}
func init() {
	objattrs.Add(treeAttrs)
	treeAttrs.Define("ChildUIDs").Args("arg1").
		Returns(
			"nil",
		).Impl(func(t *Tree, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.ChildUIDs",
			args[0])
		if err != nil {
			return nil, err
		}
		cb.RefreshWhenBusy(t.refresh)

		t.
			instance.ChildUIDs = func(uid widget.TreeNodeID) []widget.TreeNodeID {
			result, err :=
				cb.Call(object.NewString(
					string(uid)))
			if err != nil {
				cb.Report(err)
				t.cacheMu.Lock()
				defer t.cacheMu.Unlock()
				return t.lastChildren[uid]
			}
			list, ok := result.(*object.List)
			if !ok {
				cb.Report(fmt.Errorf("expected list of node IDs, got %s",

					result.Type()))
				return []widget.TreeNodeID{}
			}
			items := list.Value()
			uids := make([]widget.TreeNodeID, 0, len(
				items))
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
		return object.
			Nil, nil
	})
	treeAttrs.Define("IsBranch").Args("arg1").Returns("nil").Impl(func(t *Tree, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.IsBranch",
			args[0])
		if err != nil {
			return nil,
				err
		}
		cb.RefreshWhenBusy(t.refresh)
		t.
			instance.
			IsBranch = func(uid widget.TreeNodeID) bool {
			result,
				err := cb.Call(object.NewString(string(uid)))
			t.
				cacheMu.
				Lock()
			defer t.cacheMu.Unlock()
			if err != nil {
				cb.Report(err)
				return t.lastBranch[uid]
			}
			t.lastBranch[uid] = result.IsTruthy()
			return t.lastBranch[uid]
		}
		return object.Nil, nil
	})
	treeAttrs.Define("CreateNode").Args(
		"arg1").
		Returns(
			"nil").Impl(func(t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.CreateNode",
			args[0])
		if err != nil {
			return nil, err
		}
		cb.RefreshWhenBusy(t.refresh)

		t.createCb = cb
		t.instance.CreateNode = func(branch bool) fyne.
			CanvasObject {
			inner,
				ok := callTemplate(cb, object.
				NewBool(branch))
			if !ok {
				return newTemplateSlot()
			}
			return inner
		}
		return object.Nil, nil
	})
	treeAttrs.Define("UpdateNode").Args(
		"arg1").
		Returns(
			"nil").Impl(func(t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.UpdateNode",
			args[0])
		if err != nil {
			return nil, err
		}
		cb.RefreshWhenBusy(t.refresh)

		t.instance.UpdateNode = func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			node, ok := resolveTemplateItem(node, func() (fyne.CanvasObject, bool) {
				return callTemplate(t.createCb, object.NewBool(branch))
			})
			if !ok {
				return
			}
			_, err := cb.Call(object.NewString(string(uid)), object.NewBool(branch), wrapCanvasObject(node))
			cb.Report(err)
		}
		return object.Nil, nil
	})
	treeAttrs.Define("OnSelected").Args(
		"arg1").
		Returns(
			"nil").Impl(func(t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.OnSelected",
			args[0])
		if err != nil {
			return nil, err
		}
		t.instance.OnSelected = func(
			uid widget.TreeNodeID) {
			cb.Fire(object.NewString(string(uid)))
		}
		return object.Nil,
			nil
	})
	treeAttrs.Define("OnUnselected").Args("arg1").Returns("nil").Impl(func(
		t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "tree.OnUnselected",
			args[0])
		if err != nil {
			return nil, err
		}
		t.instance.OnUnselected =

			func(uid widget.TreeNodeID) {
				cb.Fire(object.
					NewString(string(
						uid)))
			}
		return object.
			Nil, nil
	})
	treeAttrs.Define("OpenBranch").Args(
		"arg1").
		Returns(
			"nil").Impl(func(t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		uid, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, err
		}
		t.instance.OpenBranch(widget.TreeNodeID(uid))
		return object.
			Nil, nil
	})
	treeAttrs.Define("CloseBranch").Args("arg1").Returns("nil").Impl(func(t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		uid, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		t.instance.
			CloseBranch(widget.TreeNodeID(uid))
		return object.Nil, nil
	})
	treeAttrs.Define("IsBranchOpen").Args("arg1").Returns("nil").Impl(func(
		t *Tree, ctx context.Context, args ...object.Object) (object.Object, error) {
		uid, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		return object.
			NewBool(t.instance.IsBranchOpen(widget.
				TreeNodeID(uid))), nil
	})
	treeAttrs.Define("Select").Doc("Select the tree node with the given UID.").
		Args("arg1").Returns("nil").Impl(func(t *Tree,
		ctx context.Context, args ...object.Object) (object.
		Object, error) {
		uid, err := object.AsString(args[0])
		if err != nil {
			return nil,
				err
		}
		t.instance.Select(widget.TreeNodeID(
			uid))
		return object.Nil, nil
	})
	treeAttrs.Define("Unselect").Doc("Unselect the tree node with the given UID.").
		Args("arg1").Returns("nil").Impl(func(t *Tree, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		uid, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, err
		}
		t.instance.Unselect(widget.TreeNodeID(uid))
		return object.
			Nil, nil
	})
}
