package vmexec

import (
	"sync"

	"github.com/uidbz/fynerisor/gui/internal/goid"
)

// guard is a reentrant mutex keyed by goroutine: the owning goroutine may
// enter again (a handler that calls Table.Refresh, which synchronously calls
// the table's Data getter on the same goroutine), any other goroutine either
// waits (enter) or is turned away (tryEnter).
//
// One guard exists per Run, i.e. per Risor VM — never per process — so two
// windows or embedded CanvasObjects never block each other (issue #13).
type guard struct {
	mu    sync.Mutex
	cond  *sync.Cond
	owner uint64 // goroutine inside the VM; 0 = free
	depth int
}

func newGuard() *guard {
	g := &guard{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// enter blocks until the calling goroutine owns the guard. Only the run's
// VM goroutine blocks; UI-thread callers use tryEnter so the GUI never waits
// for script code.
func (g *guard) enter() {
	id := goid.ID()
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.owner != 0 && g.owner != id {
		g.cond.Wait()
	}
	g.owner = id
	g.depth++
}

// tryEnter takes the guard if it is free or already owned by the caller.
func (g *guard) tryEnter() bool {
	id := goid.ID()
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.owner != 0 && g.owner != id {
		return false
	}
	g.owner = id
	g.depth++
	return true
}

func (g *guard) exit() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.depth--
	if g.depth == 0 {
		g.owner = 0
		g.cond.Broadcast()
	}
}

// ownedByCaller reports whether the calling goroutine is inside the VM.
func (g *guard) ownedByCaller() bool {
	id := goid.ID()
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.owner == id
}
