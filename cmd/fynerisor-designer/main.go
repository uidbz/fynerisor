// Command fynerisor-designer is the standalone scene designer.
//
//	fynerisor-designer page.scene.json
//
// A missing file starts a blank scene; the first Save creates it. The same
// editor is embedded in the browser (F4 / the pencil button).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/uidbz/fynerisor/designer"
	"github.com/uidbz/fynerisor/scene"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: fynerisor-designer [page.scene.json]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	ref := flag.Arg(0)
	if ref == "" {
		ref = "untitled.scene.json"
	}
	title := filepath.Base(ref)

	a := app.NewWithID("com.fynerisor.designer")
	w := a.NewWindow("fynerisor designer — " + title)
	w.Resize(fyne.NewSize(1280, 820))

	store := scene.FileStore{}
	var (
		doc     *scene.Document
		version string
	)
	snap, err := store.Load(context.Background(), ref)
	switch {
	case err == nil:
		version = snap.Version
		doc, err = scene.Decode(snap.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor-designer: %v\n", err)
			os.Exit(1)
		}
	case errors.Is(err, scene.ErrNotFound):
		doc = &scene.Document{
			Format: scene.Format,
			Title:  strings.TrimSuffix(title, filepath.Ext(title)),
			Root: &scene.Node{
				ID: "root", Type: "free",
				Props: map[string]any{"designW": 1600.0, "designH": 900.0, "scale": "fit"},
			},
		}
	default:
		fmt.Fprintf(os.Stderr, "fynerisor-designer: %v\n", err)
		os.Exit(1)
	}

	ed := designer.New(doc, designer.Options{
		Store:   store,
		Ref:     ref,
		Version: version,
		Window:  w,
		Status:  func(s string) { log.Println("[designer]", s) },
		OnClose: func(reload bool) { w.Close() },
		OnDirty: func(dirty bool) {
			if dirty {
				w.SetTitle("fynerisor designer — " + title + " •")
			} else {
				w.SetTitle("fynerisor designer — " + title)
			}
		},
	})
	w.SetContent(ed.Object())
	w.ShowAndRun()
}
