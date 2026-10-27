package gui

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// newTestApp returns a Fyne test app and makes the test quiesce every script
// run it started before the next test begins.
//
// Fyne's test driver runs fyne.Do inline on the calling goroutine. A run
// left alive by one test (a queued event, a refresh-when-idle) would do UI
// work on its VM goroutine while the next test renders on another — a race
// inside Fyne that only exists in tests. Cleanup stops every Window created
// during the test and waits for its VM goroutine to exit.
func newTestApp(t *testing.T) fyne.App {
	t.Helper()
	var mu sync.Mutex
	var windows []*Window
	windowCreated = func(w *Window) {
		mu.Lock()
		windows = append(windows, w)
		mu.Unlock()
	}
	a := test.NewApp()
	t.Cleanup(func() {
		windowCreated = nil
		mu.Lock()
		defer mu.Unlock()
		for _, w := range windows {
			if !w.Quiesce(5 * time.Second) {
				t.Errorf("a script run did not stop within 5s")
			}
		}
		a.Quit()
	})
	return a
}
