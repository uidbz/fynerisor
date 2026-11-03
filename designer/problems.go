package designer

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

type problemsView struct {
	ed     *Editor
	issues []scene.Issue
	list   *widget.List
	empty  *widget.Label
	box    *fyne.Container
}

func newProblems(e *Editor) *problemsView {
	p := &problemsView{ed: e}
	p.empty = widget.NewLabel("No problems.")
	p.list = widget.NewList(
		func() int { return len(p.issues) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			l := obj.(*widget.Label)
			if id >= 0 && id < len(p.issues) {
				l.SetText(p.issues[id].String())
				l.Wrapping = fyne.TextWrapWord
			}
		},
	)
	p.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(p.issues) {
			return
		}
		if p.issues[id].NodeID != "" {
			p.ed.Select(p.issues[id].NodeID)
		}
		p.list.Unselect(id)
	}
	p.box = container.NewBorder(p.empty, nil, nil, nil, p.list)
	return p
}

func (p *problemsView) object() fyne.CanvasObject { return p.box }

func (p *problemsView) set(issues []scene.Issue) {
	p.issues = issues
	p.list.Refresh()
	if len(issues) == 0 {
		p.empty.SetText("No problems.")
	} else {
		errs, warns := 0, 0
		for _, i := range issues {
			if i.Severity == scene.Error {
				errs++
			} else {
				warns++
			}
		}
		p.empty.SetText(fmt.Sprintf("%d error(s), %d warning(s)", errs, warns))
	}
}
