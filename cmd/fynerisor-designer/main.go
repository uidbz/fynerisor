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
	"github.com/uidbz/fynerisor/scene/handlers"
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
		// The CLI has no script editor. Show the sidecar and handler so the
		// author can open them; browser hosts replace this hook.
		EditHandler: func(script, handler string) {
			log.Printf("[designer] edit %s: %s", script, handler)
		},
		BeforeSave: func(doc *scene.Document) (bool, error) {
			// ApplySidecar names empty call handlers even when it cannot
			// rewrite the sidecar, so changed is valid with these errors.
			rep, changed, err := handlers.ApplySidecar(doc, ref)
			if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, handlers.ErrRemoteSidecar) {
				return false, err
			}
			if len(rep.Created) > 0 || len(rep.Converted) > 0 {
				log.Printf("[designer] handlers: created %v, converted %v", rep.Created, rep.Converted)
			}
			for _, w := range rep.Warnings {
				log.Printf("[designer] handlers: %s", w)
			}
			return changed, nil
		},
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
