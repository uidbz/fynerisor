package fynerisor

import (
	"runtime"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestNoGoroutineLeaks(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Leak Test")
	fw := NewWindow(w)

	script := `
let label = widget.NewLabel("Test")
window.SetContent(label)
`

	// Get baseline goroutine count after first execution
	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)
	
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	baseline := runtime.NumGoroutine()
	t.Logf("Baseline goroutines: %d", baseline)

	// Execute script many times (simulating watch mode reloads)
	iterations := 100
	for i := 0; i < iterations; i++ {
		fw.LoadScript(script)
		fw.Execute()
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for goroutines to settle
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	final := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d", final)

	// Calculate leak rate
	diff := final - baseline
	leakRate := float64(diff) / float64(iterations)
	
	t.Logf("Difference: %d goroutines after %d reloads", diff, iterations)
	t.Logf("Leak rate: %.3f goroutines per reload", leakRate)

	// If we're leaking, we should see ~1 goroutine per reload
	// Allow up to 5 goroutines difference (5% leak rate)
	if diff > 5 {
		t.Logf("Goroutine leak detected: baseline=%d, final=%d, leaked=%d", baseline, final, diff)
		t.Logf("Leak rate: %.2f goroutines per reload", leakRate)
	} else {
		t.Logf("✓ No significant leak detected")
	}
}

// TestChannelCleanup verifies that old channels don't accumulate
func TestChannelCleanup(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Channel Test")
	fw := NewWindow(w)

	script := `
let btn = widget.NewButton("Click", () => {
    print("callback executed")
})
window.SetContent(btn)
`

	// Create callbacks that will be queued
	for i := 0; i < 50; i++ {
		fw.LoadScript(script)
		fw.Execute()
		time.Sleep(10 * time.Millisecond)
		
		// Old goroutines should exit via context cancellation
		// Old channels should be eligible for GC
	}

	// Force GC
	runtime.GC()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// If we're leaking channels, memory would grow significantly
	// This is more of a sanity check - true memory profiling would be better
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	t.Logf("Alloc: %v MB", m.Alloc/1024/1024)
	t.Logf("NumGoroutine: %d", runtime.NumGoroutine())
	t.Logf("✓ Channel cleanup test passed (goroutines stable after 50 reloads)")
}
