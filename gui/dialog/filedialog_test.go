package dialog

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
)

func TestExtensionPatterns(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{".txt", []string{"*.txt"}},
		{"txt", []string{"*.txt"}},
		{"*.txt", []string{"*.txt"}},
		{".txt .md", []string{"*.txt", "*.md"}},
		{".png;.jpg;.jpeg", []string{"*.png", "*.jpg", "*.jpeg"}},
		{"", []string{}},
	}
	for _, c := range cases {
		if got := extensionPatterns(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("extensionPatterns(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestFileDialogWrapperOpenLifecycle(t *testing.T) {
	test.NewApp()
	guithread.SetMain()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	parent := test.NewWindow(widget.NewLabel("host"))
	defer parent.Close()
	parent.Resize(fyne.NewSize(1200, 800))

	var last string
	fd := NewFileOpen(func(path string, err error) {
		if err != nil {
			t.Errorf("unexpected dialog error: %v", err)
		}
		last = path
	}, parent)

	// Interface() is a live canvas object, as it was for the stock wrapper.
	if _, ok := fd.Interface().(fyne.CanvasObject); !ok {
		t.Fatal("Interface() does not return a fyne.CanvasObject")
	}

	// Script-style configuration followed by Show; accepting delivers the
	// absolute path of the selection.
	fd.instance.SetDirectory(dir)
	fd.instance.SetFileName("alpha.txt")
	fd.ShowUI()
	if len(parent.Canvas().Overlays().List()) != 1 {
		t.Fatalf("overlay stack = %d, want 1", len(parent.Canvas().Overlays().List()))
	}
	fd.instance.Accept()
	if last != filepath.Join(dir, "alpha.txt") {
		t.Fatalf("callback received %q", last)
	}
	if len(parent.Canvas().Overlays().List()) != 0 {
		t.Fatal("overlay not dismissed after accept")
	}

	// Cancelling delivers the empty path (script sees nil, nil).
	fd.ShowUI()
	fd.instance.Cancel()
	if last != "" {
		t.Fatalf("callback received %q after cancel, want empty", last)
	}

	// Let kdefilepicker's focus-retry timers fire before teardown.
	time.Sleep(350 * time.Millisecond)
}
