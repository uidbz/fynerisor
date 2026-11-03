package tiestore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/uidbz/tie/api"
	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/scene"
)

// fakeClient exercises the store's error paths without a server.
type fakeClient struct {
	uid    client.DirUID
	row    client.Row
	getErr error
}

func (f *fakeClient) DirUIDFromPath(string) (client.DirUID, error) { return f.uid, nil }
func (f *fakeClient) MkTieDirAll(string) (client.DirUID, error)    { return f.uid, nil }
func (f *fakeClient) GetContext(context.Context, string) (client.Row, error) {
	return f.row, f.getErr
}
func (f *fakeClient) QueryContext(context.Context, client.QuerySpec) ([]client.Row, int, error) {
	return nil, 0, client.ErrNotFound
}
func (f *fakeClient) DownloadContext(context.Context, string, string, string) error { return nil }
func (f *fakeClient) ResolveHost(string) (client.FileHost, error)                   { return client.FileHost{}, nil }
func (f *fakeClient) WriteFile(client.FileHost, client.DirUID, string, string, []string) (string, error) {
	return "", errors.New("write should not be reached")
}
func (f *fakeClient) NewBatch() *api.Batch { return &api.Batch{} }
func (f *fakeClient) BatchContext(context.Context, *api.Batch) (client.BatchReply, error) {
	return nil, errors.New("batch should not be reached")
}
func (f *fakeClient) CollectionMembers(string) ([]client.Row, error) { return nil, nil }
func (f *fakeClient) SetDirType(client.DirUID, string) error         { return nil }
func (f *fakeClient) Sync() error                                    { return nil }

var transportErr = errors.New("connection refused")

// A transport error while reading the pointer must surface as itself, not as
// a version conflict (whose Reload choice discards the user's edits).
func TestSaveDoesNotMistakeTransportErrorForConflict(t *testing.T) {
	uid := client.DirUID(fmt.Sprintf("%064d", 1))
	store := &Store{Client: &fakeClient{uid: uid, getErr: transportErr}}
	_, err := store.Save(context.Background(), "tie:/scenes/x", []byte(`{}`), "some-version")
	if !errors.Is(err, transportErr) {
		t.Fatalf("Save error = %v, want the transport error", err)
	}
	if errors.Is(err, scene.ErrConflict) {
		t.Fatal("transport error reported as a version conflict")
	}
}

// The same failure must not make Load report "not found" (which made the
// browser create a fresh starter page over an existing home scene).
func TestLoadDoesNotMistakeTransportErrorForNotFound(t *testing.T) {
	uid := client.DirUID(fmt.Sprintf("%064d", 1))
	store := &Store{Client: &fakeClient{uid: uid, getErr: transportErr}}
	_, err := store.Load(context.Background(), "tie:/scenes/x")
	if !errors.Is(err, transportErr) {
		t.Fatalf("Load error = %v, want the transport error", err)
	}
	if errors.Is(err, scene.ErrNotFound) {
		t.Fatal("transport error reported as not found")
	}
}

// A directory with no pointer is genuinely "no document".
func TestLoadMissingPointerIsNotFound(t *testing.T) {
	uid := client.DirUID(fmt.Sprintf("%064d", 1))
	store := &Store{Client: &fakeClient{uid: uid, getErr: client.ErrNotFound}}
	_, err := store.Load(context.Background(), "tie:/scenes/x")
	if !errors.Is(err, scene.ErrNotFound) {
		t.Fatalf("Load error = %v, want ErrNotFound", err)
	}
}
