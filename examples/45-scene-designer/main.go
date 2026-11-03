// Example 45: embed the scene designer in a host application.
//
// The designer is a normal Fyne canvas object. A host supplies a document, a
// store and a parent window; the editor provides the palette, canvas, outline,
// inspector, problems, undo/redo and saving.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/uidbz/fynerisor/designer"
	"github.com/uidbz/fynerisor/scene"
)

func main() {
	ref := "designer-demo.scene.json"
	if len(os.Args) > 1 {
		ref = os.Args[1]
	}
	store := scene.FileStore{}
	doc, version := loadOrNew(store, ref)

	a := app.NewWithID("com.fynerisor.example.designer")
	w := a.NewWindow("Designer example — " + filepath.Base(ref))
	w.Resize(fyne.NewSize(1280, 820))

	ed := designer.New(doc, designer.Options{
		Store:   store,
		Ref:     ref,
		Version: version,
		Window:  w,
		Status:  func(s string) { log.Println("[designer]", s) },
		OnClose: func(reload bool) { w.Close() },
		OnDirty: func(dirty bool) {
			if dirty {
				w.SetTitle("Designer example — " + filepath.Base(ref) + " •")
			} else {
				w.SetTitle("Designer example — " + filepath.Base(ref))
			}
		},
	})
	w.SetContent(ed.Object())
	w.ShowAndRun()
}

func loadOrNew(store scene.FileStore, ref string) (*scene.Document, string) {
	snap, err := store.Load(context.Background(), ref)
	switch {
	case err == nil:
		doc, err := scene.Decode(snap.Data)
		if err != nil {
			log.Fatal(err)
		}
		return doc, snap.Version
	case errors.Is(err, scene.ErrNotFound):
		return &scene.Document{
			Format: scene.Format,
			Title:  "Designer demo",
			Root: &scene.Node{
				ID: "root", Type: "free",
				Place: scene.Place{W: 900, H: 600},
				Props: map[string]any{
					"designW": 900.0, "designH": 600.0, "scale": "fit",
					"letterbox": "#101418",
				},
				Children: []*scene.Node{
					{
						ID: "title", Type: "label", Name: "Title",
						Place: scene.Place{X: 60, Y: 50, W: 500, H: 60},
						Props: map[string]any{"text": "Design me", "bold": true},
					},
					{
						ID: "door", Type: "hotspot", Name: "Door",
						Place: scene.Place{X: 520, Y: 200, W: 220, H: 180},
						Props: map[string]any{"hoverFill": "#38bdf855", "label": "Door"},
					},
				},
			},
		}, ""
	default:
		log.Fatal(err)
		return nil, ""
	}
}
