package vmexec

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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

// The queue is unbounded: a flood of events queued while the body runs is
// neither dropped nor allowed to block its sender (issue #51).
func TestEnqueueUnboundedWhileBodyBusy(t *testing.T) {
	r, _ := silentRun(t)
	const n = 5000
	release := make(chan struct{})
	inBody := make(chan struct{})
	var ran atomic.Int64
	done := make(chan struct{})

	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
	})
	<-inBody

	start := time.Now()
	for i := 0; i < n; i++ {
		if err := r.Enqueue("ev", func() {
			if ran.Add(1) == n {
				close(done)
			}
		}); err != nil {
			t.Fatalf("Enqueue: %v", err)
		}
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("Enqueue took %v; it must never block", d)
	}
	close(release)
	waitClosed(t, done, "all queued events")
}

// A backlog deeper than the configured depth is reported once, but events
// keep their place in the queue (issue #51).
func TestQueueBacklogReportedOnceNotDropped(t *testing.T) {
	var mu sync.Mutex
	var reports []string
	r := New(context.Background(), Config{QueueSize: 2, Report: func(what string, err error) {
		mu.Lock()
		reports = append(reports, what+": "+err.Error())
		mu.Unlock()
	}})
	defer r.Stop()

	// Not started: nothing drains the queue, so the depth builds up.
	ran := make(chan int, 5)
	for i := 0; i < 5; i++ {
		i := i
		if err := r.Enqueue("ev", func() { ran <- i }); err != nil {
			t.Fatalf("Enqueue %d: %v", i, err)
		}
	}

	mu.Lock()
	seen := 0
	for _, rp := range reports {
		if strings.Contains(rp, "pending") {
			seen++
		}
	}
	mu.Unlock()
	if seen != 1 {
		t.Fatalf("backlog reported %d times, want exactly 1: %q", seen, reports)
	}

	// Nothing was dropped: all five jobs run, in order.
	r.Start(func(ctx context.Context) {})
	for want := 0; want < 5; want++ {
		select {
		case got := <-ran:
			if got != want {
				t.Fatalf("jobs ran out of order: got %d, want %d", got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("job %d never ran; events were dropped", want)
		}
	}
}

// Coalescing keeps at most one pending job per key: 10 000 changes queued
// while the body is busy run as one delivery carrying the latest value,
// and never fill the queue (issue #51).
func TestCoalescedEventsKeepLatestValue(t *testing.T) {
	r, _ := silentRun(t)
	const n = 10000
	release := make(chan struct{})
	inBody := make(chan struct{})

	var mu sync.Mutex
	var calls int
	latest := ""
	var order []string

	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
	})
	<-inBody

	key := new(int)
	for i := 0; i < n; i++ {
		v := fmt.Sprint(i)
		if err := r.EnqueueCoalesced(key, "change", func() {
			mu.Lock()
			calls++
			latest = v
			order = append(order, "change")
			mu.Unlock()
		}); err != nil {
			t.Fatalf("EnqueueCoalesced: %v", err)
		}
	}
	// A discrete event queued after the burst must run after the final
	// delivery of the burst.
	after := make(chan struct{})
	if err := r.Enqueue("after", func() {
		mu.Lock()
		order = append(order, "after")
		mu.Unlock()
		close(after)
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	close(release)
	waitClosed(t, after, "marker after coalesced burst")

	mu.Lock()
	defer mu.Unlock()
	if latest != fmt.Sprint(n-1) {
		t.Fatalf("last delivered value = %q, want %q", latest, fmt.Sprint(n-1))
	}
	if calls < 1 || calls > 2 {
		t.Fatalf("10 000 coalesced changes ran %d times, want a single delivery", calls)
	}
	if len(order) < 2 || order[0] != "change" || order[len(order)-1] != "after" {
		t.Fatalf("burst ran after the later discrete event: %v", order)
	}
}

// Coalescing is per callback: two different keys each deliver their latest
// value.
func TestCoalescedKeysAreIndependent(t *testing.T) {
	r, _ := silentRun(t)
	release := make(chan struct{})
	inBody := make(chan struct{})
	r.Start(func(ctx context.Context) {
		close(inBody)
		<-release
	})
	<-inBody

	var mu sync.Mutex
	got := map[string]string{}
	done := make(chan struct{}, 2)
	k1, k2 := new(int), new(int)
	for i := 0; i < 50; i++ {
		_ = r.EnqueueCoalesced(k1, "a", func() {
			mu.Lock()
			got["a"] = "1"
			mu.Unlock()
			done <- struct{}{}
		})
		_ = r.EnqueueCoalesced(k2, "b", func() {
			mu.Lock()
			got["b"] = "2"
			mu.Unlock()
			done <- struct{}{}
		})
	}
	close(release)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("coalesced key never delivered")
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if got["a"] != "1" || got["b"] != "2" {
		t.Fatalf("keys interfered: %v", got)
	}
}

func TestEnqueueCoalescedStoppedRun(t *testing.T) {
	r, _ := silentRun(t)
	r.Start(func(ctx context.Context) {})
	r.Stop()
	waitClosed(t, r.Done(), "run exit after Stop")
	if err := r.EnqueueCoalesced(new(int), "late", func() {
		t.Error("event of a stopped run executed")
	}); !errors.Is(err, ErrStopped) {
		t.Fatalf("expected ErrStopped, got %v", err)
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
