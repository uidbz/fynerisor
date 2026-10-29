package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/gui"
)

// This file covers issue #28: navigation is async with a supersede/
// cancel guard, auth is a callback flow with a sentinel error, and the
// shared loadPage path records history exactly once.

func newTestBrowser(t *testing.T, loader func(ctx context.Context, url, u, p string) (string, error)) (*Browser, func()) {
	t.Helper()
	a := test.NewApp()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	b := New(w, Config{LoadContent: loader})
	return b, func() { w.Stop(); a.Quit() }
}

func waitFor(t *testing.T, d time.Duration, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

func pageScript(name string) string {
	return fmt.Sprintf("window.SetContent(widget.NewLabel(%q))", name)
}

func TestSentinelUnauthorized(t *testing.T) {
	wrapped := fmt.Errorf("%w: Basic realm=x", ErrUnauthorized)
	if !errors.Is(wrapped, ErrUnauthorized) {
		t.Errorf("errors.Is failed for %v", wrapped)
	}
	if errors.Is(errors.New("some other"), ErrUnauthorized) {
		t.Error("unrelated error matched the sentinel")
	}
}

// A slow fetch must not block, and a newer navigation must supersede it
// (superseded results never load, and its ctx is cancelled).
func TestNavigateSupersede(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseNow := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseNow)

	b, cleanup := newTestBrowser(t, func(ctx context.Context, url, _, _ string) (string, error) {
		if url == "http://x/slow" {
			select {
			case <-release:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		return pageScript(url), nil
	})
	defer cleanup()

	start := time.Now()
	b.Navigate("http://x/slow")
	if time.Since(start) > 500*time.Millisecond {
		t.Error("navigate blocked on the fetch")
	}
	b.Navigate("http://x/fast") // supersedes

	waitFor(t, 3*time.Second, func() bool { return b.history.Current() == "http://x/fast" }, "fast page to load")

	close2 := releaseNow
	close2() // let the superseded fetch finish late
	time.Sleep(300 * time.Millisecond)
	if got := b.history.Current(); got != "http://x/fast" {
		t.Errorf("superseded navigation loaded: current = %q", got)
	}
}

// The 401 sentinel triggers the callback auth dialog; the retry with
// credentials loads through the shared path and records history.
func TestAuthSentinelCallbackFlow(t *testing.T) {
	const secret = "http://x/secret"

	loader := func(ctx context.Context, url, user, pass string) (string, error) {
		if url == secret {
			if user != "ada" || pass != "pw" {
				return "", fmt.Errorf("%w: Basic", ErrUnauthorized)
			}
			return pageScript("secret"), nil
		}
		return pageScript(url), nil
	}

	authCalls := 0
	auth := &funcAuth{fn: func(ctx context.Context, url string, parent fyne.Window, done func(string, string, bool)) {
		authCalls++
		if parent == nil {
			t.Error("auth dialog got no parent window")
		}
		done("ada", "pw", true)
	}}

	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	b := New(w, Config{LoadContent: loader, AuthProvider: auth})
	defer func() { w.Stop(); a.Quit() }()

	b.Navigate(secret)
	waitFor(t, 3*time.Second, func() bool { return b.history.Current() == secret }, "authenticated page to load")
	if authCalls != 1 {
		t.Errorf("auth provider called %d times, want 1", authCalls)
	}
}

// Cancelled auth leaves the page unauthenticated and says so.
func TestAuthCancelled(t *testing.T) {
	const secret = "http://x/secret"
	loader := func(ctx context.Context, url, user, pass string) (string, error) {
		return "", fmt.Errorf("%w: Basic", ErrUnauthorized)
	}
	auth := &funcAuth{fn: func(ctx context.Context, url string, parent fyne.Window, done func(string, string, bool)) {
		done("", "", false)
	}}

	a := test.NewApp()
	defer a.Quit()
	host := a.NewWindow("host")
	w := gui.NewWindow(host)
	b := New(w, Config{LoadContent: loader, AuthProvider: auth})
	defer func() { w.Stop(); a.Quit() }()

	b.Navigate(secret)
	waitFor(t, 3*time.Second, func() bool {
		s, _ := b.status.Get()
		return s == "Authentication cancelled"
	}, "cancelled status")
	if got := b.history.Current(); got == secret {
		t.Error("cancelled auth must not record the visit")
	}
}

type funcAuth struct {
	fn func(ctx context.Context, url string, parent fyne.Window, done func(string, string, bool))
}

func (f *funcAuth) ShowAuthDialog(ctx context.Context, url string, parent fyne.Window, done func(string, string, bool)) {
	f.fn(ctx, url, parent, done)
}
