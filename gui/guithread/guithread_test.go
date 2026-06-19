package guithread

import "testing"

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
