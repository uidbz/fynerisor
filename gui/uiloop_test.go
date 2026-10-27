package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/gui/guithread"
)

// uiLoop gives tests a real UI goroutine.
//
// Fyne's test driver runs fyne.Do inline on the calling goroutine, so in
// plain tests the split between Fyne's UI thread and a script's VM goroutine
// never happens and -race cannot see races across it. startUILoop installs
// a Fyne app whose driver dispatches fyne.Do / DoAndWait onto one dedicated
// goroutine — like the GLFW driver — and records it as the GUI goroutine.
// Tests then perform UI actions (taps, typing, getter calls, content reads)
// with u.run, exactly as the real UI thread would.
type uiLoop struct {
	ch   chan func()
	stop chan struct{}
}

type splitApp struct {
	fyne.App
	drv fyne.Driver
}

func (a *splitApp) Driver() fyne.Driver { return a.drv }

type splitDriver struct {
	fyne.Driver
	u *uiLoop
}

func (d *splitDriver) DoFromGoroutine(f func(), wait bool) {
	if guithread.IsMain() {
		f()
		return
	}
	if !wait {
		select {
		case d.u.ch <- f:
		case <-d.u.stop:
		}
		return
	}
	done := make(chan struct{})
	select {
	case d.u.ch <- func() { f(); close(done) }:
	case <-d.u.stop:
		return
	}
	select {
	case <-done:
	case <-d.u.stop:
	}
}

func startUILoop(t *testing.T) (*uiLoop, fyne.App) {
	t.Helper()
	base := test.NewApp()
	u := &uiLoop{ch: make(chan func(), 4096), stop: make(chan struct{})}
	ready := make(chan struct{})
	go func() {
		guithread.SetMain()
		close(ready)
		for {
			select {
			case f := <-u.ch:
				f()
			case <-u.stop:
				return
			}
		}
	}()
	<-ready
	a := &splitApp{App: base, drv: &splitDriver{Driver: base.Driver(), u: u}}
	fyne.SetCurrentApp(a)
	t.Cleanup(func() {
		close(u.stop)
		base.Quit()
	})
	return u, a
}

// run executes f on the UI goroutine and waits for it.
func (u *uiLoop) run(f func()) {
	done := make(chan struct{})
	u.ch <- func() { f(); close(done) }
	<-done
}
