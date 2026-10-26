package guithread

import (
	"context"
	"testing"
)

func TestDoRunsInlineOnMain(t *testing.T) {
	SetMain()

	ran := false
	Do(func() { ran = true })

	if !ran {
		t.Fatal("Do did not run the function inline on the main goroutine")
	}
}

func TestIsMain(t *testing.T) {
	SetMain()
	if !IsMain() {
		t.Fatal("expected IsMain to be true on the goroutine that called SetMain")
	}

	done := make(chan bool, 1)
	go func() { done <- IsMain() }()
	if <-done {
		t.Fatal("expected IsMain to be false on a different goroutine")
	}
}

// TestInitDoesNotOverwrite covers issue #14: creating a window from a
// background goroutine must not steal the "GUI goroutine" from main.
func TestInitDoesNotOverwrite(t *testing.T) {
	SetMain()

	done := make(chan struct{})
	go func() {
		Init(false) // what newWindow does on whatever goroutine calls it
		close(done)
	}()
	<-done

	if !IsMain() {
		t.Fatal("Init on a background goroutine overwrote the recorded GUI goroutine")
	}
}

func TestDoAndWaitContextCancelled(t *testing.T) {
	SetMain()
	ran := false
	if err := DoAndWaitContext(context.Background(), func() { ran = true }); err != nil || !ran {
		t.Fatalf("inline DoAndWaitContext on main: err=%v ran=%v", err, ran)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errc := make(chan error, 1)
	go func() { errc <- DoAndWaitContext(ctx, func() {}) }()
	if err := <-errc; err == nil {
		t.Fatal("DoAndWaitContext with a cancelled context should fail off-main")
	}
}
