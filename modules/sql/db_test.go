package sql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestQueryContextCancellation covers issue #26: Exec/Query must be
// cancellable through the builtin's context, so stopping a script aborts
// an in-flight query instead of letting it run to completion.
func TestQueryContextCancellation(t *testing.T) {
	db, err := New(context.Background(), "sqlite3://:memory:", false)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	// A recursive CTE far too large to finish within the cancellation
	// delay (measured: ~10M rows/s, so 2B rows is ~3 minutes).
	const heavy = `WITH RECURSIVE s(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM s WHERE i < 2000000000) SELECT count(*) FROM s`
	res, err := db.Query(ctx, object.NewString(heavy))
	if err == nil {
		if res != nil {
			t.Fatalf("heavy query unexpectedly completed: %v", res.Inspect())
		}
		t.Fatal("heavy query unexpectedly completed without error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
