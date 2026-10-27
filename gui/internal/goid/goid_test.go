package goid

import (
	"sync"
	"testing"
)

func TestIDDistinguishesGoroutines(t *testing.T) {
	main := ID()
	if main == 0 {
		t.Fatal("ID returned 0")
	}
	if ID() != main {
		t.Fatal("ID not stable on one goroutine")
	}
	var other uint64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); other = ID() }()
	wg.Wait()
	if other == 0 || other == main {
		t.Fatalf("goroutine IDs not distinct: main=%d other=%d", main, other)
	}
}

func BenchmarkID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ID()
	}
}
