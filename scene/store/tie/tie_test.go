package tiestore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/uidbz/tie/client"
)

// Integration test: skipped unless the tie test environment is running
// (see github.com/uidbz/tie/test-env).
func TestStoreRoundTrip(t *testing.T) {
	cfg := client.TestingConfig()
	// The test-env server auto-creates this account (see tie/test-env).
	cfg.Username = "defaultuser"
	cfg.Password = "defaultpassword"
	tc := client.NewTieClient(cfg)
	tc.SetTimeout(2 * time.Second)
	if _, err := tc.ServerVersion(); err != nil {
		t.Skipf("tie test environment not running: %v", err)
	}
	ctx := context.Background()
	st := New(tc)
	path := fmt.Sprintf("tie:/fynerisor-scene-tests/%d", time.Now().UnixNano())
	first := []byte(`{"format":"fynerisor.scene/1","title":"one","root":{"id":"root","type":"free"}}`)
	second := []byte(`{"format":"fynerisor.scene/1","title":"two","root":{"id":"root","type":"free"}}`)

	snap, err := st.Save(ctx, path, first, "")
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if snap.Version == "" || !snap.Writable {
		t.Fatalf("snapshot = %+v", snap)
	}
	loaded, err := st.Load(ctx, path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if string(loaded.Data) != string(first) || loaded.Version != snap.Version {
		t.Fatalf("loaded = %q version=%q", loaded.Data, loaded.Version)
	}
	if _, err := st.Save(ctx, path, second, "stale"); err == nil {
		t.Fatal("expected a conflict")
	}
	next, err := st.Save(ctx, path, second, snap.Version)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	loaded, err = st.Load(ctx, path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if string(loaded.Data) != string(second) || loaded.Version != next.Version {
		t.Fatalf("after update: %q version=%q", loaded.Data, loaded.Version)
	}
	if _, err := st.Load(ctx, "tie:/fynerisor-scene-tests/does-not-exist"); err == nil {
		t.Fatal("expected not found")
	}
}
