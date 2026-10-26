package gui

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynebinding "fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	risorbinding "github.com/uidbz/fynerisor/gui/binding"
)

func findEntryInContent(obj fyne.CanvasObject) *widget.Entry {
	switch o := obj.(type) {
	case *widget.Entry:
		return o
	case *fyne.Container:
		for _, c := range o.Objects {
			if e := findEntryInContent(c); e != nil {
				return e
			}
		}
	}
	return nil
}

// TestEntryOnChangedSerialized covers issue #11: typing many characters with
// an OnChanged handler must not spawn concurrent goroutines into the VM and
// the last handler call must observe the final value. Run with -race in CI.
func TestEntryOnChangedSerialized(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	fw.LoadScript(`
let e = widget.NewEntry()
e.OnChanged((text) => { window.SetStatus(text) })
window.SetContent(e)
`)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	entry := findEntryInContent(w.Content())
	if entry == nil {
		t.Fatal("entry not found in content")
	}

	const n = 200
	for i := 0; i < n; i++ {
		text := fmt.Sprintf("v%d", i)
		// Drive the entry the way a user would: changes to Fyne widgets
		// happen on the GUI thread; the queueing happens in OnChanged.
		fyne.Do(func() { entry.SetText(text) })
		time.Sleep(time.Millisecond)
	}

	deadline := time.Now().Add(3 * time.Second)
	want := fmt.Sprintf("v%d", n-1)
	for time.Now().Before(deadline) {
		if fw.Status != want {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		return
	}
	t.Fatalf("last OnChanged call did not observe the final value, status=%q", fw.Status)
}

// TestBindingListenerSerialized covers issue #11: many binding updates with
// an AddListener handler must serialize into the VM and the last listener
// call must observe the final value.
func TestBindingListenerSerialized(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	// The binding is created on the Go side so the test can drive it; the
	// script only attaches the listener.
	d := risorbinding.NewString(fw)
	fw.RegisterGlobal("d", d)

	fw.LoadScript(`
d.AddListener(() => { window.SetStatus(d.Get()) })
`)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	fb, ok := d.Interface().(fynebinding.String)
	if !ok {
		t.Fatalf("unexpected binding type %T", d.Interface())
	}

	const n = 200
	for i := 0; i < n; i++ {
		fb.Set(fmt.Sprintf("v%d", i))
		time.Sleep(time.Millisecond)
	}

	deadline := time.Now().Add(3 * time.Second)
	want := fmt.Sprintf("v%d", n-1)
	for time.Now().Before(deadline) {
		if fw.Status == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("last listener call did not observe the final value, status=%q", fw.Status)
}
