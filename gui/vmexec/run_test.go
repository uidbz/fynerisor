package vmexec

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func waitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func silentRun(t *testing.T) (*Run, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var reports []string
	r := New(context.Background(), Config{Report: func(what string, err error) {
		mu.Lock()
		reports = append(reports, what+": "+err.Error())
		mu.Unlock()
	}})
	t.Cleanup(r.Stop)
	return r, &reports
}

func TestFromContext(t *testing.T) {
	r, _ := silentRun(t)
	if FromContext(r.Context()) != r {
		t.Fatal("FromContext did not return the run attached to its context")
	}
	child := context.WithValue(r.Context(), struct{}{}, 1)
	if FromContext(child) != r {
		t.Fatal("FromContext lost the run through a derived context")
	}
	if FromContext(context.Background()) != nil {
		t.Fatal("FromContext on a bare context should be nil")
	}
}

// Events queued while the body runs execute after it, in order, on one
// goroutine — never concurrently with the body.
func TestEventsRunAfterBodyInOrder(t *testing.T) {
	r, _ := silentRun(t)
	release := make(chan struct{})
	var mu sync.Mutex
	var order []int
	inBody := make(chan struct{})

	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
		mu.Lock()
		order = append(order, 0)
		mu.Unlock()
	})
	<-inBody
	done := make(chan struct{})
	for i := 1; i <= 5; i++ {
		i := i
		if err := r.Enqueue("ev", func() {
			mu.Lock()
			order = append(order, i)
			n := len(order)
			mu.Unlock()
			if n == 6 {
				close(done)
			}
		}); err != nil {
			t.Fatalf("Enqueue: %v", err)
		}
	}
	close(release)
	waitClosed(t, done, "queued events")

	mu.Lock()
	defer mu.Unlock()
	for i, v := range order {
		if v != i {
			t.Fatalf("events ran out of order: %v", order)
		}
	}
}

// TryCall never waits: while the VM goroutine is inside the body, another
// goroutine (the UI thread) gets ErrBusy immediately.
func TestTryCallBusyDuringBody(t *testing.T) {
	r, _ := silentRun(t)
	inBody := make(chan struct{})
	release := make(chan struct{})
	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
	})
	<-inBody

	start := time.Now()
	_, err := r.TryCall(func() (object.Object, error) {
		t.Error("TryCall entered the VM while the body held it")
		return object.Nil, nil
	})
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("expected ErrBusy, got %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("TryCall blocked instead of failing fast")
	}
	close(release)

	// Once the body is done the VM is free again.
	deadline := time.Now().Add(2 * time.Second)
	for {
		res, err := r.TryCall(func() (object.Object, error) { return object.NewString("ok"), nil })
		if err == nil {
			if s, _ := res.(*object.String); s == nil || s.Value() != "ok" {
				t.Fatalf("unexpected result %v", res)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("VM never became free: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// A job may re-enter the VM on its own goroutine (handler → Refresh →
// getter) without being rejected.
func TestTryCallReentrantFromJob(t *testing.T) {
	r, _ := silentRun(t)
	r.Start(func(ctx context.Context) {})
	done := make(chan struct{})
	var innerErr error
	_ = r.Enqueue("outer", func() {
		_, innerErr = r.TryCall(func() (object.Object, error) { return object.Nil, nil })
		close(done)
	})
	waitClosed(t, done, "reentrant job")
	if innerErr != nil {
		t.Fatalf("reentrant TryCall rejected: %v", innerErr)
	}
}

// Two runs (two windows / embedded views) never block each other.
func TestRunsAreIndependent(t *testing.T) {
	a, _ := silentRun(t)
	b, _ := silentRun(t)
	inA := make(chan struct{})
	release := make(chan struct{})
	a.Start(func(ctx context.Context) {
		close(inA)
		<-release
	})
	b.Start(func(ctx context.Context) {})
	<-inA

	deadline := time.Now().Add(2 * time.Second)
	for {
		_, err := b.TryCall(func() (object.Object, error) { return object.Nil, nil })
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("run b blocked by run a: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	close(release)
}

func TestStopDropsEvents(t *testing.T) {
	r, _ := silentRun(t)
	r.Start(func(ctx context.Context) {})
	r.Stop()
	waitClosed(t, r.Done(), "run exit after Stop")
	if err := r.Enqueue("late", func() { t.Error("event of a stopped run executed") }); !errors.Is(err, ErrStopped) {
		t.Fatalf("expected ErrStopped, got %v", err)
	}
	if _, err := r.TryCall(func() (object.Object, error) { return object.Nil, nil }); !errors.Is(err, ErrStopped) {
		t.Fatalf("expected ErrStopped from TryCall, got %v", err)
	}
}

func TestStopCancelsBodyContext(t *testing.T) {
	r, _ := silentRun(t)
	started := make(chan struct{})
	r.Start(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
	})
	<-started
	r.Stop()
	waitClosed(t, r.Done(), "body exit after Stop")
}

func TestPanicsAreRecovered(t *testing.T) {
	r, reports := silentRun(t)
	r.Start(func(ctx context.Context) { panic("body boom") })
	done := make(chan struct{})
	_ = r.Enqueue("job", func() { panic("job boom") })
	_ = r.Enqueue("after", func() { close(done) })
	waitClosed(t, done, "job after panics")

	if _, err := r.TryCall(func() (object.Object, error) { panic("getter boom") }); err == nil ||
		!strings.Contains(err.Error(), "getter boom") {
		t.Fatalf("expected recovered getter panic, got %v", err)
	}
	r.Stop()
	waitClosed(t, r.Done(), "run exit")
	joined := strings.Join(*reports, "\n")
	if !strings.Contains(joined, "body boom") || !strings.Contains(joined, "job boom") {
		t.Fatalf("panics were not reported: %q", joined)
	}
}

func TestEnqueueNeverBlocksWhenFull(t *testing.T) {
	r := New(context.Background(), Config{QueueSize: 2, Report: func(string, error) {}})
	defer r.Stop()
	// Not started: nothing drains the queue.
	_ = r.Enqueue("a", func() {})
	_ = r.Enqueue("b", func() {})
	start := time.Now()
	if err := r.Enqueue("c", func() {}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("Enqueue blocked on a full queue")
	}
}

func TestWhenIdleCoalescesAndRunsAfterBody(t *testing.T) {
	r, _ := silentRun(t)
	release := make(chan struct{})
	inBody := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	ran := make(chan struct{}, 10)
	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
	})
	<-inBody
	key := new(int)
	for i := 0; i < 3; i++ {
		r.WhenIdle(key, func() {
			mu.Lock()
			calls++
			mu.Unlock()
			ran <- struct{}{}
		})
	}
	close(release)
	select {
	case <-ran:
	case <-time.After(2 * time.Second):
		t.Fatal("idle function never ran")
	}
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("WhenIdle with one key ran %d times, want 1", calls)
	}
}

func TestInVM(t *testing.T) {
	r, _ := silentRun(t)
	var inBody bool
	done := make(chan struct{})
	r.Start(func(ctx context.Context) {
		inBody = r.InVM()
		close(done)
	})
	waitClosed(t, done, "body")
	if !inBody {
		t.Fatal("InVM false inside the body")
	}
	if r.InVM() {
		t.Fatal("InVM true on a foreign goroutine")
	}
}
