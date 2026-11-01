package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/builtin"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// runScene loads and shows a scene document. It is also reached from runRun
// when the script argument ends in .scene.json.
func runScene(argv []string) int {
	args := []string{}
	for _, a := range argv {
		if strings.HasPrefix(a, "-") {
			continue
		}
		args = append(args, a)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: fynerisor scene <page.scene.json>")
		return 2
	}
	return showScene(args[0], "", 800, 600, false, false)
}

func showScene(path, title string, width, height int, headless, verbose bool) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
		return 1
	}
	reg := renderbuiltin.NewRegistry()
	doc, issues, err := scene.DecodeValidated(data, reg.Registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return 1
	}
	for _, issue := range issues {
		if issue.Severity == scene.Error || verbose {
			fmt.Fprintf(os.Stderr, "%s: %s\n", path, issue)
		}
	}
	if scene.Errors(issues) > 0 {
		return 1
	}
	if title == "" {
		title = filepath.Base(path)
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	var a fyne.App
	var win fyne.Window
	if headless {
		a = test.NewApp()
	} else {
		a = app.New()
	}
	win = a.NewWindow(title)
	if !headless {
		win.Resize(fyne.NewSize(float32(width), float32(height)))
	}
	cfg := guiscene.Config{Registry: reg}
	sc := guiscene.NewScene(cfg, path)
	fw := gui.NewWindow(win, gui.WithAppName("fynerisor-scene"), gui.WithGlobal("scene", sc))
	view := render.NewView(doc, render.Options{Registry: reg, Mode: render.ModePlay, Base: path})

	if headless {
		view.Resize(fyne.NewSize(float32(width), float32(height)))
		fmt.Printf("scene %q: %d nodes, %d warnings\n", doc.Title, doc.Count(), len(issues))
		return 0
	}

	script := "window.SetContent(page)\n"
	if doc.Script != "" {
		ref, err := scene.ResolveRef(path, doc.Script)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: scene script: %v\n", err)
			return 1
		}
		scriptPath, err := scene.FilePath(ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: scene script: %v\n", err)
			return 1
		}
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: scene script: %v\n", err)
			return 1
		}
		// The sidecar may set content itself; the trailing bootstrap is
		// idempotent and makes a script that only binds handlers work.
		script = string(src) + "\nwindow.SetContent(page)\n"
	}
	fw.LoadScript(script)
	fw.ExecuteWith(map[string]any{"page": view})
	if verbose {
		fmt.Printf("scene %q: %d nodes\n", doc.Title, doc.Count())
	}
	fw.ShowAndRun()
	return 0
}

func checkSceneFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return 1
	}
	reg := builtin.Registry()
	doc, issues, err := scene.DecodeValidated(data, reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return 1
	}
	for _, issue := range issues {
		fmt.Fprintf(os.Stderr, "%s: %s\n", path, issue)
	}
	rc := 0
	if scene.Errors(issues) > 0 {
		rc = 1
	}
	if doc.Script != "" {
		ref, err := scene.ResolveRef(path, doc.Script)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: script: %v\n", path, err)
			return 1
		}
		scriptPath, err := scene.FilePath(ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: script: %v\n", path, err)
			return 1
		}
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: script: %v\n", path, err)
			return 1
		}
		cb := core.NewContext(
			core.WithHTTP(), core.WithOS(), core.WithStrings(), core.WithFilepath(),
			core.WithTime(), core.WithTie(), core.WithIO(),
			core.WithJSON(), core.WithCSV(), core.WithExec(),
			// Globals the browser injects for scene runs; compile checks need
			// their names to exist even though the values are placeholders.
			core.WithGlobal("page", map[string]any{}),
			core.WithGlobal("scene", map[string]any{}),
		)
		if err := cb.CompileScript(string(src)); err != nil {
			fmt.Fprintln(os.Stderr, formatScriptError(scriptPath, err))
			rc = 1
		}
	}
	return rc
}
