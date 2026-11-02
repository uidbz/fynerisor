package render

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/theme"

	"github.com/uidbz/fynerisor/scene"
)

func TestStyleTheme(t *testing.T) {
	th := newStyleTheme(&scene.Style{Fg: "#ff0000", Bg: "#000000", TextSize: 24, Bold: true})
	if got := ColorHex(th.Color(theme.ColorNameForeground, theme.VariantDark)); got != "#ff0000ff" {
		t.Fatalf("foreground = %s", got)
	}
	if got := ColorHex(th.Color(theme.ColorNameBackground, theme.VariantDark)); got != "#000000ff" {
		t.Fatalf("background = %s", got)
	}
	if got := th.Size(theme.SizeNameText); got != 24 {
		t.Fatalf("text size = %v", got)
	}
}

func TestStoreResolverLoadEmbed(t *testing.T) {
	dir := t.TempDir()
	child := filepath.Join(dir, "child.scene.json")
	if err := os.WriteFile(child, []byte(`{"format":"fynerisor.scene/1","title":"Child","root":{"id":"root","type":"free"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &StoreResolver{}
	doc, base, err := r.LoadEmbed(context.Background(), filepath.Join(dir, "parent.scene.json"), "child.scene.json")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "Child" || base != child {
		t.Fatalf("doc=%q base=%q", doc.Title, base)
	}
}
