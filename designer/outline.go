package designer

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

type outlineView struct {
	ed      *Editor
	tree    *widget.Tree
	box     *fyne.Container
	syncing bool
}

func newOutline(e *Editor) *outlineView {
	o := &outlineView{ed: e}
	o.tree = widget.NewTree(o.childUIDs, o.isBranch, o.create, o.update)
	o.tree.OnSelected = func(uid string) {
		if o.syncing || uid == "" {
			return
		}
		o.ed.Select(uid)
	}
	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { o.ed.reorderSelection(-1) })
	down := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { o.ed.reorderSelection(1) })
	hide := widget.NewButtonWithIcon("", theme.VisibilityIcon(), o.toggleHidden)
	lock := widget.NewButtonWithIcon("", theme.ConfirmIcon(), o.toggleLocked)
	del := widget.NewButtonWithIcon("", theme.DeleteIcon(), o.ed.deleteSelection)
	tools := container.NewHBox(up, down, hide, lock, del)
	o.box = container.NewBorder(tools, nil, nil, nil, container.NewVScroll(o.tree))
	return o
}

func (o *outlineView) object() fyne.CanvasObject { return o.box }

func (o *outlineView) childUIDs(uid string) []string {
	if o.ed.doc.Root == nil {
		return nil
	}
	if uid == "" {
		return []string{o.ed.doc.Root.ID}
	}
	n, ok := o.ed.doc.Find(uid)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(n.Children))
	for _, c := range n.Children {
		out = append(out, c.ID)
	}
	return out
}

func (o *outlineView) isBranch(uid string) bool {
	if uid == "" {
		return true
	}
	n, ok := o.ed.doc.Find(uid)
	if !ok {
		return false
	}
	return len(n.Children) > 0 || o.ed.view.ContainerKind(uid) != scene.Leaf
}

func (o *outlineView) create(bool) fyne.CanvasObject { return widget.NewLabel("") }

func (o *outlineView) update(uid string, _ bool, obj fyne.CanvasObject) {
	l, ok := obj.(*widget.Label)
	if !ok {
		return
	}
	if uid == "" {
		l.SetText("Page · " + o.ed.doc.Title)
		return
	}
	n, ok := o.ed.doc.Find(uid)
	if !ok {
		l.SetText(uid)
		return
	}
	marks := ""
	if n.Hidden {
		marks += " ◌"
	}
	if n.Locked {
		marks += " 🔒"
	}
	name := n.Name
	if name == "" {
		name = n.ID
	}
	l.SetText(fmt.Sprintf("%s · %s%s", n.Type, name, marks))
}

func (o *outlineView) refresh() {
	if o.ed.doc.Root == nil {
		return
	}
	o.tree.Refresh()
	o.tree.OpenAllBranches()
	o.syncSelection()
}

func (o *outlineView) syncSelection() {
	if o.ed.primary == "" {
		return
	}
	o.syncing = true
	o.tree.Select(o.ed.primary)
	o.syncing = false
	o.tree.Refresh()
}

func (o *outlineView) toggleHidden() {
	var ops []scene.Op
	for _, id := range o.ed.selectedIDs() {
		n, ok := o.ed.doc.Find(id)
		if !ok {
			continue
		}
		ops = append(ops, scene.SetFieldOp(id, "hidden", n.Hidden, !n.Hidden))
	}
	if o.ed.apply("visibility", "", ops...) {
		o.refresh()
	}
}

func (o *outlineView) toggleLocked() {
	var ops []scene.Op
	for _, id := range o.ed.selectedIDs() {
		n, ok := o.ed.doc.Find(id)
		if !ok || id == o.ed.doc.Root.ID {
			continue
		}
		ops = append(ops, scene.SetFieldOp(id, "locked", n.Locked, !n.Locked))
	}
	if o.ed.apply("lock", "", ops...) {
		o.refresh()
	}
}
