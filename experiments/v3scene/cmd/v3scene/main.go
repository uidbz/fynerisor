// Command v3scene runs one scene page with a Risor v3 update program.
//
//	go run ./cmd/v3scene -page testdata/customers/customers.scene.json \
//	    -script testdata/customers/actions.risor -mode actions
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/uidbz/fynerisor/experiments/v3scene"
	"github.com/uidbz/fynerisor/gui/guithread"
)

func main() {
	page := flag.String("page", "testdata/customers/customers.scene.json", "scene document")
	script := flag.String("script", "testdata/customers/actions.risor", "Risor v3 update program")
	mode := flag.String("mode", string(v3scene.ModeActions), "effects mode: actions or host-functions")
	title := flag.String("title", "Risor v3 scene", "window title")
	check := flag.Bool("check", false, "compile the page and exit")
	flag.Parse()

	guithread.SetMain()

	doc, err := v3scene.LoadDocument(*page)
	if err != nil {
		log.Fatalf("load page: %v", err)
	}
	src, err := os.ReadFile(*script)
	if err != nil {
		log.Fatalf("read script: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := v3scene.Config{
		Document: doc,
		Script:   string(src),
		Mode:     v3scene.Mode(*mode),
		Base:     *page,
		Status:   func(s string) { fmt.Println("status:", s) },
		Report:   func(err error) { fmt.Fprintln(os.Stderr, "v3scene:", err) },
	}
	if *check {
		if err := v3scene.Check(cfg); err != nil {
			log.Fatalf("check page: %v", err)
		}
		fmt.Println("ok")
		return
	}
	pageApp, err := v3scene.New(cfg)
	if err != nil {
		log.Fatalf("compile page: %v", err)
	}

	a := app.New()
	w := a.NewWindow(*title)
	w.SetContent(pageApp.CanvasObject())
	w.Resize(fyne.NewSize(900, 640))
	w.Show()

	pageApp.Start(ctx)
	w.ShowAndRun()
}
