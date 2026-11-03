// Example 44: a scene page hosted by a small Go program.
//
// The program decodes page.scene.json, builds a live page, runs the sidecar
// script with the page object in scope, and shows it. This is the same flow the
// browser uses for scene pages; a real browser also supplies the navigator.
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	ref := filepath.Join(dir, "page.scene.json")
	data, err := os.ReadFile(ref)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := scene.Decode(data)
	if err != nil {
		log.Fatal(err)
	}
	reg := renderbuiltin.NewRegistry()
	cfg := guiscene.Config{Registry: reg}

	fw := gui.NewApp("fynerisor scene example",
		gui.WithGlobal("scene", guiscene.NewScene(cfg, ref)),
	)
	cfg.Window = fw
	page, err := guiscene.BuildPage(context.Background(), doc, cfg, ref, scene.VersionOf(data))
	if err != nil {
		log.Fatal(err)
	}

	script := "window.SetContent(page)\n"
	if doc.Script != "" {
		sidecar, err := os.ReadFile(filepath.Join(dir, doc.Script))
		if err != nil {
			log.Fatal(err)
		}
		script = string(sidecar) + "\nwindow.SetContent(page)\n"
	}
	fw.LoadScript(script)
	fw.ExecuteWith(map[string]any{
		"page":  page,
		"scene": guiscene.NewScene(cfg, ref),
	})
	fw.ShowAndRun()
}
