// Package goid returns the current goroutine's ID.
//
// Fyne and fynerisor need to know which goroutine is "the GUI thread" and
// which goroutine currently owns a Risor VM. Go deliberately hides goroutine
// IDs, so this parses the header of runtime.Stack ("goroutine 123 [...]"),
// the same technique Fyne uses internally (its async package is internal and
// cannot be imported).
package goid

import "runtime"

// ID returns the ID of the calling goroutine.
func ID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	const prefix = "goroutine "
	if n <= len(prefix) {
		return 0
	}
	var id uint64
	for _, c := range buf[len(prefix):n] {
		if c < '0' || c > '9' {
			break
		}
		id = id*10 + uint64(c-'0')
	}
	return id
}
