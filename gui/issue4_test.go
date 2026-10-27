package gui

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	guic "github.com/uidbz/fynerisor/gui/container"
	"github.com/uidbz/fynerisor/gui/guithread"
	guiw "github.com/uidbz/fynerisor/gui/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// callAttr invokes a bound method (builtin) on a container/widget object.
func callAttr(t *testing.T, obj object.Object, name string, args ...object.Object) {
	t.Helper()
	attr, ok := obj.GetAttr(name)
	if !ok {
		t.Fatalf("%s: no attribute %q", obj.Type(), name)
	}
	b, ok := attr.(*object.Builtin)
	if !ok {
		t.Fatalf("%s.%s: not a builtin, got %s", obj.Type(), name, attr.Type())
	}
	if _, err := b.Call(context.Background(), args...); err != nil {
		t.Fatalf("%s.%s: %v", obj.Type(), name, err)
	}
}

func objectsOf(t *testing.T, obj object.Object) []fyne.CanvasObject {
	t.Helper()
	c, ok := obj.Interface().(*fyne.Container)
	if !ok {
		t.Fatalf("%s: instance is not *fyne.Container", obj.Type())
	}
	return c.Objects
}

func labelOf(text string) object.Object {
	return guiw.NewLabel(text)
}

func unwrap(t *testing.T, o object.Object) fyne.CanvasObject {
	t.Helper()
	co, ok := o.Interface().(fyne.CanvasObject)
	if !ok {
		t.Fatalf("object %s does not wrap a canvas object", o.Type())
	}
	return co
}

// Issue #4: VBox/HBox/Stack/... gain Append/Remove/Refresh.
func TestContainerAppendRemoveRefresh(t *testing.T) {
	guithread.SetMain() // run bound GUI ops inline (no running event loop)
	_ = newTestApp(t)

	vb := guic.NewVBox(unwrap(t, labelOf("a")))
	if got := objectsOf(t, vb); len(got) != 1 {
		t.Fatalf("initial VBox children: want 1, got %d", len(got))
	}

	b := labelOf("b")
	callAttr(t, vb, "Append", b, labelOf("c"))
	if got := objectsOf(t, vb); len(got) != 3 {
		t.Fatalf("after Append: want 3, got %d", len(got))
	}

	callAttr(t, vb, "Remove", b)
	if got := objectsOf(t, vb); len(got) != 2 {
		t.Fatalf("after Remove: want 2, got %d", len(got))
	}

	callAttr(t, vb, "Refresh")

	// Bad argument: error result, not a panic.
	attr, _ := vb.GetAttr("Append")
	res, err := attr.(*object.Builtin).Call(context.Background(), object.NewString("nope"))
	if err == nil && res != nil && res.Type() == object.ERROR {
		err = res.(*object.Error).Unwrap()
	}
	if err == nil {
		t.Fatal("Append(string): expected error")
	}
}

// Issue #4: Border SetTop/SetCenter/... swap slots.
func TestBorderSetters(t *testing.T) {
	guithread.SetMain()
	_ = newTestApp(t)

	top := labelOf("top")
	newTop := labelOf("newtop")
	br := guic.NewBorder(unwrap(t, top), nil, nil, nil, unwrap(t, labelOf("center")))

	callAttr(t, br, "SetTop", newTop)
	topCO, newTopCO := unwrap(t, top), unwrap(t, newTop)
	objs := objectsOf(t, br)
	foundNew, foundOld := false, false
	for _, o := range objs {
		if o == newTopCO {
			foundNew = true
		}
		if o == topCO {
			foundOld = true
		}
	}
	if !foundNew {
		t.Fatal("new top missing after SetTop")
	}
	if foundOld {
		t.Fatal("old top still present after SetTop")
	}
	callAttr(t, br, "Refresh")
}

// Issue #4: Scroll gains SetContent.
func TestScrollSetContent(t *testing.T) {
	guithread.SetMain()
	_ = newTestApp(t)

	sc := guic.NewScroll(unwrap(t, labelOf("old")))
	next := labelOf("new")
	callAttr(t, sc, "SetContent", next)
	s, ok := sc.Interface().(*container.Scroll)
	if !ok {
		t.Fatalf("Scroll instance type: %T", sc.Interface())
	}
	if s.Content != unwrap(t, next) {
		t.Fatal("Scroll.SetContent did not replace Content")
	}
	callAttr(t, sc, "Refresh")
}
