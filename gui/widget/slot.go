package widget

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/fynerisor/gui/vmexec"
)

// Fyne pools the object a template getter (List/GridWrap CreateItem, Tree
// CreateNode, the flex table's CreateCell) returns and never asks for a
// fresh one while the row lives. So a template request must never be
// answered with a fake: when the VM is busy an empty Label would be pooled
// forever, and every later UpdateItem would hand the script that Label
// where it expects its own widget (issue #50).
//
// Instead, a busy template getter returns a templateSlot: a real one-child
// container, invisible until the script's widget is installed. The update
// getter for the same row installs it the next time the VM can be entered
// and then hands the script the real widget. The RefreshWhenBusy machinery
// keeps re-running the getters until that happens.

// templateSlot stands in for a script widget that could not be created
// because the VM was busy or the run was stopped. Fyne holds the slot in
// its row pool; the real widget is adopted later, in place. Its methods
// are used only on Fyne's UI thread (it is only ever touched from render
// getters and materialize), so it needs no locking.
type templateSlot struct {
	widget.BaseWidget
	content fyne.CanvasObject // nil until adopted
}

func newTemplateSlot() *templateSlot {
	s := &templateSlot{}
	s.ExtendBaseWidget(s)
	return s
}

func (s *templateSlot) CreateRenderer() fyne.WidgetRenderer {
	if s.content == nil {
		return emptySlotRenderer{} // zero MinSize, nothing painted
	}
	return widget.NewSimpleRenderer(s.content)
}

// MinSize reports the contained widget's minimum size once materialized.
// Fyne's GridWrap re-measures the template on every refresh, so a slot
// must not shrink the measured row once its real widget is in place.
func (s *templateSlot) MinSize() fyne.Size {
	if s.content != nil {
		return s.content.MinSize()
	}
	return fyne.NewSize(0, 0)
}

// emptySlotRenderer keeps an unmaterialized slot invisible.
type emptySlotRenderer struct{}

func (emptySlotRenderer) Layout(fyne.Size)             {}
func (emptySlotRenderer) MinSize() fyne.Size           { return fyne.NewSize(0, 0) }
func (emptySlotRenderer) Objects() []fyne.CanvasObject { return nil }
func (emptySlotRenderer) Refresh()                     {}
func (emptySlotRenderer) Destroy()                     {}

// resolveTemplateItem returns the object to hand the script's update
// getter: the pooled object itself, or the script widget inside a slot
// (installing it first if needed). ok=false means the template callback
// still cannot run — the VM is busy or the run stopped — so the caller
// must skip this update; a later refresh will retry.
func resolveTemplateItem(item fyne.CanvasObject, fill func() (fyne.CanvasObject, bool)) (fyne.CanvasObject, bool) {
	slot, ok := item.(*templateSlot)
	if !ok {
		return item, true
	}
	if slot.content == nil {
		inner, filled := fill()
		if !filled {
			return nil, false
		}
		slot.content = inner
		slot.Refresh()
	}
	return slot.content, true
}

// callTemplate invokes a template callback and unwraps its widget. ok=false
// is returned only for quiet contention errors (ErrBusy/ErrStopped), meaning
// nothing was created and the caller must retry later; script errors and
// wrong return types are reported and answered with an empty Label, exactly
// as scripts have always seen them.
func callTemplate(cb *vmexec.Callback, args ...object.Object) (fyne.CanvasObject, bool) {
	if cb == nil { // UpdateItem registered before CreateItem
		return nil, false
	}
	result, err := cb.Call(args...)
	if err != nil {
		if vmexec.Quiet(err) {
			return nil, false
		}
		cb.Report(err)
		return widget.NewLabel(""), true
	}
	canvasObj, ok := result.(IsCanvasObject)
	if !ok {
		if result == nil {
			cb.Report(fmt.Errorf("template callback returned no widget"))
		} else {
			cb.Report(fmt.Errorf("expected widget, got %s", result.Type()))
		}
		return widget.NewLabel(""), true
	}
	return canvasObj.CanvasObject(), true
}
