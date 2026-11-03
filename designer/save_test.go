package designer

import (
	"context"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

func testRegistry() *render.Registry { return renderbuiltin.NewRegistry() }

// blockingStore blocks Save until release is closed, so a test can edit while
// a save is in flight.
type blockingStore struct {
	*scene.MemoryStore
	entered chan struct{}
	release chan struct{}

	mu    sync.Mutex
	saves int
}

func newBlockingStore() *blockingStore {
	return &blockingStore{
		MemoryStore: scene.NewMemoryStore(),
		entered:     make(chan struct{}, 8),
		release:     make(chan struct{}),
	}
}

func (s *blockingStore) Save(ctx context.Context, ref string, data []byte, base string) (*scene.Snapshot, error) {
	s.mu.Lock()
	s.saves++
	s.mu.Unlock()
	select {
	case s.entered <- struct{}{}:
	default:
	}
	<-s.release
	return s.MemoryStore.Save(ctx, ref, data, base)
}

func (s *blockingStore) saveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saves
}

// saveSignal synchronizes a test with completed saves. OnSaveDone fires after
// the save's UI updates, so reads after a receive are ordered after them.
type saveSignal struct{ done chan struct{} }

func newSaveSignal(e *Editor) *saveSignal {
	s := &saveSignal{done: make(chan struct{}, 8)}
	e.opts.OnSaveDone = func() { s.done <- struct{}{} }
	return s
}

func (s *saveSignal) wait(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-s.done:
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for save %d of %d", i+1, n)
		}
	}
}

func newBlockingEditor(t *testing.T) (*Editor, *blockingStore) {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	store := newBlockingStore()
	e := New(testDoc(), Options{
		Registry: testRegistry(),
		Store:    store,
		Ref:      "mem://scene",
	})
	w := test.NewWindow(e.Object())
	w.Resize(fyne.NewSize(1200, 800))
	e.view.RefreshRects()
	e.overlay.Resize(e.stage.Size())
	return e, store
}

// TestSaveKeepsEditsMadeDuringSaveDirty: the save writes the bytes it encoded;
// edits typed while it is in flight must keep the document dirty.
func TestSaveKeepsEditsMadeDuringSaveDirty(t *testing.T) {
	e, store := newBlockingEditor(t)
	sig := newSaveSignal(e)
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)
	e.Save()
	select {
	case <-store.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("save did not start")
	}
	// Edit while the save is blocked on the network.
	e.overlay.nudge(1, 0)
	close(store.release)
	sig.wait(t, 1)
	if !e.Dirty() {
		t.Fatal("edit made during the save was marked saved")
	}
	if store.saveCount() != 1 {
		t.Fatalf("saves = %d, want 1", store.saveCount())
	}
}

// TestSaveQueuesOverlappingRequests: Save during a save is coalesced into one
// follow-up save, not dropped.
func TestSaveQueuesOverlappingRequests(t *testing.T) {
	e, store := newBlockingEditor(t)
	sig := newSaveSignal(e)
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)
	e.Save()
	select {
	case <-store.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("save did not start")
	}
	e.overlay.nudge(1, 0)
	e.Save() // queued
	close(store.release)
	sig.wait(t, 2)
	if store.saveCount() != 2 {
		t.Fatalf("saves = %d, want 2", store.saveCount())
	}
	if e.Dirty() {
		t.Fatal("queued save did not persist the latest edit")
	}
}

// TestEditorEntryForwardsSave: Ctrl+S works while an inspector field has focus
// (the driver gives shortcuts to the focused widget, and Entry swallows
// unknown ones).
func TestEditorEntryForwardsSave(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	store := scene.NewMemoryStore()
	e := New(testDoc(), Options{Registry: testRegistry(), Store: store, Ref: "mem://scene"})
	sig := newSaveSignal(e)
	e.SetGrid(0)
	e.Select("door")
	e.overlay.nudge(1, 0)

	entry := newEditorEntry(e)
	entry.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl})
	sig.wait(t, 1)
	if e.Dirty() {
		t.Fatal("document still dirty after Ctrl+S from a field")
	}
}

// TestSelectingSlotChildStaysClean: the slot select fires OnChanged from
// SetSelected during a refresh; that used to push a no-op SetPlace and mark
// the document dirty just by selecting a child of a border/split/scroll.
func TestSelectingSlotChildStaysClean(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	doc := &scene.Document{Root: &scene.Node{
		ID: "root", Type: "free",
		Props: map[string]any{"designW": 800.0, "designH": 600.0, "scale": "none"},
		Children: []*scene.Node{{
			ID: "border", Type: "border", Place: scene.Place{X: 20, Y: 20, W: 400, H: 300},
			Children: []*scene.Node{{
				ID: "child", Type: "label", Place: scene.Place{Slot: "top"},
				Props: map[string]any{"text": "top"},
			}},
		}},
	}}
	e := New(doc, Options{Registry: testRegistry(), Store: scene.NewMemoryStore(), Ref: "mem://scene"})
	e.Select("child")
	if e.Dirty() {
		t.Fatal("selecting a slot child dirtied the document")
	}
}
