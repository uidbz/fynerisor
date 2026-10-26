package binding

// CallbackQueuer serializes callback invocations onto the callback queue of
// the owning window, so exactly one goroutine enters the Risor VM at a time.
// *gui.Window implements it via its Do method (issue #11).
type CallbackQueuer interface {
	Do(fn func())
}

// queueCallback runs fn through the window's callback queue when one is
// available, and inline otherwise. Binding listeners must never spawn raw
// goroutines that call into the VM: each such goroutine races every other
// entry point into the single-threaded VM.
func queueCallback(w CallbackQueuer, fn func()) {
	if w != nil {
		w.Do(fn)
		return
	}
	fn()
}
